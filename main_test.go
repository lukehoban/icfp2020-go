package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInteractEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectedError  string
		expectedResult interface{}
	}{
		{
			name:           "valid addition expression",
			method:         "POST",
			body:           InteractRequest{Expression: "ap ap add 2 3"},
			expectedStatus: 200,
			expectedResult: int64(5),
		},
		{
			name:           "valid multiplication expression",
			method:         "POST",
			body:           InteractRequest{Expression: "ap ap mul 4 5"},
			expectedStatus: 200,
			expectedResult: int64(20),
		},
		{
			name:           "valid number expression",
			method:         "POST",
			body:           InteractRequest{Expression: "42"},
			expectedStatus: 200,
			expectedResult: int64(42),
		},
		{
			name:           "partial application expression",
			method:         "POST",
			body:           InteractRequest{Expression: "ap add 5"},
			expectedStatus: 200,
			expectedResult: "ap add 5", // Partial applications return their string representation
		},
		{
			name:           "invalid method GET",
			method:         "GET",
			body:           nil,
			expectedStatus: 405,
			expectedError:  "Method not allowed",
		},
		{
			name:           "invalid JSON body",
			method:         "POST",
			body:           "invalid json",
			expectedStatus: 400,
			expectedError:  "Invalid JSON",
		},
		{
			name:           "empty expression",
			method:         "POST",
			body:           InteractRequest{Expression: ""},
			expectedStatus: 400,
			expectedError:  "Invalid expression",
		},
		{
			name:           "invalid expression",
			method:         "POST",
			body:           InteractRequest{Expression: "invalid syntax"},
			expectedStatus: 400,
			expectedError:  "Invalid expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.body == nil {
				req = httptest.NewRequest(tt.method, "/interact", nil)
			} else if bodyStr, ok := tt.body.(string); ok {
				// For invalid JSON test
				req = httptest.NewRequest(tt.method, "/interact", bytes.NewBufferString(bodyStr))
			} else {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, "/interact", bytes.NewBuffer(bodyBytes))
			}

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			interactHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response InteractResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectedError != "" {
				if response.Error != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, response.Error)
				}
			} else {
				if response.Error != "" {
					t.Errorf("Unexpected error: %s", response.Error)
				} else {
					// Handle JSON number conversion - JSON unmarshals numbers as float64 by default
					if expectedNum, ok := tt.expectedResult.(int64); ok {
						if actualNum, ok := response.Result.(float64); ok {
							if int64(actualNum) != expectedNum {
								t.Errorf("Expected result %v, got %v", expectedNum, int64(actualNum))
							}
						} else {
							t.Errorf("Expected result %v (int64), got %v (%T)", tt.expectedResult, response.Result, response.Result)
						}
					} else {
						if response.Result != tt.expectedResult {
							t.Errorf("Expected result %v, got %v", tt.expectedResult, response.Result)
						}
					}
				}
			}
		})
	}
}

func TestRootEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		contentType    string
	}{
		{
			name:           "valid GET request",
			method:         "GET",
			expectedStatus: 200,
			contentType:    "text/html",
		},
		{
			name:           "invalid POST request",
			method:         "POST",
			expectedStatus: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			rr := httptest.NewRecorder()

			rootHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.contentType != "" {
				contentType := rr.Header().Get("Content-Type")
				if contentType != tt.contentType {
					t.Errorf("Expected content type %s, got %s", tt.contentType, contentType)
				}
			}

			if tt.method == "GET" && tt.expectedStatus == 200 {
				body := rr.Body.String()
				if !bytes.Contains([]byte(body), []byte("Galaxy Interpreter")) {
					t.Error("Expected HTML to contain 'Galaxy Interpreter' title")
				}
				if !bytes.Contains([]byte(body), []byte("<html>")) {
					t.Error("Expected valid HTML document")
				}
			}
		})
	}
}

// Test helper to ensure the global program is loaded
func TestGlobalProgramLoaded(t *testing.T) {
	if globalProgram == nil {
		t.Error("globalProgram should be loaded during init")
	}
	if len(globalProgram) == 0 {
		t.Error("globalProgram should contain parsed symbols")
	}
}
