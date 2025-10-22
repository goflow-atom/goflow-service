# Authentication

This document describes the authentication and authorization mechanisms used by the GoFlow API.

## Table of Contents

- [Overview](#overview)
- [JWT-Based Authentication](#jwt-based-authentication)
- [Authorization Flow](#authorization-flow)
- [Role-Based Access Control](#role-based-access-control)
- [Security Practices](#security-practices)
- [API Key Authentication](#api-key-authentication)

## Overview

GoFlow uses JWT (JSON Web Tokens) for authentication and implements Role-Based Access Control (RBAC) for authorization. All API requests must include valid authentication credentials.

## JWT-Based Authentication

### Token Generation

Tokens are generated upon successful login:

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "your-password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9.eyJzdWIiOiJ1c2VyQGV4YW1wbGUuY29tIiwicm9sZXMiOlsiYWRtaW4iXSwiZXhwIjoxNjQwOTk1MjAwfQ.signature",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9...",
  "expires_at": "2024-01-01T12:00:00Z",
  "token_type": "Bearer"
}
```

### Token Structure

JWT tokens contain the following claims:

```json
{
  "sub": "user@example.com",
  "user_id": "usr_123abc",
  "roles": ["admin", "developer"],
  "permissions": ["workflow:create", "workflow:execute"],
  "iat": 1640908800,
  "exp": 1640995200,
  "iss": "goflow-api",
  "aud": "goflow-client"
}
```

**Claims:**
- `sub`: Subject (username/email)
- `user_id`: Unique user identifier
- `roles`: User roles
- `permissions`: Granted permissions
- `iat`: Issued at timestamp
- `exp`: Expiration timestamp
- `iss`: Issuer
- `aud`: Audience

### Using Tokens

Include the token in the `Authorization` header:

```http
GET /api/v1/workflows
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9...
```

### Token Refresh

Refresh expired tokens using the refresh token:

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9..."
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9...",
  "expires_at": "2024-01-01T13:00:00Z"
}
```

### Token Revocation

Revoke tokens (logout):

```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp9...
```

**Response:**
```json
{
  "message": "Token revoked successfully"
}
```

## Authorization Flow

### 1. User Login

```
┌──────┐                ┌──────────┐                ┌──────────┐
│Client│                │   API    │                │ Database │
└──┬───┘                └────┬─────┘                └────┬─────┘
   │                         │                           │
   │ POST /auth/login        │                           │
   │ {username, password}    │                           │
   ├────────────────────────>│                           │
   │                         │                           │
   │                         │ Verify credentials        │
   │                         ├──────────────────────────>│
   │                         │                           │
   │                         │ User + Roles              │
   │                         │<──────────────────────────┤
   │                         │                           │
   │                         │ Generate JWT              │
   │                         │                           │
   │ {token, refresh_token}  │                           │
   │<────────────────────────┤                           │
   │                         │                           │
```

### 2. Authenticated Request

```
┌──────┐                ┌──────────┐                ┌──────────┐
│Client│                │   API    │                │ Database │
└──┬───┘                └────┬─────┘                └────┬─────┘
   │                         │                           │
   │ GET /workflows          │                           │
   │ Authorization: Bearer   │                           │
   ├────────────────────────>│                           │
   │                         │                           │
   │                         │ Validate JWT              │
   │                         │                           │
   │                         │ Check permissions         │
   │                         │                           │
   │                         │ Query workflows           │
   │                         ├──────────────────────────>│
   │                         │                           │
   │                         │ Workflows                 │
   │                         │<──────────────────────────┤
   │                         │                           │
   │ {workflows}             │                           │
   │<────────────────────────┤                           │
   │                         │                           │
```

## Role-Based Access Control

### Roles

GoFlow defines the following roles:

| Role | Description |
|------|-------------|
| `admin` | Full system access |
| `developer` | Create and manage workflows |
| `operator` | Execute and monitor workflows |
| `viewer` | Read-only access |

### Permissions

Each role has specific permissions:

#### Admin Role
- `workflow:create` - Create workflows
- `workflow:read` - View workflows
- `workflow:update` - Update workflows
- `workflow:delete` - Delete workflows
- `workflow:execute` - Execute workflows
- `execution:read` - View executions
- `execution:cancel` - Cancel executions
- `webhook:create` - Create webhooks
- `webhook:read` - View webhooks
- `webhook:delete` - Delete webhooks
- `schedule:create` - Create schedules
- `schedule:read` - View schedules
- `schedule:update` - Update schedules
- `schedule:delete` - Delete schedules
- `user:manage` - Manage users

#### Developer Role
- `workflow:create`
- `workflow:read`
- `workflow:update`
- `workflow:execute`
- `execution:read`
- `webhook:create`
- `webhook:read`
- `schedule:create`
- `schedule:read`

#### Operator Role
- `workflow:read`
- `workflow:execute`
- `execution:read`
- `execution:cancel`
- `webhook:read`
- `schedule:read`

#### Viewer Role
- `workflow:read`
- `execution:read`
- `webhook:read`
- `schedule:read`

### Permission Checks

API endpoints enforce permissions:

```go
// Example: Create workflow endpoint
// Requires: workflow:create permission

POST /api/v1/workflows
Authorization: Bearer <token>

// If user lacks permission:
{
  "error": {
    "code": "AUTHORIZATION_ERROR",
    "message": "Insufficient permissions",
    "required_permission": "workflow:create"
  }
}
```

### Resource Ownership

Users can only access resources they own or have been granted access to:

```json
{
  "id": "wf_123abc",
  "name": "My Workflow",
  "owner_id": "usr_123",
  "shared_with": [
    {
      "user_id": "usr_456",
      "permissions": ["read", "execute"]
    }
  ]
}
```

## Security Practices

### Token Expiration

- **Access Token**: 1 hour expiration
- **Refresh Token**: 30 days expiration
- Tokens automatically expire and must be refreshed

### Token Storage

**Client-Side Best Practices:**
- Store tokens in memory or secure storage (not localStorage)
- Use httpOnly cookies for web applications
- Clear tokens on logout

### HTTPS Enforcement

All API requests must use HTTPS in production:
```
https://api.goflow.example.com/api/v1/...
```

HTTP requests are automatically redirected to HTTPS.

### Password Requirements

- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

### Rate Limiting

Authentication endpoints are rate limited:
- Login: 5 attempts per 15 minutes per IP
- Refresh: 10 attempts per hour per user
- Password reset: 3 attempts per hour per email

### Account Lockout

After 5 failed login attempts, accounts are locked for 15 minutes.

### Two-Factor Authentication (2FA)

Optional 2FA using TOTP (Time-based One-Time Password):

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "your-password",
  "totp_code": "123456"
}
```

## API Key Authentication

For service-to-service communication, use API keys:

### Creating API Keys

```http
POST /api/v1/auth/api-keys
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Production Service",
  "permissions": ["workflow:execute", "execution:read"],
  "expires_at": "2025-01-01T00:00:00Z"
}
```

**Response:**
```json
{
  "id": "key_123abc",
  "name": "Production Service",
  "key": "gf_live_abc123def456...",
  "permissions": ["workflow:execute", "execution:read"],
  "created_at": "2024-01-01T10:00:00Z",
  "expires_at": "2025-01-01T00:00:00Z"
}
```

### Using API Keys

Include the API key in the `X-API-Key` header:

```http
GET /api/v1/workflows
X-API-Key: gf_live_abc123def456...
```

### API Key Rotation

Rotate API keys regularly:

```http
POST /api/v1/auth/api-keys/:id/rotate
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": "key_123abc",
  "key": "gf_live_new_key_789...",
  "rotated_at": "2024-06-01T10:00:00Z"
}
```

---

For more information, see:
- [API Documentation](./api.md)
- [Security Best Practices](../guides/security.md)

