// Package lib provides a visitor pattern API for processing markdown files with YAML frontmatter.
package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	yaml "go.yaml.in/yaml/v2"
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

// parseMarkdownFile parses the file into frontmatter and content
func parseMarkdownFile(path string, frontmatter any) (string, error) {
	fh, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
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
					return "", fmt.Errorf("failed to write content: %w", err)
				}
			}
		case 1: // Scanning frontmatter
			if line == "---" {
				state = 2 // End of frontmatter, start scanning content
			} else {
				if _, err := frontMatterBytes.WriteString(line + "\n"); err != nil {
					return "", fmt.Errorf("failed to write frontmatter: %w", err)
				}
			}
		case 2: // Scanning content
			if _, err := content.WriteString(line + "\n"); err != nil {
				return "", fmt.Errorf("failed to write content: %w", err)
			}
		}
	}

	if err := s.Err(); err != nil {
		return "", fmt.Errorf("failed to scan file: %w", err)
	}

	// Parse frontmatter if we collected any
	if frontMatterBytes.Len() > 0 {
		if err := yaml.Unmarshal(frontMatterBytes.Bytes(), frontmatter); err != nil {
			return "", fmt.Errorf("failed to unmarshal frontmatter: %w", err)
		}
	}

	return content.String(), nil
}
