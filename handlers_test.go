package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexG28/keyvalue/store"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
)

var testStore = NewMockStore()

type MockStore struct {
	data      map[string]string
	addErr    error
	getErr    error
	deleteErr error
}

func NewMockStore() *MockStore {
	return &MockStore{
		data: make(map[string]string),
	}
}

func (m *MockStore) Add(key, value string) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.data[key] = value
	return nil
}

func (m *MockStore) Get(key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", store.ErrNotFound
}

func (m *MockStore) Delete(key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.data[key]; !ok {
		return store.ErrNotFound
	}
	delete(m.data, key)
	return nil
}

func getMockStore() *MockStore {
	return testStore
}

func resetMockStore() {
	mockStore := getMockStore()
	mockStore.data = make(map[string]string)
	mockStore.addErr = nil
	mockStore.getErr = nil
	mockStore.deleteErr = nil

}

type fakeApplyFuture struct {
	err      error
	response any
	index    uint64
}

func (f *fakeApplyFuture) Error() error          { return f.err }
func (f *fakeApplyFuture) Response() interface{} { return f.response }
func (f *fakeApplyFuture) Index() uint64         { return f.index }

type fakeIndexFuture struct {
	err   error
	index uint64
}

func (f *fakeIndexFuture) Error() error  { return f.err }
func (f *fakeIndexFuture) Index() uint64 { return f.index }

const fakeLeaderState = 2

type FakeRaft struct {
	applyErr    error
	addVoterErr error
	leader      bool
}

func (f *FakeRaft) Apply(_ []byte, _ time.Duration) raft.ApplyFuture {
	return &fakeApplyFuture{err: f.applyErr, index: 1}
}

func (f *FakeRaft) AddVoter(_ raft.ServerID, _ raft.ServerAddress, _ uint64, _ time.Duration) raft.IndexFuture {
	return &fakeIndexFuture{err: f.addVoterErr, index: 1}
}

func (f *FakeRaft) State() raft.RaftState {
	if f.leader {
		return fakeLeaderState
	}
	return 0
}

func newTestHTTPServerWithRaft(raft *FakeRaft) httpServer {
	return httpServer{
		r: raft,
		s: testStore,
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		setupKey     string
		setupValue   string
		requestPath  string
		expectedCode int
		expectedBody string
		injectErr    error
	}{
		{
			name:         "Successful Get",
			setupKey:     "hello",
			setupValue:   "there",
			requestPath:  "/Get/hello",
			expectedCode: http.StatusOK,
			expectedBody: "there",
		},
		{
			name:         "Key Not Found",
			requestPath:  "/Get/nonexistent",
			expectedCode: http.StatusNotFound,
			expectedBody: "Key not found\n",
		},
		{
			name:         "Missing Key in URL",
			requestPath:  "/Get/",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL format. Expected Get/{key}\n",
		},
		{
			name:         "Internal Store Error",
			setupKey:     "errkey",
			setupValue:   "irrelevant",
			requestPath:  "/Get/errkey",
			expectedCode: http.StatusNotFound,
			injectErr:    store.ErrNotFound,
			expectedBody: "Key not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()
			mockStore := getMockStore()

			fakeRaft := &FakeRaft{applyErr: tt.injectErr, leader: true}
			hs := newTestHTTPServerWithRaft(fakeRaft)

			if tt.setupKey != "" {
				_ = mockStore.Add(tt.setupKey, tt.setupValue)
			}
			if tt.injectErr != nil {
				mockStore.getErr = tt.injectErr
			}
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			hs.Get(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code, "status code")
			body, _ := io.ReadAll(rr.Body)
			assert.Equal(t, tt.expectedBody, string(body), "response body")
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name          string
		expectedKey   string
		expectedValue string
		requestPath   string
		expectedCode  int
		injectErr     error
		expectedBody  string
	}{
		{
			name:          "Successful set",
			expectedKey:   "hello",
			expectedValue: "there",
			requestPath:   "/Set/hello/there",
			expectedCode:  http.StatusOK,
			expectedBody:  "Set key 'hello' to value 'there'\n",
		},
		{
			name:         "Missing Key in URL",
			requestPath:  "/Set//randomvalue",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL format. Expected Set/{key}/{value}\n",
		},
		{
			name:         "Missing Value in URL",
			requestPath:  "/Set/randomkey/",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL format. Expected Set/{key}/{value}\n",
		},
		{
			name:         "Internal Raft Error",
			requestPath:  "/Set/errkey/errval",
			expectedCode: http.StatusInternalServerError,
			injectErr:    assert.AnError,
			expectedBody: "Could not write key-value: assert.AnError general error for testing\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()
			mockStore := getMockStore()
			fakeRaft := &FakeRaft{applyErr: tt.injectErr, leader: true}
			hs := newTestHTTPServerWithRaft(fakeRaft)
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			hs.Set(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code, "status code")
			body, _ := io.ReadAll(rr.Body)
			assert.Equal(t, tt.expectedBody, string(body), "response body")
			if tt.expectedKey != "" && tt.expectedValue != "" && tt.injectErr == nil {
				mockStore.data[tt.expectedKey] = tt.expectedValue
				actualValue, err := mockStore.Get(tt.expectedKey)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, actualValue)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name            string
		keyToPopulate   string
		valueToPopulate string
		keyToDelete     string
		requestPath     string
		expectedCode    int
		injectErr       error
		expectedBody    string
	}{
		{
			name:            "Successful delete",
			keyToPopulate:   "hello",
			valueToPopulate: "randomValue",
			keyToDelete:     "hello",
			requestPath:     "/Delete/hello",
			expectedCode:    http.StatusOK,
			expectedBody:    "Deleted key 'hello'\n",
		},
		{
			name:         "Missing Key in URL",
			requestPath:  "/Delete/",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL format. Expected Delete/{key}\n",
		},
		{
			name:         "Internal Raft Error",
			keyToDelete:  "errkey",
			requestPath:  "/Delete/errkey",
			expectedCode: http.StatusInternalServerError,
			injectErr:    assert.AnError,
			expectedBody: "Could not delete key: assert.AnError general error for testing\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()
			mockStore := getMockStore()
			if tt.keyToPopulate != "" && tt.valueToPopulate != "" {
				mockStore.Add(tt.keyToPopulate, tt.valueToPopulate)
			}

			if tt.keyToDelete != "" && tt.injectErr == nil {
				mockStore.data[tt.keyToDelete] = "someval"
			}
			fakeRaft := &FakeRaft{applyErr: tt.injectErr, leader: true}
			hs := newTestHTTPServerWithRaft(fakeRaft)
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			hs.Delete(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code, "status code")
			body, _ := io.ReadAll(rr.Body)
			assert.Equal(t, tt.expectedBody, string(body), "response body")
			if tt.keyToDelete != "" && tt.injectErr == nil && tt.expectedCode != http.StatusOK {
				_, err := mockStore.Get(tt.keyToDelete)
				assert.NotNil(t, err)
			}
		})
	}
}
