# Testing Guide

This document provides comprehensive guidelines for testing in GoFlow.

## Table of Contents

- [Testing Guide](#testing-guide)
  - [Table of Contents](#table-of-contents)
  - [Testing Strategy](#testing-strategy)
    - [Test Pyramid](#test-pyramid)
  - [Test Types](#test-types)
    - [Unit Tests](#unit-tests)
      - [Example](#example)
      - [Running Unit Tests](#running-unit-tests)
    - [Integration Tests](#integration-tests)
      - [Example](#example-1)
      - [Running Integration Tests](#running-integration-tests)
    - [End-to-End Tests](#end-to-end-tests)
  - [Mocking and Dependency Injection](#mocking-and-dependency-injection)
    - [Using testify/mock](#using-testifymock)
  - [CI Integration](#ci-integration)
    - [GitHub Actions](#github-actions)
  - [Test Isolation](#test-isolation)
    - [Database Isolation](#database-isolation)
    - [Resource Cleanup](#resource-cleanup)
  - [Coverage Requirements](#coverage-requirements)
    - [Check Coverage](#check-coverage)
  - [Best Practices](#best-practices)
    - [1. Test Naming](#1-test-naming)
    - [2. Arrange-Act-Assert Pattern](#2-arrange-act-assert-pattern)
    - [3. Use Table-Driven Tests](#3-use-table-driven-tests)
    - [4. Test Error Cases](#4-test-error-cases)
    - [5. Use Subtests](#5-use-subtests)
    - [6. Avoid Test Interdependence](#6-avoid-test-interdependence)

## Testing Strategy

GoFlow follows a comprehensive testing strategy with three levels of tests:

```
┌─────────────────────────────────────┐
│      End-to-End Tests (E2E)        │
│  - Full system integration         │
│  - Real dependencies               │
│  - Critical user flows             │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│     Integration Tests               │
│  - Multiple components              │
│  - Real database/cache              │
│  - API endpoints                    │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│        Unit Tests                   │
│  - Individual functions             │
│  - Mocked dependencies              │
│  - Fast execution                   │
└─────────────────────────────────────┘
```

### Test Pyramid

- **70% Unit Tests** - Fast, isolated, test individual components
- **20% Integration Tests** - Test component interactions
- **10% E2E Tests** - Test critical user journeys

## Test Types

### Unit Tests

Unit tests verify individual functions and methods in isolation.

#### Example

```go
// internal/service/workflow_service_test.go
package service

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "goflow-service/internal/domain"
)

// Mock repository
type MockWorkflowRepository struct {
    mock.Mock
}

func (m *MockWorkflowRepository) FindByID(ctx context.Context, id string) (*domain.Workflow, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) Save(ctx context.Context, workflow *domain.Workflow) error {
    args := m.Called(ctx, workflow)
    return args.Error(0)
}

// Test function
func TestWorkflowService_GetByID(t *testing.T) {
    // Arrange
    mockRepo := new(MockWorkflowRepository)
    service := NewWorkflowService(mockRepo, nil, nil)

    expectedWorkflow := &domain.Workflow{
        ID:   "wf_123",
        Name: "Test Workflow",
    }

    mockRepo.On("FindByID", mock.Anything, "wf_123").Return(expectedWorkflow, nil)

    // Act
    workflow, err := service.GetByID(context.Background(), "wf_123")

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedWorkflow.ID, workflow.ID)
    assert.Equal(t, expectedWorkflow.Name, workflow.Name)
    mockRepo.AssertExpectations(t)
}

// Table-driven test
func TestWorkflowService_Validate(t *testing.T) {
    tests := []struct {
        name     string
        workflow *domain.Workflow
        wantErr  bool
        errMsg   string
    }{
        {
            name: "valid workflow",
            workflow: &domain.Workflow{
                ID:      "wf_123",
                Name:    "Test",
                Version: 1,
                Nodes:   []*domain.Node{{ID: "node_1", Type: "webhook"}},
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            workflow: &domain.Workflow{
                Name: "Test",
            },
            wantErr: true,
            errMsg:  "workflow ID is required",
        },
        {
            name: "empty nodes",
            workflow: &domain.Workflow{
                ID:    "wf_123",
                Name:  "Test",
                Nodes: []*domain.Node{},
            },
            wantErr: true,
            errMsg:  "workflow must have at least one node",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewWorkflowService(nil, nil, nil)
            err := service.Validate(tt.workflow)

            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

#### Running Unit Tests

```bash
# Run all unit tests
go test ./...

# Run tests in specific package
go test ./internal/service

# Run specific test
go test ./internal/service -run TestWorkflowService_GetByID

# Run with verbose output
go test -v ./internal/service

# Run with coverage
go test -cover ./internal/service

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests

Integration tests verify interactions between multiple components with real dependencies.

#### Example

```go
// test/integration/workflow_test.go
// +build integration

package integration

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWorkflowIntegration(t *testing.T) {
    // Setup
    cfg := config.LoadTestConfig()
    db := setupTestDatabase(t, cfg)
    defer cleanupTestDatabase(t, db)

    repo := repository.NewWorkflowRepository(db)
    service := service.NewWorkflowService(repo, nil, nil)

    ctx := context.Background()

    // Test: Create workflow
    workflow := &domain.Workflow{
        ID:      "wf_test_123",
        Name:    "Integration Test Workflow",
        Version: 1,
    }

    err := service.Create(ctx, workflow)
    require.NoError(t, err)

    // Test: Retrieve workflow
    retrieved, err := service.GetByID(ctx, workflow.ID)
    require.NoError(t, err)
    assert.Equal(t, workflow.ID, retrieved.ID)
}
```

#### Running Integration Tests

```bash
# Run integration tests
go test -tags=integration ./test/integration

# Run with verbose output
go test -v -tags=integration ./test/integration
```

### End-to-End Tests

E2E tests verify complete user workflows through the API.

```bash
# Start the application
make run

# In another terminal, run E2E tests
go test -tags=e2e ./test/e2e
```

## Mocking and Dependency Injection

### Using testify/mock

```go
import "github.com/stretchr/testify/mock"

// Define mock
type MockNodeExecutor struct {
    mock.Mock
}

func (m *MockNodeExecutor) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    args := m.Called(ctx, input)
    return args.Get(0).(map[string]interface{}), args.Error(1)
}

// Use in test
func TestWithMock(t *testing.T) {
    mockExecutor := new(MockNodeExecutor)
    mockExecutor.On("Execute", mock.Anything, mock.Anything).Return(
        map[string]interface{}{"result": "success"},
        nil,
    )

    // Test code using mockExecutor

    mockExecutor.AssertExpectations(t)
}
```

## CI Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: make test

      - name: Run integration tests
        run: make test-integration
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASSWORD: postgres

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Test Isolation

### Database Isolation

```go
func setupTestDatabase(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", testDSN)
    require.NoError(t, err)

    // Create unique schema for this test
    schemaName := fmt.Sprintf("test_%d", time.Now().UnixNano())
    _, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
    require.NoError(t, err)

    t.Cleanup(func() {
        db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", schemaName))
        db.Close()
    })

    return db
}
```

### Resource Cleanup

```go
func TestWithCleanup(t *testing.T) {
    // Setup
    resource := createResource()

    // Register cleanup
    t.Cleanup(func() {
        resource.Close()
    })

    // Test code
}
```

## Coverage Requirements

- **Minimum Coverage:** 80%
- **Critical Paths:** 100%
- **New Code:** Must not decrease overall coverage

### Check Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
make test-coverage-html

# Check coverage threshold
go test -cover ./... | grep -E "coverage: [0-9]+\.[0-9]+%" | awk '{if ($2 < 80.0) exit 1}'
```

## Best Practices

### 1. Test Naming

```go
// Good
func TestWorkflowService_Execute_WithValidInput_ReturnsSuccess(t *testing.T)
func TestWorkflowService_Execute_WithInvalidInput_ReturnsError(t *testing.T)

// Bad
func TestExecute(t *testing.T)
func Test1(t *testing.T)
```

### 2. Arrange-Act-Assert Pattern

```go
func TestExample(t *testing.T) {
    // Arrange - Setup test data and dependencies
    input := "test"
    expected := "TEST"

    // Act - Execute the function under test
    result := ToUpper(input)

    // Assert - Verify the result
    assert.Equal(t, expected, result)
}
```

### 3. Use Table-Driven Tests

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "test", false},
        {"empty", "", true},
        {"too long", strings.Repeat("a", 1000), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 4. Test Error Cases

```go
func TestDivide(t *testing.T) {
    // Test success case
    result, err := Divide(10, 2)
    require.NoError(t, err)
    assert.Equal(t, 5, result)

    // Test error case
    _, err = Divide(10, 0)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "division by zero")
}
```

### 5. Use Subtests

```go
func TestWorkflow(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        // Test create
    })

    t.Run("Update", func(t *testing.T) {
        // Test update
    })

    t.Run("Delete", func(t *testing.T) {
        // Test delete
    })
}
```

### 6. Avoid Test Interdependence

```go
// Good - independent tests
func TestCreate(t *testing.T) {
    workflow := createTestWorkflow()
    // Test create
}

func TestUpdate(t *testing.T) {
    workflow := createTestWorkflow()
    // Test update
}

// Bad - tests depend on each other
var globalWorkflow *Workflow

func TestCreate(t *testing.T) {
    globalWorkflow = createTestWorkflow()
}

func TestUpdate(t *testing.T) {
    // Depends on TestCreate running first
    updateWorkflow(globalWorkflow)
}
```

---

For more information, see:
- [Contributing Guide](./contributing.md)
- [Code Style Guide](./code-style.md)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
