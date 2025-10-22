# Security Policy

## Supported Versions

We release patches for security vulnerabilities. The following versions are currently supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

The GoFlow team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by emailing:

**[security@goflow-atom.dev](mailto:security@goflow-atom.dev)**

Include the following information in your report:

- **Type of vulnerability** (e.g., SQL injection, XSS, authentication bypass)
- **Full paths of source file(s)** related to the vulnerability
- **Location of the affected source code** (tag/branch/commit or direct URL)
- **Step-by-step instructions** to reproduce the issue
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the vulnerability** - what can an attacker do?
- **Suggested fix** (if you have one)

### What to Expect

After you submit a report, you can expect:

1. **Acknowledgment** - We'll acknowledge receipt within 48 hours
2. **Initial Assessment** - We'll provide an initial assessment within 5 business days
3. **Updates** - We'll keep you informed about our progress
4. **Resolution** - We'll work on a fix and coordinate disclosure timing with you
5. **Credit** - We'll credit you in the security advisory (unless you prefer to remain anonymous)

### Response Timeline

- **Critical vulnerabilities**: Patch within 7 days
- **High severity**: Patch within 30 days
- **Medium severity**: Patch within 90 days
- **Low severity**: Patch in next regular release

### Disclosure Policy

- We follow **coordinated disclosure**
- We'll work with you to understand and resolve the issue
- We'll publicly disclose the vulnerability after a fix is released
- We'll credit you in the security advisory (with your permission)

## Security Best Practices

### For Users

When deploying GoFlow, follow these security best practices:

#### 1. Authentication & Authorization

- **Enable authentication** for all API endpoints
- **Use strong JWT secrets** (minimum 32 characters, randomly generated)
- **Implement RBAC** to control access to workflows and executions
- **Rotate secrets regularly** (at least every 90 days)

```yaml
# config.yaml
security:
  jwt_secret: "use-a-strong-randomly-generated-secret-here"
  jwt_expiration: 3600  # 1 hour
  enable_rbac: true
```

#### 2. Network Security

- **Use HTTPS/TLS** for all external communications
- **Enable TLS for database connections**
- **Use TLS for Redis connections**
- **Configure firewall rules** to restrict access
- **Use VPC/private networks** in cloud environments

```yaml
# config.yaml
database:
  ssl_mode: require
  ssl_cert: /path/to/cert.pem
  ssl_key: /path/to/key.pem

redis:
  tls_enabled: true
  tls_cert: /path/to/cert.pem
```

#### 3. Secret Management

- **Never hardcode secrets** in workflow definitions
- **Use environment variables** or secret management systems
- **Integrate with HashiCorp Vault** or AWS Secrets Manager
- **Encrypt secrets at rest** in the database

```json
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
```

#### 4. Input Validation

- **Validate all workflow definitions** before execution
- **Sanitize user inputs** to prevent injection attacks
- **Limit workflow complexity** (max nodes, max execution time)
- **Implement rate limiting** to prevent abuse

```yaml
# config.yaml
validation:
  max_nodes_per_workflow: 100
  max_execution_time: 3600  # 1 hour
  max_workflow_size: 1048576  # 1 MB

rate_limiting:
  enabled: true
  requests_per_minute: 100
```

#### 5. Database Security

- **Use parameterized queries** (we do this by default)
- **Enable database encryption** at rest
- **Use strong database passwords**
- **Limit database user permissions** (principle of least privilege)
- **Enable database audit logging**

```yaml
# config.yaml
database:
  user: goflow_app  # Not a superuser
  password: "strong-randomly-generated-password"
  ssl_mode: require
  max_connections: 100
```

#### 6. Monitoring & Logging

- **Enable audit logging** for all operations
- **Monitor for suspicious activity**
- **Set up alerts** for security events
- **Regularly review logs**
- **Use centralized logging** (ELK, Splunk, etc.)

```yaml
# config.yaml
logging:
  level: info
  audit_enabled: true
  audit_log_path: /var/log/goflow/audit.log
```

#### 7. Updates & Patches

- **Keep GoFlow updated** to the latest version
- **Subscribe to security advisories**
- **Apply security patches promptly**
- **Test updates in staging** before production

#### 8. Container Security

If running in containers:

- **Use official images** from trusted registries
- **Scan images for vulnerabilities** (Trivy, Clair)
- **Run as non-root user**
- **Use read-only file systems** where possible
- **Limit container capabilities**

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
# ... build steps ...

FROM alpine:latest
RUN addgroup -g 1000 goflow && \
    adduser -D -u 1000 -G goflow goflow
USER goflow
# ... rest of Dockerfile ...
```

### For Developers

When contributing to GoFlow:

#### 1. Secure Coding Practices

- **Validate all inputs** at API boundaries
- **Use parameterized queries** for database operations
- **Avoid eval() or similar** dynamic code execution
- **Sanitize outputs** to prevent XSS
- **Use crypto/rand** for random number generation (not math/rand)

#### 2. Dependency Management

- **Keep dependencies updated**
- **Review dependency security advisories**
- **Use `go mod verify`** to check integrity
- **Avoid dependencies with known vulnerabilities**

```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

#### 3. Code Review

- **Review code for security issues**
- **Use static analysis tools** (gosec, golangci-lint)
- **Check for hardcoded secrets**
- **Verify error handling**

```bash
# Run security scanner
gosec ./...
```

#### 4. Testing

- **Write security tests**
- **Test authentication/authorization**
- **Test input validation**
- **Test error handling**
- **Perform penetration testing**

## Known Security Considerations

### 1. Workflow Execution

Workflows can execute arbitrary HTTP requests and database queries. Ensure:
- Workflows are created by trusted users only
- Network access is restricted appropriately
- Database permissions are limited

### 2. Expression Evaluation

The expression evaluator supports JSONPath and template strings. While we sanitize inputs, be cautious with:
- User-provided expressions
- External data sources
- Complex nested expressions

### 3. OpenAI Integration

When using OpenAI nodes:
- API keys are stored encrypted
- Prompts may contain sensitive data
- Rate limiting is enforced
- Costs can accumulate quickly

### 4. Webhook Endpoints

Webhook endpoints are publicly accessible. Ensure:
- HMAC signature validation is enabled
- IP whitelisting is configured (if possible)
- Rate limiting is active
- Webhook secrets are strong

## Security Advisories

Security advisories are published at:
- [GitHub Security Advisories](https://github.com/goflow-atom/goflow-service/security/advisories)
- [Security Mailing List](mailto:security-announce@goflow-atom.dev)

Subscribe to receive notifications about security updates.

## Bug Bounty Program

We currently do not have a formal bug bounty program, but we deeply appreciate security researchers who responsibly disclose vulnerabilities. We will:

- Publicly acknowledge your contribution (with your permission)
- Provide a detailed write-up of the fix
- Consider your findings for future bug bounty programs

## Contact

For security-related questions or concerns:

- **Security Team**: [security@goflow-atom.dev](mailto:security@goflow-atom.dev)
- **PGP Key**: Available at [https://goflow-atom.dev/security/pgp-key.asc](https://goflow-atom.dev/security/pgp-key.asc)

---

**Thank you for helping keep GoFlow and our users safe!** 🔒
