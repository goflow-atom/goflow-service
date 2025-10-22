package dto

import "time"

// RegisterWebhookRequest represents the request to register a webhook
type RegisterWebhookRequest struct {
	WorkflowID   string                 `json:"workflowId" binding:"required,uuid"`
	Path         string                 `json:"path"`
	Secret       string                 `json:"secret"`
	Method       string                 `json:"method" binding:"omitempty,oneof=GET POST PUT DELETE"`
	InputMapping map[string]interface{} `json:"inputMapping"`
}

// WebhookResponse represents a webhook resource
type WebhookResponse struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflowId"`
	URL          string                 `json:"url"`
	Path         string                 `json:"path,omitempty"`
	Method       string                 `json:"method"`
	Secret       string                 `json:"secret,omitempty"`
	InputMapping map[string]interface{} `json:"inputMapping,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
}

