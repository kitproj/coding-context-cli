package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootstrapFromFile(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "setup.md")
	ruleContent := `---
---
# Development Setup

This is a setup guide.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file for the rule (setup.md -> setup-bootstrap)
	bootstrapFile := filepath.Join(rulesDir, "setup-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Running bootstrap"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that bootstrap output appears before rule content
	outputStr := string(output)
	bootstrapIdx := strings.Index(outputStr, "Running bootstrap")
	setupIdx := strings.Index(outputStr, "# Development Setup")

	if bootstrapIdx == -1 {
		t.Errorf("bootstrap output not found in stdout")
	}
	if setupIdx == -1 {
		t.Errorf("rule content not found in stdout")
	}
	if bootstrapIdx != -1 && setupIdx != -1 && bootstrapIdx > setupIdx {
		t.Errorf("bootstrap output should appear before rule content")
	}

	// Check that task content is present
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestBootstrapFileNotRequired(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file WITHOUT a bootstrap
	ruleFile := filepath.Join(rulesDir, "info.md")
	ruleContent := `---
---
# Project Info

General information about the project.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary - should succeed without a bootstrap file
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that rule content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "# Project Info") {
		t.Errorf("rule content not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestMultipleBootstrapFiles(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create first rule file with bootstrap
	ruleFile1 := filepath.Join(rulesDir, "setup.md")
	ruleContent1 := `---
---
# Setup

Setup instructions.
`
	if err := os.WriteFile(ruleFile1, []byte(ruleContent1), 0644); err != nil {
		t.Fatalf("failed to write rule file 1: %v", err)
	}

	bootstrapFile1 := filepath.Join(rulesDir, "setup-bootstrap")
	bootstrapContent1 := `#!/bin/bash
echo "Running setup bootstrap"
`
	if err := os.WriteFile(bootstrapFile1, []byte(bootstrapContent1), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file 1: %v", err)
	}

	// Create second rule file with bootstrap
	ruleFile2 := filepath.Join(rulesDir, "deploy.md")
	ruleContent2 := `---
---
# Deploy

Deployment instructions.
`
	if err := os.WriteFile(ruleFile2, []byte(ruleContent2), 0644); err != nil {
		t.Fatalf("failed to write rule file 2: %v", err)
	}

	bootstrapFile2 := filepath.Join(rulesDir, "deploy-bootstrap")
	bootstrapContent2 := `#!/bin/bash
echo "Running deploy bootstrap"
`
	if err := os.WriteFile(bootstrapFile2, []byte(bootstrapContent2), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file 2: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that both bootstrap scripts ran
	outputStr := string(output)
	if !strings.Contains(outputStr, "Running setup bootstrap") {
		t.Errorf("setup bootstrap output not found in stdout")
	}
	if !strings.Contains(outputStr, "Running deploy bootstrap") {
		t.Errorf("deploy bootstrap output not found in stdout")
	}

	// Check that both rule contents are present
	if !strings.Contains(outputStr, "# Setup") {
		t.Errorf("setup rule content not found in stdout")
	}
	if !strings.Contains(outputStr, "# Deploy") {
		t.Errorf("deploy rule content not found in stdout")
	}
}

func TestSelectorFiltering(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create rule files with different selectors
	ruleFile1 := filepath.Join(rulesDir, "python.md")
	ruleContent1 := `---
language: python
---
# Python Guidelines

Python specific guidelines.
`
	if err := os.WriteFile(ruleFile1, []byte(ruleContent1), 0644); err != nil {
		t.Fatalf("failed to write python rule file: %v", err)
	}

	ruleFile2 := filepath.Join(rulesDir, "golang.md")
	ruleContent2 := `---
language: go
---
# Go Guidelines

Go specific guidelines.
`
	if err := os.WriteFile(ruleFile2, []byte(ruleContent2), 0644); err != nil {
		t.Fatalf("failed to write go rule file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary with selector filtering for Python
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-s", "language=python", "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that only Python guidelines are included
	outputStr := string(output)
	if !strings.Contains(outputStr, "# Python Guidelines") {
		t.Errorf("Python guidelines not found in stdout")
	}
	if strings.Contains(outputStr, "# Go Guidelines") {
		t.Errorf("Go guidelines should not be in stdout when filtering for Python")
	}
}

func TestTemplateExpansionWithOsExpand(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a task file with template variables
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please work on ${component} and fix ${issue}.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary with parameters
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-p", "component=auth", "-p", "issue=login bug", "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that template variables were expanded
	outputStr := string(output)
	if !strings.Contains(outputStr, "Please work on auth and fix login bug.") {
		t.Errorf("template variables were not expanded correctly. Output:\n%s", outputStr)
	}
}

func TestWorkDirOption(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary with -C flag (from a different directory)
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	cmd.Dir = "/" // Start from root directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that task content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestMdcFileSupport(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a .mdc rule file
	ruleFile := filepath.Join(rulesDir, "custom.mdc")
	ruleContent := `---
---
# Custom Rules

This is a .mdc file.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write .mdc rule file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that .mdc file content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "# Custom Rules") {
		t.Errorf(".mdc file content not found in stdout")
	}
}

func TestMdcFileWithBootstrap(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a .mdc rule file
	ruleFile := filepath.Join(rulesDir, "custom.mdc")
	ruleContent := `---
---
# Custom Rules

This is a .mdc file with bootstrap.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write .mdc rule file: %v", err)
	}

	// Create a bootstrap file for the .mdc file (custom.mdc -> custom-bootstrap)
	bootstrapFile := filepath.Join(rulesDir, "custom-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Running custom bootstrap"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that bootstrap ran and content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "Running custom bootstrap") {
		t.Errorf("custom bootstrap output not found in stdout")
	}
	if !strings.Contains(outputStr, "# Custom Rules") {
		t.Errorf(".mdc file content not found in stdout")
	}
}

func TestBootstrapWithoutExecutePermission(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "setup.md")
	ruleContent := `---
---
# Development Setup

This is a setup guide.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file WITHOUT execute permission (0644 instead of 0755)
	// This simulates a bootstrap file that was checked out from git on Windows
	// or otherwise doesn't have the executable bit set
	bootstrapFile := filepath.Join(rulesDir, "setup-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Bootstrap executed successfully"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0644); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Verify the file is not executable initially
	fileInfo, err := os.Stat(bootstrapFile)
	if err != nil {
		t.Fatalf("failed to stat bootstrap file: %v", err)
	}
	if fileInfo.Mode()&0111 != 0 {
		t.Fatalf("bootstrap file should not be executable initially, but has mode: %v", fileInfo.Mode())
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary - this should chmod +x the bootstrap file before running it
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that bootstrap output appears (proving it ran successfully)
	outputStr := string(output)
	if !strings.Contains(outputStr, "Bootstrap executed successfully") {
		t.Errorf("bootstrap output not found in stdout, meaning it didn't run successfully")
	}

	// Check that rule content is present
	if !strings.Contains(outputStr, "# Development Setup") {
		t.Errorf("rule content not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}

	// Verify the bootstrap file is now executable
	fileInfo, err = os.Stat(bootstrapFile)
	if err != nil {
		t.Fatalf("failed to stat bootstrap file after run: %v", err)
	}
	if fileInfo.Mode()&0111 == 0 {
		t.Errorf("bootstrap file should be executable after run, but has mode: %v", fileInfo.Mode())
	}
}

func TestOpenCodeRulesSupport(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	openCodeAgentDir := filepath.Join(tmpDir, ".opencode", "agent")
	openCodeCommandDir := filepath.Join(tmpDir, ".opencode", "command")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(openCodeAgentDir, 0755); err != nil {
		t.Fatalf("failed to create opencode agent dir: %v", err)
	}
	if err := os.MkdirAll(openCodeCommandDir, 0755); err != nil {
		t.Fatalf("failed to create opencode command dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create an agent rule file in .opencode/agent
	agentFile := filepath.Join(openCodeAgentDir, "docs.md")
	agentContent := `# Documentation Agent

This agent helps with documentation.
`
	if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	// Create a command rule file in .opencode/command
	commandFile := filepath.Join(openCodeCommandDir, "commit.md")
	commandContent := `# Commit Command

This command helps create commits.
`
	if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
		t.Fatalf("failed to write command file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-opencode.md")
	taskContent := `# Test OpenCode Task

This is a test task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-opencode")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	outputStr := string(output)

	// Check that agent rule content is present
	if !strings.Contains(outputStr, "# Documentation Agent") {
		t.Errorf("OpenCode agent rule content not found in stdout")
	}

	// Check that command rule content is present
	if !strings.Contains(outputStr, "# Commit Command") {
		t.Errorf("OpenCode command rule content not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(outputStr, "# Test OpenCode Task") {
		t.Errorf("task content not found in stdout")
	}
}
