// Package lib provides a visitor pattern API for processing markdown files with YAML frontmatter.
package lib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kitproj/coding-context-cli/internal/parser"
)

// FrontMatter represents the YAML frontmatter of a markdown file
type FrontMatter map[string]any

// MarkdownVisitor is a function that processes a markdown file's frontmatter and content
type MarkdownVisitor func(path string, frontMatter FrontMatter, content string) error

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
		content, err := parser.ParseMarkdownFile(path, &frontMatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file %s: %w", path, err)
		}

		// Call the visitor
		if err := visitor(path, frontMatter, content); err != nil {
			return err
		}
	}

	return nil
}

// VisitPath processes markdown files in the given path (file or directory).
// If the path is a file, it processes that single file.
// If the path is a directory, it walks the directory recursively processing all .md and .mdc files.
// It stops on the first error.
func VisitPath(path string, visitor MarkdownVisitor) error {
	// Check if the path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Path doesn't exist, skip silently
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	// If it's a file, process it directly
	if !info.IsDir() {
		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".mdc" {
			return nil
		}

		var frontMatter FrontMatter
		content, err := parser.ParseMarkdownFile(path, &frontMatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file %s: %w", path, err)
		}

		return visitor(path, frontMatter, content)
	}

	// If it's a directory, walk it
	return filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			return nil
		}

		// Only process .md and .mdc files
		ext := filepath.Ext(filePath)
		if ext != ".md" && ext != ".mdc" {
			return nil
		}

		// Parse the markdown file
		var frontMatter FrontMatter
		content, err := parser.ParseMarkdownFile(filePath, &frontMatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file %s: %w", filePath, err)
		}

		// Call the visitor
		return visitor(filePath, frontMatter, content)
	})
}

// VisitPaths processes markdown files in multiple paths.
// Each path can be a file or directory.
// It stops on the first error.
func VisitPaths(paths []string, visitor MarkdownVisitor) error {
	for _, path := range paths {
		if err := VisitPath(path, visitor); err != nil {
			return err
		}
	}
	return nil
}
