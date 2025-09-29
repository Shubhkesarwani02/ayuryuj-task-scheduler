package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "task-scheduler/internal/metrics"
)

type MetricsHandler struct {
    metrics *metrics.Metrics
}

func NewMetricsHandler(metrics *metrics.Metrics) *MetricsHandler {
    return &MetricsHandler{metrics: metrics}
}

// GetMetrics godoc
// @Summary Get system metrics
// @Description Get execution metrics and statistics
// @Tags metrics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
    c.JSON(http.StatusOK, h.metrics.GetMetrics())
}
