package codingcontext

// Markdown represents a markdown file with frontmatter and content
type Markdown[T any] struct {
	FrontMatter T      // Parsed YAML frontmatter
	Content     string // Expanded content of the markdown
	Tokens      int    // Estimated token count
}

// TaskMarkdown is a Markdown with TaskFrontMatter
type TaskMarkdown = Markdown[TaskFrontMatter]

// RuleMarkdown is a Markdown with RuleFrontMatter
type RuleMarkdown = Markdown[RuleFrontMatter]

// Result holds the assembled context from running a task
type Result struct {
	Rules  []Markdown[RuleFrontMatter] // List of included rule files
	Task   Markdown[TaskFrontMatter]   // Task file with frontmatter and content
	Tokens int                         // Total token count
	Agent  Agent                       // The agent used (from task or -a flag)
}

// MCPServers returns all MCP server configurations from both rules and the task.
// Each rule and the task can specify one MCP server configuration.
// Returns a slice of all configured MCP servers.
func (r *Result) MCPServers() []MCPServerConfig {
	var servers []MCPServerConfig

	// Add server from each rule that has one
	for _, rule := range r.Rules {
		// Check if the MCPServer is not empty (has at least one field set)
		if rule.FrontMatter.MCPServer.Command != "" || rule.FrontMatter.MCPServer.URL != "" {
			servers = append(servers, rule.FrontMatter.MCPServer)
		}
	}

	// Add server from task if it has one
	if r.Task.FrontMatter.MCPServer.Command != "" || r.Task.FrontMatter.MCPServer.URL != "" {
		servers = append(servers, r.Task.FrontMatter.MCPServer)
	}

	return servers
}
