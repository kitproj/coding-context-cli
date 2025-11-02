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

// Create AGENTS.md for Codex agent
agentsFile := filepath.Join(tmpDir, "AGENTS.md")
agentsContent := `# Test Agents

This is a test agents file.
`
if err := os.WriteFile(agentsFile, []byte(agentsContent), 0644); err != nil {
t.Fatalf("failed to write AGENTS.md: %v", err)
}

// Run the import command
cmd = exec.Command(binaryPath, "import", "Codex")
cmd.Dir = tmpDir
output, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("failed to run import command: %v\n%s", err, output)
}

// Check that rules.md was created
rulesOutput := filepath.Join(tmpDir, "rules.md")
if _, err := os.Stat(rulesOutput); os.IsNotExist(err) {
t.Errorf("rules.md file was not created")
}

// Check content
content, err := os.ReadFile(rulesOutput)
if err != nil {
t.Fatalf("failed to read rules.md: %v", err)
}
if !strings.Contains(string(content), "# Test Agents") {
t.Errorf("Expected test content in rules.md")
}
}

func TestImportUnknownAgent(t *testing.T) {
// Build the binary
binaryPath := filepath.Join(t.TempDir(), "coding-context")
cmd := exec.Command("go", "build", "-o", binaryPath, ".")
if output, err := cmd.CombinedOutput(); err != nil {
t.Fatalf("failed to build binary: %v\n%s", err, output)
}

tmpDir := t.TempDir()

// Run with unknown agent
cmd = exec.Command(binaryPath, "import", "UnknownAgent")
cmd.Dir = tmpDir
output, err := cmd.CombinedOutput()

if err == nil {
t.Errorf("Expected error for unknown agent")
}

if !strings.Contains(string(output), "unknown agent") {
t.Errorf("Expected 'unknown agent' error message, got: %s", string(output))
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
rulesDir := filepath.Join(tmpDir, ".prompts", "rules")
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
cmd = exec.Command(binaryPath, "bootstrap")
cmd.Dir = tmpDir
if output, err := cmd.CombinedOutput(); err != nil {
t.Fatalf("failed to run bootstrap command: %v\n%s", err, output)
}

// Check that the marker file was created
if _, err := os.Stat(markerFile); os.IsNotExist(err) {
t.Errorf("marker file was not created, bootstrap script did not run")
}
}
