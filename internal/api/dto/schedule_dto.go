package dto

import "time"

// CreateScheduleRequest represents the request to create a schedule
type CreateScheduleRequest struct {
	WorkflowID string                 `json:"workflowId" binding:"required,uuid"`
	Cron       string                 `json:"cron" binding:"required"`
	Input      map[string]interface{} `json:"input"`
	Timezone   string                 `json:"timezone"`
	Enabled    bool                   `json:"enabled"`
}

// ScheduleResponse represents a schedule resource
type ScheduleResponse struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflowId"`
	Cron       string                 `json:"cron"`
	Input      map[string]interface{} `json:"input,omitempty"`
	Timezone   string                 `json:"timezone,omitempty"`
	Enabled    bool                   `json:"enabled"`
	NextRunAt  *time.Time             `json:"nextRunAt,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// ListSchedulesResponse represents the response for listing schedules
type ListSchedulesResponse struct {
	Schedules []ScheduleResponse `json:"schedules"`
}

