---
task_name: example-mcp-arbitrary-fields
agent: cursor
mcp_server: filesystem
---

# Example Task with MCP Server Field

This task demonstrates the simplified MCP server field.

## The `mcp_server` Field

Instead of a complex map of server configurations, the `mcp_server` field is now a simple string that specifies the name of the MCP server to use. The name typically matches the filename or task name.

**Example:**
```yaml
---
mcp_server: filesystem
---
```

## Why Simplify?

Previously, the `mcp_servers` field was a complex map with detailed configurations:

```yaml
mcp_servers:
  filesystem:
    type: stdio
    command: filesystem
  git:
    type: stdio
    command: git
```

This was overly complex for most use cases. The new simplified format just specifies the server name:

```yaml
mcp_server: filesystem
```

## How It Works

The `mcp_server` field is a **standard frontmatter field** that provides metadata about which MCP server should be used for the task. It does not act as a selector and does not filter rules.

## Example Usage

The example above shows a task that uses the `filesystem` MCP server. This is just a name - the actual configuration of the server is handled elsewhere (typically in your AI agent's configuration).

All fields are preserved when the configuration is parsed and appear in the task frontmatter output.
