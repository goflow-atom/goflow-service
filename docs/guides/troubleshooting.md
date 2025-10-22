# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with GoFlow.

## Table of Contents

- [Common Issues](#common-issues)
  - [Database Connection Issues](#database-connection-issues)
  - [Node Execution Failures](#node-execution-failures)
  - [Workflow Stuck or Hanging](#workflow-stuck-or-hanging)
  - [Memory Leaks](#memory-leaks)
  - [Performance Issues](#performance-issues)
- [Debugging Guide](#debugging-guide)
  - [Logs](#logs)
  - [Distributed Tracing](#distributed-tracing)
  - [Metrics](#metrics)
- [Resiliency Features](#resiliency-features)
  - [Automatic Retries](#automatic-retries)
  - [State Restoration](#state-restoration)
  - [Circuit Breaker](#circuit-breaker)
- [Performance Troubleshooting](#performance-troubleshooting)
- [Escalation Procedures](#escalation-procedures)

## Common Issues

### Database Connection Issues

#### Symptom

```
Error: failed to connect to database: dial tcp: lookup postgres on 127.0.0.1:53: no such host
```

#### Possible Causes

1. Database host is incorrect or unreachable
2. Database credentials are invalid
3. Network connectivity issues
4. Database is not running

#### Solutions

**1. Verify database connection:**

```bash
# Test PostgreSQL connection
psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Check if database is running
docker ps | grep postgres
kubectl get pods -n goflow | grep postgres
```

**2. Check environment variables:**

```bash
echo $DB_HOST
echo $DB_PORT
echo $DB_NAME
echo $DB_USER
```

**3. Verify network connectivity:**

```bash
# Ping database host
ping $DB_HOST

# Test port connectivity
telnet $DB_HOST $DB_PORT
nc -zv $DB_HOST $DB_PORT
```

**4. Check database logs:**

```bash
# Docker
docker logs postgres

# Kubernetes
kubectl logs -n goflow postgres-0
```

**5. Verify SSL configuration:**

```bash
# If using SSL, ensure certificates are valid
DB_SSL_MODE=require
DB_SSL_CERT=/path/to/client-cert.pem
DB_SSL_KEY=/path/to/client-key.pem
DB_SSL_ROOT_CERT=/path/to/ca-cert.pem
```

### Node Execution Failures

#### Symptom

```
Error: node execution failed: http_request node failed: context deadline exceeded
```

#### Possible Causes

1. External API is slow or unresponsive
2. Timeout is too short
3. Network issues
4. Invalid configuration

#### Solutions

**1. Check node configuration:**

```bash
# View workflow definition
curl http://localhost:8080/api/v1/workflows/{workflow_id}

# Check node timeout settings
```

**2. Increase timeout:**

```json
{
  "id": "api_call",
  "type": "http_request",
  "config": {
    "url": "https://api.example.com",
    "timeout": 60
  }
}
```

**3. Enable retries:**

```json
{
  "id": "api_call",
  "type": "http_request",
  "config": {
    "url": "https://api.example.com",
    "retry": {
      "max_attempts": 3,
      "initial_delay": "1s",
      "max_delay": "10s",
      "multiplier": 2
    }
  }
}
```

**4. Check external API status:**

```bash
# Test API directly
curl -v https://api.example.com

# Check API response time
time curl https://api.example.com
```

**5. Review execution logs:**

```bash
# Get execution details
curl http://localhost:8080/api/v1/executions/{execution_id}

# View logs
kubectl logs -n goflow deployment/goflow-service | grep execution_id
```

### Workflow Stuck or Hanging

#### Symptom

Workflow execution status remains "running" indefinitely.

#### Possible Causes

1. Deadlock in workflow graph
2. Node is waiting for external event
3. Worker pool exhausted
4. Database connection pool exhausted

#### Solutions

**1. Check execution status:**

```bash
# Get execution details
curl http://localhost:8080/api/v1/executions/{execution_id}

# Check current node
```

**2. Review workflow graph:**

```bash
# Visualize workflow DAG
# Check for circular dependencies or missing edges
```

**3. Check worker pool status:**

```bash
# View metrics
curl http://localhost:8080/metrics | grep worker_pool

# Check active workers
goflow_worker_pool_active_workers
goflow_worker_pool_queue_size
```

**4. Increase worker pool size:**

```yaml
worker_pool:
  num_workers: 100
  queue_size: 2000
```

**5. Cancel stuck execution:**

```bash
# Cancel execution
curl -X POST http://localhost:8080/api/v1/executions/{execution_id}/cancel
```

### Memory Leaks

#### Symptom

Memory usage continuously increases over time.

#### Possible Causes

1. Goroutine leaks
2. Unclosed database connections
3. Large workflow outputs not being cleaned up
4. Cache not being evicted

#### Solutions

**1. Check memory metrics:**

```bash
# View memory usage
curl http://localhost:8080/metrics | grep go_memstats

# Check goroutine count
curl http://localhost:8080/metrics | grep go_goroutines
```

**2. Enable pprof profiling:**

```bash
# Get heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# Analyze with pprof
go tool pprof heap.prof
```

**3. Check database connections:**

```bash
# View connection pool metrics
curl http://localhost:8080/metrics | grep db_connections

# Verify max connections
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

**4. Configure garbage collection:**

```bash
# Adjust GC settings
GOGC=100
GOMEMLIMIT=4GiB
```

**5. Implement execution cleanup:**

```bash
# Clean up old executions
curl -X DELETE http://localhost:8080/api/v1/executions?older_than=7d
```

### Performance Issues

#### Symptom

Slow workflow execution or high API latency.

#### Possible Causes

1. Database queries are slow
2. Too many concurrent executions
3. Insufficient resources
4. Network latency

#### Solutions

**1. Check performance metrics:**

```bash
# View execution duration
curl http://localhost:8080/metrics | grep workflow_execution_duration

# Check API latency
curl http://localhost:8080/metrics | grep http_request_duration
```

**2. Analyze slow queries:**

```sql
-- Enable query logging
ALTER DATABASE goflow SET log_min_duration_statement = 1000;

-- View slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

**3. Add database indexes:**

```sql
-- Add indexes for common queries
CREATE INDEX idx_workflows_created_at ON workflows(created_at);
CREATE INDEX idx_executions_workflow_id ON executions(workflow_id);
CREATE INDEX idx_executions_status ON executions(status);
```

**4. Enable Redis caching:**

```yaml
cache:
  enabled: true
  ttl: 300
  workflows: true
  executions: false
```

**5. Scale horizontally:**

```bash
# Increase replicas
kubectl scale deployment goflow-service --replicas=5 -n goflow
```

## Debugging Guide

### Logs

#### Log Levels

Set appropriate log level based on debugging needs:

```bash
# Development
LOG_LEVEL=debug

# Production
LOG_LEVEL=info

# Troubleshooting
LOG_LEVEL=debug
```

#### Viewing Logs

**Docker:**

```bash
# View all logs
docker logs goflow-service

# Follow logs
docker logs -f goflow-service

# Last 100 lines
docker logs --tail 100 goflow-service

# Filter by level
docker logs goflow-service 2>&1 | grep ERROR
```

**Kubernetes:**

```bash
# View pod logs
kubectl logs -n goflow deployment/goflow-service

# Follow logs
kubectl logs -f -n goflow deployment/goflow-service

# Previous container logs
kubectl logs -n goflow deployment/goflow-service --previous

# Filter by execution ID
kubectl logs -n goflow deployment/goflow-service | grep exec_123
```

#### Structured Logging

GoFlow uses structured JSON logging:

```json
{
  "level": "error",
  "timestamp": "2024-01-01T10:00:00Z",
  "caller": "service/workflow_service.go:123",
  "message": "Failed to execute workflow",
  "workflow_id": "wf_123",
  "execution_id": "exec_456",
  "error": "node execution failed",
  "stack_trace": "..."
}
```

#### Log Aggregation

**Using kubectl:**

```bash
# Aggregate logs from all pods
kubectl logs -n goflow -l app=goflow-service --tail=100

# Export logs to file
kubectl logs -n goflow deployment/goflow-service > goflow.log
```

**Using Elasticsearch:**

```bash
# Query logs by execution ID
curl -X GET "elasticsearch:9200/goflow-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "execution_id": "exec_123"
    }
  }
}'
```

### Distributed Tracing

GoFlow supports OpenTelemetry for distributed tracing.

#### Enable Tracing

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=goflow-service
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=0.1
```

#### View Traces in Jaeger

```bash
# Access Jaeger UI
http://jaeger:16686

# Search by execution ID
# Filter by service: goflow-service
# View trace timeline and spans
```

#### Trace Context

Each workflow execution creates a trace with spans for:

- Workflow execution
- Node executions
- Database queries
- HTTP requests
- Cache operations

#### Example Trace

```
Workflow Execution (exec_123) - 1.2s
├── Node: webhook_trigger - 10ms
├── Node: http_request - 500ms
│   ├── HTTP GET https://api.example.com - 480ms
│   └── Response parsing - 20ms
├── Node: transform - 50ms
└── Node: database - 100ms
    └── SQL INSERT - 95ms
```

### Metrics

#### Prometheus Metrics

Access metrics at `http://localhost:8080/metrics`

**Key Metrics:**

```bash
# Workflow metrics
goflow_workflow_executions_total{status="success"} 1234
goflow_workflow_executions_total{status="failed"} 56
goflow_workflow_execution_duration_seconds_bucket{le="1"} 800
goflow_workflow_execution_duration_seconds_bucket{le="5"} 1200

# Node metrics
goflow_node_executions_total{type="http_request",status="success"} 5678
goflow_node_execution_duration_seconds{type="http_request"} 0.5

# System metrics
goflow_worker_pool_active_workers 45
goflow_worker_pool_queue_size 123
go_goroutines 234
go_memstats_alloc_bytes 123456789

# HTTP metrics
goflow_http_requests_total{method="POST",path="/api/v1/workflows",status="200"} 1000
goflow_http_request_duration_seconds{method="POST",path="/api/v1/workflows"} 0.1
```

#### Querying Metrics

**Using curl:**

```bash
# Get all metrics
curl http://localhost:8080/metrics

# Filter specific metric
curl http://localhost:8080/metrics | grep workflow_executions_total
```

**Using PromQL:**

```promql
# Success rate
rate(goflow_workflow_executions_total{status="success"}[5m]) /
rate(goflow_workflow_executions_total[5m])

# Average execution duration
rate(goflow_workflow_execution_duration_seconds_sum[5m]) /
rate(goflow_workflow_execution_duration_seconds_count[5m])

# 95th percentile latency
histogram_quantile(0.95,
  rate(goflow_workflow_execution_duration_seconds_bucket[5m])
)

# Active workers trend
avg_over_time(goflow_worker_pool_active_workers[1h])
```

## Resiliency Features

### Automatic Retries

GoFlow automatically retries failed operations with exponential backoff.

#### Retry Configuration

```json
{
  "id": "api_call",
  "type": "http_request",
  "config": {
    "url": "https://api.example.com",
    "retry": {
      "max_attempts": 5,
      "initial_delay": "1s",
      "max_delay": "30s",
      "multiplier": 2,
      "retryable_status_codes": [408, 429, 500, 502, 503, 504]
    }
  }
}
```

#### Retry Behavior

```
Attempt 1: Immediate
Attempt 2: Wait 1s
Attempt 3: Wait 2s
Attempt 4: Wait 4s
Attempt 5: Wait 8s
```

#### Monitoring Retries

```bash
# View retry metrics
curl http://localhost:8080/metrics | grep retry

goflow_node_retries_total{type="http_request"} 123
goflow_node_retry_exhausted_total{type="http_request"} 5
```

### State Restoration

GoFlow persists execution state to enable recovery after failures.

#### State Persistence

- Workflow execution state saved to PostgreSQL
- Node outputs cached in Redis
- Write-Ahead Log (WAL) for durability

#### Recovery Process

```go
// Automatic recovery on startup
func (s *WorkflowService) RecoverExecutions(ctx context.Context) error {
    // Find incomplete executions
    executions, err := s.repo.FindIncompleteExecutions(ctx)
    if err != nil {
        return err
    }

    // Resume each execution
    for _, exec := range executions {
        go s.ResumeExecution(ctx, exec.ID)
    }

    return nil
}
```

#### Manual Recovery

```bash
# List incomplete executions
curl http://localhost:8080/api/v1/executions?status=running

# Resume execution
curl -X POST http://localhost:8080/api/v1/executions/{execution_id}/resume
```

### Circuit Breaker

Circuit breaker prevents cascading failures by stopping requests to failing services.

#### Circuit Breaker States

1. **Closed** - Normal operation, requests pass through
2. **Open** - Too many failures, requests fail immediately
3. **Half-Open** - Testing if service recovered

#### Configuration

```yaml
circuit_breaker:
  failure_threshold: 5
  success_threshold: 2
  timeout: 60s
  half_open_max_requests: 3
```

#### Monitoring Circuit Breaker

```bash
# View circuit breaker state
curl http://localhost:8080/metrics | grep circuit_breaker

goflow_circuit_breaker_state{service="external_api"} 0  # 0=closed, 1=open, 2=half-open
goflow_circuit_breaker_failures_total{service="external_api"} 123
```

## Performance Troubleshooting

### Profiling

#### CPU Profiling

```bash
# Capture CPU profile for 30 seconds
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze profile
go tool pprof cpu.prof

# Top functions by CPU time
(pprof) top10

# View call graph
(pprof) web
```

#### Memory Profiling

```bash
# Capture heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# Analyze profile
go tool pprof heap.prof

# Top allocations
(pprof) top10

# View allocation sources
(pprof) list FunctionName
```

#### Goroutine Profiling

```bash
# View goroutine dump
curl http://localhost:8080/debug/pprof/goroutine?debug=2

# Analyze goroutine profile
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

### Load Testing

#### Using Apache Bench

```bash
# Test workflow creation
ab -n 1000 -c 10 -T 'application/json' \
  -p workflow.json \
  http://localhost:8080/api/v1/workflows

# Test workflow execution
ab -n 1000 -c 10 -X POST \
  http://localhost:8080/api/v1/workflows/{workflow_id}/execute
```

#### Using k6

```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  vus: 50,
  duration: '5m',
};

export default function() {
  let res = http.post('http://localhost:8080/api/v1/workflows/{workflow_id}/execute',
    JSON.stringify({ input: { test: 'data' } }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(res, {
    'status is 200': (r) => r.status === 200,
    'duration < 1s': (r) => r.timings.duration < 1000,
  });
}
```

## Escalation Procedures

### Severity Levels

#### P0 - Critical

- Service is completely down
- Data loss or corruption
- Security breach

**Response Time:** Immediate
**Action:** Page on-call engineer, create incident

#### P1 - High

- Major feature not working
- Significant performance degradation
- Affecting multiple users

**Response Time:** 1 hour
**Action:** Notify team, create ticket

#### P2 - Medium

- Minor feature not working
- Workaround available
- Affecting few users

**Response Time:** 4 hours
**Action:** Create ticket, schedule fix

#### P3 - Low

- Cosmetic issues
- Feature requests
- Documentation updates

**Response Time:** Next sprint
**Action:** Add to backlog

### Incident Response

#### 1. Identify

```bash
# Check service health
curl http://localhost:8080/health

# Check metrics
curl http://localhost:8080/metrics

# Check logs
kubectl logs -n goflow deployment/goflow-service --tail=100
```

#### 2. Assess

- Determine severity level
- Identify affected users/workflows
- Estimate impact

#### 3. Mitigate

```bash
# Rollback deployment
kubectl rollout undo deployment/goflow-service -n goflow

# Scale up resources
kubectl scale deployment goflow-service --replicas=10 -n goflow

# Disable problematic feature
kubectl set env deployment/goflow-service FEATURE_FLAG_X=false -n goflow
```

#### 4. Communicate

- Update status page
- Notify affected users
- Post in incident channel

#### 5. Resolve

- Deploy fix
- Verify resolution
- Monitor for recurrence

#### 6. Post-Mortem

- Document root cause
- Identify action items
- Update runbooks

### Contact Information

- **On-Call Engineer:** Check PagerDuty
- **Team Slack:** #goflow-team
- **Email:** goflow-support@example.com
- **Documentation:** https://docs.goflow.example.com

---

For more information, see:
- [Architecture Documentation](../architecture.md)
- [Deployment Guide](./deployment.md)
- [Monitoring Setup](../../monitoring/README.md)
