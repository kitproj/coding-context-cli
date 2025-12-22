package codingcontext

import (
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

// Result holds the assembled context from running a task
type Result struct {
	Rules  []markdown.Markdown[markdown.RuleFrontMatter] // List of included rule files
	Task   markdown.Markdown[markdown.TaskFrontMatter]   // Task file with frontmatter and content
	Tokens int                                           // Total token count
	Agent  Agent                                         // The agent used (from task or -a flag)
}

// MCPServers returns all MCP server configurations from rules.
// Each rule can specify one MCP server configuration.
// Returns a slice of all configured MCP servers from rules only.
func (r *Result) MCPServers() []mcp.MCPServerConfig {
	var servers []mcp.MCPServerConfig

	// Add server from each rule
	for _, rule := range r.Rules {
		servers = append(servers, rule.FrontMatter.MCPServer)
	}

	return servers
}
