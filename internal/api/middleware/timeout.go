// Package middleware provides HTTP middleware for the GoFlow API.
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goflow-atom/goflow-service/internal/api/dto"
)

// TimeoutMiddleware creates a middleware that enforces a timeout on requests.
//
// Parameters:
//   - timeout: Maximum duration for request processing
//
// Returns:
//   - gin.HandlerFunc: Timeout middleware
//
// Example:
//
//	router.Use(middleware.TimeoutMiddleware(30 * time.Second))
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal completion
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, dto.ErrorResponse{
					Error: dto.ErrorDetail{
						Code:    "REQUEST_TIMEOUT",
						Message: "Request processing timeout exceeded",
					},
				})
				c.Abort()
			}
		}
	}
}
