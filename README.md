# GoFlow Workflow Engine

<div align="center">

![GoFlow Logo](docs/assets/logo.png)

**A powerful, scalable workflow orchestration engine built in Go**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/goflow-atom/goflow-service/CI)](https://github.com/goflow-atom/goflow-service/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/goflow-atom/goflow-service)](https://goreportcard.com/report/github.com/goflow-atom/goflow-service)
[![Coverage](https://img.shields.io/codecov/c/github/goflow-atom/goflow-service)](https://codecov.io/gh/goflow-atom/goflow-service)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](docs/)

[Features](#-features) •
[Quick Start](#-quick-start) •
[Documentation](#-documentation) •
[Architecture](#-architecture) •
[Contributing](#-contributing) •
[License](#-license)

</div>

---

## 📋 Overview

GoFlow is a production-grade workflow orchestration engine that enables you to design, execute, and monitor complex workflows with ease. Built with Go's performance and concurrency in mind, GoFlow provides a flexible node-based architecture supporting various operations including HTTP requests, webhooks, conditional logic, loops, parallel execution, and AI-powered transformations.

### Why GoFlow?

- **🚀 High Performance**: Built in Go for speed and efficiency
- **🔄 DAG-Based Execution**: Workflows as Directed Acyclic Graphs with intelligent dependency resolution
- **⚡ Parallel Processing**: Execute independent nodes concurrently for optimal performance
- **🔌 Extensible**: 11+ built-in node types with easy custom node creation
- **💾 Durable Execution**: State persistence and automatic recovery from failures
- **🤖 AI-Powered**: Native OpenAI integration for intelligent workflows
- **📊 Observable**: Comprehensive logging, metrics, and tracing
- **🔒 Secure**: Built-in authentication, authorization, and encryption
- **☁️ Cloud-Native**: Kubernetes-ready with Docker support
- **📈 Scalable**: Horizontal scaling with stateless architecture

## ✨ Features

### Core Capabilities

- **Workflow Management**: Create, update, version, and manage workflows via REST API
- **DAG Execution Engine**: Intelligent topological sorting and dependency resolution
- **Node Types**: 11+ built-in node types for various operations
- **Parallel Execution**: Concurrent execution of independent workflow branches
- **Conditional Logic**: Dynamic branching based on runtime conditions
- **Loop Support**: Iterate over collections with configurable loop nodes
- **Expression Evaluation**: JSONPath and template-based dynamic data transformation
- **Error Handling**: Configurable retry policies with exponential backoff
- **State Management**: Persistent execution state with resume capability
- **Webhook Support**: Trigger workflows via HTTP webhooks with HMAC validation
- **Scheduling**: Cron-based workflow scheduling
- **Event Streaming**: Kafka integration for event-driven workflows
- **Caching**: Redis-based caching for improved performance
- **Monitoring**: Prometheus metrics and Grafana dashboards

### Node Types

| Node Type | Description | Use Case |
|-----------|-------------|----------|
| **Webhook** | Receive HTTP requests | API endpoints, external triggers |
| **HTTP Request** | Make HTTP calls | API integration, data fetching |
| **Conditional** | Branch based on conditions | Decision logic, routing |
| **Loop** | Iterate over collections | Batch processing, data transformation |
| **Parallel** | Execute nodes concurrently | Performance optimization |
| **Transform** | Transform data | Data mapping, filtering |
| **Delay** | Wait for duration | Rate limiting, scheduling |
| **Database** | Execute SQL queries | Data persistence, retrieval |
| **Email** | Send emails | Notifications, alerts |
| **OpenAI Completion** | Generate text with GPT | Content generation, analysis |
| **OpenAI Embedding** | Create vector embeddings | Semantic search, similarity |

### Advanced Features

- **Durable Execution**: Inngest integration for long-running workflows
- **Circuit Breaker**: Prevent cascading failures
- **Distributed Locking**: Redis-based coordination for multi-instance deployments
- **Rate Limiting**: Protect against abuse and manage costs
- **Audit Logging**: Complete audit trail for compliance
- **RBAC**: Role-based access control for workflows and executions
- **Secret Management**: Secure storage and retrieval of sensitive data
- **Multi-tenancy**: Isolated workspaces for different teams

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 6+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/goflow-atom/goflow-service.git
   cd goflow-service
   ```

2. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start dependencies** (using Docker Compose):
   ```bash
   docker-compose up -d postgres redis kafka
   ```

4. **Run database migrations**:
   ```bash
   make migrate-up
   ```

5. **Start the service**:
   ```bash
   # Development mode with hot-reload
   make dev

   # Or standard mode
   make run
   ```

6. **Verify the service**:
   ```bash
   curl http://localhost:8080/health
   ```

### Your First Workflow

Create a simple workflow that fetches data from an API:

```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fetch User Data",
    "description": "Fetch user data from API",
    "nodes": [
      {
        "id": "fetch-user",
        "type": "http_request",
        "config": {
          "method": "GET",
          "url": "https://jsonplaceholder.typicode.com/users/1"
        }
      }
    ],
    "edges": []
  }'
```

Execute the workflow:

```bash
curl -X POST http://localhost:8080/api/v1/workflows/{workflow_id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": {}
  }'
```

Check execution status:

```bash
curl http://localhost:8080/api/v1/executions/{execution_id}
```

## 📚 Documentation

Comprehensive documentation is available in the [docs/](docs/) directory:

- **[Getting Started Guide](docs/guides/getting-started.md)** - Detailed setup and configuration
- **[Architecture Overview](docs/architecture.md)** - System design and components
- **[Workflow Definition](docs/guides/workflow-definition.md)** - How to define workflows
- **[Node Types Reference](docs/guides/node-types.md)** - Complete node type documentation
- **[API Reference](docs/api/)** - REST API documentation
- **[Deployment Guide](docs/guides/deployment.md)** - Production deployment
- **[Development Guide](docs/development/)** - Contributing and development workflow
- **[Examples](examples/)** - Sample workflows and use cases

## 🏗️ Architecture

GoFlow follows a layered architecture pattern with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                           │
│  (REST API, gRPC, Client SDKs)                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      API Layer                              │
│  (Gin Router, Handlers, Middleware, Validation)             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                            │
│  (Workflow Service, Execution Service, Scheduler)           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Engine Layer                             │
│  (Workflow Engine, DAG Resolver, Executor, State Manager)   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                       │
│  (PostgreSQL, Redis, Kafka, OpenAI, HTTP Client)            │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

- **API Layer**: HTTP routing, request/response handling, authentication
- **Service Layer**: Business logic orchestration, transaction management
- **Engine Layer**: Workflow execution, DAG resolution, state management
- **Domain Layer**: Business entities, validation rules
- **Repository Layer**: Data access and persistence
- **Infrastructure Layer**: External service integrations

See [Architecture Documentation](docs/architecture.md) for detailed information.

## 🛠️ Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin (HTTP routing and middleware)
- **Workflow Engine**: Inngest (durable execution)
- **Database**: PostgreSQL 14+ (primary data store)
- **Cache**: Redis 6+ (caching and distributed locking)
- **Message Queue**: Kafka (event streaming)
- **AI Integration**: OpenAI API (GPT completions and embeddings)
- **Logging**: Zap (structured logging)
- **Configuration**: Viper (configuration management)
- **Testing**: Testify (assertions and mocking)
- **Linting**: golangci-lint
- **Containerization**: Docker & Docker Compose
- **Orchestration**: Kubernetes
- **Infrastructure**: Terraform
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus & Grafana

## 📁 Project Structure

```
goflow-service/
├── cmd/
│   └── server/              # Application entry point
├── internal/
│   ├── api/                 # API layer (handlers, middleware, DTOs)
│   ├── service/             # Service layer (business logic)
│   ├── engine/              # Workflow engine (DAG, execution, expressions)
│   ├── domain/              # Domain models and validation
│   ├── repository/          # Data access layer
│   └── infrastructure/      # External integrations (Redis, Kafka, etc.)
├── pkg/                     # Public libraries
│   ├── client/              # Go client SDK
│   ├── logger/              # Logging utilities
│   ├── validator/           # Validation utilities
│   └── utils/               # Common utilities
├── test/
│   ├── unit/                # Unit tests
│   ├── integration/         # Integration tests
│   └── e2e/                 # End-to-end tests
├── migrations/              # Database migrations
├── configs/                 # Configuration files
├── deployments/             # Deployment configurations
│   ├── docker/              # Dockerfiles
│   ├── kubernetes/          # Kubernetes manifests
│   └── terraform/           # Terraform configurations
├── docs/                    # Documentation
├── examples/                # Example workflows and clients
├── tools/                   # Development tools
└── api/                     # API specifications (OpenAPI, Proto)
```

## 🔧 Development

### Prerequisites

Install development tools:

```bash
# Install Air for hot-reload
go install github.com/cosmtrek/air@latest

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Common Commands

```bash
# Development with hot-reload
make dev

# Build
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt

# Run database migrations
make migrate-up

# Rollback migrations
make migrate-down

# Generate mocks
make mocks

# Build Docker image
make docker-build

# Run in Docker
make docker-run
```

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# E2E tests
make test-e2e

# With race detection
make test-race

# With coverage report
make test-coverage
open coverage.html
```

### Code Quality

```bash
# Run linter
make lint

# Auto-fix linting issues
make lint-fix

# Check code formatting
make fmt-check

# Format code
make fmt

# Security scan
make security-scan
```

## 🚢 Deployment

### Docker

```bash
# Build image
docker build -t goflow-service:latest .

# Run container
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e REDIS_HOST=redis \
  goflow-service:latest
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Kubernetes

```bash
# Apply configurations
kubectl apply -f deployments/kubernetes/

# Check status
kubectl get pods -n goflow

# View logs
kubectl logs -f deployment/goflow-service -n goflow
```

### Environment Variables

Key environment variables:

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goflow_db
DB_USER=goflow
DB_PASSWORD=your_password
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=goflow-events

# OpenAI
OPENAI_API_KEY=your_api_key
OPENAI_ORG_ID=your_org_id

# Security
JWT_SECRET=your_jwt_secret
JWT_EXPIRATION=3600

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

See [.env.example](.env.example) for complete configuration.

## 📊 Monitoring

### Metrics

GoFlow exposes Prometheus metrics at `/metrics`:

- Workflow execution count
- Execution duration
- Node execution count
- Error rates
- Queue depth
- Database connection pool stats
- Cache hit/miss rates

### Health Checks

- **Liveness**: `GET /health/live`
- **Readiness**: `GET /health/ready`

### Logging

Structured JSON logging with configurable levels:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "workflow execution started",
  "workflow_id": "wf_123",
  "execution_id": "exec_456",
  "user_id": "user_789"
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Quick Contribution Steps

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run tests and linting (`make test && make lint`)
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to your fork (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## 🔒 Security

Security is a top priority. Please see our [Security Policy](SECURITY.md) for:

- Reporting vulnerabilities
- Security best practices
- Supported versions

**Do not report security vulnerabilities through public GitHub issues.**

Email: [security@goflow-atom.dev](mailto:security@goflow-atom.dev)

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [Inngest](https://www.inngest.com/) - Durable execution platform
- [Testify](https://github.com/stretchr/testify) - Testing toolkit
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [Viper](https://github.com/spf13/viper) - Configuration management

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/goflow-atom/goflow-service/issues)
- **Discussions**: [GitHub Discussions](https://github.com/goflow-atom/goflow-service/discussions)
- **Email**: [support@goflow-atom.dev](mailto:support@goflow-atom.dev)

## 🗺️ Roadmap

See our [Implementation Roadmap](docs/tasks/01_IMPLEMENTATION_ROADMAP.md) for planned features and progress.

### Upcoming Features

- [ ] GraphQL API
- [ ] gRPC support
- [ ] WebSocket support for real-time updates
- [ ] Visual workflow editor
- [ ] Workflow templates marketplace
- [ ] Multi-region deployment
- [ ] Advanced analytics and reporting
- [ ] Custom node type SDK
- [ ] Workflow versioning and rollback
- [ ] A/B testing support

## 📈 Status

- **Version**: 0.1.0 (Alpha)
- **Status**: Active Development
- **Stability**: Experimental

## ⭐ Star History

If you find GoFlow useful, please consider giving it a star! ⭐

---

<div align="center">

**Built with ❤️ by the GoFlow Team**

[Website](https://goflow-atom.dev) •
[Documentation](docs/) •
[GitHub](https://github.com/goflow-atom/goflow-service) •
[Twitter](https://twitter.com/goflow_atom)

</div>

