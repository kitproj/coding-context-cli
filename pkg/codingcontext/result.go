package codingcontext

import (
	"fmt"
	"path/filepath"
	"strings"

	yaml "github.com/goccy/go-yaml"
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

// ParseFrontmatter unmarshals the frontmatter into the provided struct.
// The target parameter should be a pointer to a struct with yaml tags.
// Returns an error if the frontmatter cannot be unmarshaled into the target.
//
// Example:
//
//	type TaskMeta struct {
//	    TaskName string   `yaml:"task_name"`
//	    Resume   bool     `yaml:"resume"`
//	    Priority string   `yaml:"priority"`
//	}
//
//	var meta TaskMeta
//	if err := markdown.ParseFrontmatter(&meta); err != nil {
//	    // handle error
//	}
func (m *Markdown) ParseFrontmatter(target any) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if m.FrontMatter == nil {
		return fmt.Errorf("frontmatter is nil")
	}

	// Marshal the frontmatter map to YAML bytes, then unmarshal into target
	// This approach leverages the existing YAML library without adding new dependencies
	yamlBytes, err := yaml.Marshal(m.FrontMatter)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Unmarshal the YAML bytes into the target struct
	if err := yaml.Unmarshal(yamlBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal frontmatter into target: %w", err)
	}

	return nil
}

// Result holds the assembled context from running a task
type Result struct {
	Rules []Markdown // List of included rule files
	Task  Markdown   // Task file with frontmatter and content
}
