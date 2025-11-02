package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// runExport exports rules from the default agent to the specified agent
func runExport(ctx context.Context, agentRules map[Agent][]RulePath, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: coding-context export <agent>")
	}

	agentName := Agent(args[0])

	// Check if agent is valid and not Default
	if agentName == Default {
		return fmt.Errorf("cannot export to default agent")
	}

	targetRulePaths, ok := agentRules[agentName]
	if !ok {
		return fmt.Errorf("unknown agent: %s", agentName)
	}

	// Get Default agent rules
	defaultRulePaths := agentRules[Default]

	fmt.Fprintf(os.Stderr, "Exporting to %s...\n", agentName)

	// Build a map from normalized paths to target paths
	normalizedToTarget := make(map[string]string)
	for _, rp := range targetRulePaths {
		normalizedToTarget[rp.Normalized()] = rp.Source()
	}

	// Process default agent rules and copy to target agent locations
	for _, defaultRP := range defaultRulePaths {
		sourcePath := defaultRP.Source()
		normalizedPath := defaultRP.Normalized()

		// Skip if the path doesn't exist
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(sourcePath, func(filePath string, info os.FileInfo, err error) error {
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

			// Find target path from normalized path
			if targetPath, ok := normalizedToTarget[normalizedPath]; ok {
				// Create directory if needed
				targetDir := filepath.Dir(targetPath)
				if err := os.MkdirAll(targetDir, 0755); err != nil {
					return fmt.Errorf("failed to create target directory: %w", err)
				}

				// Write content to target file
				if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
					return fmt.Errorf("failed to write target file: %w", err)
				}

				fmt.Fprintf(os.Stderr, "  Exported %s to %s\n", filePath, targetPath)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
