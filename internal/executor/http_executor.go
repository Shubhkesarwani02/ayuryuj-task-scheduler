package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"task-scheduler/internal/models"
)

type HTTPExecutor struct {
	client *http.Client
}

func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (e *HTTPExecutor) Execute(task *models.Task) *models.TaskResult {
	startTime := time.Now()

	result := &models.TaskResult{
		ID:              uuid.New(),
		TaskID:          task.ID,
		RunAt:           startTime,
		Success:         false,
		ResponseHeaders: make(models.Headers),
		CreatedAt:       time.Now(),
	}

	// Prepare request
	req, err := e.prepareRequest(task)
	if err != nil {
		result.ErrorMessage = stringPtr(fmt.Sprintf("Failed to prepare request: %v", err))
		result.DurationMs = int(time.Since(startTime).Milliseconds())
		return result
	}

	// Execute request
	resp, err := e.client.Do(req)
	if err != nil {
		result.ErrorMessage = stringPtr(fmt.Sprintf("HTTP request failed: %v", err))
		result.DurationMs = int(time.Since(startTime).Milliseconds())
		return result
	}
	defer resp.Body.Close()

	// Calculate duration
	result.DurationMs = int(time.Since(startTime).Milliseconds())

	// Set status code
	result.StatusCode = &resp.StatusCode

	// Determine success (2xx status codes are considered successful)
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	// Extract response headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.ResponseHeaders[key] = values[0] // Take first value if multiple
		}
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.ErrorMessage = stringPtr(fmt.Sprintf("Failed to read response body: %v", err))
		return result
	}

	// Store response body (limit size to prevent database issues)
	bodyStr := string(body)
	if len(bodyStr) > 10000 { // Limit to 10KB
		bodyStr = bodyStr[:10000] + "... (truncated)"
	}
	result.ResponseBody = &bodyStr

	// If request failed, add status code to error message
	if !result.Success {
		if result.ErrorMessage == nil {
			result.ErrorMessage = stringPtr(fmt.Sprintf("HTTP request returned status %d", resp.StatusCode))
		} else {
			*result.ErrorMessage = fmt.Sprintf("%s (status: %d)", *result.ErrorMessage, resp.StatusCode)
		}
	}

	return result
}

func (e *HTTPExecutor) prepareRequest(task *models.Task) (*http.Request, error) {
	var body io.Reader

	// Prepare request body if payload exists
	if task.Payload != nil && *task.Payload != "" {
		body = strings.NewReader(*task.Payload)
	}

	// Create request
	req, err := http.NewRequest(strings.ToUpper(task.Method), task.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if task.Headers != nil {
		for key, value := range task.Headers {
			req.Header.Set(key, value)
		}
	}

	// Set default Content-Type for requests with payload
	if task.Payload != nil && *task.Payload != "" && req.Header.Get("Content-Type") == "" {
		// Try to detect if payload is JSON
		var js json.RawMessage
		if json.Unmarshal([]byte(*task.Payload), &js) == nil {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// Set User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "TaskScheduler/1.0")
	}

	return req, nil
}

func (e *HTTPExecutor) ExecuteWithTimeout(task *models.Task, timeout time.Duration) *models.TaskResult {
	// Create a new client with custom timeout
	client := &http.Client{
		Timeout:   timeout,
		Transport: e.client.Transport,
	}

	originalClient := e.client
	e.client = client

	result := e.Execute(task)

	// Restore original client
	e.client = originalClient

	return result
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// ExecutorInterface for dependency injection
type ExecutorInterface interface {
	Execute(task *models.Task) *models.TaskResult
	ExecuteWithTimeout(task *models.Task, timeout time.Duration) *models.TaskResult
}
