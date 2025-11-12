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
task_name: test-task
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
task_name: test-task
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
task_name: test-task
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
task_name: test-task
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
task_name: test-task
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
task_name: test-task
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
task_name: test-task
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
task_name: test-task
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
task_name: test-task
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
	taskContent := `---
task_name: test-opencode
---
# Test OpenCode Task

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

func TestTaskSelectionByFrontmatter(t *testing.T) {
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

	// Create a task file with a different filename than task_name
	// This tests that filename doesn't matter, only task_name matters
	taskFile := filepath.Join(tasksDir, "arbitrary-filename.md")
	taskContent := `---
task_name: my-special-task
---
# My Special Task

This task has a different filename than task_name.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary with task name matching the task_name frontmatter, not filename
	cmd = exec.Command(binaryPath, "-C", tmpDir, "my-special-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that task content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "# My Special Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestTaskMissingTaskNameError(t *testing.T) {
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

	// Create a task file WITHOUT task_name in frontmatter
	taskFile := filepath.Join(tasksDir, "bad-task.md")
	taskContent := `---
description: A task without task_name
---
# Bad Task

This task is missing task_name in frontmatter.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary - should fail with an error
	cmd = exec.Command(binaryPath, "-C", tmpDir, "bad-task")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected binary to fail, but it succeeded")
	}

	// Check that error message mentions missing task_name
	outputStr := string(output)
	if !strings.Contains(outputStr, "missing required 'task_name' field in frontmatter") {
		t.Errorf("expected error about missing task_name, got: %s", outputStr)
	}
}

func TestMultipleTasksWithSameNameError(t *testing.T) {
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

	// Create two task files with the SAME task_name
	taskFile1 := filepath.Join(tasksDir, "file1.md")
	taskContent1 := `---
task_name: duplicate-task
---
# Task File 1

This is the first file.
`
	if err := os.WriteFile(taskFile1, []byte(taskContent1), 0644); err != nil {
		t.Fatalf("failed to write task file 1: %v", err)
	}

	taskFile2 := filepath.Join(tasksDir, "file2.md")
	taskContent2 := `---
task_name: duplicate-task
---
# Task File 2

This is the second file.
`
	if err := os.WriteFile(taskFile2, []byte(taskContent2), 0644); err != nil {
		t.Fatalf("failed to write task file 2: %v", err)
	}

	// Run the binary - should fail with an error about duplicate task names
	cmd = exec.Command(binaryPath, "-C", tmpDir, "duplicate-task")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected binary to fail with duplicate task names, but it succeeded")
	}

	// Check that error message mentions multiple task files
	outputStr := string(output)
	if !strings.Contains(outputStr, "multiple task files found") {
		t.Errorf("expected error about multiple task files, got: %s", outputStr)
	}
}

func TestTaskSelectionWithSelectors(t *testing.T) {
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

	// Create two task files with the same task_name but different environments
	taskFile1 := filepath.Join(tasksDir, "deploy-staging.md")
	taskContent1 := `---
task_name: deploy
environment: staging
---
# Deploy to Staging

Deploy to the staging environment.
`
	if err := os.WriteFile(taskFile1, []byte(taskContent1), 0644); err != nil {
		t.Fatalf("failed to write staging task file: %v", err)
	}

	taskFile2 := filepath.Join(tasksDir, "deploy-production.md")
	taskContent2 := `---
task_name: deploy
environment: production
---
# Deploy to Production

Deploy to the production environment.
`
	if err := os.WriteFile(taskFile2, []byte(taskContent2), 0644); err != nil {
		t.Fatalf("failed to write production task file: %v", err)
	}

	// Run the binary with selector for staging
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-s", "environment=staging", "deploy")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary for staging: %v\n%s", err, output)
	}

	// Check that staging task content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "# Deploy to Staging") {
		t.Errorf("staging task content not found in stdout")
	}
	if strings.Contains(outputStr, "# Deploy to Production") {
		t.Errorf("production task content should not be in stdout when selecting staging")
	}

	// Run the binary with selector for production
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-s", "environment=production", "deploy")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary for production: %v\n%s", err, output)
	}

	// Check that production task content is present
	outputStr = string(output)
	if !strings.Contains(outputStr, "# Deploy to Production") {
		t.Errorf("production task content not found in stdout")
	}
	if strings.Contains(outputStr, "# Deploy to Staging") {
		t.Errorf("staging task content should not be in stdout when selecting production")
	}
}

func TestResumeMode(t *testing.T) {
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

	// Create a rule file that should be included in normal mode
	ruleFile := filepath.Join(rulesDir, "coding-standards.md")
	ruleContent := `---
---
# Coding Standards

These are the coding standards for the project.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a normal task file (with resume: false)
	normalTaskFile := filepath.Join(tasksDir, "fix-bug-initial.md")
	normalTaskContent := `---
task_name: fix-bug
resume: false
---
# Fix Bug (Initial)

This is the initial task prompt for fixing a bug.
`
	if err := os.WriteFile(normalTaskFile, []byte(normalTaskContent), 0644); err != nil {
		t.Fatalf("failed to write normal task file: %v", err)
	}

	// Create a resume task file (with resume: true)
	resumeTaskFile := filepath.Join(tasksDir, "fix-bug-resume.md")
	resumeTaskContent := `---
task_name: fix-bug
resume: true
---
# Fix Bug (Resume)

This is the resume task prompt for continuing the bug fix.
`
	if err := os.WriteFile(resumeTaskFile, []byte(resumeTaskContent), 0644); err != nil {
		t.Fatalf("failed to write resume task file: %v", err)
	}

	// Test 1: Run in normal mode (with -s resume=false to select non-resume task)
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-s", "resume=false", "fix-bug")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary in normal mode: %v\n%s", err, output)
	}

	outputStr := string(output)

	// In normal mode, rules should be included
	if !strings.Contains(outputStr, "# Coding Standards") {
		t.Errorf("normal mode: rule content not found in stdout")
	}

	// In normal mode, should use the normal task (not resume task)
	if !strings.Contains(outputStr, "# Fix Bug (Initial)") {
		t.Errorf("normal mode: normal task content not found in stdout")
	}
	if strings.Contains(outputStr, "# Fix Bug (Resume)") {
		t.Errorf("normal mode: resume task content should not be in stdout")
	}

	// Test 2: Run in resume mode (with -r flag)
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-r", "fix-bug")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary in resume mode: %v\n%s", err, output)
	}

	outputStr = string(output)

	// In resume mode, rules should NOT be included
	if strings.Contains(outputStr, "# Coding Standards") {
		t.Errorf("resume mode: rule content should not be in stdout")
	}

	// In resume mode, should use the resume task
	if !strings.Contains(outputStr, "# Fix Bug (Resume)") {
		t.Errorf("resume mode: resume task content not found in stdout")
	}
	if strings.Contains(outputStr, "# Fix Bug (Initial)") {
		t.Errorf("resume mode: normal task content should not be in stdout")
	}
}

func TestRemoteRuleFromHTTP(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a remote directory structure to serve
	remoteDir := t.TempDir()
	rulesDir := filepath.Join(remoteDir, ".agents", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create remote rules dir: %v", err)
	}

	// Create a remote rule file
	remoteRuleFile := filepath.Join(rulesDir, "remote-rule.md")
	remoteRuleContent := `---
---
# Remote Rule

This is a rule loaded from a remote directory.
`
	if err := os.WriteFile(remoteRuleFile, []byte(remoteRuleContent), 0644); err != nil {
		t.Fatalf("failed to write remote rule file: %v", err)
	}

	// Create a temporary directory structure for local task
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
task_name: test-task
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary with remote directory (using file:// URL)
	remoteURL := "file://" + remoteDir
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-d", remoteURL, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that remote rule content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "# Remote Rule") {
		t.Errorf("remote rule content not found in stdout")
	}
	if !strings.Contains(outputStr, "This is a rule loaded from a remote directory") {
		t.Errorf("remote rule description not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestPrintTaskFrontmatter(t *testing.T) {
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
	ruleFile := filepath.Join(rulesDir, "test-rule.md")
	ruleContent := `---
language: go
---
# Test Rule

This is a test rule.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a task file with frontmatter
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
task_name: test-task
author: tester
version: 1.0
---
# Test Task

This is a test task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Test without -t flag (should not print frontmatter)
	cmd = exec.Command(binaryPath, "-C", tmpDir, "test-task")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary without -t: %v\n%s", err, output)
	}

	outputStr := string(output)
	// Should not contain frontmatter delimiters in the main output
	lines := strings.Split(outputStr, "\n")
	if len(lines) > 0 && lines[0] == "---" {
		t.Errorf("frontmatter should not be printed without -t flag")
	}
	// Task content should be present
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in output without -t flag")
	}
	// Rule content should be present
	if !strings.Contains(outputStr, "# Test Rule") {
		t.Errorf("rule content not found in output without -t flag")
	}

	// Test with -t flag (should print frontmatter)
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-t", "test-task")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary with -t: %v\n%s", err, output)
	}

	outputStr = string(output)
	lines = strings.Split(outputStr, "\n")
	
	// First line should be frontmatter delimiter
	if lines[0] != "---" {
		t.Errorf("expected first line to be '---', got %q", lines[0])
	}
	
	// Should contain task frontmatter fields
	if !strings.Contains(outputStr, "task_name: test-task") {
		t.Errorf("task frontmatter field 'task_name' not found in output")
	}
	if !strings.Contains(outputStr, "author: tester") {
		t.Errorf("task frontmatter field 'author' not found in output")
	}
	if !strings.Contains(outputStr, "version: 1.0") {
		t.Errorf("task frontmatter field 'version' not found in output")
	}
	
	// Find the second --- (end of frontmatter)
	secondDelimiterIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			secondDelimiterIdx = i
			break
		}
	}
	if secondDelimiterIdx == -1 {
		t.Errorf("expected to find closing frontmatter delimiter '---'")
	}
	
	// Rule content should appear after frontmatter
	if !strings.Contains(outputStr, "# Test Rule") {
		t.Errorf("rule content not found in output with -t flag")
	}
	
	// Task content should appear after rules
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in output with -t flag")
	}
	
	// Verify order: frontmatter should come before rules, rules before task content
	frontmatterIdx := strings.Index(outputStr, "task_name: test-task")
	ruleIdx := strings.Index(outputStr, "# Test Rule")
	taskIdx := strings.Index(outputStr, "# Test Task")
	
	if frontmatterIdx == -1 || ruleIdx == -1 || taskIdx == -1 {
		t.Fatalf("could not find all required sections in output")
	}
	
	if frontmatterIdx > ruleIdx {
		t.Errorf("frontmatter should appear before rules")
	}
	if ruleIdx > taskIdx {
		t.Errorf("rules should appear before task content")
	}
	
	// Rule frontmatter should NOT be printed
	if strings.Contains(outputStr, "language: go") {
		t.Errorf("rule frontmatter should not be printed in output")
	}
}
