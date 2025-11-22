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
	Rules []Markdown[RuleFrontMatter] // List of included rule files
	Task  Markdown[TaskFrontMatter]   // Task file with frontmatter and content
}

// MCPServers returns all MCP servers from both rules and the task.
// Servers from the task are included first, followed by servers from rules.
// Duplicate servers may be present if the same server is specified in multiple places.
func (r *Result) MCPServers() []MCPServerConfig {
	var servers []MCPServerConfig

	// Add servers from task first
	if r.Task.FrontMatter.MCPServers != nil {
		servers = append(servers, r.Task.FrontMatter.MCPServers...)
	}

	// Add servers from all rules
	for _, rule := range r.Rules {
		if rule.FrontMatter.MCPServers != nil {
			servers = append(servers, rule.FrontMatter.MCPServers...)
		}
	}

	return servers
}
