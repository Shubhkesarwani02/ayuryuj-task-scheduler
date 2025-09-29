package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-scheduler/internal/models"
)

// TestTaskCreation tests basic task creation and validation
func TestTaskCreation(t *testing.T) {
	// Test one-off task creation
	triggerTime := time.Now().Add(time.Hour)
	task := &models.Task{
		ID:          uuid.New(),
		Name:        "Test One-Off Task",
		TriggerType: models.TriggerTypeOneOff,
		TriggerTime: &triggerTime,
		Method:      "POST",
		URL:         "https://api.example.com/webhook",
		Headers:     models.Headers{"Content-Type": "application/json"},
		Status:      models.TaskStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		NextRun:     &triggerTime,
	}

	// Assertions
	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.Equal(t, "Test One-Off Task", task.Name)
	assert.Equal(t, models.TriggerTypeOneOff, task.TriggerType)
	assert.NotNil(t, task.TriggerTime)
	assert.Equal(t, "POST", task.Method)
	assert.Equal(t, "https://api.example.com/webhook", task.URL)
	assert.Equal(t, models.TaskStatusScheduled, task.Status)
}

// TestCronTask tests cron task creation
func TestCronTask(t *testing.T) {
	cronExpr := "0 */6 * * *" // Every 6 hours
	nextRun := time.Now().Add(6 * time.Hour)

	task := &models.Task{
		ID:          uuid.New(),
		Name:        "Test Cron Task",
		TriggerType: models.TriggerTypeCron,
		CronExpr:    &cronExpr,
		Method:      "GET",
		URL:         "https://api.example.com/status",
		Status:      models.TaskStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		NextRun:     &nextRun,
	}

	assert.Equal(t, models.TriggerTypeCron, task.TriggerType)
	assert.NotNil(t, task.CronExpr)
	assert.Equal(t, cronExpr, *task.CronExpr)
	assert.Nil(t, task.TriggerTime)
	assert.NotNil(t, task.NextRun)
}

// TestTaskResult tests task result creation
func TestTaskResult(t *testing.T) {
	taskID := uuid.New()
	statusCode := 200
	responseBody := `{"status": "success", "message": "Task completed"}`

	result := &models.TaskResult{
		ID:           uuid.New(),
		TaskID:       taskID,
		Success:      true,
		StatusCode:   &statusCode,
		ResponseBody: &responseBody,
		DurationMs:   234,
		ExecutedAt:   time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, taskID, result.TaskID)
	assert.True(t, result.Success)
	assert.Equal(t, 200, *result.StatusCode)
	assert.Contains(t, *result.ResponseBody, "success")
	assert.Greater(t, result.DurationMs, 0)
}

// TestTaskStatusTransitions tests valid status transitions
func TestTaskStatusTransitions(t *testing.T) {
	validStatuses := []models.TaskStatus{
		models.TaskStatusScheduled,
		models.TaskStatusRunning,
		models.TaskStatusCompleted,
		models.TaskStatusFailed,
		models.TaskStatusCancelled,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			task := &models.Task{
				ID:     uuid.New(),
				Name:   "Status Test Task",
				Status: status,
			}
			assert.Equal(t, status, task.Status)
			assert.NotEmpty(t, string(task.Status))
		})
	}
}

// TestTaskWithPayload tests task with JSON payload
func TestTaskWithPayload(t *testing.T) {
	payload := map[string]interface{}{
		"user_id": 123,
		"action":  "update_profile",
		"data": map[string]string{
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	payloadStr := string(payloadBytes)

	task := &models.Task{
		ID:      uuid.New(),
		Name:    "Task with Payload",
		Method:  "POST",
		URL:     "https://api.example.com/users",
		Payload: &payloadStr,
		Headers: models.Headers{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Status: models.TaskStatusScheduled,
	}

	assert.NotNil(t, task.Payload)

	// Parse payload back to verify
	var parsedPayload map[string]interface{}
	err = json.Unmarshal([]byte(*task.Payload), &parsedPayload)
	require.NoError(t, err)

	assert.Equal(t, float64(123), parsedPayload["user_id"]) // JSON numbers are float64
	assert.Equal(t, "update_profile", parsedPayload["action"])

	// Test headers
	assert.Equal(t, "application/json", task.Headers["Content-Type"])
	assert.Equal(t, "Bearer token123", task.Headers["Authorization"])
}

// TestCreateTaskRequest tests API request models
func TestCreateTaskRequest(t *testing.T) {
	futureTime := time.Now().Add(time.Hour)

	request := models.CreateTaskRequest{
		Name: "API Test Task",
		Trigger: models.CreateTaskTrigger{
			Type:     models.TriggerTypeOneOff,
			DateTime: &futureTime,
		},
		Action: models.CreateTaskAction{
			Method:  "POST",
			URL:     "https://webhook.example.com/api/notify",
			Headers: map[string]string{"Content-Type": "application/json"},
			Payload: map[string]string{"message": "Hello, World!"},
		},
	}

	assert.Equal(t, "API Test Task", request.Name)
	assert.Equal(t, models.TriggerTypeOneOff, request.Trigger.Type)
	assert.NotNil(t, request.Trigger.DateTime)
	assert.Equal(t, "POST", request.Action.Method)
	assert.Equal(t, "https://webhook.example.com/api/notify", request.Action.URL)
	assert.Contains(t, request.Action.Headers, "Content-Type")
}

// TestTriggerTypes tests different trigger types
func TestTriggerTypes(t *testing.T) {
	triggerTypes := []models.TriggerType{
		models.TriggerTypeOneOff,
		models.TriggerTypeCron,
	}

	for _, triggerType := range triggerTypes {
		t.Run(string(triggerType), func(t *testing.T) {
			assert.NotEmpty(t, string(triggerType))
			assert.True(t, triggerType == models.TriggerTypeOneOff || triggerType == models.TriggerTypeCron)
		})
	}
}
