package dto

import "time"

// CreateWorkflowRequest represents the request to create a new workflow
type CreateWorkflowRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	Description string                 `json:"description"`
	Status      string                 `json:"status" binding:"omitempty,oneof=draft active inactive"`
	Definition  map[string]interface{} `json:"definition" binding:"required"`
	Tags        []string               `json:"tags"`
}

// UpdateWorkflowRequest represents the request to update an existing workflow
type UpdateWorkflowRequest struct {
	Name        string                 `json:"name" binding:"omitempty,min=1,max=255"`
	Description string                 `json:"description"`
	Status      string                 `json:"status" binding:"omitempty,oneof=draft active inactive"`
	Definition  map[string]interface{} `json:"definition" binding:"required"`
	Tags        []string               `json:"tags"`
}

// WorkflowResponse represents a workflow resource
type WorkflowResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     int                    `json:"version"`
	Status      string                 `json:"status"`
	Definition  map[string]interface{} `json:"definition"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	CreatedBy   string                 `json:"createdBy,omitempty"`
}

// ListWorkflowsResponse represents the response for listing workflows
type ListWorkflowsResponse struct {
	Workflows  []WorkflowResponse `json:"workflows"`
	Pagination Pagination         `json:"pagination"`
}
