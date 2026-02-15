package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/tokencount"
	"gopkg.in/yaml.v3"
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
				// From here on, just copy everything as-is to content
			} else {
				if _, err := frontMatterBytes.WriteString(line + "\n"); err != nil {
					return Markdown[T]{}, fmt.Errorf("failed to write frontmatter: %w", err)
				}
			}
		case 2: // Scanning content - copy everything as-is
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

	return Markdown[T]{
		FrontMatter: *frontMatter,
		Content:     content.String(),
		Tokens:      tokencount.EstimateTokens(content.String()),
	}, nil
}
