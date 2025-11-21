package codingcontext

// Config represents the standardized JSON structure for MCP configurations.
// Compatible with:
// - Claude Desktop: 'claude_desktop_config.json'
// - Claude Code:    '.mcp.json'
// - Cursor:         'mcp.json'
type Config struct {
	// MCPServers maps server names to their configuration.
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}
