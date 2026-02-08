# Scaling http2cli

This guide explains how to deploy http2cli in a multi-container setup for isolation, modularity, and horizontal scaling.

## Architecture

```
                         ┌──────────────────────────────┐
                         │         nginx/traefik        │
    HTTP request ──────▶ │         (reverse proxy)      │
                         └──────────────────────────────┘
                                      │
              ┌───────────────────────┼───────────────────────┐
              ▼                       ▼                       ▼
     ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
     │   pdf-tools     │    │   ocr-tools     │    │   base-tools    │
     │   (debian)      │    │   (alpine)      │    │   (alpine)      │
     │                 │    │                 │    │                 │
     │ http2cli        │    │ http2cli        │    │ http2cli        │
     │ wkhtmltopdf     │    │ tesseract       │    │ cat, date, rev  │
     │                 │    │                 │    │                 │
     │ :8080           │    │ :8080           │    │ :8080           │
     └─────────────────┘    └─────────────────┘    └─────────────────┘
```

Each tool container:
- Has its own http2cli binary (~5MB, cheap to duplicate)
- Exposes only the CLI tools it contains
- Runs independently with its own configuration
- Can be scaled horizontally

## Why This Approach?

**Isolation** - Heavy tools (wkhtmltopdf, tesseract) run in separate containers. A crash or resource exhaustion in one doesn't affect others.

**Modularity** - Compose only the tools you need. Don't need OCR? Don't deploy the ocr-tools container.

**Scaling** - Scale bottleneck tools independently. If PDF generation is slow, run 3 instances of pdf-tools while keeping 1 instance of base-tools.

**Security** - No Docker socket mounting. No inter-container trust. Each container only knows about its local commands.

## Example Setup

### docker-compose.yml

```yaml
services:
  # Reverse proxy - routes requests to appropriate tool containers
  proxy:
    image: nginx:alpine
    ports:
      - "8080:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - base-tools
      - pdf-tools

  # Base tools (lightweight, fast)
  base-tools:
    image: http2cli-base:latest
    build:
      context: ../..
      dockerfile: Dockerfile.dev
    volumes:
      - ./configs/base-tools.yaml:/etc/http2cli/config.yaml:ro
    user: "1000:1000"
    read_only: true
    security_opt:
      - no-new-privileges:true
    deploy:
      replicas: 1

  # PDF tools (heavy, slow - may need scaling)
  pdf-tools:
    image: http2cli-pdf:latest
    build:
      context: .
      dockerfile: Dockerfile.pdf
    volumes:
      - ./configs/pdf-tools.yaml:/etc/http2cli/config.yaml:ro
    user: "1000:1000"
    read_only: true
    security_opt:
      - no-new-privileges:true
    deploy:
      replicas: 2  # Run 2 instances for load balancing
```

### nginx.conf

```nginx
events {
    worker_connections 1024;
}

http {
    # Upstreams - nginx automatically load balances across replicas
    upstream base_tools {
        server base-tools:8080;
    }

    upstream pdf_tools {
        server pdf-tools:8080;
    }

    server {
        listen 80;

        # Timeouts for long-running operations
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;

        # Base tools endpoints
        location /cat {
            proxy_pass http://base_tools;
        }

        location /date {
            proxy_pass http://base_tools;
        }

        location /rev {
            proxy_pass http://base_tools;
        }

        # PDF tools endpoints (note trailing slashes for path rewriting)
        location /pdf/ {
            proxy_pass http://pdf_tools/;
        }

        # Health check (pick any backend)
        location /health {
            proxy_pass http://base_tools;
        }

        # API discovery
        location /api/tools {
            # This would need aggregation logic for multiple backends
            # For simplicity, just return base tools
            proxy_pass http://base_tools;
        }
    }
}
```

### Tool Container Dockerfile

Example for PDF tools using wkhtmltopdf:

```dockerfile
# Dockerfile.pdf
FROM surnet/alpine-wkhtmltopdf:3.18.0-0.12.6-small

# Copy http2cli binary from the build
COPY --from=http2cli:latest /http2cli /usr/local/bin/http2cli

EXPOSE 8080
USER 1000:1000

CMD ["http2cli", "--config", "/etc/http2cli/config.yaml"]
```

### Per-Container Configuration

**configs/base-tools.yaml:**
```yaml
server:
  host: "0.0.0.0"
  port: 8080

security:
  allowed_commands:
    - /bin/cat
    - /bin/date
    - /bin/rev
  command_timeout: 30s
  max_concurrent_commands: 10

tools:
  - name: cat
    endpoint: /cat
    method: POST
    command: /bin/cat
    input:
      type: stdin
      form_field: file
    output:
      type: stdout
      content_type: application/octet-stream

  - name: date
    endpoint: /date
    method: GET
    command: /bin/date
    input:
      type: none
    output:
      type: stdout
      content_type: text/plain

  - name: rev
    endpoint: /rev
    method: POST
    command: /bin/rev
    input:
      type: stdin
      form_field: text
    output:
      type: stdout
      content_type: text/plain
```

**configs/pdf-tools.yaml:**
```yaml
server:
  host: "0.0.0.0"
  port: 8080

security:
  allowed_commands:
    - /usr/bin/wkhtmltopdf
  command_timeout: 120s
  max_concurrent_commands: 5  # Limit concurrent PDF generations

tools:
  - name: html-to-pdf
    endpoint: /convert
    method: POST
    command: /usr/bin/wkhtmltopdf
    timeout: 120s
    input:
      type: stdin
      form_field: html
    output:
      type: stdout
      content_type: application/pdf
    arguments:
      static:
        - "-"   # Read from stdin
        - "-"   # Write to stdout
      mapped:
        - name: page_size
          flag: "--page-size"
          source: query
          type: enum
          values: [A4, Letter, Legal]
          default: A4
```

## Scaling Operations

### Scale a specific service

```bash
# Scale PDF tools to 3 instances
docker compose up -d --scale pdf-tools=3

# Check running instances
docker compose ps
```

### Monitor health

```bash
# Check all container health
docker compose ps

# View logs for specific service
docker compose logs -f pdf-tools
```

### Rolling updates

```bash
# Rebuild and update PDF tools without downtime
docker compose build pdf-tools
docker compose up -d --no-deps pdf-tools
```

## Production Considerations

### Load Balancing

Nginx automatically load balances across container replicas. For more advanced load balancing:

```nginx
upstream pdf_tools {
    least_conn;  # Send to least busy instance
    server pdf-tools:8080;
}
```

### Health Checks

Add container health checks for automatic restart:

```yaml
pdf-tools:
  healthcheck:
    test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 10s
```

### Resource Limits

Prevent runaway containers:

```yaml
pdf-tools:
  deploy:
    resources:
      limits:
        cpus: '2'
        memory: 2G
      reservations:
        cpus: '0.5'
        memory: 512M
```

### Kubernetes

This architecture maps directly to Kubernetes:
- Each tool container becomes a Deployment
- nginx becomes an Ingress or Service mesh
- Scaling is handled by HPA (Horizontal Pod Autoscaler)

## Tradeoffs

| Aspect | Single Container | Multi-Container |
|--------|------------------|-----------------|
| Simplicity | Simpler | More complex |
| Isolation | All tools share resources | Tools isolated |
| Scaling | Scale everything together | Scale tools independently |
| Image size | One large image | Multiple smaller images |
| Startup time | One container | Multiple containers |
| Memory | Shared binary | Duplicated binary (~5MB each) |

Choose multi-container when:
- You have heavy tools that need isolation
- You need to scale specific tools independently
- Different tools have different resource requirements
- You want to deploy only specific tools to specific environments
