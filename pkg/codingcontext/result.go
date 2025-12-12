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

// MCPServers returns all MCP servers from both rules and the task.
// Servers from the task take precedence over servers from rules.
// If multiple rules define the same server name, the behavior is non-deterministic.
func (r *Result) MCPServers() MCPServerConfigs {
	servers := make(MCPServerConfigs)

	// Add servers from rules first (so task can override)
	for _, rule := range r.Rules {
		if rule.FrontMatter.MCPServers != nil {
			for name, config := range rule.FrontMatter.MCPServers {
				servers[name] = config
			}
		}
	}

	// Add servers from task (overriding any from rules)
	if r.Task.FrontMatter.MCPServers != nil {
		for name, config := range r.Task.FrontMatter.MCPServers {
			servers[name] = config
		}
	}

	return servers
}
