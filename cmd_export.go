package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// runExport exports rules from the default agent to the specified agent
func runExport(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: coding-context export <agent>")
	}

	agentName := Agent(args[0])

	// Check if agent is valid and not Default
	if agentName == Default {
		return fmt.Errorf("cannot export to default agent")
	}

	targetLevels, ok := agentRules[agentName]
	if !ok {
		return fmt.Errorf("unknown agent: %s", agentName)
	}

	// Get Default agent rules
	defaultLevels := agentRules[Default]

	fmt.Fprintf(os.Stderr, "Exporting to %s...\n", agentName)

	// Process default agent rules and copy to target agent locations
	for level := ProjectLevel; level <= SystemLevel; level++ {
		defaultPaths, ok := defaultLevels[level]
		if !ok {
			continue
		}

		targetPaths, ok := targetLevels[level]
		if !ok {
			continue
		}

		for _, defaultPath := range defaultPaths {
			// Skip if the path doesn't exist
			if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
				continue
			}

			err := filepath.Walk(defaultPath, func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}

				// Only process .md and .mdc files
				ext := filepath.Ext(filePath)
				if ext != ".md" && ext != ".mdc" {
					return nil
				}

				// Parse frontmatter
				var frontmatter map[string]string
				content, err := parseMarkdownFile(filePath, &frontmatter)
				if err != nil {
					return fmt.Errorf("failed to parse markdown file: %w", err)
				}

				// Copy to target agent paths
				// Use first target path for this level
				if len(targetPaths) > 0 {
					targetPath := targetPaths[0]

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
	}

	return nil
}
