# API Layer

This directory contains the complete API layer scaffolding for the GoFlow Workflow Engine, generated from the OpenAPI specification defined in `docs/swagger.yml`.

## Directory Structure

```
internal/api/
├── dto/                    # Data Transfer Objects (Request/Response models)
│   ├── common_dto.go      # Common DTOs (ErrorResponse, Pagination, Health, Logs)
│   ├── workflow_dto.go    # Workflow-related DTOs
│   ├── execution_dto.go   # Execution-related DTOs
│   ├── schedule_dto.go    # Schedule-related DTOs
│   └── webhook_dto.go     # Webhook-related DTOs
├── handler/               # HTTP request handlers
│   ├── health_handler.go     # Health check endpoint
│   ├── workflow_handler.go   # Workflow CRUD and execution
│   ├── execution_handler.go  # Execution monitoring and control
│   ├── schedule_handler.go   # Schedule management
│   └── webhook_handler.go    # Webhook registration
├── middleware/            # HTTP middleware
│   └── auth.go           # JWT authentication middleware
├── router.go             # Route registration and setup
└── README.md             # This file
```

## Implemented Endpoints

### Health
- `GET /health` - Health check (public)

### Workflows
- `GET /api/v1/workflows` - List workflows with pagination and filtering
- `POST /api/v1/workflows` - Create a new workflow
- `GET /api/v1/workflows/:id` - Get workflow by ID
- `PUT /api/v1/workflows/:id` - Update workflow
- `DELETE /api/v1/workflows/:id` - Delete workflow (soft/hard delete)
- `POST /api/v1/workflows/:id/execute` - Execute workflow (sync/async)

### Executions
- `GET /api/v1/executions/:id` - Get execution details
- `POST /api/v1/executions/:id` - Cancel execution
- `GET /api/v1/executions/:id/logs` - Get execution logs with filtering

### Schedules
- `GET /api/v1/schedules` - List all schedules
- `POST /api/v1/schedules` - Create a new schedule
- `GET /api/v1/schedules/:id` - Get schedule by ID
- `DELETE /api/v1/schedules/:id` - Delete schedule

### Webhooks
- `POST /api/v1/webhooks` - Register a webhook
- `GET /api/v1/webhooks/:id` - Get webhook details
- `DELETE /api/v1/webhooks/:id` - Unregister webhook

## Features

### Request Validation
All handlers include:
- JSON binding with validation tags
- UUID format validation for path parameters
- Query parameter parsing with defaults
- Proper error responses with standard format

### Response Structure
All responses follow the OpenAPI specification:
- Success responses with appropriate HTTP status codes (200, 201, 202, 204)
- Error responses with standard `ErrorResponse` structure
- Pagination metadata for list endpoints
- Proper JSON serialization with field tags

### Mock Implementations
All handlers currently return mock/stub responses that match the expected schema:
- Realistic sample data
- Proper status codes
- Complete response structures
- TODO comments indicating where business logic should be implemented

## Usage

### Setting Up Routes

```go
package main

import (
    "github.com/goflow-atom/goflow-service/internal/api"
)

func main() {
    // Create router with version
    router := api.NewRouter("1.0.0")
    
    // Setup all routes
    engine := router.Setup()
    
    // Start server
    engine.Run(":8080")
}
```

### Implementing Business Logic

To implement actual business logic, inject services into handlers:

```go
// In handler constructor
func NewWorkflowHandler(workflowService service.WorkflowService) *WorkflowHandler {
    return &WorkflowHandler{
        workflowService: workflowService,
    }
}

// In handler method
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
    var req dto.CreateWorkflowRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ... error handling
    }
    
    // Call service layer
    workflow, err := h.workflowService.Create(c.Request.Context(), req)
    if err != nil {
        // ... error handling
    }
    
    c.JSON(http.StatusCreated, workflow)
}
```

### Enabling Authentication

Uncomment the authentication middleware in `router.go`:

```go
// In router.go Setup() method
v1 := r.engine.Group("/api/v1")
{
    // Uncomment to enable authentication
    v1.Use(middleware.AuthMiddleware())
    
    // ... routes
}
```

## Next Steps

1. **Implement Service Layer**: Create service interfaces and implementations for business logic
2. **JWT Authentication**: Complete the JWT validation in `middleware/auth.go`
3. **Database Integration**: Connect handlers to repository layer for data persistence
4. **Error Handling**: Enhance error handling with custom error types and proper logging
5. **Validation**: Add more sophisticated validation rules using validator package
6. **Testing**: Write unit tests for all handlers
7. **Documentation**: Generate Swagger documentation using swag annotations

## OpenAPI Compliance

All endpoints, request/response structures, and validation rules are generated from the OpenAPI specification in `docs/swagger.yml`. Any changes to the API contract should be made in the Swagger file first, then reflected in the code.

## Dependencies

Required Go packages:
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/google/uuid` - UUID generation and validation

These should be added to `go.mod`:
```bash
go get github.com/gin-gonic/gin
go get github.com/google/uuid
```

## Notes

- All handlers include Swagger annotations for documentation generation
- UUID validation is performed on all ID path parameters
- Query parameters have sensible defaults (page=1, limit=20, etc.)
- Soft delete is the default for workflow deletion (use `hardDelete=true` for permanent deletion)
- Async execution is supported via the `async` query parameter
- All timestamps use RFC3339 format
- Pagination follows a consistent structure across all list endpoints

