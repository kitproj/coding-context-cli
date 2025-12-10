package codingcontext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// expandFileReferences replaces @path references with file content
// The @ symbol must be followed by a path, and the path ends at the first unescaped space or end of string
// Spaces in paths can be escaped with backslash: @src/My\ File.txt
func expandFileReferences(content string, baseDir string) string {
	var result strings.Builder
	i := 0

	for i < len(content) {
		// Look for @ symbol
		if content[i] == '@' {
			// Check if this could be a file reference (not an email)
			// It's a file reference if @ is at start or preceded by whitespace/punctuation
			isFileRef := i == 0 ||
				content[i-1] == ' ' || content[i-1] == '\t' || content[i-1] == '\n' ||
				content[i-1] == '(' || content[i-1] == '[' || content[i-1] == '{' ||
				content[i-1] == ',' || content[i-1] == '.' || content[i-1] == ';' ||
				content[i-1] == ':' || content[i-1] == '!'

			if isFileRef {
				// Extract the path - read until unescaped space or end of string
				start := i + 1
				pathEnd := start
				escaped := false

				for pathEnd < len(content) {
					if escaped {
						escaped = false
						pathEnd++
						continue
					}

					if content[pathEnd] == '\\' {
						escaped = true
						pathEnd++
						continue
					}

					// Stop at unescaped space or newline (these are the main delimiters)
					if content[pathEnd] == ' ' || content[pathEnd] == '\t' || content[pathEnd] == '\n' {
						break
					}

					// Stop at closing punctuation
					if content[pathEnd] == ')' || content[pathEnd] == ']' || content[pathEnd] == '}' {
						break
					}

					// Stop at comma followed by space (list separator)
					if content[pathEnd] == ',' && pathEnd+1 < len(content) &&
						(content[pathEnd+1] == ' ' || content[pathEnd+1] == '\t' || content[pathEnd+1] == '\n') {
						break
					}

					// Stop at period at end of content (sentence end)
					if content[pathEnd] == '.' && pathEnd+1 >= len(content) {
						break
					}

					// Stop at period followed by space or newline (sentence end)
					if content[pathEnd] == '.' && pathEnd+1 < len(content) &&
						(content[pathEnd+1] == ' ' || content[pathEnd+1] == '\t' || content[pathEnd+1] == '\n') {
						break
					}

					pathEnd++
				}

				if pathEnd > start {
					// Extract and unescape the path
					rawPath := content[start:pathEnd]
					filePath := unescapePath(rawPath)

					// Try to read the file
					fileContent, err := readFileReference(filePath, baseDir)
					if err == nil {
						// Successfully read file, write the formatted content
						result.WriteString(formatFileContent(filePath, fileContent))
						i = pathEnd
						continue
					}
					// If file read failed, keep the original @path
				}
			}
		}

		// Write current character and move to next
		result.WriteByte(content[i])
		i++
	}

	return result.String()
}

// unescapePath removes backslash escaping from a path string
func unescapePath(path string) string {
	var result strings.Builder
	escaped := false

	for i := 0; i < len(path); i++ {
		if escaped {
			result.WriteByte(path[i])
			escaped = false
		} else if path[i] == '\\' {
			escaped = true
		} else {
			result.WriteByte(path[i])
		}
	}

	return result.String()
}

// readFileReference reads a file from disk, resolving relative paths from baseDir
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
