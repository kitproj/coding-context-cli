package codingcontext

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// fileReferencePattern matches @filepath patterns
// Matches: @path/to/file.ext, @./relative/path, @../parent/path, @file.txt
// Does not match: email@domain.com (word boundary before @ prevents email matching)
// Must have: path separator (/), relative prefix (./ or ../), or be a file with extension
// Trailing punctuation is handled in the expansion logic
var fileReferencePattern = regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])@((?:[a-zA-Z]:)?(?:[./][^\s<>"|*?\n]+|[a-zA-Z0-9_-]+/[^\s<>"|*?\n]+|[a-zA-Z0-9_][a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]+))`)

// expandFileReferences replaces @filepath references with the actual file content
// Returns the expanded content and any error encountered
func expandFileReferences(content string, baseDir string) (string, error) {
	var expandErr error

	// Find all matches first
	matches := fileReferencePattern.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	// Build the result by replacing each match
	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		// match[0] is start of full match, match[1] is end of full match
		// match[2] is start of captured filepath, match[3] is end of captured filepath

		// Write everything before this match
		result.WriteString(content[lastEnd:match[0]])

		// Extract the filepath (match[2]:match[3])
		filepath := content[match[2]:match[3]]

		// Strip trailing punctuation (period, comma, semicolon, etc.) from filepath
		// This handles cases like "@file.txt." where the period is sentence punctuation
		filepath = strings.TrimRight(filepath, ".,;:!?)")

		// Read the file content
		fileContent, err := readFileReference(filepath, baseDir)
		if err != nil {
			expandErr = err
			// Return original content on error
			return content, expandErr
		}

		// Find any prefix character before @ (start of match to @ position)
		prefix := ""
		atPos := strings.Index(content[match[0]:match[1]], "@")
		if atPos > 0 {
			prefix = content[match[0] : match[0]+atPos]
		}

		// Write the prefix and file content
		result.WriteString(prefix)
		result.WriteString(formatFileContent(filepath, fileContent))

		// Advance past the original match
		lastEnd = match[1]
	}

	// Write any remaining content after the last match
	result.WriteString(content[lastEnd:])

	return result.String(), nil
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
