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

	"github.com/goflow-atom/goflow-service/internal/config"
	"github.com/goflow-atom/goflow-service/internal/core"
	"go.uber.org/zap"
)

func main() {
	// Initialize application with Wire dependency injection
	srv, cleanup, err := InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer cleanup()

	// Get configuration and logger for logging version info
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	var logger *zap.Logger
	if cfg.App.Environment == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}

	// Print version information
	versionInfo := core.GetVersionInfo()
	logger.Info("GoFlow Workflow Engine",
		zap.String("version", versionInfo.Version),
		zap.String("go_version", versionInfo.GoVersion),
		zap.String("module_path", versionInfo.ModulePath),
		zap.String("environment", cfg.App.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the HTTP server
	if err := srv.Start(ctx); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	logger.Info("Server is running. Press Ctrl+C to stop.",
		zap.String("address", srv.GetAddress()),
	)
	<-sigChan

	logger.Info("Shutdown signal received, stopping server...")

	// Stop the server
	if err := srv.Stop(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	fmt.Println("Server stopped successfully")
}
