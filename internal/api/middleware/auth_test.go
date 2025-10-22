package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/goflow-atom/goflow-service/internal/api/middleware"
	"github.com/goflow-atom/goflow-service/internal/config"
	"github.com/goflow-atom/goflow-service/internal/domain"
)

// Test configuration
var (
	testSecret   = "test-secret-key-for-jwt-signing"
	testIssuer   = "goflow-test"
	testAudience = "goflow-api-test"
)

// Helper function to create a test JWT token
func createTestToken(secret string, userID, email string, roles []string, expiresIn time.Duration) (string, error) {
	claims := &middleware.JWTClaims{
		UserID:      userID,
		Email:       email,
		Name:        "Test User",
		Roles:       roles,
		Permissions: []string{"workflow:read", "workflow:create"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testIssuer,
			Audience:  jwt.ClaimStrings{testAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Helper function to create test authenticator
func createTestAuthenticator() middleware.Authenticator {
	authConfig := config.AuthConfig{
		JWTSecret:          testSecret,
		JWTExpirationHours: 24,
		JWTIssuer:          testIssuer,
		JWTAudience:        testAudience,
	}

	logger, _ := zap.NewDevelopment()
	return middleware.NewJWTAuthenticator(authConfig, logger)
}

// TestJWTAuthenticator_ValidateToken_Success tests successful token validation
func TestJWTAuthenticator_ValidateToken_Success(t *testing.T) {
	auth := createTestAuthenticator()

	// Create a valid token
	tokenString, err := createTestToken(testSecret, "user123", "test@example.com", []string{"admin"}, 1*time.Hour)
	require.NoError(t, err)

	// Validate the token
	claims, err := auth.ValidateToken(tokenString)
	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "Test User", claims.Name)
	assert.Equal(t, []string{"admin"}, claims.Roles)
	assert.Equal(t, testIssuer, claims.Issuer)
	assert.Contains(t, claims.Audience, testAudience)
}

// TestJWTAuthenticator_ValidateToken_ExpiredToken tests expired token validation
func TestJWTAuthenticator_ValidateToken_ExpiredToken(t *testing.T) {
	auth := createTestAuthenticator()

	// Create an expired token
	tokenString, err := createTestToken(testSecret, "user123", "test@example.com", []string{"admin"}, -1*time.Hour)
	require.NoError(t, err)

	// Validate the token - should fail
	claims, err := auth.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is expired")
}

// TestJWTAuthenticator_ValidateToken_InvalidSignature tests invalid signature
func TestJWTAuthenticator_ValidateToken_InvalidSignature(t *testing.T) {
	auth := createTestAuthenticator()

	// Create a token with different secret
	tokenString, err := createTestToken("wrong-secret", "user123", "test@example.com", []string{"admin"}, 1*time.Hour)
	require.NoError(t, err)

	// Validate the token - should fail
	claims, err := auth.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestJWTAuthenticator_ValidateToken_MalformedToken tests malformed token
func TestJWTAuthenticator_ValidateToken_MalformedToken(t *testing.T) {
	auth := createTestAuthenticator()

	// Test with malformed token
	claims, err := auth.ValidateToken("not-a-valid-jwt-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestJWTAuthenticator_ValidateToken_InvalidIssuer tests invalid issuer
func TestJWTAuthenticator_ValidateToken_InvalidIssuer(t *testing.T) {
	auth := createTestAuthenticator()

	// Create a token with wrong issuer
	claims := &middleware.JWTClaims{
		UserID: "user123",
		Email:  "test@example.com",
		Roles:  []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wrong-issuer",
			Audience:  jwt.ClaimStrings{testAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	// Validate the token - should fail
	validatedClaims, err := auth.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
	assert.Contains(t, err.Error(), "invalid token issuer")
}

// TestJWTAuthenticator_ExtractUserContext_Success tests successful user context extraction
func TestJWTAuthenticator_ExtractUserContext_Success(t *testing.T) {
	auth := createTestAuthenticator()

	claims := &middleware.JWTClaims{
		UserID:      "user123",
		Email:       "test@example.com",
		Name:        "Test User",
		Roles:       []string{"admin", "developer"},
		Permissions: []string{"workflow:read", "workflow:create"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testIssuer,
			Audience:  jwt.ClaimStrings{testAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	userContext, err := auth.ExtractUserContext(claims)
	require.NoError(t, err)
	assert.NotNil(t, userContext)
	assert.Equal(t, "user123", userContext.UserID)
	assert.Equal(t, "test@example.com", userContext.Email)
	assert.Equal(t, "Test User", userContext.Name)
	assert.Equal(t, []string{"admin", "developer"}, userContext.Roles)
	assert.Equal(t, []string{"workflow:read", "workflow:create"}, userContext.Permissions)
	assert.Equal(t, testIssuer, userContext.Issuer)
	assert.Equal(t, testAudience, userContext.Audience)
}

// TestJWTAuthenticator_ExtractUserContext_MissingUserID tests missing user ID
func TestJWTAuthenticator_ExtractUserContext_MissingUserID(t *testing.T) {
	auth := createTestAuthenticator()

	claims := &middleware.JWTClaims{
		UserID: "", // Missing user ID
		Email:  "test@example.com",
		Roles:  []string{"admin"},
	}

	userContext, err := auth.ExtractUserContext(claims)
	assert.Error(t, err)
	assert.Nil(t, userContext)
	assert.Contains(t, err.Error(), "user_id claim is required")
}

// TestJWTAuthenticator_ExtractUserContext_MissingEmail tests missing email
func TestJWTAuthenticator_ExtractUserContext_MissingEmail(t *testing.T) {
	auth := createTestAuthenticator()

	claims := &middleware.JWTClaims{
		UserID: "user123",
		Email:  "", // Missing email
		Roles:  []string{"admin"},
	}

	userContext, err := auth.ExtractUserContext(claims)
	assert.Error(t, err)
	assert.Nil(t, userContext)
	assert.Contains(t, err.Error(), "email claim is required")
}

// TestAuthMiddleware_ValidJWT tests the middleware with a valid JWT
func TestAuthMiddleware_ValidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := createTestAuthenticator()
	router := gin.New()
	router.Use(auth.GetMiddleware())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		userCtx := middleware.GetUserContext(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userCtx.UserID,
			"email":   userCtx.Email,
		})
	})

	// Create a valid token
	tokenString, err := createTestToken(testSecret, "user123", "test@example.com", []string{"admin"}, 1*time.Hour)
	require.NoError(t, err)

	// Make request with valid token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user123")
	assert.Contains(t, w.Body.String(), "test@example.com")
}

// TestAuthMiddleware_NoJWTProvided tests the middleware without JWT
func TestAuthMiddleware_NoJWTProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := createTestAuthenticator()
	router := gin.New()
	router.Use(auth.GetMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authorization header")
}

// TestAuthMiddleware_InvalidJWT tests the middleware with invalid JWT
func TestAuthMiddleware_InvalidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := createTestAuthenticator()
	router := gin.New()
	router.Use(auth.GetMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request with invalid token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// TestAuthMiddleware_ExpiredJWT tests the middleware with expired JWT
func TestAuthMiddleware_ExpiredJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := createTestAuthenticator()
	router := gin.New()
	router.Use(auth.GetMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create an expired token
	tokenString, err := createTestToken(testSecret, "user123", "test@example.com", []string{"admin"}, -1*time.Hour)
	require.NoError(t, err)

	// Make request with expired token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// TestAuthMiddleware_InvalidHeaderFormat tests invalid authorization header format
func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := createTestAuthenticator()
	router := gin.New()
	router.Use(auth.GetMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "missing Bearer prefix",
			header: "some-token",
		},
		{
			name:   "wrong prefix",
			header: "Basic some-token",
		},
		{
			name:   "only Bearer",
			header: "Bearer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid authorization header format")
		})
	}
}

// TestGetUserContext tests the GetUserContext helper function
func TestGetUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test with user context set
	t.Run("with user context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		expectedUserCtx := &domain.UserContext{
			UserID: "user123",
			Email:  "test@example.com",
			Roles:  []string{"admin"},
		}
		c.Set("user_context", expectedUserCtx)

		userCtx := middleware.GetUserContext(c)
		assert.NotNil(t, userCtx)
		assert.Equal(t, expectedUserCtx.UserID, userCtx.UserID)
		assert.Equal(t, expectedUserCtx.Email, userCtx.Email)
	})

	// Test without user context
	t.Run("without user context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userCtx := middleware.GetUserContext(c)
		assert.Nil(t, userCtx)
	})

	// Test with wrong type
	t.Run("with wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_context", "not-a-user-context")

		userCtx := middleware.GetUserContext(c)
		assert.Nil(t, userCtx)
	})
}

// TestRequireRole tests the RequireRole middleware
func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRoles      []string
		requiredRoles  []string
		expectedStatus int
		shouldPass     bool
	}{
		{
			name:           "user has required role",
			userRoles:      []string{"admin"},
			requiredRoles:  []string{"admin"},
			expectedStatus: http.StatusOK,
			shouldPass:     true,
		},
		{
			name:           "user has one of required roles",
			userRoles:      []string{"developer"},
			requiredRoles:  []string{"admin", "developer"},
			expectedStatus: http.StatusOK,
			shouldPass:     true,
		},
		{
			name:           "user does not have required role",
			userRoles:      []string{"viewer"},
			requiredRoles:  []string{"admin"},
			expectedStatus: http.StatusForbidden,
			shouldPass:     false,
		},
		{
			name:           "user has no roles",
			userRoles:      []string{},
			requiredRoles:  []string{"admin"},
			expectedStatus: http.StatusForbidden,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Set user context
				userCtx := &domain.UserContext{
					UserID: "user123",
					Email:  "test@example.com",
					Roles:  tt.userRoles,
				}
				c.Set("user_context", userCtx)
				c.Next()
			})
			router.Use(middleware.RequireRole(tt.requiredRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.shouldPass {
				assert.Contains(t, w.Body.String(), "success")
			} else {
				assert.Contains(t, w.Body.String(), "FORBIDDEN")
			}
		})
	}
}

// TestRequireRole_NoUserContext tests RequireRole without user context
func TestRequireRole_NoUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication required")
}

// TestRequirePermission tests the RequirePermission middleware
func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                string
		userPermissions     []string
		requiredPermissions []string
		expectedStatus      int
		shouldPass          bool
	}{
		{
			name:                "user has required permission",
			userPermissions:     []string{"workflow:create"},
			requiredPermissions: []string{"workflow:create"},
			expectedStatus:      http.StatusOK,
			shouldPass:          true,
		},
		{
			name:                "user has one of required permissions",
			userPermissions:     []string{"workflow:read"},
			requiredPermissions: []string{"workflow:create", "workflow:read"},
			expectedStatus:      http.StatusOK,
			shouldPass:          true,
		},
		{
			name:                "user does not have required permission",
			userPermissions:     []string{"workflow:read"},
			requiredPermissions: []string{"workflow:delete"},
			expectedStatus:      http.StatusForbidden,
			shouldPass:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Set user context
				userCtx := &domain.UserContext{
					UserID:      "user123",
					Email:       "test@example.com",
					Permissions: tt.userPermissions,
				}
				c.Set("user_context", userCtx)
				c.Next()
			})
			router.Use(middleware.RequirePermission(tt.requiredPermissions...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.shouldPass {
				assert.Contains(t, w.Body.String(), "success")
			} else {
				assert.Contains(t, w.Body.String(), "FORBIDDEN")
			}
		})
	}
}

// TestUserContext_HasRole tests the HasRole method
func TestUserContext_HasRole(t *testing.T) {
	userCtx := &domain.UserContext{
		Roles: []string{"admin", "developer"},
	}

	assert.True(t, userCtx.HasRole("admin"))
	assert.True(t, userCtx.HasRole("developer"))
	assert.False(t, userCtx.HasRole("viewer"))
}

// TestUserContext_HasPermission tests the HasPermission method
func TestUserContext_HasPermission(t *testing.T) {
	userCtx := &domain.UserContext{
		Permissions: []string{"workflow:create", "workflow:read"},
	}

	assert.True(t, userCtx.HasPermission("workflow:create"))
	assert.True(t, userCtx.HasPermission("workflow:read"))
	assert.False(t, userCtx.HasPermission("workflow:delete"))
}

// TestUserContext_HasAnyRole tests the HasAnyRole method
func TestUserContext_HasAnyRole(t *testing.T) {
	userCtx := &domain.UserContext{
		Roles: []string{"developer"},
	}

	assert.True(t, userCtx.HasAnyRole("admin", "developer"))
	assert.True(t, userCtx.HasAnyRole("developer"))
	assert.False(t, userCtx.HasAnyRole("admin", "viewer"))
}

// TestUserContext_IsExpired tests the IsExpired method
func TestUserContext_IsExpired(t *testing.T) {
	// Not expired
	userCtx := &domain.UserContext{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	assert.False(t, userCtx.IsExpired())

	// Expired
	userCtx.ExpiresAt = time.Now().Add(-1 * time.Hour)
	assert.True(t, userCtx.IsExpired())
}

// TestUserContext_IsAdmin tests the IsAdmin method
func TestUserContext_IsAdmin(t *testing.T) {
	// Is admin
	userCtx := &domain.UserContext{
		Roles: []string{"admin"},
	}
	assert.True(t, userCtx.IsAdmin())

	// Not admin
	userCtx.Roles = []string{"developer"}
	assert.False(t, userCtx.IsAdmin())
}

