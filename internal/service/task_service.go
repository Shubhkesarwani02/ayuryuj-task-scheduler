package service

import (
    "github.com/google/uuid"

    "task-scheduler/internal/models"
    "task-scheduler/internal/repository"
    "task-scheduler/internal/scheduler"
)

type TaskService struct {
    taskRepo  *repository.TaskRepository
    scheduler scheduler.TaskScheduler
}

func NewTaskService(taskRepo *repository.TaskRepository, scheduler scheduler.TaskScheduler) *TaskService {
    return &TaskService{
        taskRepo:  taskRepo,
        scheduler: scheduler,
    }
}

func (s *TaskService) CreateTask(task *models.Task) error {
    // Save to database first
    if err := s.taskRepo.Create(task); err != nil {
        return err
    }
    
    // Schedule the task
    return s.scheduler.ScheduleTask(task)
}

func (s *TaskService) UpdateTask(task *models.Task) error {
    // Unschedule the old version
    s.scheduler.UnscheduleTask(task.ID)
    
    // Update in database
    if err := s.taskRepo.Update(task); err != nil {
        return err
    }
    
    // Reschedule if still active
    if task.Status == models.TaskStatusScheduled {
        return s.scheduler.ScheduleTask(task)
    }
    
    return nil
}

func (s *TaskService) DeleteTask(id uuid.UUID) error {
    // Unschedule the task
    s.scheduler.UnscheduleTask(id)
    
    // Mark as cancelled in database
    return s.taskRepo.Delete(id)
}

func (s *TaskService) GetTask(id uuid.UUID) (*models.Task, error) {
    return s.taskRepo.GetByID(id)
}

func (s *TaskService) ListTasks(limit, offset int, status string) ([]models.Task, int64, error) {
    return s.taskRepo.List(limit, offset, status)
}
