package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// runExport exports rules from the default agent to the specified agent
func runExport(ctx context.Context, exportRules map[Agent][]RulePath, args []string) error {
	// Define flags for export command
	var includes selectorMap
	var excludes selectorMap
	exportFlags := flag.NewFlagSet("export", flag.ExitOnError)
	exportFlags.Var(&includes, "s", "Include rules with matching frontmatter (key=value)")
	exportFlags.Var(&excludes, "S", "Exclude rules with matching frontmatter (key=value)")

	if err := exportFlags.Parse(args); err != nil {
		return err
	}

	exportArgs := exportFlags.Args()
	if len(exportArgs) < 1 {
		return fmt.Errorf("usage: coding-context export <agent> [-s key=value] [-S key=value]")
	}

	agentName := Agent(exportArgs[0])

	// Check if agent is valid and not Default
	if agentName == Default {
		return fmt.Errorf("cannot export to default agent")
	}

	targetRulePaths, ok := exportRules[agentName]
	if !ok {
		return fmt.Errorf("unknown agent: %s", agentName)
	}

	fmt.Fprintf(os.Stderr, "Exporting to %s...\n", agentName)

	// Process default agent rules and copy to target agent locations
	for _, targetRP := range targetRulePaths {
		sourcePath := targetRP.SourcePath()
		targetPath := targetRP.TargetPath()

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

			// Check if file matches include and exclude selectors
			if !includes.matchesIncludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (does not match include selectors): %s\n", filePath)
				return nil
			}
			if !excludes.matchesExcludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (matches exclude selectors): %s\n", filePath)
				return nil
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

			// Write content to target file
			if err := os.WriteFile(actualTarget, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write target file: %w", err)
			}

			fmt.Fprintf(os.Stderr, "  Exported %s to %s\n", filePath, actualTarget)

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
