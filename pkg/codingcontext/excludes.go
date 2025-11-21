package codingcontext

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CLIExcludes stores CLI names to exclude
type CLIExcludes map[string]bool

// String implements the fmt.Stringer interface for CLIExcludes
func (e *CLIExcludes) String() string {
	if *e == nil {
		return ""
	}
	var names []string
	for name := range *e {
		names = append(names, name)
	}
	return strings.Join(names, ",")
}

// Set implements the flag.Value interface for CLIExcludes
func (e *CLIExcludes) Set(value string) error {
	if *e == nil {
		*e = make(CLIExcludes)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("CLI name cannot be empty")
	}
	// Normalize to lowercase for case-insensitive matching
	(*e)[strings.ToLower(value)] = true
	return nil
}

// ShouldExcludePath returns true if the given path should be excluded based on CLI exclusions
func (e *CLIExcludes) ShouldExcludePath(path string) bool {
	if *e == nil || len(*e) == 0 {
		return false
	}

	// Normalize path separators
	normalizedPath := filepath.ToSlash(path)

	// Check each CLI exclusion
	for cliName := range *e {
		patterns := cliPathPatterns[cliName]
		for _, pattern := range patterns {
			if strings.Contains(normalizedPath, pattern) {
				return true
			}
		}
	}

	return false
}

// cliPathPatterns maps CLI names to their associated path patterns
// This maps agent CLI names to the directory/file patterns they use
var cliPathPatterns = map[string][]string{
	"cursor": {
		".cursor/",
		".cursorrules",
	},
	"opencode": {
		".opencode/",
	},
	"copilot": {
		".github/copilot-instructions.md",
		".github/agents",
	},
	"claude": {
		".claude/",
		"CLAUDE.md",
		"CLAUDE.local.md",
	},
	"gemini": {
		".gemini/",
		"GEMINI.md",
	},
	"augment": {
		".augment/",
	},
	"windsurf": {
		".windsurf/",
		".windsurfrules",
	},
	"codex": {
		".codex/",
	},
}
