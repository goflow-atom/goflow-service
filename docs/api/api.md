# API Documentation

This document describes the RESTful API provided by the GoFlow Workflow Engine.

## Table of Contents

- [Overview](#overview)
- [Base URL](#base-url)
- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [Workflows](#workflows)
  - [Executions](#executions)
  - [Webhooks](#webhooks)
  - [Schedules](#schedules)
- [Request/Response Format](#requestresponse-format)
- [Status Codes](#status-codes)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Pagination](#pagination)

## Overview

The GoFlow API is a RESTful API that allows you to:
- Create, read, update, and delete workflows
- Trigger and monitor workflow executions
- Register and manage webhooks
- Schedule recurring workflows
- Query execution logs and metrics

All API endpoints return JSON responses and accept JSON request bodies.

## Base URL

```
Development: http://localhost:8080/api/v1
Production:  https://api.goflow.example.com/api/v1
```

## Authentication

All API requests require authentication using JWT (JSON Web Tokens).

### Obtaining a Token

```bash
POST /auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "your-password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9...",
  "expires_at": "2024-01-01T12:00:00Z"
}
```

### Using the Token

Include the token in the `Authorization` header:

```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9...
```

## Endpoints

### Workflows

#### List Workflows

```http
GET /workflows
```

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 20, max: 100)
- `sort` (string): Sort field (default: created_at)
- `order` (string): Sort order - asc or desc (default: desc)
- `search` (string): Search by name or description

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/workflows?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "wf_123abc",
      "name": "User Onboarding",
      "description": "Automated user onboarding workflow",
      "version": 1,
      "status": "active",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

#### Get Workflow

```http
GET /workflows/:id
```

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/workflows/wf_123abc" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "id": "wf_123abc",
  "name": "User Onboarding",
  "description": "Automated user onboarding workflow",
  "version": 1,
  "status": "active",
  "nodes": [
    {
      "id": "node_1",
      "type": "webhook",
      "config": {
        "method": "POST",
        "path": "/webhooks/user-signup"
      }
    },
    {
      "id": "node_2",
      "type": "http_request",
      "config": {
        "url": "https://api.example.com/users",
        "method": "POST",
        "body": {
          "email": "${input.email}",
          "name": "${input.name}"
        }
      }
    }
  ],
  "edges": [
    {
      "from": "node_1",
      "to": "node_2"
    }
  ],
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

#### Create Workflow

```http
POST /workflows
```

**Request Body:**
```json
{
  "name": "User Onboarding",
  "description": "Automated user onboarding workflow",
  "nodes": [
    {
      "id": "node_1",
      "type": "webhook",
      "config": {
        "method": "POST",
        "path": "/webhooks/user-signup"
      }
    },
    {
      "id": "node_2",
      "type": "http_request",
      "config": {
        "url": "https://api.example.com/users",
        "method": "POST",
        "body": {
          "email": "${input.email}",
          "name": "${input.name}"
        }
      }
    }
  ],
  "edges": [
    {
      "from": "node_1",
      "to": "node_2"
    }
  ]
}
```

**Response:**
```json
{
  "id": "wf_123abc",
  "name": "User Onboarding",
  "description": "Automated user onboarding workflow",
  "version": 1,
  "status": "active",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

#### Update Workflow

```http
PUT /workflows/:id
```

**Request Body:**
```json
{
  "name": "Updated User Onboarding",
  "description": "Updated description",
  "nodes": [...],
  "edges": [...]
}
```

**Response:**
```json
{
  "id": "wf_123abc",
  "name": "Updated User Onboarding",
  "version": 2,
  "status": "active",
  "updated_at": "2024-01-02T10:00:00Z"
}
```

#### Delete Workflow

```http
DELETE /workflows/:id
```

**Response:**
```json
{
  "message": "Workflow deleted successfully"
}
```

### Executions

#### List Executions

```http
GET /executions
```

**Query Parameters:**
- `workflow_id` (string): Filter by workflow ID
- `status` (string): Filter by status (pending, running, completed, failed)
- `page` (integer): Page number
- `limit` (integer): Items per page

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/executions?workflow_id=wf_123abc&status=completed" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "exec_456def",
      "workflow_id": "wf_123abc",
      "status": "completed",
      "started_at": "2024-01-01T11:00:00Z",
      "completed_at": "2024-01-01T11:00:05Z",
      "duration_ms": 5000
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

#### Get Execution

```http
GET /executions/:id
```

**Response:**
```json
{
  "id": "exec_456def",
  "workflow_id": "wf_123abc",
  "status": "completed",
  "input": {
    "email": "user@example.com",
    "name": "John Doe"
  },
  "output": {
    "user_id": "usr_789",
    "status": "created"
  },
  "node_executions": [
    {
      "node_id": "node_1",
      "status": "completed",
      "started_at": "2024-01-01T11:00:00Z",
      "completed_at": "2024-01-01T11:00:02Z"
    },
    {
      "node_id": "node_2",
      "status": "completed",
      "started_at": "2024-01-01T11:00:02Z",
      "completed_at": "2024-01-01T11:00:05Z"
    }
  ],
  "started_at": "2024-01-01T11:00:00Z",
  "completed_at": "2024-01-01T11:00:05Z"
}
```

#### Trigger Execution

```http
POST /workflows/:id/execute
```

**Request Body:**
```json
{
  "input": {
    "email": "user@example.com",
    "name": "John Doe"
  },
  "async": true
}
```

**Response (async=true):**
```json
{
  "execution_id": "exec_456def",
  "status": "pending",
  "message": "Execution started"
}
```

**Response (async=false):**
```json
{
  "execution_id": "exec_456def",
  "status": "completed",
  "output": {
    "user_id": "usr_789",
    "status": "created"
  }
}
```

#### Cancel Execution

```http
POST /executions/:id/cancel
```

**Response:**
```json
{
  "message": "Execution cancelled successfully"
}
```

#### Get Execution Logs

```http
GET /executions/:id/logs
```

**Query Parameters:**
- `level` (string): Filter by log level (debug, info, warn, error)
- `node_id` (string): Filter by node ID

**Response:**
```json
{
  "logs": [
    {
      "timestamp": "2024-01-01T11:00:00Z",
      "level": "info",
      "node_id": "node_1",
      "message": "Node execution started",
      "metadata": {}
    },
    {
      "timestamp": "2024-01-01T11:00:02Z",
      "level": "info",
      "node_id": "node_1",
      "message": "Node execution completed",
      "metadata": {
        "duration_ms": 2000
      }
    }
  ]
}
```

### Webhooks

#### List Webhooks

```http
GET /webhooks
```

**Response:**
```json
{
  "data": [
    {
      "id": "wh_123",
      "workflow_id": "wf_123abc",
      "path": "/webhooks/user-signup",
      "method": "POST",
      "enabled": true,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

#### Get Webhook

```http
GET /webhooks/:id
```

**Response:**
```json
{
  "id": "wh_123",
  "workflow_id": "wf_123abc",
  "path": "/webhooks/user-signup",
  "method": "POST",
  "enabled": true,
  "secret": "whsec_...",
  "created_at": "2024-01-01T10:00:00Z"
}
```

#### Create Webhook

```http
POST /webhooks
```

**Request Body:**
```json
{
  "workflow_id": "wf_123abc",
  "path": "/webhooks/user-signup",
  "method": "POST",
  "enabled": true
}
```

**Response:**
```json
{
  "id": "wh_123",
  "workflow_id": "wf_123abc",
  "path": "/webhooks/user-signup",
  "method": "POST",
  "enabled": true,
  "secret": "whsec_abc123...",
  "created_at": "2024-01-01T10:00:00Z"
}
```

#### Delete Webhook

```http
DELETE /webhooks/:id
```

**Response:**
```json
{
  "message": "Webhook deleted successfully"
}
```

### Schedules

#### List Schedules

```http
GET /schedules
```

**Response:**
```json
{
  "data": [
    {
      "id": "sch_123",
      "workflow_id": "wf_123abc",
      "cron": "0 0 * * *",
      "enabled": true,
      "next_run": "2024-01-02T00:00:00Z",
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

#### Create Schedule

```http
POST /schedules
```

**Request Body:**
```json
{
  "workflow_id": "wf_123abc",
  "cron": "0 0 * * *",
  "timezone": "UTC",
  "enabled": true,
  "input": {
    "param1": "value1"
  }
}
```

**Response:**
```json
{
  "id": "sch_123",
  "workflow_id": "wf_123abc",
  "cron": "0 0 * * *",
  "timezone": "UTC",
  "enabled": true,
  "next_run": "2024-01-02T00:00:00Z",
  "created_at": "2024-01-01T10:00:00Z"
}
```

#### Update Schedule

```http
PUT /schedules/:id
```

**Request Body:**
```json
{
  "cron": "0 12 * * *",
  "enabled": false
}
```

#### Delete Schedule

```http
DELETE /schedules/:id
```

## Request/Response Format

### Content Type

All requests and responses use JSON:
```
Content-Type: application/json
```

### Date Format

All timestamps use ISO 8601 format:
```
2024-01-01T10:00:00Z
```

### ID Format

Resource IDs use prefixed format:
- Workflows: `wf_` prefix (e.g., `wf_123abc`)
- Executions: `exec_` prefix (e.g., `exec_456def`)
- Webhooks: `wh_` prefix (e.g., `wh_789ghi`)
- Schedules: `sch_` prefix (e.g., `sch_012jkl`)

## Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created successfully |
| 204 | No Content - Request succeeded with no response body |
| 400 | Bad Request - Invalid request format or parameters |
| 401 | Unauthorized - Missing or invalid authentication |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists or conflict |
| 422 | Unprocessable Entity - Validation error |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |
| 503 | Service Unavailable - Service temporarily unavailable |

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid workflow definition",
    "details": [
      {
        "field": "nodes[0].config.url",
        "message": "URL is required"
      }
    ],
    "request_id": "req_123abc"
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `AUTHENTICATION_ERROR` | Authentication failed |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource conflict |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `INTERNAL_ERROR` | Internal server error |
| `WORKFLOW_EXECUTION_FAILED` | Workflow execution failed |
| `INVALID_WORKFLOW` | Invalid workflow definition |

## Rate Limiting

API requests are rate limited to prevent abuse.

### Rate Limit Headers

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

### Rate Limits

| Tier | Requests per Hour |
|------|-------------------|
| Free | 1,000 |
| Pro | 10,000 |
| Enterprise | 100,000 |

### Rate Limit Exceeded Response

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "retry_after": 3600
  }
}
```

## Pagination

List endpoints support pagination using query parameters.

### Pagination Parameters

- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

### Pagination Response

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

## Examples

### Complete Workflow Creation and Execution

```bash
# 1. Create workflow
curl -X POST "http://localhost:8080/api/v1/workflows" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "User Onboarding",
    "nodes": [
      {
        "id": "webhook",
        "type": "webhook",
        "config": {"method": "POST", "path": "/webhooks/signup"}
      },
      {
        "id": "create_user",
        "type": "http_request",
        "config": {
          "url": "https://api.example.com/users",
          "method": "POST",
          "body": {"email": "${input.email}"}
        }
      }
    ],
    "edges": [{"from": "webhook", "to": "create_user"}]
  }'

# 2. Trigger execution
curl -X POST "http://localhost:8080/api/v1/workflows/wf_123abc/execute" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "input": {"email": "user@example.com"},
    "async": true
  }'

# 3. Check execution status
curl -X GET "http://localhost:8080/api/v1/executions/exec_456def" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

For more information, see:
- [Authentication Documentation](./authentication.md)
- [Webhook Documentation](./webhooks.md)
- [OpenAPI Specification](../../api/openapi.yaml)

