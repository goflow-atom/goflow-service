// Package app provides application-level dependency injection and initialization.
//
// This package uses Google Wire for compile-time dependency injection,
// ensuring type-safe and efficient dependency management.
package app

import (
	"fmt"

	"github.com/goflow-atom/goflow-service/internal/config"
	"github.com/goflow-atom/goflow-service/internal/server"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ProviderSet is the Wire provider set for the application.
// It includes all core dependencies needed to run the GoFlow server.
var ProviderSet = wire.NewSet(
	ProvideConfig,
	ProvideLogger,
	ProvideServerConfig,
	ProvideServer,
	ProvideCleanup,
)

// Cleanup is a function type for cleanup operations.
type Cleanup func()

// ProvideConfig loads and provides the application configuration.
//
// This provider loads configuration from environment variables and
// validates it before returning.
//
// Returns:
//   - *config.Config: Application configuration
//   - error: Error if configuration loading or validation fails
//
// Example:
//
//	cfg, err := ProvideConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
func ProvideConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return cfg, nil
}

// ProvideLogger creates and provides a Zap logger instance.
//
// The logger is configured based on the application environment:
//   - Production: JSON-formatted structured logging
//   - Development: Console-formatted with color and stack traces
//
// Parameters:
//   - cfg: Application configuration
//
// Returns:
//   - *zap.Logger: Configured Zap logger
//   - error: Error if logger initialization fails
//
// Example:
//
//	logger, err := ProvideLogger(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ProvideLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.App.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}

// ProvideServerConfig creates the HTTP server configuration from app config.
//
// This provider extracts server-specific configuration and creates
// a server.Config instance.
//
// Parameters:
//   - cfg: Application configuration
//
// Returns:
//   - server.Config: HTTP server configuration
//
// Example:
//
//	serverCfg := ProvideServerConfig(cfg)
func ProvideServerConfig(cfg *config.Config) server.Config {
	return server.Config{
		Port:            cfg.Server.Port,
		Host:            cfg.Server.Host,
		Mode:            cfg.Server.Mode,
		ReadTimeout:     cfg.Server.ReadTimeout,
		WriteTimeout:    cfg.Server.WriteTimeout,
		IdleTimeout:     cfg.Server.IdleTimeout,
		ShutdownTimeout: cfg.Server.ShutdownTimeout,
	}
}

// ProvideServer creates and provides the HTTP server instance.
//
// This provider initializes the Gin HTTP server with all middleware
// and route configurations.
//
// Parameters:
//   - serverCfg: Server configuration
//   - logger: Zap logger instance
//
// Returns:
//   - *server.Server: Configured HTTP server
//   - error: Error if server creation fails
//
// Example:
//
//	srv, err := ProvideServer(serverCfg, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ProvideServer(serverCfg server.Config, logger *zap.Logger) (*server.Server, error) {
	srv, err := server.New(serverCfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	// Setup routes
	srv.SetupRoutes()

	return srv, nil
}

// ProvideCleanup creates a cleanup function that handles resource cleanup.
//
// This provider returns a function that should be called on application shutdown
// to properly clean up resources like logger buffers.
//
// Parameters:
//   - logger: Zap logger instance to cleanup
//
// Returns:
//   - func(): Cleanup function to be called on shutdown
//
// Example:
//
//	cleanup := ProvideCleanup(logger)
//	defer cleanup()
func ProvideCleanup(logger *zap.Logger) func() {
	return func() {
		if logger != nil {
			_ = logger.Sync()
		}
	}
}

// ProvideDatabaseConnection is a placeholder for future database connection.
// This will be implemented in CONN-001 task.
//
// TODO: Implement in CONN-001
// func ProvideDatabaseConnection(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
//     // Database connection logic will be implemented here
//     return nil, fmt.Errorf("not implemented")
// }

// ProvideRedisClient is a placeholder for future Redis connection.
// This will be implemented in CONN-101 task.
//
// TODO: Implement in CONN-101
// func ProvideRedisClient(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
//     // Redis connection logic will be implemented here
//     return nil, fmt.Errorf("not implemented")
// }

// ProvideKafkaProducer is a placeholder for future Kafka producer.
// This will be implemented in CONN-201 task.
//
// TODO: Implement in CONN-201
// func ProvideKafkaProducer(cfg *config.Config, logger *zap.Logger) (*kafka.Writer, error) {
//     // Kafka producer logic will be implemented here
//     return nil, fmt.Errorf("not implemented")
// }

// ProvideKafkaConsumer is a placeholder for future Kafka consumer.
// This will be implemented in CONN-202 task.
//
// TODO: Implement in CONN-202
// func ProvideKafkaConsumer(cfg *config.Config, logger *zap.Logger) (*kafka.Reader, error) {
//     // Kafka consumer logic will be implemented here
//     return nil, fmt.Errorf("not implemented")
// }

