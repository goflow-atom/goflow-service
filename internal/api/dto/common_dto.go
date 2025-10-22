package dto

import "time"

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	Uptime  string `json:"uptime,omitempty"`
}

// ErrorDetails represents detailed error information for executions
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

// NodeExecution represents a single node execution within a workflow
type NodeExecution struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"`
	Input    map[string]interface{} `json:"input,omitempty"`
	Output   map[string]interface{} `json:"output,omitempty"`
	Duration int                    `json:"duration,omitempty"` // milliseconds
	Error    *ErrorDetails          `json:"error,omitempty"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	ID          string                 `json:"id"`
	ExecutionID string                 `json:"executionId"`
	NodeID      string                 `json:"nodeId,omitempty"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ListLogsResponse represents the response for listing execution logs
type ListLogsResponse struct {
	Logs       []LogEntry `json:"logs"`
	Pagination Pagination `json:"pagination"`
}
