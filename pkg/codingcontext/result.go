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
