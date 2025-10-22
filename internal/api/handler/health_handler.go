package handler

import (
	"net/http"
	"time"

	"github.com/goflow-atom/goflow-service/internal/api/dto"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

// HealthCheck handles GET /health
// @Summary Health Check
// @Description Returns the current status of the API server
// @Tags Health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /health [get]
// @ID healthCheck
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(h.startTime)

	response := dto.HealthResponse{
		Status:  "healthy",
		Version: h.version,
		Uptime:  uptime.String(),
	}

	c.JSON(http.StatusOK, response)
}
