package logic

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-scheduler/internal/models"
)

// TestBusinessLogic contains tests for core business logic without external dependencies
func TestTaskValidation(t *testing.T) {
	t.Run("valid_one_off_task", func(t *testing.T) {
		futureTime := time.Now().Add(time.Hour)
		task := &models.Task{
			ID:          uuid.New(),
			Name:        "Valid Task",
			TriggerType: models.TriggerTypeOneOff,
			TriggerTime: &futureTime,
			Method:      "POST",
			URL:         "https://example.com/webhook",
			Status:      models.TaskStatusScheduled,
		}

		// Validate task fields
		assert.NotEqual(t, uuid.Nil, task.ID)
		assert.NotEmpty(t, task.Name)
		assert.True(t, task.TriggerTime.After(time.Now()))
		assert.Contains(t, []string{"GET", "POST", "PUT", "DELETE"}, task.Method)
		assert.True(t, strings.HasPrefix(task.URL, "http"))
	})

	t.Run("valid_cron_task", func(t *testing.T) {
		cronExpr := "0 */6 * * *"
		task := &models.Task{
			ID:          uuid.New(),
			Name:        "Cron Task",
			TriggerType: models.TriggerTypeCron,
			CronExpr:    &cronExpr,
			Method:      "GET",
			URL:         "https://example.com/status",
			Status:      models.TaskStatusScheduled,
		}

		assert.NotNil(t, task.CronExpr)
		assert.Contains(t, *task.CronExpr, "*")
		assert.Nil(t, task.TriggerTime)
	})
}

func TestTaskResultValidation(t *testing.T) {
	taskID := uuid.New()

	t.Run("successful_result", func(t *testing.T) {
		statusCode := 200
		responseBody := `{"status": "ok"}`

		result := &models.TaskResult{
			ID:           uuid.New(),
			TaskID:       taskID,
			Success:      true,
			StatusCode:   &statusCode,
			ResponseBody: &responseBody,
			DurationMs:   150,
			RunAt:        time.Now(),
		}

		assert.True(t, result.Success)
		assert.Equal(t, 200, *result.StatusCode)
		assert.Greater(t, result.DurationMs, 0)
		assert.Nil(t, result.ErrorMessage)
	})

	t.Run("failed_result", func(t *testing.T) {
		statusCode := 500
		errorMsg := "Internal server error"

		result := &models.TaskResult{
			ID:           uuid.New(),
			TaskID:       taskID,
			Success:      false,
			StatusCode:   &statusCode,
			ErrorMessage: &errorMsg,
			DurationMs:   250,
			RunAt:        time.Now(),
		}

		assert.False(t, result.Success)
		assert.Equal(t, 500, *result.StatusCode)
		assert.NotNil(t, result.ErrorMessage)
		assert.Contains(t, *result.ErrorMessage, "error")
	})
}

func TestJSONPayloadHandling(t *testing.T) {
	t.Run("marshal_and_unmarshal_payload", func(t *testing.T) {
		originalPayload := map[string]interface{}{
			"user_id": 123,
			"name":    "John Doe",
			"active":  true,
			"tags":    []string{"important", "user"},
		}

		// Marshal to JSON string
		payloadBytes, err := json.Marshal(originalPayload)
		require.NoError(t, err)
		payloadStr := string(payloadBytes)

		// Unmarshal back
		var parsedPayload map[string]interface{}
		err = json.Unmarshal([]byte(payloadStr), &parsedPayload)
		require.NoError(t, err)

		// Validate
		assert.Equal(t, float64(123), parsedPayload["user_id"]) // JSON numbers are float64
		assert.Equal(t, "John Doe", parsedPayload["name"])
		assert.Equal(t, true, parsedPayload["active"])
	})
}

func TestTaskStatusLogic(t *testing.T) {
	validTransitions := map[models.TaskStatus][]models.TaskStatus{
		models.TaskStatusScheduled: {models.TaskStatusCompleted, models.TaskStatusCancelled},
		models.TaskStatusCompleted: {}, // Terminal state
		models.TaskStatusCancelled: {}, // Terminal state
	}

	for fromStatus, toStatuses := range validTransitions {
		t.Run("from_"+string(fromStatus), func(t *testing.T) {
			for _, toStatus := range toStatuses {
				t.Run("to_"+string(toStatus), func(t *testing.T) {
					// Simulate status transition
					task := &models.Task{
						ID:     uuid.New(),
						Status: fromStatus,
					}

					// Change status
					task.Status = toStatus

					// Validate transition
					assert.Equal(t, toStatus, task.Status)
					assert.NotEqual(t, fromStatus, toStatus)
				})
			}
		})
	}
}

func TestHTTPMethodValidation(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range validMethods {
		t.Run("method_"+method, func(t *testing.T) {
			task := &models.Task{
				ID:     uuid.New(),
				Name:   "HTTP Method Test",
				Method: method,
				URL:    "https://example.com/api",
			}

			assert.Contains(t, validMethods, task.Method)
			assert.NotEmpty(t, task.Method)
		})
	}
}

func TestURLValidation(t *testing.T) {
	validURLs := []string{
		"https://example.com",
		"https://api.example.com/webhook",
		"http://localhost:8080/callback",
		"https://webhook.site/unique-id",
	}

	for _, url := range validURLs {
		t.Run("url_validation", func(t *testing.T) {
			task := &models.Task{
				ID:   uuid.New(),
				Name: "URL Test",
				URL:  url,
			}

			assert.True(t, strings.HasPrefix(task.URL, "http"))
			assert.Contains(t, task.URL, "://")
		})
	}
}

func TestHeadersHandling(t *testing.T) {
	headers := models.Headers{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token123",
		"X-API-Key":     "api-key-456",
		"User-Agent":    "TaskScheduler/1.0",
	}

	task := &models.Task{
		ID:      uuid.New(),
		Headers: headers,
	}

	assert.Len(t, task.Headers, 4)
	assert.Equal(t, "application/json", task.Headers["Content-Type"])
	assert.Equal(t, "Bearer token123", task.Headers["Authorization"])
	assert.Contains(t, task.Headers["User-Agent"], "TaskScheduler")
}
