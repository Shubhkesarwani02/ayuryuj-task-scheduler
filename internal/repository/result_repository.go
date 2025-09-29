package repository

import (
    "github.com/google/uuid"
    "gorm.io/gorm"

    "task-scheduler/internal/models"
)

type ResultRepository struct {
    db *gorm.DB
}

func NewResultRepository(db *gorm.DB) *ResultRepository {
    return &ResultRepository{db: db}
}

func (r *ResultRepository) Create(result *models.TaskResult) error {
    return r.db.Create(result).Error
}

func (r *ResultRepository) GetByTaskID(taskID uuid.UUID, limit, offset int) ([]models.TaskResult, int64, error) {
    var results []models.TaskResult
    var total int64

    query := r.db.Model(&models.TaskResult{}).Where("task_id = ?", taskID)
    
    err := query.Count(&total).Error
    if err != nil {
        return nil, 0, err
    }

    err = query.Limit(limit).Offset(offset).Order("run_at DESC").Find(&results).Error
    return results, total, err
}

func (r *ResultRepository) List(limit, offset int, taskID *uuid.UUID, success *bool) ([]models.TaskResult, int64, error) {
    var results []models.TaskResult
    var total int64

    query := r.db.Model(&models.TaskResult{}).Preload("Task")
    
    if taskID != nil {
        query = query.Where("task_id = ?", *taskID)
    }
    
    if success != nil {
        query = query.Where("success = ?", *success)
    }

    err := query.Count(&total).Error
    if err != nil {
        return nil, 0, err
    }

    err = query.Limit(limit).Offset(offset).Order("run_at DESC").Find(&results).Error
    return results, total, err
}
