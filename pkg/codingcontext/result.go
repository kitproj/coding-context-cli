package codingcontext

import (
	"maps"

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

// MCPServers returns all MCP servers from both rules and the task.
// Servers from the task take precedence over servers from rules.
// If multiple rules define the same server name, the behavior is non-deterministic.
func (r *Result) MCPServers() mcp.MCPServerConfigs {
	servers := make(mcp.MCPServerConfigs)

	// Add servers from rules first (so task can override)
	for _, rule := range r.Rules {
		if rule.FrontMatter.MCPServers != nil {
			maps.Copy(servers, rule.FrontMatter.MCPServers)
		}
	}

	// Add servers from task (overriding any from rules)
	if r.Task.FrontMatter.MCPServers != nil {
		maps.Copy(servers, r.Task.FrontMatter.MCPServers)
	}

	return servers
}
