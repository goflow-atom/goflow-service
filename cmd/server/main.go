// Package main is the entry point for the GoFlow Workflow Engine server.
//
// This application provides a production-grade workflow orchestration system
// with support for complex DAG-based workflows, conditional execution,
// parallel processing, and external integrations.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goflow-atom/goflow-service/internal/app"
	"github.com/goflow-atom/goflow-service/internal/core"
	"go.uber.org/zap"
)

// mockWorkflowManager is a temporary implementation for testing
type mockWorkflowManager struct{}

func (m *mockWorkflowManager) Start(ctx context.Context, defID string, input []byte) (string, error) {
	return "mock-instance-id", nil
}

func (m *mockWorkflowManager) GetStatus(ctx context.Context, instanceID string) (app.WorkflowStatus, error) {
	return app.WorkflowStatus{
		InstanceID:   instanceID,
		DefinitionID: "mock-def",
		State:        "running",
		StartedAt:    time.Now(),
	}, nil
}

func (m *mockWorkflowManager) Signal(ctx context.Context, instanceID string, signalName string, payload []byte) error {
	return nil
}

func (m *mockWorkflowManager) Terminate(ctx context.Context, instanceID string, reason string) error {
	return nil
}

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Print version information
	versionInfo := core.GetVersionInfo()
	logger.Info("GoFlow Workflow Engine",
		zap.String("version", versionInfo.Version),
		zap.String("go_version", versionInfo.GoVersion),
		zap.String("module_path", versionInfo.ModulePath),
	)

	// Create application configuration
	config := app.Config{
		ServerPort:  8080,
		LogLevel:    "debug",
		Environment: "development",
	}

	// Create workflow manager (mock for now)
	manager := &mockWorkflowManager{}

	// Create GoFlow application
	goflowApp, err := app.New(config, logger, manager)
	if err != nil {
		logger.Fatal("Failed to create application", zap.Error(err))
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the application
	if err := goflowApp.Start(ctx); err != nil {
		logger.Fatal("Failed to start application", zap.Error(err))
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	logger.Info("Application is running. Press Ctrl+C to stop.")
	<-sigChan

	logger.Info("Shutdown signal received, stopping application...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop the application
	if err := goflowApp.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	fmt.Println("Application stopped successfully")
}
