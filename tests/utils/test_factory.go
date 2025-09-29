package utils

import (
	"encoding/json"
	"time"

	"task-scheduler/internal/models"

	"github.com/google/uuid"
)

// TaskFactory provides methods to create test tasks
type TaskFactory struct{}

// NewTaskFactory creates a new task factory
func NewTaskFactory() *TaskFactory {
	return &TaskFactory{}
}

// CreateOneOffTask creates a one-off task for testing
func (f *TaskFactory) CreateOneOffTask(name, url string, triggerTime time.Time) *models.Task {
	return &models.Task{
		ID:          uuid.New(),
		Name:        name,
		TriggerType: models.TriggerTypeOneOff,
		TriggerTime: &triggerTime,
		Method:      "POST",
		URL:         url,
		Headers:     models.Headers{"Content-Type": "application/json"},
		Payload:     stringPtr(`{"test": "data"}`),
		Status:      models.TaskStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		NextRun:     &triggerTime,
	}
}

// CreateCronTask creates a cron task for testing
func (f *TaskFactory) CreateCronTask(name, url, cronExpr string) *models.Task {
	nextRun := time.Now().Add(time.Minute) // Next run in 1 minute
	return &models.Task{
		ID:          uuid.New(),
		Name:        name,
		TriggerType: models.TriggerTypeCron,
		CronExpr:    &cronExpr,
		Method:      "GET",
		URL:         url,
		Headers:     models.Headers{"User-Agent": "task-scheduler-test"},
		Status:      models.TaskStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		NextRun:     &nextRun,
	}
}

// CreateHTTPTask creates a task with specific HTTP configuration
func (f *TaskFactory) CreateHTTPTask(method, url string, headers models.Headers, payload interface{}) *models.Task {
	triggerTime := time.Now().Add(time.Second * 30)

	var payloadStr *string
	if payload != nil {
		if payloadBytes, err := json.Marshal(payload); err == nil {
			payloadStr = stringPtr(string(payloadBytes))
		}
	}

	return &models.Task{
		ID:          uuid.New(),
		Name:        "Test HTTP Task",
		TriggerType: models.TriggerTypeOneOff,
		TriggerTime: &triggerTime,
		Method:      method,
		URL:         url,
		Headers:     headers,
		Payload:     payloadStr,
		Status:      models.TaskStatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		NextRun:     &triggerTime,
	}
}

// CreateTaskRequest creates a task creation request for API testing
func (f *TaskFactory) CreateTaskRequest(name, method, url string, triggerType models.TriggerType) models.CreateTaskRequest {
	req := models.CreateTaskRequest{
		Name: name,
		Action: models.CreateTaskAction{
			Method:  method,
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
		},
	}

	if triggerType == models.TriggerTypeOneOff {
		triggerTime := time.Now().Add(time.Minute)
		req.Trigger = models.CreateTaskTrigger{
			Type:     models.TriggerTypeOneOff,
			DateTime: &triggerTime,
		}
	} else {
		cronExpr := "*/5 * * * *" // Every 5 minutes
		req.Trigger = models.CreateTaskTrigger{
			Type: models.TriggerTypeCron,
			Cron: &cronExpr,
		}
	}

	return req
}

// CreateTaskWithPayload creates a task request with JSON payload
func (f *TaskFactory) CreateTaskWithPayload(name, url string, payload interface{}) models.CreateTaskRequest {
	triggerTime := time.Now().Add(time.Minute)

	return models.CreateTaskRequest{
		Name: name,
		Trigger: models.CreateTaskTrigger{
			Type:     models.TriggerTypeOneOff,
			DateTime: &triggerTime,
		},
		Action: models.CreateTaskAction{
			Method:  "POST",
			URL:     url,
			Headers: map[string]string{"Content-Type": "application/json"},
			Payload: payload,
		},
	}
}

// ResultFactory provides methods to create test task results
type ResultFactory struct{}

// NewResultFactory creates a new result factory
func NewResultFactory() *ResultFactory {
	return &ResultFactory{}
}

// CreateSuccessResult creates a successful task result
func (f *ResultFactory) CreateSuccessResult(taskID uuid.UUID) *models.TaskResult {
	return &models.TaskResult{
		ID:              uuid.New(),
		TaskID:          taskID,
		RunAt:           time.Now(),
		StatusCode:      intPtr(200),
		Success:         true,
		ResponseHeaders: models.Headers{"Content-Type": "application/json"},
		ResponseBody:    stringPtr(`{"status": "success"}`),
		DurationMs:      150,
		CreatedAt:       time.Now(),
	}
}

// CreateErrorResult creates a failed task result
func (f *ResultFactory) CreateErrorResult(taskID uuid.UUID, statusCode int, errorMsg string) *models.TaskResult {
	return &models.TaskResult{
		ID:           uuid.New(),
		TaskID:       taskID,
		RunAt:        time.Now(),
		StatusCode:   intPtr(statusCode),
		Success:      false,
		ErrorMessage: stringPtr(errorMsg),
		DurationMs:   250,
		CreatedAt:    time.Now(),
	}
}

// CreateTimeoutResult creates a timeout task result
func (f *ResultFactory) CreateTimeoutResult(taskID uuid.UUID) *models.TaskResult {
	return &models.TaskResult{
		ID:           uuid.New(),
		TaskID:       taskID,
		RunAt:        time.Now(),
		Success:      false,
		ErrorMessage: stringPtr("request timeout"),
		DurationMs:   30000, // 30 seconds
		CreatedAt:    time.Now(),
	}
}

// TestScenarios provides common test scenarios
type TestScenarios struct {
	TaskFactory   *TaskFactory
	ResultFactory *ResultFactory
}

// NewTestScenarios creates test scenarios helper
func NewTestScenarios() *TestScenarios {
	return &TestScenarios{
		TaskFactory:   NewTaskFactory(),
		ResultFactory: NewResultFactory(),
	}
}

// CreateValidOneOffTaskScenario creates a valid one-off task scenario
func (s *TestScenarios) CreateValidOneOffTaskScenario(mockServerURL string) (*models.Task, models.CreateTaskRequest) {
	triggerTime := time.Now().Add(time.Second * 10)

	task := s.TaskFactory.CreateOneOffTask(
		"Test One-Off Task",
		mockServerURL+"/webhook",
		triggerTime,
	)

	request := models.CreateTaskRequest{
		Name: task.Name,
		Trigger: models.CreateTaskTrigger{
			Type:     models.TriggerTypeOneOff,
			DateTime: &triggerTime,
		},
		Action: models.CreateTaskAction{
			Method:  task.Method,
			URL:     task.URL,
			Headers: map[string]string(task.Headers),
			Payload: map[string]string{"test": "data"},
		},
	}

	return task, request
}

// CreateValidCronTaskScenario creates a valid cron task scenario
func (s *TestScenarios) CreateValidCronTaskScenario(mockServerURL string) (*models.Task, models.CreateTaskRequest) {
	cronExpr := "*/2 * * * *" // Every 2 minutes

	task := s.TaskFactory.CreateCronTask(
		"Test Cron Task",
		mockServerURL+"/api/endpoint",
		cronExpr,
	)

	request := models.CreateTaskRequest{
		Name: task.Name,
		Trigger: models.CreateTaskTrigger{
			Type: models.TriggerTypeCron,
			Cron: &cronExpr,
		},
		Action: models.CreateTaskAction{
			Method:  task.Method,
			URL:     task.URL,
			Headers: map[string]string(task.Headers),
		},
	}

	return task, request
}

// CreateInvalidTaskScenarios creates various invalid task scenarios
func (s *TestScenarios) CreateInvalidTaskScenarios() []models.CreateTaskRequest {
	pastTime := time.Now().Add(-time.Hour)

	return []models.CreateTaskRequest{
		// Invalid URL
		{
			Name: "Invalid URL Task",
			Trigger: models.CreateTaskTrigger{
				Type:     models.TriggerTypeOneOff,
				DateTime: &pastTime,
			},
			Action: models.CreateTaskAction{
				Method: "GET",
				URL:    "not-a-valid-url",
			},
		},
		// Past datetime
		{
			Name: "Past DateTime Task",
			Trigger: models.CreateTaskTrigger{
				Type:     models.TriggerTypeOneOff,
				DateTime: &pastTime,
			},
			Action: models.CreateTaskAction{
				Method: "GET",
				URL:    "https://example.com",
			},
		},
		// Invalid cron expression
		{
			Name: "Invalid Cron Task",
			Trigger: models.CreateTaskTrigger{
				Type: models.TriggerTypeCron,
				Cron: stringPtr("invalid-cron-expression"),
			},
			Action: models.CreateTaskAction{
				Method: "GET",
				URL:    "https://example.com",
			},
		},
		// Missing required fields
		{
			Name: "",
			Trigger: models.CreateTaskTrigger{
				Type: models.TriggerTypeOneOff,
			},
			Action: models.CreateTaskAction{
				Method: "GET",
			},
		},
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
