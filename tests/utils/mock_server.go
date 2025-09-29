package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// MockHTTPServer represents a mock HTTP server for testing task execution
type MockHTTPServer struct {
	Server    *httptest.Server
	Requests  []CapturedRequest
	Responses map[string]MockResponse
	mutex     sync.RWMutex
}

// CapturedRequest represents a captured HTTP request
type CapturedRequest struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	Timestamp time.Time         `json:"timestamp"`
}

// MockResponse represents a configured mock response
type MockResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Delay      time.Duration     `json:"delay"`
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer() *MockHTTPServer {
	mock := &MockHTTPServer{
		Requests:  make([]CapturedRequest, 0),
		Responses: make(map[string]MockResponse),
	}

	// Create the test server
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	return mock
}

// handleRequest handles incoming requests to the mock server
func (m *MockHTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Read request body
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body.Close()

	// Capture the request
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	captured := CapturedRequest{
		Method:    r.Method,
		URL:       r.URL.String(),
		Headers:   headers,
		Body:      string(bodyBytes),
		Timestamp: time.Now(),
	}
	m.Requests = append(m.Requests, captured)

	// Generate response key
	responseKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)

	// Check if we have a configured response
	if response, exists := m.Responses[responseKey]; exists {
		// Add delay if configured
		if response.Delay > 0 {
			time.Sleep(response.Delay)
		}

		// Set response headers
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}

		// Set status code and body
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))
		return
	}

	// Default response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Mock server response",
		"timestamp": time.Now().Format(time.RFC3339),
		"method":    r.Method,
		"path":      r.URL.Path,
	})
}

// SetResponse configures a mock response for a specific method and path
func (m *MockHTTPServer) SetResponse(method, path string, response MockResponse) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := fmt.Sprintf("%s:%s", method, path)
	m.Responses[key] = response
}

// GetRequests returns all captured requests
func (m *MockHTTPServer) GetRequests() []CapturedRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to avoid race conditions
	requests := make([]CapturedRequest, len(m.Requests))
	copy(requests, m.Requests)
	return requests
}

// GetRequestCount returns the number of captured requests
func (m *MockHTTPServer) GetRequestCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.Requests)
}

// GetLastRequest returns the last captured request
func (m *MockHTTPServer) GetLastRequest() *CapturedRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.Requests) == 0 {
		return nil
	}
	return &m.Requests[len(m.Requests)-1]
}

// GetRequestsByMethod returns requests filtered by HTTP method
func (m *MockHTTPServer) GetRequestsByMethod(method string) []CapturedRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var filtered []CapturedRequest
	for _, req := range m.Requests {
		if req.Method == method {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// GetRequestsByPath returns requests filtered by URL path
func (m *MockHTTPServer) GetRequestsByPath(path string) []CapturedRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var filtered []CapturedRequest
	for _, req := range m.Requests {
		if req.URL == path {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// ClearRequests clears all captured requests
func (m *MockHTTPServer) ClearRequests() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Requests = make([]CapturedRequest, 0)
}

// ClearResponses clears all configured responses
func (m *MockHTTPServer) ClearResponses() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Responses = make(map[string]MockResponse)
}

// Close shuts down the mock server
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// GetURL returns the base URL of the mock server
func (m *MockHTTPServer) GetURL() string {
	return m.Server.URL
}

// WaitForRequests waits for a specific number of requests with timeout
func (m *MockHTTPServer) WaitForRequests(expectedCount int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if m.GetRequestCount() >= expectedCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// SetSuccessResponse sets a successful JSON response for a path
func (m *MockHTTPServer) SetSuccessResponse(method, path string, body interface{}) {
	bodyBytes, _ := json.Marshal(body)
	m.SetResponse(method, path, MockResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(bodyBytes),
	})
}

// SetErrorResponse sets an error response for a path
func (m *MockHTTPServer) SetErrorResponse(method, path string, statusCode int, message string) {
	errorBody := map[string]string{"error": message}
	bodyBytes, _ := json.Marshal(errorBody)
	m.SetResponse(method, path, MockResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(bodyBytes),
	})
}

// SetTimeoutResponse sets a response that will timeout
func (m *MockHTTPServer) SetTimeoutResponse(method, path string, delay time.Duration) {
	m.SetResponse(method, path, MockResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"message": "delayed response"}`,
		Delay:      delay,
	})
}
