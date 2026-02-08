# http2cli

A lightweight HTTP wrapper for CLI tools. Expose any command-line tool as an HTTP API with just a YAML configuration file.

## Quick Start

```bash
# Build
go build -o http2cli ./cmd/http2cli

# Run
./http2cli --config configs/config.yaml
```

## Usage

```bash
# Get current date
curl "http://localhost:8080/date?format=%Y-%m-%d"

# Pass file through cat
curl -F "file=@input.txt" http://localhost:8080/cat

# Reverse text
echo "hello" | curl -F "text=@-" http://localhost:8080/rev
```

## Docker

```dockerfile
FROM alpine:3.19
COPY --from=http2cli:latest /http2cli /usr/local/bin/http2cli
COPY config.yaml /etc/http2cli/config.yaml
CMD ["http2cli", "--config", "/etc/http2cli/config.yaml"]
```

## Documentation

See [docs/](docs/) for detailed documentation:
- [Configuration Guide](docs/configuration.md)
- [API Reference](docs/api.md)
- [Security](docs/security.md)
