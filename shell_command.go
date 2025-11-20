package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// shellCommandPattern matches lines that start with !`command`
// The pattern captures the command inside backticks
var shellCommandPattern = regexp.MustCompile(`(?m)^!\x60([^\x60]*)\x60\s*$`)

// processShellCommands finds all shell commands in the format !`command`
// and replaces them with the command output.
// Commands are executed in the current working directory.
func processShellCommands(ctx context.Context, content string) (string, error) {
	var errs []string

	result := shellCommandPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the command from the match
		submatches := shellCommandPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		command := strings.TrimSpace(submatches[1])

		// Execute the command using sh -c to support shell features
		cmd := exec.CommandContext(ctx, "sh", "-c", command)

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			errMsg := fmt.Sprintf("Error executing command '%s': %v", command, err)
			if stderr.Len() > 0 {
				errMsg += fmt.Sprintf("\nStderr: %s", stderr.String())
			}
			errs = append(errs, errMsg)
			return fmt.Sprintf("<!-- %s -->", errMsg)
		}

		// Return the command output, trimming trailing newline if present
		output := stdout.String()
		return strings.TrimSuffix(output, "\n")
	})

	if len(errs) > 0 {
		return result, fmt.Errorf("shell command errors: %s", strings.Join(errs, "; "))
	}

	return result, nil
}
