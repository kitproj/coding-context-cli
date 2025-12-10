package codingcontext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readFileReference reads a file from disk, resolving relative paths from baseDir
// This function is used by the parameter expansion mechanism when encountering ${file:path} references
func readFileReference(filePath string, baseDir string) (string, error) {
	// Resolve the path relative to baseDir if it's not absolute
	var fullPath string
	if filepath.IsAbs(filePath) {
		fullPath = filePath
	} else {
		fullPath = filepath.Join(baseDir, filePath)
	}

	// Read the file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file reference %s: %w", filePath, err)
	}

	return string(content), nil
}

// formatFileContent formats file content with a header showing the filename
func formatFileContent(filePath string, content string) string {
	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString("File: ")
	sb.WriteString(filePath)
	sb.WriteString("\n")
	sb.WriteString("```\n")
	sb.WriteString(content)
	// Ensure content ends with newline before closing backticks
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("```\n\n")
	return sb.String()
}
