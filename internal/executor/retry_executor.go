package executor

import (
    "log"
    "time"

    "task-scheduler/internal/models"
)

type RetryExecutor struct {
    executor    ExecutorInterface
    maxRetries  int
    retryDelay  time.Duration
}

func NewRetryExecutor(executor ExecutorInterface, maxRetries int, retryDelay time.Duration) *RetryExecutor {
    return &RetryExecutor{
        executor:   executor,
        maxRetries: maxRetries,
        retryDelay: retryDelay,
    }
}

func (r *RetryExecutor) Execute(task *models.Task) *models.TaskResult {
    var lastResult *models.TaskResult
    
    for attempt := 0; attempt <= r.maxRetries; attempt++ {
        if attempt > 0 {
            log.Printf("Retrying task %s (attempt %d/%d)", task.ID, attempt+1, r.maxRetries+1)
            time.Sleep(r.retryDelay)
        }
        
        result := r.executor.Execute(task)
        lastResult = result
        
        // If successful, return immediately
        if result.Success {
            if attempt > 0 {
                log.Printf("Task %s succeeded on attempt %d", task.ID, attempt+1)
            }
            return result
        }
        
        // Log failure
        errorMsg := "unknown error"
        if result.ErrorMessage != nil {
            errorMsg = *result.ErrorMessage
        }
        log.Printf("Task %s failed on attempt %d: %s", task.ID, attempt+1, errorMsg)
    }
    
    // All attempts failed
    log.Printf("Task %s failed after %d attempts", task.ID, r.maxRetries+1)
    return lastResult
}

func (r *RetryExecutor) ExecuteWithTimeout(task *models.Task, timeout time.Duration) *models.TaskResult {
    var lastResult *models.TaskResult
    
    for attempt := 0; attempt <= r.maxRetries; attempt++ {
        if attempt > 0 {
            log.Printf("Retrying task %s with timeout (attempt %d/%d)", task.ID, attempt+1, r.maxRetries+1)
            time.Sleep(r.retryDelay)
        }
        
        result := r.executor.ExecuteWithTimeout(task, timeout)
        lastResult = result
        
        // If successful, return immediately
        if result.Success {
            if attempt > 0 {
                log.Printf("Task %s succeeded on attempt %d", task.ID, attempt+1)
            }
            return result
        }
        
        // Log failure
        errorMsg := "unknown error"
        if result.ErrorMessage != nil {
            errorMsg = *result.ErrorMessage
        }
        log.Printf("Task %s failed on attempt %d: %s", task.ID, attempt+1, errorMsg)
    }
    
    // All attempts failed
    log.Printf("Task %s failed after %d attempts", task.ID, r.maxRetries+1)
    return lastResult
}
