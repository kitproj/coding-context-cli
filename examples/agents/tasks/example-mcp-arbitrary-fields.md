---
task_name: example-mcp-arbitrary-fields
agent: cursor
mcp_server:
  command: python
  args: ["-m", "server"]
  env:
    PYTHON_PATH: /usr/bin/python3
  custom_config:
    host: localhost
    port: 5432
    ssl: true
    pool:
      min: 2
      max: 10
  monitoring:
    enabled: true
    metrics_port: 9090
---

# Example Task with MCP Server Configuration

This task demonstrates the MCP server configuration with arbitrary custom fields.

## The `mcp_server` Field

The `mcp_server` field specifies a single MCP server configuration with both standard and arbitrary custom fields. Each task or rule can specify one MCP server configuration.

**Standard fields:**
- `command`: The executable to run (e.g., "python", "npx", "docker")
- `args`: Array of command-line arguments
- `env`: Environment variables for the server process
- `type`: Connection protocol ("stdio", "http", "sse") - optional, defaults to stdio
- `url`: Endpoint URL for HTTP/SSE types
- `headers`: Custom HTTP headers for HTTP/SSE types

**Arbitrary custom fields:**
You can add any additional fields for your specific MCP server needs:
- `custom_config`: Nested configuration objects
- `monitoring`: Monitoring settings
- `cache_enabled`, `max_retries`, `timeout_seconds`, etc.

## Why Arbitrary Fields?

Different MCP servers may need different configuration options beyond the standard fields. Arbitrary fields allow you to:

1. **Add custom metadata**: Version info, regions, endpoints, etc.
2. **Configure behavior**: Caching, retry policies, timeouts, rate limits
3. **Include nested config**: Complex configuration objects specific to your server
4. **Future-proof**: Add new fields without changing the schema

## How It Works

The `MCPServerConfig` struct includes a `Content` field that captures all fields from YAML/JSON:

```go
type MCPServerConfig struct {
    // Standard fields
    Type    TransportType
    Command string
    Args    []string
    Env     map[string]string
    URL     string
    Headers map[string]string
    
    // Arbitrary fields via inline map
    Content map[string]any `yaml:",inline"`
}
```

All fields (both standard and custom) are preserved when the configuration is parsed and can be accessed via the struct fields or the `Content` map.
