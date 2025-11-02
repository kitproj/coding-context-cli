package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// runRules prints all default agent rules to stdout
func runRules(ctx context.Context, agentRules map[Agent][]RulePath, args []string) error {
	// Get the Default agent's rules
	rulePaths := agentRules[Default]

	var totalTokens int

	// Walk through all rule paths and collect content
	for _, rp := range rulePaths {
		path := rp.Source()

		// Skip if the path doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Only process .md files
			ext := filepath.Ext(filePath)
			if ext != ".md" {
				return nil
			}

			// Parse frontmatter
			var frontmatter map[string]string
			content, err := parseMarkdownFile(filePath, &frontmatter)
			if err != nil {
				return fmt.Errorf("failed to parse markdown file: %w", err)
			}

			// Estimate tokens
			tokens := estimateTokens(content)
			totalTokens += tokens

			// Log to stderr
			fmt.Fprintf(os.Stderr, "Including rule file: %s (~%d tokens)\n", filePath, tokens)

			// Print content to stdout
			fmt.Fprint(os.Stdout, content)
			fmt.Fprintln(os.Stdout)

			return nil
		})
		if err != nil {
			return err
		}
	}

	// Log total tokens to stderr
	fmt.Fprintf(os.Stderr, "Total estimated tokens: %d\n", totalTokens)

	return nil
}
