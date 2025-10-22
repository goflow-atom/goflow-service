// Package middleware provides HTTP middleware for the GoFlow API.
//
// This package includes authentication, authorization, CORS, rate limiting,
// and other cross-cutting concerns for HTTP request processing.
package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/goflow-atom/goflow-service/internal/config"
	"github.com/goflow-atom/goflow-service/internal/domain"
)

// JWTClaims represents the custom claims in the JWT token.
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Name        string   `json:"name,omitempty"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// Authenticator defines the interface for JWT authentication.
type Authenticator interface {
	// ValidateToken validates a JWT token and returns the claims
	ValidateToken(tokenString string) (*JWTClaims, error)
	// ExtractUserContext extracts user context from JWT claims
	ExtractUserContext(claims *JWTClaims) (*domain.UserContext, error)
	// GetMiddleware returns the Gin middleware handler
	GetMiddleware() gin.HandlerFunc
}

// JWTAuthenticator implements the Authenticator interface.
type JWTAuthenticator struct {
	config *config.AuthConfig
	logger *zap.Logger
}

// NewJWTAuthenticator creates a new JWT authenticator.
//
// Parameters:
//   - authConfig: Authentication configuration
//   - logger: Zap logger instance
//
// Returns:
//   - Authenticator: JWT authenticator instance
//
// Example:
//
//	auth := middleware.NewJWTAuthenticator(cfg.Auth, logger)
//	router.Use(auth.GetMiddleware())
func NewJWTAuthenticator(authConfig config.AuthConfig, logger *zap.Logger) Authenticator {
	return &JWTAuthenticator{
		config: &authConfig,
		logger: logger,
	}
}

// ValidateToken validates a JWT token and returns the claims.
//
// This method verifies the token signature, expiration, and standard claims.
//
// Parameters:
//   - tokenString: The JWT token string
//
// Returns:
//   - *JWTClaims: Parsed and validated claims
//   - error: Error if validation fails
func (a *JWTAuthenticator) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify issuer if configured
	if a.config.JWTIssuer != "" && claims.Issuer != a.config.JWTIssuer {
		return nil, fmt.Errorf("invalid token issuer: expected %s, got %s", a.config.JWTIssuer, claims.Issuer)
	}

	// Verify audience if configured
	if a.config.JWTAudience != "" {
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == a.config.JWTAudience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return nil, fmt.Errorf("invalid token audience")
		}
	}

	return claims, nil
}

// ExtractUserContext extracts user context from JWT claims.
//
// Parameters:
//   - claims: JWT claims
//
// Returns:
//   - *domain.UserContext: User context
//   - error: Error if extraction fails
func (a *JWTAuthenticator) ExtractUserContext(claims *JWTClaims) (*domain.UserContext, error) {
	if claims.UserID == "" {
		return nil, fmt.Errorf("user_id claim is required")
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("email claim is required")
	}

	userContext := &domain.UserContext{
		UserID:      claims.UserID,
		Email:       claims.Email,
		Name:        claims.Name,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		IssuedAt:    claims.IssuedAt.Time,
		ExpiresAt:   claims.ExpiresAt.Time,
		Issuer:      claims.Issuer,
	}

	// Set audience (first one if multiple)
	if len(claims.Audience) > 0 {
		userContext.Audience = claims.Audience[0]
	}

	return userContext, nil
}

// GetMiddleware returns the Gin middleware handler for JWT authentication.
//
// This middleware:
//  1. Extracts the JWT token from the Authorization header
//  2. Validates the token signature and claims
//  3. Extracts user context from the token
//  4. Attaches user context to the Gin context
//  5. Returns 401 Unauthorized for invalid or missing tokens
//
// Returns:
//   - gin.HandlerFunc: Middleware handler function
//
// Example:
//
//	auth := middleware.NewJWTAuthenticator(cfg.Auth, logger)
//	router.Use(auth.GetMiddleware())
func (a *JWTAuthenticator) GetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			a.logger.Warn("Missing authorization header",
				zap.String("request_id", GetRequestID(c)),
				zap.String("path", c.Request.URL.Path),
			)
			HandleError(c, 401, "UNAUTHORIZED", "Missing authorization header")
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			a.logger.Warn("Invalid authorization header format",
				zap.String("request_id", GetRequestID(c)),
				zap.String("path", c.Request.URL.Path),
			)
			HandleError(c, 401, "UNAUTHORIZED", "Invalid authorization header format. Expected: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			a.logger.Warn("Token validation failed",
				zap.String("request_id", GetRequestID(c)),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)
			HandleError(c, 401, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		// Extract user context
		userContext, err := a.ExtractUserContext(claims)
		if err != nil {
			a.logger.Error("Failed to extract user context",
				zap.String("request_id", GetRequestID(c)),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)
			HandleError(c, 401, "UNAUTHORIZED", "Invalid token claims")
			c.Abort()
			return
		}

		// Attach user context to Gin context
		c.Set("user_context", userContext)
		c.Set("user_id", userContext.UserID)
		c.Set("user_email", userContext.Email)
		c.Set("user_roles", userContext.Roles)

		a.logger.Debug("User authenticated",
			zap.String("request_id", GetRequestID(c)),
			zap.String("user_id", userContext.UserID),
			zap.String("email", userContext.Email),
			zap.Strings("roles", userContext.Roles),
		)

		c.Next()
	}
}

// GetUserContext retrieves the user context from the Gin context.
//
// This helper function should be used by handlers to access the authenticated
// user's information.
//
// Parameters:
//   - c: Gin context
//
// Returns:
//   - *domain.UserContext: User context, or nil if not found
//
// Example:
//
//	func (h *Handler) CreateWorkflow(c *gin.Context) {
//	    userCtx := middleware.GetUserContext(c)
//	    if userCtx == nil {
//	        // User not authenticated
//	        return
//	    }
//	    // Use userCtx.UserID, userCtx.Roles, etc.
//	}
func GetUserContext(c *gin.Context) *domain.UserContext {
	value, exists := c.Get("user_context")
	if !exists {
		return nil
	}

	userContext, ok := value.(*domain.UserContext)
	if !ok {
		return nil
	}

	return userContext
}

// RequireRole returns a middleware that checks if the user has a specific role.
//
// This middleware should be used after the authentication middleware to enforce
// role-based access control.
//
// Parameters:
//   - roles: Required roles (user must have at least one)
//
// Returns:
//   - gin.HandlerFunc: Middleware handler function
//
// Example:
//
//	// Require admin role
//	router.POST("/admin/users", middleware.RequireRole("admin"), handler.CreateUser)
//
//	// Require admin or developer role
//	router.POST("/workflows", middleware.RequireRole("admin", "developer"), handler.CreateWorkflow)
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext := GetUserContext(c)
		if userContext == nil {
			HandleError(c, 401, "UNAUTHORIZED", "Authentication required")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if userContext.HasRole(role) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			HandleError(c, 403, "FORBIDDEN", fmt.Sprintf("Required role: %v", roles))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission returns a middleware that checks if the user has a specific permission.
//
// This middleware should be used after the authentication middleware to enforce
// permission-based access control.
//
// Parameters:
//   - permissions: Required permissions (user must have at least one)
//
// Returns:
//   - gin.HandlerFunc: Middleware handler function
//
// Example:
//
//	// Require workflow:create permission
//	router.POST("/workflows", middleware.RequirePermission("workflow:create"), handler.CreateWorkflow)
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext := GetUserContext(c)
		if userContext == nil {
			HandleError(c, 401, "UNAUTHORIZED", "Authentication required")
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		hasPermission := false
		for _, permission := range permissions {
			if userContext.HasPermission(permission) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			HandleError(c, 403, "FORBIDDEN", fmt.Sprintf("Required permission: %v", permissions))
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthMiddleware is a convenience function that creates a JWT authentication middleware
// with default configuration. This is kept for backward compatibility.
//
// Deprecated: Use NewJWTAuthenticator and GetMiddleware instead for better testability.
//
// Example:
//
//	// Old way (deprecated)
//	router.Use(middleware.AuthMiddleware())
//
//	// New way (recommended)
//	auth := middleware.NewJWTAuthenticator(cfg.Auth, logger)
//	router.Use(auth.GetMiddleware())
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a placeholder for backward compatibility
		// In production, use NewJWTAuthenticator instead
		c.Next()
	}
}
