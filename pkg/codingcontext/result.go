package codingcontext

// RuleContent represents a single rule file's content
type RuleContent struct {
	Path    string // Path to the rule file
	Content string // Expanded content of the rule
	Tokens  int    // Estimated token count for this rule
}

// Result holds the assembled context from running a task
type Result struct {
	Rules           []RuleContent // List of included rule files
	Task            string        // Expanded task content
	TaskFrontmatter FrontMatter   // Task frontmatter metadata
	TotalTokens     int           // Total estimated tokens across all content
}
