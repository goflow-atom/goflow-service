package dto

import "time"

// ExecuteWorkflowRequest represents the request to execute a workflow
type ExecuteWorkflowRequest struct {
	Input         map[string]interface{} `json:"input" binding:"required"`
	WebhookSecret string                 `json:"webhookSecret,omitempty"`
}

// ExecutionResponse represents a workflow execution
type ExecutionResponse struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflowId"`
	Status     string                 `json:"status"`
	Input      map[string]interface{} `json:"input,omitempty"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Nodes      []NodeExecution        `json:"nodes,omitempty"`
	StartedAt  time.Time              `json:"startedAt"`
	EndedAt    *time.Time             `json:"endedAt,omitempty"`
	Duration   int                    `json:"duration,omitempty"` // milliseconds
	Error      *ErrorDetails          `json:"error,omitempty"`
}

// CancelExecutionRequest represents the request to cancel an execution
type CancelExecutionRequest struct {
	Reason string `json:"reason" binding:"required"`
}
