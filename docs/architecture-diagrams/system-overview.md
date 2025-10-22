# Workflow Engine System Overview

## Introduction

The GoFlow Workflow Engine is a modular, event-driven workflow orchestration system built in Go. It provides a robust platform for defining, executing, and managing complex workflows as Directed Acyclic Graphs (DAGs). The architecture emphasizes scalability, reliability, and extensibility through a layered design pattern with clear separation of concerns.

### Design Philosophy

- **Modular Architecture**: Clear separation between API, business logic, execution engine, and infrastructure layers
- **Event-Driven**: Asynchronous execution with event streaming and durable execution guarantees
- **Extensible**: Interface-based design allows easy addition of new node types and integrations
- **Observable**: Comprehensive monitoring, logging, and tracing capabilities
- **Resilient**: Built-in retry mechanisms, circuit breakers, and state persistence for fault tolerance

## Architecture Layers

### 1. API Layer (Gin Framework)

The API layer provides RESTful HTTP endpoints for all workflow operations, built on the high-performance Gin web framework.

**Key Components:**

- **HTTP Handlers**: Process incoming requests and delegate to service layer
  - `WorkflowHandler`: CRUD operations for workflow definitions
  - `ExecutionHandler`: Workflow execution management and monitoring
  - `WebhookHandler`: Webhook trigger endpoints
  - `ScheduleHandler`: Cron schedule management

- **Middleware Stack**:
  - **Authentication Middleware**: JWT token validation and user identity extraction
  - **Authorization Middleware**: RBAC enforcement (admin, developer, operator, viewer roles)
  - **Logging Middleware**: Request/response logging with correlation IDs
  - **Error Handling Middleware**: Standardized error responses
  - **Rate Limiting Middleware**: Request throttling per user/IP
  - **CORS Middleware**: Cross-origin resource sharing configuration

- **Request Validation**:
  - JSON schema validation using `go-playground/validator`
  - Input sanitization and type checking
  - Workflow definition validation (DAG structure, node types, edge validity)

- **DTO Models**: Data Transfer Objects for API contracts
  - `CreateWorkflowRequest`, `UpdateWorkflowRequest`
  - `ExecuteWorkflowRequest`, `ExecutionResponse`
  - `WebhookTriggerRequest`, `ScheduleRequest`

**Endpoints:**
```
POST   /api/v1/workflows              - Create workflow
GET    /api/v1/workflows              - List workflows
GET    /api/v1/workflows/:id          - Get workflow by ID
PUT    /api/v1/workflows/:id          - Update workflow
DELETE /api/v1/workflows/:id          - Delete workflow
POST   /api/v1/workflows/:id/execute  - Execute workflow
GET    /api/v1/executions             - List executions
GET    /api/v1/executions/:id         - Get execution details
POST   /api/v1/executions/:id/cancel  - Cancel execution
POST   /webhooks/:path                - Webhook triggers
GET    /health                        - Health check
GET    /metrics                       - Prometheus metrics
```

### 2. Service Layer

The service layer encapsulates business logic and orchestrates operations across multiple repositories and external services.

**Key Services:**

- **WorkflowService**:
  - Workflow CRUD operations with validation
  - Version management and publishing
  - Workflow definition parsing and DAG validation
  - Cycle detection and dependency resolution

- **ExecutionService**:
  - Execution lifecycle management
  - Input/output data handling
  - Status tracking and updates
  - Execution cancellation and cleanup

- **ScheduleService**:
  - Cron expression parsing and validation
  - Schedule creation and management
  - Next run time calculation
  - Automatic workflow triggering

- **WebhookService**:
  - Webhook registration and configuration
  - HMAC signature validation
  - Webhook event routing to workflows

**Business Logic:**
- Transaction management across multiple repository operations
- Complex validation rules (workflow constraints, execution prerequisites)
- Data transformation between domain models and DTOs
- Event publishing for workflow lifecycle events

### 3. Workflow Engine Core

The heart of the system, responsible for executing workflows as DAGs with parallel processing capabilities.

**Key Components:**

- **DAG Executor**:
  - Topological sort for execution order determination
  - Dependency resolution between nodes
  - Parallel execution of independent nodes
  - Sequential execution of dependent nodes

- **Goroutine Worker Pool**:
  - Configurable worker count (`WORKER_POOL_SIZE`, default: 50)
  - Buffered job queue (`QUEUE_SIZE`, default: 1000)
  - Dynamic worker scaling based on load
  - Graceful shutdown with job completion

- **Execution Controls**:
  - **Retry Logic**: Exponential backoff with configurable max attempts
  - **Timeout Management**: Per-node and per-workflow timeouts
  - **Conditional Branching**: JavaScript expression evaluation for control flow
  - **Loop Execution**: Iteration over collections with parallel/sequential modes
  - **Parallel Execution**: Concurrent branch execution with synchronization

- **State Management**:
  - In-memory execution context during processing
  - Periodic checkpointing to database
  - State restoration after failures or restarts

**Execution Flow:**
```
1. Load workflow definition from database
2. Validate DAG structure (no cycles, valid edges)
3. Create execution context with input data
4. Perform topological sort to determine execution order
5. Initialize worker pool
6. Dispatch nodes to workers based on dependencies
7. Execute nodes with retry and timeout logic
8. Collect and aggregate node outputs
9. Update execution status and persist results
10. Trigger completion webhooks
```

### 4. Node Executor Factory

Implements the Factory pattern for polymorphic node execution, supporting 11+ node types with extensibility for custom nodes.

**Architecture:**

```go
type NodeExecutor interface {
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
    Validate() error
    Type() string
}

type NodeExecutorFactory struct {
    httpClient   *http.Client
    dbClient     *database.Client
    openaiClient *openai.Client
    emailClient  *email.Client
    redisClient  *redis.Client
}

func (f *NodeExecutorFactory) Create(node *domain.Node) (NodeExecutor, error) {
    switch node.Type {
    case "webhook":
        return NewWebhookExecutor(node.Config), nil
    case "http_request":
        return NewHTTPRequestExecutor(node.Config, f.httpClient), nil
    // ... other node types
    default:
        return nil, fmt.Errorf("unknown node type: %s", node.Type)
    }
}
```

**Supported Node Types:**

1. **Webhook**: Receives HTTP requests from external services
   - Methods: GET, POST, PUT, DELETE
   - Signature validation (HMAC-SHA256)
   - Custom response configuration

2. **HTTP Request**: Makes HTTP calls to external APIs
   - All HTTP methods supported
   - Header and query parameter configuration
   - Request/response body transformation
   - Timeout and retry configuration

3. **Conditional**: Branches execution based on expressions
   - JavaScript expression evaluation
   - Access to input data and previous node outputs
   - True/false branch routing

4. **Loop**: Iterates over collections
   - Sequential or parallel iteration
   - Item variable binding
   - Max iteration limits
   - Continue-on-error option

5. **Parallel**: Executes multiple branches concurrently
   - Independent branch execution
   - Wait-for-all or fail-fast modes
   - Result aggregation

6. **Transform**: Data transformation using JavaScript
   - Full JavaScript runtime (goja)
   - Access to input data and node outputs
   - Timeout protection

7. **Delay**: Pauses execution for specified duration
   - Seconds, minutes, hours units
   - Non-blocking implementation

8. **Database**: Executes SQL queries
   - PostgreSQL connection pooling
   - Parameterized queries (SQL injection prevention)
   - Transaction support
   - Query timeout configuration

9. **Email**: Sends emails via SMTP
   - HTML and plain text support
   - Attachments
   - Template rendering
   - CC/BCC recipients

10. **OpenAI Completion**: Generates text using GPT models
    - Model selection (GPT-4, GPT-3.5-turbo)
    - Temperature and token configuration
    - Streaming support

11. **OpenAI Embedding**: Creates vector embeddings
    - text-embedding-ada-002 model
    - Batch embedding support

**Extensibility:**
New node types can be added by:
1. Implementing the `NodeExecutor` interface
2. Registering in the factory's `Create` method
3. Adding configuration schema
4. Writing unit tests

### 5. Scheduler Service

Manages cron-based recurring workflow executions with state resumption for long-running workflows.

**Features:**

- **Cron Expression Parsing**: Standard cron syntax support (minute, hour, day, month, weekday)
- **Schedule Management**: Create, update, enable/disable schedules
- **Next Run Calculation**: Automatic calculation of next execution time
- **Execution Triggering**: Automatic workflow execution at scheduled times
- **Missed Execution Handling**: Configurable behavior for missed schedules
- **State Resumption**: Resume long-running workflows after system restarts

**Implementation:**
```go
type SchedulerService struct {
    repo          repository.ScheduleRepository
    workflowSvc   *WorkflowService
    executionSvc  *ExecutionService
    ticker        *time.Ticker
    stopChan      chan struct{}
}

func (s *SchedulerService) Start() {
    s.ticker = time.NewTicker(1 * time.Minute)
    go func() {
        for {
            select {
            case <-s.ticker.C:
                s.checkAndExecuteSchedules()
            case <-s.stopChan:
                return
            }
        }
    }()
}
```

### 6. Execution State Manager

Manages execution state with both in-memory and persistent storage for performance and durability.

**State Storage:**

- **In-Memory State**:
  - Active execution context during processing
  - Node input/output cache
  - Execution metadata (start time, current node, etc.)
  - Fast access for running workflows

- **Persistent State** (PostgreSQL):
  - Execution records in `workflow_executions` table
  - Node execution details in `node_executions` table
  - Execution logs in `execution_logs` table
  - Durable storage for recovery and auditing

**State Transitions:**
```
pending → running → completed
                 → failed
                 → cancelled
```

**Checkpointing:**
- Periodic state snapshots to database (every 10 seconds or after each node)
- Write-Ahead Log (WAL) pattern for crash recovery
- Atomic state updates using database transactions

### 7. Repository Layer

Provides data access abstraction for PostgreSQL operations using the Repository pattern.

**Key Repositories:**

- **WorkflowRepository**: CRUD operations for workflow definitions
- **ExecutionRepository**: Execution record management
- **NodeExecutionRepository**: Node-level execution tracking
- **ScheduleRepository**: Schedule management
- **WebhookRepository**: Webhook configuration storage

**Features:**
- Interface-based design for testability
- Transaction support with `sql.Tx`
- Query builder integration (sqlx)
- Connection pooling configuration
- Prepared statement caching

### 8. Infrastructure Layer

Integrates with external services and manages infrastructure dependencies.

**Components:**

- **PostgreSQL Client**:
  - Connection pool management
  - Transaction handling
  - Migration support (golang-migrate)
  - Query logging and metrics

- **Redis Client**:
  - Distributed locking (Redlock algorithm)
  - Caching layer for workflow definitions
  - Session storage
  - Rate limiting counters

- **Kafka Producer** (optional):
  - Event streaming for workflow lifecycle events
  - Async event publishing
  - Partition key strategy

- **Inngest Client**:
  - Durable function execution
  - Automatic retry with exponential backoff
  - Event-driven workflow triggers
  - State persistence across restarts

- **OpenAI Client**:
  - API key management
  - Request rate limiting
  - Response caching
  - Error handling and retries

- **SMTP Client**:
  - Email sending with TLS
  - Template rendering
  - Attachment handling

## External Dependencies

### PostgreSQL (Primary Database)

**Purpose**: Durable storage for all workflow data with ACID guarantees

**Usage:**
- Workflow definitions with JSONB for flexible schema
- Execution records and node results
- Audit logs and execution history
- Schedule configurations

**Configuration:**
```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goflow
DB_USER=goflow
DB_PASSWORD=secure_password
DB_SSL_MODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

**Features Used:**
- JSONB for flexible workflow definitions
- GIN indexes for fast JSON queries
- Foreign key constraints for referential integrity
- Transactions for consistency
- Connection pooling for performance

### Redis (Cache and Distributed Locking)

**Purpose**: Performance optimization and coordination

**Usage:**
- Distributed locking to prevent concurrent execution conflicts
- Caching frequently accessed workflow definitions
- Rate limiting counters
- Session storage for API authentication

**Configuration:**
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10
```

**Lock Pattern:**
```go
lock := redis.NewLock(client, "workflow:"+workflowID, 30*time.Second)
if err := lock.Acquire(); err != nil {
    return fmt.Errorf("failed to acquire lock: %w", err)
}
defer lock.Release()
```

### Kafka (Event Streaming)

**Purpose**: Asynchronous event streaming for workflow events (optional)

**Events Published:**
- `workflow.created`
- `workflow.updated`
- `workflow.deleted`
- `execution.started`
- `execution.completed`
- `execution.failed`
- `node.executed`

**Configuration:**
```bash
KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
KAFKA_TOPIC=workflow-events
KAFKA_CONSUMER_GROUP=goflow-consumers
```

### Inngest (Durable Execution)

**Purpose**: Event-driven durable execution with automatic retries and state persistence

**Features:**
- Automatic retry with exponential backoff (1s, 2s, 4s, 8s, 16s, 32s)
- Workflow resumption after process restarts
- Event-based triggering
- Built-in observability

**Configuration:**
```bash
INNGEST_EVENT_KEY=your_inngest_event_key
INNGEST_SIGNING_KEY=your_inngest_signing_key
INNGEST_APP_ID=goflow-service
```

**Integration:**
```go
inngest.CreateFunction(
    inngest.FunctionOpts{
        Name: "execute-workflow",
        Retries: 5,
    },
    inngest.EventTrigger("workflow/execute", nil),
    func(ctx context.Context, input inngest.Input[WorkflowEvent]) (interface{}, error) {
        return executionService.Execute(ctx, input.Event.Data.WorkflowID, input.Event.Data.Input)
    },
)
```

### OpenAI APIs

**Purpose**: AI-powered node types for text generation and embeddings

**Models Used:**
- GPT-4, GPT-3.5-turbo for completions
- text-embedding-ada-002 for embeddings

**Configuration:**
```bash
OPENAI_API_KEY=sk-...
OPENAI_ORG_ID=org-...
OPENAI_TIMEOUT=60s
OPENAI_MAX_RETRIES=3
```

### External Webhooks and HTTP APIs

**Purpose**: Integration with third-party services

**Features:**
- Configurable HTTP client with timeout and retry
- TLS certificate validation
- Custom headers and authentication
- Request/response logging

## Monitoring and Observability

### Prometheus Metrics

**Metrics Exposed** (at `/metrics` endpoint):

```
# Workflow metrics
goflow_workflow_executions_total{status="success|failed|cancelled"}
goflow_workflow_execution_duration_seconds{workflow_id}
goflow_workflow_active_executions{workflow_id}

# Node metrics
goflow_node_executions_total{type,status}
goflow_node_execution_duration_seconds{type}
goflow_node_retries_total{type}

# System metrics
goflow_worker_pool_active_workers
goflow_worker_pool_queue_size
goflow_worker_pool_queue_wait_duration_seconds

# HTTP metrics
goflow_http_requests_total{method,path,status}
goflow_http_request_duration_seconds{method,path}

# Database metrics
goflow_db_connections_open
goflow_db_connections_idle
goflow_db_query_duration_seconds{query}
```

### Grafana Dashboards

**Pre-built Dashboards** (in `monitoring/grafana/dashboards/`):

1. **Workflow Overview**: Execution metrics, success rates, duration percentiles
2. **Node Performance**: Node-level metrics by type, error rates
3. **System Health**: CPU, memory, goroutines, database connections
4. **API Performance**: HTTP request metrics, latency, error rates

### OpenTelemetry Tracing

**Trace Spans Created:**
- HTTP request handling
- Workflow execution lifecycle
- Individual node executions
- Database queries
- External API calls
- Cache operations

**Configuration:**
```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=goflow-service
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=0.1
```

**Trace Context Propagation:**
- W3C Trace Context headers
- Correlation IDs across services
- Parent-child span relationships

### Structured Logging (Zap)

**Log Levels:**
- `DEBUG`: Detailed debugging information
- `INFO`: General informational messages
- `WARN`: Warning messages for potential issues
- `ERROR`: Error messages for failures

**Log Format:**
```json
{
  "level": "info",
  "timestamp": "2024-01-01T10:00:00Z",
  "caller": "service/workflow_service.go:123",
  "message": "Workflow executed successfully",
  "workflow_id": "wf_123",
  "execution_id": "exec_456",
  "duration_ms": 1234,
  "trace_id": "abc123",
  "span_id": "def456"
}
```

## Data Flow Description

### Typical Workflow Execution Flow

**Step 1: Trigger**
- API call: `POST /api/v1/workflows/{workflow_id}/execute` with input JSON
- Webhook: `POST /webhooks/{path}` from external service
- Cron: Automatic trigger from SchedulerService

**Step 2: Authentication & Authorization**
- JWT token validation
- User identity extraction
- RBAC permission check (execute permission required)

**Step 3: Workflow Loading**
- Fetch workflow definition from PostgreSQL
- Parse JSONB definition into domain model
- Validate workflow status (must be 'published')

**Step 4: Execution Initialization**
- Create execution record with `status='pending'`
- Generate unique execution ID
- Store input data in `workflow_executions.input_data`
- Initialize execution context with variables and secrets

**Step 5: DAG Validation**
- Validate all nodes have valid types
- Check all edges reference existing nodes
- Detect cycles using depth-first search
- Build dependency graph

**Step 6: Worker Pool Spawning**
- Initialize worker pool with configured size
- Create buffered job queue
- Start worker goroutines

**Step 7: Node Execution**
- Perform topological sort to determine execution order
- Dispatch nodes to workers when dependencies are satisfied
- Execute node-specific logic via NodeExecutor interface
- Collect node output and store in execution context
- Update `node_executions` table with results
- Emit telemetry events

**Step 8: Status Updates**
- Update execution status after each node completion
- Persist intermediate state to database (checkpointing)
- Publish events to Kafka (if configured)
- Update Prometheus metrics

**Step 9: Result Storage**
- Aggregate outputs from all terminal nodes
- Store final result in `workflow_executions.output_data`
- Update execution status to 'completed' or 'failed'
- Record completion timestamp

**Step 10: Completion Webhooks**
- Trigger configured completion webhooks
- Send execution summary with status and outputs
- Retry webhook delivery on failure

**Step 11: Cleanup**
- Release distributed locks
- Clear in-memory execution context
- Close database transactions
- Shutdown worker pool gracefully

## Diagram Reference

For a visual representation of the system architecture, see [system-overview.png](./system-overview.png).

The diagram illustrates:
- All architecture layers and their interactions
- External dependencies and integration points
- Data flow between components
- Monitoring and observability stack

---

**Related Documentation:**
- [Database Schema](./database-schema.md)
- [Flow Diagram](./flow-diagram.md)
- [Architecture Overview](../architecture.md)
- [Deployment Guide](../guides/deployment.md)
