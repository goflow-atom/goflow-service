package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/goflow-atom/goflow-service/internal/api/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExecutionHandler handles execution-related endpoints
type ExecutionHandler struct {
	// executionService service.ExecutionService // TODO: inject service when implemented
}

// NewExecutionHandler creates a new execution handler
func NewExecutionHandler() *ExecutionHandler {
	return &ExecutionHandler{}
}

// GetExecution handles GET /api/v1/executions/:id
// @Summary Get Execution
// @Description Retrieve details of a specific workflow execution
// @Tags Executions
// @Accept json
// @Produce json
// @Param id path string true "Execution ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} dto.ExecutionResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/executions/{id} [get]
// @ID getExecution
func (h *ExecutionHandler) GetExecution(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid execution ID format",
			},
		})
		return
	}

	// TODO: Call service layer to fetch execution
	// For now, return mock response
	endedAt := time.Now()
	response := dto.ExecutionResponse{
		ID:         id,
		WorkflowID: uuid.New().String(),
		Status:     "completed",
		Input:      map[string]interface{}{"userId": 123, "email": "user@example.com"},
		Output:     map[string]interface{}{"result": "success"},
		Nodes: []dto.NodeExecution{
			{
				ID:       "node-1",
				Type:     "webhook",
				Status:   "completed",
				Duration: 100,
			},
			{
				ID:       "node-2",
				Type:     "http_request",
				Status:   "completed",
				Duration: 500,
			},
		},
		StartedAt: time.Now().Add(-2 * time.Second),
		EndedAt:   &endedAt,
		Duration:  2000,
	}

	c.JSON(http.StatusOK, response)
}

// CancelExecution handles POST /api/v1/executions/:id
// @Summary Cancel Execution
// @Description Cancel a running execution
// @Tags Executions
// @Accept json
// @Produce json
// @Param id path string true "Execution ID" format(uuid)
// @Param request body dto.CancelExecutionRequest true "Cancellation reason"
// @Security BearerAuth
// @Success 200 {object} dto.ExecutionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/executions/{id} [post]
// @ID cancelExecution
func (h *ExecutionHandler) CancelExecution(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid execution ID format",
			},
		})
		return
	}

	var req dto.CancelExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// TODO: Call service layer to cancel execution
	endedAt := time.Now()
	response := dto.ExecutionResponse{
		ID:         id,
		WorkflowID: uuid.New().String(),
		Status:     "cancelled",
		Input:      map[string]interface{}{"userId": 123},
		StartedAt:  time.Now().Add(-1 * time.Minute),
		EndedAt:    &endedAt,
		Duration:   60000,
	}

	c.JSON(http.StatusOK, response)
}

// GetExecutionLogs handles GET /api/v1/executions/:id/logs
// @Summary Get Execution Logs
// @Description Retrieve logs for a specific execution with pagination and filtering
// @Tags Executions
// @Accept json
// @Produce json
// @Param id path string true "Execution ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Logs per page" default(50)
// @Param level query string false "Filter by log level" Enums(info, warn, error)
// @Param nodeId query string false "Filter logs by node ID"
// @Security BearerAuth
// @Success 200 {object} dto.ListLogsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/executions/{id}/logs [get]
// @ID getExecutionLogs
func (h *ExecutionHandler) GetExecutionLogs(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	level := c.Query("level")
	nodeID := c.Query("nodeId")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid execution ID format",
			},
		})
		return
	}

	// TODO: Call service layer to fetch logs
	_ = level
	_ = nodeID

	// Mock response
	response := dto.ListLogsResponse{
		Logs: []dto.LogEntry{
			{
				ID:          uuid.New().String(),
				ExecutionID: id,
				NodeID:      "node-1",
				Level:       "info",
				Message:     "Node started execution",
				Timestamp:   time.Now().Add(-2 * time.Second),
				Metadata:    map[string]interface{}{"duration": 100},
			},
			{
				ID:          uuid.New().String(),
				ExecutionID: id,
				NodeID:      "node-2",
				Level:       "info",
				Message:     "HTTP request completed",
				Timestamp:   time.Now().Add(-1 * time.Second),
				Metadata:    map[string]interface{}{"statusCode": 200},
			},
		},
		Pagination: dto.Pagination{
			Page:  page,
			Limit: limit,
			Total: 2,
			Pages: 1,
		},
	}

	c.JSON(http.StatusOK, response)
}
