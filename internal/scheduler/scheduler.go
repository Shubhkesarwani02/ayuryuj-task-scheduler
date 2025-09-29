package scheduler

import (
    "context"
    "log"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/robfig/cron/v3"

    "task-scheduler/internal/executor"
    "task-scheduler/internal/logger"
    "task-scheduler/internal/metrics"
    "task-scheduler/internal/models"
    "task-scheduler/internal/repository"
)

type Scheduler struct {
    taskRepo     *repository.TaskRepository
    resultRepo   *repository.ResultRepository
    executor     executor.ExecutorInterface
    taskLogger   *logger.TaskLogger
    metrics      *metrics.Metrics
    cron         *cron.Cron
    oneOffTasks  map[uuid.UUID]*time.Timer
    mu           sync.RWMutex
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup
}

func NewScheduler(taskRepo *repository.TaskRepository, resultRepo *repository.ResultRepository, 
    httpExecutor *executor.HTTPExecutor, taskLogger *logger.TaskLogger, metrics *metrics.Metrics) *Scheduler {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Wrap executor with retry logic
    retryExecutor := executor.NewRetryExecutor(httpExecutor, 2, 5*time.Second)
    
    return &Scheduler{
        taskRepo:    taskRepo,
        resultRepo:  resultRepo,
        executor:    retryExecutor,
        taskLogger:  taskLogger,
        metrics:     metrics,
        cron:        cron.New(cron.WithSeconds()),
        oneOffTasks: make(map[uuid.UUID]*time.Timer),
        ctx:         ctx,
        cancel:      cancel,
    }
}

func (s *Scheduler) Start() error {
    log.Println("Starting task scheduler...")
    
    // Start cron scheduler
    s.cron.Start()
    
    // Load existing tasks from database
    if err := s.loadExistingTasks(); err != nil {
        return err
    }
    
    // Start periodic task checker for missed tasks
    s.wg.Add(1)
    go s.periodicTaskChecker()
    
    log.Println("Task scheduler started successfully")
    return nil
}

func (s *Scheduler) Stop() {
    log.Println("Stopping task scheduler...")
    
    // Cancel context to stop all goroutines
    s.cancel()
    
    // Stop cron scheduler
    s.cron.Stop()
    
    // Cancel all one-off timers
    s.mu.Lock()
    for _, timer := range s.oneOffTasks {
        timer.Stop()
    }
    s.oneOffTasks = make(map[uuid.UUID]*time.Timer)
    s.mu.Unlock()
    
    // Wait for all goroutines to finish
    s.wg.Wait()
    
    log.Println("Task scheduler stopped")
}

func (s *Scheduler) ScheduleTask(task *models.Task) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    switch task.TriggerType {
    case models.TriggerTypeOneOff:
        return s.scheduleOneOffTask(task)
    case models.TriggerTypeCron:
        return s.scheduleCronTask(task)
    default:
        log.Printf("Unknown trigger type: %s for task %s", task.TriggerType, task.ID)
        return nil
    }
}

func (s *Scheduler) UnscheduleTask(taskID uuid.UUID) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Remove from one-off tasks if exists
    if timer, exists := s.oneOffTasks[taskID]; exists {
        timer.Stop()
        delete(s.oneOffTasks, taskID)
        log.Printf("Unscheduled one-off task: %s", taskID)
    }
    
    // Note: Cron tasks are handled by checking status in the execution function
    log.Printf("Task unscheduled: %s", taskID)
}

func (s *Scheduler) scheduleOneOffTask(task *models.Task) error {
    if task.TriggerTime == nil {
        log.Printf("One-off task %s has no trigger time", task.ID)
        return nil
    }
    
    // Cancel existing timer if any
    if timer, exists := s.oneOffTasks[task.ID]; exists {
        timer.Stop()
    }
    
    duration := time.Until(*task.TriggerTime)
    if duration <= 0 {
        // Task should run immediately
        go s.executeTask(task)
        return nil
    }
    
    timer := time.AfterFunc(duration, func() {
        s.executeTask(task)
        
        // Clean up timer reference
        s.mu.Lock()
        delete(s.oneOffTasks, task.ID)
        s.mu.Unlock()
    })
    
    s.oneOffTasks[task.ID] = timer
    log.Printf("Scheduled one-off task %s to run at %s", task.ID, task.TriggerTime.Format(time.RFC3339))
    
    return nil
}

func (s *Scheduler) scheduleCronTask(task *models.Task) error {
    if task.CronExpr == nil {
        log.Printf("Cron task %s has no cron expression", task.ID)
        return nil
    }
    
    _, err := s.cron.AddFunc(*task.CronExpr, func() {
        // Check if task is still active before executing
        currentTask, err := s.taskRepo.GetByID(task.ID)
        if err != nil {
            log.Printf("Failed to get task %s: %v", task.ID, err)
            return
        }
        
        if currentTask.Status != models.TaskStatusScheduled {
            log.Printf("Skipping execution of task %s (status: %s)", task.ID, currentTask.Status)
            return
        }
        
        s.executeTask(currentTask)
        
        // Update next run time
        if nextRun, err := s.calculateNextCronRun(*currentTask.CronExpr); err == nil {
            s.taskRepo.UpdateNextRun(currentTask.ID, &nextRun)
        }
    })
    
    if err != nil {
        log.Printf("Failed to schedule cron task %s: %v", task.ID, err)
        return err
    }
    
    log.Printf("Scheduled cron task %s with expression: %s", task.ID, *task.CronExpr)
    return nil
}

func (s *Scheduler) executeTask(task *models.Task) {
    log.Printf("Executing task: %s (%s)", task.ID, task.Name)
    
    startTime := time.Now()
    result := s.executor.Execute(task)
    result.RunAt = startTime
    result.TaskID = task.ID
    
    // Record metrics
    duration := time.Duration(result.DurationMs) * time.Millisecond
    s.metrics.RecordTaskExecution(duration, result.Success)
    
    // Log execution
    s.taskLogger.LogTaskExecution(task, result)
    
    // Save result to database
    if err := s.resultRepo.Create(result); err != nil {
        log.Printf("Failed to save result for task %s: %v", task.ID, err)
    }
    
    // Update task status if it's a one-off task
    if task.TriggerType == models.TriggerTypeOneOff {
        task.Status = models.TaskStatusCompleted
        if err := s.taskRepo.Update(task); err != nil {
            log.Printf("Failed to update task status for %s: %v", task.ID, err)
        }
    }
    
    log.Printf("Task execution completed: %s (success: %t, duration: %dms)", 
        task.ID, result.Success, result.DurationMs)
}

func (s *Scheduler) loadExistingTasks() error {
    log.Println("Loading existing tasks from database...")
    
    tasks, err := s.taskRepo.GetScheduledTasks()
    if err != nil {
        return err
    }
    
    for _, task := range tasks {
        if err := s.ScheduleTask(&task); err != nil {
            log.Printf("Failed to schedule task %s: %v", task.ID, err)
            continue
        }
    }
    
    log.Printf("Loaded %d existing tasks", len(tasks))
    return nil
}

func (s *Scheduler) periodicTaskChecker() {
    defer s.wg.Done()
    
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.checkMissedTasks()
        }
    }
}

func (s *Scheduler) checkMissedTasks() {
    tasks, err := s.taskRepo.GetScheduledTasks()
    if err != nil {
        log.Printf("Failed to check for missed tasks: %v", err)
        return
    }
    
    now := time.Now()
    for _, task := range tasks {
        if task.NextRun != nil && task.NextRun.Before(now) {
            // This task should have run already
            if task.TriggerType == models.TriggerTypeOneOff {
                // Execute missed one-off task
                go s.executeTask(&task)
            } else if task.TriggerType == models.TriggerTypeCron {
                // Update next run time for cron task
                if nextRun, err := s.calculateNextCronRun(*task.CronExpr); err == nil {
                    s.taskRepo.UpdateNextRun(task.ID, &nextRun)
                }
            }
        }
    }
}

func (s *Scheduler) calculateNextCronRun(cronExpr string) (time.Time, error) {
    parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    schedule, err := parser.Parse(cronExpr)
    if err != nil {
        return time.Time{}, err
    }
    return schedule.Next(time.Now()), nil
}

// TaskScheduler interface for dependency injection
type TaskScheduler interface {
    Start() error
    Stop()
    ScheduleTask(task *models.Task) error
    UnscheduleTask(taskID uuid.UUID)
}
