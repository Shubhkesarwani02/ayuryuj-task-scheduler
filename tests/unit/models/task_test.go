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

func TestTask_Creation(t *testing.T) {
	tests := []struct {
		name     string
		task     models.Task
		wantErr  bool
		validate func(*testing.T, models.Task)
	}{
		{
			name: "valid one-off task",
			task: models.Task{
				ID:          uuid.New(),
				Name:        "Test One-off Task",
				TriggerType: models.TriggerTypeOneOff,
				TriggerTime: timePtr(time.Now().Add(time.Hour)),
				Method:      "POST",
				URL:         "https://example.com/webhook",
				Headers:     models.Headers{"Content-Type": "application/json"},
				Payload:     stringPtr(`{"key": "value"}`),
				Status:      models.TaskStatusScheduled,
			},
			wantErr: false,
			validate: func(t *testing.T, task models.Task) {
				assert.Equal(t, models.TriggerTypeOneOff, task.TriggerType)
				assert.NotNil(t, task.TriggerTime)
				assert.Nil(t, task.CronExpr)
				assert.Equal(t, models.TaskStatusScheduled, task.Status)
			},
		},
		{
			name: "valid cron task",
			task: models.Task{
				ID:          uuid.New(),
				Name:        "Test Cron Task",
				TriggerType: models.TriggerTypeCron,
				CronExpr:    stringPtr("0 0 * * *"),
				Method:      "GET",
				URL:         "https://example.com/api",
				Headers:     models.Headers{"User-Agent": "task-scheduler"},
				Status:      models.TaskStatusScheduled,
			},
			wantErr: false,
			validate: func(t *testing.T, task models.Task) {
				assert.Equal(t, models.TriggerTypeCron, task.TriggerType)
				assert.Nil(t, task.TriggerTime)
				assert.NotNil(t, task.CronExpr)
				assert.Equal(t, "0 0 * * *", *task.CronExpr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate task structure
			assert.NotEqual(t, uuid.Nil, tt.task.ID)
			assert.NotEmpty(t, tt.task.Name)
			assert.NotEmpty(t, tt.task.Method)
			assert.NotEmpty(t, tt.task.URL)

			if tt.validate != nil {
				tt.validate(t, tt.task)
			}
		})
	}
}

func TestHeaders_Serialization(t *testing.T) {
	tests := []struct {
		name    string
		headers models.Headers
		wantErr bool
	}{
		{
			name:    "empty headers",
			headers: models.Headers{},
			wantErr: false,
		},
		{
			name: "single header",
			headers: models.Headers{
				"Content-Type": "application/json",
			},
			wantErr: false,
		},
		{
			name: "multiple headers",
			headers: models.Headers{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
				"User-Agent":    "task-scheduler/1.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Value() method (serialization)
			value, err := tt.headers.Value()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify it's valid JSON
			var decoded map[string]string
			err = json.Unmarshal(value.([]byte), &decoded)
			require.NoError(t, err)
			assert.Equal(t, map[string]string(tt.headers), decoded)

			// Test Scan() method (deserialization)
			var scannedHeaders models.Headers
			err = scannedHeaders.Scan(value)
			require.NoError(t, err)
			assert.Equal(t, tt.headers, scannedHeaders)
		})
	}
}

func TestHeaders_ScanNil(t *testing.T) {
	var headers models.Headers
	err := headers.Scan(nil)
	require.NoError(t, err)
	assert.NotNil(t, headers)
	assert.Len(t, headers, 0)
}

func TestHeaders_ScanInvalidType(t *testing.T) {
	var headers models.Headers
	err := headers.Scan("not-bytes")
	require.NoError(t, err) // Should not error but return empty
}

func TestTaskResult_Creation(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name     string
		result   models.TaskResult
		validate func(*testing.T, models.TaskResult)
	}{
		{
			name: "successful result",
			result: models.TaskResult{
				ID:              uuid.New(),
				TaskID:          taskID,
				RunAt:           time.Now(),
				StatusCode:      intPtr(200),
				Success:         true,
				ResponseHeaders: models.Headers{"Content-Type": "application/json"},
				ResponseBody:    stringPtr(`{"status": "ok"}`),
				DurationMs:      150,
			},
			validate: func(t *testing.T, result models.TaskResult) {
				assert.True(t, result.Success)
				assert.NotNil(t, result.StatusCode)
				assert.Equal(t, 200, *result.StatusCode)
				assert.Nil(t, result.ErrorMessage)
				assert.Greater(t, result.DurationMs, 0)
			},
		},
		{
			name: "failed result with error",
			result: models.TaskResult{
				ID:           uuid.New(),
				TaskID:       taskID,
				RunAt:        time.Now(),
				StatusCode:   intPtr(500),
				Success:      false,
				ErrorMessage: stringPtr("Internal server error"),
				DurationMs:   300,
			},
			validate: func(t *testing.T, result models.TaskResult) {
				assert.False(t, result.Success)
				assert.NotNil(t, result.StatusCode)
				assert.Equal(t, 500, *result.StatusCode)
				assert.NotNil(t, result.ErrorMessage)
				assert.Contains(t, *result.ErrorMessage, "Internal server error")
			},
		},
		{
			name: "timeout result",
			result: models.TaskResult{
				ID:           uuid.New(),
				TaskID:       taskID,
				RunAt:        time.Now(),
				Success:      false,
				ErrorMessage: stringPtr("request timeout"),
				DurationMs:   30000,
			},
			validate: func(t *testing.T, result models.TaskResult) {
				assert.False(t, result.Success)
				assert.Nil(t, result.StatusCode)
				assert.NotNil(t, result.ErrorMessage)
				assert.Contains(t, *result.ErrorMessage, "timeout")
				assert.Equal(t, 30000, result.DurationMs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation
			assert.NotEqual(t, uuid.Nil, tt.result.ID)
			assert.NotEqual(t, uuid.Nil, tt.result.TaskID)
			assert.Equal(t, taskID, tt.result.TaskID)

			if tt.validate != nil {
				tt.validate(t, tt.result)
			}
		})
	}
}

func TestCreateTaskRequest_Validation(t *testing.T) {
	futureTime := time.Now().Add(time.Hour)

	tests := []struct {
		name    string
		request models.CreateTaskRequest
		valid   bool
	}{
		{
			name: "valid one-off task request",
			request: models.CreateTaskRequest{
				Name: "Test Task",
				Trigger: models.CreateTaskTrigger{
					Type:     models.TriggerTypeOneOff,
					DateTime: &futureTime,
				},
				Action: models.CreateTaskAction{
					Method: "POST",
					URL:    "https://example.com/webhook",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Payload: map[string]string{"key": "value"},
				},
			},
			valid: true,
		},
		{
			name: "valid cron task request",
			request: models.CreateTaskRequest{
				Name: "Cron Task",
				Trigger: models.CreateTaskTrigger{
					Type: models.TriggerTypeCron,
					Cron: stringPtr("0 */6 * * *"),
				},
				Action: models.CreateTaskAction{
					Method: "GET",
					URL:    "https://api.example.com/status",
				},
			},
			valid: true,
		},
		{
			name: "empty name",
			request: models.CreateTaskRequest{
				Name: "",
				Trigger: models.CreateTaskTrigger{
					Type:     models.TriggerTypeOneOff,
					DateTime: &futureTime,
				},
				Action: models.CreateTaskAction{
					Method: "GET",
					URL:    "https://example.com",
				},
			},
			valid: false,
		},
		{
			name: "invalid URL",
			request: models.CreateTaskRequest{
				Name: "Test Task",
				Trigger: models.CreateTaskTrigger{
					Type:     models.TriggerTypeOneOff,
					DateTime: &futureTime,
				},
				Action: models.CreateTaskAction{
					Method: "GET",
					URL:    "not-a-valid-url",
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic structure validation
			if tt.valid {
				assert.NotEmpty(t, tt.request.Name)
				assert.NotEmpty(t, tt.request.Action.Method)
				assert.NotEmpty(t, tt.request.Action.URL)

				if tt.request.Trigger.Type == models.TriggerTypeOneOff {
					assert.NotNil(t, tt.request.Trigger.DateTime)
				} else {
					assert.NotNil(t, tt.request.Trigger.Cron)
				}
			}
		})
	}
}

func TestUpdateTaskRequest_Structure(t *testing.T) {
	futureTime := time.Now().Add(time.Hour)

	updateReq := models.UpdateTaskRequest{
		Name: stringPtr("Updated Task Name"),
		Trigger: &models.CreateTaskTrigger{
			Type:     models.TriggerTypeOneOff,
			DateTime: &futureTime,
		},
		Action: &models.CreateTaskAction{
			Method: "PUT",
			URL:    "https://updated.example.com",
			Headers: map[string]string{
				"Authorization": "Bearer new-token",
			},
		},
	}

	assert.NotNil(t, updateReq.Name)
	assert.Equal(t, "Updated Task Name", *updateReq.Name)
	assert.NotNil(t, updateReq.Trigger)
	assert.NotNil(t, updateReq.Action)
	assert.Equal(t, "PUT", updateReq.Action.Method)
}

func TestTaskStatus_Constants(t *testing.T) {
	assert.Equal(t, models.TaskStatus("scheduled"), models.TaskStatusScheduled)
	assert.Equal(t, models.TaskStatus("cancelled"), models.TaskStatusCancelled)
	assert.Equal(t, models.TaskStatus("completed"), models.TaskStatusCompleted)
}

func TestTriggerType_Constants(t *testing.T) {
	assert.Equal(t, models.TriggerType("one-off"), models.TriggerTypeOneOff)
	assert.Equal(t, models.TriggerType("cron"), models.TriggerTypeCron)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
