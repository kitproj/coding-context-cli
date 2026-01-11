---
rule_name: go-linting-standards
tool_name: golangci-lint
languages:
  - go
stage: implementation
mcp_server:
  type: stdio
  command: golangci-lint
  args:
    - run
    - --config
    - .golangci.yml
---

# Go Linting Standards with Tool Integration

This rule defines linting standards for Go code and provides integration with the `golangci-lint` tool.

## Linting Standards

- Run `golangci-lint` before committing code
- Fix all errors and warnings
- Use the project's `.golangci.yml` configuration
- Enable key linters: `gofmt`, `govet`, `staticcheck`, `errcheck`

## Tool Integration

This rule includes MCP server configuration for the `golangci-lint` tool. When used with compatible AI agents, the tool can be invoked directly to check code quality.

## Usage

The linter checks for:
- Code formatting issues
- Common bugs and errors
- Code complexity
- Unused code
- Security vulnerabilities

## Configuration

Customize linting rules in `.golangci.yml`:
```yaml
linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - errcheck
    - gosec
```
