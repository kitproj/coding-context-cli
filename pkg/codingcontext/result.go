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
// Servers from the task are included first, with rule servers added afterwards.
// If the same server name appears in multiple places, the task's version takes precedence
// over rule versions, and earlier rules take precedence over later rules.
func (r *Result) MCPServers() map[string]MCPServerConfig {
	servers := make(map[string]MCPServerConfig)

	// Add servers from rules first (so task can override)
	for i := len(r.Rules) - 1; i >= 0; i-- {
		rule := r.Rules[i]
		if rule.FrontMatter.MCPServers != nil {
			for name, config := range rule.FrontMatter.MCPServers {
				if _, exists := servers[name]; !exists {
					servers[name] = config
				}
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
