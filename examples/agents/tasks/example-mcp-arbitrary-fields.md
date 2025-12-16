---
task_name: example-mcp-arbitrary-fields
agent: cursor
mcp_servers:
  # Example with standard fields only
  filesystem:
    type: stdio
    command: filesystem
  
  # Example with standard fields plus arbitrary custom fields
  custom-database:
    type: stdio
    command: database-mcp
    args: ["--verbose"]
    # Arbitrary fields below
    cache_enabled: true
    max_cache_size: 1000
    connection_pool_size: 10
    
  # Example HTTP server with custom metadata
  api-server:
    type: http
    url: https://api.example.com
    headers:
      Authorization: Bearer token123
    # Arbitrary fields below
    api_version: v2
    rate_limit: 100
    timeout_seconds: 30
    retry_policy: exponential
    region: us-west-2
    
  # Example with nested custom configuration
  advanced-server:
    type: stdio
    command: python
    args: ["-m", "server"]
    env:
      PYTHON_PATH: /usr/bin/python3
    # Arbitrary nested fields below
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

# Example Task with Arbitrary MCP Server Fields

This task demonstrates the ability to add arbitrary fields to MCP server configurations, just like we can with FrontMatter.

## Why Arbitrary Fields?

Different MCP servers may need different configuration options beyond the standard fields (`type`, `command`, `args`, `env`, `url`, `headers`). Arbitrary fields allow you to:

1. **Add custom metadata**: Version info, regions, endpoints, etc.
2. **Configure behavior**: Caching, retry policies, timeouts, rate limits
3. **Include nested config**: Complex configuration objects specific to your server
4. **Future-proof**: Add new fields without changing the schema

## How It Works

The `MCPServerConfig` struct now includes a `Content` field (similar to `BaseFrontMatter`) that captures all fields from YAML/JSON:

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

## Example Usage

The examples above show:
- **Simple custom fields**: `cache_enabled`, `max_cache_size`
- **API configuration**: `api_version`, `rate_limit`, `timeout_seconds`
- **Nested objects**: `custom_config` with sub-fields like `host`, `port`, `ssl`
- **Multiple custom sections**: `custom_config` and `monitoring` as separate objects

All these fields are preserved when the configuration is parsed and can be accessed via the `Content` map.
