# Webhooks

This document describes webhook functionality in GoFlow, including configuration, security, and use cases.

## Table of Contents

- [Overview](#overview)
- [Webhook Triggers](#webhook-triggers)
- [Webhook Security](#webhook-security)
- [Configuration](#configuration)
- [Payload Format](#payload-format)
- [Retry Behavior](#retry-behavior)
- [Use Cases](#use-cases)

## Overview

Webhooks allow external services to trigger workflow executions by sending HTTP requests to GoFlow. When a webhook receives a request, it automatically starts the associated workflow with the request data as input.

## Webhook Triggers

### Creating a Webhook

Webhooks are created as part of workflow definitions or via the API:

#### Via Workflow Definition

```json
{
  "name": "GitHub Integration",
  "nodes": [
    {
      "id": "webhook_trigger",
      "type": "webhook",
      "config": {
        "method": "POST",
        "path": "/webhooks/github-push",
        "validate_signature": true
      }
    },
    {
      "id": "process_push",
      "type": "transform",
      "config": {
        "script": "return { repo: input.repository.name, branch: input.ref }"
      }
    }
  ],
  "edges": [
    {
      "from": "webhook_trigger",
      "to": "process_push"
    }
  ]
}
```

#### Via API

```http
POST /api/v1/webhooks
Authorization: Bearer <token>
Content-Type: application/json

{
  "workflow_id": "wf_123abc",
  "path": "/webhooks/github-push",
  "method": "POST",
  "enabled": true,
  "validate_signature": true
}
```

**Response:**
```json
{
  "id": "wh_123",
  "workflow_id": "wf_123abc",
  "url": "https://api.goflow.example.com/webhooks/github-push",
  "method": "POST",
  "enabled": true,
  "secret": "whsec_abc123def456...",
  "created_at": "2024-01-01T10:00:00Z"
}
```

### Triggering a Webhook

Send an HTTP request to the webhook URL:

```bash
curl -X POST "https://api.goflow.example.com/webhooks/github-push" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Signature: sha256=..." \
  -d '{
    "repository": {
      "name": "my-repo",
      "url": "https://github.com/user/my-repo"
    },
    "ref": "refs/heads/main",
    "commits": [...]
  }'
```

**Response:**
```json
{
  "execution_id": "exec_456def",
  "status": "pending",
  "message": "Workflow execution started"
}
```

## Webhook Security

### Signature Validation

GoFlow uses HMAC-SHA256 signatures to verify webhook authenticity.

#### Generating Signatures

When creating a webhook, GoFlow provides a secret key. Use this key to sign requests:

```python
import hmac
import hashlib
import json

def generate_signature(payload, secret):
    """Generate HMAC-SHA256 signature for webhook payload"""
    payload_bytes = json.dumps(payload).encode('utf-8')
    signature = hmac.new(
        secret.encode('utf-8'),
        payload_bytes,
        hashlib.sha256
    ).hexdigest()
    return f"sha256={signature}"

# Example usage
payload = {"event": "user.created", "data": {...}}
secret = "whsec_abc123def456..."
signature = generate_signature(payload, secret)

# Include in request header
headers = {
    "X-Webhook-Signature": signature,
    "Content-Type": "application/json"
}
```

#### Verifying Signatures (Server-Side)

GoFlow automatically verifies signatures when `validate_signature` is enabled:

```go
func (h *WebhookHandler) VerifySignature(payload []byte, signature string, secret string) bool {
    expectedSignature := fmt.Sprintf("sha256=%s", 
        hmac.New(sha256.New, []byte(secret)).Sum(payload))
    
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

### Secret Rotation

Rotate webhook secrets regularly for security:

```http
POST /api/v1/webhooks/:id/rotate-secret
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": "wh_123",
  "secret": "whsec_new_secret_789...",
  "rotated_at": "2024-06-01T10:00:00Z"
}
```

### IP Whitelisting

Restrict webhook access to specific IP addresses:

```http
PUT /api/v1/webhooks/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "allowed_ips": [
    "192.30.252.0/22",
    "185.199.108.0/22"
  ]
}
```

### Rate Limiting

Webhooks are rate limited per IP address:
- 100 requests per minute per IP
- 1000 requests per hour per IP

## Configuration

### Webhook Configuration Options

```json
{
  "id": "wh_123",
  "workflow_id": "wf_123abc",
  "path": "/webhooks/custom-path",
  "method": "POST",
  "enabled": true,
  "validate_signature": true,
  "allowed_ips": ["192.168.1.0/24"],
  "timeout": 30,
  "retry_config": {
    "max_attempts": 3,
    "initial_delay": 1000,
    "max_delay": 10000,
    "multiplier": 2.0
  },
  "headers": {
    "X-Custom-Header": "value"
  },
  "query_params": {
    "source": "webhook"
  }
}
```

**Configuration Fields:**
- `path`: Webhook URL path (must be unique)
- `method`: HTTP method (GET, POST, PUT, DELETE)
- `enabled`: Enable/disable webhook
- `validate_signature`: Require signature validation
- `allowed_ips`: IP whitelist (CIDR notation)
- `timeout`: Request timeout in seconds
- `retry_config`: Retry configuration for failed executions
- `headers`: Custom headers to include in workflow input
- `query_params`: Custom query parameters

### Response Configuration

Configure webhook response behavior:

```json
{
  "response_config": {
    "sync": false,
    "include_execution_id": true,
    "success_status": 202,
    "success_body": {
      "message": "Workflow triggered successfully"
    }
  }
}
```

**Synchronous Response (sync=true):**
```json
{
  "execution_id": "exec_456def",
  "status": "completed",
  "output": {
    "result": "success"
  },
  "duration_ms": 1234
}
```

**Asynchronous Response (sync=false):**
```json
{
  "execution_id": "exec_456def",
  "status": "pending",
  "message": "Workflow execution started"
}
```

## Payload Format

### Request Payload

Webhook requests can include any JSON payload:

```json
{
  "event": "user.created",
  "timestamp": "2024-01-01T10:00:00Z",
  "data": {
    "user_id": "usr_123",
    "email": "user@example.com",
    "name": "John Doe"
  },
  "metadata": {
    "source": "api",
    "version": "1.0"
  }
}
```

### Workflow Input

The webhook payload becomes the workflow input:

```json
{
  "input": {
    "event": "user.created",
    "timestamp": "2024-01-01T10:00:00Z",
    "data": {
      "user_id": "usr_123",
      "email": "user@example.com",
      "name": "John Doe"
    },
    "metadata": {
      "source": "api",
      "version": "1.0"
    },
    "_webhook": {
      "id": "wh_123",
      "path": "/webhooks/user-created",
      "method": "POST",
      "headers": {
        "content-type": "application/json",
        "user-agent": "GitHub-Hookshot/abc123"
      },
      "query": {},
      "ip": "192.30.252.1"
    }
  }
}
```

### Accessing Webhook Data

Access webhook data in node configurations using expressions:

```json
{
  "id": "send_email",
  "type": "email",
  "config": {
    "to": "${input.data.email}",
    "subject": "Welcome ${input.data.name}!",
    "body": "Your account was created at ${input.timestamp}"
  }
}
```

## Retry Behavior

### Automatic Retries

If workflow execution fails, GoFlow automatically retries based on configuration:

```json
{
  "retry_config": {
    "max_attempts": 3,
    "initial_delay": 1000,
    "max_delay": 10000,
    "multiplier": 2.0
  }
}
```

**Retry Schedule:**
- Attempt 1: Immediate
- Attempt 2: After 1 second
- Attempt 3: After 2 seconds
- Attempt 4: After 4 seconds (capped at max_delay)

### Retry Response

During retries, webhook returns:

```json
{
  "execution_id": "exec_456def",
  "status": "retrying",
  "attempt": 2,
  "max_attempts": 3,
  "next_retry_at": "2024-01-01T10:00:05Z"
}
```

### Failed Webhooks

After exhausting retries:

```json
{
  "execution_id": "exec_456def",
  "status": "failed",
  "error": {
    "code": "EXECUTION_FAILED",
    "message": "Workflow execution failed after 3 attempts"
  }
}
```

## Use Cases

### 1. GitHub Integration

Trigger workflows on GitHub events:

```json
{
  "name": "CI/CD Pipeline",
  "nodes": [
    {
      "id": "github_webhook",
      "type": "webhook",
      "config": {
        "path": "/webhooks/github-push",
        "method": "POST",
        "validate_signature": true
      }
    },
    {
      "id": "run_tests",
      "type": "http_request",
      "config": {
        "url": "https://ci.example.com/run-tests",
        "method": "POST",
        "body": {
          "repo": "${input.repository.name}",
          "branch": "${input.ref}"
        }
      }
    }
  ]
}
```

### 2. Payment Gateway Integration

Process payment events from Stripe:

```json
{
  "name": "Payment Processing",
  "nodes": [
    {
      "id": "stripe_webhook",
      "type": "webhook",
      "config": {
        "path": "/webhooks/stripe-payment",
        "method": "POST",
        "validate_signature": true
      }
    },
    {
      "id": "update_order",
      "type": "database",
      "config": {
        "query": "UPDATE orders SET status = 'paid' WHERE id = $1",
        "params": ["${input.data.object.metadata.order_id}"]
      }
    },
    {
      "id": "send_confirmation",
      "type": "email",
      "config": {
        "to": "${input.data.object.billing_details.email}",
        "subject": "Payment Confirmed",
        "template": "payment_confirmation"
      }
    }
  ]
}
```

### 3. Form Submission

Handle form submissions from websites:

```json
{
  "name": "Contact Form Handler",
  "nodes": [
    {
      "id": "form_webhook",
      "type": "webhook",
      "config": {
        "path": "/webhooks/contact-form",
        "method": "POST"
      }
    },
    {
      "id": "validate_input",
      "type": "conditional",
      "config": {
        "condition": "${input.email != null && input.message != null}",
        "on_true": "save_to_db",
        "on_false": "send_error"
      }
    },
    {
      "id": "save_to_db",
      "type": "database",
      "config": {
        "query": "INSERT INTO contacts (email, message) VALUES ($1, $2)",
        "params": ["${input.email}", "${input.message}"]
      }
    }
  ]
}
```

---

For more information, see:
- [API Documentation](./api.md)
- [Workflow Definition Guide](../guides/workflow-definition.md)
- [Node Types Documentation](../guides/node-types.md)

