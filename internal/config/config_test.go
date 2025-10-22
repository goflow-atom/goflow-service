package config

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	config, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, config)

	// Check server defaults
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, gin.DebugMode, config.Server.Mode)
	assert.Equal(t, 15*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 15*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.Server.IdleTimeout)
	assert.Equal(t, 30*time.Second, config.Server.ShutdownTimeout)

	// Check app defaults
	assert.Equal(t, "goflow-workflow-engine", config.App.Name)
	assert.Equal(t, "development", config.App.Environment)
	assert.Equal(t, "info", config.App.LogLevel)
	assert.Equal(t, "1.0.0", config.App.Version)
}

func TestLoad_FromEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("HOST", "localhost")
	os.Setenv("GIN_MODE", "release")
	os.Setenv("APP_ENV", "production")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("READ_TIMEOUT", "30")
	os.Setenv("WRITE_TIMEOUT", "30")
	os.Setenv("IDLE_TIMEOUT", "120")
	os.Setenv("SHUTDOWN_TIMEOUT", "60")
	defer clearEnvVars()

	config, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "localhost", config.Server.Host)
	assert.Equal(t, gin.ReleaseMode, config.Server.Mode)
	assert.Equal(t, "production", config.App.Environment)
	assert.Equal(t, "error", config.App.LogLevel)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, config.Server.IdleTimeout)
	assert.Equal(t, 60*time.Second, config.Server.ShutdownTimeout)
}

func TestValidate_Success(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
			Mode: gin.DebugMode,
		},
		App: AppConfig{
			Environment: "development",
			LogLevel:    "info",
		},
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestValidate_InvalidPort(t *testing.T) {
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
			config := &Config{
				Server: ServerConfig{
					Port: tt.port,
					Mode: gin.DebugMode,
				},
				App: AppConfig{
					Environment: "development",
					LogLevel:    "info",
				},
			}

			err := config.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid server port")
		})
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "invalid",
		},
		App: AppConfig{
			Environment: "development",
			LogLevel:    "info",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Gin mode")
}

func TestValidate_InvalidEnvironment(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: gin.DebugMode,
		},
		App: AppConfig{
			Environment: "invalid",
			LogLevel:    "info",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid environment")
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: gin.DebugMode,
		},
		App: AppConfig{
			Environment: "development",
			LogLevel:    "invalid",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestValidate_AllValidModes(t *testing.T) {
	modes := []string{gin.DebugMode, gin.ReleaseMode, gin.TestMode}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{
					Port: 8080,
					Mode: mode,
				},
				App: AppConfig{
					Environment: "development",
					LogLevel:    "info",
				},
			}

			err := config.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestValidate_AllValidEnvironments(t *testing.T) {
	envs := []string{"development", "staging", "production", "test"}

	for _, env := range envs {
		t.Run(env, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{
					Port: 8080,
					Mode: gin.DebugMode,
				},
				App: AppConfig{
					Environment: env,
					LogLevel:    "info",
				},
			}

			err := config.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestValidate_AllValidLogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{
					Port: 8080,
					Mode: gin.DebugMode,
				},
				App: AppConfig{
					Environment: "development",
					LogLevel:    level,
				},
			}

			err := config.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	value := getEnv("TEST_VAR", "default")
	assert.Equal(t, "test_value", value)

	value = getEnv("NON_EXISTENT", "default")
	assert.Equal(t, "default", value)
}

func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	value := getEnvAsInt("TEST_INT", 10)
	assert.Equal(t, 42, value)

	value = getEnvAsInt("NON_EXISTENT", 10)
	assert.Equal(t, 10, value)

	os.Setenv("TEST_INVALID", "not_a_number")
	defer os.Unsetenv("TEST_INVALID")

	value = getEnvAsInt("TEST_INVALID", 10)
	assert.Equal(t, 10, value)
}

func TestGetEnvAsDuration(t *testing.T) {
	os.Setenv("TEST_DURATION", "60")
	defer os.Unsetenv("TEST_DURATION")

	value := getEnvAsDuration("TEST_DURATION", 30)
	assert.Equal(t, time.Duration(60), value)

	value = getEnvAsDuration("NON_EXISTENT", 30)
	assert.Equal(t, time.Duration(30), value)

	os.Setenv("TEST_INVALID", "not_a_number")
	defer os.Unsetenv("TEST_INVALID")

	value = getEnvAsDuration("TEST_INVALID", 30)
	assert.Equal(t, time.Duration(30), value)
}

// Helper function to clear environment variables
func clearEnvVars() {
	os.Unsetenv("PORT")
	os.Unsetenv("HOST")
	os.Unsetenv("GIN_MODE")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("READ_TIMEOUT")
	os.Unsetenv("WRITE_TIMEOUT")
	os.Unsetenv("IDLE_TIMEOUT")
	os.Unsetenv("SHUTDOWN_TIMEOUT")
	os.Unsetenv("APP_NAME")
	os.Unsetenv("APP_VERSION")
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	clearEnvVars()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load()
	}
}

func BenchmarkValidate(b *testing.B) {
	config := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: gin.DebugMode,
		},
		App: AppConfig{
			Environment: "development",
			LogLevel:    "info",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

