# Deployment Guide

This guide covers deploying GoFlow to production environments.

## Table of Contents

- [Overview](#overview)
- [Deployment Methods](#deployment-methods)
  - [Docker Compose](#docker-compose)
  - [Kubernetes](#kubernetes)
  - [CI/CD Pipeline](#cicd-pipeline)
- [Environment Configuration](#environment-configuration)
- [Database Setup](#database-setup)
- [Monitoring and Logging](#monitoring-and-logging)
- [Security Best Practices](#security-best-practices)
- [Scaling Considerations](#scaling-considerations)

## Overview

GoFlow can be deployed using various methods depending on your infrastructure requirements. This guide covers the most common deployment scenarios.

### Prerequisites

- Docker 20.10+
- Kubernetes 1.24+ (for Kubernetes deployment)
- PostgreSQL 14+
- Redis 6+
- Kafka 3.0+ (optional, for event streaming)

## Deployment Methods

### Docker Compose

Docker Compose is suitable for development, staging, and small production deployments.

#### Step 1: Prepare Environment

Create a `.env` file:

```bash
# Application
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info

# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=goflow
DB_USER=goflow
DB_PASSWORD=your_secure_password
DB_SSL_MODE=require

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=workflow-events

# Inngest
INNGEST_EVENT_KEY=your_inngest_event_key
INNGEST_SIGNING_KEY=your_inngest_signing_key

# OpenAI
OPENAI_API_KEY=your_openai_api_key

# JWT
JWT_SECRET=your_jwt_secret_key_min_32_chars
JWT_EXPIRATION=24h

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=your_smtp_password
SMTP_FROM=noreply@example.com
```

#### Step 2: Deploy with Docker Compose

```bash
# Pull latest images
docker-compose pull

# Start services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f goflow-service
```

#### Step 3: Run Database Migrations

```bash
docker-compose exec goflow-service ./goflow-service migrate up
```

#### Step 4: Verify Deployment

```bash
# Health check
curl http://localhost:8080/health

# API check
curl http://localhost:8080/api/v1/workflows
```

### Kubernetes

Kubernetes is recommended for production deployments requiring high availability and scalability.

#### Step 1: Create Namespace

```bash
kubectl create namespace goflow
```

#### Step 2: Create Secrets

```bash
# Create database secret
kubectl create secret generic goflow-db-secret \
  --from-literal=username=goflow \
  --from-literal=password=your_secure_password \
  -n goflow

# Create JWT secret
kubectl create secret generic goflow-jwt-secret \
  --from-literal=secret=your_jwt_secret_key_min_32_chars \
  -n goflow

# Create OpenAI secret
kubectl create secret generic goflow-openai-secret \
  --from-literal=api-key=your_openai_api_key \
  -n goflow

# Create Inngest secret
kubectl create secret generic goflow-inngest-secret \
  --from-literal=event-key=your_inngest_event_key \
  --from-literal=signing-key=your_inngest_signing_key \
  -n goflow
```

#### Step 3: Create ConfigMap

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: goflow-config
  namespace: goflow
data:
  APP_ENV: "production"
  APP_PORT: "8080"
  LOG_LEVEL: "info"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "goflow"
  DB_SSL_MODE: "require"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  KAFKA_BROKERS: "kafka-service:9092"
  KAFKA_TOPIC: "workflow-events"
EOF
```

#### Step 4: Deploy Application

```bash
# Apply Kubernetes manifests
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl get deployments -n goflow
kubectl get pods -n goflow
kubectl get services -n goflow

# View logs
kubectl logs -f deployment/goflow-service -n goflow
```

#### Step 5: Run Database Migrations

```bash
# Create migration job
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: goflow-migrate
  namespace: goflow
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: goflow/goflow-service:latest
        command: ["./goflow-service", "migrate", "up"]
        envFrom:
        - configMapRef:
            name: goflow-config
        - secretRef:
            name: goflow-db-secret
      restartPolicy: OnFailure
EOF

# Check migration status
kubectl logs job/goflow-migrate -n goflow
```

#### Step 6: Configure Ingress

```bash
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: goflow-ingress
  namespace: goflow
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - goflow.example.com
    secretName: goflow-tls
  rules:
  - host: goflow.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: goflow-service
            port:
              number: 8080
EOF
```

### CI/CD Pipeline

Automate deployments using GitHub Actions.

#### GitHub Actions Workflow

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Production

on:
  push:
    branches:
      - main
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: make test

      - name: Build binary
        run: make build

      - name: Build Docker image
        run: |
          docker build -t goflow/goflow-service:${{ github.sha }} .
          docker tag goflow/goflow-service:${{ github.sha }} goflow/goflow-service:latest

      - name: Push to registry
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push goflow/goflow-service:${{ github.sha }}
          docker push goflow/goflow-service:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG }}

      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/goflow-service \
            goflow-service=goflow/goflow-service:${{ github.sha }} \
            -n goflow
          kubectl rollout status deployment/goflow-service -n goflow

      - name: Run migrations
        run: |
          kubectl apply -f deployments/kubernetes/migration-job.yaml
```

## Environment Configuration

### Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `APP_ENV` | Yes | - | Environment (development, staging, production) |
| `APP_PORT` | No | 8080 | HTTP server port |
| `LOG_LEVEL` | No | info | Log level (debug, info, warn, error) |
| `DB_HOST` | Yes | - | PostgreSQL host |
| `DB_PORT` | No | 5432 | PostgreSQL port |
| `DB_NAME` | Yes | - | Database name |
| `DB_USER` | Yes | - | Database user |
| `DB_PASSWORD` | Yes | - | Database password |
| `DB_SSL_MODE` | No | disable | SSL mode (disable, require, verify-full) |
| `DB_MAX_OPEN_CONNS` | No | 25 | Maximum open connections |
| `DB_MAX_IDLE_CONNS` | No | 5 | Maximum idle connections |
| `REDIS_HOST` | Yes | - | Redis host |
| `REDIS_PORT` | No | 6379 | Redis port |
| `REDIS_PASSWORD` | No | - | Redis password |
| `REDIS_DB` | No | 0 | Redis database number |
| `KAFKA_BROKERS` | No | - | Kafka broker addresses (comma-separated) |
| `KAFKA_TOPIC` | No | workflow-events | Kafka topic name |
| `INNGEST_EVENT_KEY` | Yes | - | Inngest event key |
| `INNGEST_SIGNING_KEY` | Yes | - | Inngest signing key |
| `OPENAI_API_KEY` | No | - | OpenAI API key |
| `JWT_SECRET` | Yes | - | JWT signing secret (min 32 chars) |
| `JWT_EXPIRATION` | No | 24h | JWT token expiration |
| `SMTP_HOST` | No | - | SMTP server host |
| `SMTP_PORT` | No | 587 | SMTP server port |
| `SMTP_USER` | No | - | SMTP username |
| `SMTP_PASSWORD` | No | - | SMTP password |
| `SMTP_FROM` | No | - | Default sender email |

### Configuration Files

GoFlow uses Viper for configuration management. Configuration can be provided via:

1. **Environment variables** (highest priority)
2. **Configuration files** (`configs/config.yaml`)
3. **Default values** (lowest priority)

Example `configs/config.production.yaml`:

```yaml
app:
  env: production
  port: 8080
  log_level: info

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  name: ${DB_NAME}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  ssl_mode: require
  max_open_conns: 50
  max_idle_conns: 10

redis:
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
  password: ${REDIS_PASSWORD}
  db: 0

kafka:
  brokers:
    - ${KAFKA_BROKERS}
  topic: workflow-events

inngest:
  event_key: ${INNGEST_EVENT_KEY}
  signing_key: ${INNGEST_SIGNING_KEY}

openai:
  api_key: ${OPENAI_API_KEY}

jwt:
  secret: ${JWT_SECRET}
  expiration: 24h

smtp:
  host: ${SMTP_HOST}
  port: ${SMTP_PORT}
  user: ${SMTP_USER}
  password: ${SMTP_PASSWORD}
  from: ${SMTP_FROM}
```

## Database Setup

### Running Migrations

GoFlow uses database migrations to manage schema changes.

#### Using CLI

```bash
# Run all pending migrations
./goflow-service migrate up

# Rollback last migration
./goflow-service migrate down

# Check migration status
./goflow-service migrate status

# Create new migration
./goflow-service migrate create add_user_table
```

#### Using Docker

```bash
docker run --rm \
  -e DB_HOST=postgres \
  -e DB_NAME=goflow \
  -e DB_USER=goflow \
  -e DB_PASSWORD=password \
  goflow/goflow-service:latest \
  migrate up
```

#### Using Kubernetes Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: goflow-migrate
  namespace: goflow
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: goflow/goflow-service:latest
        command: ["./goflow-service", "migrate", "up"]
        envFrom:
        - configMapRef:
            name: goflow-config
        - secretRef:
            name: goflow-db-secret
      restartPolicy: OnFailure
  backoffLimit: 3
```

### Database Backup

#### PostgreSQL Backup

```bash
# Create backup
pg_dump -h localhost -U goflow -d goflow > backup.sql

# Restore backup
psql -h localhost -U goflow -d goflow < backup.sql
```

#### Automated Backups with CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: goflow-db-backup
  namespace: goflow
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:14
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | \
              gzip > /backups/goflow-$(date +%Y%m%d-%H%M%S).sql.gz
            envFrom:
            - secretRef:
                name: goflow-db-secret
            volumeMounts:
            - name: backups
              mountPath: /backups
          volumes:
          - name: backups
            persistentVolumeClaim:
              claimName: goflow-backups
          restartPolicy: OnFailure
```

## Monitoring and Logging

### Prometheus Metrics

GoFlow exposes Prometheus metrics at `/metrics` endpoint.

#### Key Metrics

- `goflow_workflow_executions_total` - Total workflow executions
- `goflow_workflow_execution_duration_seconds` - Execution duration histogram
- `goflow_workflow_execution_errors_total` - Total execution errors
- `goflow_node_executions_total` - Total node executions by type
- `goflow_node_execution_duration_seconds` - Node execution duration
- `goflow_worker_pool_active_workers` - Active workers in pool
- `goflow_worker_pool_queue_size` - Current queue size
- `goflow_http_requests_total` - Total HTTP requests
- `goflow_http_request_duration_seconds` - HTTP request duration

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'goflow'
    static_configs:
      - targets: ['goflow-service:8080']
    metrics_path: '/metrics'
```

#### ServiceMonitor for Kubernetes

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: goflow-service
  namespace: goflow
spec:
  selector:
    matchLabels:
      app: goflow-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### Grafana Dashboards

Import pre-built dashboards from `monitoring/grafana/dashboards/`:

1. **Workflow Overview** - Execution metrics, success rates, duration
2. **Node Performance** - Node-level metrics by type
3. **System Health** - CPU, memory, goroutines, connections
4. **API Performance** - HTTP request metrics, latency, errors

#### Import Dashboard

```bash
# Using Grafana API
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -d @monitoring/grafana/dashboards/workflow-overview.json
```

### Logging

GoFlow uses structured logging with Zap.

#### Log Levels

- `debug` - Detailed debugging information
- `info` - General informational messages
- `warn` - Warning messages
- `error` - Error messages

#### Log Format

```json
{
  "level": "info",
  "timestamp": "2024-01-01T10:00:00Z",
  "caller": "service/workflow_service.go:123",
  "message": "Workflow executed successfully",
  "workflow_id": "wf_123",
  "execution_id": "exec_456",
  "duration_ms": 1234
}
```

#### Centralized Logging with ELK Stack

```yaml
# filebeat.yml
filebeat.inputs:
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'
  processors:
  - add_kubernetes_metadata:
      host: ${NODE_NAME}
      matchers:
      - logs_path:
          logs_path: "/var/lib/docker/containers/"

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "goflow-%{+yyyy.MM.dd}"
```

### OpenTelemetry Integration

GoFlow supports OpenTelemetry for distributed tracing.

#### Configuration

```yaml
# otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024

exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger]
```

#### Environment Variables

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=goflow-service
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=0.1
```

## Security Best Practices

### 1. Secrets Management

**Never commit secrets to version control.**

#### Using Kubernetes Secrets

```bash
# Create from file
kubectl create secret generic goflow-secrets \
  --from-file=jwt-secret=./jwt-secret.txt \
  --from-file=db-password=./db-password.txt \
  -n goflow

# Create from literal
kubectl create secret generic goflow-secrets \
  --from-literal=jwt-secret=your_secret \
  --from-literal=db-password=your_password \
  -n goflow
```

#### Using External Secrets Operator

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: goflow-secrets
  namespace: goflow
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: goflow-secrets
  data:
  - secretKey: jwt-secret
    remoteRef:
      key: goflow/jwt-secret
  - secretKey: db-password
    remoteRef:
      key: goflow/db-password
```

### 2. Network Security

#### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: goflow-network-policy
  namespace: goflow
spec:
  podSelector:
    matchLabels:
      app: goflow-service
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

### 3. TLS/SSL Configuration

#### Enable SSL for PostgreSQL

```bash
DB_SSL_MODE=verify-full
DB_SSL_CERT=/path/to/client-cert.pem
DB_SSL_KEY=/path/to/client-key.pem
DB_SSL_ROOT_CERT=/path/to/ca-cert.pem
```

#### Enable TLS for Redis

```bash
REDIS_TLS_ENABLED=true
REDIS_TLS_CERT=/path/to/client-cert.pem
REDIS_TLS_KEY=/path/to/client-key.pem
REDIS_TLS_CA=/path/to/ca-cert.pem
```

### 4. RBAC Configuration

#### Kubernetes RBAC

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: goflow-service
  namespace: goflow
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: goflow-role
  namespace: goflow
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: goflow-rolebinding
  namespace: goflow
subjects:
- kind: ServiceAccount
  name: goflow-service
  namespace: goflow
roleRef:
  kind: Role
  name: goflow-role
  apiGroup: rbac.authorization.k8s.io
```

### 5. Security Scanning

#### Container Image Scanning

```yaml
# .github/workflows/security.yml
name: Security Scan

on: [push, pull_request]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build image
        run: docker build -t goflow/goflow-service:${{ github.sha }} .

      - name: Run Trivy scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: goflow/goflow-service:${{ github.sha }}
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
```

## Scaling Considerations

### Horizontal Scaling

#### Kubernetes HPA

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: goflow-service-hpa
  namespace: goflow
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: goflow-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
```

### Database Scaling

#### Read Replicas

```yaml
# PostgreSQL read replica configuration
database:
  primary:
    host: postgres-primary
    port: 5432
  replicas:
    - host: postgres-replica-1
      port: 5432
    - host: postgres-replica-2
      port: 5432
  read_write_split: true
```

#### Connection Pooling

```bash
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=10m
```

### Redis Scaling

#### Redis Cluster

```yaml
redis:
  mode: cluster
  nodes:
    - redis-node-1:6379
    - redis-node-2:6379
    - redis-node-3:6379
    - redis-node-4:6379
    - redis-node-5:6379
    - redis-node-6:6379
```

### Worker Pool Configuration

```yaml
worker_pool:
  num_workers: 50
  queue_size: 1000
  max_retries: 3
  retry_delay: 5s
```

### Performance Tuning

#### Go Runtime Settings

```bash
GOMAXPROCS=8
GOGC=100
GOMEMLIMIT=4GiB
```

#### Resource Limits

```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

---

For more information, see:
- [Architecture Documentation](../architecture.md)
- [Monitoring Setup](../../monitoring/README.md)
- [Security Guide](../../SECURITY.md)
