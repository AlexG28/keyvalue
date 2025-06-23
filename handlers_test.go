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
			setupKey:     "",
			setupValue:   "",
			requestPath:  "/Get/nonexistent",
			expectedCode: http.StatusNotFound,
			expectedBody: "Key not found\n",
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

func TestSet(t *testing.T) {
	tests := []struct {
		name          string
		expectedKey   string
		expectedValue string
		requestPath   string
		expectedCode  int
	}{
		{
			name:          "Successful set",
			expectedKey:   "hello",
			expectedValue: "there",
			requestPath:   "/Set/hello/there",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "Missing Key in URL",
			expectedKey:   "",
			expectedValue: "",
			requestPath:   "/Set//randomvalue",
			expectedCode:  http.StatusBadRequest,
		},
		{
			name:          "Missing Value in URL",
			expectedKey:   "",
			expectedValue: "",
			requestPath:   "/Get/randomkey/",
			expectedCode:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()

			req, err := http.NewRequest(http.MethodGet, tt.requestPath, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			rr := httptest.NewRecorder()

			Set(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %d want %d",
					rr.Code, tt.expectedCode)
			}
			if tt.expectedKey != "" && tt.expectedValue != "" {

				actualValue, err := localStore.Get(tt.expectedKey)

				if err != nil {
					t.Errorf("key mising from map: %q", tt.expectedKey)
				}

				if actualValue != tt.expectedValue {
					t.Errorf("handler returned unexpected value:\nGOT: %q\nWANT: %q",
						actualValue, tt.expectedValue)
				}
			} else {
				_, err := localStore.Get(tt.expectedKey)
				if err == nil {
					t.Errorf("Key somehow in the map: %q", tt.expectedKey)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name          string
		keyToDelete   string
		expectedValue string
		requestPath   string
		expectedCode  int
	}{
		{
			name:          "Successful delete",
			keyToDelete:   "hello",
			expectedValue: "",
			requestPath:   "/Delete/hello",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "Missing Key in URL",
			keyToDelete:   "",
			expectedValue: "",
			requestPath:   "/Delete/",
			expectedCode:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMockStore()
			populateMockStore()

			req, err := http.NewRequest(http.MethodGet, tt.requestPath, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			rr := httptest.NewRecorder()

			Delete(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %d want %d",
					rr.Code, tt.expectedCode)
			}
			if tt.keyToDelete != "" && tt.expectedValue != "" {

				actualValue, err := localStore.Get(tt.keyToDelete)

				if err != nil {
					t.Errorf("key mising from map: %q", tt.keyToDelete)
				}

				if actualValue != tt.expectedValue {
					t.Errorf("handler returned unexpected value:\nGOT: %q\nWANT: %q",
						actualValue, tt.expectedValue)
				}
			} else {
				_, err := localStore.Get(tt.keyToDelete)
				if err == nil {
					t.Errorf("Key somehow in the map: %q", tt.keyToDelete)
				}
			}
		})
	}
}
