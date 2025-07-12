package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvalEndpoint(t *testing.T) {
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
			body:           EvalRequest{Expression: "ap ap add 2 3"},
			expectedStatus: 200,
			expectedResult: int64(5),
		},
		{
			name:           "valid multiplication expression",
			method:         "POST",
			body:           EvalRequest{Expression: "ap ap mul 4 5"},
			expectedStatus: 200,
			expectedResult: int64(20),
		},
		{
			name:           "valid number expression",
			method:         "POST",
			body:           EvalRequest{Expression: "42"},
			expectedStatus: 200,
			expectedResult: int64(42),
		},
		{
			name:           "partial application expression",
			method:         "POST",
			body:           EvalRequest{Expression: "ap add 5"},
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
			body:           EvalRequest{Expression: ""},
			expectedStatus: 400,
			expectedError:  "Invalid expression",
		},
		{
			name:           "invalid expression",
			method:         "POST",
			body:           EvalRequest{Expression: "invalid syntax"},
			expectedStatus: 400,
			expectedError:  "Invalid expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.body == nil {
				req = httptest.NewRequest(tt.method, "/eval", nil)
			} else if bodyStr, ok := tt.body.(string); ok {
				// For invalid JSON test
				req = httptest.NewRequest(tt.method, "/eval", bytes.NewBufferString(bodyStr))
			} else {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, "/eval", bytes.NewBuffer(bodyBytes))
			}

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			evalHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response EvalResponse
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

func TestInteractEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "valid interact request",
			method: "POST",
			body: InteractRequest{
				State: "nil",
				Point: struct {
					X int `json:"x"`
					Y int `json:"y"`
				}{X: 0, Y: 0},
			},
			expectedStatus: 200,
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
			name:   "empty state",
			method: "POST",
			body: InteractRequest{
				State: "",
				Point: struct {
					X int `json:"x"`
					Y int `json:"y"`
				}{X: 0, Y: 0},
			},
			expectedStatus: 400,
			expectedError:  "Invalid state",
		},
		{
			name:   "invalid state expression",
			method: "POST",
			body: InteractRequest{
				State: "invalid syntax",
				Point: struct {
					X int `json:"x"`
					Y int `json:"y"`
				}{X: 0, Y: 0},
			},
			expectedStatus: 400,
			expectedError:  "Invalid state expression",
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
				}
				// For successful requests, check that we get valid response structure
				if response.NewState == "" && response.Images == nil {
					t.Error("Expected non-empty response for successful interaction")
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
func TestGalaxyLoaded(t *testing.T) {
	if galaxy == nil {
		t.Error("galaxy should be loaded during init")
	}
	if len(galaxy) == 0 {
		t.Error("galaxy should contain parsed symbols")
	}
}
