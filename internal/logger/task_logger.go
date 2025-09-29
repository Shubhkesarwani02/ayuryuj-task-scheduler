package logger

import (
    "encoding/json"
    "log"
    "os"
    "time"

    "task-scheduler/internal/models"
)

type TaskLogger struct {
    file *os.File
}

func NewTaskLogger(filename string) (*TaskLogger, error) {
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, err
    }
    
    return &TaskLogger{file: file}, nil
}

func (l *TaskLogger) LogTaskExecution(task *models.Task, result *models.TaskResult) {
    logEntry := map[string]interface{}{
        "timestamp":        time.Now().Format(time.RFC3339),
        "task_id":          task.ID,
        "task_name":        task.Name,
        "task_method":      task.Method,
        "task_url":         task.URL,
        "execution_time":   result.RunAt.Format(time.RFC3339),
        "duration_ms":      result.DurationMs,
        "status_code":      result.StatusCode,
        "success":          result.Success,
        "error_message":    result.ErrorMessage,
    }
    
    jsonData, err := json.Marshal(logEntry)
    if err != nil {
        log.Printf("Failed to marshal log entry: %v", err)
        return
    }
    
    if _, err := l.file.WriteString(string(jsonData) + "\n"); err != nil {
        log.Printf("Failed to write log entry: %v", err)
    }
}

func (l *TaskLogger) LogTaskScheduled(task *models.Task) {
    logEntry := map[string]interface{}{
        "timestamp":    time.Now().Format(time.RFC3339),
        "event":        "task_scheduled",
        "task_id":      task.ID,
        "task_name":    task.Name,
        "trigger_type": task.TriggerType,
        "next_run":     task.NextRun,
    }
    
    jsonData, err := json.Marshal(logEntry)
    if err != nil {
        log.Printf("Failed to marshal log entry: %v", err)
        return
    }
    
    if _, err := l.file.WriteString(string(jsonData) + "\n"); err != nil {
        log.Printf("Failed to write log entry: %v", err)
    }
}

func (l *TaskLogger) LogTaskCancelled(taskID string, taskName string) {
    logEntry := map[string]interface{}{
        "timestamp": time.Now().Format(time.RFC3339),
        "event":     "task_cancelled",
        "task_id":   taskID,
        "task_name": taskName,
    }
    
    jsonData, err := json.Marshal(logEntry)
    if err != nil {
        log.Printf("Failed to marshal log entry: %v", err)
        return
    }
    
    if _, err := l.file.WriteString(string(jsonData) + "\n"); err != nil {
        log.Printf("Failed to write log entry: %v", err)
    }
}

func (l *TaskLogger) Close() error {
    return l.file.Close()
}
