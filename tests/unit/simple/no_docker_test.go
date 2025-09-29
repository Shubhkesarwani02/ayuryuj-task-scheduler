package simple

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-scheduler/internal/executor"
	"task-scheduler/internal/models"
	"task-scheduler/tests/utils"
)

// TestHTTPExecutorBasic tests the HTTP executor with mock server (no database required)
func TestHTTPExecutorBasic(t *testing.T) {
	// Create HTTP executor
	httpExecutor := executor.NewHTTPExecutor()

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

// TestTaskFactoryBasic tests task creation without database
func TestTaskFactoryBasic(t *testing.T) {
	factory := utils.NewTaskFactory()

	// Test one-off task creation
	futureTime := time.Now().Add(time.Hour)
	task := factory.CreateOneOffTask("Test Task", "https://example.com/webhook", futureTime)

	assert.Equal(t, "Test Task", task.Name)
	assert.Equal(t, models.TriggerTypeOneOff, task.TriggerType)
	assert.Equal(t, "https://example.com/webhook", task.URL)
	assert.NotNil(t, task.TriggerTime)
	assert.Equal(t, models.TaskStatusScheduled, task.Status)

	// Test cron task creation
	cronTask := factory.CreateCronTask("Cron Task", "https://example.com/api", "0 */6 * * *")

	assert.Equal(t, "Cron Task", cronTask.Name)
	assert.Equal(t, models.TriggerTypeCron, cronTask.TriggerType)
	assert.NotNil(t, cronTask.CronExpr)
	assert.Equal(t, "0 */6 * * *", *cronTask.CronExpr)
	assert.Nil(t, cronTask.TriggerTime)
}

// TestMockServerBasic tests mock server functionality
func TestMockServerBasic(t *testing.T) {
	mockServer := utils.NewMockHTTPServer()
	defer mockServer.Close()

	// Configure responses
	mockServer.SetSuccessResponse("GET", "/api/status", map[string]string{"status": "ok"})
	mockServer.SetErrorResponse("POST", "/api/error", 500, "Internal Server Error")

	// Test URL generation
	url := mockServer.GetURL()
	assert.Contains(t, url, "http://")

	// Clear requests/responses
	mockServer.ClearRequests()
	mockServer.ClearResponses()

	// Verify cleared
	assert.Equal(t, 0, mockServer.GetRequestCount())
}

// TestResultFactoryBasic tests result creation without database
func TestResultFactoryBasic(t *testing.T) {
	factory := utils.NewResultFactory()
	taskID := utils.NewTaskFactory().CreateOneOffTask("Test", "https://example.com", time.Now().Add(time.Hour)).ID

	// Test success result
	successResult := factory.CreateSuccessResult(taskID)

	assert.Equal(t, taskID, successResult.TaskID)
	assert.True(t, successResult.Success)
	assert.Equal(t, 200, *successResult.StatusCode)
	assert.Greater(t, successResult.DurationMs, 0)
	assert.Nil(t, successResult.ErrorMessage)

	// Test error result
	errorResult := factory.CreateErrorResult(taskID, 500, "Connection timeout")

	assert.Equal(t, taskID, errorResult.TaskID)
	assert.False(t, errorResult.Success)
	assert.Equal(t, 500, *errorResult.StatusCode)
	assert.NotNil(t, errorResult.ErrorMessage)
	assert.Contains(t, *errorResult.ErrorMessage, "timeout")

	// Test timeout result
	timeoutResult := factory.CreateTimeoutResult(taskID)

	assert.Equal(t, taskID, timeoutResult.TaskID)
	assert.False(t, timeoutResult.Success)
	assert.NotNil(t, timeoutResult.ErrorMessage)
	assert.Contains(t, *timeoutResult.ErrorMessage, "timeout")
	assert.Equal(t, 30000, timeoutResult.DurationMs)
}
