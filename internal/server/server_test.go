package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "0.0.0.0", config.Host)
	assert.Equal(t, gin.DebugMode, config.Mode)
	assert.Equal(t, 15*time.Second, config.ReadTimeout)
	assert.Equal(t, 15*time.Second, config.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.IdleTimeout)
	assert.Equal(t, 30*time.Second, config.ShutdownTimeout)
}

func TestNew_Success(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	srv, err := New(config, logger)

	require.NoError(t, err)
	assert.NotNil(t, srv)
	assert.Equal(t, config.Port, srv.config.Port)
	assert.Equal(t, config.Host, srv.config.Host)
	assert.NotNil(t, srv.router)
}

func TestNew_NilLogger(t *testing.T) {
	config := DefaultConfig()
	srv, err := New(config, nil)

	assert.Error(t, err)
	assert.Nil(t, srv)
	assert.Contains(t, err.Error(), "logger cannot be nil")
}

func TestNew_InvalidPort(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"port too large", 65536},
		{"port too large 2", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Port = tt.port

			srv, err := New(config, logger)

			assert.Error(t, err)
			assert.Nil(t, srv)
			assert.Contains(t, err.Error(), "invalid port")
		})
	}
}

func TestNew_EmptyHost(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	config.Host = ""

	srv, err := New(config, logger)

	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", srv.config.Host)
}

func TestNew_EmptyMode(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	config.Mode = ""

	srv, err := New(config, logger)

	require.NoError(t, err)
	assert.Equal(t, gin.DebugMode, srv.config.Mode)
}

func TestSetupRoutes(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	srv, err := New(config, logger)
	require.NoError(t, err)

	srv.SetupRoutes()

	// Test that routes are registered
	router := srv.GetRouter()
	assert.NotNil(t, router)

	// Test health endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test root endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test API v1 status endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/status", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheck(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	srv, err := New(config, logger)
	require.NoError(t, err)

	srv.SetupRoutes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	assert.Contains(t, w.Body.String(), "goflow-workflow-engine")
}

func TestRoot(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	srv, err := New(config, logger)
	require.NoError(t, err)

	srv.SetupRoutes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "GoFlow Workflow Engine")
	assert.Contains(t, w.Body.String(), "Welcome")
}

func TestStatus(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	srv, err := New(config, logger)
	require.NoError(t, err)

	srv.SetupRoutes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "running")
	assert.Contains(t, w.Body.String(), "v1")
}

func TestGetRouter(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	srv, err := New(config, logger)
	require.NoError(t, err)

	router := srv.GetRouter()
	assert.NotNil(t, router)
	assert.IsType(t, &gin.Engine{}, router)
}

func TestGetAddress(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	config.Port = 9090
	config.Host = "localhost"

	srv, err := New(config, logger)
	require.NoError(t, err)

	address := srv.GetAddress()
	assert.Equal(t, "localhost:9090", address)
}

func TestStart_Success(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	config.Port = 0 // Use random available port
	// Override port validation for testing
	config.Port = 18080

	srv, err := New(config, logger)
	require.NoError(t, err)

	srv.SetupRoutes()

	ctx := context.Background()
	err = srv.Start(ctx)
	require.NoError(t, err)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is running by making a request
	resp, err := http.Get(fmt.Sprintf("http://%s/health", srv.GetAddress()))
	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Stop the server
	err = srv.Stop(ctx)
	assert.NoError(t, err)
}

func TestStop_NotStarted(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := DefaultConfig()
	srv, err := New(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = srv.Stop(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server not started")
}

func TestGinLogger(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	middleware := ginLogger(logger)
	assert.NotNil(t, middleware)

	// Test middleware with a mock request
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = New(config, logger)
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	config := DefaultConfig()
	srv, _ := New(config, logger)
	srv.SetupRoutes()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv.router.ServeHTTP(w, req)
	}
}

