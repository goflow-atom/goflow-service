// Package middleware provides HTTP middleware for the GoFlow API.
package middleware

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"

	"go.uber.org/zap"
)

// PolicyConfig represents the authorization policy configuration.
//
// This structure defines the complete authorization policy for the application,
// including resource-action-role mappings and default behavior.
type PolicyConfig struct {
	// DefaultDeny determines whether to deny access by default when no policy is found
	DefaultDeny bool `json:"default_deny" yaml:"default_deny"`

	// Resources maps resource names to their authorization policies
	Resources map[string]ResourcePolicy `json:"resources" yaml:"resources"`
}

// ResourcePolicy defines the authorization policy for a single resource.
type ResourcePolicy struct {
	// Name is the resource name
	Name string `json:"name" yaml:"name"`

	// Description provides documentation for the resource
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Actions maps action names to the roles allowed to perform them
	Actions map[string][]string `json:"actions" yaml:"actions"`
}

// DefaultPolicyConfig returns the default authorization policy configuration.
//
// This provides a sensible default policy for common GoFlow resources:
// - Workflows: admin and developer can create/update/delete, all roles can read
// - Executions: admin and developer can create/delete, all roles can read
// - Schedules: admin and developer can create/update/delete, all roles can read
// - Logs: all roles can read
//
// Returns:
//   - *PolicyConfig: Default policy configuration
//
// Example:
//
//	config := middleware.DefaultPolicyConfig()
//	pm := middleware.NewPolicyManagerFromConfig(config, logger)
func DefaultPolicyConfig() *PolicyConfig {
	return &PolicyConfig{
		DefaultDeny: true,
		Resources: map[string]ResourcePolicy{
			"workflow": {
				Name:        "workflow",
				Description: "Workflow definitions and templates",
				Actions: map[string][]string{
					"create": {"admin", "developer"},
					"read":   {"admin", "developer", "operator", "viewer"},
					"update": {"admin", "developer"},
					"delete": {"admin"},
					"execute": {"admin", "developer", "operator"},
				},
			},
			"execution": {
				Name:        "execution",
				Description: "Workflow execution instances",
				Actions: map[string][]string{
					"create": {"admin", "developer", "operator"},
					"read":   {"admin", "developer", "operator", "viewer"},
					"cancel": {"admin", "developer", "operator"},
					"retry":  {"admin", "developer", "operator"},
					"delete": {"admin"},
				},
			},
			"schedule": {
				Name:        "schedule",
				Description: "Workflow schedules and triggers",
				Actions: map[string][]string{
					"create": {"admin", "developer"},
					"read":   {"admin", "developer", "operator", "viewer"},
					"update": {"admin", "developer"},
					"delete": {"admin", "developer"},
					"enable": {"admin", "developer"},
					"disable": {"admin", "developer"},
				},
			},
			"log": {
				Name:        "log",
				Description: "Execution logs and audit trails",
				Actions: map[string][]string{
					"read": {"admin", "developer", "operator", "viewer"},
				},
			},
			"user": {
				Name:        "user",
				Description: "User management",
				Actions: map[string][]string{
					"create": {"admin"},
					"read":   {"admin", "developer", "operator", "viewer"},
					"update": {"admin"},
					"delete": {"admin"},
				},
			},
		},
	}
}

// LoadPolicyConfigFromFile loads authorization policy configuration from a JSON file.
//
// Parameters:
//   - filePath: Path to the JSON configuration file
//
// Returns:
//   - *PolicyConfig: Loaded policy configuration
//   - error: Error if file cannot be read or parsed
//
// Example:
//
//	config, err := middleware.LoadPolicyConfigFromFile("./config/authz_policy.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	pm := middleware.NewPolicyManagerFromConfig(config, logger)
func LoadPolicyConfigFromFile(filePath string) (*PolicyConfig, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy config file: %w", err)
	}

	// Parse JSON
	var config PolicyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse policy config: %w", err)
	}

	return &config, nil
}

// SavePolicyConfigToFile saves authorization policy configuration to a JSON file.
//
// Parameters:
//   - config: Policy configuration to save
//   - filePath: Path to the JSON configuration file
//
// Returns:
//   - error: Error if file cannot be written
//
// Example:
//
//	config := middleware.DefaultPolicyConfig()
//	err := middleware.SavePolicyConfigToFile(config, "./config/authz_policy.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
func SavePolicyConfigToFile(config *PolicyConfig, filePath string) error {
	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write policy config file: %w", err)
	}

	return nil
}

// NewPolicyManagerFromConfig creates a new policy manager from a configuration.
//
// This is a convenience function that creates a PolicyManager and loads all
// policies from the configuration.
//
// Parameters:
//   - config: Policy configuration
//   - logger: Zap logger instance
//
// Returns:
//   - *PolicyManager: New policy manager with loaded policies
//
// Example:
//
//	config := middleware.DefaultPolicyConfig()
//	pm := middleware.NewPolicyManagerFromConfig(config, logger)
//	router.POST("/workflows",
//	    middleware.RequireResourceAction(pm, "workflow", "create"),
//	    handler.CreateWorkflow)
func NewPolicyManagerFromConfig(config *PolicyConfig, logger *zap.Logger) *PolicyManager {
	pm := NewPolicyManager(config.DefaultDeny, logger)

	// Load all resource policies
	for resourceName, resourcePolicy := range config.Resources {
		pm.AddPolicy(resourceName, resourcePolicy.Actions)
	}

	logger.Info("Loaded authorization policies",
		zap.Int("resource_count", len(config.Resources)),
		zap.Bool("default_deny", config.DefaultDeny),
	)

	return pm
}

// LoadPolicyManagerFromFile creates a new policy manager from a configuration file.
//
// This is a convenience function that loads the configuration from a file and
// creates a PolicyManager with the loaded policies.
//
// Parameters:
//   - filePath: Path to the JSON configuration file
//   - logger: Zap logger instance
//
// Returns:
//   - *PolicyManager: New policy manager with loaded policies
//   - error: Error if file cannot be read or parsed
//
// Example:
//
//	pm, err := middleware.LoadPolicyManagerFromFile("./config/authz_policy.json", logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	router.POST("/workflows",
//	    middleware.RequireResourceAction(pm, "workflow", "create"),
//	    handler.CreateWorkflow)
func LoadPolicyManagerFromFile(filePath string, logger *zap.Logger) (*PolicyManager, error) {
	config, err := LoadPolicyConfigFromFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewPolicyManagerFromConfig(config, logger), nil
}

// ValidatePolicyConfig validates a policy configuration.
//
// This function checks for common configuration errors:
// - Empty resource names
// - Empty action names
// - Empty role lists
// - Duplicate resources
//
// Parameters:
//   - config: Policy configuration to validate
//
// Returns:
//   - error: Error if configuration is invalid, nil if valid
//
// Example:
//
//	config := middleware.DefaultPolicyConfig()
//	if err := middleware.ValidatePolicyConfig(config); err != nil {
//	    log.Fatal(err)
//	}
func ValidatePolicyConfig(config *PolicyConfig) error {
	if config == nil {
		return fmt.Errorf("policy config cannot be nil")
	}

	if len(config.Resources) == 0 {
		return fmt.Errorf("policy config must have at least one resource")
	}

	// Validate each resource
	for resourceName, resourcePolicy := range config.Resources {
		if resourceName == "" {
			return fmt.Errorf("resource name cannot be empty")
		}

		if len(resourcePolicy.Actions) == 0 {
			return fmt.Errorf("resource %s must have at least one action", resourceName)
		}

		// Validate each action
		for actionName, roles := range resourcePolicy.Actions {
			if actionName == "" {
				return fmt.Errorf("action name cannot be empty for resource %s", resourceName)
			}

			if len(roles) == 0 {
				return fmt.Errorf("action %s on resource %s must have at least one role", actionName, resourceName)
			}

			// Validate each role
			if slices.Contains(roles, "") {
				return fmt.Errorf("role name cannot be empty for action %s on resource %s", actionName, resourceName)
			}
		}
	}

	return nil
}

// MergePolicyConfigs merges multiple policy configurations into one.
//
// Later configurations override earlier ones for the same resource.
// This is useful for combining a base configuration with environment-specific overrides.
//
// Parameters:
//   - configs: Policy configurations to merge (in order of precedence)
//
// Returns:
//   - *PolicyConfig: Merged policy configuration
//
// Example:
//
//	baseConfig := middleware.DefaultPolicyConfig()
//	customConfig, _ := middleware.LoadPolicyConfigFromFile("./config/custom_policy.json")
//	merged := middleware.MergePolicyConfigs(baseConfig, customConfig)
func MergePolicyConfigs(configs ...*PolicyConfig) *PolicyConfig {
	if len(configs) == 0 {
		return DefaultPolicyConfig()
	}

	// Start with the first config
	merged := &PolicyConfig{
		DefaultDeny: configs[0].DefaultDeny,
		Resources:   make(map[string]ResourcePolicy),
	}

	// Copy resources from first config
	for k, v := range configs[0].Resources {
		merged.Resources[k] = v
	}

	// Merge subsequent configs
	for i := 1; i < len(configs); i++ {
		config := configs[i]

		// Override default deny if specified
		merged.DefaultDeny = config.DefaultDeny

		// Merge resources
		maps.Copy(merged.Resources, config.Resources)
	}

	return merged
}

