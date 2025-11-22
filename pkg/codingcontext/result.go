package codingcontext

import (
	"path/filepath"
	"strings"
)

// Markdown represents a markdown file with frontmatter and content
type Markdown[T FrontMatter] struct {
	FrontMatter T      // Parsed YAML frontmatter
	Content     string // Expanded content of the markdown
	Tokens      int    // Estimated token count
}

// NewMarkdown creates a new Markdown with the given frontmatter
func NewMarkdown[T FrontMatter](t T) Markdown[T] {
	return Markdown[T]{
		FrontMatter: t,
	}
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
