package codingcontext

import (
	"path/filepath"
	"strings"
)

// Markdown represents a markdown file with frontmatter and content
type Markdown[T any] struct {
	Path        string // Path to the markdown file
	FrontMatter T      // Parsed YAML frontmatter
	Content     string // Expanded content of the markdown
	Tokens      int    // Estimated token count
}

// TaskMarkdown is a Markdown with TaskFrontMatter
type TaskMarkdown = Markdown[TaskFrontMatter]

// RuleMarkdown is a Markdown with RuleFrontMatter
type RuleMarkdown = Markdown[RuleFrontMatter]

// BootstrapPath returns the path to the bootstrap script for this markdown file, if it exists.
// Returns empty string if the path is empty.
func (m *Markdown[T]) BootstrapPath() string {
	if m.Path == "" {
		return ""
	}
	ext := filepath.Ext(m.Path)
	baseNameWithoutExt := strings.TrimSuffix(m.Path, ext)
	return baseNameWithoutExt + "-bootstrap"
}

// Result holds the assembled context from running a task
type Result struct {
	Rules []Markdown[FrontMatter] // List of included rule files
	Task  Markdown[FrontMatter]   // Task file with frontmatter and content
}
