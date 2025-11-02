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

// Check that .agents directory was created
agentsDir := filepath.Join(tmpDir, ".agents")
if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
t.Errorf(".agents directory was not created")
}
}

func TestExportBasic(t *testing.T) {
// Build the binary
binaryPath := filepath.Join(t.TempDir(), "coding-context")
cmd := exec.Command("go", "build", "-o", binaryPath, ".")
if output, err := cmd.CombinedOutput(); err != nil {
t.Fatalf("failed to build binary: %v\n%s", err, output)
}

tmpDir := t.TempDir()

// Create default agent rules
agentsRulesDir := filepath.Join(tmpDir, ".agents", "rules")
if err := os.MkdirAll(agentsRulesDir, 0755); err != nil {
t.Fatalf("failed to create .agents/rules: %v", err)
}

ruleFile := filepath.Join(agentsRulesDir, "test.md")
if err := os.WriteFile(ruleFile, []byte("# Test Rule\n"), 0644); err != nil {
t.Fatalf("failed to write rule file: %v", err)
}

// Run export to Claude
cmd = exec.Command(binaryPath, "-C", tmpDir, "export", "Claude")
if output, err := cmd.CombinedOutput(); err != nil {
t.Fatalf("failed to run export command: %v\n%s", err, output)
}

// Check that CLAUDE.local.md was created
claudeFile := filepath.Join(tmpDir, "CLAUDE.local.md")
if _, err := os.Stat(claudeFile); os.IsNotExist(err) {
t.Errorf("CLAUDE.local.md was not created")
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
cmd = exec.Command(binaryPath, "-C", tmpDir, "prompt", "test-task", "taskName=MyTask", "language=Go")
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
