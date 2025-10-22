// Package middleware provides HTTP middleware for the GoFlow API.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDMiddleware adds a unique request ID to each request.
// The request ID is stored in the Gin context and can be used for tracing.
//
// The request ID is:
// - Generated as a UUID v4
// - Stored in the context with key "request_id"
// - Added to response headers as "X-Request-ID"
// - Included in all log messages for the request
//
// Example:
//
//	router.Use(middleware.RequestIDMiddleware())
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already set in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}

		// Store request ID in context
		c.Set("request_id", requestID)

		// Add request ID to response headers
		c.Header("X-Request-ID", requestID)

		// Continue processing
		c.Next()
	}
}

// LoggerMiddleware creates a Gin middleware that logs requests using Zap.
// This middleware includes request ID tracking for distributed tracing.
//
// Logged fields:
// - request_id: Unique identifier for the request
// - status: HTTP status code
// - method: HTTP method
// - path: Request path
// - query: Query parameters
// - ip: Client IP address
// - latency: Request processing time
// - user_agent: Client user agent
// - error: Error message (if any)
//
// Example:
//
//	router.Use(middleware.LoggerMiddleware(logger))
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Get request ID from context
		requestID, _ := c.Get("request_id")
		reqID, _ := requestID.(string)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", reqID),
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", userAgent),
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

// GetRequestID retrieves the request ID from the Gin context.
//
// Parameters:
//   - c: Gin context
//
// Returns:
//   - string: Request ID, or empty string if not found
//
// Example:
//
//	requestID := middleware.GetRequestID(c)
//	logger.Info("Processing request", zap.String("request_id", requestID))
func GetRequestID(c *gin.Context) string {
	requestID, exists := c.Get("request_id")
	if !exists {
		return ""
	}
	reqID, ok := requestID.(string)
	if !ok {
		return ""
	}
	return reqID
}

// LogWithRequestID creates a logger with the request ID field pre-populated.
//
// Parameters:
//   - logger: Base Zap logger
//   - c: Gin context
//
// Returns:
//   - *zap.Logger: Logger with request ID field
//
// Example:
//
//	reqLogger := middleware.LogWithRequestID(logger, c)
//	reqLogger.Info("Processing workflow")
func LogWithRequestID(logger *zap.Logger, c *gin.Context) *zap.Logger {
	requestID := GetRequestID(c)
	if requestID == "" {
		return logger
	}
	return logger.With(zap.String("request_id", requestID))
}
