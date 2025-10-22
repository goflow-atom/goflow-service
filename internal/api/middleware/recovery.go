// Package middleware provides HTTP middleware for the GoFlow API.
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/goflow-atom/goflow-service/internal/api/dto"
	"go.uber.org/zap"
)

// ErrorHandler is a centralized error handler middleware that catches panics
// and errors, returning consistent JSON error responses.
//
// This middleware:
// - Catches panics and converts them to 500 Internal Server Error responses
// - Logs all errors with stack traces
// - Returns standardized ErrorResponse format
// - Includes request ID in error logs for tracing
//
// Example:
//
//	router.Use(middleware.ErrorHandler(logger))
func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID from context
				requestID, _ := c.Get("request_id")
				reqID, _ := requestID.(string)

				// Log the panic with stack trace
				logger.Error("Panic recovered",
					zap.String("request_id", reqID),
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// Return standardized error response
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					Error: dto.ErrorDetail{
						Code:    "INTERNAL_ERROR",
						Message: "An internal server error occurred",
						Details: map[string]interface{}{
							"request_id": reqID,
						},
					},
				})

				// Abort the request
				c.Abort()
			}
		}()

		// Process request
		c.Next()

		// Check if there are any errors in the context
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Get request ID from context
			requestID, _ := c.Get("request_id")
			reqID, _ := requestID.(string)

			// Log the error
			logger.Error("Request error",
				zap.String("request_id", reqID),
				zap.Error(err.Err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			// If response hasn't been written yet, return error response
			if !c.Writer.Written() {
				statusCode := http.StatusInternalServerError
				errorCode := "INTERNAL_ERROR"
				errorMessage := err.Error()

				// Try to determine appropriate status code from error type
				if err.Type == gin.ErrorTypeBind {
					statusCode = http.StatusBadRequest
					errorCode = "VALIDATION_ERROR"
				}

				c.JSON(statusCode, dto.ErrorResponse{
					Error: dto.ErrorDetail{
						Code:    errorCode,
						Message: errorMessage,
						Details: map[string]interface{}{
							"request_id": reqID,
						},
					},
				})
			}
		}
	}
}

// HandleError is a helper function to handle errors in handlers.
// It sets the appropriate HTTP status code and error response.
//
// Parameters:
//   - c: Gin context
//   - statusCode: HTTP status code
//   - errorCode: Application error code
//   - message: Error message
//
// Example:
//
//	if err != nil {
//	    middleware.HandleError(c, http.StatusNotFound, "NOT_FOUND", "Workflow not found")
//	    return
//	}
func HandleError(c *gin.Context, statusCode int, errorCode, message string) {
	requestID, _ := c.Get("request_id")
	reqID, _ := requestID.(string)

	c.JSON(statusCode, dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    errorCode,
			Message: message,
			Details: map[string]interface{}{
				"request_id": reqID,
			},
		},
	})
}

// HandleErrorWithDetails is a helper function to handle errors with additional details.
//
// Parameters:
//   - c: Gin context
//   - statusCode: HTTP status code
//   - errorCode: Application error code
//   - message: Error message
//   - details: Additional error details
//
// Example:
//
//	middleware.HandleErrorWithDetails(c, http.StatusBadRequest, "VALIDATION_ERROR",
//	    "Invalid input", map[string]interface{}{"field": "email", "issue": "invalid format"})
func HandleErrorWithDetails(c *gin.Context, statusCode int, errorCode, message string, details map[string]interface{}) {
	requestID, _ := c.Get("request_id")
	reqID, _ := requestID.(string)

	if details == nil {
		details = make(map[string]interface{})
	}
	details["request_id"] = reqID

	c.JSON(statusCode, dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	})
}

// MapErrorToHTTP maps common error types to HTTP status codes and error codes.
//
// Parameters:
//   - err: The error to map
//
// Returns:
//   - statusCode: HTTP status code
//   - errorCode: Application error code
//   - message: Error message
func MapErrorToHTTP(err error) (statusCode int, errorCode, message string) {
	if err == nil {
		return http.StatusOK, "", ""
	}

	// Default to internal server error
	statusCode = http.StatusInternalServerError
	errorCode = "INTERNAL_ERROR"
	message = err.Error()

	// Check error message for common patterns
	errMsg := err.Error()
	switch {
	case contains(errMsg, "not found"):
		statusCode = http.StatusNotFound
		errorCode = "NOT_FOUND"
	case contains(errMsg, "already exists"), contains(errMsg, "duplicate"):
		statusCode = http.StatusConflict
		errorCode = "CONFLICT"
	case contains(errMsg, "invalid"), contains(errMsg, "validation"):
		statusCode = http.StatusBadRequest
		errorCode = "VALIDATION_ERROR"
	case contains(errMsg, "unauthorized"), contains(errMsg, "authentication"):
		statusCode = http.StatusUnauthorized
		errorCode = "UNAUTHORIZED"
	case contains(errMsg, "forbidden"), contains(errMsg, "permission"):
		statusCode = http.StatusForbidden
		errorCode = "FORBIDDEN"
	}

	return statusCode, errorCode, message
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if matchAt(s, substr, i) {
			return true
		}
	}
	return false
}

func matchAt(s, substr string, pos int) bool {
	for i := 0; i < len(substr); i++ {
		if toLower(s[pos+i]) != toLower(substr[i]) {
			return false
		}
	}
	return true
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// WrapError wraps an error with additional context and returns it as a Gin error.
//
// Parameters:
//   - c: Gin context
//   - err: The error to wrap
//   - context: Additional context message
//
// Example:
//
//	if err := service.CreateWorkflow(ctx, workflow); err != nil {
//	    middleware.WrapError(c, err, "failed to create workflow")
//	    return
//	}
func WrapError(c *gin.Context, err error, context string) {
	wrappedErr := fmt.Errorf("%s: %w", context, err)
	_ = c.Error(wrappedErr)
}
