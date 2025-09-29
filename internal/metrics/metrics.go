package metrics

import (
    "sync"
    "time"
)

type Metrics struct {
    mu                    sync.RWMutex
    TotalTasksExecuted    int64
    SuccessfulTasks       int64
    FailedTasks           int64
    TotalExecutionTime    time.Duration
    AverageExecutionTime  time.Duration
    TasksPerMinute        float64
    lastMinuteExecutions  []time.Time
}

func NewMetrics() *Metrics {
    return &Metrics{
        lastMinuteExecutions: make([]time.Time, 0),
    }
}

func (m *Metrics) RecordTaskExecution(duration time.Duration, success bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    now := time.Now()
    
    m.TotalTasksExecuted++
    m.TotalExecutionTime += duration
    m.AverageExecutionTime = m.TotalExecutionTime / time.Duration(m.TotalTasksExecuted)
    
    if success {
        m.SuccessfulTasks++
    } else {
        m.FailedTasks++
    }
    
    // Track executions in the last minute
    m.lastMinuteExecutions = append(m.lastMinuteExecutions, now)
    
    // Remove executions older than 1 minute
    cutoff := now.Add(-time.Minute)
    for i, t := range m.lastMinuteExecutions {
        if t.After(cutoff) {
            m.lastMinuteExecutions = m.lastMinuteExecutions[i:]
            break
        }
    }
    
    m.TasksPerMinute = float64(len(m.lastMinuteExecutions))
}

func (m *Metrics) GetMetrics() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    successRate := float64(0)
    if m.TotalTasksExecuted > 0 {
        successRate = float64(m.SuccessfulTasks) / float64(m.TotalTasksExecuted) * 100
    }
    
    return map[string]interface{}{
        "total_tasks_executed":    m.TotalTasksExecuted,
        "successful_tasks":        m.SuccessfulTasks,
        "failed_tasks":           m.FailedTasks,
        "success_rate_percent":   successRate,
        "average_execution_ms":   m.AverageExecutionTime.Milliseconds(),
        "tasks_per_minute":       m.TasksPerMinute,
    }
}

func (m *Metrics) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.TotalTasksExecuted = 0
    m.SuccessfulTasks = 0
    m.FailedTasks = 0
    m.TotalExecutionTime = 0
    m.AverageExecutionTime = 0
    m.TasksPerMinute = 0
    m.lastMinuteExecutions = make([]time.Time, 0)
}
