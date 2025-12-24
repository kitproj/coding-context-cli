package codingcontext

import (
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

// Result holds the assembled context from running a task
type Result struct {
	Rules  []markdown.Markdown[markdown.RuleFrontMatter]  // List of included rule files
	Skills []markdown.Markdown[markdown.SkillFrontMatter] // List of included skill files
	Task   markdown.Markdown[markdown.TaskFrontMatter]    // Task file with frontmatter and content
	Tokens int                                            // Total token count
	Agent  Agent                                          // The agent used (from task or -a flag)
	Prompt string                                         // Combined prompt: all rules and task content
}

// MCPServers returns all MCP server configurations from rules.
// Each rule can specify one MCP server configuration.
// Returns a slice of all configured MCP servers from rules only.
// Empty/zero-value MCP server configurations are filtered out.
func (r *Result) MCPServers() []mcp.MCPServerConfig {
	var servers []mcp.MCPServerConfig

	// Add server from each rule, filtering out empty configs
	for _, rule := range r.Rules {
		server := rule.FrontMatter.MCPServer
		// Skip empty MCP server configs (no command and no URL means empty)
		if server.Command != "" || server.URL != "" {
			servers = append(servers, server)
		}
	}

	return servers
}
