package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// runImport imports rules from all known agents into the default agent locations
func runImport(ctx context.Context, args []string) error {
	// Iterate over all agents except Default
	for agent, levels := range agentRules {
		if agent == Default {
			continue
		}

		fmt.Fprintf(os.Stderr, "Importing from %s...\n", agent)

		// Process rules in level order (0, 1, 2, 3)
		for level := ProjectLevel; level <= SystemLevel; level++ {
			paths, ok := levels[level]
			if !ok {
				continue
			}

			for _, path := range paths {
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

					// Only process .md and .mdc files as rule files
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

					// Determine target path in default agent
					var targetPath string
					switch level {
					case ProjectLevel:
						targetPath = filepath.Join(".agents", "rules", fmt.Sprintf("%s.md", agent))
					case AncestorLevel:
						// Write to .agents/AGENTS.md
						targetPath = filepath.Join(".agents", "AGENTS.md")
					case UserLevel:
						homeDir, _ := os.UserHomeDir()
						targetPath = filepath.Join(homeDir, ".agents", "AGENTS.md")
					case SystemLevel:
						targetPath = filepath.Join("/etc", "agents", "rules", fmt.Sprintf("%s.md", agent))
					}

					// Create directory if needed
					targetDir := filepath.Dir(targetPath)
					if err := os.MkdirAll(targetDir, 0755); err != nil {
						return fmt.Errorf("failed to create target directory: %w", err)
					}

					// Append content to target file
					f, err := os.OpenFile(targetPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return fmt.Errorf("failed to open target file: %w", err)
					}
					defer f.Close()

					if _, err := f.WriteString(content + "\n\n"); err != nil {
						return fmt.Errorf("failed to write content: %w", err)
					}

					fmt.Fprintf(os.Stderr, "  Imported %s to %s\n", filePath, targetPath)

					return nil
				})
				if err != nil {
					return fmt.Errorf("failed to walk rule path: %w", err)
				}
			}
		}
	}

	return nil
}
