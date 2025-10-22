// Package middleware provides HTTP middleware for the GoFlow API.
//
// This package includes authorization middleware for role-based access control (RBAC)
// and resource-based authorization.
package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

// Authorizer defines the interface for authorization operations.
//
// This interface provides methods for checking user permissions and roles
// against resources and actions.
type Authorizer interface {
	// Authorize checks if the user has permission to perform an action on a resource
	Authorize(ctx context.Context, resource string, action string) error

	// HasRole checks if the user has a specific role
	HasRole(ctx context.Context, role string) bool

	// GetAllowedActions returns the list of actions the user can perform on a resource
	GetAllowedActions(ctx context.Context, resource string) ([]string, error)
}

// AuthorizationPolicy defines the authorization policy for a resource.
//
// A policy maps actions to the roles that are allowed to perform those actions.
// For example:
//
//	{
//	  "workflow": {
//	    "create": ["admin", "developer"],
//	    "read":   ["admin", "developer", "operator", "viewer"],
//	    "update": ["admin", "developer"],
//	    "delete": ["admin"]
//	  }
//	}
type AuthorizationPolicy struct {
	// Resource is the name of the resource (e.g., "workflow", "execution")
	Resource string `json:"resource"`

	// Actions maps action names to the roles allowed to perform them
	Actions map[string][]string `json:"actions"`
}

// PolicyManager manages authorization policies.
type PolicyManager struct {
	// policies maps resource names to their authorization policies
	policies map[string]*AuthorizationPolicy

	// defaultDeny determines whether to deny access by default when no policy is found
	defaultDeny bool

	// logger for structured logging
	logger *zap.Logger
}

// NewPolicyManager creates a new policy manager.
//
// Parameters:
//   - defaultDeny: If true, deny access when no policy is found; if false, allow access
//   - logger: Zap logger instance
//
// Returns:
//   - *PolicyManager: New policy manager instance
//
// Example:
//
//	pm := middleware.NewPolicyManager(true, logger)
//	pm.AddPolicy("workflow", map[string][]string{
//	    "create": {"admin", "developer"},
//	    "read":   {"admin", "developer", "operator", "viewer"},
//	})
func NewPolicyManager(defaultDeny bool, logger *zap.Logger) *PolicyManager {
	return &PolicyManager{
		policies:    make(map[string]*AuthorizationPolicy),
		defaultDeny: defaultDeny,
		logger:      logger,
	}
}

// AddPolicy adds an authorization policy for a resource.
//
// Parameters:
//   - resource: The resource name
//   - actions: Map of action names to allowed roles
//
// Example:
//
//	pm.AddPolicy("workflow", map[string][]string{
//	    "create": {"admin", "developer"},
//	    "read":   {"admin", "developer", "operator", "viewer"},
//	    "update": {"admin", "developer"},
//	    "delete": {"admin"},
//	})
func (pm *PolicyManager) AddPolicy(resource string, actions map[string][]string) {
	pm.policies[resource] = &AuthorizationPolicy{
		Resource: resource,
		Actions:  actions,
	}
}

// GetPolicy retrieves the authorization policy for a resource.
//
// Parameters:
//   - resource: The resource name
//
// Returns:
//   - *AuthorizationPolicy: The policy, or nil if not found
func (pm *PolicyManager) GetPolicy(resource string) *AuthorizationPolicy {
	return pm.policies[resource]
}

// Authorize checks if the user has permission to perform an action on a resource.
//
// This method retrieves the user context from the Gin context, looks up the
// authorization policy for the resource, and checks if the user's roles are
// allowed to perform the action.
//
// Parameters:
//   - ctx: Context containing user information
//   - resource: The resource name (e.g., "workflow", "execution")
//   - action: The action name (e.g., "create", "read", "update", "delete")
//
// Returns:
//   - error: Error if authorization fails, nil if authorized
//
// Example:
//
//	err := pm.Authorize(c.Request.Context(), "workflow", "create")
//	if err != nil {
//	    // User is not authorized
//	}
func (pm *PolicyManager) Authorize(ctx context.Context, resource string, action string) error {
	// Extract user context from Gin context
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	userContext := GetUserContext(ginCtx)
	if userContext == nil {
		return fmt.Errorf("user context not found")
	}

	// Get policy for resource
	policy := pm.GetPolicy(resource)
	if policy == nil {
		if pm.defaultDeny {
			pm.logger.Warn("No policy found for resource, denying access",
				zap.String("resource", resource),
				zap.String("action", action),
				zap.String("user_id", userContext.UserID),
			)
			return fmt.Errorf("no authorization policy found for resource: %s", resource)
		}
		// No policy and default allow
		return nil
	}

	// Get allowed roles for action
	allowedRoles, exists := policy.Actions[action]
	if !exists {
		if pm.defaultDeny {
			pm.logger.Warn("No policy found for action, denying access",
				zap.String("resource", resource),
				zap.String("action", action),
				zap.String("user_id", userContext.UserID),
			)
			return fmt.Errorf("no authorization policy found for action: %s on resource: %s", action, resource)
		}
		// No action policy and default allow
		return nil
	}

	// Check if user has any of the allowed roles
	for _, userRole := range userContext.Roles {
		if slices.Contains(allowedRoles, userRole) {
			pm.logger.Debug("Authorization granted",
				zap.String("resource", resource),
				zap.String("action", action),
				zap.String("user_id", userContext.UserID),
				zap.String("role", userRole),
			)
			return nil
		}
	}

	// User does not have required role
	pm.logger.Warn("Authorization denied",
		zap.String("resource", resource),
		zap.String("action", action),
		zap.String("user_id", userContext.UserID),
		zap.Strings("user_roles", userContext.Roles),
		zap.Strings("required_roles", allowedRoles),
	)

	return fmt.Errorf("user does not have required role for action: %s on resource: %s", action, resource)
}

// HasRole checks if the user has a specific role.
//
// Parameters:
//   - ctx: Context containing user information
//   - role: The role to check for
//
// Returns:
//   - bool: true if the user has the role, false otherwise
func (pm *PolicyManager) HasRole(ctx context.Context, role string) bool {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return false
	}

	userContext := GetUserContext(ginCtx)
	if userContext == nil {
		return false
	}

	return userContext.HasRole(role)
}

// GetAllowedActions returns the list of actions the user can perform on a resource.
//
// Parameters:
//   - ctx: Context containing user information
//   - resource: The resource name
//
// Returns:
//   - []string: List of allowed actions
//   - error: Error if user context is not found
func (pm *PolicyManager) GetAllowedActions(ctx context.Context, resource string) ([]string, error) {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	userContext := GetUserContext(ginCtx)
	if userContext == nil {
		return nil, fmt.Errorf("user context not found")
	}

	policy := pm.GetPolicy(resource)
	if policy == nil {
		if pm.defaultDeny {
			return []string{}, nil
		}
		// No policy and default allow - return all actions
		return []string{"*"}, nil
	}

	allowedActions := make([]string, 0)
	for action, allowedRoles := range policy.Actions {
		for _, userRole := range userContext.Roles {
			if slices.Contains(allowedRoles, userRole) {
				allowedActions = append(allowedActions, action)
				break
			}
		}
	}

	return allowedActions, nil
}

// RequireResourceAction returns a middleware that checks if the user can perform
// an action on a resource using the policy manager.
//
// This middleware should be used after the authentication middleware to enforce
// resource-based access control.
//
// Parameters:
//   - pm: Policy manager instance
//   - resource: The resource name
//   - action: The action name
//
// Returns:
//   - gin.HandlerFunc: Middleware handler function
//
// Example:
//
//	pm := middleware.NewPolicyManager(true, logger)
//	pm.AddPolicy("workflow", map[string][]string{
//	    "create": {"admin", "developer"},
//	})
//	router.POST("/workflows",
//	    middleware.RequireResourceAction(pm, "workflow", "create"),
//	    handler.CreateWorkflow)
func RequireResourceAction(pm *PolicyManager, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := pm.Authorize(c, resource, action)
		if err != nil {
			HandleError(c, http.StatusForbidden, "FORBIDDEN",
				fmt.Sprintf("You do not have permission to %s %s", action, resource))
			c.Abort()
			return
		}

		c.Next()
	}
}

