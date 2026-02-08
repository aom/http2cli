# API Reference

## Standard Endpoints

### Health Check

```
GET /health
```

Returns server health status.

**Response:**
```json
{"status": "ok"}
```

### Readiness Check

```
GET /ready
```

Checks if all configured tools are available.

**Response (ready):**
```json
{"status": "ready"}
```

**Response (not ready):**
```json
{
  "status": "not ready",
  "errors": ["command /usr/bin/foo not allowed"]
}
```

### List Tools

```
GET /api/tools
```

Returns list of available tools.

**Response:**
```json
{
  "tools": [
    {
      "name": "cat",
      "description": "Pass file contents through",
      "endpoint": "/cat",
      "method": "POST"
    }
  ]
}
```

## Tool Endpoints

Tool endpoints are defined in configuration. Each tool has its own endpoint.

### Request Format

**For POST endpoints with file input:**

```http
POST /endpoint?param=value HTTP/1.1
Content-Type: multipart/form-data; boundary=----boundary

------boundary
Content-Disposition: form-data; name="file"; filename="input.txt"
Content-Type: application/octet-stream

<file content>
------boundary--
```

**For GET endpoints:**

```http
GET /endpoint?param=value HTTP/1.1
```

### Response Format

**Success:**
```http
HTTP/1.1 200 OK
Content-Type: <configured content type>
X-Processing-Time: 1.234ms

<command output>
```

**Error:**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "Bad Request",
  "message": "required file 'file' not provided"
}
```

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad Request - invalid parameters or missing required fields |
| 401 | Unauthorized - invalid or missing API key |
| 405 | Method Not Allowed |
| 422 | Unprocessable Entity - command failed |
| 500 | Internal Server Error |
| 504 | Gateway Timeout - command timed out |

## Authentication

If `security.api_key` is configured, all requests (except /health and /ready) require the `X-API-Key` header.

```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8080/api/tools
```

## Examples

### Using curl

```bash
# Date with format
curl "http://localhost:8080/date?format=%Y-%m-%d"

# Cat with file upload
curl -F "file=@input.txt" http://localhost:8080/cat

# Rev from stdin
echo "hello" | curl -F "text=@-" http://localhost:8080/rev

# With API key
curl -H "X-API-Key: secret" -F "file=@input.txt" http://localhost:8080/cat
```
