package codingcontext

import (
	"fmt"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/skills"
)

// Result holds the assembled context from running a task
type Result struct {
	Rules  []markdown.Markdown[markdown.RuleFrontMatter] // List of included rule files
	Task   markdown.Markdown[markdown.TaskFrontMatter]   // Task file with frontmatter and content
	Skills skills.AvailableSkills                        // List of discovered skills (metadata only)
	Tokens int                                           // Total token count
	Agent  Agent                                         // The agent used (from task or -a flag)
	Prompt string                                        // Combined prompt: all rules and task content
}

// MCPServers returns all MCP server configurations from rules as a map.
// Each rule can specify one MCP server configuration.
// Returns a map from rule ID to MCP server configuration.
// Empty/zero-value MCP server configurations are filtered out.
// Rules without an ID field are included with a generated key "rule-<index>".
func (r *Result) MCPServers() map[string]mcp.MCPServerConfig {
	servers := make(map[string]mcp.MCPServerConfig)

	// Add server from each rule, filtering out empty configs
	for i, rule := range r.Rules {
		server := rule.FrontMatter.MCPServer
		// Skip empty MCP server configs (no command and no URL means empty)
		if server.Command != "" || server.URL != "" {
			// Use the rule's ID as the key, or generate one if not present
			key := rule.FrontMatter.ID
			if key == "" {
				key = fmt.Sprintf("rule-%d", i)
			}
			servers[key] = server
		}
	}

	return servers
}
