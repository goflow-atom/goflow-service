package handler

import (
	"net/http"
	"time"

	"github.com/goflow-atom/goflow-service/internal/api/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WebhookHandler handles webhook-related endpoints
type WebhookHandler struct {
	// webhookService service.WebhookService // TODO: inject service when implemented
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

// RegisterWebhook handles POST /api/v1/webhooks
// @Summary Register Webhook
// @Description Register a dynamic webhook endpoint for triggering workflows
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param webhook body dto.RegisterWebhookRequest true "Webhook configuration"
// @Security BearerAuth
// @Success 201 {object} dto.WebhookResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/webhooks [post]
// @ID registerWebhook
func (h *WebhookHandler) RegisterWebhook(c *gin.Context) {
	var req dto.RegisterWebhookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// TODO: Call service layer to register webhook
	webhookID := uuid.New().String()
	path := req.Path
	if path == "" {
		path = "/webhook/" + webhookID[:8]
	}

	method := req.Method
	if method == "" {
		method = "POST"
	}

	response := dto.WebhookResponse{
		ID:           webhookID,
		WorkflowID:   req.WorkflowID,
		URL:          "https://api.example.com" + path,
		Path:         path,
		Method:       method,
		Secret:       "********", // Masked for security
		InputMapping: req.InputMapping,
		CreatedAt:    time.Now(),
	}

	c.JSON(http.StatusCreated, response)
}

// GetWebhook handles GET /api/v1/webhooks/:id
// @Summary Get Webhook
// @Description Retrieve webhook details
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} dto.WebhookResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/webhooks/{id} [get]
// @ID getWebhook
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid webhook ID format",
			},
		})
		return
	}

	// TODO: Call service layer to fetch webhook
	response := dto.WebhookResponse{
		ID:           id,
		WorkflowID:   uuid.New().String(),
		URL:          "https://api.example.com/webhook/" + id[:8],
		Path:         "/webhook/" + id[:8],
		Method:       "POST",
		Secret:       "********", // Masked for security
		InputMapping: map[string]interface{}{"userEmail": "request.body.email"},
		CreatedAt:    time.Now().Add(-7 * 24 * time.Hour),
	}

	c.JSON(http.StatusOK, response)
}

// UnregisterWebhook handles DELETE /api/v1/webhooks/:id
// @Summary Unregister Webhook
// @Description Remove a webhook registration
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID" format(uuid)
// @Security BearerAuth
// @Success 204 "Webhook unregistered"
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/webhooks/{id} [delete]
// @ID unregisterWebhook
func (h *WebhookHandler) UnregisterWebhook(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid webhook ID format",
			},
		})
		return
	}

	// TODO: Call service layer to unregister webhook

	c.Status(http.StatusNoContent)
}
