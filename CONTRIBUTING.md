# Contributing to GoFlow Workflow Engine

First off, thank you for considering contributing to GoFlow! It's people like you that make GoFlow such a great tool.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [How Can I Contribute?](#how-can-i-contribute)
- [Style Guidelines](#style-guidelines)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Community](#community)

## Code of Conduct

This project and everyone participating in it is governed by the [GoFlow Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [conduct@goflow-atom.dev](mailto:conduct@goflow-atom.dev).

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 6+
- Docker & Docker Compose
- Make
- Air (for hot-reload development)
- golangci-lint

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/goflow-service.git
   cd goflow-service
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/goflow-atom/goflow-service.git
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   go mod verify
   ```

5. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your local configuration
   ```

6. **Start dependencies**:
   ```bash
   docker-compose up -d postgres redis kafka
   ```

7. **Run database migrations**:
   ```bash
   make migrate-up
   ```

8. **Verify setup**:
   ```bash
   make test
   make lint
   ```

## Development Workflow

### Creating a Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or modifications
- `chore/` - Maintenance tasks

### Running the Application

**Development mode with hot-reload**:
```bash
make dev
# or
air
```

**Standard mode**:
```bash
make run
```

### Running Tests

**All tests**:
```bash
make test
```

**Unit tests only**:
```bash
make test-unit
```

**Integration tests**:
```bash
make test-integration
```

**With coverage**:
```bash
make test-coverage
```

**Race detection**:
```bash
make test-race
```

### Linting

**Run linter**:
```bash
make lint
```

**Auto-fix issues**:
```bash
make lint-fix
```

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs **actual behavior**
- **Environment details** (OS, Go version, etc.)
- **Logs or error messages**
- **Screenshots** if applicable

Use the bug report template when creating an issue.

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear title and description**
- **Use case** - Why is this enhancement needed?
- **Proposed solution** - How should it work?
- **Alternatives considered**
- **Additional context** - Mockups, examples, etc.

### Your First Code Contribution

Unsure where to begin? Look for issues labeled:
- `good first issue` - Simple issues for newcomers
- `help wanted` - Issues where we need community help
- `beginner friendly` - Issues suitable for first-time contributors

### Pull Requests

1. **Update your fork**:
   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature
   ```

3. **Make your changes** following our style guidelines

4. **Add tests** for your changes

5. **Run tests and linting**:
   ```bash
   make test
   make lint
   ```

6. **Commit your changes** following commit guidelines

7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature
   ```

8. **Open a Pull Request** on GitHub

## Style Guidelines

### Go Code Style

We follow standard Go conventions and use `golangci-lint` for enforcement.

**Key principles**:
- Use `gofmt` for formatting (enforced by CI)
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable names (no single-letter names except in loops)
- Keep functions small and focused (max 50-100 lines)
- Use early returns to reduce nesting
- Add GoDoc comments for all exported functions and types

**Example**:
```go
// WorkflowService handles workflow-related business logic.
type WorkflowService struct {
    repo   repository.WorkflowRepository
    logger *zap.Logger
}

// GetByID retrieves a workflow by its unique identifier.
// Returns NotFoundError if the workflow does not exist.
func (s *WorkflowService) GetByID(ctx context.Context, id string) (*domain.Workflow, error) {
    if id == "" {
        return nil, fmt.Errorf("workflow ID cannot be empty")
    }

    workflow, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get workflow: %w", err)
    }

    return workflow, nil
}
```

### Error Handling

**Always wrap errors with context**:
```go
if err != nil {
    return fmt.Errorf("failed to create workflow: %w", err)
}
```

**Use custom error types for domain errors**:
```go
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}
```

### Testing Guidelines

**Test file naming**: `*_test.go`

**Test function naming**: `Test<StructName>_<MethodName>_<Scenario>`

**Example**:
```go
func TestWorkflowService_GetByID_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockWorkflowRepository)
    service := NewWorkflowService(mockRepo, logger)

    expectedWorkflow := &domain.Workflow{ID: "wf_123", Name: "Test"}
    mockRepo.On("FindByID", mock.Anything, "wf_123").Return(expectedWorkflow, nil)

    // Act
    result, err := service.GetByID(context.Background(), "wf_123")

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedWorkflow, result)
    mockRepo.AssertExpectations(t)
}
```

**Use table-driven tests** for multiple scenarios:
```go
func TestValidateWorkflow(t *testing.T) {
    tests := []struct {
        name    string
        input   *Workflow
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid workflow",
            input:   &Workflow{Name: "Test", Nodes: []Node{{ID: "1"}}},
            wantErr: false,
        },
        {
            name:    "empty name",
            input:   &Workflow{Name: "", Nodes: []Node{{ID: "1"}}},
            wantErr: true,
            errMsg:  "name cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateWorkflow(tt.input)
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

**Target coverage**: Minimum 80% code coverage

### Documentation

**Package-level documentation**:
```go
// Package workflow provides workflow management functionality.
//
// This package handles workflow creation, validation, and execution
// orchestration. It implements the core business logic for the
// GoFlow workflow engine.
package workflow
```

**Function documentation**:
```go
// CreateWorkflow creates a new workflow with the given definition.
//
// The workflow definition is validated before creation. If validation
// fails, an error is returned. The created workflow is persisted to
// the database and cached for quick retrieval.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - def: Workflow definition containing nodes and edges
//
// Returns:
//   - *Workflow: The created workflow with generated ID
//   - error: Validation or persistence error
//
// Example:
//
//     workflow, err := service.CreateWorkflow(ctx, definition)
//     if err != nil {
//         log.Fatal(err)
//     }
//     fmt.Printf("Created workflow: %s\n", workflow.ID)
func CreateWorkflow(ctx context.Context, def *WorkflowDefinition) (*Workflow, error) {
    // Implementation
}
```

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes
- `build`: Build system changes

### Scope

The scope should be the name of the affected component:
- `api` - API layer changes
- `service` - Service layer changes
- `engine` - Engine layer changes
- `domain` - Domain model changes
- `repo` - Repository layer changes
- `infra` - Infrastructure changes
- `config` - Configuration changes

### Examples

```
feat(engine): add support for parallel node execution

Implement parallel execution capability for workflow nodes that
have no dependencies on each other. This improves execution time
for workflows with independent branches.

Closes #123
```

```
fix(api): handle nil pointer in workflow handler

Add nil check for workflow object before accessing properties
to prevent panic when workflow is not found.

Fixes #456
```

```
docs(readme): update installation instructions

Add missing step for database migration and clarify
environment variable configuration.
```

### Breaking Changes

If your commit introduces breaking changes, add `BREAKING CHANGE:` in the footer:

```
feat(api): change workflow execution response format

BREAKING CHANGE: The execution response now returns a structured
object instead of a plain string. Update clients to handle the
new response format.
```

## Pull Request Process

### Before Submitting

1. ✅ Code compiles without errors
2. ✅ All tests pass
3. ✅ Linting passes
4. ✅ Code coverage is maintained or improved
5. ✅ Documentation is updated
6. ✅ Commit messages follow guidelines
7. ✅ Branch is up to date with main

### PR Title

Use the same format as commit messages:
```
feat(engine): add parallel node execution support
```

### PR Description

Use the pull request template and include:

- **Summary** - What does this PR do?
- **Motivation** - Why is this change needed?
- **Changes** - List of changes made
- **Testing** - How was this tested?
- **Screenshots** - If applicable
- **Related Issues** - Link to related issues
- **Checklist** - Complete the PR checklist

### Review Process

1. **Automated checks** must pass (CI/CD pipeline)
2. **Code review** by at least one maintainer
3. **Address feedback** - Make requested changes
4. **Approval** - PR is approved by maintainer
5. **Merge** - Maintainer merges the PR

### After Merge

- Delete your feature branch
- Update your local main branch
- Close related issues if applicable

## Community

### Communication Channels

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - General questions and discussions
- **Discord** - Real-time chat (coming soon)
- **Email** - [dev@goflow-atom.dev](mailto:dev@goflow-atom.dev)

### Getting Help

- Check the [documentation](docs/)
- Search existing [issues](https://github.com/goflow-atom/goflow-service/issues)
- Ask in [GitHub Discussions](https://github.com/goflow-atom/goflow-service/discussions)
- Join our Discord community (coming soon)

### Recognition

Contributors are recognized in:
- [CONTRIBUTORS.md](CONTRIBUTORS.md) file
- Release notes
- Project README

## Additional Resources

- [Architecture Documentation](docs/architecture.md)
- [API Documentation](docs/api/)
- [Development Guide](docs/development/)
- [Testing Guide](docs/development/testing.md)
- [Code Style Guide](docs/development/code-style.md)

---

Thank you for contributing to GoFlow! 🚀

