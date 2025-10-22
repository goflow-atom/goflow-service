# Code Style Guide

This document defines the coding standards and best practices for GoFlow.

## Table of Contents

- [Go Conventions](#go-conventions)
  - [Naming](#naming)
  - [Function Size](#function-size)
  - [Error Handling](#error-handling)
  - [Comments](#comments)
- [Directory Policy](#directory-policy)
- [Linting Tools](#linting-tools)
- [Code Organization](#code-organization)
- [Best Practices](#best-practices)

## Go Conventions

### Naming

#### Packages

- Use lowercase, single-word names
- Avoid underscores or mixed caps
- Use descriptive names that reflect purpose

```go
// Good
package workflow
package repository
package config

// Bad
package workflowService
package workflow_service
package ws
```

#### Variables

- Use camelCase for local variables
- Use PascalCase for exported variables
- Use descriptive names, avoid abbreviations

```go
// Good
var workflowID string
var executionCount int
var UserRepository repository.UserRepository

// Bad
var wfId string
var cnt int
var usrRepo repository.UserRepository
```

#### Functions

- Use PascalCase for exported functions
- Use camelCase for private functions
- Use verb-noun naming pattern

```go
// Good
func ExecuteWorkflow(ctx context.Context, id string) error
func validateInput(input map[string]interface{}) error

// Bad
func execute_workflow(ctx context.Context, id string) error
func Validate(input map[string]interface{}) error
```

#### Constants

- Use PascalCase for exported constants
- Use camelCase for private constants
- Group related constants

```go
// Good
const (
    StatusPending   = "pending"
    StatusRunning   = "running"
    StatusCompleted = "completed"
    StatusFailed    = "failed"
)

const defaultTimeout = 30 * time.Second

// Bad
const STATUS_PENDING = "pending"
const Default_Timeout = 30
```

#### Interfaces

- Use noun or adjective names
- Single-method interfaces should end with "-er"

```go
// Good
type WorkflowExecutor interface {
    Execute(ctx context.Context) error
}

type NodeValidator interface {
    Validate() error
}

// Bad
type IWorkflow interface {
    Execute(ctx context.Context) error
}

type WorkflowInterface interface {
    Execute(ctx context.Context) error
}
```

### Function Size

- Keep functions small and focused (< 50 lines)
- Each function should do one thing well
- Extract complex logic into helper functions

```go
// Good
func (s *WorkflowService) Execute(ctx context.Context, workflowID string, input map[string]interface{}) (*domain.Execution, error) {
    workflow, err := s.getWorkflow(ctx, workflowID)
    if err != nil {
        return nil, err
    }

    if err := s.validateInput(workflow, input); err != nil {
        return nil, err
    }

    execution := s.createExecution(workflow, input)

    if err := s.engine.Execute(ctx, execution); err != nil {
        return nil, err
    }

    return execution, nil
}

func (s *WorkflowService) getWorkflow(ctx context.Context, id string) (*domain.Workflow, error) {
    // Implementation
}

func (s *WorkflowService) validateInput(workflow *domain.Workflow, input map[string]interface{}) error {
    // Implementation
}

// Bad - too long, does too many things
func (s *WorkflowService) Execute(ctx context.Context, workflowID string, input map[string]interface{}) (*domain.Execution, error) {
    // 200 lines of code doing everything
}
```

### Error Handling

#### Always Check Errors

```go
// Good
result, err := someFunction()
if err != nil {
    return nil, fmt.Errorf("failed to execute: %w", err)
}

// Bad
result, _ := someFunction()
```

#### Wrap Errors with Context

```go
// Good
if err := s.repo.Save(ctx, workflow); err != nil {
    return fmt.Errorf("failed to save workflow %s: %w", workflow.ID, err)
}

// Bad
if err := s.repo.Save(ctx, workflow); err != nil {
    return err
}
```

#### Use Custom Error Types

```go
// Good
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// Usage
if input == nil {
    return &ValidationError{
        Field:   "input",
        Message: "input cannot be nil",
    }
}
```

#### Don't Panic

```go
// Good
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Bad
func divide(a, b int) int {
    if b == 0 {
        panic("division by zero")
    }
    return a / b
}
```

### Comments

#### Package Comments

```go
// Package workflow provides workflow orchestration functionality.
//
// It includes workflow definition, execution, and management capabilities.
// Workflows are defined as directed acyclic graphs (DAGs) of nodes.
package workflow
```

#### Function Comments

```go
// Execute runs a workflow with the given input and returns the execution result.
//
// The function validates the workflow, creates an execution record, and
// processes all nodes in topological order. If any node fails, the entire
// execution is marked as failed.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - workflowID: Unique identifier of the workflow to execute
//   - input: Input data passed to the workflow
//
// Returns:
//   - *domain.Execution: The execution result with status and outputs
//   - error: Any error that occurred during execution
func (s *WorkflowService) Execute(ctx context.Context, workflowID string, input map[string]interface{}) (*domain.Execution, error) {
    // Implementation
}
```

#### Inline Comments

```go
// Good - explain why, not what
// Use exponential backoff to avoid overwhelming the API
delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second

// Bad - obvious comment
// Set delay to 2^attempt seconds
delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
```

## Directory Policy

### `internal/` vs `pkg/`

#### `internal/` - Private Application Code

Use `internal/` for code that should not be imported by external projects:

```
internal/
├── api/          # HTTP handlers and middleware
├── config/       # Configuration management
├── domain/       # Domain models and business logic
├── engine/       # Workflow execution engine
├── infrastructure/ # External service clients
├── repository/   # Data access layer
└── service/      # Business logic services
```

**Characteristics:**
- Cannot be imported by external projects
- Application-specific logic
- Implementation details

#### `pkg/` - Public Libraries

Use `pkg/` for code that can be imported by external projects:

```
pkg/
├── client/       # Go client SDK
├── constants/    # Public constants
├── crypto/       # Cryptography utilities
├── logger/       # Logging utilities
├── utils/        # General utilities
└── validator/    # Validation utilities
```

**Characteristics:**
- Can be imported by external projects
- Reusable libraries
- Well-documented public APIs
- Stable interfaces

### Directory Structure Best Practices

```go
// Good - clear separation of concerns
internal/
├── api/
│   ├── handler/
│   │   ├── workflow_handler.go
│   │   └── execution_handler.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── logging.go
│   └── router.go
├── service/
│   ├── workflow_service.go
│   └── execution_service.go
└── repository/
    ├── workflow_repository.go
    └── execution_repository.go

// Bad - mixed concerns
internal/
├── workflow.go
├── execution.go
├── handler.go
└── repository.go
```

## Linting Tools

### golangci-lint Configuration

GoFlow uses `golangci-lint` for code quality checks.

#### Configuration File

`.golangci.yml`:

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gosec
    - gocyclo
    - dupl
    - misspell
    - lll
    - unparam
    - nakedret
    - prealloc
    - exportloopref

linters-settings:
  gocyclo:
    min-complexity: 15
  lll:
    line-length: 120
  errcheck:
    check-type-assertions: true
    check-blank: true
  govet:
    check-shadowing: true
  misspell:
    locale: US

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - dupl
        - gosec
```

#### Running Linters

```bash
# Run all linters
make lint

# Run specific linter
golangci-lint run --enable=gofmt

# Auto-fix issues
golangci-lint run --fix

# Run on specific files
golangci-lint run internal/service/
```

### Pre-commit Hooks

Install pre-commit hooks to run linters automatically:

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "Running linters..."
make lint

if [ $? -ne 0 ]; then
    echo "Linting failed. Please fix the issues before committing."
    exit 1
fi

echo "Running tests..."
make test

if [ $? -ne 0 ]; then
    echo "Tests failed. Please fix the issues before committing."
    exit 1
fi

echo "All checks passed!"
```

## Code Organization

### Layered Architecture

```
┌─────────────────────────────────────┐
│         API Layer (HTTP)            │
│  - Handlers                         │
│  - Middleware                       │
│  - Request/Response DTOs            │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│        Service Layer                │
│  - Business Logic                   │
│  - Orchestration                    │
│  - Validation                       │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│        Domain Layer                 │
│  - Domain Models                    │
│  - Business Rules                   │
│  - Interfaces                       │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│      Repository Layer               │
│  - Data Access                      │
│  - Query Building                   │
│  - Transactions                     │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│    Infrastructure Layer             │
│  - Database                         │
│  - Cache                            │
│  - External APIs                    │
└─────────────────────────────────────┘
```

### Dependency Injection

```go
// Good - dependencies injected via constructor
type WorkflowService struct {
    repo   repository.WorkflowRepository
    engine *engine.WorkflowEngine
    logger *zap.Logger
}

func NewWorkflowService(
    repo repository.WorkflowRepository,
    engine *engine.WorkflowEngine,
    logger *zap.Logger,
) *WorkflowService {
    return &WorkflowService{
        repo:   repo,
        engine: engine,
        logger: logger,
    }
}

// Bad - global dependencies
var globalRepo repository.WorkflowRepository

type WorkflowService struct{}

func (s *WorkflowService) Execute(ctx context.Context, id string) error {
    workflow, err := globalRepo.FindByID(ctx, id)
    // ...
}
```

### Interface-Based Design

```go
// Good - depend on interfaces
type WorkflowService struct {
    repo repository.WorkflowRepository // interface
}

type WorkflowRepository interface {
    FindByID(ctx context.Context, id string) (*domain.Workflow, error)
    Save(ctx context.Context, workflow *domain.Workflow) error
}

// Bad - depend on concrete types
type WorkflowService struct {
    repo *PostgresWorkflowRepository // concrete type
}
```

## Best Practices

### Context Usage

```go
// Good - pass context as first parameter
func (s *WorkflowService) Execute(ctx context.Context, workflowID string) error {
    // Use context for cancellation and timeouts
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue execution
    }
}

// Bad - no context
func (s *WorkflowService) Execute(workflowID string) error {
    // No way to cancel or timeout
}
```

### Goroutine Management

```go
// Good - use sync.WaitGroup
func (s *WorkflowService) ExecuteParallel(ctx context.Context, nodes []*domain.Node) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(nodes))

    for _, node := range nodes {
        wg.Add(1)
        go func(n *domain.Node) {
            defer wg.Done()
            if err := s.executeNode(ctx, n); err != nil {
                errChan <- err
            }
        }(node)
    }

    wg.Wait()
    close(errChan)

    for err := range errChan {
        if err != nil {
            return err
        }
    }

    return nil
}

// Bad - no synchronization
func (s *WorkflowService) ExecuteParallel(ctx context.Context, nodes []*domain.Node) error {
    for _, node := range nodes {
        go s.executeNode(ctx, node) // Fire and forget
    }
    return nil // Returns immediately
}
```

### Resource Cleanup

```go
// Good - use defer for cleanup
func (s *WorkflowService) Execute(ctx context.Context, workflowID string) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback() // Always rollback, commit will override

    // Do work

    return tx.Commit()
}

// Bad - manual cleanup
func (s *WorkflowService) Execute(ctx context.Context, workflowID string) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    // Do work

    tx.Commit()
    return nil // Forgot to rollback on error
}
```

### Struct Initialization

```go
// Good - use named fields
workflow := &domain.Workflow{
    ID:        "wf_123",
    Name:      "My Workflow",
    Version:   1,
    CreatedAt: time.Now(),
}

// Bad - positional fields
workflow := &domain.Workflow{
    "wf_123",
    "My Workflow",
    1,
    time.Now(),
}
```

### Table-Driven Tests

```go
// Good - table-driven tests
func TestValidateWorkflow(t *testing.T) {
    tests := []struct {
        name    string
        workflow *domain.Workflow
        wantErr bool
    }{
        {
            name: "valid workflow",
            workflow: &domain.Workflow{
                ID:   "wf_123",
                Name: "Test",
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            workflow: &domain.Workflow{
                Name: "Test",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateWorkflow(tt.workflow)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateWorkflow() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

For more information, see:
- [Contributing Guide](./contributing.md)
- [Testing Guide](./testing.md)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
