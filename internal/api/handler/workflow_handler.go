package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/goflow-atom/goflow-service/internal/api/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WorkflowHandler handles workflow-related endpoints
type WorkflowHandler struct {
	// workflowService service.WorkflowService // TODO: inject service when implemented
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{}
}

// ListWorkflows handles GET /api/v1/workflows
// @Summary List Workflows
// @Description Retrieve a paginated list of workflows with optional filtering
// @Tags Workflows
// @Accept json
// @Produce json
// @Param page query int false "Page number (1-based)" default(1) minimum(1)
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param status query string false "Filter by workflow status" Enums(active, inactive, draft)
// @Param name query string false "Filter by workflow name (partial match)"
// @Security BearerAuth
// @Success 200 {object} dto.ListWorkflowsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/workflows [get]
// @ID listWorkflows
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	name := c.Query("name")

	// TODO: Call service layer to fetch workflows
	_ = status
	_ = name

	// Mock response
	response := dto.ListWorkflowsResponse{
		Workflows: []dto.WorkflowResponse{
			{
				ID:          uuid.New().String(),
				Name:        "Sample Workflow",
				Description: "A sample workflow",
				Version:     1,
				Status:      "active",
				Definition:  map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}},
				Tags:        []string{"sample"},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				CreatedBy:   "user@example.com",
			},
		},
		Pagination: dto.Pagination{
			Page:  page,
			Limit: limit,
			Total: 1,
			Pages: 1,
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateWorkflow handles POST /api/v1/workflows
// @Summary Create Workflow
// @Description Create a new workflow definition
// @Tags Workflows
// @Accept json
// @Produce json
// @Param workflow body dto.CreateWorkflowRequest true "Workflow definition"
// @Security BearerAuth
// @Success 201 {object} dto.WorkflowResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/workflows [post]
// @ID createWorkflow
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req dto.CreateWorkflowRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// TODO: Call service layer to create workflow
	// For now, return mock response
	response := dto.WorkflowResponse{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Version:     1,
		Status:      req.Status,
		Definition:  req.Definition,
		Tags:        req.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   "user@example.com",
	}

	c.JSON(http.StatusCreated, response)
}

// GetWorkflow handles GET /api/v1/workflows/:id
// @Summary Get Workflow
// @Description Retrieve details of a specific workflow by ID
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} dto.WorkflowResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/workflows/{id} [get]
// @ID getWorkflow
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid workflow ID format",
			},
		})
		return
	}

	// TODO: Call service layer to fetch workflow
	// For now, return mock response
	response := dto.WorkflowResponse{
		ID:          id,
		Name:        "Sample Workflow",
		Description: "A sample workflow",
		Version:     1,
		Status:      "active",
		Definition:  map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}},
		Tags:        []string{"sample"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   "user@example.com",
	}

	c.JSON(http.StatusOK, response)
}

// UpdateWorkflow handles PUT /api/v1/workflows/:id
// @Summary Update Workflow
// @Description Update an existing workflow
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID" format(uuid)
// @Param workflow body dto.UpdateWorkflowRequest true "Updated workflow definition"
// @Security BearerAuth
// @Success 200 {object} dto.WorkflowResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/workflows/{id} [put]
// @ID updateWorkflow
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid workflow ID format",
			},
		})
		return
	}

	var req dto.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// TODO: Call service layer to update workflow
	// For now, return mock response
	response := dto.WorkflowResponse{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Version:     2, // Incremented version
		Status:      req.Status,
		Definition:  req.Definition,
		Tags:        req.Tags,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		CreatedBy:   "user@example.com",
	}

	c.JSON(http.StatusOK, response)
}

// DeleteWorkflow handles DELETE /api/v1/workflows/:id
// @Summary Delete Workflow
// @Description Delete a workflow and all associated executions (soft delete by default)
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID" format(uuid)
// @Param hardDelete query bool false "If true, perform hard delete" default(false)
// @Security BearerAuth
// @Success 204 "Workflow deleted successfully"
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/workflows/{id} [delete]
// @ID deleteWorkflow
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")
	hardDelete := c.DefaultQuery("hardDelete", "false") == "true"

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid workflow ID format",
			},
		})
		return
	}

	// TODO: Call service layer to delete workflow
	_ = hardDelete

	c.Status(http.StatusNoContent)
}

// ExecuteWorkflow handles POST /api/v1/workflows/:id/execute
// @Summary Execute Workflow
// @Description Trigger a synchronous or asynchronous execution of a workflow
// @Tags Workflows
// @Accept json
// @Produce json
// @Param id path string true "Workflow ID" format(uuid)
// @Param async query bool false "If true, return immediately with execution ID" default(false)
// @Param input body dto.ExecuteWorkflowRequest false "Workflow input parameters"
// @Security BearerAuth
// @Success 200 {object} dto.ExecutionResponse "Synchronous execution completed"
// @Success 202 {object} dto.ExecutionResponse "Asynchronous execution started"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/workflows/{id}/execute [post]
// @ID executeWorkflow
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	id := c.Param("id")
	async := c.DefaultQuery("async", "false") == "true"

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid workflow ID format",
			},
		})
		return
	}

	var req dto.ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// TODO: Call service layer to execute workflow
	executionID := uuid.New().String()
	status := "completed"
	statusCode := http.StatusOK

	if async {
		status = "pending"
		statusCode = http.StatusAccepted
	}

	response := dto.ExecutionResponse{
		ID:         executionID,
		WorkflowID: id,
		Status:     status,
		Input:      req.Input,
		Output:     map[string]interface{}{"result": "success"},
		Nodes:      []dto.NodeExecution{},
		StartedAt:  time.Now(),
	}

	if !async {
		endedAt := time.Now()
		response.EndedAt = &endedAt
		response.Duration = 1500 // milliseconds
	}

	c.JSON(statusCode, response)
}
