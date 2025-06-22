package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexG28/keyvalue/store"
)

type MockStore struct {
	data map[string]string
}

func NewMockStore() *MockStore {
	return &MockStore{
		data: make(map[string]string),
	}
}

func (m *MockStore) Add(key, value string) error {
	m.data[key] = value
	return nil
}

func (m *MockStore) Get(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", store.ErrNotFound
}

func (m *MockStore) Delete(key string) error {
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
}

func TestGet(t *testing.T) {
	// Use table-driven tests for multiple scenarios
	tests := []struct {
		name         string
		setupKey     string
		setupValue   string
		requestPath  string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Successful Get",
			setupKey:     "hello",
			setupValue:   "there",
			requestPath:  "/Get/hello",
			expectedCode: http.StatusOK,
			expectedBody: "there", // This is the expected value from your mock store
		},
		{
			name:         "Key Not Found",
			setupKey:     "", // No key set
			setupValue:   "",
			requestPath:  "/Get/nonexistent",
			expectedCode: http.StatusNotFound,
			expectedBody: "Failed to get from store\n",
		},
		{
			name:         "Missing Key in URL",
			setupKey:     "",
			setupValue:   "",
			requestPath:  "/Get/",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid URL format. Expected Get/{key}\n",
		},
		{
			name:         "Incorrect Body Check",
			setupKey:     "another",
			setupValue:   "correct value",
			requestPath:  "/Get/another",
			expectedCode: http.StatusOK,
			expectedBody: "correct value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()
			if tt.setupKey != "" {
				err := localStore.Add(tt.setupKey, tt.setupValue)
				if err != nil {
					t.Fatalf("Failed to add setup data: %v", err)
				}
			}
			req, err := http.NewRequest(http.MethodGet, tt.requestPath, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			rr := httptest.NewRecorder()

			Get(rr, req)

			// Assert
			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %d want %d",
					rr.Code, tt.expectedCode)
			}

			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("Could not read response body: %v", err)
			}

			if string(body) != tt.expectedBody {
				t.Errorf("handler returned unexpected body:\nGOT: %q\nWANT: %q",
					string(body), tt.expectedBody)
			}
		})
	}
}
