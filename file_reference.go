package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// fileReferencePattern matches @filepath patterns in markdown
// Matches: @path/to/file.ext (stops at whitespace or punctuation except /, -, _, and .)
var fileReferencePattern = regexp.MustCompile(`@([a-zA-Z0-9_./\-]+[a-zA-Z0-9_/\-])`)

// expandFileReferences expands file references in the content.
// A file reference is denoted by @filepath (e.g., @src/components/Button.tsx).
// The file content is read and included inline in the output.
func expandFileReferences(content string, workDir string) (string, error) {
	// Find all file references
	matches := fileReferencePattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	// Process matches in reverse order to maintain correct string positions
	result := content
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		// match[0], match[1] are the full match bounds
		// match[2], match[3] are the capture group bounds (the filepath)
		fullMatchStart := match[0]
		fullMatchEnd := match[1]
		filepathStart := match[2]
		filepathEnd := match[3]

		filepath := content[filepathStart:filepathEnd]

		// Read the referenced file
		fileContent, err := readReferencedFile(filepath, workDir)
		if err != nil {
			return "", fmt.Errorf("failed to read referenced file %s: %w", filepath, err)
		}

		// Format the file content as a code block with the file path
		replacement := formatFileContent(filepath, fileContent)

		// Replace the reference with the file content
		result = result[:fullMatchStart] + replacement + result[fullMatchEnd:]
	}

	return result, nil
}

// readReferencedFile reads the content of a file referenced by @filepath
func readReferencedFile(path string, workDir string) (string, error) {
	// If path is relative, resolve it relative to workDir
	var fullPath string
	if filepath.IsAbs(path) {
		fullPath = path
	} else {
		fullPath = filepath.Join(workDir, path)
	}

	// Read the file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// formatFileContent formats the file content for inclusion in the output
func formatFileContent(path string, content string) string {
	// Determine the file extension for syntax highlighting
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		ext = "text"
	}

	// Format as a markdown code block with the file path as a comment
	return fmt.Sprintf("```%s\n# File: %s\n%s```", ext, path, content)
}
