package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runBootstrap runs bootstrap scripts for the default agent
func runBootstrap(ctx context.Context, exportRules map[Agent][]RulePath, args []string) error {
	// Get the Default agent's rules
	rulePaths := exportRules[Default]

	// Walk through all rule paths and find bootstrap scripts
	for _, rp := range rulePaths {
		path := rp.SourcePath()

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

			// Check for a bootstrap file named <markdown-file-without-md-suffix>-bootstrap
			baseNameWithoutExt := strings.TrimSuffix(filePath, ext)
			bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

			// Check if bootstrap file exists
			if _, err := os.Stat(bootstrapFilePath); os.IsNotExist(err) {
				return nil
			}

			// Get absolute path for execution
			absBootstrapPath, err := filepath.Abs(bootstrapFilePath)
			if err != nil {
				absBootstrapPath = bootstrapFilePath
			}

			// Run the bootstrap script
			fmt.Fprintf(os.Stdout, "Running bootstrap script: %s\n", absBootstrapPath)

			cmd := exec.CommandContext(ctx, absBootstrapPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = filepath.Dir(absBootstrapPath)

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run bootstrap script %s: %w", absBootstrapPath, err)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
