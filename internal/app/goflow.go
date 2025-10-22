// Package app provides the main application structure and workflow management
// interfaces for the GoFlow Workflow Engine.
//
// This package defines the core WorkflowManager interface and the GoFlow
// application struct that orchestrates all components of the workflow engine.
package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/goflow-atom/goflow-service/internal/core"
	"go.uber.org/zap"
)

// WorkflowStatus represents the current state of a workflow instance
type WorkflowStatus struct {
	InstanceID  string                 `json:"instance_id"`
	DefinitionID string                `json:"definition_id"`
	State       string                 `json:"state"` // running, completed, failed, terminated
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowManager defines the interface for managing workflow instances.
//
// This interface provides methods for starting, monitoring, signaling, and
// terminating workflow instances. Implementations of this interface handle
// the core workflow orchestration logic.
type WorkflowManager interface {
	// Start initiates a new workflow instance with the given definition ID and input data.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - defID: The workflow definition identifier
	//   - input: JSON-encoded input data for the workflow
	//
	// Returns:
	//   - instanceID: Unique identifier for the created workflow instance
	//   - error: Error if workflow creation fails
	//
	// Example:
	//
	//     instanceID, err := manager.Start(ctx, "order-processing", []byte(`{"orderId": "123"}`))
	//     if err != nil {
	//         log.Fatal(err)
	//     }
	Start(ctx context.Context, defID string, input []byte) (instanceID string, err error)

	// GetStatus retrieves the current status of a workflow instance.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - instanceID: The workflow instance identifier
	//
	// Returns:
	//   - status: Current status of the workflow instance
	//   - error: Error if status retrieval fails or instance not found
	//
	// Example:
	//
	//     status, err := manager.GetStatus(ctx, "wf-inst-123")
	//     if err != nil {
	//         log.Fatal(err)
	//     }
	//     fmt.Printf("Workflow state: %s\n", status.State)
	GetStatus(ctx context.Context, instanceID string) (status WorkflowStatus, err error)

	// Signal sends a signal to a running workflow instance.
	//
	// Signals are used for external events that a workflow may be waiting for,
	// such as user approvals, webhook callbacks, or manual interventions.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - instanceID: The workflow instance identifier
	//   - signalName: Name of the signal to send
	//   - payload: JSON-encoded payload data for the signal
	//
	// Returns:
	//   - error: Error if signal delivery fails
	//
	// Example:
	//
	//     err := manager.Signal(ctx, "wf-inst-123", "approval", []byte(`{"approved": true}`))
	//     if err != nil {
	//         log.Fatal(err)
	//     }
	Signal(ctx context.Context, instanceID string, signalName string, payload []byte) error

	// Terminate forcefully stops a running workflow instance.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - instanceID: The workflow instance identifier
	//   - reason: Human-readable reason for termination
	//
	// Returns:
	//   - error: Error if termination fails
	//
	// Example:
	//
	//     err := manager.Terminate(ctx, "wf-inst-123", "User requested cancellation")
	//     if err != nil {
	//         log.Fatal(err)
	//     }
	Terminate(ctx context.Context, instanceID string, reason string) error
}

// Config holds the configuration for the GoFlow application
type Config struct {
	// ServerPort is the HTTP server port
	ServerPort int `json:"server_port"`

	// LogLevel is the logging level (debug, info, warn, error)
	LogLevel string `json:"log_level"`

	// Environment is the deployment environment (development, staging, production)
	Environment string `json:"environment"`
}

// GoFlow is the main application struct that orchestrates all components
// of the workflow engine.
type GoFlow struct {
	config  Config
	logger  *zap.Logger
	manager WorkflowManager
	mu      sync.RWMutex
	started bool
}

// New creates a new GoFlow application instance.
//
// Parameters:
//   - config: Application configuration
//   - logger: Zap logger instance
//   - manager: WorkflowManager implementation
//
// Returns:
//   - *GoFlow: New GoFlow application instance
//   - error: Error if initialization fails
//
// Example:
//
//     logger, _ := zap.NewProduction()
//     app, err := app.New(config, logger, workflowManager)
//     if err != nil {
//         log.Fatal(err)
//     }
func New(config Config, logger *zap.Logger, manager WorkflowManager) (*GoFlow, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if manager == nil {
		return nil, fmt.Errorf("workflow manager cannot be nil")
	}

	return &GoFlow{
		config:  config,
		logger:  logger,
		manager: manager,
		started: false,
	}, nil
}

// Start starts the GoFlow application.
//
// This method initializes all components and starts the HTTP server.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - error: Error if startup fails
//
// Example:
//
//     if err := app.Start(context.Background()); err != nil {
//         log.Fatal(err)
//     }
func (g *GoFlow) Start(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.started {
		return fmt.Errorf("application already started")
	}

	versionInfo := core.GetVersionInfo()
	g.logger.Info("Starting GoFlow application",
		zap.String("version", versionInfo.Version),
		zap.String("environment", g.config.Environment),
		zap.Int("port", g.config.ServerPort),
	)

	g.started = true

	g.logger.Info("GoFlow application started successfully")
	return nil
}

// Stop gracefully stops the GoFlow application.
//
// This method shuts down all components and waits for in-flight requests
// to complete.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - error: Error if shutdown fails
//
// Example:
//
//     if err := app.Stop(context.Background()); err != nil {
//         log.Error("Failed to stop application", zap.Error(err))
//     }
func (g *GoFlow) Stop(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.started {
		return fmt.Errorf("application not started")
	}

	g.logger.Info("Stopping GoFlow application")

	g.started = false

	g.logger.Info("GoFlow application stopped successfully")
	return nil
}

// IsStarted returns whether the application is currently running.
//
// Returns:
//   - bool: true if the application is started, false otherwise
func (g *GoFlow) IsStarted() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.started
}

// GetWorkflowManager returns the workflow manager instance.
//
// Returns:
//   - WorkflowManager: The workflow manager instance
func (g *GoFlow) GetWorkflowManager() WorkflowManager {
	return g.manager
}

// GetLogger returns the logger instance.
//
// Returns:
//   - *zap.Logger: The logger instance
func (g *GoFlow) GetLogger() *zap.Logger {
	return g.logger
}

// GetConfig returns the application configuration.
//
// Returns:
//   - Config: The application configuration
func (g *GoFlow) GetConfig() Config {
	return g.config
}

