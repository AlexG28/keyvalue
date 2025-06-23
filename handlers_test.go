package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexG28/keyvalue/store"
	"github.com/stretchr/testify/assert"
)

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

func TestMain(m *testing.M) {
	localStore = NewMockStore()
	m.Run()
}

func getMockStore() *MockStore {
	return localStore.(*MockStore)
}

func resetMockStore() {
	mockStore := getMockStore()
	mockStore.data = make(map[string]string)
	mockStore.addErr = nil
	mockStore.getErr = nil
	mockStore.deleteErr = nil
}

func populateMockStore() {
	localStore.Add("hello", "there")
	localStore.Add("random", "value")
	localStore.Add("foo", "bar")
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
			if tt.setupKey != "" {
				_ = mockStore.Add(tt.setupKey, tt.setupValue)
			}
			if tt.injectErr != nil {
				mockStore.getErr = tt.injectErr
			}
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			Get(rr, req)
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
			name:         "Internal Store Error",
			requestPath:  "/Set/errkey/errval",
			expectedCode: http.StatusInternalServerError,
			injectErr:    assert.AnError,
			expectedBody: "Failed to add to store\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()
			mockStore := getMockStore()
			if tt.injectErr != nil {
				mockStore.addErr = tt.injectErr
			}
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			Set(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code, "status code")
			body, _ := io.ReadAll(rr.Body)
			assert.Equal(t, tt.expectedBody, string(body), "response body")
			if tt.expectedKey != "" && tt.expectedValue != "" && tt.injectErr == nil {
				actualValue, err := mockStore.Get(tt.expectedKey)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, actualValue)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		keyToDelete  string
		requestPath  string
		expectedCode int
		injectErr    error
		expectedBody string
	}{
		{
			name:         "Successful delete",
			keyToDelete:  "hello",
			requestPath:  "/Delete/hello",
			expectedCode: http.StatusOK,
			expectedBody: "Deleted key 'hello'\n",
		},
		{
			name:         "Missing Key in URL",
			requestPath:  "/Delete/",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL format. Expected Delete/{key}\n",
		},
		{
			name:         "Internal Store Error",
			keyToDelete:  "errkey",
			requestPath:  "/Delete/errkey",
			expectedCode: http.StatusInternalServerError,
			injectErr:    assert.AnError,
			expectedBody: "Failed to delete key\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()
			mockStore := getMockStore()
			if tt.keyToDelete != "" && tt.injectErr == nil {
				mockStore.data[tt.keyToDelete] = "someval"
			}
			if tt.injectErr != nil {
				mockStore.deleteErr = tt.injectErr
			}
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			Delete(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code, "status code")
			body, _ := io.ReadAll(rr.Body)
			assert.Equal(t, tt.expectedBody, string(body), "response body")
		})
	}
}
