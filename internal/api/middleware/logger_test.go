package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		capturedRequestID = GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedRequestID)
	
	// Verify it's a valid UUID
	_, err := uuid.Parse(capturedRequestID)
	assert.NoError(t, err)
	
	// Verify it's in the response header
	responseRequestID := w.Header().Get("X-Request-ID")
	assert.Equal(t, capturedRequestID, responseRequestID)
}

func TestRequestIDMiddleware_UsesExistingID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	existingID := "test-request-id-123"
	var capturedRequestID string
	
	router.GET("/test", func(c *gin.Context) {
		capturedRequestID = GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, existingID, capturedRequestID)
	assert.Equal(t, existingID, w.Header().Get("X-Request-ID"))
}

func TestLoggerMiddleware_LogsRequest(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(LoggerMiddleware(logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?foo=bar", nil)
	req.Header.Set("User-Agent", "test-agent")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// Middleware should not affect response
	assert.Contains(t, w.Body.String(), "ok")
}

func TestLoggerMiddleware_LogsError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(LoggerMiddleware(logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLoggerMiddleware_LogsClientError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(LoggerMiddleware(logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetRequestID_Exists(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	var requestID string
	router.GET("/test", func(c *gin.Context) {
		requestID = GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.NotEmpty(t, requestID)
	assert.Contains(t, w.Body.String(), requestID)
}

func TestGetRequestID_NotExists(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var requestID string
	router.GET("/test", func(c *gin.Context) {
		requestID = GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Empty(t, requestID)
}

func TestLogWithRequestID_WithID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	router.GET("/test", func(c *gin.Context) {
		reqLogger := LogWithRequestID(logger, c)
		assert.NotNil(t, reqLogger)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogWithRequestID_WithoutID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		reqLogger := LogWithRequestID(logger, c)
		assert.NotNil(t, reqLogger)
		assert.Equal(t, logger, reqLogger) // Should return same logger
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestIDMiddleware_MultipleRequests(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	requestIDs := make([]string, 0)
	router.GET("/test", func(c *gin.Context) {
		requestIDs = append(requestIDs, GetRequestID(c))
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act - Make multiple requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Assert - All request IDs should be unique
	assert.Len(t, requestIDs, 5)
	uniqueIDs := make(map[string]bool)
	for _, id := range requestIDs {
		assert.NotEmpty(t, id)
		uniqueIDs[id] = true
	}
	assert.Len(t, uniqueIDs, 5)
}

func TestLoggerMiddleware_CapturesLatency(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(LoggerMiddleware(logger))

	router.GET("/test", func(c *gin.Context) {
		// Simulate some processing time
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// The middleware should log the latency (we can't easily test the log output)
}

