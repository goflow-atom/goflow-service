// Package server provides HTTP server initialization and management for the GoFlow Workflow Engine.
//
// This package handles the creation and lifecycle management of the HTTP server,
// including Gin router setup, graceful shutdown, and configuration management.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Config holds the HTTP server configuration.
type Config struct {
	// Port is the HTTP server port
	Port int `json:"port"`
	// Host is the HTTP server host (default: "0.0.0.0")
	Host string `json:"host"`
	// Mode is the Gin mode (debug, release, test)
	Mode string `json:"mode"`
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration `json:"read_timeout"`
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `json:"write_timeout"`
	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration `json:"idle_timeout"`
	// ShutdownTimeout is the maximum duration to wait for graceful shutdown
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

// DefaultConfig returns the default server configuration.
func DefaultConfig() Config {
	return Config{
		Port:            8080,
		Host:            "0.0.0.0",
		Mode:            gin.DebugMode,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Server represents the HTTP server.
type Server struct {
	config     Config
	logger     *zap.Logger
	router     *gin.Engine
	httpServer *http.Server
}

// New creates a new HTTP server instance.
//
// Parameters:
//   - config: Server configuration
//   - logger: Zap logger instance
//
// Returns:
//   - *Server: New server instance
//   - error: Error if validation fails
//
// Example:
//
//	config := server.DefaultConfig()
//	logger, _ := zap.NewProduction()
//	srv, err := server.New(config, logger)
func New(config Config, logger *zap.Logger) (*Server, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Validate configuration
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d (must be between 1 and 65535)", config.Port)
	}

	if config.Host == "" {
		config.Host = "0.0.0.0"
	}

	if config.Mode == "" {
		config.Mode = gin.DebugMode
	}

	// Set Gin mode
	gin.SetMode(config.Mode)

	// Create Gin router
	router := gin.New()

	// Add default middleware
	router.Use(gin.Recovery())
	router.Use(ginLogger(logger))

	return &Server{
		config: config,
		logger: logger,
		router: router,
	}, nil
}

// SetupRoutes configures the API routes.
//
// This method sets up the route groups for API versioning and defines
// the initial routes for the workflow engine.
func (s *Server) SetupRoutes() {
	// Health check endpoint (no versioning)
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/", s.root)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Placeholder routes - will be implemented in subsequent tasks
		v1.GET("/status", s.status)
	}

	s.logger.Info("Routes configured",
		zap.String("api_version", "v1"),
		zap.String("base_path", "/api/v1"),
	)
}

// Start starts the HTTP server.
//
// This method initializes the HTTP server and starts listening for requests.
// It returns immediately and does not block. Use the context to control shutdown.
//
// Parameters:
//   - ctx: Context for server lifecycle management
//
// Returns:
//   - error: Error if server fails to start
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.logger.Info("Starting HTTP server",
		zap.String("address", addr),
		zap.String("mode", s.config.Mode),
	)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	s.logger.Info("HTTP server started successfully",
		zap.String("address", addr),
	)

	return nil
}

// Stop gracefully shuts down the HTTP server.
//
// This method attempts to gracefully shutdown the server within the configured
// shutdown timeout. It waits for active connections to complete before closing.
//
// Parameters:
//   - ctx: Context for shutdown timeout control
//
// Returns:
//   - error: Error if shutdown fails or times out
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return fmt.Errorf("server not started")
	}

	s.logger.Info("Shutting down HTTP server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.logger.Info("HTTP server stopped successfully")
	return nil
}

// GetRouter returns the Gin router instance.
//
// This is useful for testing and for adding custom routes or middleware.
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// GetAddress returns the server address.
func (s *Server) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}

// healthCheck handles the /health endpoint.
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "goflow-workflow-engine",
		"version": "1.0.0",
	})
}

// root handles the root endpoint.
func (s *Server) root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "GoFlow Workflow Engine",
		"version": "1.0.0",
		"message": "Welcome to GoFlow Workflow Engine API",
		"docs":    "/api/v1/docs",
	})
}

// status handles the /api/v1/status endpoint.
func (s *Server) status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "running",
		"api_version": "v1",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// ginLogger creates a Gin middleware that logs requests using Zap.
func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("Client error", fields...)
		} else {
			logger.Info("Request completed", fields...)
		}
	}
}

