package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "github.com/goccy/go-yaml"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/tokencount"
)

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

// ParseMarkdownFile parses a markdown file into frontmatter and content
func ParseMarkdownFile[T any](path string, frontMatter *T) (Markdown[T], error) {
	fh, err := os.Open(path)
	if err != nil {
		return Markdown[T]{}, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer fh.Close()

	s := bufio.NewScanner(fh)

	var content bytes.Buffer
	var frontMatterBytes bytes.Buffer

	// State machine: 0 = unknown, 1 = scanning frontmatter, 2 = scanning content
	state := 0

	for s.Scan() {
		line := s.Text()

		switch state {
		case 0: // State unknown - first line
			if line == "---" {
				state = 1 // Start scanning frontmatter
			} else {
				state = 2 // No frontmatter, start scanning content
				if _, err := content.WriteString(line + "\n"); err != nil {
					return Markdown[T]{}, fmt.Errorf("failed to write content: %w", err)
				}
			}
		case 1: // Scanning frontmatter
			if line == "---" {
				state = 2 // End of frontmatter, start scanning content
			} else {
				if _, err := frontMatterBytes.WriteString(line + "\n"); err != nil {
					return Markdown[T]{}, fmt.Errorf("failed to write frontmatter: %w", err)
				}
			}
		case 2: // Scanning content
			if _, err := content.WriteString(line + "\n"); err != nil {
				return Markdown[T]{}, fmt.Errorf("failed to write content: %w", err)
			}
		}
	}

	if err := s.Err(); err != nil {
		return Markdown[T]{}, fmt.Errorf("failed to scan file %s: %w", path, err)
	}

	// Parse frontmatter if we collected any
	if frontMatterBytes.Len() > 0 {
		if err := yaml.Unmarshal(frontMatterBytes.Bytes(), frontMatter); err != nil {
			return Markdown[T]{}, fmt.Errorf("failed to unmarshal frontmatter in file %s: %w", path, err)
		}
	}

	// Default the Name field to filename without extension if not specified
	setDefaultName(frontMatter, path)

	return Markdown[T]{
		FrontMatter: *frontMatter,
		Content:     content.String(),
		Tokens:      tokencount.EstimateTokens(content.String()),
	}, nil
}

// setDefaultName sets the Name field to the filename without extension if not already set
func setDefaultName(frontMatter any, path string) {
	// Use type assertion to check if frontMatter has a Name field via BaseFrontMatter
	switch fm := frontMatter.(type) {
	case *TaskFrontMatter:
		if fm.Name == "" {
			fm.Name = getDefaultName(path)
		}
	case *RuleFrontMatter:
		if fm.Name == "" {
			fm.Name = getDefaultName(path)
		}
	case *CommandFrontMatter:
		if fm.Name == "" {
			fm.Name = getDefaultName(path)
		}
	case *SkillFrontMatter:
		// Skills already have a required Name field (shadows BaseFrontMatter.Name), don't override
		// The skill Name field is validated separately in skill discovery
	case *BaseFrontMatter:
		if fm.Name == "" {
			fm.Name = getDefaultName(path)
		}
	}
}

// getDefaultName extracts the filename without extension
func getDefaultName(path string) string {
	baseName := filepath.Base(path)
	ext := filepath.Ext(baseName)
	return strings.TrimSuffix(baseName, ext)
}
