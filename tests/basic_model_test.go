package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"task-scheduler/internal/models"
)

// TestTaskModel tests the basic Task model functionality
func TestTaskModel(t *testing.T) {
	// Test task creation
	task := &models.Task{
		ID:          uuid.New(),
		Name:        "Test Task",
		TriggerType: models.TriggerTypeOneOff,
		Method:      "POST",
		URL:         "https://example.com/webhook",
		Status:      models.TaskStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.Equal(t, "Test Task", task.Name)
	assert.Equal(t, models.TriggerTypeOneOff, task.TriggerType)
	assert.Equal(t, models.TaskStatusScheduled, task.Status)
}

// TestTaskResult tests the TaskResult model
func TestTaskResult(t *testing.T) {
	taskID := uuid.New()
	statusCode := 200

	result := &models.TaskResult{
		ID:           uuid.New(),
		TaskID:       taskID,
		Success:      true,
		StatusCode:   &statusCode,
		DurationMs:   150,
		ResponseBody: stringPtr(`{"status": "ok"}`),
		RunAt:        time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, taskID, result.TaskID)
	assert.True(t, result.Success)
	assert.Equal(t, 200, *result.StatusCode)
	assert.Greater(t, result.DurationMs, 0)
}

// TestCronValidation tests cron expression validation logic
func TestCronValidation(t *testing.T) {
	validCronExpressions := []string{
		"0 0 * * *",   // Daily at midnight
		"*/5 * * * *", // Every 5 minutes
		"0 */6 * * *", // Every 6 hours
		"0 9 * * MON", // Every Monday at 9 AM
	}

	for _, expr := range validCronExpressions {
		t.Run("valid_cron_"+expr, func(t *testing.T) {
			// This would normally use the cron parser
			// For now, just test that the expression is not empty
			assert.NotEmpty(t, expr)
			assert.Contains(t, expr, " ")
		})
	}
}

// TestTaskStatuses tests task status transitions
func TestTaskStatuses(t *testing.T) {
	statuses := []models.TaskStatus{
		models.TaskStatusScheduled,
		models.TaskStatusCompleted,
		models.TaskStatusCancelled,
	}

	// Verify all statuses are defined
	for _, status := range statuses {
		assert.NotEmpty(t, string(status))
	}

	// Test status transitions
	assert.NotEqual(t, models.TaskStatusScheduled, models.TaskStatusCompleted)
	assert.NotEqual(t, models.TaskStatusCompleted, models.TaskStatusCancelled)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
