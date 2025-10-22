// Package domain provides domain models and business logic for the GoFlow Workflow Engine.
package domain

import (
	"slices"
	"time"
)

// UserContext represents the authenticated user context extracted from JWT.
// This context is attached to the Gin context after successful authentication
// and can be used by handlers to access user information.
type UserContext struct {
	// UserID is the unique identifier of the authenticated user
	UserID string `json:"user_id"`

	// Email is the user's email address
	Email string `json:"email"`

	// Name is the user's display name
	Name string `json:"name,omitempty"`

	// Roles is the list of roles assigned to the user
	// Common roles: admin, developer, operator, viewer
	Roles []string `json:"roles"`

	// Permissions is the list of specific permissions granted to the user
	Permissions []string `json:"permissions,omitempty"`

	// IssuedAt is the timestamp when the token was issued
	IssuedAt time.Time `json:"issued_at"`

	// ExpiresAt is the timestamp when the token expires
	ExpiresAt time.Time `json:"expires_at"`

	// Issuer is the JWT issuer claim
	Issuer string `json:"issuer,omitempty"`

	// Audience is the JWT audience claim
	Audience string `json:"audience,omitempty"`
}

// HasRole checks if the user has a specific role.
//
// Parameters:
//   - role: The role to check for
//
// Returns:
//   - bool: true if the user has the role, false otherwise
//
// Example:
//
//	if userCtx.HasRole("admin") {
//	    // User is an admin
//	}
func (u *UserContext) HasRole(role string) bool {
	return slices.Contains(u.Roles, role)
}

// HasPermission checks if the user has a specific permission.
//
// Parameters:
//   - permission: The permission to check for
//
// Returns:
//   - bool: true if the user has the permission, false otherwise
//
// Example:
//
//	if userCtx.HasPermission("workflow:create") {
//	    // User can create workflows
//	}
func (u *UserContext) HasPermission(permission string) bool {
	return slices.Contains(u.Permissions, permission)
}

// HasAnyRole checks if the user has any of the specified roles.
//
// Parameters:
//   - roles: The roles to check for
//
// Returns:
//   - bool: true if the user has at least one of the roles, false otherwise
//
// Example:
//
//	if userCtx.HasAnyRole("admin", "developer") {
//	    // User is either an admin or developer
//	}
func (u *UserContext) HasAnyRole(roles ...string) bool {
	return slices.ContainsFunc(roles, u.HasRole)
}

// IsExpired checks if the user's token has expired.
//
// Returns:
//   - bool: true if the token has expired, false otherwise
//
// Example:
//
//	if userCtx.IsExpired() {
//	    // Token has expired
//	}
func (u *UserContext) IsExpired() bool {
	return time.Now().After(u.ExpiresAt)
}

// IsAdmin checks if the user has the admin role.
//
// Returns:
//   - bool: true if the user is an admin, false otherwise
//
// Example:
//
//	if userCtx.IsAdmin() {
//	    // User is an admin
//	}
func (u *UserContext) IsAdmin() bool {
	return u.HasRole("admin")
}

