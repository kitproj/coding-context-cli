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
func runBootstrap(ctx context.Context, args []string) error {
	// Get the Default agent's rules
	levels := agentRules[Default]

	// Walk through all rule paths and find bootstrap scripts
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

				// Only process .md and .mdc files
				ext := filepath.Ext(filePath)
				if ext != ".md" && ext != ".mdc" {
					return nil
				}

				// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
				baseNameWithoutExt := strings.TrimSuffix(filePath, ext)
				bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

				// Check if bootstrap file exists
				if _, err := os.Stat(bootstrapFilePath); os.IsNotExist(err) {
					return nil
				}

				// Get absolute path
				absBootstrapPath, err := filepath.Abs(bootstrapFilePath)
				if err != nil {
					return fmt.Errorf("failed to get absolute path for bootstrap script: %w", err)
				}

				// Make it executable
				if err := os.Chmod(absBootstrapPath, 0755); err != nil {
					return fmt.Errorf("failed to make bootstrap script executable: %w", err)
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
	}

	return nil
}
