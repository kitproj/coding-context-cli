package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// runImport imports rules from all known agents into the default agent locations
func runImport(ctx context.Context, importRules map[Agent][]RulePath, args []string) error {
	// Iterate over all agents except Default
	for agent, rulePaths := range importRules {
		if agent == Default {
			continue
		}

		fmt.Fprintf(os.Stderr, "Importing from %s...\n", agent)

		for _, rp := range rulePaths {
			sourcePath := rp.SourcePath()
			targetPath := rp.TargetPath()

			// Skip if the source path doesn't exist
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

				// Determine actual target path
				var actualTarget string
				if info.IsDir() || filepath.Ext(sourcePath) == "" {
					// If source is a directory, map the file relative to it
					relPath, _ := filepath.Rel(sourcePath, filePath)
					actualTarget = filepath.Join(targetPath, relPath)
				} else {
					// Single file mapping
					actualTarget = targetPath
				}

				// Create directory if needed
				targetDir := filepath.Dir(actualTarget)
				if err := os.MkdirAll(targetDir, 0755); err != nil {
					return fmt.Errorf("failed to create target directory: %w", err)
				}

				// Append content to target file (for deduplication later)
				f, err := os.OpenFile(actualTarget, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to open target file: %w", err)
				}
				defer f.Close()

				if _, err := f.WriteString(content + "\n\n"); err != nil {
					return fmt.Errorf("failed to write content: %w", err)
				}

				// Estimate and log token count
				tokens := estimateTokens(content)
				fmt.Fprintf(os.Stderr, "  Imported %s to %s (~%d tokens)\n", filePath, actualTarget, tokens)

				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to walk rule path: %w", err)
			}
		}
	}

	return nil
}
