package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/goflow-atom/goflow-service/internal/domain"
)

// setupTestRouter creates a test Gin router with the given middleware.
func setupTestRouter(middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware...)
	return router
}

// createTestUserContext creates a test user context with the given roles.
func createTestUserContext(roles ...string) *domain.UserContext {
	return &domain.UserContext{
		UserID:      "test-user-123",
		Email:       "test@example.com",
		Name:        "Test User",
		Roles:       roles,
		Permissions: []string{},
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Issuer:      "goflow-test",
		Audience:    "goflow-api",
	}
}

// setupTestContext creates a test Gin context with user context attached.
func setupTestContext(userContext *domain.UserContext) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	if userContext != nil {
		c.Set("user_context", userContext)
		c.Set("user_id", userContext.UserID)
		c.Set("user_email", userContext.Email)
		c.Set("user_roles", userContext.Roles)
	}

	return c, w
}

func TestNewPolicyManager(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name        string
		defaultDeny bool
	}{
		{
			name:        "default deny enabled",
			defaultDeny: true,
		},
		{
			name:        "default deny disabled",
			defaultDeny: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPolicyManager(tt.defaultDeny, logger)

			require.NotNil(t, pm)
			assert.Equal(t, tt.defaultDeny, pm.defaultDeny)
			assert.NotNil(t, pm.policies)
			assert.NotNil(t, pm.logger)
		})
	}
}

func TestPolicyManager_AddPolicy(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	actions := map[string][]string{
		"create": {"admin", "developer"},
		"read":   {"admin", "developer", "operator", "viewer"},
		"update": {"admin", "developer"},
		"delete": {"admin"},
	}

	pm.AddPolicy("workflow", actions)

	policy := pm.GetPolicy("workflow")
	require.NotNil(t, policy)
	assert.Equal(t, "workflow", policy.Resource)
	assert.Equal(t, actions, policy.Actions)
}

func TestPolicyManager_GetPolicy(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	actions := map[string][]string{
		"create": {"admin"},
	}
	pm.AddPolicy("workflow", actions)

	tests := []struct {
		name     string
		resource string
		wantNil  bool
	}{
		{
			name:     "existing policy",
			resource: "workflow",
			wantNil:  false,
		},
		{
			name:     "non-existing policy",
			resource: "execution",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := pm.GetPolicy(tt.resource)
			if tt.wantNil {
				assert.Nil(t, policy)
			} else {
				assert.NotNil(t, policy)
			}
		})
	}
}

func TestPolicyManager_Authorize_AllowedRoleAccess(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin", "developer"},
		"read":   {"admin", "developer", "operator", "viewer"},
	})

	tests := []struct {
		name     string
		roles    []string
		resource string
		action   string
		wantErr  bool
	}{
		{
			name:     "admin can create workflow",
			roles:    []string{"admin"},
			resource: "workflow",
			action:   "create",
			wantErr:  false,
		},
		{
			name:     "developer can create workflow",
			roles:    []string{"developer"},
			resource: "workflow",
			action:   "create",
			wantErr:  false,
		},
		{
			name:     "viewer can read workflow",
			roles:    []string{"viewer"},
			resource: "workflow",
			action:   "read",
			wantErr:  false,
		},
		{
			name:     "user with multiple roles can access",
			roles:    []string{"viewer", "developer"},
			resource: "workflow",
			action:   "create",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userContext := createTestUserContext(tt.roles...)
			c, _ := setupTestContext(userContext)

			err := pm.Authorize(c, tt.resource, tt.action)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPolicyManager_Authorize_ForbiddenRoleAccess(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin", "developer"},
		"delete": {"admin"},
	})

	tests := []struct {
		name     string
		roles    []string
		resource string
		action   string
	}{
		{
			name:     "viewer cannot create workflow",
			roles:    []string{"viewer"},
			resource: "workflow",
			action:   "create",
		},
		{
			name:     "operator cannot create workflow",
			roles:    []string{"operator"},
			resource: "workflow",
			action:   "create",
		},
		{
			name:     "developer cannot delete workflow",
			roles:    []string{"developer"},
			resource: "workflow",
			action:   "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userContext := createTestUserContext(tt.roles...)
			c, _ := setupTestContext(userContext)

			err := pm.Authorize(c, tt.resource, tt.action)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "does not have required role")
		})
	}
}

func TestPolicyManager_Authorize_NoRoleUnauthorized(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin", "developer"},
	})

	userContext := createTestUserContext() // No roles
	c, _ := setupTestContext(userContext)

	err := pm.Authorize(c, "workflow", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not have required role")
}

func TestPolicyManager_Authorize_MissingUserContext(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin"},
	})

	c, _ := setupTestContext(nil) // No user context

	err := pm.Authorize(c, "workflow", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user context not found")
}

func TestPolicyManager_Authorize_NoPolicyDefaultDeny(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger) // Default deny

	userContext := createTestUserContext("admin")
	c, _ := setupTestContext(userContext)

	err := pm.Authorize(c, "nonexistent", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authorization policy found")
}

func TestPolicyManager_Authorize_NoPolicyDefaultAllow(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(false, logger) // Default allow

	userContext := createTestUserContext("admin")
	c, _ := setupTestContext(userContext)

	err := pm.Authorize(c, "nonexistent", "create")

	assert.NoError(t, err)
}

func TestPolicyManager_Authorize_NoActionPolicyDefaultDeny(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin"},
	})

	userContext := createTestUserContext("admin")
	c, _ := setupTestContext(userContext)

	err := pm.Authorize(c, "workflow", "nonexistent_action")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authorization policy found for action")
}

func TestPolicyManager_HasRole(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	tests := []struct {
		name     string
		roles    []string
		checkRole string
		want     bool
	}{
		{
			name:      "user has role",
			roles:     []string{"admin", "developer"},
			checkRole: "admin",
			want:      true,
		},
		{
			name:      "user does not have role",
			roles:     []string{"viewer"},
			checkRole: "admin",
			want:      false,
		},
		{
			name:      "user has no roles",
			roles:     []string{},
			checkRole: "admin",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userContext := createTestUserContext(tt.roles...)
			c, _ := setupTestContext(userContext)

			result := pm.HasRole(c, tt.checkRole)

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestPolicyManager_HasRole_NoUserContext(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	c, _ := setupTestContext(nil)

	result := pm.HasRole(c, "admin")

	assert.False(t, result)
}

func TestPolicyManager_GetAllowedActions(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin", "developer"},
		"read":   {"admin", "developer", "operator", "viewer"},
		"update": {"admin", "developer"},
		"delete": {"admin"},
	})

	tests := []struct {
		name     string
		roles    []string
		resource string
		want     []string
	}{
		{
			name:     "admin has all actions",
			roles:    []string{"admin"},
			resource: "workflow",
			want:     []string{"create", "read", "update", "delete"},
		},
		{
			name:     "developer has create, read, update",
			roles:    []string{"developer"},
			resource: "workflow",
			want:     []string{"create", "read", "update"},
		},
		{
			name:     "viewer has only read",
			roles:    []string{"viewer"},
			resource: "workflow",
			want:     []string{"read"},
		},
		{
			name:     "operator has only read",
			roles:    []string{"operator"},
			resource: "workflow",
			want:     []string{"read"},
		},
		{
			name:     "user with no matching roles has no actions",
			roles:    []string{"unknown"},
			resource: "workflow",
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userContext := createTestUserContext(tt.roles...)
			c, _ := setupTestContext(userContext)

			actions, err := pm.GetAllowedActions(c, tt.resource)

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, actions)
		})
	}
}

func TestPolicyManager_GetAllowedActions_NoPolicy(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name        string
		defaultDeny bool
		want        []string
	}{
		{
			name:        "default deny returns empty list",
			defaultDeny: true,
			want:        []string{},
		},
		{
			name:        "default allow returns wildcard",
			defaultDeny: false,
			want:        []string{"*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPolicyManager(tt.defaultDeny, logger)
			userContext := createTestUserContext("admin")
			c, _ := setupTestContext(userContext)

			actions, err := pm.GetAllowedActions(c, "nonexistent")

			require.NoError(t, err)
			assert.Equal(t, tt.want, actions)
		})
	}
}

func TestPolicyManager_GetAllowedActions_NoUserContext(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	c, _ := setupTestContext(nil)

	actions, err := pm.GetAllowedActions(c, "workflow")

	assert.Error(t, err)
	assert.Nil(t, actions)
	assert.Contains(t, err.Error(), "user context not found")
}

func TestRequireResourceAction_Allowed(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin", "developer"},
	})

	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		// Attach user context
		userContext := createTestUserContext("admin")
		c.Set("user_context", userContext)
		c.Set("user_id", userContext.UserID)
		c.Set("user_email", userContext.Email)
		c.Set("user_roles", userContext.Roles)
	}, RequireResourceAction(pm, "workflow", "create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireResourceAction_Forbidden(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin", "developer"},
	})

	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		// Attach user context with viewer role
		userContext := createTestUserContext("viewer")
		c.Set("user_context", userContext)
		c.Set("user_id", userContext.UserID)
		c.Set("user_email", userContext.Email)
		c.Set("user_roles", userContext.Roles)
	}, RequireResourceAction(pm, "workflow", "create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireResourceAction_NoUserContext(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin"},
	})

	router := setupTestRouter()
	router.GET("/test", RequireResourceAction(pm, "workflow", "create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireResourceAction_MultipleRoles(t *testing.T) {
	logger := zap.NewNop()
	pm := NewPolicyManager(true, logger)

	pm.AddPolicy("workflow", map[string][]string{
		"create": {"admin", "developer"},
		"delete": {"admin"},
	})

	tests := []struct {
		name       string
		roles      []string
		action     string
		wantStatus int
	}{
		{
			name:       "user with developer role can create",
			roles:      []string{"developer", "viewer"},
			action:     "create",
			wantStatus: http.StatusOK,
		},
		{
			name:       "user with developer role cannot delete",
			roles:      []string{"developer", "viewer"},
			action:     "delete",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "user with admin role can delete",
			roles:      []string{"admin"},
			action:     "delete",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/test", func(c *gin.Context) {
				userContext := createTestUserContext(tt.roles...)
				c.Set("user_context", userContext)
				c.Set("user_id", userContext.UserID)
				c.Set("user_email", userContext.Email)
				c.Set("user_roles", userContext.Roles)
			}, RequireResourceAction(pm, "workflow", tt.action), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

