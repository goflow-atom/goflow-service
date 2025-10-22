package app

import (
	"os"
	"testing"

	"github.com/goflow-atom/goflow-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvideConfig tests the configuration provider.
func TestProvideConfig(t *testing.T) {
	// Set required environment variables
	os.Setenv("PORT", "8080")
	os.Setenv("APP_ENV", "test")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("APP_ENV")

	cfg, err := ProvideConfig()

	require.NoError(t, err, "ProvideConfig should not return an error")
	require.NotNil(t, cfg, "Config should not be nil")
	assert.Equal(t, 8080, cfg.Server.Port, "Port should be 8080")
	assert.Equal(t, "test", cfg.App.Environment, "Environment should be test")
}

// TestProvideConfigError tests configuration provider with invalid config.
func TestProvideConfigError(t *testing.T) {
	// Set invalid port
	os.Setenv("PORT", "invalid")
	defer os.Unsetenv("PORT")

	cfg, err := ProvideConfig()

	// Should still work because getEnvAsInt has a default
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

// TestProvideLogger tests the logger provider.
func TestProvideLogger(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		wantErr     bool
	}{
		{
			name:        "Development logger",
			environment: "development",
			wantErr:     false,
		},
		{
			name:        "Production logger",
			environment: "production",
			wantErr:     false,
		},
		{
			name:        "Test logger",
			environment: "test",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{
					Environment: tt.environment,
				},
			}

			logger, err := ProvideLogger(cfg)

			if tt.wantErr {
				assert.Error(t, err, "ProvideLogger should return an error")
				assert.Nil(t, logger, "Logger should be nil on error")
			} else {
				require.NoError(t, err, "ProvideLogger should not return an error")
				require.NotNil(t, logger, "Logger should not be nil")
			}
		})
	}
}

// TestProvideServerConfig tests the server configuration provider.
func TestProvideServerConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
			Mode: "debug",
		},
	}

	serverCfg := ProvideServerConfig(cfg)

	assert.Equal(t, 8080, serverCfg.Port, "Port should match")
	assert.Equal(t, "localhost", serverCfg.Host, "Host should match")
	assert.Equal(t, "debug", serverCfg.Mode, "Mode should match")
}

// TestProvideServer tests the server provider.
func TestProvideServer(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
			Mode: "test",
		},
	}

	// Create logger
	logger, err := ProvideLogger(cfg)
	require.NoError(t, err)

	// Create server config
	serverCfg := ProvideServerConfig(cfg)

	// Create server
	srv, err := ProvideServer(serverCfg, logger)

	require.NoError(t, err, "ProvideServer should not return an error")
	require.NotNil(t, srv, "Server should not be nil")
	assert.Equal(t, "localhost:8080", srv.GetAddress(), "Server address should match")
}

// TestProvideCleanup tests the cleanup provider.
func TestProvideCleanup(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "test",
		},
	}

	logger, err := ProvideLogger(cfg)
	require.NoError(t, err)

	cleanup := ProvideCleanup(logger)

	require.NotNil(t, cleanup, "Cleanup function should not be nil")

	// Should not panic when called
	assert.NotPanics(t, func() {
		cleanup()
	}, "Cleanup should not panic")
}

// TestProvideCleanupWithNilLogger tests cleanup with nil logger.
func TestProvideCleanupWithNilLogger(t *testing.T) {
	cleanup := ProvideCleanup(nil)

	require.NotNil(t, cleanup, "Cleanup function should not be nil")

	// Should not panic even with nil logger
	assert.NotPanics(t, func() {
		cleanup()
	}, "Cleanup should not panic with nil logger")
}

// TestProviderSetIntegration tests the full provider set integration.
func TestProviderSetIntegration(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("APP_ENV", "test")
	os.Setenv("GIN_MODE", "test")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("APP_ENV")
	defer os.Unsetenv("GIN_MODE")

	// Test the full dependency chain
	cfg, err := ProvideConfig()
	require.NoError(t, err)

	logger, err := ProvideLogger(cfg)
	require.NoError(t, err)

	serverCfg := ProvideServerConfig(cfg)
	require.NotNil(t, serverCfg)

	srv, err := ProvideServer(serverCfg, logger)
	require.NoError(t, err)
	require.NotNil(t, srv)

	cleanup := ProvideCleanup(logger)
	require.NotNil(t, cleanup)

	// Cleanup
	cleanup()
}

// BenchmarkProvideConfig benchmarks the configuration provider.
func BenchmarkProvideConfig(b *testing.B) {
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ProvideConfig()
	}
}

// BenchmarkProvideLogger benchmarks the logger provider.
func BenchmarkProvideLogger(b *testing.B) {
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "production",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger, _ := ProvideLogger(cfg)
		if logger != nil {
			_ = logger.Sync()
		}
	}
}

// BenchmarkProvideServer benchmarks the server provider.
func BenchmarkProvideServer(b *testing.B) {
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
			Mode: "test",
		},
	}

	logger, _ := ProvideLogger(cfg)
	serverCfg := ProvideServerConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ProvideServer(serverCfg, logger)
	}
}

