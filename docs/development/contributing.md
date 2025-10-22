# Contributing to GoFlow

Thank you for your interest in contributing to GoFlow! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting Pull Requests](#submitting-pull-requests)
- [Development Workflow](#development-workflow)
- [Branch Naming Conventions](#branch-naming-conventions)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Code Review Process](#code-review-process)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](../../CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to goflow-conduct@example.com.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 14+
- Redis 6+
- Git

### Setup Development Environment

1. **Fork the repository**

```bash
# Fork on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/goflow-service.git
cd goflow-service
```

2. **Add upstream remote**

```bash
git remote add upstream https://github.com/goflow/goflow-service.git
```

3. **Install dependencies**

```bash
go mod download
```

4. **Start development services**

```bash
docker-compose up -d postgres redis
```

5. **Run migrations**

```bash
make migrate-up
```

6. **Run the application**

```bash
make run
```

7. **Verify setup**

```bash
curl http://localhost:8080/health
```

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

**Bug Report Template:**

```markdown
**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Create workflow with '...'
2. Execute with input '...'
3. See error

**Expected behavior**
A clear description of what you expected to happen.

**Actual behavior**
What actually happened.

**Environment:**
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.21.0]
- GoFlow version: [e.g., v1.0.0]

**Logs**
```
Paste relevant logs here
```

**Additional context**
Add any other context about the problem here.
```

### Suggesting Features

We welcome feature suggestions! Please create an issue with the following information:

**Feature Request Template:**

```markdown
**Is your feature request related to a problem?**
A clear description of the problem. Ex. I'm always frustrated when [...]

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
A clear description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.

**Proposed Implementation**
If you have ideas about how to implement this feature, please share them.
```

### Submitting Pull Requests

1. **Create a feature branch**

```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes**

- Follow the [Code Style Guide](./code-style.md)
- Write tests for new functionality
- Update documentation as needed

3. **Run tests**

```bash
make test
make lint
```

4. **Commit your changes**

```bash
git add .
git commit -m "feat: add new feature"
```

5. **Push to your fork**

```bash
git push origin feature/your-feature-name
```

6. **Create a Pull Request**

- Go to the original repository on GitHub
- Click "New Pull Request"
- Select your fork and branch
- Fill out the PR template
- Submit the PR

## Development Workflow

### 1. Sync with Upstream

Before starting work, sync your fork with upstream:

```bash
git checkout main
git fetch upstream
git merge upstream/main
git push origin main
```

### 2. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 3. Develop and Test

```bash
# Run tests continuously
make test-watch

# Run specific test
go test -v ./internal/service -run TestWorkflowService_Execute

# Check coverage
make test-coverage
```

### 4. Commit Changes

Follow [Conventional Commits](#commit-message-guidelines) specification.

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Branch Naming Conventions

Use the following prefixes for branch names:

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Adding or updating tests
- `chore/` - Maintenance tasks

**Examples:**

```
feature/add-webhook-node
fix/database-connection-leak
docs/update-api-documentation
refactor/simplify-executor-factory
test/add-integration-tests
chore/update-dependencies
```

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `chore` - Maintenance tasks
- `ci` - CI/CD changes
- `build` - Build system changes

### Scope

The scope should be the name of the affected component:

- `api` - API layer
- `service` - Service layer
- `engine` - Workflow engine
- `repository` - Data access layer
- `node` - Node types
- `config` - Configuration
- `docs` - Documentation

### Examples

```bash
# Feature
feat(node): add OpenAI completion node type

# Bug fix
fix(engine): resolve deadlock in parallel execution

# Documentation
docs(api): update authentication examples

# Refactoring
refactor(service): simplify workflow validation logic

# Performance
perf(repository): optimize workflow query with indexes

# Breaking change
feat(api)!: change workflow execution response format

BREAKING CHANGE: The execution response now includes additional metadata.
```

## Code Review Process

### Pull Request Template

When creating a PR, please include:

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Related Issues
Fixes #(issue number)

## Changes Made
- Change 1
- Change 2
- Change 3

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] My code follows the code style of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Screenshots (if applicable)
Add screenshots to help explain your changes
```

### Review Criteria

Reviewers will check for:

1. **Functionality**
   - Does the code work as intended?
   - Are edge cases handled?
   - Are there any bugs?

2. **Code Quality**
   - Follows Go best practices
   - Adheres to project code style
   - Proper error handling
   - No code duplication

3. **Testing**
   - Adequate test coverage
   - Tests are meaningful
   - Tests pass consistently

4. **Documentation**
   - Code is well-commented
   - Public APIs are documented
   - README/docs updated if needed

5. **Performance**
   - No obvious performance issues
   - Efficient algorithms used
   - Resource usage is reasonable

### Review Process

1. **Automated Checks**
   - CI/CD pipeline runs automatically
   - Tests must pass
   - Linting must pass
   - Coverage must meet threshold (80%)

2. **Peer Review**
   - At least one approval required
   - Address all review comments
   - Request re-review after changes

3. **Maintainer Review**
   - Final approval from maintainer
   - Merge when all checks pass

### Addressing Review Comments

```bash
# Make requested changes
git add .
git commit -m "fix: address review comments"
git push origin feature/your-feature-name

# If you need to amend the last commit
git add .
git commit --amend --no-edit
git push origin feature/your-feature-name --force
```

## Testing Requirements

### Minimum Requirements

- **Unit Test Coverage:** 80% minimum
- **Integration Tests:** For all API endpoints
- **End-to-End Tests:** For critical workflows

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test -v ./internal/service

# Run specific test
go test -v ./internal/service -run TestWorkflowService_Execute

# Run integration tests
make test-integration

# Run e2e tests
make test-e2e
```

### Writing Tests

See [Testing Guide](./testing.md) for detailed testing guidelines.

## Documentation

### Code Documentation

- Use GoDoc comments for all exported functions, types, and constants
- Include examples in documentation where helpful
- Keep comments up-to-date with code changes

**Example:**

```go
// WorkflowService handles workflow operations
type WorkflowService struct {
    repo   repository.WorkflowRepository
    engine *engine.WorkflowEngine
}

// Execute runs a workflow with the given input
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - workflowID: Unique identifier of the workflow
//   - input: Input data for the workflow
//
// Returns:
//   - *domain.Execution: The execution result
//   - error: Any error that occurred during execution
//
// Example:
//   execution, err := service.Execute(ctx, "wf_123", map[string]interface{}{
//       "user_id": "user_456",
//   })
func (s *WorkflowService) Execute(ctx context.Context, workflowID string, input map[string]interface{}) (*domain.Execution, error) {
    // Implementation
}
```

### Documentation Files

Update relevant documentation when making changes:

- **README.md** - Project overview and quick start
- **docs/architecture.md** - Architecture changes
- **docs/api/** - API changes
- **docs/guides/** - User guides
- **docs/development/** - Development guides

### API Documentation

- Update OpenAPI specification in `docs/api/openapi.yaml`
- Update Postman collection in `docs/api/postman/`
- Add examples for new endpoints

## Community

### Communication Channels

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - General questions and discussions
- **Slack** - Real-time chat (join at goflow.slack.com)
- **Email** - goflow-dev@example.com

### Getting Help

If you need help:

1. Check the [documentation](../README.md)
2. Search [existing issues](https://github.com/goflow/goflow-service/issues)
3. Ask in [GitHub Discussions](https://github.com/goflow/goflow-service/discussions)
4. Join our [Slack channel](https://goflow.slack.com)

### Recognition

Contributors are recognized in:

- **CONTRIBUTORS.md** - List of all contributors
- **Release Notes** - Acknowledgment in release notes
- **GitHub** - Contributor badge on profile

## License

By contributing to GoFlow, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to GoFlow! 🎉

For more information, see:
- [Code Style Guide](./code-style.md)
- [Testing Guide](./testing.md)
- [Architecture Documentation](../architecture.md)
