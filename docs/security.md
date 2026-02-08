# Security

http2cli is designed with security in mind. This document describes the security measures in place.

## Command Injection Prevention

The most critical security concern is command injection. http2cli prevents this through several measures:

### No Shell Interpreter

Commands are executed directly using Go's `exec.Command`, not through a shell. This means:

- No shell metacharacters are interpreted (`;`, `|`, `&`, etc.)
- No environment variable expansion
- No glob expansion
- No command substitution

### Argument Separation

Arguments are passed as separate strings to `exec.Command`, not as a single concatenated string. This prevents injection through argument manipulation.

```go
// Safe - arguments are separate
exec.Command("/bin/date", "+%Y-%m-%d")

// Unsafe - would allow injection (NOT USED)
exec.Command("sh", "-c", "/bin/date +%Y-%m-%d")
```

### Command Allowlist

Only commands explicitly listed in `security.allowed_commands` can be executed. Commands not in this list are rejected before execution.

```yaml
security:
  allowed_commands:
    - /bin/cat
    - /bin/date
```

## API Authentication

Optional API key authentication using `X-API-Key` header:

```yaml
security:
  api_key: "your-secret-key"
```

When configured:
- All requests (except /health and /ready) require the header
- Requests with invalid or missing keys receive 401 Unauthorized

## Resource Protection

### Timeouts

Every command execution has a timeout to prevent resource exhaustion:

```yaml
security:
  command_timeout: 30s

tools:
  - name: slow-tool
    timeout: 120s  # Override for specific tool
```

### Concurrent Execution Limits

Limit the number of commands running simultaneously:

```yaml
security:
  max_concurrent_commands: 10
```

Requests exceeding this limit wait or timeout.

## Input Validation

### Parameter Validation

All parameters are validated against their type definitions:

- `enum`: Must be one of allowed values
- `int`: Must be valid integer within min/max bounds
- `string`: Optional regex pattern matching
- `bool`: Must be valid boolean string

Invalid parameters receive 400 Bad Request.

### File Handling

- Files are read into memory, processed, and discarded
- No user-provided filenames are used in command execution
- File extensions can be validated per-tool (planned feature)

## Best Practices

1. **Use full paths** for commands in `allowed_commands`
2. **Set appropriate timeouts** to prevent resource exhaustion
3. **Enable API key** for public-facing deployments
4. **Use HTTPS** via reverse proxy (nginx, traefik, etc.)
5. **Run as non-root** user in Docker
6. **Use read-only filesystem** to prevent modification of binaries at runtime
7. **Limit exposed commands** to only what's necessary

### Docker Security Example

```yaml
services:
  http2cli:
    image: http2cli:latest
    user: "1000:1000"           # Run as non-root
    read_only: true             # Prevent filesystem modifications
    security_opt:
      - no-new-privileges:true  # Prevent privilege escalation
    volumes:
      - ./config.yaml:/etc/http2cli/config.yaml:ro
```
