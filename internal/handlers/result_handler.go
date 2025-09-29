package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "task-scheduler/internal/repository"
)

type ResultHandler struct {
    resultRepo *repository.ResultRepository
}

func NewResultHandler(resultRepo *repository.ResultRepository) *ResultHandler {
    return &ResultHandler{resultRepo: resultRepo}
}

// GetResults godoc
// @Summary Get all task execution results
// @Description Get paginated list of all task execution results with optional filtering
// @Tags results
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param task_id query string false "Filter by task ID"
// @Param success query bool false "Filter by success status"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /results [get]
func (h *ResultHandler) GetResults(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }

    offset := (page - 1) * limit

    var taskID *uuid.UUID
    if taskIDStr := c.Query("task_id"); taskIDStr != "" {
        if id, err := uuid.Parse(taskIDStr); err == nil {
            taskID = &id
        }
    }

    var success *bool
    if successStr := c.Query("success"); successStr != "" {
        if successBool, err := strconv.ParseBool(successStr); err == nil {
            success = &successBool
        }
    }

    results, total, err := h.resultRepo.List(limit, offset, taskID, success)
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
