package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-scheduler/internal/executor"
	"task-scheduler/internal/models"
	"task-scheduler/tests/utils"
)

func TestHTTPExecutorWithMockServer(t *testing.T) {
	// Create HTTP executor
	httpExecutor := executor.NewHTTPExecutor()
	require.NotNil(t, httpExecutor)

	// Create mock server
	mockServer := utils.NewMockHTTPServer()
	defer mockServer.Close()

	// Setup success response
	mockServer.SetSuccessResponse("POST", "/webhook", map[string]interface{}{
		"status": "received",
		"id":     "12345",
	})

	// Create task factory
	factory := utils.NewTaskFactory()

	// Create test task
	task := factory.CreateHTTPTask(
		"POST",
		mockServer.GetURL()+"/webhook",
		models.Headers{"Content-Type": "application/json"},
		map[string]string{"test": "data"},
	)

	// Execute task
	result := httpExecutor.Execute(task)

	// Verify result
	require.NotNil(t, result)
	assert.Equal(t, task.ID, result.TaskID)
	assert.True(t, result.Success)
	assert.Equal(t, 200, *result.StatusCode)
	assert.Greater(t, result.DurationMs, 0)

	// Verify mock server received request
	assert.True(t, mockServer.WaitForRequests(1, 2*time.Second))
	requests := mockServer.GetRequests()
	require.Len(t, requests, 1)

	req := requests[0]
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/webhook", req.URL)
	assert.Contains(t, req.Body, "test")
}

func TestHTTPExecutorGETRequest(t *testing.T) {
	// Create HTTP executor
	httpExecutor := executor.NewHTTPExecutor()

	// Create mock server
	mockServer := utils.NewMockHTTPServer()
	defer mockServer.Close()

	// Setup success response
	mockServer.SetSuccessResponse("GET", "/api/status", map[string]interface{}{
		"status": "healthy",
		"uptime": "5m30s",
	})

	// Create task factory
	factory := utils.NewTaskFactory()

	// Create GET task
	task := factory.CreateHTTPTask(
		"GET",
		mockServer.GetURL()+"/api/status",
		models.Headers{"User-Agent": "task-scheduler/1.0"},
		nil, // No payload for GET
	)

	// Execute task
	result := httpExecutor.Execute(task)

	// Verify result
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 200, *result.StatusCode)

	// Verify request received
	assert.True(t, mockServer.WaitForRequests(1, 2*time.Second))
	requests := mockServer.GetRequests()
	require.Len(t, requests, 1)

	req := requests[0]
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "/api/status", req.URL)
}

func TestHTTPExecutorErrorResponse(t *testing.T) {
	// Create HTTP executor
	httpExecutor := executor.NewHTTPExecutor()

	// Create mock server
	mockServer := utils.NewMockHTTPServer()
	defer mockServer.Close()

	// Setup error response
	mockServer.SetErrorResponse("POST", "/api/error", 500, "Internal Server Error")

	// Create task factory
	factory := utils.NewTaskFactory()

	// Create task that will fail
	task := factory.CreateHTTPTask(
		"POST",
		mockServer.GetURL()+"/api/error",
		models.Headers{"Content-Type": "application/json"},
		map[string]string{"data": "test"},
	)

	// Execute task
	result := httpExecutor.Execute(task)

	// Verify result shows failure
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, 500, *result.StatusCode)
	assert.NotNil(t, result.ErrorMessage)
}
