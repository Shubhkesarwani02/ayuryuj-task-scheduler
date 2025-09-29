package scheduler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"task-scheduler/internal/models"
)

// MockExecutor for testing scheduler
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(task *models.Task) *models.TaskResult {
	args := m.Called(task)
	return args.Get(0).(*models.TaskResult)
}

// MockTaskRepository for scheduler tests
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) GetByStatus(status models.TaskStatus) ([]models.Task, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func TestCronExpressionParsing(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectValid bool
	}{
		{
			name:        "valid every minute",
			expression:  "* * * * *",
			expectValid: true,
		},
		{
			name:        "valid every hour",
			expression:  "0 * * * *",
			expectValid: true,
		},
		{
			name:        "valid daily at midnight",
			expression:  "0 0 * * *",
			expectValid: true,
		},
		{
			name:        "valid weekly on monday",
			expression:  "0 0 * * 1",
			expectValid: true,
		},
		{
			name:        "invalid expression",
			expression:  "invalid cron",
			expectValid: false,
		},
		{
			name:        "empty expression",
			expression:  "",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

			_, err := parser.Parse(tt.expression)

			if tt.expectValid {
				assert.NoError(t, err, "Expected valid cron expression")
			} else {
				assert.Error(t, err, "Expected invalid cron expression")
			}
		})
	}
}

func TestSchedulerTaskProcessing(t *testing.T) {
	tests := []struct {
		name         string
		tasks        []models.Task
		setupMocks   func(*MockExecutor, *MockTaskRepository)
		expectations func(*testing.T, *MockExecutor, *MockTaskRepository)
	}{
		{
			name: "process scheduled one-off tasks",
			tasks: []models.Task{
				{
					ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
					Name:        "One-off Task",
					TriggerType: models.TriggerTypeOneOff,
					Method:      "GET",
					URL:         "https://example.com/api1",
					Status:      models.TaskStatusScheduled,
					TriggerTime: timePtr(time.Now().Add(-time.Minute)), // Past time
				},
			},
			setupMocks: func(executor *MockExecutor, repo *MockTaskRepository) {
				// Mock successful execution
				taskID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
				result := &models.TaskResult{
					TaskID:     taskID,
					Success:    true,
					StatusCode: intPtr(200),
					DurationMs: 150,
				}
				executor.On("Execute", mock.MatchedBy(func(task *models.Task) bool {
					return task.ID == taskID
				})).Return(result)

				// Mock task status update to completed
				repo.On("Update", mock.MatchedBy(func(task *models.Task) bool {
					return task.ID == taskID && task.Status == models.TaskStatusCompleted
				})).Return(nil)
			},
			expectations: func(t *testing.T, executor *MockExecutor, repo *MockTaskRepository) {
				executor.AssertCalled(t, "Execute", mock.AnythingOfType("*models.Task"))
				repo.AssertCalled(t, "Update", mock.AnythingOfType("*models.Task"))
			},
		},
		{
			name: "skip future one-off tasks",
			tasks: []models.Task{
				{
					ID:          "task-2",
					Name:        "Future Task",
					TriggerType: models.TriggerTypeOneOff,
					Method:      "POST",
					URL:         "https://example.com/api2",
					Status:      models.TaskStatusScheduled,
					ScheduledAt: timePtr(time.Now().Add(time.Hour)), // Future time
				},
			},
			setupMocks: func(executor *MockExecutor, repo *MockTaskRepository) {
				// No execution expected for future tasks
			},
			expectations: func(t *testing.T, executor *MockExecutor, repo *MockTaskRepository) {
				executor.AssertNotCalled(t, "Execute", mock.AnythingOfType("*models.Task"))
				repo.AssertNotCalled(t, "Update", mock.AnythingOfType("*models.Task"))
			},
		},
		{
			name: "process cron tasks",
			tasks: []models.Task{
				{
					ID:             "task-3",
					Name:           "Cron Task",
					TriggerType:    models.TriggerTypeCron,
					Method:         "GET",
					URL:            "https://example.com/api3",
					Status:         models.TaskStatusScheduled,
					CronExpression: stringPtr("* * * * *"),                    // Every minute
					LastExecutedAt: timePtr(time.Now().Add(-2 * time.Minute)), // 2 minutes ago
				},
			},
			setupMocks: func(executor *MockExecutor, repo *MockTaskRepository) {
				// Mock successful execution
				result := &models.TaskResult{
					TaskID:     "task-3",
					Success:    true,
					StatusCode: intPtr(200),
					DurationMs: 200,
				}
				executor.On("Execute", mock.MatchedBy(func(task *models.Task) bool {
					return task.ID == "task-3"
				})).Return(result)

				// Mock task update with new LastExecutedAt
				repo.On("Update", mock.MatchedBy(func(task *models.Task) bool {
					return task.ID == "task-3" && task.LastExecutedAt != nil
				})).Return(nil)
			},
			expectations: func(t *testing.T, executor *MockExecutor, repo *MockTaskRepository) {
				executor.AssertCalled(t, "Execute", mock.AnythingOfType("*models.Task"))
				repo.AssertCalled(t, "Update", mock.AnythingOfType("*models.Task"))
			},
		},
		{
			name: "handle execution failure",
			tasks: []models.Task{
				{
					ID:          "task-4",
					Name:        "Failing Task",
					TriggerType: models.TriggerTypeOneOff,
					Method:      "POST",
					URL:         "https://invalid-url.example.com",
					Status:      models.TaskStatusScheduled,
					ScheduledAt: timePtr(time.Now().Add(-time.Minute)),
				},
			},
			setupMocks: func(executor *MockExecutor, repo *MockTaskRepository) {
				// Mock failed execution
				result := &models.TaskResult{
					TaskID:       "task-4",
					Success:      false,
					ErrorMessage: stringPtr("Connection failed"),
					DurationMs:   5000,
				}
				executor.On("Execute", mock.MatchedBy(func(task *models.Task) bool {
					return task.ID == "task-4"
				})).Return(result)

				// Mock task status update to completed (even on failure)
				repo.On("Update", mock.MatchedBy(func(task *models.Task) bool {
					return task.ID == "task-4" && task.Status == models.TaskStatusCompleted
				})).Return(nil)
			},
			expectations: func(t *testing.T, executor *MockExecutor, repo *MockTaskRepository) {
				executor.AssertCalled(t, "Execute", mock.AnythingOfType("*models.Task"))
				repo.AssertCalled(t, "Update", mock.AnythingOfType("*models.Task"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			mockRepo := new(MockTaskRepository)

			// Setup mocks
			tt.setupMocks(mockExecutor, mockRepo)

			// Simulate scheduler processing logic
			for _, task := range tt.tasks {
				processed := simulateTaskProcessing(&task, mockExecutor, mockRepo)

				// Verify processing logic
				if task.TriggerType == models.TriggerTypeOneOff {
					if task.ScheduledAt != nil && task.ScheduledAt.Before(time.Now()) {
						assert.True(t, processed, "One-off task in the past should be processed")
					} else {
						assert.False(t, processed, "Future one-off task should not be processed")
					}
				} else if task.TriggerType == models.TriggerTypeCron {
					if shouldExecuteCronTask(&task) {
						assert.True(t, processed, "Cron task should be processed if due")
					}
				}
			}

			// Verify expectations
			tt.expectations(t, mockExecutor, mockRepo)
		})
	}
}

func TestCronScheduling(t *testing.T) {
	tests := []struct {
		name           string
		cronExpression string
		lastExecuted   time.Time
		currentTime    time.Time
		shouldExecute  bool
	}{
		{
			name:           "every minute - due",
			cronExpression: "* * * * *",
			lastExecuted:   time.Now().Add(-2 * time.Minute),
			currentTime:    time.Now(),
			shouldExecute:  true,
		},
		{
			name:           "every minute - not due",
			cronExpression: "* * * * *",
			lastExecuted:   time.Now().Add(-30 * time.Second),
			currentTime:    time.Now(),
			shouldExecute:  false,
		},
		{
			name:           "hourly - due",
			cronExpression: "0 * * * *",
			lastExecuted:   time.Now().Add(-61 * time.Minute),
			currentTime:    time.Now(),
			shouldExecute:  true,
		},
		{
			name:           "daily - not due",
			cronExpression: "0 0 * * *",
			lastExecuted:   time.Now().Add(-12 * time.Hour),
			currentTime:    time.Now(),
			shouldExecute:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &models.Task{
				TriggerType:    models.TriggerTypeCron,
				CronExpression: &tt.cronExpression,
				LastExecutedAt: &tt.lastExecuted,
			}

			shouldExecute := shouldExecuteCronTask(task)
			assert.Equal(t, tt.shouldExecute, shouldExecute)
		})
	}
}

func TestSchedulerErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		task          *models.Task
		executorError bool
		repoError     bool
		expectPanic   bool
	}{
		{
			name: "normal execution",
			task: &models.Task{
				ID:          "task-1",
				TriggerType: models.TriggerTypeOneOff,
				Status:      models.TaskStatusScheduled,
				ScheduledAt: timePtr(time.Now().Add(-time.Minute)),
			},
			executorError: false,
			repoError:     false,
			expectPanic:   false,
		},
		{
			name: "executor returns error result",
			task: &models.Task{
				ID:          "task-2",
				TriggerType: models.TriggerTypeOneOff,
				Status:      models.TaskStatusScheduled,
				ScheduledAt: timePtr(time.Now().Add(-time.Minute)),
			},
			executorError: true,
			repoError:     false,
			expectPanic:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			mockRepo := new(MockTaskRepository)

			// Setup mocks based on test case
			if tt.executorError {
				result := &models.TaskResult{
					TaskID:       tt.task.ID,
					Success:      false,
					ErrorMessage: stringPtr("Execution failed"),
				}
				mockExecutor.On("Execute", tt.task).Return(result)
			} else {
				result := &models.TaskResult{
					TaskID:  tt.task.ID,
					Success: true,
				}
				mockExecutor.On("Execute", tt.task).Return(result)
			}

			if tt.repoError {
				mockRepo.On("Update", tt.task).Return(assert.AnError)
			} else {
				mockRepo.On("Update", tt.task).Return(nil)
			}

			// Test execution
			if tt.expectPanic {
				assert.Panics(t, func() {
					simulateTaskProcessing(tt.task, mockExecutor, mockRepo)
				})
			} else {
				assert.NotPanics(t, func() {
					simulateTaskProcessing(tt.task, mockExecutor, mockRepo)
				})
			}
		})
	}
}

// Helper functions to simulate scheduler logic

func simulateTaskProcessing(task *models.Task, executor *MockExecutor, repo *MockTaskRepository) bool {
	// Check if task should be executed
	shouldExecute := false

	if task.TriggerType == models.TriggerTypeOneOff {
		if task.ScheduledAt != nil && task.ScheduledAt.Before(time.Now()) {
			shouldExecute = true
		}
	} else if task.TriggerType == models.TriggerTypeCron {
		shouldExecute = shouldExecuteCronTask(task)
	}

	if !shouldExecute {
		return false
	}

	// Execute task
	result := executor.Execute(task)

	// Update task status based on result
	if task.TriggerType == models.TriggerTypeOneOff {
		task.Status = models.TaskStatusCompleted
	} else if task.TriggerType == models.TriggerTypeCron {
		now := time.Now()
		task.LastExecutedAt = &now
	}

	// Save task updates
	repo.Update(task)

	return true
}

func shouldExecuteCronTask(task *models.Task) bool {
	if task.CronExpression == nil {
		return false
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(*task.CronExpression)
	if err != nil {
		return false
	}

	lastExec := time.Time{}
	if task.LastExecutedAt != nil {
		lastExec = *task.LastExecutedAt
	}

	nextExec := schedule.Next(lastExec)
	return nextExec.Before(time.Now()) || nextExec.Equal(time.Now())
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(i int) *int {
	return &i
}
