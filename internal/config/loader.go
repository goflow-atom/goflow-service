// Package config provides configuration management for the GoFlow Workflow Engine.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// LoadFromFile loads configuration from a YAML file.
//
// This function uses Viper to load configuration from a YAML file
// and merge it with environment variables. Environment variables
// take precedence over file values.
//
// Parameters:
//   - configPath: Path to the configuration file
//
// Returns:
//   - *Config: Loaded configuration
//   - error: Error if file cannot be read or parsed
//
// Example:
//
//	config, err := config.LoadFromFile("configs/config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
func LoadFromFile(configPath string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Initialize Viper
	v := viper.New()

	// Set config file
	v.SetConfigFile(configPath)

	// Enable environment variable override
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadFromEnv loads configuration from environment variables only.
//
// This is a convenience function that calls Load() and is provided
// for clarity when explicitly loading from environment variables.
//
// Returns:
//   - *Config: Loaded configuration
//   - error: Error if configuration is invalid
func LoadFromEnv() (*Config, error) {
	return Load()
}

// LoadWithDefaults loads configuration with the following precedence:
//  1. Environment variables (highest priority)
//  2. Configuration file (if exists)
//  3. Default values (lowest priority)
//
// Parameters:
//   - configPath: Path to the configuration file (optional, can be empty)
//
// Returns:
//   - *Config: Loaded configuration
//   - error: Error if configuration is invalid
//
// Example:
//
//	// Try to load from file, fall back to env vars and defaults
//	config, err := config.LoadWithDefaults("configs/config.yaml")
func LoadWithDefaults(configPath string) (*Config, error) {
	// If config path is provided and file exists, load from file
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return LoadFromFile(configPath)
		}
	}

	// Otherwise, load from environment variables with defaults
	return Load()
}

// FindConfigFile searches for a configuration file in common locations.
//
// Search order:
//  1. ./configs/config.yaml
//  2. ./config.yaml
//  3. /etc/goflow/config.yaml
//  4. $HOME/.goflow/config.yaml
//
// Returns:
//   - string: Path to the found config file
//   - error: Error if no config file is found
func FindConfigFile() (string, error) {
	searchPaths := []string{
		"configs/config.yaml",
		"config.yaml",
		"/etc/goflow/config.yaml",
	}

	// Add home directory path
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(home, ".goflow", "config.yaml"))
	}

	// Search for config file
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no config file found in standard locations")
}

// MustLoad loads configuration and panics if it fails.
//
// This is useful for application startup where configuration
// errors should be fatal.
//
// Returns:
//   - *Config: Loaded configuration
func MustLoad() *Config {
	config, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return config
}

// MustLoadFromFile loads configuration from a file and panics if it fails.
//
// Parameters:
//   - configPath: Path to the configuration file
//
// Returns:
//   - *Config: Loaded configuration
func MustLoadFromFile(configPath string) *Config {
	config, err := LoadFromFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration from %s: %v", configPath, err))
	}
	return config
}
