package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Task prompt paths for the default agent
var taskPaths = []string{
	".agents/tasks",
	// User and system paths will be added dynamically
}

// runPrompt finds and prints a task prompt
func runPrompt(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: coding-context prompt <name>")
	}

	promptName := args[0]
	
	// Build full task paths list
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	allTaskPaths := []string{
		".agents/tasks",
		filepath.Join(homeDir, ".agents", "tasks"),
		"/etc/agents/tasks",
	}

	// Get parameters from remaining args
	params := make(map[string]string)
	for i := 1; i < len(args); i++ {
		// Parse key=value pairs
		if kv := args[i]; len(kv) > 0 {
			// Simple parsing - split on first =
			for j := 0; j < len(kv); j++ {
				if kv[j] == '=' {
					key := kv[:j]
					value := kv[j+1:]
					params[key] = value
					break
				}
			}
		}
	}

	// Search for prompt file in task paths
	var promptContent string
	var promptPath string
	var totalTokens int

	for _, taskPath := range allTaskPaths {
		// Check if directory exists
		if _, err := os.Stat(taskPath); os.IsNotExist(err) {
			continue
		}

		// Check for prompt file
		candidatePath := filepath.Join(taskPath, promptName+".md")
		if _, err := os.Stat(candidatePath); os.IsNotExist(err) {
			continue
		}

		// Found the prompt file
		var frontmatter map[string]string
		content, err := parseMarkdownFile(candidatePath, &frontmatter)
		if err != nil {
			return fmt.Errorf("failed to parse prompt file: %w", err)
		}

		promptContent = content
		promptPath = candidatePath
		break
	}

	if promptContent == "" {
		return fmt.Errorf("prompt file not found for: %s", promptName)
	}

	// Template the prompt using os.Expand
	templated := os.Expand(promptContent, func(key string) string {
		if val, ok := params[key]; ok {
			return val
		}
		// Return original placeholder if not found
		return fmt.Sprintf("${%s}", key)
	})

	// Estimate tokens
	totalTokens = estimateTokens(templated)

	// Log to stderr
	fmt.Fprintf(os.Stderr, "Using prompt file: %s (~%d tokens)\n", promptPath, totalTokens)

	// Print to stdout
	fmt.Fprint(os.Stdout, templated)

	return nil
}
