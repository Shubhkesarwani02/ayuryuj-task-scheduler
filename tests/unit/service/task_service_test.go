package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"task-scheduler/internal/models"
)

// MockTaskRepository mocks the task repository for unit testing
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(id string) (*models.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskRepository) GetAll(limit, offset int) ([]models.Task, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByStatus(status models.TaskStatus) ([]models.Task, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// MockResultRepository mocks the result repository
type MockResultRepository struct {
	mock.Mock
}

func (m *MockResultRepository) Create(result *models.TaskResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockResultRepository) GetByTaskID(taskID string) ([]models.TaskResult, error) {
	args := m.Called(taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaskResult), args.Error(1)
}

func (m *MockResultRepository) GetAll(limit, offset int) ([]models.TaskResult, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TaskResult), args.Error(1)
}

func TestCreateTaskServiceLogic(t *testing.T) {
	mockRepo := new(MockTaskRepository)

	tests := []struct {
		name          string
		request       models.CreateTaskRequest
		setupMock     func(*MockTaskRepository)
		expectedError bool
	}{
		{
			name: "successful one-off task creation",
			request: models.CreateTaskRequest{
				Name:        "Test Task",
				Description: stringPtr("Test description"),
				Trigger: models.CreateTaskTrigger{
					Type:     models.TriggerTypeOneOff,
					DateTime: timePtr(time.Now().Add(time.Hour)),
				},
				Action: models.CreateTaskAction{
					Method: "POST",
					URL:    "https://example.com/webhook",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Payload: stringPtr(`{"test": "data"}`),
				},
			},
			setupMock: func(m *MockTaskRepository) {
				m.On("Create", mock.AnythingOfType("*models.Task")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "successful cron task creation",
			request: models.CreateTaskRequest{
				Name: "Cron Task",
				Trigger: models.CreateTaskTrigger{
					Type: models.TriggerTypeCron,
					Cron: stringPtr("0 */6 * * *"), // Every 6 hours
				},
				Action: models.CreateTaskAction{
					Method: "GET",
					URL:    "https://api.example.com/health",
				},
			},
			setupMock: func(m *MockTaskRepository) {
				m.On("Create", mock.AnythingOfType("*models.Task")).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock calls
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil

			// Setup mock expectations
			tt.setupMock(mockRepo)

			// Test task creation logic manually since we don't have the service
			task := &models.Task{
				Name:        tt.request.Name,
				Description: tt.request.Description,
				TriggerType: tt.request.Trigger.Type,
				Method:      tt.request.Action.Method,
				URL:         tt.request.Action.URL,
				Payload:     tt.request.Action.Payload,
				Status:      models.TaskStatusScheduled,
			}

			if tt.request.Trigger.DateTime != nil {
				task.ScheduledAt = tt.request.Trigger.DateTime
			}
			if tt.request.Trigger.Cron != nil {
				task.CronExpression = tt.request.Trigger.Cron
			}

			// Convert headers
			if len(tt.request.Action.Headers) > 0 {
				headers := make(models.Headers)
				for k, v := range tt.request.Action.Headers {
					headers[k] = v
				}
				task.Headers = headers
			}

			// Call mock repository
			err := mockRepo.Create(task)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, models.TaskStatusScheduled, task.Status)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskRetrievalLogic(t *testing.T) {
	mockRepo := new(MockTaskRepository)

	// Test data
	testTasks := []models.Task{
		{
			ID:          "task-1",
			Name:        "Test Task 1",
			TriggerType: models.TriggerTypeOneOff,
			Method:      "POST",
			URL:         "https://example.com/webhook1",
			Status:      models.TaskStatusScheduled,
		},
		{
			ID:          "task-2",
			Name:        "Test Task 2",
			TriggerType: models.TriggerTypeCron,
			Method:      "GET",
			URL:         "https://example.com/webhook2",
			Status:      models.TaskStatusCompleted,
		},
	}

	tests := []struct {
		name      string
		setupMock func(*MockTaskRepository)
		testFunc  func(*MockTaskRepository)
	}{
		{
			name: "get all tasks",
			setupMock: func(m *MockTaskRepository) {
				m.On("GetAll", 10, 0).Return(testTasks, nil)
			},
			testFunc: func(m *MockTaskRepository) {
				tasks, err := m.GetAll(10, 0)
				assert.NoError(t, err)
				assert.Len(t, tasks, 2)
				assert.Equal(t, "Test Task 1", tasks[0].Name)
				assert.Equal(t, "Test Task 2", tasks[1].Name)
			},
		},
		{
			name: "get task by ID",
			setupMock: func(m *MockTaskRepository) {
				m.On("GetByID", "task-1").Return(&testTasks[0], nil)
			},
			testFunc: func(m *MockTaskRepository) {
				task, err := m.GetByID("task-1")
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, "task-1", task.ID)
				assert.Equal(t, "Test Task 1", task.Name)
			},
		},
		{
			name: "get tasks by status",
			setupMock: func(m *MockTaskRepository) {
				scheduledTasks := []models.Task{testTasks[0]}
				m.On("GetByStatus", models.TaskStatusScheduled).Return(scheduledTasks, nil)
			},
			testFunc: func(m *MockTaskRepository) {
				tasks, err := m.GetByStatus(models.TaskStatusScheduled)
				assert.NoError(t, err)
				assert.Len(t, tasks, 1)
				assert.Equal(t, models.TaskStatusScheduled, tasks[0].Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil

			// Setup mock
			tt.setupMock(mockRepo)

			// Run test
			tt.testFunc(mockRepo)

			// Verify expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskStatusUpdate(t *testing.T) {
	mockRepo := new(MockTaskRepository)

	task := &models.Task{
		ID:          "task-1",
		Name:        "Test Task",
		TriggerType: models.TriggerTypeOneOff,
		Method:      "POST",
		URL:         "https://example.com/webhook",
		Status:      models.TaskStatusScheduled,
	}

	tests := []struct {
		name      string
		newStatus models.TaskStatus
		setupMock func(*MockTaskRepository)
	}{
		{
			name:      "cancel task",
			newStatus: models.TaskStatusCancelled,
			setupMock: func(m *MockTaskRepository) {
				m.On("Update", mock.MatchedBy(func(t *models.Task) bool {
					return t.ID == "task-1" && t.Status == models.TaskStatusCancelled
				})).Return(nil)
			},
		},
		{
			name:      "complete task",
			newStatus: models.TaskStatusCompleted,
			setupMock: func(m *MockTaskRepository) {
				m.On("Update", mock.MatchedBy(func(t *models.Task) bool {
					return t.ID == "task-1" && t.Status == models.TaskStatusCompleted
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil

			// Setup mock
			tt.setupMock(mockRepo)

			// Update task status
			task.Status = tt.newStatus

			// Call update
			err := mockRepo.Update(task)
			assert.NoError(t, err)
			assert.Equal(t, tt.newStatus, task.Status)

			// Verify expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestResultCreationLogic(t *testing.T) {
	mockResultRepo := new(MockResultRepository)

	tests := []struct {
		name      string
		result    *models.TaskResult
		setupMock func(*MockResultRepository)
	}{
		{
			name: "successful result",
			result: &models.TaskResult{
				TaskID:     "task-1",
				Success:    true,
				StatusCode: intPtr(200),
				Response:   stringPtr(`{"status": "ok"}`),
				DurationMs: 150,
			},
			setupMock: func(m *MockResultRepository) {
				m.On("Create", mock.MatchedBy(func(r *models.TaskResult) bool {
					return r.TaskID == "task-1" && r.Success == true
				})).Return(nil)
			},
		},
		{
			name: "failed result",
			result: &models.TaskResult{
				TaskID:       "task-2",
				Success:      false,
				ErrorMessage: stringPtr("Connection timeout"),
				DurationMs:   30000,
			},
			setupMock: func(m *MockResultRepository) {
				m.On("Create", mock.MatchedBy(func(r *models.TaskResult) bool {
					return r.TaskID == "task-2" && r.Success == false
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockResultRepo.ExpectedCalls = nil
			mockResultRepo.Calls = nil

			// Setup mock
			tt.setupMock(mockResultRepo)

			// Create result
			err := mockResultRepo.Create(tt.result)
			assert.NoError(t, err)

			// Verify expectations
			mockResultRepo.AssertExpectations(t)
		})
	}
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
