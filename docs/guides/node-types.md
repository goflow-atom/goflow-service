# Node Types

This document provides comprehensive documentation for all available node types in GoFlow.

## Table of Contents

- [Node Types](#node-types)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
    - [NodeExecutor Interface](#nodeexecutor-interface)
    - [Polymorphic Design](#polymorphic-design)
  - [Available Node Types](#available-node-types)
    - [Webhook](#webhook)
    - [HTTP Request](#http-request)
    - [Conditional](#conditional)
    - [Loop](#loop)
    - [Parallel](#parallel)
    - [Transform](#transform)
    - [Delay](#delay)
    - [Database](#database)
    - [Email](#email)
    - [OpenAI Completion](#openai-completion)
    - [OpenAI Embedding](#openai-embedding)
  - [Adding New Node Types](#adding-new-node-types)
    - [Step 1: Define Node Interface](#step-1-define-node-interface)
    - [Step 2: Register in Factory](#step-2-register-in-factory)
    - [Step 3: Add Constants](#step-3-add-constants)
    - [Step 4: Write Tests](#step-4-write-tests)
    - [Step 5: Document the Node Type](#step-5-document-the-node-type)
    - [Step 6: Add Example Workflow](#step-6-add-example-workflow)

## Overview

Nodes are the building blocks of workflows in GoFlow. Each node type implements specific functionality and can be configured with parameters, retry policies, and error handling.

### NodeExecutor Interface

All node types implement the `NodeExecutor` interface:

```go
// NodeExecutor defines the interface for executing workflow nodes
type NodeExecutor interface {
    // Execute runs the node with the given context and input
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

    // Validate checks if the node configuration is valid
    Validate() error

    // Type returns the node type identifier
    Type() string
}
```

### Polymorphic Design

The Node Executor Factory creates appropriate executors based on node type:

```go
// NodeExecutorFactory creates node executors
type NodeExecutorFactory struct {
    httpClient    *http.Client
    dbClient      *database.Client
    openaiClient  *openai.Client
    emailClient   *email.Client
}

func (f *NodeExecutorFactory) Create(node *domain.Node) (NodeExecutor, error) {
    switch node.Type {
    case "webhook":
        return NewWebhookExecutor(node.Config), nil
    case "http_request":
        return NewHTTPRequestExecutor(node.Config, f.httpClient), nil
    case "conditional":
        return NewConditionalExecutor(node.Config), nil
    // ... other node types
    default:
        return nil, fmt.Errorf("unknown node type: %s", node.Type)
    }
}
```

## Available Node Types

### Webhook

**Type**: `webhook`

**Description**: Receives HTTP requests from external services to trigger workflow execution.

**Configuration**:

```json
{
  "id": "webhook_trigger",
  "type": "webhook",
  "config": {
    "method": "POST",
    "path": "/webhooks/my-webhook",
    "validate_signature": true,
    "response": {
      "status": 200,
      "body": {
        "message": "Webhook received"
      }
    }
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `method` | string | Yes | HTTP method (GET, POST, PUT, DELETE) |
| `path` | string | Yes | Webhook URL path |
| `validate_signature` | boolean | No | Enable signature validation (default: false) |
| `response` | object | No | Custom response configuration |

**Input**: HTTP request data

```json
{
  "body": {...},
  "headers": {...},
  "query": {...},
  "method": "POST",
  "path": "/webhooks/my-webhook"
}
```

**Output**: Same as input (passes through request data)

**Example**:

```json
{
  "name": "GitHub Webhook Handler",
  "nodes": [
    {
      "id": "github_webhook",
      "type": "webhook",
      "config": {
        "method": "POST",
        "path": "/webhooks/github",
        "validate_signature": true
      }
    },
    {
      "id": "process_push",
      "type": "transform",
      "config": {
        "script": "return { repo: input.body.repository.name, commits: input.body.commits.length }"
      }
    }
  ],
  "edges": [
    {"from": "github_webhook", "to": "process_push"}
  ]
}
```

### HTTP Request

**Type**: `http_request`

**Description**: Makes HTTP requests to external APIs.

**Configuration**:

```json
{
  "id": "api_call",
  "type": "http_request",
  "config": {
    "url": "https://api.example.com/users",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json",
      "Authorization": "Bearer ${secrets.api_token}"
    },
    "body": {
      "name": "${input.name}",
      "email": "${input.email}"
    },
    "timeout": 30,
    "follow_redirects": true,
    "validate_ssl": true
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | Yes | Request URL (supports expressions) |
| `method` | string | Yes | HTTP method (GET, POST, PUT, PATCH, DELETE) |
| `headers` | object | No | Request headers |
| `body` | object/string | No | Request body |
| `query` | object | No | Query parameters |
| `timeout` | integer | No | Timeout in seconds (default: 30) |
| `follow_redirects` | boolean | No | Follow redirects (default: true) |
| `validate_ssl` | boolean | No | Validate SSL certificates (default: true) |

**Input**: Data for URL/body interpolation

**Output**:

```json
{
  "status": 200,
  "headers": {...},
  "body": {...},
  "duration_ms": 123
}
```

**Example**:

```json
{
  "id": "create_user",
  "type": "http_request",
  "config": {
    "url": "https://api.example.com/users",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "name": "${input.name}",
      "email": "${input.email}"
    }
  }
}
```

### Conditional

**Type**: `conditional`

**Description**: Branches workflow execution based on conditions.

**Configuration**:

```json
{
  "id": "check_age",
  "type": "conditional",
  "config": {
    "condition": "${input.age >= 18}",
    "on_true": "adult_flow",
    "on_false": "minor_flow"
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `condition` | string | Yes | Boolean expression to evaluate |
| `on_true` | string | Yes | Node ID to execute if condition is true |
| `on_false` | string | No | Node ID to execute if condition is false |

**Input**: Data for condition evaluation

**Output**:

```json
{
  "condition_result": true,
  "next_node": "adult_flow"
}
```

**Example**:

```json
{
  "nodes": [
    {
      "id": "check_status",
      "type": "conditional",
      "config": {
        "condition": "${nodes.api_call.output.status === 'success'}",
        "on_true": "send_success_email",
        "on_false": "send_error_email"
      }
    }
  ]
}
```

### Loop

**Type**: `loop`

**Description**: Iterates over a collection and executes nodes for each item.

**Configuration**:

```json
{
  "id": "process_users",
  "type": "loop",
  "config": {
    "items": "${input.users}",
    "item_variable": "user",
    "max_iterations": 100,
    "parallel": false,
    "continue_on_error": true
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `items` | string/array | Yes | Array to iterate over (expression or literal) |
| `item_variable` | string | No | Variable name for current item (default: "item") |
| `max_iterations` | integer | No | Maximum iterations (default: 1000) |
| `parallel` | boolean | No | Execute iterations in parallel (default: false) |
| `continue_on_error` | boolean | No | Continue on iteration error (default: false) |

**Input**: Data containing the array to iterate

**Output**:

```json
{
  "iterations": 10,
  "results": [
    {"index": 0, "output": {...}},
    {"index": 1, "output": {...}}
  ],
  "errors": []
}
```

**Example**:

```json
{
  "nodes": [
    {
      "id": "loop_users",
      "type": "loop",
      "config": {
        "items": "${input.users}",
        "item_variable": "user",
        "parallel": true
      }
    },
    {
      "id": "send_email",
      "type": "email",
      "config": {
        "to": "${user.email}",
        "subject": "Welcome!",
        "body": "Hello ${user.name}"
      }
    }
  ],
  "edges": [
    {"from": "loop_users", "to": "send_email"}
  ]
}
```

### Parallel

**Type**: `parallel`

**Description**: Executes multiple branches concurrently.

**Configuration**:

```json
{
  "id": "parallel_tasks",
  "type": "parallel",
  "config": {
    "branches": [
      {"nodes": ["fetch_user", "process_user"]},
      {"nodes": ["fetch_orders", "process_orders"]},
      {"nodes": ["fetch_analytics", "process_analytics"]}
    ],
    "wait_for_all": true,
    "fail_fast": false,
    "timeout": 60
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `branches` | array | Yes | Array of branch configurations |
| `wait_for_all` | boolean | No | Wait for all branches (default: true) |
| `fail_fast` | boolean | No | Stop all on first failure (default: false) |
| `timeout` | integer | No | Timeout for all branches in seconds |

**Input**: Data passed to all branches

**Output**:

```json
{
  "branches": [
    {"index": 0, "status": "completed", "output": {...}},
    {"index": 1, "status": "completed", "output": {...}},
    {"index": 2, "status": "failed", "error": "..."}
  ],
  "all_succeeded": false
}
```

**Example**:

```json
{
  "nodes": [
    {
      "id": "parallel_apis",
      "type": "parallel",
      "config": {
        "branches": [
          {"nodes": ["call_api_1"]},
          {"nodes": ["call_api_2"]},
          {"nodes": ["call_api_3"]}
        ],
        "wait_for_all": true
      }
    }
  ]
}
```

### Transform

**Type**: `transform`

**Description**: Transforms data using JavaScript expressions.

**Configuration**:

```json
{
  "id": "transform_data",
  "type": "transform",
  "config": {
    "script": "return { fullName: input.firstName + ' ' + input.lastName, age: input.age }",
    "timeout": 5
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `script` | string | Yes | JavaScript code to execute |
| `timeout` | integer | No | Script timeout in seconds (default: 5) |

**Input**: Data available as `input` variable in script

**Output**: Result returned by the script

**Example**:

```json
{
  "id": "calculate_total",
  "type": "transform",
  "config": {
    "script": "return { total: input.items.reduce((sum, item) => sum + item.price, 0) }"
  }
}
```

### Delay

**Type**: `delay`

**Description**: Pauses workflow execution for a specified duration.

**Configuration**:

```json
{
  "id": "wait_5_seconds",
  "type": "delay",
  "config": {
    "duration": 5,
    "unit": "seconds"
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `duration` | integer | Yes | Delay duration |
| `unit` | string | No | Time unit: seconds, minutes, hours (default: seconds) |

**Input**: Any (passed through)

**Output**: Same as input (passes through)

**Example**:

```json
{
  "nodes": [
    {
      "id": "send_email",
      "type": "email",
      "config": {...}
    },
    {
      "id": "wait_1_hour",
      "type": "delay",
      "config": {
        "duration": 1,
        "unit": "hours"
      }
    },
    {
      "id": "send_followup",
      "type": "email",
      "config": {...}
    }
  ]
}
```

### Database

**Type**: `database`

**Description**: Executes SQL queries against a database.

**Configuration**:

```json
{
  "id": "query_users",
  "type": "database",
  "config": {
    "connection": "default",
    "query": "SELECT * FROM users WHERE age > $1 AND status = $2",
    "params": ["${input.min_age}", "active"],
    "timeout": 30
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `connection` | string | No | Database connection name (default: "default") |
| `query` | string | Yes | SQL query to execute |
| `params` | array | No | Query parameters (prevents SQL injection) |
| `timeout` | integer | No | Query timeout in seconds (default: 30) |

**Input**: Data for parameter interpolation

**Output**:

```json
{
  "rows": [
    {"id": 1, "name": "John", "age": 25},
    {"id": 2, "name": "Jane", "age": 30}
  ],
  "row_count": 2,
  "duration_ms": 45
}
```

**Example**:

```json
{
  "id": "insert_user",
  "type": "database",
  "config": {
    "query": "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
    "params": ["${input.name}", "${input.email}"]
  }
}
```

### Email

**Type**: `email`

**Description**: Sends emails via SMTP.

**Configuration**:

```json
{
  "id": "send_welcome_email",
  "type": "email",
  "config": {
    "to": "${input.email}",
    "cc": ["admin@example.com"],
    "subject": "Welcome to GoFlow!",
    "body": "Hello ${input.name}, welcome to our platform!",
    "html": true,
    "attachments": [
      {
        "filename": "welcome.pdf",
        "path": "/path/to/welcome.pdf"
      }
    ]
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `to` | string/array | Yes | Recipient email address(es) |
| `cc` | array | No | CC recipients |
| `bcc` | array | No | BCC recipients |
| `subject` | string | Yes | Email subject |
| `body` | string | Yes | Email body |
| `html` | boolean | No | Send as HTML (default: false) |
| `attachments` | array | No | Email attachments |
| `template` | string | No | Email template name |

**Input**: Data for email interpolation

**Output**:

```json
{
  "message_id": "abc123@mail.example.com",
  "sent_at": "2024-01-01T10:00:00Z",
  "recipients": ["user@example.com"]
}
```

**Example**:

```json
{
  "id": "send_notification",
  "type": "email",
  "config": {
    "to": "${input.user.email}",
    "subject": "Order Confirmation #${input.order_id}",
    "template": "order_confirmation",
    "html": true
  }
}
```

### OpenAI Completion

**Type**: `openai_completion`

**Description**: Generates text using OpenAI's GPT models.

**Configuration**:

```json
{
  "id": "generate_summary",
  "type": "openai_completion",
  "config": {
    "model": "gpt-4",
    "prompt": "Summarize the following text:\n\n${input.text}",
    "max_tokens": 150,
    "temperature": 0.7,
    "top_p": 1.0,
    "frequency_penalty": 0.0,
    "presence_penalty": 0.0
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `model` | string | Yes | OpenAI model (gpt-4, gpt-3.5-turbo, etc.) |
| `prompt` | string | Yes | Prompt text |
| `max_tokens` | integer | No | Maximum tokens to generate (default: 100) |
| `temperature` | float | No | Sampling temperature 0-2 (default: 0.7) |
| `top_p` | float | No | Nucleus sampling (default: 1.0) |
| `frequency_penalty` | float | No | Frequency penalty 0-2 (default: 0.0) |
| `presence_penalty` | float | No | Presence penalty 0-2 (default: 0.0) |
| `stop` | array | No | Stop sequences |

**Input**: Data for prompt interpolation

**Output**:

```json
{
  "text": "Generated text response...",
  "model": "gpt-4",
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 100,
    "total_tokens": 150
  },
  "finish_reason": "stop"
}
```

**Example**:

```json
{
  "id": "generate_response",
  "type": "openai_completion",
  "config": {
    "model": "gpt-4",
    "prompt": "Write a professional email response to: ${input.customer_message}",
    "max_tokens": 200,
    "temperature": 0.8
  }
}
```

### OpenAI Embedding

**Type**: `openai_embedding`

**Description**: Creates vector embeddings for semantic search.

**Configuration**:

```json
{
  "id": "create_embedding",
  "type": "openai_embedding",
  "config": {
    "model": "text-embedding-ada-002",
    "input": "${input.text}"
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `model` | string | Yes | Embedding model (text-embedding-ada-002, etc.) |
| `input` | string/array | Yes | Text to embed |

**Input**: Data for embedding

**Output**:

```json
{
  "embedding": [0.123, -0.456, 0.789, ...],
  "model": "text-embedding-ada-002",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

**Example**:

```json
{
  "id": "embed_document",
  "type": "openai_embedding",
  "config": {
    "model": "text-embedding-ada-002",
    "input": "${input.document_text}"
  }
}
```

## Adding New Node Types

### Step 1: Define Node Interface

Create a new file in `internal/domain/node_types/`:

```go
// internal/domain/node_types/custom_node.go
package node_types

import (
    "context"
    "fmt"
)

// CustomNodeConfig defines configuration for custom node
type CustomNodeConfig struct {
    Parameter1 string `json:"parameter1"`
    Parameter2 int    `json:"parameter2"`
}

// CustomNodeExecutor implements NodeExecutor for custom nodes
type CustomNodeExecutor struct {
    config CustomNodeConfig
}

// NewCustomNodeExecutor creates a new custom node executor
func NewCustomNodeExecutor(config map[string]interface{}) (*CustomNodeExecutor, error) {
    var nodeConfig CustomNodeConfig
    if err := mapToStruct(config, &nodeConfig); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    return &CustomNodeExecutor{
        config: nodeConfig,
    }, nil
}

// Execute runs the custom node
func (e *CustomNodeExecutor) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    // Implement your custom logic here
    result := map[string]interface{}{
        "output": "custom result",
    }

    return result, nil
}

// Validate checks if the configuration is valid
func (e *CustomNodeExecutor) Validate() error {
    if e.config.Parameter1 == "" {
        return fmt.Errorf("parameter1 is required")
    }
    return nil
}

// Type returns the node type identifier
func (e *CustomNodeExecutor) Type() string {
    return "custom_node"
}
```

### Step 2: Register in Factory

Update `internal/service/node_executor_factory.go`:

```go
func (f *NodeExecutorFactory) Create(node *domain.Node) (NodeExecutor, error) {
    switch node.Type {
    case "webhook":
        return node_types.NewWebhookExecutor(node.Config)
    case "http_request":
        return node_types.NewHTTPRequestExecutor(node.Config, f.httpClient)
    // ... other node types
    case "custom_node":
        return node_types.NewCustomNodeExecutor(node.Config)
    default:
        return nil, fmt.Errorf("unknown node type: %s", node.Type)
    }
}
```

### Step 3: Add Constants

Update `pkg/constants/node_types.go`:

```go
package constants

const (
    NodeTypeWebhook          = "webhook"
    NodeTypeHTTPRequest      = "http_request"
    // ... other node types
    NodeTypeCustomNode       = "custom_node"
)
```

### Step 4: Write Tests

Create `internal/domain/node_types/custom_node_test.go`:

```go
package node_types

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCustomNodeExecutor_Execute(t *testing.T) {
    config := map[string]interface{}{
        "parameter1": "value1",
        "parameter2": 42,
    }

    executor, err := NewCustomNodeExecutor(config)
    require.NoError(t, err)

    input := map[string]interface{}{
        "test": "data",
    }

    output, err := executor.Execute(context.Background(), input)
    require.NoError(t, err)
    assert.NotNil(t, output)
}

func TestCustomNodeExecutor_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  map[string]interface{}
        wantErr bool
    }{
        {
            name: "valid config",
            config: map[string]interface{}{
                "parameter1": "value1",
                "parameter2": 42,
            },
            wantErr: false,
        },
        {
            name: "missing parameter1",
            config: map[string]interface{}{
                "parameter2": 42,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            executor, err := NewCustomNodeExecutor(tt.config)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.NoError(t, executor.Validate())
            }
        })
    }
}
```

### Step 5: Document the Node Type

Add documentation to this file with:
- Node type identifier
- Description
- Configuration parameters
- Input/output specifications
- Usage examples

### Step 6: Add Example Workflow

Create an example in `examples/workflows/`:

```json
{
  "name": "Custom Node Example",
  "nodes": [
    {
      "id": "custom_1",
      "type": "custom_node",
      "config": {
        "parameter1": "value1",
        "parameter2": 42
      }
    }
  ]
}
```

---

For more information, see:
- [Workflow Definition Guide](./workflow-definition.md)
- [Architecture Documentation](../architecture.md)
- [Contributing Guide](../development/contributing.md)
---

For more information, see:
- [Workflow Definition Guide](./workflow-definition.md)
- [Architecture Documentation](../architecture.md)
- [Contributing Guide](../development/contributing.md)
