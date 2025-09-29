package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TriggerType string

const (
	TriggerTypeOneOff TriggerType = "one-off"
	TriggerTypeCron   TriggerType = "cron"
)

type TaskStatus string

const (
	TaskStatusScheduled TaskStatus = "scheduled"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusCompleted TaskStatus = "completed"
)

type Headers map[string]string

func (h Headers) Value() (driver.Value, error) {
	return json.Marshal(h)
}

func (h *Headers) Scan(value interface{}) error {
	if value == nil {
		*h = make(Headers)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, h)
}

type Task struct {
	ID          uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string      `json:"name" gorm:"not null"`
	TriggerType TriggerType `json:"trigger_type" gorm:"not null"`
	TriggerTime *time.Time  `json:"trigger_time,omitempty"`
	CronExpr    *string     `json:"cron_expr,omitempty"`
	Method      string      `json:"method" gorm:"not null"`
	URL         string      `json:"url" gorm:"not null"`
	Headers     Headers     `json:"headers,omitempty" gorm:"type:jsonb"`
	Payload     *string     `json:"payload,omitempty" gorm:"type:jsonb"`
	Status      TaskStatus  `json:"status" gorm:"default:scheduled"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	NextRun     *time.Time  `json:"next_run,omitempty"`
}

type TaskResult struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TaskID          uuid.UUID `json:"task_id" gorm:"not null"`
	RunAt           time.Time `json:"run_at" gorm:"not null"`
	StatusCode      *int      `json:"status_code"`
	Success         bool      `json:"success"`
	ResponseHeaders Headers   `json:"response_headers,omitempty" gorm:"type:jsonb"`
	ResponseBody    *string   `json:"response_body,omitempty"`
	ErrorMessage    *string   `json:"error_message,omitempty"`
	DurationMs      int       `json:"duration_ms"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationship
	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// CreateTaskRequest represents the request payload for creating a task
type CreateTaskRequest struct {
	Name    string            `json:"name" binding:"required"`
	Trigger CreateTaskTrigger `json:"trigger" binding:"required"`
	Action  CreateTaskAction  `json:"action" binding:"required"`
}

type CreateTaskTrigger struct {
	Type     TriggerType `json:"type" binding:"required,oneof=one-off cron"`
	DateTime *time.Time  `json:"datetime,omitempty"`
	Cron     *string     `json:"cron,omitempty"`
}

type CreateTaskAction struct {
	Method  string            `json:"method" binding:"required"`
	URL     string            `json:"url" binding:"required,url"`
	Headers map[string]string `json:"headers,omitempty"`
	Payload interface{}       `json:"payload,omitempty"`
}

// UpdateTaskRequest represents the request payload for updating a task
type UpdateTaskRequest struct {
	Name    *string            `json:"name,omitempty"`
	Trigger *CreateTaskTrigger `json:"trigger,omitempty"`
	Action  *CreateTaskAction  `json:"action,omitempty"`
}
