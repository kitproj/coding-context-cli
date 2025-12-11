package codingcontext

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// expand performs all types of expansion on the content in a single pass:
// 1. Parameter expansion: ${param_name}
// 2. Command expansion: !`command`
// 3. Path expansion: @path
// SECURITY: Processes rune-by-rune to prevent injection attacks where expanded
// content contains further expansion sequences (e.g., command output with ${param}).
func expand(content string, params map[string]string, logger *slog.Logger) string {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	var result strings.Builder
	runes := []rune(content)
	i := 0

	for i < len(runes) {
		// Check for parameter expansion: ${...}
		if i+2 < len(runes) && runes[i] == '$' && runes[i+1] == '{' {
			// Find the closing }
			end := i + 2
			for end < len(runes) && runes[end] != '}' {
				end++
			}
			if end < len(runes) {
				// Extract parameter name
				paramName := string(runes[i+2 : end])
				if val, ok := params[paramName]; ok {
					result.WriteString(val)
				} else {
					logger.Warn("parameter not found", "param", paramName)
					result.WriteString(string(runes[i : end+1]))
				}
				i = end + 1
				continue
			}
		}

		// Check for command expansion: !`...`
		if i+2 < len(runes) && runes[i] == '!' && runes[i+1] == '`' {
			// Find the closing `
			end := i + 2
			for end < len(runes) && runes[end] != '`' {
				end++
			}
			if end < len(runes) {
				// Extract command
				command := string(runes[i+2 : end])
				cmd := exec.Command("sh", "-c", command)
				output, err := cmd.CombinedOutput()
				if err != nil {
					logger.Warn("command expansion failed", "command", command, "error", err)
					// Return the original !`command` if command fails
					result.WriteString(string(runes[i : end+1]))
				} else {
					// Write command output (trimming trailing newline)
					result.WriteString(strings.TrimSuffix(string(output), "\n"))
				}
				i = end + 1
				continue
			}
		}

		// Check for path expansion: @path
		if runes[i] == '@' && (i == 0 || isWhitespaceRune(runes[i-1])) {
			// Found potential path expansion at start or after whitespace
			pathStart := i + 1
			pathEnd := pathStart

			// Scan for the end of the path (whitespace or end of string)
			// Handle escaped spaces
			for pathEnd < len(runes) {
				if runes[pathEnd] == '\\' && pathEnd+1 < len(runes) && runes[pathEnd+1] == ' ' {
					// Escaped space, skip both characters
					pathEnd += 2
				} else if isWhitespaceRune(runes[pathEnd]) {
					// Unescaped whitespace marks end of path
					break
				} else {
					pathEnd++
				}
			}

			if pathEnd > pathStart {
				// Extract and unescape the path
				path := unescapePath(string(runes[pathStart:pathEnd]))

				// Validate the path
				if err := validatePath(path); err != nil {
					logger.Warn("path validation failed", "path", path, "error", err)
					// Return the original @path if validation fails
					result.WriteString(string(runes[i:pathEnd]))
					i = pathEnd
					continue
				}

				// Read the file
				fileContent, err := os.ReadFile(path)
				if err != nil {
					logger.Warn("path expansion failed", "path", path, "error", err)
					// Return the original @path if file doesn't exist
					result.WriteString(string(runes[i:pathEnd]))
				} else {
					// Expand to file content
					result.Write(fileContent)
				}

				i = pathEnd
				continue
			}
		}

		// No expansion found, write the current rune
		result.WriteRune(runes[i])
		i++
	}

	return result.String()
}

// isWhitespaceRune checks if a rune is whitespace (space, tab, newline, carriage return)
func isWhitespaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// unescapePath removes escape sequences from a path (specifically \<space>)
func unescapePath(path string) string {
	return strings.ReplaceAll(path, "\\ ", " ")
}

// validatePath validates a file path for basic safety checks.
// Note: This tool is designed to work with user-created markdown files in their
// workspace and grants read access to files the user can read. The primary
// defense is that users should only use trusted markdown files.
func validatePath(path string) error {
	// Check for null bytes which are never valid in file paths
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	// We intentionally allow paths with .. components as they may be
	// legitimate references to files in parent directories within the
	// user's workspace. The security model is that users should only
	// use trusted markdown files, similar to running trusted shell scripts.

	return nil
}
