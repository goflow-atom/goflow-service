# Workflow Definition

This guide explains how to define workflows in GoFlow using JSON/YAML format.

## Table of Contents

- [Workflow Definition](#workflow-definition)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Workflow Structure](#workflow-structure)
    - [Basic Workflow](#basic-workflow)
    - [Complete Workflow Schema](#complete-workflow-schema)
  - [Workflow JSON Schema](#workflow-json-schema)
    - [Top-Level Fields](#top-level-fields)
    - [Node Fields](#node-fields)
    - [Edge Fields](#edge-fields)
  - [Node Parameters](#node-parameters)
    - [Input and Output](#input-and-output)
  - [Expressions and Variables](#expressions-and-variables)
    - [Expression Syntax](#expression-syntax)
    - [Available Context](#available-context)
    - [Examples](#examples)
    - [Workflow Variables](#workflow-variables)
    - [Secrets Management](#secrets-management)
  - [Workflow Versioning](#workflow-versioning)
    - [Version Management](#version-management)
    - [Version Strategy](#version-strategy)
    - [Creating New Version](#creating-new-version)
    - [Publishing Version](#publishing-version)
  - [Complete Examples](#complete-examples)
    - [Example 1: User Onboarding Workflow](#example-1-user-onboarding-workflow)
    - [Example 2: Data Processing Pipeline](#example-2-data-processing-pipeline)
    - [Example 3: AI-Powered Content Generation](#example-3-ai-powered-content-generation)
  - [Best Practices](#best-practices)
    - [1. Use Descriptive Node IDs](#1-use-descriptive-node-ids)
    - [2. Add Descriptions](#2-add-descriptions)
    - [3. Handle Errors Gracefully](#3-handle-errors-gracefully)
    - [4. Use Variables for Configuration](#4-use-variables-for-configuration)
    - [5. Keep Workflows Focused](#5-keep-workflows-focused)

## Overview

Workflows in GoFlow are defined as Directed Acyclic Graphs (DAGs) where:
- **Nodes** represent individual tasks or operations
- **Edges** define the execution order and data flow between nodes
- **Expressions** enable dynamic data transformation

## Workflow Structure

### Basic Workflow

```json
{
  "name": "My Workflow",
  "description": "Description of what this workflow does",
  "version": 1,
  "nodes": [
    {
      "id": "node_1",
      "type": "http_request",
      "config": {...}
    }
  ],
  "edges": [
    {
      "from": "node_1",
      "to": "node_2"
    }
  ],
  "config": {
    "timeout": 300,
    "retry_policy": "exponential"
  }
}
```

### Complete Workflow Schema

```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "version": "integer (default: 1)",
  "tags": ["string"],
  "nodes": [
    {
      "id": "string (required, unique)",
      "type": "string (required)",
      "name": "string (optional)",
      "description": "string (optional)",
      "config": {
        "...": "node-specific configuration"
      },
      "retry": {
        "max_attempts": 3,
        "initial_delay": 1000,
        "max_delay": 10000,
        "multiplier": 2.0
      },
      "timeout": 30,
      "on_error": "fail|continue|retry",
      "conditions": {
        "skip_if": "${expression}",
        "run_if": "${expression}"
      }
    }
  ],
  "edges": [
    {
      "from": "string (required)",
      "to": "string (required)",
      "condition": "${expression} (optional)"
    }
  ],
  "config": {
    "timeout": 300,
    "max_concurrent_nodes": 10,
    "retry_policy": "exponential|linear|fixed",
    "on_failure": "stop|continue",
    "notifications": {
      "on_success": ["email@example.com"],
      "on_failure": ["email@example.com"]
    }
  },
  "variables": {
    "key": "value"
  },
  "secrets": {
    "api_key": "${secrets.my_api_key}"
  }
}
```

## Workflow JSON Schema

### Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Workflow name (unique) |
| `description` | string | No | Workflow description |
| `version` | integer | No | Workflow version (default: 1) |
| `tags` | array | No | Tags for categorization |
| `nodes` | array | Yes | Array of workflow nodes |
| `edges` | array | Yes | Array of node connections |
| `config` | object | No | Workflow-level configuration |
| `variables` | object | No | Workflow variables |
| `secrets` | object | No | Secret references |

### Node Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique node identifier |
| `type` | string | Yes | Node type (see [Node Types](./node-types.md)) |
| `name` | string | No | Human-readable node name |
| `description` | string | No | Node description |
| `config` | object | Yes | Node-specific configuration |
| `retry` | object | No | Retry configuration |
| `timeout` | integer | No | Timeout in seconds |
| `on_error` | string | No | Error handling: fail, continue, retry |
| `conditions` | object | No | Conditional execution |

### Edge Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from` | string | Yes | Source node ID |
| `to` | string | Yes | Target node ID |
| `condition` | string | No | Conditional edge (expression) |

## Node Parameters

### Input and Output

Nodes receive input from:
1. **Workflow Input**: Initial data passed to workflow
2. **Previous Node Output**: Data from upstream nodes
3. **Workflow Variables**: Defined in workflow config
4. **Secrets**: Secure credentials

```json
{
  "id": "process_data",
  "type": "transform",
  "config": {
    "script": "return { result: input.data * 2 }"
  }
}
```

## Expressions and Variables

### Expression Syntax

GoFlow supports JavaScript-like expressions using `${}` syntax:

```json
{
  "id": "send_email",
  "type": "email",
  "config": {
    "to": "${input.user.email}",
    "subject": "Hello ${input.user.name}!",
    "body": "Your order #${nodes.create_order.output.order_id} has been confirmed."
  }
}
```

### Available Context

- `input` - Workflow input data
- `nodes.<node_id>.output` - Output from specific node
- `variables` - Workflow variables
- `secrets` - Secure credentials
- `execution` - Execution metadata

### Examples

```json
{
  "config": {
    "url": "${variables.api_base_url}/users/${input.user_id}",
    "headers": {
      "Authorization": "Bearer ${secrets.api_token}"
    },
    "timeout": "${variables.default_timeout || 30}"
  }
}
```

### Workflow Variables

Define reusable variables at the workflow level:

```json
{
  "name": "API Integration Workflow",
  "variables": {
    "api_base_url": "https://api.example.com",
    "default_timeout": 30,
    "max_retries": 3,
    "environment": "production"
  },
  "nodes": [
    {
      "id": "api_call",
      "type": "http_request",
      "config": {
        "url": "${variables.api_base_url}/endpoint",
        "timeout": "${variables.default_timeout}"
      }
    }
  ]
}
```

### Secrets Management

Store sensitive data securely:

```json
{
  "name": "Secure Workflow",
  "secrets": {
    "api_key": "encrypted_value",
    "db_password": "encrypted_value",
    "jwt_secret": "encrypted_value"
  },
  "nodes": [
    {
      "id": "api_call",
      "type": "http_request",
      "config": {
        "headers": {
          "X-API-Key": "${secrets.api_key}"
        }
      }
    }
  ]
}
```

## Workflow Versioning

### Version Management

Workflows support versioning for safe updates:

```json
{
  "id": "wf_123",
  "name": "My Workflow",
  "version": 2,
  "changelog": "Added error handling node",
  "nodes": [...]
}
```

### Version Strategy

1. **Immutable Versions** - Published versions cannot be modified
2. **Draft Versions** - Work-in-progress versions can be edited
3. **Semantic Versioning** - Major.Minor.Patch format

### Creating New Version

```bash
# Create new version
POST /api/v1/workflows/{workflow_id}/versions
{
  "changelog": "Added new transformation step"
}

# Response
{
  "id": "wf_123",
  "version": 3,
  "status": "draft"
}
```

### Publishing Version

```bash
# Publish version
POST /api/v1/workflows/{workflow_id}/versions/{version}/publish

# Response
{
  "id": "wf_123",
  "version": 3,
  "status": "published",
  "published_at": "2024-01-01T10:00:00Z"
}
```

## Complete Examples

### Example 1: User Onboarding Workflow

```json
{
  "name": "User Onboarding",
  "version": 1,
  "description": "Automated user onboarding process",
  "variables": {
    "welcome_email_template": "welcome_v2",
    "slack_channel": "#new-users"
  },
  "nodes": [
    {
      "id": "webhook_trigger",
      "type": "webhook",
      "config": {
        "method": "POST",
        "path": "/webhooks/user-signup"
      }
    },
    {
      "id": "create_user",
      "type": "database",
      "config": {
        "query": "INSERT INTO users (email, name, created_at) VALUES ($1, $2, NOW()) RETURNING id",
        "params": ["${input.email}", "${input.name}"]
      }
    },
    {
      "id": "send_welcome_email",
      "type": "email",
      "config": {
        "to": "${input.email}",
        "subject": "Welcome to GoFlow!",
        "template": "${variables.welcome_email_template}",
        "data": {
          "name": "${input.name}",
          "user_id": "${nodes.create_user.output.rows[0].id}"
        }
      }
    },
    {
      "id": "notify_slack",
      "type": "http_request",
      "config": {
        "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
        "method": "POST",
        "body": {
          "channel": "${variables.slack_channel}",
          "text": "New user signed up: ${input.name} (${input.email})"
        }
      }
    }
  ],
  "edges": [
    {"from": "webhook_trigger", "to": "create_user"},
    {"from": "create_user", "to": "send_welcome_email"},
    {"from": "create_user", "to": "notify_slack"}
  ]
}
```

### Example 2: Data Processing Pipeline

```json
{
  "name": "Data Processing Pipeline",
  "version": 1,
  "description": "Process and analyze incoming data",
  "nodes": [
    {
      "id": "fetch_data",
      "type": "http_request",
      "config": {
        "url": "https://api.example.com/data",
        "method": "GET"
      }
    },
    {
      "id": "validate_data",
      "type": "conditional",
      "config": {
        "condition": "${nodes.fetch_data.output.body.records.length > 0}",
        "on_true": "process_records",
        "on_false": "send_error_notification"
      }
    },
    {
      "id": "process_records",
      "type": "loop",
      "config": {
        "items": "${nodes.fetch_data.output.body.records}",
        "item_variable": "record",
        "parallel": true
      }
    },
    {
      "id": "transform_record",
      "type": "transform",
      "config": {
        "script": "return { id: record.id, value: record.value * 2, processed_at: new Date().toISOString() }"
      }
    },
    {
      "id": "save_record",
      "type": "database",
      "config": {
        "query": "INSERT INTO processed_data (id, value, processed_at) VALUES ($1, $2, $3)",
        "params": ["${record.id}", "${record.value}", "${record.processed_at}"]
      }
    },
    {
      "id": "send_error_notification",
      "type": "email",
      "config": {
        "to": "admin@example.com",
        "subject": "Data Processing Error",
        "body": "No records found to process"
      }
    }
  ],
  "edges": [
    {"from": "fetch_data", "to": "validate_data"},
    {"from": "process_records", "to": "transform_record"},
    {"from": "transform_record", "to": "save_record"}
  ]
}
```

### Example 3: AI-Powered Content Generation

```json
{
  "name": "AI Content Generator",
  "version": 1,
  "description": "Generate and publish AI-powered content",
  "secrets": {
    "openai_api_key": "encrypted_key"
  },
  "nodes": [
    {
      "id": "generate_content",
      "type": "openai_completion",
      "config": {
        "model": "gpt-4",
        "prompt": "Write a blog post about: ${input.topic}",
        "max_tokens": 1000,
        "temperature": 0.7
      }
    },
    {
      "id": "review_content",
      "type": "conditional",
      "config": {
        "condition": "${nodes.generate_content.output.text.length > 500}",
        "on_true": "create_embedding",
        "on_false": "regenerate_content"
      }
    },
    {
      "id": "create_embedding",
      "type": "openai_embedding",
      "config": {
        "model": "text-embedding-ada-002",
        "input": "${nodes.generate_content.output.text}"
      }
    },
    {
      "id": "save_content",
      "type": "database",
      "config": {
        "query": "INSERT INTO content (topic, text, embedding, created_at) VALUES ($1, $2, $3, NOW())",
        "params": [
          "${input.topic}",
          "${nodes.generate_content.output.text}",
          "${nodes.create_embedding.output.embedding}"
        ]
      }
    },
    {
      "id": "regenerate_content",
      "type": "openai_completion",
      "config": {
        "model": "gpt-4",
        "prompt": "Write a longer, more detailed blog post about: ${input.topic}",
        "max_tokens": 2000,
        "temperature": 0.8
      }
    }
  ],
  "edges": [
    {"from": "generate_content", "to": "review_content"},
    {"from": "create_embedding", "to": "save_content"}
  ]
}
```

## Best Practices

### 1. Use Descriptive Node IDs

```json
// Good
{"id": "fetch_user_data", "type": "http_request"}
{"id": "validate_email", "type": "conditional"}

// Bad
{"id": "node1", "type": "http_request"}
{"id": "n2", "type": "conditional"}
```

### 2. Add Descriptions

```json
{
  "id": "process_payment",
  "type": "http_request",
  "description": "Charges the customer's credit card via Stripe API",
  "config": {...}
}
```

### 3. Handle Errors Gracefully

```json
{
  "id": "api_call",
  "type": "http_request",
  "config": {
    "url": "https://api.example.com",
    "retry": {
      "max_attempts": 3,
      "initial_delay": "1s"
    },
    "on_error": "send_error_notification"
  }
}
```

### 4. Use Variables for Configuration

```json
{
  "variables": {
    "api_base_url": "https://api.example.com",
    "timeout": 30
  },
  "nodes": [
    {
      "config": {
        "url": "${variables.api_base_url}/endpoint",
        "timeout": "${variables.timeout}"
      }
    }
  ]
}
```

### 5. Keep Workflows Focused

- One workflow should handle one business process
- Break complex workflows into smaller, reusable workflows
- Use sub-workflows for common patterns

---

For more information, see:
- [Node Types Documentation](./node-types.md)
- [API Documentation](../api/api.md)
- [Architecture Overview](../architecture.md)
