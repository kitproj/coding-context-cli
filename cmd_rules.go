package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// runRules prints all default agent rules to stdout
func runRules(ctx context.Context, agentRules map[Agent][]RulePath, args []string) error {
	// Define flags for rules command
	var includes selectorMap
	var excludes selectorMap
	rulesFlags := flag.NewFlagSet("rules", flag.ExitOnError)
	rulesFlags.Var(&includes, "s", "Include rules with matching frontmatter (key=value)")
	rulesFlags.Var(&excludes, "S", "Exclude rules with matching frontmatter (key=value)")

	if err := rulesFlags.Parse(args); err != nil {
		return err
	}

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

			// Check if file matches include and exclude selectors
			if !includes.matchesIncludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (does not match include selectors): %s\n", filePath)
				return nil
			}
			if !excludes.matchesExcludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (matches exclude selectors): %s\n", filePath)
				return nil
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
