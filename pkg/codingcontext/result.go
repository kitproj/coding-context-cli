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

// MCPServer returns the MCP server name from the task.
// If the task doesn't specify an MCP server, returns an empty string.
// Note: MCP servers specified in rules are intentionally ignored - only the task's
// MCP server is used. This ensures a single, clear source of truth for the MCP server.
func (r *Result) MCPServer() string {
	// Return the MCP server from task
	if r.Task.FrontMatter.MCPServer != "" {
		return r.Task.FrontMatter.MCPServer
	}

	return ""
}
