package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// runPrompt finds and prints a task prompt
func runPrompt(ctx context.Context, args []string) error {
	// Define flags for prompt command
	var params paramMap
	var includes selectorMap
	var excludes selectorMap
	promptFlags := flag.NewFlagSet("prompt", flag.ExitOnError)
	promptFlags.Var(&params, "p", "Template parameter (key=value)")
	promptFlags.Var(&includes, "s", "Include rules with matching frontmatter (key=value)")
	promptFlags.Var(&excludes, "S", "Exclude rules with matching frontmatter (key=value)")
	
	if err := promptFlags.Parse(args); err != nil {
		return err
	}

	promptArgs := promptFlags.Args()
	if len(promptArgs) < 1 {
		return fmt.Errorf("usage: coding-context prompt [-p key=value] [-s key=value] [-S key=value] <name>")
	}

	promptName := promptArgs[0]
	
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

	// Search for prompt file in task paths
	var promptPath string

	for _, taskPath := range allTaskPaths {
		// Check if directory exists
		if _, err := os.Stat(taskPath); os.IsNotExist(err) {
			continue
		}

		// Check for prompt file
		candidatePath := filepath.Join(taskPath, promptName+".md")
		if _, err := os.Stat(candidatePath); err == nil {
			promptPath = candidatePath
			break
		}
	}

	if promptPath == "" {
		return fmt.Errorf("prompt file not found for: %s", promptName)
	}

	// Read the prompt file
	var frontmatter map[string]string
	content, err := parseMarkdownFile(promptPath, &frontmatter)
	if err != nil {
		return fmt.Errorf("failed to parse prompt file: %w", err)
	}

	// Check if file matches include and exclude selectors
	if !includes.matchesIncludes(frontmatter) {
		return fmt.Errorf("prompt file does not match include selectors: %s", promptPath)
	}
	if !excludes.matchesExcludes(frontmatter) {
		return fmt.Errorf("prompt file matches exclude selectors: %s", promptPath)
	}

	// Template the prompt using os.Expand
	templated := os.Expand(content, func(key string) string {
		if val, ok := params[key]; ok {
			return val
		}
		// Return original placeholder if not found
		return fmt.Sprintf("${%s}", key)
	})

	// Estimate tokens and log to stderr
	tokens := estimateTokens(templated)
	fmt.Fprintf(os.Stderr, "Using prompt file: %s (~%d tokens)\n", promptPath, tokens)

	// Print to stdout
	fmt.Fprint(os.Stdout, templated)

	return nil
}
