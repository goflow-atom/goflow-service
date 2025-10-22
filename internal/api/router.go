package api

import (
	"github.com/goflow-atom/goflow-service/internal/api/handler"

	"github.com/gin-gonic/gin"
)

// Router sets up all API routes
type Router struct {
	engine            *gin.Engine
	healthHandler     *handler.HealthHandler
	workflowHandler   *handler.WorkflowHandler
	executionHandler  *handler.ExecutionHandler
	scheduleHandler   *handler.ScheduleHandler
	webhookHandler    *handler.WebhookHandler
}

// NewRouter creates a new router with all handlers
func NewRouter(version string) *Router {
	return &Router{
		engine:            gin.Default(),
		healthHandler:     handler.NewHealthHandler(version),
		workflowHandler:   handler.NewWorkflowHandler(),
		executionHandler:  handler.NewExecutionHandler(),
		scheduleHandler:   handler.NewScheduleHandler(),
		webhookHandler:    handler.NewWebhookHandler(),
	}
}

// Setup configures all routes
func (r *Router) Setup() *gin.Engine {
	// Public routes
	r.engine.GET("/health", r.healthHandler.HealthCheck)

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Apply authentication middleware to all v1 routes
		// TODO: Implement and uncomment when auth middleware is ready
		// v1.Use(middleware.AuthMiddleware())

		// Workflow routes
		workflows := v1.Group("/workflows")
		{
			workflows.GET("", r.workflowHandler.ListWorkflows)
			workflows.POST("", r.workflowHandler.CreateWorkflow)
			workflows.GET("/:id", r.workflowHandler.GetWorkflow)
			workflows.PUT("/:id", r.workflowHandler.UpdateWorkflow)
			workflows.DELETE("/:id", r.workflowHandler.DeleteWorkflow)
			workflows.POST("/:id/execute", r.workflowHandler.ExecuteWorkflow)
		}

		// Execution routes
		executions := v1.Group("/executions")
		{
			executions.GET("/:id", r.executionHandler.GetExecution)
			executions.POST("/:id", r.executionHandler.CancelExecution)
			executions.GET("/:id/logs", r.executionHandler.GetExecutionLogs)
		}

		// Schedule routes
		schedules := v1.Group("/schedules")
		{
			schedules.GET("", r.scheduleHandler.ListSchedules)
			schedules.POST("", r.scheduleHandler.CreateSchedule)
			schedules.GET("/:id", r.scheduleHandler.GetSchedule)
			schedules.DELETE("/:id", r.scheduleHandler.DeleteSchedule)
		}

		// Webhook routes
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("", r.webhookHandler.RegisterWebhook)
			webhooks.GET("/:id", r.webhookHandler.GetWebhook)
			webhooks.DELETE("/:id", r.webhookHandler.UnregisterWebhook)
		}
	}

	return r.engine
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
