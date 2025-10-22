# Authorization Middleware (MW-002)

## Overview

The Authorization Middleware provides role-based access control (RBAC) for the GoFlow Workflow Engine. It works in conjunction with the Authentication Middleware (MW-001) to enforce fine-grained access control on API endpoints.

## Features

- **Policy-Based Authorization**: Declarative policies mapping resources → actions → roles
- **Flexible Configuration**: Load policies from files or use programmatic defaults
- **Default Deny Strategy**: Secure by default - access must be explicitly granted
- **Comprehensive Logging**: All authorization decisions are logged for audit
- **Integration with JWT**: Seamlessly retrieves user roles from authenticated context

## Quick Start

### 1. Basic Usage with Default Policies

```go
import (
    "github.com/goflow-atom/goflow-service/internal/api/middleware"
    "go.uber.org/zap"
)

// Create policy manager with default policies
logger := zap.NewProduction()
pm := middleware.NewPolicyManagerFromConfig(
    middleware.DefaultPolicyConfig(),
    logger,
)

// Protect endpoints
router.POST("/workflows",
    auth.GetMiddleware(),  // MW-001: Authentication
    middleware.RequireResourceAction(pm, "workflow", "create"),
    handler.CreateWorkflow,
)

router.GET("/workflows/:id",
    auth.GetMiddleware(),
    middleware.RequireResourceAction(pm, "workflow", "read"),
    handler.GetWorkflow,
)

router.DELETE("/workflows/:id",
    auth.GetMiddleware(),
    middleware.RequireResourceAction(pm, "workflow", "delete"),
    handler.DeleteWorkflow,
)
```

### 2. Loading Policies from File

```go
// Load from configuration file
pm, err := middleware.LoadPolicyManagerFromFile(
    "./config/authz_policy.json",
    logger,
)
if err != nil {
    log.Fatal(err)
}

// Use in routes
router.Use(auth.GetMiddleware())
router.POST("/workflows", 
    middleware.RequireResourceAction(pm, "workflow", "create"),
    handler.CreateWorkflow,
)
```

### 3. Custom Policy Configuration

```go
config := &middleware.PolicyConfig{
    DefaultDeny: true,
    Resources: map[string]middleware.ResourcePolicy{
        "custom_resource": {
            Name: "custom_resource",
            Description: "My custom resource",
            Actions: map[string][]string{
                "special_action": {"admin", "power_user"},
                "read": {"admin", "power_user", "viewer"},
            },
        },
    },
}

pm := middleware.NewPolicyManagerFromConfig(config, logger)
```

## Default Policies

The middleware comes with sensible default policies for GoFlow resources:

### Workflow
- **create**: admin, developer
- **read**: admin, developer, operator, viewer
- **update**: admin, developer
- **delete**: admin
- **execute**: admin, developer, operator

### Execution
- **create**: admin, developer, operator
- **read**: admin, developer, operator, viewer
- **cancel**: admin, developer, operator
- **retry**: admin, developer, operator
- **delete**: admin

### Schedule
- **create**: admin, developer
- **read**: admin, developer, operator, viewer
- **update**: admin, developer
- **delete**: admin, developer
- **enable**: admin, developer
- **disable**: admin, developer

### Log
- **read**: admin, developer, operator, viewer

### User
- **create**: admin
- **read**: admin, developer, operator, viewer
- **update**: admin
- **delete**: admin

## Role Hierarchy

- **admin**: Full access to all resources and actions
- **developer**: Can create, read, update workflows and executions
- **operator**: Can execute workflows and read resources
- **viewer**: Read-only access to all resources

## Configuration

### Environment Variables

```bash
# Path to authorization policy configuration file (optional)
AUTHZ_POLICY_CONFIG_PATH=./config/authz_policy.json

# Default deny strategy (default: true)
AUTHZ_DEFAULT_DENY=true
```

### Policy File Format

Create a JSON file with the following structure:

```json
{
  "default_deny": true,
  "resources": {
    "resource_name": {
      "name": "resource_name",
      "description": "Resource description",
      "actions": {
        "action_name": ["role1", "role2"]
      }
    }
  }
}
```

See `config/authz_policy.example.json` for a complete example.

## API Reference

### PolicyManager

The core authorization engine.

#### Methods

- `Authorize(ctx context.Context, resource string, action string) error`
  - Checks if the user can perform an action on a resource
  - Returns error if unauthorized

- `HasRole(ctx context.Context, role string) bool`
  - Checks if the user has a specific role

- `GetAllowedActions(ctx context.Context, resource string) ([]string, error)`
  - Returns list of actions the user can perform on a resource

- `AddPolicy(resource string, actions map[string][]string)`
  - Adds or updates a policy for a resource

- `GetPolicy(resource string) *AuthorizationPolicy`
  - Retrieves the policy for a resource

### Middleware Functions

- `RequireResourceAction(pm *PolicyManager, resource, action string) gin.HandlerFunc`
  - Returns middleware that enforces authorization for a resource-action pair

### Configuration Functions

- `DefaultPolicyConfig() *PolicyConfig`
  - Returns default policy configuration

- `LoadPolicyConfigFromFile(filePath string) (*PolicyConfig, error)`
  - Loads policy configuration from JSON file

- `SavePolicyConfigToFile(config *PolicyConfig, filePath string) error`
  - Saves policy configuration to JSON file

- `NewPolicyManagerFromConfig(config *PolicyConfig, logger *zap.Logger) *PolicyManager`
  - Creates policy manager from configuration

- `LoadPolicyManagerFromFile(filePath string, logger *zap.Logger) (*PolicyManager, error)`
  - Loads policy manager from configuration file

- `ValidatePolicyConfig(config *PolicyConfig) error`
  - Validates policy configuration

- `MergePolicyConfigs(configs ...*PolicyConfig) *PolicyConfig`
  - Merges multiple policy configurations

## Advanced Usage

### Merging Policies

Combine base policies with custom overrides:

```go
// Start with defaults
baseConfig := middleware.DefaultPolicyConfig()

// Load custom overrides
customConfig, _ := middleware.LoadPolicyConfigFromFile("./config/custom.json")

// Merge (custom overrides base)
merged := middleware.MergePolicyConfigs(baseConfig, customConfig)
pm := middleware.NewPolicyManagerFromConfig(merged, logger)
```

### Programmatic Policy Management

```go
pm := middleware.NewPolicyManager(true, logger)

// Add policies programmatically
pm.AddPolicy("workflow", map[string][]string{
    "create": {"admin", "developer"},
    "read":   {"admin", "developer", "viewer"},
})

pm.AddPolicy("execution", map[string][]string{
    "create": {"admin", "operator"},
    "read":   {"admin", "operator", "viewer"},
})
```

### Checking Permissions in Handlers

```go
func (h *Handler) CreateWorkflow(c *gin.Context) {
    // Get allowed actions for the user
    actions, err := h.policyManager.GetAllowedActions(c, "workflow")
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to check permissions"})
        return
    }

    // Return allowed actions in response
    c.JSON(200, gin.H{
        "allowed_actions": actions,
    })
}
```

## Error Responses

### 403 Forbidden

When a user lacks the required role:

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "You do not have permission to create workflow",
    "details": {
      "request_id": "abc-123-def"
    }
  }
}
```

## Testing

The middleware includes comprehensive test coverage:

```bash
# Run authorization tests
go test -v ./internal/api/middleware -run "TestPolicyManager|TestRequireResourceAction"

# Check coverage
go test -coverprofile=coverage.out ./internal/api/middleware
go tool cover -func=coverage.out | grep authz.go
```

## Security Best Practices

1. **Always use default deny**: Set `default_deny: true` in production
2. **Explicit permissions**: Grant only necessary permissions to each role
3. **Regular audits**: Review authorization logs regularly
4. **Principle of least privilege**: Start with minimal permissions and add as needed
5. **Test thoroughly**: Test all authorization scenarios before deployment

## Troubleshooting

### Issue: Authorization always denies

**Cause**: No policy defined for resource or action

**Solution**: Check that policy exists and includes the action:
```go
policy := pm.GetPolicy("resource_name")
if policy == nil {
    // Policy not found
}
```

### Issue: User has role but still denied

**Cause**: Role not in allowed roles for action

**Solution**: Verify policy configuration:
```json
{
  "resources": {
    "workflow": {
      "actions": {
        "create": ["admin", "developer"]  // Add user's role here
      }
    }
  }
}
```

### Issue: Authorization not enforced

**Cause**: Middleware not applied to route

**Solution**: Ensure middleware is added:
```go
router.POST("/workflows",
    auth.GetMiddleware(),  // Required: Authentication first
    middleware.RequireResourceAction(pm, "workflow", "create"),  // Then authorization
    handler.CreateWorkflow,
)
```

## Related Documentation

- [MW-001 Authentication Middleware](./auth.go)
- [Implementation Summary](../../../docs/summaries/MW-002_Authorization_Middleware_SUMMARY.md)
- [GoFlow Implementation Roadmap](../../../docs/tasks/01_IMPLEMENTATION_ROADMAP.md)

## Support

For issues or questions:
1. Check the implementation summary: `docs/summaries/MW-002_Authorization_Middleware_SUMMARY.md`
2. Review test cases: `internal/api/middleware/authz_test.go`
3. Check logs for authorization decisions

