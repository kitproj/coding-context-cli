package codingcontext

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Expander handles content expansion for parameters, commands, and file paths
type Expander struct {
	params Params
	logger *slog.Logger
}

// NewExpander creates a new Expander with the given parameters and logger
func NewExpander(params Params, logger *slog.Logger) *Expander {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return &Expander{
		params: params,
		logger: logger,
	}
}

// Expand performs all types of expansion on the content:
// 1. Parameter expansion: ${param_name}
// 2. Command expansion: !`command`
// 3. Path expansion: @path
func (e *Expander) Expand(content string) string {
	// First expand commands, then paths, then parameters
	// This order allows commands and paths to generate parameter references
	content = e.expandCommands(content)
	content = e.expandPaths(content)
	content = e.expandParameters(content)
	return content
}

// expandParameters handles ${param_name} expansion
func (e *Expander) expandParameters(content string) string {
	return os.Expand(content, func(key string) string {
		if val, ok := e.params[key]; ok {
			return val
		}
		// Return original placeholder if not found and log warning
		e.logger.Warn("parameter not found", "param", key)
		return fmt.Sprintf("${%s}", key)
	})
}

// expandCommands handles !`command` expansion
func (e *Expander) expandCommands(content string) string {
	// Match !`...` where ... is the command
	// The backtick content can span multiple lines
	re := regexp.MustCompile("!`([^`]*)`")

	result := re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract command from !`command`
		command := match[2 : len(match)-1] // Remove !` and `

		// Execute the command using sh -c
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			e.logger.Warn("command expansion failed", "command", command, "error", err)
			// Return the original match if command fails
			return match
		}

		// Return the command output, trimming trailing newline if present
		return strings.TrimSuffix(string(output), "\n")
	})

	return result
}

// expandPaths handles @path expansion
func (e *Expander) expandPaths(content string) string {
	// Match @path where path is delimited by whitespace
	// Paths can have escaped spaces (\ )
	var result strings.Builder
	i := 0

	for i < len(content) {
		if content[i] == '@' && (i == 0 || isWhitespace(content[i-1])) {
			// Found potential path expansion at start or after whitespace
			pathStart := i + 1
			pathEnd := pathStart

			// Scan for the end of the path (whitespace or end of string)
			// Handle escaped spaces
			for pathEnd < len(content) {
				if content[pathEnd] == '\\' && pathEnd+1 < len(content) && content[pathEnd+1] == ' ' {
					// Escaped space, skip both characters
					pathEnd += 2
				} else if isWhitespace(content[pathEnd]) {
					// Unescaped whitespace marks end of path
					break
				} else {
					pathEnd++
				}
			}

			if pathEnd > pathStart {
				// Extract and unescape the path
				path := unescapePath(content[pathStart:pathEnd])

				// Read the file
				fileContent, err := os.ReadFile(path)
				if err != nil {
					e.logger.Warn("path expansion failed", "path", path, "error", err)
					// Return the original @path if file doesn't exist
					result.WriteString(content[i:pathEnd])
				} else {
					// Expand to file content
					result.Write(fileContent)
				}

				i = pathEnd
				continue
			}
		}

		result.WriteByte(content[i])
		i++
	}

	return result.String()
}

// isWhitespace checks if a byte is whitespace (space, tab, newline, carriage return)
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// unescapePath removes escape sequences from a path (specifically \<space>)
func unescapePath(path string) string {
	return strings.ReplaceAll(path, "\\ ", " ")
}
