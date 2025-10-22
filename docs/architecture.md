# GoFlow Architecture

This document provides a comprehensive overview of the GoFlow Workflow Engine architecture, including system design, component responsibilities, data flow, and operational characteristics.

## Table of Contents

- [System Overview](#system-overview)
- [Components Overview](#components-overview)
- [Execution Lifecycle](#execution-lifecycle)
- [Concurrency Model](#concurrency-model)
- [Durability](#durability)
- [Error Handling](#error-handling)
- [Performance & Scaling](#performance--scaling)
- [Security Considerations](#security-considerations)

## System Overview

GoFlow is a distributed workflow orchestration engine designed to execute complex, multi-step workflows reliably and efficiently. The system follows a layered architecture pattern with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                            │
│  (REST API, gRPC, Client SDKs)                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      API Layer                               │
│  (Gin Router, Handlers, Middleware, Validation)            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
│  (Workflow Service, Execution Service, Scheduler)           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Engine Layer                              │
│  (Workflow Engine, DAG Resolver, Executor, State Manager)   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                        │
│  (PostgreSQL, Redis, Kafka, OpenAI, HTTP Client)           │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Separation of Concerns**: Each layer has distinct responsibilities
2. **Dependency Injection**: Components receive dependencies through constructors
3. **Interface-Based Design**: Components depend on interfaces, not concrete implementations
4. **Fail-Fast**: Validate early and provide clear error messages
5. **Observability**: Comprehensive logging, metrics, and tracing
6. **Scalability**: Horizontal scaling through stateless design

## Components Overview

### 1. API Layer (`internal/api`)

**Responsibility**: Handle HTTP requests, routing, authentication, and request/response transformation.

**Key Components**:
- **Router** (`router.go`): Gin-based HTTP router configuration
- **Handlers** (`handler/`): Request handlers for workflows, executions, webhooks
- **Middleware** (`middleware/`): Authentication, logging, rate limiting, CORS
- **DTOs** (`dto/`): Data Transfer Objects for API contracts

**Technologies**: Gin, JWT, OpenAPI

### 2. Service Layer (`internal/service`)

**Responsibility**: Implement business logic and coordinate between API and engine layers.

**Key Components**:
- **Workflow Service** (`workflow_service.go`): CRUD operations for workflows
- **Execution Service** (`execution_service.go`): Trigger and manage workflow executions
- **Scheduler Service** (`scheduler_service.go`): Cron-based workflow scheduling
- **Webhook Registry Service** (`webhook_registry_service.go`): Manage webhook endpoints
- **Notification Service** (`notification_service.go`): Send execution notifications
- **Node Executor Factory** (`node_executor_factory.go`): Create node executors

### 3. Engine Layer (`internal/engine`)

**Responsibility**: Core workflow execution logic, state management, and orchestration.

**Key Components**:

#### Workflow Engine (`engine.go`)
- Orchestrates workflow execution
- Manages execution context
- Coordinates with Inngest for durable execution

#### DAG Resolver (`dag_resolver.go`)
- Parses workflow definitions into Directed Acyclic Graphs
- Validates graph structure (no cycles, valid connections)
- Determines execution order based on dependencies

#### Executor (`executor.go`)
- Executes individual workflow nodes
- Manages node lifecycle (pending → running → completed/failed)
- Handles node-to-node data passing

#### State Manager (`state_manager.go`)
- Persists execution state to PostgreSQL
- Enables workflow resumption after failures
- Tracks node execution history

#### Context Builder (`context_builder.go`)
- Builds execution context with workflow data
- Provides data access for expression evaluation
- Manages variable scoping

#### Expression Evaluator (`expression_evaluator.go`)
- Evaluates dynamic expressions in node configurations
- Supports JSONPath, template strings, and conditional logic
- Provides access to workflow context and previous node outputs

#### Retry Handler (`retry_handler.go`)
- Implements exponential backoff retry logic
- Configurable retry policies per node type
- Tracks retry attempts and failures

#### Worker Pool (`worker_pool.go`)
- Manages goroutine pool for concurrent execution
- Configurable pool size and queue depth
- Load balancing across workers

#### Inngest Client (`inngest_client.go`)
- Integration with Inngest for durable execution
- Handles workflow pause/resume
- Manages long-running workflows

### 4. Domain Layer (`internal/domain`)

**Responsibility**: Define core business entities and domain logic.

**Key Entities**:

#### Workflow (`workflow.go`)
```go
type Workflow struct {
    ID          string
    Name        string
    Description string
    Version     int
    Nodes       []Node
    Edges       []Edge
    Config      WorkflowConfig
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### Execution (`execution.go`)
```go
type Execution struct {
    ID          string
    WorkflowID  string
    Status      ExecutionStatus
    Input       map[string]interface{}
    Output      map[string]interface{}
    StartedAt   time.Time
    CompletedAt *time.Time
    Error       *string
}
```

#### Node (`node.go`)
```go
type Node struct {
    ID       string
    Type     NodeType
    Config   map[string]interface{}
    Retry    RetryConfig
    Timeout  time.Duration
}
```

#### Node Types (`node_types/`)
- **Webhook** (`webhook.go`): Receive external HTTP requests
- **HTTP Request** (`http_request.go`): Make HTTP calls
- **Conditional** (`conditional.go`): Branch based on conditions
- **Loop** (`loop.go`): Iterate over collections
- **Parallel** (`parallel.go`): Execute nodes concurrently
- **Transform** (`transform.go`): Transform data
- **Delay** (`delay.go`): Wait for specified duration
- **Database** (`database.go`): Execute SQL queries
- **Email** (`email.go`): Send emails
- **OpenAI Completion** (`openai_completion.go`): Generate text
- **OpenAI Embedding** (`openai_embedding.go`): Create embeddings

### 5. Repository Layer (`internal/repository`)

**Responsibility**: Data access and persistence operations.

**Key Repositories**:
- **Workflow Repository** (`workflow_repo.go`): Workflow CRUD operations
- **Execution Repository** (`execution_repo.go`): Execution state management
- **Node Execution Repository** (`node_execution_repo.go`): Node-level execution tracking
- **Schedule Repository** (`schedule_repo.go`): Scheduled workflow management
- **Log Repository** (`log_repo.go`): Execution log storage

**Technologies**: PostgreSQL, GORM/sqlx

### 6. Infrastructure Layer (`internal/infrastructure`)

**Responsibility**: Integration with external services and systems.

**Key Components**:

#### Database (`database/`)
- **Connection Pool** (`connection_pool.go`): Manage database connections
- **PostgreSQL Client** (`postgres.go`): PostgreSQL operations
- **Transaction Manager** (`transaction.go`): Handle database transactions

#### Cache (`cache/`)
- **Redis Client** (`redis.go`): Redis operations
- **Cache Manager** (`cache_manager.go`): Caching strategies
- **Distributed Lock** (`distributed_lock.go`): Distributed locking for coordination

#### Queue (`queue/`)
- **Kafka Client** (`kafka.go`): Kafka producer/consumer
- **Queue Manager** (`queue_manager.go`): Message queue abstraction

#### HTTP (`http/`)
- **HTTP Client** (`client.go`): Configurable HTTP client
- **Retry Logic** (`retry.go`): HTTP request retry handling

#### OpenAI (`openai/`)
- **OpenAI Client** (`client.go`): OpenAI API integration
- **Completion** (`completion.go`): Text generation
- **Embedding** (`embedding.go`): Vector embeddings
- **Rate Limiter** (`rate_limiter.go`): API rate limiting

## Execution Lifecycle

The workflow execution follows a well-defined lifecycle:

### 1. Workflow Submission
```
Client → API Handler → Workflow Service → Validation → Repository
```
- Client submits workflow definition via REST API
- API handler validates request format
- Workflow service validates business rules
- Repository persists workflow to PostgreSQL

### 2. Execution Trigger
```
Trigger (API/Webhook/Schedule) → Execution Service → Engine → DAG Resolver
```
- Execution triggered by API call, webhook, or schedule
- Execution service creates execution record
- Engine receives execution request
- DAG resolver parses workflow into execution graph

### 3. Node Execution
```
Engine → Worker Pool → Node Executor → Infrastructure → State Manager
```
- Engine submits nodes to worker pool based on dependencies
- Worker picks up node from queue
- Node executor executes node logic
- Infrastructure layer handles external calls
- State manager persists node execution state

### 4. Data Flow
```
Node Output → Context Builder → Expression Evaluator → Next Node Input
```
- Node output stored in execution context
- Expression evaluator processes dynamic expressions
- Context builder provides data to downstream nodes
- Next nodes receive transformed data as input

### 5. Completion
```
Final Node → State Manager → Execution Service → Notification Service
```
- Final node completes execution
- State manager updates execution status
- Execution service triggers completion handlers
- Notification service sends alerts (if configured)

### Execution State Diagram

```
                    ┌─────────┐
                    │ PENDING │
                    └────┬────┘
                         │
                         ↓
                    ┌─────────┐
                    │ RUNNING │
                    └────┬────┘
                         │
            ┌────────────┼────────────┐
            ↓            ↓            ↓
       ┌─────────┐  ┌─────────┐  ┌─────────┐
       │ PAUSED  │  │COMPLETED│  │ FAILED  │
       └────┬────┘  └─────────┘  └─────────┘
            │
            ↓
       ┌─────────┐
       │ RUNNING │
       └─────────┘
```

## Concurrency Model

GoFlow leverages Go's concurrency primitives to achieve high-performance parallel execution.

### Worker Pool Architecture

```
                    ┌──────────────────┐
                    │  Workflow Engine │
                    └────────┬─────────┘
                             │
                             ↓
                    ┌──────────────────┐
                    │   Job Queue      │
                    │  (Buffered Chan) │
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              ↓              ↓              ↓
         ┌────────┐     ┌────────┐     ┌────────┐
         │Worker 1│     │Worker 2│ ... │Worker N│
         └────────┘     └────────┘     └────────┘
              │              │              │
              └──────────────┼──────────────┘
                             ↓
                    ┌──────────────────┐
                    │  Node Executors  │
                    └──────────────────┘
```

### Key Characteristics

1. **Fixed-Size Worker Pool**: Configurable number of goroutines (default: CPU cores * 2)
2. **Buffered Job Queue**: Prevents blocking when submitting jobs
3. **Graceful Shutdown**: Workers complete current tasks before terminating
4. **Context Propagation**: Cancellation signals propagate through execution tree

### Concurrency Control

```go
// Worker pool configuration
type WorkerPoolConfig struct {
    NumWorkers    int           // Number of concurrent workers
    QueueSize     int           // Job queue buffer size
    MaxRetries    int           // Maximum retry attempts
    RetryDelay    time.Duration // Delay between retries
}

// Execution with concurrency control
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflow *Workflow) error {
    // Create execution context with timeout
    execCtx, cancel := context.WithTimeout(ctx, workflow.Timeout)
    defer cancel()

    // Submit nodes to worker pool based on DAG dependencies
    for _, node := range resolvedDAG {
        select {
        case e.jobQueue <- Job{Node: node, Context: execCtx}:
            // Job submitted successfully
        case <-execCtx.Done():
            return execCtx.Err()
        }
    }

    return nil
}
```

### Parallel Node Execution

The **Parallel** node type enables concurrent execution of multiple branches:

```json
{
  "id": "parallel-1",
  "type": "parallel",
  "config": {
    "branches": [
      {"nodes": ["fetch-user", "process-user"]},
      {"nodes": ["fetch-orders", "process-orders"]},
      {"nodes": ["fetch-analytics", "process-analytics"]}
    ],
    "waitForAll": true,
    "failFast": false
  }
}
```

- **waitForAll**: Wait for all branches to complete
- **failFast**: Stop all branches if one fails

### Synchronization Mechanisms

1. **WaitGroups**: Coordinate parallel branch completion
2. **Channels**: Pass data between nodes
3. **Mutexes**: Protect shared state (execution context)
4. **Context**: Propagate cancellation signals

## Durability

GoFlow ensures workflow durability through multiple mechanisms:

### 1. Inngest Integration

**Inngest** provides durable execution capabilities:

- **Automatic Retries**: Failed steps retry automatically
- **Pause/Resume**: Long-running workflows can pause and resume
- **Event-Driven**: Workflows can wait for external events
- **Step Functions**: Each node is a durable step

```go
// Inngest function definition
func (e *Engine) RegisterWorkflow(workflow *Workflow) error {
    inngestFunc := inngest.CreateFunction(
        inngest.FunctionOpts{
            ID:   workflow.ID,
            Name: workflow.Name,
        },
        inngest.EventTrigger("workflow.execute", nil),
        e.executeWorkflowSteps,
    )

    return e.inngestClient.Register(inngestFunc)
}

// Durable step execution
func (e *Engine) executeWorkflowSteps(ctx context.Context, input inngest.Input) error {
    for _, node := range workflow.Nodes {
        // Each step is durable - if it fails, Inngest will retry
        result, err := inngest.Step(ctx, node.ID, func() (interface{}, error) {
            return e.executeNode(ctx, node)
        })

        if err != nil {
            return err
        }

        // Store result for next steps
        ctx = context.WithValue(ctx, node.ID, result)
    }

    return nil
}
```

### 2. State Persistence

All execution state is persisted to PostgreSQL:

```sql
-- Executions table
CREATE TABLE executions (
    id UUID PRIMARY KEY,
    workflow_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    input JSONB,
    output JSONB,
    error TEXT,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Node executions table
CREATE TABLE node_executions (
    id UUID PRIMARY KEY,
    execution_id UUID NOT NULL,
    node_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    input JSONB,
    output JSONB,
    error TEXT,
    retry_count INT DEFAULT 0,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    FOREIGN KEY (execution_id) REFERENCES executions(id)
);

-- Execution logs table
CREATE TABLE execution_logs (
    id BIGSERIAL PRIMARY KEY,
    execution_id UUID NOT NULL,
    node_id VARCHAR(255),
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (execution_id) REFERENCES executions(id)
);
```

### 3. Write-Ahead Logging

Critical state changes are logged before execution:

1. **Before Node Execution**: Log node start with input
2. **After Node Execution**: Log node completion with output
3. **On Error**: Log error details and stack trace
4. **On Retry**: Log retry attempt number

### 4. Idempotency

Nodes can be safely retried without side effects:

- **Idempotency Keys**: HTTP requests include unique keys
- **Conditional Writes**: Database operations use optimistic locking
- **Deduplication**: Duplicate events are detected and ignored

## Error Handling

Comprehensive error handling ensures system reliability:

### Error Types

```go
// Domain errors
type ErrorCode string

const (
    ErrWorkflowNotFound     ErrorCode = "WORKFLOW_NOT_FOUND"
    ErrInvalidWorkflow      ErrorCode = "INVALID_WORKFLOW"
    ErrExecutionFailed      ErrorCode = "EXECUTION_FAILED"
    ErrNodeExecutionFailed  ErrorCode = "NODE_EXECUTION_FAILED"
    ErrTimeout              ErrorCode = "TIMEOUT"
    ErrCancelled            ErrorCode = "CANCELLED"
    ErrRetryExhausted       ErrorCode = "RETRY_EXHAUSTED"
)

type DomainError struct {
    Code    ErrorCode
    Message string
    Cause   error
    Context map[string]interface{}
}
```

### Retry Strategies

#### Exponential Backoff
```go
type RetryConfig struct {
    MaxAttempts int           // Maximum retry attempts
    InitialDelay time.Duration // Initial delay
    MaxDelay     time.Duration // Maximum delay
    Multiplier   float64       // Backoff multiplier
}

func (r *RetryHandler) ExecuteWithRetry(ctx context.Context, fn func() error, config RetryConfig) error {
    delay := config.InitialDelay

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }

        if !isRetryable(err) {
            return err
        }

        select {
        case <-time.After(delay):
            delay = time.Duration(float64(delay) * config.Multiplier)
            if delay > config.MaxDelay {
                delay = config.MaxDelay
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    return ErrRetryExhausted
}
```

### Error Recovery

1. **Node-Level Recovery**: Individual nodes can fail without affecting the entire workflow
2. **Compensation Actions**: Failed nodes can trigger rollback operations
3. **Dead Letter Queue**: Failed executions are moved to DLQ for manual review
4. **Alerting**: Critical errors trigger notifications

### Circuit Breaker

Prevent cascading failures with circuit breaker pattern:

```go
type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    state        State // CLOSED, OPEN, HALF_OPEN
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if cb.state == OPEN {
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = HALF_OPEN
        } else {
            return ErrCircuitOpen
        }
    }

    err := fn()
    if err != nil {
        cb.recordFailure()
        return err
    }

    cb.recordSuccess()
    return nil
}
```

## Performance & Scaling

GoFlow is designed for high performance and horizontal scalability:

### Performance Optimizations

#### 1. Connection Pooling
```go
// Database connection pool
type DBConfig struct {
    MaxOpenConns    int           // Maximum open connections
    MaxIdleConns    int           // Maximum idle connections
    ConnMaxLifetime time.Duration // Connection lifetime
    ConnMaxIdleTime time.Duration // Idle connection timeout
}

// Recommended settings for high throughput
dbConfig := DBConfig{
    MaxOpenConns:    100,
    MaxIdleConns:    25,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 1 * time.Minute,
}
```

#### 2. Redis Caching

Cache frequently accessed data:

- **Workflow Definitions**: Cache parsed workflows (TTL: 5 minutes)
- **Node Configurations**: Cache node executor configurations
- **Expression Results**: Cache expression evaluation results
- **API Responses**: Cache external API responses (configurable TTL)

```go
// Cache workflow definition
func (s *WorkflowService) GetWorkflow(ctx context.Context, id string) (*Workflow, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("workflow:%s", id)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached.(*Workflow), nil
    }

    // Fetch from database
    workflow, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Store in cache
    s.cache.Set(ctx, cacheKey, workflow, 5*time.Minute)

    return workflow, nil
}
```

#### 3. Database Indexing

Critical indexes for query performance:

```sql
-- Workflow queries
CREATE INDEX idx_workflows_name ON workflows(name);
CREATE INDEX idx_workflows_created_at ON workflows(created_at DESC);

-- Execution queries
CREATE INDEX idx_executions_workflow_id ON executions(workflow_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_started_at ON executions(started_at DESC);
CREATE INDEX idx_executions_workflow_status ON executions(workflow_id, status);

-- Node execution queries
CREATE INDEX idx_node_executions_execution_id ON node_executions(execution_id);
CREATE INDEX idx_node_executions_node_id ON node_executions(node_id);
CREATE INDEX idx_node_executions_status ON node_executions(status);

-- Log queries
CREATE INDEX idx_execution_logs_execution_id ON execution_logs(execution_id);
CREATE INDEX idx_execution_logs_created_at ON execution_logs(created_at DESC);
CREATE INDEX idx_execution_logs_level ON execution_logs(level);
```

#### 4. Batch Operations

Reduce database round-trips with batch operations:

```go
// Batch insert node executions
func (r *NodeExecutionRepo) BatchCreate(ctx context.Context, executions []*NodeExecution) error {
    query := `
        INSERT INTO node_executions (id, execution_id, node_id, status, input, started_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    batch := &pgx.Batch{}
    for _, exec := range executions {
        batch.Queue(query, exec.ID, exec.ExecutionID, exec.NodeID, exec.Status, exec.Input, exec.StartedAt)
    }

    return r.db.SendBatch(ctx, batch).Close()
}
```

### Horizontal Scaling

#### Stateless Design

All application state is externalized:
- **Execution State**: PostgreSQL
- **Session State**: Redis
- **Job Queue**: Kafka

This enables running multiple instances behind a load balancer:

```
                    ┌──────────────┐
                    │ Load Balancer│
                    └──────┬───────┘
                           │
          ┌────────────────┼────────────────┐
          ↓                ↓                ↓
    ┌──────────┐     ┌──────────┐     ┌──────────┐
    │Instance 1│     │Instance 2│     │Instance N│
    └─────┬────┘     └─────┬────┘     └─────┬────┘
          │                │                │
          └────────────────┼────────────────┘
                           ↓
          ┌────────────────┴────────────────┐
          ↓                ↓                ↓
    ┌──────────┐     ┌──────────┐     ┌──────────┐
    │PostgreSQL│     │  Redis   │     │  Kafka   │
    └──────────┘     └──────────┘     └──────────┘
```

#### Distributed Locking

Prevent duplicate execution with Redis-based distributed locks:

```go
func (e *ExecutionService) TriggerWorkflow(ctx context.Context, workflowID string) error {
    lockKey := fmt.Sprintf("lock:workflow:%s", workflowID)

    // Acquire distributed lock
    lock, err := e.lockManager.Acquire(ctx, lockKey, 30*time.Second)
    if err != nil {
        return err
    }
    defer lock.Release()

    // Execute workflow
    return e.engine.Execute(ctx, workflowID)
}
```

#### Worker Pool Scaling

Configure worker pool size based on workload:

```yaml
# Development
worker_pool:
  num_workers: 10
  queue_size: 100

# Production
worker_pool:
  num_workers: 100
  queue_size: 1000
```

#### Database Scaling

- **Read Replicas**: Route read queries to replicas
- **Connection Pooling**: Reuse database connections
- **Partitioning**: Partition large tables by date
- **Archiving**: Move old executions to archive tables

### Performance Metrics

Key metrics to monitor:

1. **Throughput**: Workflows executed per second
2. **Latency**: P50, P95, P99 execution times
3. **Error Rate**: Failed executions percentage
4. **Queue Depth**: Pending jobs in worker queue
5. **Database Connections**: Active/idle connection count
6. **Cache Hit Rate**: Redis cache effectiveness
7. **Resource Usage**: CPU, memory, disk I/O

### Benchmarks

Expected performance characteristics:

- **Simple Workflow** (3 nodes): ~50ms
- **Complex Workflow** (20 nodes): ~500ms
- **Parallel Workflow** (10 branches): ~200ms
- **Throughput**: 1000+ workflows/second (with 100 workers)
- **Concurrent Executions**: 10,000+ simultaneous workflows

## Security Considerations

### Authentication & Authorization

#### JWT-Based Authentication

```go
// JWT middleware
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }

        claims, err := jwt.ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("roles", claims.Roles)
        c.Next()
    }
}
```

#### Role-Based Access Control (RBAC)

```go
type Permission string

const (
    PermissionWorkflowCreate Permission = "workflow:create"
    PermissionWorkflowRead   Permission = "workflow:read"
    PermissionWorkflowUpdate Permission = "workflow:update"
    PermissionWorkflowDelete Permission = "workflow:delete"
    PermissionWorkflowExecute Permission = "workflow:execute"
)

func RequirePermission(perm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        roles := c.GetStringSlice("roles")
        if !hasPermission(roles, perm) {
            c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
            return
        }
        c.Next()
    }
}
```

### Data Security

#### Encryption at Rest

Sensitive data encrypted in database:

```go
// Encrypt sensitive fields
func (w *Workflow) BeforeSave() error {
    if w.Secrets != nil {
        encrypted, err := crypto.Encrypt(w.Secrets)
        if err != nil {
            return err
        }
        w.EncryptedSecrets = encrypted
    }
    return nil
}
```

#### Encryption in Transit

- **HTTPS**: All API endpoints use TLS 1.3
- **Database**: PostgreSQL connections use SSL
- **Redis**: TLS-enabled Redis connections
- **Kafka**: SASL/SSL authentication

### Secret Management

Secrets stored securely and never logged:

```go
// Secret reference in workflow
{
  "id": "api-call",
  "type": "http_request",
  "config": {
    "url": "https://api.example.com",
    "headers": {
      "Authorization": "Bearer ${secrets.api_token}"
    }
  }
}

// Secret resolution
func (e *ExpressionEvaluator) ResolveSecrets(expr string) (string, error) {
    // Fetch from secret store (HashiCorp Vault, AWS Secrets Manager, etc.)
    secret, err := e.secretStore.Get(expr)
    if err != nil {
        return "", err
    }

    // Never log secret values
    e.logger.Info("resolved secret", zap.String("key", expr))

    return secret, nil
}
```

### Input Validation

Strict validation prevents injection attacks:

```go
// Validate workflow definition
func (v *WorkflowValidator) Validate(workflow *Workflow) error {
    // Validate structure
    if err := v.validateStructure(workflow); err != nil {
        return err
    }

    // Validate node configurations
    for _, node := range workflow.Nodes {
        if err := v.validateNode(node); err != nil {
            return fmt.Errorf("invalid node %s: %w", node.ID, err)
        }
    }

    // Validate expressions (prevent code injection)
    if err := v.validateExpressions(workflow); err != nil {
        return err
    }

    return nil
}
```

### Rate Limiting

Prevent abuse with rate limiting:

```go
// Rate limit middleware
func RateLimitMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.AbortWithStatusJSON(429, gin.H{
                "error": "rate limit exceeded",
                "retry_after": limiter.RetryAfter(),
            })
            return
        }
        c.Next()
    }
}
```

### Audit Logging

All operations are logged for audit:

```go
// Audit log entry
type AuditLog struct {
    Timestamp time.Time
    UserID    string
    Action    string
    Resource  string
    Result    string
    IPAddress string
    UserAgent string
}

// Log workflow execution
func (s *ExecutionService) Execute(ctx context.Context, workflowID string) error {
    userID := ctx.Value("user_id").(string)

    s.auditLogger.Log(AuditLog{
        Timestamp: time.Now(),
        UserID:    userID,
        Action:    "workflow.execute",
        Resource:  workflowID,
        IPAddress: ctx.Value("ip").(string),
    })

    return s.engine.Execute(ctx, workflowID)
}
```

## Related Diagrams

Visual representations of the architecture are available in `docs/architecture-diagrams/`:

1. **system-overview.png**: High-level system architecture showing all components and their interactions
2. **database-schema.png**: Entity-relationship diagram for PostgreSQL schema
3. **flow-diagram.png**: Detailed workflow execution flow across goroutines and node types
4. **deployment-architecture.png**: Production deployment topology with Kubernetes
5. **data-flow-diagram.png**: Data flow through the system from input to output

---

For more detailed information, see:
- [Getting Started Guide](./guides/getting-started.md)
- [Node Types Documentation](./guides/node-types.md)
- [Deployment Guide](./guides/deployment.md)
- [API Documentation](./api/)
