// Package config provides configuration management for the GoFlow Workflow Engine.
//
// This package handles loading configuration from environment variables,
// configuration files, and provides default values for all settings.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Config holds all application configuration.
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server" mapstructure:"server"`

	// Application configuration
	App AppConfig `json:"app" mapstructure:"app"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	// Port is the HTTP server port
	Port int `json:"port" mapstructure:"port"`
	// Host is the HTTP server host
	Host string `json:"host" mapstructure:"host"`
	// Mode is the Gin mode (debug, release, test)
	Mode string `json:"mode" mapstructure:"mode"`
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"`
	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration `json:"idle_timeout" mapstructure:"idle_timeout"`
	// ShutdownTimeout is the maximum duration to wait for graceful shutdown
	ShutdownTimeout time.Duration `json:"shutdown_timeout" mapstructure:"shutdown_timeout"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	// Name is the application name
	Name string `json:"name" mapstructure:"name"`
	// Environment is the deployment environment (development, staging, production)
	Environment string `json:"environment" mapstructure:"environment"`
	// LogLevel is the logging level (debug, info, warn, error)
	LogLevel string `json:"log_level" mapstructure:"log_level"`
	// Version is the application version
	Version string `json:"version" mapstructure:"version"`
}

// Load loads configuration with the following precedence:
//  1. Environment variables (highest priority)
//  2. Configuration file from CONFIG_PATH or default locations
//  3. Default values (lowest priority)
//
// Environment variables:
//   - CONFIG_PATH: Path to config file (default: searches standard locations)
//   - PORT: HTTP server port (default: 8080)
//   - HOST: HTTP server host (default: 0.0.0.0)
//   - GIN_MODE: Gin mode - debug, release, test (default: debug)
//   - APP_ENV: Application environment (default: development)
//   - LOG_LEVEL: Log level (default: info)
//   - READ_TIMEOUT: Read timeout in seconds (default: 15)
//   - WRITE_TIMEOUT: Write timeout in seconds (default: 15)
//   - IDLE_TIMEOUT: Idle timeout in seconds (default: 60)
//   - SHUTDOWN_TIMEOUT: Shutdown timeout in seconds (default: 30)
//
// Returns:
//   - *Config: Loaded configuration
//   - error: Error if configuration is invalid
//
// Example:
//
//	config, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
func Load() (*Config, error) {
	// Check if config path is specified in environment
	configPath := os.Getenv("CONFIG_PATH")

	// If not specified, try to find config file in standard locations
	if configPath == "" {
		foundPath, err := FindConfigFile()
		if err == nil {
			configPath = foundPath
		}
	}

	// If config file exists, load from file (with env override)
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return LoadFromFile(configPath)
		}
	}

	// Fall back to environment variables with defaults
	return loadFromEnvWithDefaults()
}

// loadFromEnvWithDefaults loads configuration from environment variables with default values.
// This is used as a fallback when no config file is found.
func loadFromEnvWithDefaults() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("PORT", 8080),
			Host:            getEnv("HOST", "0.0.0.0"),
			Mode:            getEnv("GIN_MODE", gin.DebugMode),
			ReadTimeout:     getEnvAsDuration("READ_TIMEOUT", 15) * time.Second,
			WriteTimeout:    getEnvAsDuration("WRITE_TIMEOUT", 15) * time.Second,
			IdleTimeout:     getEnvAsDuration("IDLE_TIMEOUT", 60) * time.Second,
			ShutdownTimeout: getEnvAsDuration("SHUTDOWN_TIMEOUT", 30) * time.Second,
		},
		App: AppConfig{
			Name:        getEnv("APP_NAME", "goflow-workflow-engine"),
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate server port
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be between 1 and 65535)", c.Server.Port)
	}

	// Validate Gin mode
	validModes := map[string]bool{
		gin.DebugMode:   true,
		gin.ReleaseMode: true,
		gin.TestMode:    true,
	}
	if !validModes[c.Server.Mode] {
		return fmt.Errorf("invalid Gin mode: %s (must be debug, release, or test)", c.Server.Mode)
	}

	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[c.App.Environment] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, production, or test)", c.App.Environment)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.App.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.App.LogLevel)
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsDuration retrieves an environment variable as a duration (in seconds) or returns a default value.
func getEnvAsDuration(key string, defaultValue int) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return time.Duration(defaultValue)
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.Duration(defaultValue)
	}

	return time.Duration(value)
}
