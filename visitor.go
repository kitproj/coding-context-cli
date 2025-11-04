package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// FrontMatter represents the YAML frontmatter of a markdown file
type FrontMatter map[string]any

// MarkdownVisitor is a function that processes a markdown file's frontmatter and content
type MarkdownVisitor func(frontMatter FrontMatter, content string) error

// Visit parses markdown files matching the given pattern and calls the visitor for each file.
// It stops on the first error.
// The pattern follows the same rules as filepath.Glob.
func Visit(pattern string, visitor MarkdownVisitor) error {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob pattern %s: %w", pattern, err)
	}

	for _, path := range matches {
		// Check if it's a file
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", path, err)
		}
		if info.IsDir() {
			continue
		}

		// Parse the markdown file
		var frontMatter FrontMatter
		content, err := parseMarkdownFile(path, &frontMatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file %s: %w", path, err)
		}

		// Call the visitor
		if err := visitor(frontMatter, content); err != nil {
			return err
		}
	}

	return nil
}
