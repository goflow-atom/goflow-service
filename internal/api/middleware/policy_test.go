package middleware

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDefaultPolicyConfig(t *testing.T) {
	config := DefaultPolicyConfig()

	require.NotNil(t, config)
	assert.True(t, config.DefaultDeny)
	assert.NotEmpty(t, config.Resources)

	// Check that default resources are present
	assert.Contains(t, config.Resources, "workflow")
	assert.Contains(t, config.Resources, "execution")
	assert.Contains(t, config.Resources, "schedule")
	assert.Contains(t, config.Resources, "log")
	assert.Contains(t, config.Resources, "user")

	// Check workflow policy
	workflowPolicy := config.Resources["workflow"]
	assert.Equal(t, "workflow", workflowPolicy.Name)
	assert.NotEmpty(t, workflowPolicy.Description)
	assert.Contains(t, workflowPolicy.Actions, "create")
	assert.Contains(t, workflowPolicy.Actions, "read")
	assert.Contains(t, workflowPolicy.Actions, "update")
	assert.Contains(t, workflowPolicy.Actions, "delete")
	assert.Contains(t, workflowPolicy.Actions, "execute")

	// Check that admin has all permissions
	assert.Contains(t, workflowPolicy.Actions["create"], "admin")
	assert.Contains(t, workflowPolicy.Actions["read"], "admin")
	assert.Contains(t, workflowPolicy.Actions["update"], "admin")
	assert.Contains(t, workflowPolicy.Actions["delete"], "admin")

	// Check that viewer can only read
	assert.Contains(t, workflowPolicy.Actions["read"], "viewer")
	assert.NotContains(t, workflowPolicy.Actions["create"], "viewer")
	assert.NotContains(t, workflowPolicy.Actions["update"], "viewer")
	assert.NotContains(t, workflowPolicy.Actions["delete"], "viewer")
}

func TestLoadPolicyConfigFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			fileContent: `{
				"default_deny": true,
				"resources": {
					"workflow": {
						"name": "workflow",
						"actions": {
							"create": ["admin", "developer"],
							"read": ["admin", "developer", "viewer"]
						}
					}
				}
			}`,
			wantErr: false,
		},
		{
			name:        "invalid json",
			fileContent: `{invalid json}`,
			wantErr:     true,
			errContains: "failed to parse policy config",
		},
		{
			name:        "empty file",
			fileContent: ``,
			wantErr:     true,
			errContains: "failed to parse policy config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filePath := filepath.Join(tempDir, tt.name+".json")
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			require.NoError(t, err)

			// Load config
			config, err := LoadPolicyConfigFromFile(filePath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.True(t, config.DefaultDeny)
				assert.Contains(t, config.Resources, "workflow")
			}
		})
	}
}

func TestLoadPolicyConfigFromFile_NonExistentFile(t *testing.T) {
	_, err := LoadPolicyConfigFromFile("/nonexistent/path/config.json")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read policy config file")
}

func TestSavePolicyConfigToFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "policy.json")

	config := &PolicyConfig{
		DefaultDeny: true,
		Resources: map[string]ResourcePolicy{
			"workflow": {
				Name: "workflow",
				Actions: map[string][]string{
					"create": {"admin"},
					"read":   {"admin", "viewer"},
				},
			},
		},
	}

	err := SavePolicyConfigToFile(config, filePath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Load and verify content
	loadedConfig, err := LoadPolicyConfigFromFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, config.DefaultDeny, loadedConfig.DefaultDeny)
	assert.Equal(t, len(config.Resources), len(loadedConfig.Resources))
}

func TestNewPolicyManagerFromConfig(t *testing.T) {
	logger := zap.NewNop()
	config := &PolicyConfig{
		DefaultDeny: true,
		Resources: map[string]ResourcePolicy{
			"workflow": {
				Name: "workflow",
				Actions: map[string][]string{
					"create": {"admin", "developer"},
					"read":   {"admin", "developer", "viewer"},
				},
			},
			"execution": {
				Name: "execution",
				Actions: map[string][]string{
					"create": {"admin"},
					"read":   {"admin", "viewer"},
				},
			},
		},
	}

	pm := NewPolicyManagerFromConfig(config, logger)

	require.NotNil(t, pm)
	assert.Equal(t, config.DefaultDeny, pm.defaultDeny)

	// Verify policies were loaded
	workflowPolicy := pm.GetPolicy("workflow")
	assert.NotNil(t, workflowPolicy)
	assert.Equal(t, "workflow", workflowPolicy.Resource)

	executionPolicy := pm.GetPolicy("execution")
	assert.NotNil(t, executionPolicy)
	assert.Equal(t, "execution", executionPolicy.Resource)
}

func TestLoadPolicyManagerFromFile(t *testing.T) {
	logger := zap.NewNop()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "policy.json")

	// Create test config file
	config := DefaultPolicyConfig()
	err := SavePolicyConfigToFile(config, filePath)
	require.NoError(t, err)

	// Load policy manager
	pm, err := LoadPolicyManagerFromFile(filePath, logger)

	require.NoError(t, err)
	require.NotNil(t, pm)

	// Verify policies were loaded
	workflowPolicy := pm.GetPolicy("workflow")
	assert.NotNil(t, workflowPolicy)
}

func TestLoadPolicyManagerFromFile_NonExistentFile(t *testing.T) {
	logger := zap.NewNop()

	pm, err := LoadPolicyManagerFromFile("/nonexistent/path/config.json", logger)

	assert.Error(t, err)
	assert.Nil(t, pm)
}

func TestValidatePolicyConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *PolicyConfig
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil config",
			config:      nil,
			wantErr:     true,
			errContains: "policy config cannot be nil",
		},
		{
			name: "empty resources",
			config: &PolicyConfig{
				DefaultDeny: true,
				Resources:   map[string]ResourcePolicy{},
			},
			wantErr:     true,
			errContains: "must have at least one resource",
		},
		{
			name: "empty resource name",
			config: &PolicyConfig{
				DefaultDeny: true,
				Resources: map[string]ResourcePolicy{
					"": {
						Name: "",
						Actions: map[string][]string{
							"create": {"admin"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "resource name cannot be empty",
		},
		{
			name: "empty actions",
			config: &PolicyConfig{
				DefaultDeny: true,
				Resources: map[string]ResourcePolicy{
					"workflow": {
						Name:    "workflow",
						Actions: map[string][]string{},
					},
				},
			},
			wantErr:     true,
			errContains: "must have at least one action",
		},
		{
			name: "empty action name",
			config: &PolicyConfig{
				DefaultDeny: true,
				Resources: map[string]ResourcePolicy{
					"workflow": {
						Name: "workflow",
						Actions: map[string][]string{
							"": {"admin"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "action name cannot be empty",
		},
		{
			name: "empty roles",
			config: &PolicyConfig{
				DefaultDeny: true,
				Resources: map[string]ResourcePolicy{
					"workflow": {
						Name: "workflow",
						Actions: map[string][]string{
							"create": {},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "must have at least one role",
		},
		{
			name: "empty role name",
			config: &PolicyConfig{
				DefaultDeny: true,
				Resources: map[string]ResourcePolicy{
					"workflow": {
						Name: "workflow",
						Actions: map[string][]string{
							"create": {"admin", ""},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "role name cannot be empty",
		},
		{
			name:    "valid config",
			config:  DefaultPolicyConfig(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePolicyConfig(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergePolicyConfigs(t *testing.T) {
	tests := []struct {
		name     string
		configs  []*PolicyConfig
		validate func(t *testing.T, merged *PolicyConfig)
	}{
		{
			name:    "no configs returns default",
			configs: []*PolicyConfig{},
			validate: func(t *testing.T, merged *PolicyConfig) {
				assert.NotNil(t, merged)
				assert.True(t, merged.DefaultDeny)
				assert.NotEmpty(t, merged.Resources)
			},
		},
		{
			name: "single config",
			configs: []*PolicyConfig{
				{
					DefaultDeny: true,
					Resources: map[string]ResourcePolicy{
						"workflow": {
							Name: "workflow",
							Actions: map[string][]string{
								"create": {"admin"},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, merged *PolicyConfig) {
				assert.True(t, merged.DefaultDeny)
				assert.Contains(t, merged.Resources, "workflow")
			},
		},
		{
			name: "merge two configs",
			configs: []*PolicyConfig{
				{
					DefaultDeny: true,
					Resources: map[string]ResourcePolicy{
						"workflow": {
							Name: "workflow",
							Actions: map[string][]string{
								"create": {"admin"},
							},
						},
					},
				},
				{
					DefaultDeny: false,
					Resources: map[string]ResourcePolicy{
						"execution": {
							Name: "execution",
							Actions: map[string][]string{
								"create": {"admin", "developer"},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, merged *PolicyConfig) {
				assert.False(t, merged.DefaultDeny) // Second config overrides
				assert.Contains(t, merged.Resources, "workflow")
				assert.Contains(t, merged.Resources, "execution")
			},
		},
		{
			name: "override resource",
			configs: []*PolicyConfig{
				{
					DefaultDeny: true,
					Resources: map[string]ResourcePolicy{
						"workflow": {
							Name: "workflow",
							Actions: map[string][]string{
								"create": {"admin"},
							},
						},
					},
				},
				{
					DefaultDeny: true,
					Resources: map[string]ResourcePolicy{
						"workflow": {
							Name: "workflow",
							Actions: map[string][]string{
								"create": {"admin", "developer"},
								"read":   {"admin", "developer", "viewer"},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, merged *PolicyConfig) {
				assert.Contains(t, merged.Resources, "workflow")
				workflowPolicy := merged.Resources["workflow"]
				assert.Contains(t, workflowPolicy.Actions, "create")
				assert.Contains(t, workflowPolicy.Actions, "read")
				assert.Contains(t, workflowPolicy.Actions["create"], "developer")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := MergePolicyConfigs(tt.configs...)
			require.NotNil(t, merged)
			tt.validate(t, merged)
		})
	}
}

