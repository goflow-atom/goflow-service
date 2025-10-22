package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestErrorHandler_PanicRecovery(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorHandler(logger))

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
	assert.Contains(t, w.Body.String(), "internal server error")
}

func TestErrorHandler_GinError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorHandler(logger))

	router.GET("/error", func(c *gin.Context) {
		_ = c.Error(errors.New("test error"))
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
}

func TestErrorHandler_BindError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorHandler(logger))

	type TestRequest struct {
		Name string `json:"name" binding:"required"`
	}

	router.POST("/bind", func(c *gin.Context) {
		var req TestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(err).SetType(gin.ErrorTypeBind)
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/bind", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestHandleError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	router.GET("/test", func(c *gin.Context) {
		HandleError(c, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
	assert.Contains(t, w.Body.String(), "Resource not found")
}

func TestHandleErrorWithDetails(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	router.GET("/test", func(c *gin.Context) {
		details := map[string]interface{}{
			"field": "email",
			"issue": "invalid format",
		}
		HandleErrorWithDetails(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input", details)
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
	assert.Contains(t, w.Body.String(), "email")
	assert.Contains(t, w.Body.String(), "invalid format")
}

func TestMapErrorToHTTP(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
		},
		{
			name:           "not found error",
			err:            errors.New("resource not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "duplicate error",
			err:            errors.New("duplicate key"),
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
		},
		{
			name:           "validation error",
			err:            errors.New("invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name:           "unauthorized error",
			err:            errors.New("unauthorized access"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "forbidden error",
			err:            errors.New("permission denied"),
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
		},
		{
			name:           "generic error",
			err:            errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, _ := MapErrorToHTTP(tt.err)
			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestWrapError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		err := errors.New("original error")
		WrapError(c, err, "failed to process request")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorHandler_WithRequestID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorHandler(logger))

	router.GET("/test", func(c *gin.Context) {
		HandleError(c, http.StatusBadRequest, "TEST_ERROR", "Test error message")
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "request_id")

	// Verify request ID is in response header
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
}

func TestErrorHandler_NoRequestID(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.GET("/test", func(c *gin.Context) {
		HandleError(c, http.StatusBadRequest, "TEST_ERROR", "Test error message")
	})

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Should still work without request ID
	assert.Contains(t, w.Body.String(), "TEST_ERROR")
}

