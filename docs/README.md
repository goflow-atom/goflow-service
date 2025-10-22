# GoFlow Workflow Engine Documentation

Welcome to the GoFlow Workflow Engine documentation. This guide will help you understand, deploy, and extend the workflow orchestration system.

## 📋 Project Overview

GoFlow is a powerful, scalable workflow orchestration engine built in Go that enables you to design, execute, and monitor complex workflows with ease. It provides a flexible node-based architecture that supports various operations including HTTP requests, webhooks, conditional logic, loops, parallel execution, and AI-powered transformations using OpenAI.

## ✨ Key Features

- **Concurrent Workflow Execution**: Execute multiple workflows simultaneously with efficient goroutine-based worker pools
- **Rich Node Types**: Support for diverse operations including:
  - **Webhook**: Trigger workflows from external events
  - **HTTP Request**: Make REST API calls with full configuration
  - **Conditional**: Branch execution based on dynamic conditions
  - **Loop**: Iterate over collections with configurable behavior
  - **Parallel**: Execute multiple nodes concurrently
  - **Transform**: Manipulate and transform data using expressions
  - **Delay**: Add time-based pauses in workflow execution
  - **Database**: Execute SQL queries and commands
  - **Email**: Send notifications and alerts
  - **OpenAI Completion**: Generate text using GPT models
  - **OpenAI Embedding**: Create vector embeddings for semantic search
- **Durable Execution**: Integration with Inngest for reliable, resumable workflows
- **DAG-based Orchestration**: Define workflows as Directed Acyclic Graphs
- **Expression Evaluation**: Dynamic data transformation and condition evaluation
- **Retry & Error Handling**: Configurable retry policies and error recovery
- **State Management**: Persistent workflow state with PostgreSQL
- **Caching Layer**: Redis-based caching for performance optimization
- **Event-Driven Architecture**: Kafka integration for event streaming
- **RESTful API**: Comprehensive API for workflow management
- **Monitoring & Observability**: Built-in metrics and logging support

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 6+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/goflow-atom/goflow-service.git
   cd goflow-service
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start dependencies with Docker Compose**
   ```bash
   docker-compose up -d postgres redis kafka
   ```

4. **Run database migrations**
   ```bash
   make migrate-up
   ```

5. **Start the service**
   ```bash
   make run
   # Or for hot-reload development
   make dev
   ```

6. **Verify the service is running**
   ```bash
   curl http://localhost:8080/health
   ```

## 📁 Directory Structure

```
goflow-service/
├── api/                    # API specifications
│   ├── openapi.yaml       # OpenAPI 3.0 specification
│   ├── postman/           # Postman collections
│   └── proto/             # Protocol Buffer definitions
├── cmd/                   # Application entry points
│   └── server/            # Main server application
├── configs/               # Configuration files
│   ├── config.yaml        # Default configuration
│   ├── config.dev.yaml    # Development config
│   ├── config.staging.yaml
│   └── config.prod.yaml
├── deployments/           # Deployment configurations
│   ├── docker/            # Dockerfiles
│   ├── kubernetes/        # K8s manifests
│   └── terraform/         # Infrastructure as Code
├── docs/                  # Documentation
│   ├── api/               # API documentation
│   ├── guides/            # User guides
│   ├── development/       # Developer guides
│   └── architecture-diagrams/
├── examples/              # Example workflows and clients
│   ├── workflows/         # Sample workflow definitions
│   ├── clients/           # Client SDK examples
│   └── scripts/           # Utility scripts
├── internal/              # Private application code
│   ├── api/               # HTTP handlers and routing
│   ├── config/            # Configuration management
│   ├── domain/            # Domain models and business logic
│   ├── engine/            # Workflow execution engine
│   ├── infrastructure/    # External service integrations
│   ├── repository/        # Data access layer
│   └── service/           # Business services
├── pkg/                   # Public libraries
│   ├── client/            # Go client SDK
│   ├── constants/         # Shared constants
│   ├── crypto/            # Cryptography utilities
│   ├── logger/            # Logging utilities
│   ├── utils/             # Common utilities
│   └── validator/         # Validation utilities
├── test/                  # Test suites
│   ├── integration/       # Integration tests
│   ├── e2e/               # End-to-end tests
│   └── mocks/             # Test mocks
└── tools/                 # Development tools
    ├── code-generator/    # Code generation utilities
    └── monitoring/        # Monitoring configurations
```

## 🛠 Technology Stack

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

## 📚 Documentation

- **[Architecture](./architecture.md)**: System design and component overview
- **[Getting Started](./guides/getting-started.md)**: Detailed setup instructions
- **[Workflow Definition](./guides/workflow-definition.md)**: How to define workflows
- **[Node Types](./guides/node-types.md)**: Available node types and usage
- **[Deployment](./guides/deployment.md)**: Production deployment guide
- **[API Reference](./api/)**: REST API documentation
- **[Contributing](./development/contributing.md)**: Contribution guidelines
- **[Code Style](./development/code-style.md)**: Coding standards
- **[Testing](./development/testing.md)**: Testing strategies

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](./development/contributing.md) for details on:
- Code of Conduct
- Development workflow
- Pull request process
- Coding standards

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check our comprehensive [guides](./guides/)
- **Issues**: Report bugs or request features on [GitHub Issues](https://github.com/yourusername/goflow-service/issues)
- **Discussions**: Join conversations on [GitHub Discussions](https://github.com/yourusername/goflow-service/discussions)
- **Troubleshooting**: See our [Troubleshooting Guide](./guides/troubleshooting.md)

## 🔗 Quick Links

- [API Documentation](./api/)
- [Example Workflows](../examples/workflows/)
- [Client SDKs](../examples/clients/)
- [Deployment Templates](../deployments/)
- [Monitoring Setup](../tools/monitoring/)

---

**Built with ❤️ by the GoFlow Team**
