package codingcontext

// Markdown represents a markdown file with frontmatter and content
type Markdown struct {
	Path        string      // Path to the markdown file
	FrontMatter FrontMatter // Parsed YAML frontmatter
	Content     string      // Expanded content of the markdown
	Tokens      int         // Estimated token count
}

// Result holds the assembled context from running a task
type Result struct {
	Rules []Markdown // List of included rule files
	Task  Markdown   // Task file with frontmatter and content
}
