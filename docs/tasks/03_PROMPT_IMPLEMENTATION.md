# GoFlow AI Implementation Prompt Template

## Purpose

This document provides a comprehensive, reusable prompt template for AI-assisted code generation when implementing tasks from the GoFlow Implementation Roadmap. Follow this template to ensure consistent, high-quality implementations that meet all project requirements.

---

## How to Use This Template

1. Copy the entire "AI Implementation Prompt" section below
2. Replace all placeholders (marked with `{{PLACEHOLDER}}`) with actual values from the roadmap
3. Add any task-specific requirements or constraints
4. Provide the completed prompt to your AI coding assistant
5. Review and validate the generated code using the validation checklist

---

## AI Implementation Prompt

```markdown
# Task Implementation Request: {{TASK_ID}} - {{COMPONENT_NAME}}

## Context

You are implementing a task for the GoFlow Workflow Engine, a production-grade workflow orchestration system built in Go. This is part of a larger implementation roadmap with 238 tasks organized into 15 phases.

### Project Information
- **Project**: GoFlow Workflow Engine
- **Language**: Go 1.21+
- **Architecture**: Layered architecture (API → Service → Engine → Domain → Repository → Infrastructure)
- **Testing Framework**: testify
- **Documentation**: GoDoc standard

### Task Details
- **Task ID**: {{TASK_ID}}
- **Component**: {{COMPONENT_NAME}}
- **Phase**: {{PHASE_NUMBER}} - {{PHASE_NAME}}
- **Priority**: {{PRIORITY}} (P0=Critical, P1=High, P2=Medium, P3=Low)
- **Status**: ⭕ Not Started → 🚧 In Progress
- **Dependencies**: {{DEPENDENCIES}}
  - Must complete before starting: {{BLOCKING_DEPENDENCIES}}
  - Related tasks: {{RELATED_TASKS}}

### Task Description
{{TASK_DESCRIPTION}}

### Acceptance Criteria
{{ACCEPTANCE_CRITERIA}}

---

## Architecture Context

### GoFlow Project Structure
```
goflow-service/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/             # API layer (handlers, middleware, DTOs)
│   ├── service/         # Service layer (business logic)
│   ├── engine/          # Workflow engine (DAG, execution, expressions)
│   ├── domain/          # Domain models and validation
│   ├── repository/      # Data access layer
│   └── infrastructure/  # External integrations (Redis, Kafka, etc.)
├── test/
│   ├── unit/            # Unit tests
│   ├── integration/     # Integration tests
│   └── e2e/             # End-to-end tests
├── migrations/          # Database migrations
└── docs/                # Documentation
```

### Layer Responsibilities
- **API Layer**: HTTP routing, request/response handling, validation, middleware
- **Service Layer**: Business logic orchestration, transaction management
- **Engine Layer**: Workflow execution, DAG resolution, expression evaluation
- **Domain Layer**: Business entities, domain logic, validation rules
- **Repository Layer**: Database operations, query building
- **Infrastructure Layer**: External service integrations (Redis, Kafka, OpenAI, etc.)

### Target Layer for This Task
**{{TARGET_LAYER}}** (e.g., Repository Layer, Service Layer, etc.)

---

## Implementation Requirements

### 1. File Locations

Create/modify the following files:

**Primary Implementation Files**:
```
{{PRIMARY_FILES}}
Example:
- internal/repository/workflow_repository.go
- internal/repository/workflow_repository_test.go
```

**Supporting Files** (if needed):
```
{{SUPPORTING_FILES}}
Example:
- internal/domain/workflow.go (if new domain model needed)
- migrations/000001_create_workflows_table.up.sql
```

### 2. Required Interfaces and Contracts

Implement the following interfaces/contracts:

```go
{{INTERFACE_DEFINITIONS}}

Example:
type WorkflowRepository interface {
    FindByID(ctx context.Context, id string) (*domain.Workflow, error)
    Save(ctx context.Context, workflow *domain.Workflow) error
    Update(ctx context.Context, workflow *domain.Workflow) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filters ListFilters) ([]*domain.Workflow, error)
}
```

### 3. Dependencies and Imports

Required dependencies:
```
{{REQUIRED_DEPENDENCIES}}

Example:
- github.com/gin-gonic/gin
- github.com/stretchr/testify
- github.com/lib/pq (PostgreSQL driver)
```

Standard imports to use:
```go
{{STANDARD_IMPORTS}}

Example:
import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "goflow-service/internal/domain"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

### 4. Configuration Requirements

Add the following configuration (if applicable):

**Environment Variables**:
```
{{ENV_VARIABLES}}

Example:
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goflow_db
```

**Config Struct Fields**:
```go
{{CONFIG_FIELDS}}

Example:
type DatabaseConfig struct {
    Host     string `mapstructure:"db_host"`
    Port     int    `mapstructure:"db_port"`
    Name     string `mapstructure:"db_name"`
}
```

---

## Code Generation Instructions

### Step 1: Understand the Context

Before writing code, review:
1. ✅ Task dependencies are complete (check {{DEPENDENCIES}})
2. ✅ Related documentation sections:
   - {{DOC_REFERENCES}}
3. ✅ Existing code patterns in similar components
4. ✅ Domain models and their relationships

### Step 2: Implement Core Functionality

Follow these guidelines:

**Code Style**:
- Use Go standard formatting (gofmt)
- Follow effective Go conventions
- Use meaningful variable names (no single-letter names except for loops)
- Keep functions small and focused (max 50 lines)
- Use early returns to reduce nesting

**Error Handling**:
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to {{action}}: %w", err)
}

// Use custom error types for domain errors
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}
```

**Logging**:
```go
// Use structured logging with Zap
logger.Info("{{action}} completed",
    zap.String("{{entity}}_id", id),
    zap.Duration("duration", time.Since(start)),
)

logger.Error("{{action}} failed",
    zap.String("{{entity}}_id", id),
    zap.Error(err),
)
```

**Context Handling**:
```go
// Always accept context as first parameter
func (r *Repository) FindByID(ctx context.Context, id string) (*domain.Entity, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Use context in database calls
    row := r.db.QueryRowContext(ctx, query, id)
    // ...
}
```

### Step 3: Implement Required Methods

For each method in the interface:

1. **Add GoDoc comment** describing what the method does
2. **Validate inputs** (nil checks, empty strings, etc.)
3. **Implement core logic** following the patterns above
4. **Handle errors** with proper wrapping and context
5. **Add logging** for important operations
6. **Return appropriate values**

Example:
```go
// FindByID retrieves a workflow by its unique identifier.
// Returns NotFoundError if the workflow does not exist.
func (r *WorkflowRepository) FindByID(ctx context.Context, id string) (*domain.Workflow, error) {
    if id == "" {
        return nil, fmt.Errorf("workflow ID cannot be empty")
    }

    query := `SELECT id, name, version, definition, created_at, updated_at
              FROM workflows WHERE id = $1 AND deleted_at IS NULL`

    var workflow domain.Workflow
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &workflow.ID,
        &workflow.Name,
        &workflow.Version,
        &workflow.Definition,
        &workflow.CreatedAt,
        &workflow.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, &NotFoundError{Resource: "workflow", ID: id}
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query workflow: %w", err)
    }

    return &workflow, nil
}
```

### Step 4: Add Comprehensive Tests

Create test file: `{{TEST_FILE_PATH}}`

**Test Structure**:
```go
package {{PACKAGE_NAME}}_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)

// Test naming: Test{{StructName}}_{{MethodName}}_{{Scenario}}
func Test{{STRUCT_NAME}}_{{METHOD_NAME}}_{{SCENARIO}}(t *testing.T) {
    // Arrange
    {{SETUP_CODE}}

    // Act
    {{EXECUTION_CODE}}

    // Assert
    {{ASSERTIONS}}
}
```

**Required Test Cases**:
1. ✅ Happy path (success case)
2. ✅ Error cases (invalid input, not found, database errors)
3. ✅ Edge cases (empty values, nil pointers, boundary conditions)
4. ✅ Context cancellation
5. ✅ Concurrent access (if applicable)

**Use Table-Driven Tests** for multiple scenarios:
```go
func Test{{STRUCT_NAME}}_{{METHOD_NAME}}(t *testing.T) {
    tests := []struct {
        name    string
        input   {{INPUT_TYPE}}
        want    {{OUTPUT_TYPE}}
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid input",
            input:   {{VALID_INPUT}},
            want:    {{EXPECTED_OUTPUT}},
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   {{INVALID_INPUT}},
            want:    nil,
            wantErr: true,
            errMsg:  "expected error message",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Mock Dependencies**:
```go
type Mock{{DEPENDENCY_NAME}} struct {
    mock.Mock
}

func (m *Mock{{DEPENDENCY_NAME}}) {{METHOD_NAME}}({{PARAMS}}) {{RETURNS}} {
    args := m.Called({{PARAM_NAMES}})
    return args.Get(0).({{RETURN_TYPE}}), args.Error(1)
}
```

**Target Coverage**: Minimum 80% code coverage

---

## Testing Requirements

### Unit Tests

**File**: `{{TEST_FILE_PATH}}`

**Required Test Functions**:
```
{{TEST_FUNCTIONS}}

Example:
- TestWorkflowRepository_FindByID_Success
- TestWorkflowRepository_FindByID_NotFound
- TestWorkflowRepository_FindByID_DatabaseError
- TestWorkflowRepository_Save_Success
- TestWorkflowRepository_Save_DuplicateID
```

**Test Helpers**:
```go
// Create test fixtures
func createTestWorkflow() *domain.Workflow {
    return &domain.Workflow{
        ID:      "wf_test_123",
        Name:    "Test Workflow",
        Version: 1,
    }
}

// Setup test database (for integration tests)
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", testDSN)
    require.NoError(t, err)

    t.Cleanup(func() {
        db.Close()
    })

    return db
}
```

### Integration Tests (if applicable)

**File**: `test/integration/{{COMPONENT_NAME}}_test.go`

**Build Tag**: Add `// +build integration` at the top

**Requirements**:
- Test with real database/cache/external services
- Use test containers or docker-compose for dependencies
- Clean up test data after each test
- Test transaction rollback scenarios

---

## Documentation Requirements

### 1. GoDoc Comments

**Package-level documentation** (at top of main file):
```go
// Package {{PACKAGE_NAME}} provides {{DESCRIPTION}}.
//
// {{DETAILED_DESCRIPTION}}
//
// Example usage:
//
//     {{EXAMPLE_CODE}}
//
package {{PACKAGE_NAME}}
```

**Function/Method documentation**:
```go
// {{METHOD_NAME}} {{DESCRIPTION}}.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - {{PARAM_NAME}}: {{PARAM_DESCRIPTION}}
//
// Returns:
//   - {{RETURN_DESCRIPTION}}
//   - error: {{ERROR_DESCRIPTION}}
//
// Example:
//
//     {{EXAMPLE_CODE}}
//
func {{METHOD_NAME}}({{PARAMETERS}}) {{RETURNS}} {
    // Implementation
}
```

### 2. Code Examples

Add examples for complex functionality:
```go
func Example{{FUNCTION_NAME}}() {
    {{EXAMPLE_CODE}}
    // Output:
    // {{EXPECTED_OUTPUT}}
}
```

---

## Post-Implementation Actions

### 1. Update Task Tracking

**File**: `docs/tasks/01_IMPLEMENTATION_ROADMAP.md`

Find the task entry and update status:
```markdown
| {{TASK_ID}} | {{COMPONENT_NAME}} | {{DESCRIPTION}} | {{PRIORITY}} | ✅ | {{DEPENDENCIES}} | 100% |
```

Update phase coverage:
```markdown
**Phase {{PHASE_NUMBER}} Coverage**: {{COMPLETED}}/{{TOTAL}} tasks complete ({{PERCENTAGE}}%)
```

**File**: `docs/tasks/02_QUICK_REFERENCE.md`

Add to "Completed This Week" section:
```markdown
| {{TASK_ID}} | {{COMPONENT_NAME}} | {{YOUR_NAME}} | {{DATE}} |
```

Update progress bars and metrics.

### 2. Create Implementation Summary

**File**: `docs/summaries/{{TASK_ID}}_{{COMPONENT_NAME}}_SUMMARY.md`

Use this template:
```markdown
# {{TASK_ID}}: {{COMPONENT_NAME}} Implementation Summary

## Overview
- **Task ID**: {{TASK_ID}}
- **Component**: {{COMPONENT_NAME}}
- **Implemented By**: {{DEVELOPER_NAME}}
- **Date**: {{DATE}}
- **Status**: ✅ Complete

## What Was Implemented

### Core Functionality
{{DESCRIPTION_OF_IMPLEMENTATION}}

### Key Features
- Feature 1
- Feature 2
- Feature 3

## Files Created/Modified

### New Files
- `{{FILE_PATH}}` - {{DESCRIPTION}}
- `{{FILE_PATH}}` - {{DESCRIPTION}}

### Modified Files
- `{{FILE_PATH}}` - {{CHANGES_MADE}}

### Total Lines of Code
- Implementation: {{LOC}} lines
- Tests: {{TEST_LOC}} lines
- Total: {{TOTAL_LOC}} lines

## Test Coverage

### Unit Tests
- Test file: `{{TEST_FILE}}`
- Test functions: {{TEST_COUNT}}
- Coverage: {{COVERAGE_PERCENTAGE}}%
- All tests passing: ✅

### Integration Tests (if applicable)
- Test file: `{{INTEGRATION_TEST_FILE}}`
- Test scenarios: {{SCENARIO_COUNT}}
- All tests passing: ✅

## Dependencies

### Completed Dependencies
- {{DEP_1}} ✅
- {{DEP_2}} ✅

### Enables These Tasks
- {{ENABLED_TASK_1}}
- {{ENABLED_TASK_2}}

## Deviations from Original Plan

{{DEVIATIONS_OR_NONE}}

## Challenges and Solutions

### Challenge 1
**Problem**: {{PROBLEM_DESCRIPTION}}
**Solution**: {{SOLUTION_DESCRIPTION}}

## Next Steps

1. {{NEXT_STEP_1}}
2. {{NEXT_STEP_2}}
3. {{NEXT_STEP_3}}

## Related Documentation
- [{{DOC_NAME}}]({{DOC_PATH}})
- [{{DOC_NAME}}]({{DOC_PATH}})
```

---

## Validation Checklist

Before marking the task as complete, verify:

### Code Quality
- [ ] Code compiles without errors: `go build ./...`
- [ ] All tests pass: `go test ./...`
- [ ] Test coverage ≥ 80%: `go test -cover ./...`
- [ ] Linting passes: `golangci-lint run`
- [ ] No race conditions: `go test -race ./...`
- [ ] Code formatted: `gofmt -w .` or `go fmt ./...`

### Documentation
- [ ] All exported functions have GoDoc comments
- [ ] Package-level documentation exists
- [ ] Complex logic has inline comments
- [ ] Examples provided for public APIs
- [ ] README updated (if applicable)

### Testing
- [ ] Unit tests cover happy path
- [ ] Unit tests cover error cases
- [ ] Unit tests cover edge cases
- [ ] Integration tests pass (if applicable)
- [ ] Test fixtures are clean and reusable
- [ ] Mocks are properly implemented

### Architecture
- [ ] Follows GoFlow layered architecture
- [ ] Proper separation of concerns
- [ ] Dependencies injected correctly
- [ ] Error handling is consistent
- [ ] Logging is structured and meaningful
- [ ] Context is propagated correctly

### Task Tracking
- [ ] Task status updated in roadmap (⭕ → ✅)
- [ ] Phase coverage percentage updated
- [ ] Quick reference updated
- [ ] Implementation summary created
- [ ] Related tasks identified

### Performance
- [ ] No obvious performance issues
- [ ] Database queries are optimized (if applicable)
- [ ] Proper use of indexes (if applicable)
- [ ] Connection pooling configured (if applicable)
- [ ] Caching implemented where appropriate (if applicable)

### Security
- [ ] Input validation implemented
- [ ] SQL injection prevented (parameterized queries)
- [ ] Secrets not hardcoded
- [ ] Proper authentication/authorization (if applicable)
- [ ] Error messages don't leak sensitive info

---

## Example: Completed Prompt for DB-001

Here's an example of how to fill out this template for task DB-001:

```markdown
# Task Implementation Request: DB-001 - Schema Setup

## Context

You are implementing a task for the GoFlow Workflow Engine, a production-grade workflow orchestration system built in Go.

### Task Details
- **Task ID**: DB-001
- **Component**: Database Schema
- **Phase**: 1 - Foundation & Core Infrastructure
- **Priority**: P0 (Critical)
- **Status**: ⭕ Not Started → 🚧 In Progress
- **Dependencies**: None (this is a foundational task)

### Task Description
Create all database tables for the GoFlow system including: workflows, workflow_executions, node_executions, execution_logs, workflow_schedules, webhooks, and users tables.

### Acceptance Criteria
- All 7 tables created with proper columns and data types
- Primary keys, foreign keys, and constraints defined
- Indexes created for performance
- Soft delete support (deleted_at column)
- Timestamps (created_at, updated_at) on all tables
- JSONB columns for flexible data storage
- Migration files created using golang-migrate

---

## Architecture Context

### Target Layer for This Task
**Database Layer** - Foundation for all data persistence

---

## Implementation Requirements

### 1. File Locations

**Primary Implementation Files**:
```
- migrations/000001_create_workflows_table.up.sql
- migrations/000001_create_workflows_table.down.sql
- migrations/000002_create_workflow_executions_table.up.sql
- migrations/000002_create_workflow_executions_table.down.sql
- migrations/000003_create_node_executions_table.up.sql
- migrations/000003_create_node_executions_table.down.sql
- migrations/000004_create_execution_logs_table.up.sql
- migrations/000004_create_execution_logs_table.down.sql
- migrations/000005_create_workflow_schedules_table.up.sql
- migrations/000005_create_workflow_schedules_table.down.sql
- migrations/000006_create_webhooks_table.up.sql
- migrations/000006_create_webhooks_table.down.sql
- migrations/000007_create_users_table.up.sql
- migrations/000007_create_users_table.down.sql
```

### 2. Required Interfaces and Contracts

N/A - This is a database schema task

### 3. Dependencies and Imports

Required tools:
```
- golang-migrate/migrate (for running migrations)
- PostgreSQL 14+
```

### 4. Configuration Requirements

**Environment Variables**:
```
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goflow_db
DB_USER=goflow
DB_PASSWORD=goflow_password
DB_SSL_MODE=disable
```

[... rest of the filled template ...]
```

---

## Tips for Success

1. **Start Small**: Implement the minimum viable functionality first, then iterate
2. **Test Early**: Write tests as you implement, not after
3. **Review Dependencies**: Make sure all dependent tasks are truly complete
4. **Follow Patterns**: Look at existing code for similar components
5. **Ask Questions**: If requirements are unclear, clarify before implementing
6. **Document Decisions**: Note any architectural decisions in the summary
7. **Keep It Simple**: Avoid over-engineering; implement what's needed
8. **Performance Later**: Focus on correctness first, optimize later (unless P0 requirement)

---

## Common Patterns by Layer

### Repository Layer Pattern
```go
type WorkflowRepository struct {
    db     *sql.DB
    logger *zap.Logger
}

func NewWorkflowRepository(db *sql.DB, logger *zap.Logger) *WorkflowRepository {
    return &WorkflowRepository{
        db:     db,
        logger: logger,
    }
}

func (r *WorkflowRepository) FindByID(ctx context.Context, id string) (*domain.Workflow, error) {
    // Implementation
}
```

### Service Layer Pattern
```go
type WorkflowService struct {
    repo   repository.WorkflowRepository
    cache  cache.Cache
    logger *zap.Logger
}

func NewWorkflowService(repo repository.WorkflowRepository, cache cache.Cache, logger *zap.Logger) *WorkflowService {
    return &WorkflowService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}

func (s *WorkflowService) GetByID(ctx context.Context, id string) (*domain.Workflow, error) {
    // Check cache first
    // If not in cache, fetch from repository
    // Store in cache
    // Return result
}
```

### API Handler Pattern
```go
type WorkflowHandler struct {
    service service.WorkflowService
    logger  *zap.Logger
}

func NewWorkflowHandler(service service.WorkflowService, logger *zap.Logger) *WorkflowHandler {
    return &WorkflowHandler{
        service: service,
        logger:  logger,
    }
}

func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
    id := c.Param("id")

    workflow, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, workflow)
}
```

---

## Troubleshooting

### Common Issues

**Issue**: Tests fail with "cannot find package"
**Solution**: Run `go mod tidy` to update dependencies

**Issue**: Linter complains about unused variables
**Solution**: Remove unused variables or prefix with underscore: `_unused`

**Issue**: Race detector finds issues
**Solution**: Use proper synchronization (mutexes, channels) or atomic operations

**Issue**: Coverage is below 80%
**Solution**: Add more test cases, especially for error paths and edge cases

**Issue**: Database tests fail
**Solution**: Ensure test database is running and migrations are applied

---

## Quick Reference: Placeholder List

Copy this list when filling out the template:

```
{{TASK_ID}}
{{COMPONENT_NAME}}
{{PHASE_NUMBER}}
{{PHASE_NAME}}
{{PRIORITY}}
{{DEPENDENCIES}}
{{BLOCKING_DEPENDENCIES}}
{{RELATED_TASKS}}
{{TASK_DESCRIPTION}}
{{ACCEPTANCE_CRITERIA}}
{{TARGET_LAYER}}
{{PRIMARY_FILES}}
{{SUPPORTING_FILES}}
{{INTERFACE_DEFINITIONS}}
{{REQUIRED_DEPENDENCIES}}
{{STANDARD_IMPORTS}}
{{ENV_VARIABLES}}
{{CONFIG_FIELDS}}
{{DOC_REFERENCES}}
{{TEST_FILE_PATH}}
{{PACKAGE_NAME}}
{{STRUCT_NAME}}
{{METHOD_NAME}}
{{SCENARIO}}
{{TEST_FUNCTIONS}}
{{DEVELOPER_NAME}}
{{DATE}}
```

---

**Template Version**: 1.0
**Last Updated**: 2024-01-01
**Maintained By**: Tech Lead
**Feedback**: Submit improvements via pull request
```
