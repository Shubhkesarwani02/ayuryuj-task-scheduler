package repository

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "task-scheduler/internal/models"
)

type TaskRepository struct {
    db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
    return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *models.Task) error {
    return r.db.Create(task).Error
}

func (r *TaskRepository) GetByID(id uuid.UUID) (*models.Task, error) {
    var task models.Task
    err := r.db.First(&task, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &task, nil
}

func (r *TaskRepository) List(limit, offset int, status string) ([]models.Task, int64, error) {
    var tasks []models.Task
    var total int64

    query := r.db.Model(&models.Task{})
    
    if status != "" {
        query = query.Where("status = ?", status)
    }

    err := query.Count(&total).Error
    if err != nil {
        return nil, 0, err
    }

    err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&tasks).Error
    return tasks, total, err
}

func (r *TaskRepository) Update(task *models.Task) error {
    task.UpdatedAt = time.Now()
    return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(id uuid.UUID) error {
    return r.db.Model(&models.Task{}).Where("id = ?", id).Update("status", models.TaskStatusCancelled).Error
}

func (r *TaskRepository) GetScheduledTasks() ([]models.Task, error) {
    var tasks []models.Task
    err := r.db.Where("status = ? AND (next_run IS NULL OR next_run <= ?)", 
        models.TaskStatusScheduled, time.Now()).Find(&tasks).Error
    return tasks, err
}

func (r *TaskRepository) UpdateNextRun(id uuid.UUID, nextRun *time.Time) error {
    return r.db.Model(&models.Task{}).Where("id = ?", id).Update("next_run", nextRun).Error
}
