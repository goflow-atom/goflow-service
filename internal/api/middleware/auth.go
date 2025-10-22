package middleware

import (
	"net/http"
	"strings"

	"github.com/goflow-atom/goflow-service/internal/api/dto"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens for protected routes
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Missing authorization header",
				},
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]

		// TODO: Implement JWT validation
		// For now, just check if token is not empty
		if token == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		// TODO: Extract user information from token and set in context
		// c.Set("user_id", userID)
		// c.Set("user_email", userEmail)

		c.Next()
	}
}
