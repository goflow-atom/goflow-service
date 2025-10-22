# Database Schema for Workflow Engine

## Introduction

The GoFlow Workflow Engine uses PostgreSQL as its primary database for durable storage of workflow definitions, execution state, and audit logs. The schema is designed with ACID compliance, scalability, and performance in mind, leveraging PostgreSQL's advanced features including JSONB for flexible schema-less storage, GIN indexes for fast JSON queries, and robust transaction support.

### Design Principles

- **ACID Compliance**: All critical operations use transactions to ensure data consistency
- **Scalability**: Indexed columns and optimized queries for high-throughput operations
- **Flexibility**: JSONB columns allow schema evolution without migrations
- **Auditability**: Comprehensive logging and soft deletes for data recovery
- **Performance**: Strategic indexing and connection pooling for low-latency access

## Core Tables

### Table: `workflows`

**Purpose**: Store workflow template definitions and metadata

**Schema:**

```sql
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    definition JSONB NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT unique_workflow_name_version UNIQUE (name, version),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'published', 'archived', 'deprecated'))
);

CREATE INDEX idx_workflows_name ON workflows(name);
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflows_version ON workflows(version);
CREATE INDEX idx_workflows_created_at ON workflows(created_at);
CREATE INDEX idx_workflows_created_by ON workflows(created_by);
CREATE INDEX idx_workflows_definition_gin ON workflows USING GIN (definition);
```

**Column Descriptions:**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique workflow identifier |
| `name` | VARCHAR(255) | Human-readable workflow name |
| `version` | INTEGER | Version number for workflow evolution |
| `status` | VARCHAR(50) | Workflow lifecycle status (draft, published, archived, deprecated) |
| `definition` | JSONB | Complete workflow definition including nodes, edges, variables |
| `description` | TEXT | Optional detailed description |
| `created_by` | UUID | User who created the workflow |
| `created_at` | TIMESTAMP | Creation timestamp |
| `updated_at` | TIMESTAMP | Last modification timestamp |
| `deleted_at` | TIMESTAMP | Soft delete timestamp (NULL if not deleted) |

**JSONB Definition Structure:**
```json
{
  "nodes": [
    {
      "id": "node_1",
      "type": "webhook",
      "config": {...}
    }
  ],
  "edges": [
    {
      "from": "node_1",
      "to": "node_2"
    }
  ],
  "variables": {
    "api_url": "https://api.example.com"
  },
  "secrets": {
    "api_key": "encrypted_value"
  }
}
```

**Indexes:**
- `idx_workflows_name`: Fast lookup by workflow name
- `idx_workflows_status`: Filter workflows by status
- `idx_workflows_version`: Version-based queries
- `idx_workflows_created_at`: Time-based queries and sorting
- `idx_workflows_definition_gin`: Fast JSON queries on workflow definition

### Table: `workflow_executions`

**Purpose**: Track individual workflow execution instances

**Schema:**

```sql
CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    input_data JSONB,
    output_data JSONB,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    triggered_by VARCHAR(100),
    parent_execution_id UUID REFERENCES workflow_executions(id),
    CONSTRAINT valid_execution_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled', 'timeout'))
);

CREATE INDEX idx_executions_workflow_id ON workflow_executions(workflow_id);
CREATE INDEX idx_executions_status ON workflow_executions(status);
CREATE INDEX idx_executions_started_at ON workflow_executions(started_at);
CREATE INDEX idx_executions_completed_at ON workflow_executions(completed_at);
CREATE INDEX idx_executions_created_at ON workflow_executions(created_at);
CREATE INDEX idx_executions_triggered_by ON workflow_executions(triggered_by);
CREATE INDEX idx_executions_parent_id ON workflow_executions(parent_execution_id);
```

**Column Descriptions:**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique execution identifier |
| `workflow_id` | UUID | Reference to workflow definition |
| `status` | VARCHAR(50) | Current execution status |
| `input_data` | JSONB | Input parameters provided at execution start |
| `output_data` | JSONB | Final output after successful completion |
| `error_message` | TEXT | Error details if execution failed |
| `started_at` | TIMESTAMP | When execution actually started processing |
| `completed_at` | TIMESTAMP | When execution finished (success or failure) |
| `created_at` | TIMESTAMP | When execution record was created |
| `triggered_by` | VARCHAR(100) | Trigger source (api, webhook, cron, manual) |
| `parent_execution_id` | UUID | Parent execution for sub-workflows |

**Status Values:**
- `pending`: Execution created but not yet started
- `running`: Currently executing
- `completed`: Successfully finished
- `failed`: Execution failed with error
- `cancelled`: Manually cancelled by user
- `timeout`: Execution exceeded timeout limit

### Table: `node_executions`

**Purpose**: Store granular execution details for each node within a workflow execution

**Schema:**

```sql
CREATE TABLE node_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    input_data JSONB,
    output_data JSONB,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_node_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'skipped', 'cancelled'))
);

CREATE INDEX idx_node_executions_execution_id ON node_executions(execution_id);
CREATE INDEX idx_node_executions_node_id ON node_executions(node_id);
CREATE INDEX idx_node_executions_node_type ON node_executions(node_type);
CREATE INDEX idx_node_executions_status ON node_executions(status);
CREATE INDEX idx_node_executions_started_at ON node_executions(started_at);
CREATE INDEX idx_node_executions_composite ON node_executions(execution_id, node_id);
```

**Column Descriptions:**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique node execution identifier |
| `execution_id` | UUID | Parent workflow execution |
| `node_id` | VARCHAR(255) | Node ID from workflow definition |
| `node_type` | VARCHAR(100) | Type of node (webhook, http_request, etc.) |
| `status` | VARCHAR(50) | Node execution status |
| `input_data` | JSONB | Input data passed to this node |
| `output_data` | JSONB | Output data produced by this node |
| `error_message` | TEXT | Error details if node failed |
| `retry_count` | INTEGER | Number of retry attempts |
| `started_at` | TIMESTAMP | When node execution started |
| `completed_at` | TIMESTAMP | When node execution finished |
| `created_at` | TIMESTAMP | Record creation timestamp |

**Node Types Tracked:**
- `webhook`, `http_request`, `conditional`, `loop`, `parallel`, `transform`, `delay`, `database`, `email`, `openai_completion`, `openai_embedding`

### Table: `execution_logs`

**Purpose**: Store detailed logs for debugging and auditing

**Schema:**

```sql
CREATE TABLE execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    node_id VARCHAR(255),
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_log_level CHECK (level IN ('debug', 'info', 'warn', 'error'))
);

CREATE INDEX idx_logs_execution_id ON execution_logs(execution_id);
CREATE INDEX idx_logs_node_id ON execution_logs(node_id);
CREATE INDEX idx_logs_level ON execution_logs(level);
CREATE INDEX idx_logs_created_at ON execution_logs(created_at);
CREATE INDEX idx_logs_composite ON execution_logs(execution_id, created_at);
```

**Column Descriptions:**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique log entry identifier |
| `execution_id` | UUID | Associated workflow execution |
| `node_id` | VARCHAR(255) | Associated node (NULL for workflow-level logs) |
| `level` | VARCHAR(20) | Log level (debug, info, warn, error) |
| `message` | TEXT | Log message |
| `metadata` | JSONB | Additional structured data |
| `created_at` | TIMESTAMP | Log entry timestamp |

**Metadata Examples:**
```json
{
  "http_status": 200,
  "response_time_ms": 123,
  "retry_attempt": 2,
  "error_code": "TIMEOUT"
}
```

### Table: `workflow_schedules`

**Purpose**: Manage cron-based recurring workflow triggers

**Schema:**

```sql
CREATE TABLE workflow_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    cron_expression VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    input_data JSONB,
    next_run_at TIMESTAMP,
    last_run_at TIMESTAMP,
    last_execution_id UUID REFERENCES workflow_executions(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

CREATE INDEX idx_schedules_workflow_id ON workflow_schedules(workflow_id);
CREATE INDEX idx_schedules_next_run_at ON workflow_schedules(next_run_at);
CREATE INDEX idx_schedules_enabled ON workflow_schedules(enabled);
CREATE INDEX idx_schedules_composite ON workflow_schedules(enabled, next_run_at);
```

**Column Descriptions:**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique schedule identifier |
| `workflow_id` | UUID | Workflow to execute |
| `cron_expression` | VARCHAR(100) | Cron syntax (e.g., "0 0 * * *") |
| `enabled` | BOOLEAN | Whether schedule is active |
| `input_data` | JSONB | Default input data for scheduled executions |
| `next_run_at` | TIMESTAMP | Next scheduled execution time |
| `last_run_at` | TIMESTAMP | Last execution time |
| `last_execution_id` | UUID | Reference to last execution |
| `created_at` | TIMESTAMP | Schedule creation timestamp |
| `updated_at` | TIMESTAMP | Last modification timestamp |
| `created_by` | UUID | User who created the schedule |

**Cron Expression Examples:**
- `0 0 * * *` - Daily at midnight
- `*/15 * * * *` - Every 15 minutes
- `0 9 * * 1-5` - Weekdays at 9 AM

### Table: `webhooks`

**Purpose**: Store webhook configurations for workflow triggers

**Schema:**

```sql
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    path VARCHAR(255) NOT NULL UNIQUE,
    method VARCHAR(10) NOT NULL DEFAULT 'POST',
    secret VARCHAR(255),
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_triggered_at TIMESTAMP,
    trigger_count INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT valid_http_method CHECK (method IN ('GET', 'POST', 'PUT', 'DELETE', 'PATCH'))
);

CREATE INDEX idx_webhooks_workflow_id ON webhooks(workflow_id);
CREATE INDEX idx_webhooks_path ON webhooks(path);
CREATE INDEX idx_webhooks_enabled ON webhooks(enabled);
```

**Column Descriptions:**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique webhook identifier |
| `workflow_id` | UUID | Workflow to trigger |
| `path` | VARCHAR(255) | Webhook URL path (e.g., "/webhooks/user-signup") |
| `method` | VARCHAR(10) | HTTP method |
| `secret` | VARCHAR(255) | HMAC secret for signature validation |
| `enabled` | BOOLEAN | Whether webhook is active |
| `config` | JSONB | Additional configuration (headers, validation rules) |
| `created_at` | TIMESTAMP | Webhook creation timestamp |
| `updated_at` | TIMESTAMP | Last modification timestamp |
| `last_triggered_at` | TIMESTAMP | Last trigger timestamp |
| `trigger_count` | INTEGER | Total number of triggers |

### Table: `users`

**Purpose**: Store user accounts for authentication and authorization

**Schema:**

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP,
    CONSTRAINT valid_role CHECK (role IN ('admin', 'developer', 'operator', 'viewer'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_enabled ON users(enabled);
```

**Column Descriptions:**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique user identifier |
| `email` | VARCHAR(255) | User email (login credential) |
| `password_hash` | VARCHAR(255) | Bcrypt password hash |
| `name` | VARCHAR(255) | User display name |
| `role` | VARCHAR(50) | RBAC role (admin, developer, operator, viewer) |
| `enabled` | BOOLEAN | Account status |
| `created_at` | TIMESTAMP | Account creation timestamp |
| `updated_at` | TIMESTAMP | Last modification timestamp |
| `last_login_at` | TIMESTAMP | Last successful login |

**Role Permissions:**
- `admin`: Full access to all operations
- `developer`: Create, update, delete workflows; execute workflows
- `operator`: Execute workflows, view executions
- `viewer`: Read-only access

## Entity Relationships

### Relationship Diagram

```
users
  ↓ (created_by)
workflows ←──────────────────┐
  ↓ (workflow_id)            │
  ├─→ workflow_executions    │
  │     ↓ (execution_id)     │
  │     ├─→ node_executions  │
  │     └─→ execution_logs   │
  │                           │
  ├─→ workflow_schedules ─────┘
  │     ↓ (last_execution_id)
  │     └─→ workflow_executions
  │
  └─→ webhooks
```

### Relationship Details

| Parent Table | Child Table | Relationship Type | Foreign Key | Cascade Behavior |
|--------------|-------------|-------------------|-------------|------------------|
| `workflows` | `workflow_executions` | One-to-Many | `workflow_id` | CASCADE |
| `workflow_executions` | `node_executions` | One-to-Many | `execution_id` | CASCADE |
| `workflow_executions` | `execution_logs` | One-to-Many | `execution_id` | CASCADE |
| `workflows` | `workflow_schedules` | One-to-Many | `workflow_id` | CASCADE |
| `workflows` | `webhooks` | One-to-Many | `workflow_id` | CASCADE |
| `users` | `workflows` | One-to-Many | `created_by` | SET NULL |
| `users` | `workflow_schedules` | One-to-Many | `created_by` | SET NULL |
| `workflow_executions` | `workflow_executions` | Self-referencing | `parent_execution_id` | SET NULL |
| `workflow_schedules` | `workflow_executions` | One-to-Many | `last_execution_id` | SET NULL |

### Cascade Behavior Explanation

**ON DELETE CASCADE:**
- When a workflow is deleted, all associated executions, schedules, and webhooks are automatically deleted
- When an execution is deleted, all associated node executions and logs are automatically deleted
- Ensures referential integrity and prevents orphaned records

**ON DELETE SET NULL:**
- When a user is deleted, workflows they created remain but `created_by` is set to NULL
- Preserves workflow history even after user account deletion

## Advanced Database Features

### JSONB Usage

**Purpose**: Flexible schema-less storage for complex data structures

**Benefits:**
- Store workflow definitions without rigid schema
- Query nested JSON data efficiently
- Index JSON fields for fast lookups
- Schema evolution without migrations

**Example Queries:**

```sql
-- Find workflows with specific node type
SELECT * FROM workflows
WHERE definition @> '{"nodes": [{"type": "openai_completion"}]}';

-- Query workflow variables
SELECT name, definition->'variables' AS variables
FROM workflows
WHERE definition->'variables'->>'api_url' LIKE '%example.com%';

-- Find executions with specific input
SELECT * FROM workflow_executions
WHERE input_data @> '{"user_id": "123"}';

-- Query node execution outputs
SELECT node_id, output_data->'result' AS result
FROM node_executions
WHERE execution_id = 'exec_123';
```

### GIN Indexes

**Purpose**: Fast indexing and querying of JSONB columns

**Indexes Created:**
```sql
CREATE INDEX idx_workflows_definition_gin ON workflows USING GIN (definition);
CREATE INDEX idx_executions_input_gin ON workflow_executions USING GIN (input_data);
CREATE INDEX idx_executions_output_gin ON workflow_executions USING GIN (output_data);
CREATE INDEX idx_node_executions_output_gin ON node_executions USING GIN (output_data);
```

**Performance Impact:**
- 10-100x faster JSON queries compared to sequential scans
- Supports containment operators (`@>`, `<@`)
- Supports existence operators (`?`, `?|`, `?&`)
- Supports path/value operators (`@?`, `@@`)

### Soft Deletes

**Purpose**: Reversible deletion with audit trail

**Implementation:**
```sql
-- Soft delete a workflow
UPDATE workflows
SET deleted_at = NOW()
WHERE id = 'wf_123';

-- Query only active workflows
SELECT * FROM workflows
WHERE deleted_at IS NULL;

-- Restore a soft-deleted workflow
UPDATE workflows
SET deleted_at = NULL
WHERE id = 'wf_123';

-- Permanently delete old soft-deleted records
DELETE FROM workflows
WHERE deleted_at < NOW() - INTERVAL '90 days';
```

**Benefits:**
- Accidental deletion recovery
- Audit trail preservation
- Compliance with data retention policies

### Transaction Management

**Purpose**: Ensure data consistency across multiple operations

**Example Transaction:**
```go
func (r *WorkflowRepository) CreateWithSchedule(ctx context.Context, workflow *Workflow, schedule *Schedule) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Insert workflow
    if err := r.insertWorkflow(ctx, tx, workflow); err != nil {
        return err
    }

    // Insert schedule
    if err := r.insertSchedule(ctx, tx, schedule); err != nil {
        return err
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

**Use Cases:**
- Creating workflow with associated schedules/webhooks
- Updating execution status with node results
- Batch operations requiring atomicity

### Distributed Locking (Redis)

**Purpose**: Prevent race conditions during concurrent workflow executions

**Implementation:**
```go
func (s *ExecutionService) Execute(ctx context.Context, workflowID string, input map[string]interface{}) error {
    // Acquire distributed lock
    lockKey := fmt.Sprintf("workflow:lock:%s", workflowID)
    lock := redis.NewLock(s.redisClient, lockKey, 30*time.Second)

    if err := lock.Acquire(); err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }
    defer lock.Release()

    // Execute workflow with exclusive access
    return s.executeWorkflow(ctx, workflowID, input)
}
```

**Use Cases:**
- Prevent duplicate executions from concurrent triggers
- Ensure schedule execution happens only once
- Coordinate distributed workers

### Connection Pooling

**Purpose**: Optimize database connection usage and performance

**Configuration:**
```go
db.SetMaxOpenConns(25)        // Maximum open connections
db.SetMaxIdleConns(5)         // Maximum idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
db.SetConnMaxIdleTime(1 * time.Minute)  // Idle connection timeout
```

**Environment Variables:**
```bash
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=1m
```

**Best Practices:**
- Set `MaxOpenConns` based on database server capacity
- Keep `MaxIdleConns` lower to reduce resource usage
- Use `ConnMaxLifetime` to prevent stale connections
- Monitor connection pool metrics via Prometheus

## Migration Strategy

### Migration Tool

The project uses **golang-migrate** for database schema migrations.

**Installation:**
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Migration Files

Migrations are stored in `migrations/` directory with sequential numbering:

```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_workflows_table.up.sql
├── 000002_create_workflows_table.down.sql
├── 000003_create_workflow_executions_table.up.sql
├── 000003_create_workflow_executions_table.down.sql
└── ...
```

### Migration Commands

**Apply all pending migrations:**
```bash
migrate -path migrations -database "postgresql://user:pass@localhost:5432/goflow?sslmode=disable" up
```

**Rollback last migration:**
```bash
migrate -path migrations -database "postgresql://user:pass@localhost:5432/goflow?sslmode=disable" down 1
```

**Check migration version:**
```bash
migrate -path migrations -database "postgresql://user:pass@localhost:5432/goflow?sslmode=disable" version
```

**Force migration version (use with caution):**
```bash
migrate -path migrations -database "postgresql://user:pass@localhost:5432/goflow?sslmode=disable" force 5
```

### Versioning Approach

**Naming Convention:**
```
{version}_{description}.{up|down}.sql
```

**Example Migration (000001_create_users_table.up.sql):**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP,
    CONSTRAINT valid_role CHECK (role IN ('admin', 'developer', 'operator', 'viewer'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_enabled ON users(enabled);
```

**Example Rollback (000001_create_users_table.down.sql):**
```sql
DROP TABLE IF EXISTS users CASCADE;
```

### Schema Change Best Practices

1. **Always create both up and down migrations**
   - Enables rollback in case of issues
   - Maintains migration reversibility

2. **Test migrations in development first**
   - Verify migration applies cleanly
   - Test rollback functionality
   - Check data integrity after migration

3. **Use transactions for complex migrations**
   ```sql
   BEGIN;

   -- Migration statements
   ALTER TABLE workflows ADD COLUMN new_field VARCHAR(255);
   UPDATE workflows SET new_field = 'default_value';
   ALTER TABLE workflows ALTER COLUMN new_field SET NOT NULL;

   COMMIT;
   ```

4. **Avoid breaking changes in production**
   - Add new columns as nullable first
   - Backfill data in separate step
   - Make column NOT NULL in subsequent migration

5. **Document migration dependencies**
   - Note any required data transformations
   - Document expected downtime
   - Include rollback procedures

### Zero-Downtime Migrations

**Strategy for adding a new column:**

**Step 1: Add nullable column**
```sql
ALTER TABLE workflows ADD COLUMN new_field VARCHAR(255);
```

**Step 2: Deploy application code that writes to new column**
```go
// Application now writes to both old and new fields
```

**Step 3: Backfill existing data**
```sql
UPDATE workflows SET new_field = old_field WHERE new_field IS NULL;
```

**Step 4: Make column NOT NULL**
```sql
ALTER TABLE workflows ALTER COLUMN new_field SET NOT NULL;
```

**Step 5: Remove old column (in future migration)**
```sql
ALTER TABLE workflows DROP COLUMN old_field;
```

### Migration Monitoring

**Track migration status:**
```sql
SELECT * FROM schema_migrations;
```

**Monitor migration performance:**
- Log migration execution time
- Alert on migrations taking > 5 minutes
- Monitor database locks during migration

## Performance Optimization

### Query Optimization Tips

1. **Use indexes effectively**
   ```sql
   -- Good: Uses index
   SELECT * FROM workflows WHERE status = 'published';

   -- Bad: Full table scan
   SELECT * FROM workflows WHERE LOWER(name) = 'my workflow';
   ```

2. **Limit result sets**
   ```sql
   -- Good: Pagination
   SELECT * FROM workflow_executions
   ORDER BY created_at DESC
   LIMIT 50 OFFSET 0;

   -- Bad: Fetching all records
   SELECT * FROM workflow_executions;
   ```

3. **Use EXPLAIN ANALYZE**
   ```sql
   EXPLAIN ANALYZE
   SELECT * FROM workflows
   WHERE definition @> '{"nodes": [{"type": "webhook"}]}';
   ```

### Index Maintenance

**Rebuild indexes periodically:**
```sql
REINDEX TABLE workflows;
REINDEX TABLE workflow_executions;
```

**Monitor index usage:**
```sql
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan ASC;
```

**Remove unused indexes:**
```sql
-- Find indexes with zero scans
SELECT schemaname, tablename, indexname
FROM pg_stat_user_indexes
WHERE idx_scan = 0 AND schemaname = 'public';
```

### Vacuum and Analyze

**Regular maintenance:**
```sql
-- Vacuum to reclaim storage
VACUUM ANALYZE workflows;
VACUUM ANALYZE workflow_executions;

-- Full vacuum (requires table lock)
VACUUM FULL workflows;
```

**Autovacuum configuration:**
```sql
ALTER TABLE workflow_executions SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE node_executions SET (autovacuum_vacuum_scale_factor = 0.1);
```

## Diagram Reference

For a visual representation of the database schema, see [database-schema.png](./database-schema.png).

The Entity-Relationship Diagram (ERD) illustrates:
- All tables and their columns
- Primary and foreign key relationships
- Cardinality (one-to-many, many-to-one)
- Cascade behaviors
- Index strategies

---

**Related Documentation:**
- [System Overview](./system-overview.md)
- [Flow Diagram](./flow-diagram.md)
- [Architecture Overview](../architecture.md)
- [Deployment Guide](../guides/deployment.md)
