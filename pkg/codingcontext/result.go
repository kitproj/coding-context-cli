package codingcontext

import (
	"path/filepath"
	"strings"
)

// Markdown represents a markdown file with frontmatter and content
type Markdown struct {
	Path        string      // Path to the markdown file
	FrontMatter FrontMatter // Parsed YAML frontmatter
	Content     string      // Expanded content of the markdown
	Tokens      int         // Estimated token count
}

// BootstrapPath returns the path to the bootstrap script for this markdown file, if it exists.
// Returns empty string if the path is empty.
func (m *Markdown) BootstrapPath() string {
	if m.Path == "" {
		return ""
	}
	ext := filepath.Ext(m.Path)
	baseNameWithoutExt := strings.TrimSuffix(m.Path, ext)
	return baseNameWithoutExt + "-bootstrap"
}

// Result holds the assembled context from running a task
type Result struct {
	Rules []Markdown // List of included rule files
	Task  Markdown   // Task file with frontmatter and content
}
