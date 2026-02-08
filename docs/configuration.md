# Configuration Guide

http2cli uses YAML for configuration. The configuration file defines the server settings, security constraints, and the CLI tools to expose.

## Configuration File Structure

```yaml
server:
  # Server configuration

security:
  # Security settings

tools:
  # List of CLI tools to expose
```

## Server Configuration

```yaml
server:
  host: "0.0.0.0"       # Bind address
  port: 8080            # Listen port
  read_timeout: 30s     # HTTP read timeout
  write_timeout: 300s   # HTTP write timeout
  max_upload_size: 10MB # Maximum upload size (not yet implemented)
```

## Security Configuration

```yaml
security:
  # API key for authentication (empty = disabled)
  api_key: "your-secret-key"

  # Allowlist of commands that can be executed
  # Commands not in this list will be rejected
  allowed_commands:
    - /bin/cat
    - /bin/date
    - /usr/bin/rev

  # Default timeout for command execution
  command_timeout: 30s

  # Maximum concurrent command executions
  max_concurrent_commands: 10
```

## Tool Configuration

Each tool maps an HTTP endpoint to a CLI command.

```yaml
tools:
  - name: example           # Unique identifier
    description: "..."      # Human-readable description
    endpoint: /example      # HTTP endpoint path
    method: POST            # HTTP method (GET, POST)
    command: /usr/bin/cmd   # Path to executable (must be in allowed_commands)
    timeout: 60s            # Command timeout (optional, uses security.command_timeout if not set)

    input:                  # Input configuration
      type: stdin           # stdin, file, or none
      form_field: file      # Multipart form field name
      required: true        # Whether input is required

    output:                 # Output configuration
      type: stdout          # stdout or file
      content_type: text/plain  # Response content type
      filename: output.txt  # Download filename (optional)

    arguments:              # CLI argument configuration
      static:               # Arguments always included
        - "-flag"
      mapped:               # Arguments from HTTP parameters
        - name: param
          flag: "--param"
          source: query     # query, form, or header
          type: string      # string, int, bool, or enum
```

## Argument Types

### string

Free-form string value. Optionally validate with regex pattern.

```yaml
- name: format
  flag: "--format"
  source: query
  type: string
  pattern: "^[a-zA-Z]+$"  # Optional regex pattern
  default: "text"
```

### int

Integer value with optional min/max bounds.

```yaml
- name: count
  flag: "-n"
  source: query
  type: int
  min: 1
  max: 100
  default: "10"
```

### bool

Boolean flag. When true, the flag is added without a value.

```yaml
- name: verbose
  flag: "-v"
  source: query
  type: bool
```

Usage: `?verbose=true` adds `-v` to the command.

### enum

Value must be one of the allowed values.

```yaml
- name: output_format
  flag: "-o"
  source: query
  type: enum
  values: [json, xml, text]
  default: json
```

## Input Types

### stdin

File content is passed to the command's stdin.

```yaml
input:
  type: stdin
  form_field: file
```

### file

File is saved temporarily and path is passed to command. Use `{{input_file}}` placeholder in static arguments.

```yaml
input:
  type: file
  form_field: image

arguments:
  static:
    - "{{input_file}}"
    - "stdout"
```

### none

No file input. Used for commands that only need arguments.

```yaml
input:
  type: none
```

## Environment Variables

Configuration values can be overridden with environment variables:

- `HTTP2CLI_SERVER_HOST` - Server bind address
- `HTTP2CLI_SERVER_PORT` - Server port
- `HTTP2CLI_SECURITY_API_KEY` - API key for authentication

## Examples

See the [configs/config.yaml](../configs/config.yaml) file for a complete working example.
