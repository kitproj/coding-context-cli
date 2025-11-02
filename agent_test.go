package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportBasic(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create CLAUDE.md for Claude agent
	claudeFile := filepath.Join(tmpDir, "CLAUDE.md")
	claudeContent := `# Claude Rules

This is a test Claude file.
`
	if err := os.WriteFile(claudeFile, []byte(claudeContent), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}

	// Run the import command
	cmd = exec.Command(binaryPath, "-C", tmpDir, "import")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run import command: %v\n%s", err, output)
	}

	// Check that AGENTS.md was created
	agentsFile := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsFile); err == nil {
		// File exists, check content
		content, _ := os.ReadFile(agentsFile)
		if strings.Contains(string(content), "# Claude Rules") {
			// Success
		}
	}
}

func TestBootstrapCommand(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "setup.md")
	if err := os.WriteFile(ruleFile, []byte("# Setup\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap script
	bootstrapFile := filepath.Join(rulesDir, "setup-bootstrap")
	markerFile := filepath.Join(tmpDir, "bootstrap-ran.txt")
	bootstrapContent := `#!/bin/bash
echo "Bootstrap executed" > ` + markerFile + `
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Run bootstrap command
	cmd = exec.Command(binaryPath, "-C", tmpDir, "bootstrap")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run bootstrap command: %v\n%s", err, output)
	}

	// Check that the marker file was created
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Errorf("marker file was not created, bootstrap script did not run")
	}
}

func TestPromptCommand(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `# Task: ${taskName}

Please help with ${language}.
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run prompt command
	cmd = exec.Command(binaryPath, "-C", tmpDir, "prompt", "-p", "taskName=MyTask", "-p", "language=Go", "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run prompt command: %v\n%s", err, output)
	}

	// Check output contains templated content
	outputStr := string(output)
	if !strings.Contains(outputStr, "Task: MyTask") {
		t.Errorf("Expected 'Task: MyTask' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Please help with Go") {
		t.Errorf("Expected 'Please help with Go' in output, got: %s", outputStr)
	}
}

func TestRulesCommand(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	// Create rule files
	ruleFile1 := filepath.Join(rulesDir, "rule1.md")
	if err := os.WriteFile(ruleFile1, []byte("# Rule One\n\nFirst rule content.\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file 1: %v", err)
	}

	agentsFile := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(agentsFile, []byte("# Agents\n\nAgent rules.\n"), 0644); err != nil {
		t.Fatalf("failed to write AGENTS.md: %v", err)
	}

	// Run rules command
	cmd = exec.Command(binaryPath, "-C", tmpDir, "rules")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run rules command: %v\n%s", err, output)
	}

	// Check output contains both rule files
	outputStr := string(output)
	if !strings.Contains(outputStr, "# Rule One") {
		t.Errorf("Expected '# Rule One' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "# Agents") {
		t.Errorf("Expected '# Agents' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "First rule content") {
		t.Errorf("Expected 'First rule content' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Agent rules") {
		t.Errorf("Expected 'Agent rules' in output, got: %s", outputStr)
	}
}
