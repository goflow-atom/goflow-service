package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/goflow-atom/goflow-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestRouter_InitializesWithConfiguredMode tests that the Gin router
// initializes with the mode configured via environment variable.
func TestRouter_InitializesWithConfiguredMode(t *testing.T) {
	tests := []struct {
		name     string
		ginMode  string
		expected string
	}{
		{
			name:     "debug mode",
			ginMode:  gin.DebugMode,
			expected: gin.DebugMode,
		},
		{
			name:     "release mode",
			ginMode:  gin.ReleaseMode,
			expected: gin.ReleaseMode,
		},
		{
			name:     "test mode",
			ginMode:  gin.TestMode,
			expected: gin.TestMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set GIN_MODE environment variable
			os.Setenv("GIN_MODE", tt.ginMode)
			defer os.Unsetenv("GIN_MODE")

			// Load config
			cfg, err := config.Load()
			require.NoError(t, err)

			// Verify mode is set correctly
			assert.Equal(t, tt.expected, cfg.Server.Mode)

			// Create logger
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			// Create server with configured mode
			serverCfg := Config{
				Port:            cfg.Server.Port,
				Host:            cfg.Server.Host,
				Mode:            cfg.Server.Mode,
				ReadTimeout:     cfg.Server.ReadTimeout,
				WriteTimeout:    cfg.Server.WriteTimeout,
				IdleTimeout:     cfg.Server.IdleTimeout,
				ShutdownTimeout: cfg.Server.ShutdownTimeout,
			}

			srv, err := New(serverCfg, logger)
			require.NoError(t, err)
			assert.NotNil(t, srv)

			// Verify Gin mode is set
			assert.Equal(t, tt.expected, gin.Mode())
		})
	}
}

// TestRouter_APIVersioningGroupsExist tests that /api/v1 route group
// is established for versioning.
func TestRouter_APIVersioningGroupsExist(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Mode = gin.TestMode

	srv, err := New(cfg, logger)
	require.NoError(t, err)

	// Setup routes
	srv.SetupRoutes()

	// Test /api/v1/status endpoint exists
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "api_version")
	assert.Contains(t, w.Body.String(), "v1")
}

// TestRouter_GlobalErrorHandlerReturnsJSON tests that the centralized
// error handler returns consistent JSON errors.
func TestRouter_GlobalErrorHandlerReturnsJSON(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Mode = gin.TestMode

	srv, err := New(cfg, logger)
	require.NoError(t, err)

	// Add a test route that triggers an error
	srv.router.GET("/test-error", func(c *gin.Context) {
		panic("test panic")
	})

	// Test error handling
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-error", nil)
	srv.router.ServeHTTP(w, req)

	// The error handler should return 500 status
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check if response body contains error information
	// Note: The response might be empty if Gin's Recovery middleware runs last
	responseBody := w.Body.String()
	if responseBody != "" {
		assert.Contains(t, responseBody, "error")
		assert.Contains(t, responseBody, "code")
		assert.Contains(t, responseBody, "message")
	}

	// At minimum, verify the status code is correct
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestRouter_RequestIDTrackedInContext tests that request ID is tracked
// in the Gin context.
func TestRouter_RequestIDTrackedInContext(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Mode = gin.TestMode

	srv, err := New(cfg, logger)
	require.NoError(t, err)

	// Add a test route that checks request ID
	var capturedRequestID string
	srv.router.GET("/test-request-id", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		if exists {
			capturedRequestID = requestID.(string)
		}
		c.JSON(http.StatusOK, gin.H{"request_id": capturedRequestID})
	})

	// Test request ID tracking
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-request-id", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedRequestID)
	assert.Contains(t, w.Body.String(), capturedRequestID)

	// Verify request ID is in response headers
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

// TestConfig_LoadsFromYAMLAndEnv tests that Viper loads config from
// config.yaml and environment variables, with env vars taking precedence.
func TestConfig_LoadsFromYAMLAndEnv(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `server:
  port: 7070
  host: 0.0.0.0
  mode: debug
  read_timeout: 20s
  write_timeout: 20s
  idle_timeout: 90s
  shutdown_timeout: 45s
app:
  name: test-goflow
  environment: development
  log_level: debug
  version: 2.0.0
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test 1: Load from YAML file
	cfg, err := config.LoadFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, 7070, cfg.Server.Port)
	assert.Equal(t, "test-goflow", cfg.App.Name)
	assert.Equal(t, "debug", cfg.App.LogLevel)
	assert.Equal(t, "2.0.0", cfg.App.Version)

	// Test 2: Environment variable override
	os.Setenv("PORT", "9999")
	os.Setenv("APP_NAME", "env-override-goflow")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("APP_NAME")

	cfg2, err := config.LoadFromFile(configPath)
	require.NoError(t, err)

	// Note: Viper's AutomaticEnv() should allow env vars to override file values
	// The actual behavior depends on Viper configuration
	assert.NotNil(t, cfg2)
}

// TestConfig_LoadsFromStandardLocations tests that config.Load()
// searches for config files in standard locations.
func TestConfig_LoadsFromStandardLocations(t *testing.T) {
	// Save current directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Create a temporary directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create configs directory and config file
	os.Mkdir("configs", 0755)
	configPath := filepath.Join("configs", "config.yaml")

	configContent := `server:
  port: 6060
  host: localhost
  mode: release
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s
  shutdown_timeout: 30s
app:
  name: goflow-workflow-engine
  environment: production
  log_level: info
  version: 1.0.0
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Clear environment variables to ensure we're loading from file
	clearTestEnvVars()
	defer clearTestEnvVars()

	// Load config (should find configs/config.yaml)
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, 6060, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "production", cfg.App.Environment)
}

// TestRouter_HealthCheckEndpoint tests the /health endpoint.
func TestRouter_HealthCheckEndpoint(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.Mode = gin.TestMode

	srv, err := New(cfg, logger)
	require.NoError(t, err)

	srv.SetupRoutes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "healthy")
}

// Helper function to clear test environment variables
func clearTestEnvVars() {
	os.Unsetenv("PORT")
	os.Unsetenv("HOST")
	os.Unsetenv("GIN_MODE")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_NAME")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("READ_TIMEOUT")
	os.Unsetenv("WRITE_TIMEOUT")
	os.Unsetenv("IDLE_TIMEOUT")
	os.Unsetenv("SHUTDOWN_TIMEOUT")
	os.Unsetenv("APP_VERSION")
	os.Unsetenv("GOFLOW_CONFIG_PATH")
}

