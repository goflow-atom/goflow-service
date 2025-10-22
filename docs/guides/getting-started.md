# Getting Started

This guide will help you set up and run GoFlow Workflow Engine locally for development.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Environment Setup](#environment-setup)
- [Running Locally](#running-locally)
- [Testing Workflow Execution](#testing-workflow-execution)
- [Development Tools](#development-tools)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **Operating System**: Linux, macOS, or Windows (with WSL2)
- **CPU**: 2+ cores recommended
- **RAM**: 4GB minimum, 8GB recommended
- **Disk**: 10GB free space

### Required Software

1. **Go** (version 1.21 or higher)
   ```bash
   # Check Go version
   go version

   # Install Go (if needed)
   # Visit: https://golang.org/dl/
   ```

2. **PostgreSQL** (version 14 or higher)
   ```bash
   # Check PostgreSQL version
   psql --version

   # Install PostgreSQL
   # Ubuntu/Debian
   sudo apt-get install postgresql-14

   # macOS
   brew install postgresql@14
   ```

3. **Redis** (version 6 or higher)
   ```bash
   # Check Redis version
   redis-cli --version

   # Install Redis
   # Ubuntu/Debian
   sudo apt-get install redis-server

   # macOS
   brew install redis
   ```

4. **Docker & Docker Compose** (optional, recommended)
   ```bash
   # Check Docker version
   docker --version
   docker-compose --version

   # Install Docker
   # Visit: https://docs.docker.com/get-docker/
   ```

### Optional Software

- **Make**: For using Makefile commands
- **Air**: For hot-reload development
- **golangci-lint**: For code linting

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/goflow-service.git
cd goflow-service
```

### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download

# Verify dependencies
go mod verify
```

### 3. Install Development Tools

```bash
# Install Air for hot-reload
go install github.com/cosmtrek/air@latest

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Environment Setup

### 1. Create Environment File

```bash
# Copy example environment file
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` file with your configuration:

```bash
# Application
APP_ENV=development
APP_PORT=8080
APP_LOG_LEVEL=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=goflow
DB_PASSWORD=goflow_password
DB_NAME=goflow_db
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka (optional)
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=goflow-events

# Inngest
INNGEST_EVENT_KEY=your_inngest_event_key
INNGEST_SIGNING_KEY=your_inngest_signing_key

# OpenAI (optional)
OPENAI_API_KEY=your_openai_api_key
OPENAI_ORG_ID=your_openai_org_id

# JWT
JWT_SECRET=your_jwt_secret_key_change_this_in_production
JWT_EXPIRATION=3600

# SMTP (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=noreply@goflow.example.com
```

### 3. Start Dependencies with Docker Compose

```bash
# Start PostgreSQL, Redis, and Kafka
docker-compose up -d

# Check services are running
docker-compose ps

# View logs
docker-compose logs -f
```

**Alternative: Manual Setup**

If not using Docker, start services manually:

```bash
# Start PostgreSQL
sudo systemctl start postgresql

# Start Redis
sudo systemctl start redis

# Create database
createdb goflow_db
```

### 4. Run Database Migrations

```bash
# Run migrations
make migrate-up

# Or manually
migrate -path migrations -database "postgresql://goflow:goflow_password@localhost:5432/goflow_db?sslmode=disable" up

# Verify migrations
make migrate-version
```

## Running Locally

### Option 1: Using Make (Recommended)

```bash
# Build and run
make run

# Run with hot-reload (development)
make dev

# Run tests
make test

# Run linter
make lint
```

### Option 2: Using Go Commands

```bash
# Build the application
go build -o bin/goflow cmd/server/main.go

# Run the application
./bin/goflow

# Or build and run in one command
go run cmd/server/main.go
```

### Option 3: Using Air (Hot Reload)

```bash
# Start with hot-reload
air

# Air will watch for file changes and automatically rebuild
```

### Verify Service is Running

```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","timestamp":"2024-01-01T10:00:00Z"}

# Check API version
curl http://localhost:8080/api/v1/version

# Expected response:
# {"version":"1.0.0","build":"dev","go_version":"go1.21.0"}
```

## Testing Workflow Execution

### 1. Create a Simple Workflow

Create a file `examples/test-workflow.json`:

```json
{
  "name": "Hello World Workflow",
  "description": "A simple test workflow",
  "nodes": [
    {
      "id": "start",
      "type": "transform",
      "config": {
        "script": "return { message: 'Hello, ' + input.name + '!' }"
      }
    },
    {
      "id": "log",
      "type": "transform",
      "config": {
        "script": "console.log(input.message); return input"
      }
    }
  ],
  "edges": [
    {
      "from": "start",
      "to": "log"
    }
  ]
}
```

### 2. Create Workflow via API

```bash
# Create workflow
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d @examples/test-workflow.json

# Save the workflow ID from response
WORKFLOW_ID="wf_123abc"
```

### 3. Execute Workflow

```bash
# Trigger execution
curl -X POST "http://localhost:8080/api/v1/workflows/${WORKFLOW_ID}/execute" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "input": {
      "name": "World"
    },
    "async": false
  }'

# Expected response:
# {
#   "execution_id": "exec_456def",
#   "status": "completed",
#   "output": {
#     "message": "Hello, World!"
#   }
# }
```

### 4. View Execution Details

```bash
# Get execution details
curl -X GET "http://localhost:8080/api/v1/executions/exec_456def" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get execution logs
curl -X GET "http://localhost:8080/api/v1/executions/exec_456def/logs" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Development Tools

### Makefile Commands

The project includes a Makefile with common development tasks:

```bash
# Build
make build          # Build the application
make build-linux    # Build for Linux
make build-docker   # Build Docker image

# Run
make run            # Run the application
make dev            # Run with hot-reload

# Test
make test           # Run all tests
make test-unit      # Run unit tests
make test-integration  # Run integration tests
make test-e2e       # Run end-to-end tests
make test-coverage  # Run tests with coverage

# Database
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations
make migrate-create # Create new migration
make db-seed        # Seed database with test data

# Code Quality
make lint           # Run linter
make fmt            # Format code
make vet            # Run go vet
make check          # Run all checks

# Clean
make clean          # Clean build artifacts
make clean-all      # Clean everything including dependencies
```

### Hot Reload with Air

Air configuration (`.air.toml`):

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

### VS Code Configuration

Recommended VS Code settings (`.vscode/settings.json`):

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "go.coverOnSave": true,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

### Debugging

#### VS Code Launch Configuration

`.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/server",
      "env": {
        "APP_ENV": "development"
      },
      "args": []
    },
    {
      "name": "Attach to Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": "${command:pickProcess}"
    }
  ]
}
```

#### Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start with debugger
dlv debug cmd/server/main.go

# Common commands
(dlv) break main.main
(dlv) continue
(dlv) next
(dlv) step
(dlv) print variable
(dlv) quit
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Error

**Error:**
```
failed to connect to database: connection refused
```

**Solution:**
```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Start PostgreSQL
sudo systemctl start postgresql

# Check connection
psql -h localhost -U goflow -d goflow_db
```

#### 2. Redis Connection Error

**Error:**
```
failed to connect to redis: connection refused
```

**Solution:**
```bash
# Check Redis is running
redis-cli ping

# Start Redis
sudo systemctl start redis

# Check Redis logs
sudo journalctl -u redis -f
```

#### 3. Port Already in Use

**Error:**
```
bind: address already in use
```

**Solution:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
APP_PORT=8081 make run
```

#### 4. Migration Errors

**Error:**
```
migration failed: duplicate key value
```

**Solution:**
```bash
# Check migration status
make migrate-version

# Rollback last migration
make migrate-down

# Re-run migrations
make migrate-up

# Force migration version (use with caution)
migrate -path migrations -database "..." force <version>
```

#### 5. Module Download Errors

**Error:**
```
go: module not found
```

**Solution:**
```bash
# Clean module cache
go clean -modcache

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Logs and Debugging

#### View Application Logs

```bash
# Tail logs
tail -f logs/app.log

# Search logs
grep "ERROR" logs/app.log

# View structured logs
cat logs/app.log | jq '.'
```

#### Enable Debug Logging

```bash
# Set log level to debug
export APP_LOG_LEVEL=debug

# Run with debug logging
make run
```

#### Database Query Logging

Enable SQL query logging in `.env`:

```bash
DB_LOG_QUERIES=true
DB_LOG_SLOW_QUERIES=true
DB_SLOW_QUERY_THRESHOLD=100ms
```

### Getting Help

If you encounter issues not covered here:

1. **Check Documentation**: Review the [Troubleshooting Guide](./troubleshooting.md)
2. **Search Issues**: Look for similar issues on [GitHub Issues](https://github.com/yourusername/goflow-service/issues)
3. **Ask Questions**: Post on [GitHub Discussions](https://github.com/yourusername/goflow-service/discussions)
4. **Report Bugs**: Create a new issue with detailed information

## Next Steps

Now that you have GoFlow running locally, explore these topics:

1. **[Workflow Definition](./workflow-definition.md)**: Learn how to define workflows
2. **[Node Types](./node-types.md)**: Explore available node types
3. **[API Documentation](../api/api.md)**: Learn about the REST API
4. **[Deployment](./deployment.md)**: Deploy to production
5. **[Contributing](../development/contributing.md)**: Contribute to the project

---

**Happy Building! 🚀**
