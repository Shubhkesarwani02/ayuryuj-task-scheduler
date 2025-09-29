package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/robfig/cron/v3"

    "task-scheduler/internal/models"
    "task-scheduler/internal/repository"
)

type TaskHandler struct {
    taskRepo   *repository.TaskRepository
    resultRepo *repository.ResultRepository
}

func NewTaskHandler(taskRepo *repository.TaskRepository, resultRepo *repository.ResultRepository) *TaskHandler {
    return &TaskHandler{
        taskRepo:   taskRepo,
        resultRepo: resultRepo,
    }
}

// CreateTask godoc
// @Summary Create a new task
// @Description Create a new scheduled task with trigger and action configuration
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body models.CreateTaskRequest true "Task creation request"
// @Success 201 {object} models.Task
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
    var req models.CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate trigger configuration
    if err := h.validateTrigger(&req.Trigger); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    task := &models.Task{
        Name:        req.Name,
        TriggerType: req.Trigger.Type,
        Method:      req.Action.Method,
        URL:         req.Action.URL,
        Headers:     req.Action.Headers,
        Status:      models.TaskStatusScheduled,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // Set trigger-specific fields
    if req.Trigger.Type == models.TriggerTypeOneOff {
        task.TriggerTime = req.Trigger.DateTime
        task.NextRun = req.Trigger.DateTime
    } else {
        task.CronExpr = req.Trigger.Cron
        // Calculate next run time for cron
        if nextRun, err := h.calculateNextCronRun(*req.Trigger.Cron); err == nil {
            task.NextRun = &nextRun
        }
    }

    // Handle payload
    if req.Action.Payload != nil {
        payloadBytes, _ := json.Marshal(req.Action.Payload)
        payloadStr := string(payloadBytes)
        task.Payload = &payloadStr
    }

    if err := h.taskRepo.Create(task); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
        return
    }

    c.JSON(http.StatusCreated, task)
}

// GetTasks godoc
// @Summary List all tasks
// @Description Get a paginated list of tasks with optional status filtering
// @Tags tasks
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status" Enums(scheduled,cancelled,completed)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    status := c.Query("status")

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }

    offset := (page - 1) * limit

    tasks, total, err := h.taskRepo.List(limit, offset, status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "tasks": tasks,
        "pagination": gin.H{
            "page":       page,
            "limit":      limit,
            "total":      total,
            "total_pages": (total + int64(limit) - 1) / int64(limit),
        },
    })
}

// GetTask godoc
// @Summary Get task by ID
// @Description Get detailed information about a specific task
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.Task
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    task, err := h.taskRepo.GetByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
        return
    }

    c.JSON(http.StatusOK, task)
}

// UpdateTask godoc
// @Summary Update a task
// @Description Update an existing task's configuration
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body models.UpdateTaskRequest true "Task update request"
// @Success 200 {object} models.Task
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    task, err := h.taskRepo.GetByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
        return
    }

    var req models.UpdateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update fields if provided
    if req.Name != nil {
        task.Name = *req.Name
    }

    if req.Trigger != nil {
        if err := h.validateTrigger(req.Trigger); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        task.TriggerType = req.Trigger.Type
        if req.Trigger.Type == models.TriggerTypeOneOff {
            task.TriggerTime = req.Trigger.DateTime
            task.CronExpr = nil
            task.NextRun = req.Trigger.DateTime
        } else {
            task.CronExpr = req.Trigger.Cron
            task.TriggerTime = nil
            if nextRun, err := h.calculateNextCronRun(*req.Trigger.Cron); err == nil {
                task.NextRun = &nextRun
            }
        }
    }

    if req.Action != nil {
        task.Method = req.Action.Method
        task.URL = req.Action.URL
        task.Headers = req.Action.Headers

        if req.Action.Payload != nil {
            payloadBytes, _ := json.Marshal(req.Action.Payload)
            payloadStr := string(payloadBytes)
            task.Payload = &payloadStr
        } else {
            task.Payload = nil
        }
    }

    if err := h.taskRepo.Update(task); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
        return
    }

    c.JSON(http.StatusOK, task)
}

// DeleteTask godoc
// @Summary Cancel a task
// @Description Cancel a task (marks as cancelled, doesn't delete)
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    // Check if task exists
    _, err = h.taskRepo.GetByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
        return
    }

    if err := h.taskRepo.Delete(id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel task"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Task cancelled successfully"})
}

// GetTaskResults godoc
// @Summary Get task execution results
// @Description Get paginated list of execution results for a specific task
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks/{id}/results [get]
func (h *TaskHandler) GetTaskResults(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    // Check if task exists
    _, err = h.taskRepo.GetByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
        return
    }

    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }

    offset := (page - 1) * limit

    results, total, err := h.resultRepo.GetByTaskID(id, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "results": results,
        "pagination": gin.H{
            "page":       page,
            "limit":      limit,
            "total":      total,
            "total_pages": (total + int64(limit) - 1) / int64(limit),
        },
    })
}

func (h *TaskHandler) validateTrigger(trigger *models.CreateTaskTrigger) error {
    if trigger.Type == models.TriggerTypeOneOff {
        if trigger.DateTime == nil {
            return gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePublic, Meta: "datetime is required for one-off triggers"}
        }
        if trigger.DateTime.Before(time.Now()) {
            return gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePublic, Meta: "datetime must be in the future"}
        }
    } else if trigger.Type == models.TriggerTypeCron {
        if trigger.Cron == nil || *trigger.Cron == "" {
            return gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePublic, Meta: "cron expression is required for cron triggers"}
        }
        if _, err := cron.ParseStandard(*trigger.Cron); err != nil {
            return gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePublic, Meta: "invalid cron expression"}
        }
    }
    return nil
}

func (h *TaskHandler) calculateNextCronRun(cronExpr string) (time.Time, error) {
    parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    schedule, err := parser.Parse(cronExpr)
    if err != nil {
        return time.Time{}, err
    }
    return schedule.Next(time.Now()), nil
}
