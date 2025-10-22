package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockWorkflowManager is a mock implementation of WorkflowManager
type MockWorkflowManager struct {
	mock.Mock
}

func (m *MockWorkflowManager) Start(ctx context.Context, defID string, input []byte) (string, error) {
	args := m.Called(ctx, defID, input)
	return args.String(0), args.Error(1)
}

func (m *MockWorkflowManager) GetStatus(ctx context.Context, instanceID string) (WorkflowStatus, error) {
	args := m.Called(ctx, instanceID)
	return args.Get(0).(WorkflowStatus), args.Error(1)
}

func (m *MockWorkflowManager) Signal(ctx context.Context, instanceID string, signalName string, payload []byte) error {
	args := m.Called(ctx, instanceID, signalName, payload)
	return args.Error(0)
}

func (m *MockWorkflowManager) Terminate(ctx context.Context, instanceID string, reason string) error {
	args := m.Called(ctx, instanceID, reason)
	return args.Error(0)
}

// createTestLogger creates a test logger
func createTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// createTestConfig creates a test configuration
func createTestConfig() Config {
	return Config{
		ServerPort:  8080,
		LogLevel:    "debug",
		Environment: "test",
	}
}

// TestNew_Success verifies successful creation of GoFlow app
func TestNew_Success(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)

	// Act
	app, err := New(config, logger, manager)

	// Assert
	require.NoError(t, err, "New should not return error")
	require.NotNil(t, app, "App should not be nil")
	assert.Equal(t, config, app.GetConfig(), "Config should match")
	assert.Equal(t, logger, app.GetLogger(), "Logger should match")
	assert.Equal(t, manager, app.GetWorkflowManager(), "Manager should match")
	assert.False(t, app.IsStarted(), "App should not be started initially")
}

// TestNew_NilLogger verifies error when logger is nil
func TestNew_NilLogger(t *testing.T) {
	// Arrange
	config := createTestConfig()
	manager := new(MockWorkflowManager)

	// Act
	app, err := New(config, nil, manager)

	// Assert
	require.Error(t, err, "New should return error for nil logger")
	assert.Nil(t, app, "App should be nil")
	assert.Contains(t, err.Error(), "logger cannot be nil", "Error message should mention logger")
}

// TestNew_NilManager verifies error when manager is nil
func TestNew_NilManager(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()

	// Act
	app, err := New(config, logger, nil)

	// Assert
	require.Error(t, err, "New should return error for nil manager")
	assert.Nil(t, app, "App should be nil")
	assert.Contains(t, err.Error(), "workflow manager cannot be nil", "Error message should mention manager")
}

// TestGoFlow_Start_Success verifies successful app start
func TestGoFlow_Start_Success(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)
	ctx := context.Background()

	// Act
	err := app.Start(ctx)

	// Assert
	require.NoError(t, err, "Start should not return error")
	assert.True(t, app.IsStarted(), "App should be started")
}

// TestGoFlow_Start_AlreadyStarted verifies error when starting already started app
func TestGoFlow_Start_AlreadyStarted(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)
	ctx := context.Background()
	app.Start(ctx)

	// Act
	err := app.Start(ctx)

	// Assert
	require.Error(t, err, "Start should return error when already started")
	assert.Contains(t, err.Error(), "already started", "Error message should mention already started")
}

// TestGoFlow_Stop_Success verifies successful app stop
func TestGoFlow_Stop_Success(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)
	ctx := context.Background()
	app.Start(ctx)

	// Act
	err := app.Stop(ctx)

	// Assert
	require.NoError(t, err, "Stop should not return error")
	assert.False(t, app.IsStarted(), "App should not be started after stop")
}

// TestGoFlow_Stop_NotStarted verifies error when stopping non-started app
func TestGoFlow_Stop_NotStarted(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)
	ctx := context.Background()

	// Act
	err := app.Stop(ctx)

	// Assert
	require.Error(t, err, "Stop should return error when not started")
	assert.Contains(t, err.Error(), "not started", "Error message should mention not started")
}

// TestGoFlow_StartStop_Cycle verifies multiple start/stop cycles
func TestGoFlow_StartStop_Cycle(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)
	ctx := context.Background()

	// Act & Assert - First cycle
	err := app.Start(ctx)
	require.NoError(t, err, "First start should succeed")
	assert.True(t, app.IsStarted(), "App should be started")

	err = app.Stop(ctx)
	require.NoError(t, err, "First stop should succeed")
	assert.False(t, app.IsStarted(), "App should be stopped")

	// Act & Assert - Second cycle
	err = app.Start(ctx)
	require.NoError(t, err, "Second start should succeed")
	assert.True(t, app.IsStarted(), "App should be started again")

	err = app.Stop(ctx)
	require.NoError(t, err, "Second stop should succeed")
	assert.False(t, app.IsStarted(), "App should be stopped again")
}

// TestGoFlow_GetWorkflowManager verifies GetWorkflowManager returns correct manager
func TestGoFlow_GetWorkflowManager(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)

	// Act
	result := app.GetWorkflowManager()

	// Assert
	assert.Equal(t, manager, result, "GetWorkflowManager should return the same manager")
}

// TestGoFlow_GetLogger verifies GetLogger returns correct logger
func TestGoFlow_GetLogger(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)

	// Act
	result := app.GetLogger()

	// Assert
	assert.Equal(t, logger, result, "GetLogger should return the same logger")
}

// TestGoFlow_GetConfig verifies GetConfig returns correct config
func TestGoFlow_GetConfig(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)

	// Act
	result := app.GetConfig()

	// Assert
	assert.Equal(t, config, result, "GetConfig should return the same config")
}

// TestGoFlow_IsStarted_InitialState verifies initial state is not started
func TestGoFlow_IsStarted_InitialState(t *testing.T) {
	// Arrange
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	app, _ := New(config, logger, manager)

	// Act
	result := app.IsStarted()

	// Assert
	assert.False(t, result, "IsStarted should return false initially")
}

// TestWorkflowStatus_Structure verifies WorkflowStatus struct
func TestWorkflowStatus_Structure(t *testing.T) {
	// Arrange
	now := time.Now()
	completedAt := now.Add(1 * time.Hour)
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	// Act
	status := WorkflowStatus{
		InstanceID:   "inst-123",
		DefinitionID: "def-456",
		State:        "running",
		StartedAt:    now,
		CompletedAt:  &completedAt,
		Error:        "test error",
		Metadata:     metadata,
	}

	// Assert
	assert.Equal(t, "inst-123", status.InstanceID)
	assert.Equal(t, "def-456", status.DefinitionID)
	assert.Equal(t, "running", status.State)
	assert.Equal(t, now, status.StartedAt)
	assert.Equal(t, &completedAt, status.CompletedAt)
	assert.Equal(t, "test error", status.Error)
	assert.Equal(t, metadata, status.Metadata)
}

// TestConfig_Structure verifies Config struct
func TestConfig_Structure(t *testing.T) {
	// Act
	config := Config{
		ServerPort:  9090,
		LogLevel:    "info",
		Environment: "production",
	}

	// Assert
	assert.Equal(t, 9090, config.ServerPort)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "production", config.Environment)
}

// BenchmarkNew benchmarks New function
func BenchmarkNew(b *testing.B) {
	config := createTestConfig()
	logger := createTestLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := new(MockWorkflowManager)
		_, _ = New(config, logger, manager)
	}
}

// BenchmarkGoFlow_Start benchmarks Start method
func BenchmarkGoFlow_Start(b *testing.B) {
	config := createTestConfig()
	logger := createTestLogger()
	manager := new(MockWorkflowManager)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app, _ := New(config, logger, manager)
		_ = app.Start(ctx)
	}
}

