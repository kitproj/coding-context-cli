package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testDirs holds the directory structure for a test
type testDirs struct {
	tmpDir   string
	rulesDir string
	tasksDir string
}

// setupTestDirs creates the standard directory structure for tests
func setupTestDirs(t *testing.T) testDirs {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	return testDirs{
		tmpDir:   tmpDir,
		rulesDir: rulesDir,
		tasksDir: tasksDir,
	}
}

// runTool executes the program using "go run ." with the given arguments
// It fatally fails the test if the command returns an error.
func runTool(t *testing.T, args ...string) string {
	output, err := runToolWithError(args...)
	if err != nil {
		t.Fatalf("failed to run tool: %v\n%s", err, output)
	}
	return string(output)
}

// runToolWithError executes the program using "go run ." with the given arguments
// and returns both output and error (for tests that expect errors).
func runToolWithError(args ...string) (string, error) {
	// Get the current working directory to use as the source path for go run
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("go", append([]string{"run", wd}, args...)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// createStandardTask creates a standard task file with the given task name
func createStandardTask(t *testing.T, tasksDir, taskName string) {
	taskFile := filepath.Join(tasksDir, taskName+".md")
	taskContent := `---
task_name: ` + taskName + `
---
# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
}

func TestBootstrapFromFile(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file
	ruleFile := filepath.Join(dirs.rulesDir, "setup.md")
	ruleContent := `---
---
# Development Setup

This is a setup guide.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file for the rule (setup.md -> setup-bootstrap)
	bootstrapFile := filepath.Join(dirs.rulesDir, "setup-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Running bootstrap"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0o755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that bootstrap output appears before rule content
	bootstrapIdx := strings.Index(output, "Running bootstrap")
	setupIdx := strings.Index(output, "# Development Setup")

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
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestBootstrapFileNotRequired(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file WITHOUT a bootstrap
	ruleFile := filepath.Join(dirs.rulesDir, "info.md")
	ruleContent := `---
---
# Project Info

General information about the project.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run the program - should succeed without a bootstrap file
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that rule content is present
	if !strings.Contains(output, "# Project Info") {
		t.Errorf("rule content not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestMultipleBootstrapFiles(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create first rule file with bootstrap
	ruleFile1 := filepath.Join(dirs.rulesDir, "setup.md")
	ruleContent1 := `---
---
# Setup

Setup instructions.
`
	if err := os.WriteFile(ruleFile1, []byte(ruleContent1), 0o644); err != nil {
		t.Fatalf("failed to write rule file 1: %v", err)
	}

	bootstrapFile1 := filepath.Join(dirs.rulesDir, "setup-bootstrap")
	bootstrapContent1 := `#!/bin/bash
echo "Running setup bootstrap"
`
	if err := os.WriteFile(bootstrapFile1, []byte(bootstrapContent1), 0o755); err != nil {
		t.Fatalf("failed to write bootstrap file 1: %v", err)
	}

	// Create second rule file with bootstrap
	ruleFile2 := filepath.Join(dirs.rulesDir, "deploy.md")
	ruleContent2 := `---
---
# Deploy

Deployment instructions.
`
	if err := os.WriteFile(ruleFile2, []byte(ruleContent2), 0o644); err != nil {
		t.Fatalf("failed to write rule file 2: %v", err)
	}

	bootstrapFile2 := filepath.Join(dirs.rulesDir, "deploy-bootstrap")
	bootstrapContent2 := `#!/bin/bash
echo "Running deploy bootstrap"
`
	if err := os.WriteFile(bootstrapFile2, []byte(bootstrapContent2), 0o755); err != nil {
		t.Fatalf("failed to write bootstrap file 2: %v", err)
	}

	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that both bootstrap scripts ran
	if !strings.Contains(output, "Running setup bootstrap") {
		t.Errorf("setup bootstrap output not found in stdout")
	}
	if !strings.Contains(output, "Running deploy bootstrap") {
		t.Errorf("deploy bootstrap output not found in stdout")
	}

	// Check that both rule contents are present
	if !strings.Contains(output, "# Setup") {
		t.Errorf("setup rule content not found in stdout")
	}
	if !strings.Contains(output, "# Deploy") {
		t.Errorf("deploy rule content not found in stdout")
	}
}

func TestSelectorFiltering(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create rule files with different Selectors
	ruleFile1 := filepath.Join(dirs.rulesDir, "python.md")
	ruleContent1 := `---
language: python
---
# Python Guidelines

Python specific guidelines.
`
	if err := os.WriteFile(ruleFile1, []byte(ruleContent1), 0o644); err != nil {
		t.Fatalf("failed to write python rule file: %v", err)
	}

	ruleFile2 := filepath.Join(dirs.rulesDir, "golang.md")
	ruleContent2 := `---
language: go
---
# Go Guidelines

Go specific guidelines.
`
	if err := os.WriteFile(ruleFile2, []byte(ruleContent2), 0o644); err != nil {
		t.Fatalf("failed to write go rule file: %v", err)
	}

	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run the program with selector filtering for Python
	output := runTool(t, "-C", dirs.tmpDir, "-s", "language=python", "test-task")

	// Check that only Python guidelines are included
	if !strings.Contains(output, "# Python Guidelines") {
		t.Errorf("Python guidelines not found in stdout")
	}
	if strings.Contains(output, "# Go Guidelines") {
		t.Errorf("Go guidelines should not be in stdout when filtering for Python")
	}
}

func TestTemplateExpansionWithOsExpand(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
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
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program with parameters
	output := runTool(t, "-C", tmpDir, "-p", "component=auth", "-p", "issue=login bug", "test-task")

	// Check that template variables were expanded
	if !strings.Contains(output, "Please work on auth and fix login bug.") {
		t.Errorf("template variables were not expanded correctly. Output:\n%s", output)
	}
}

func TestExpanderIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a test file for path expansion
	dataFile := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(dataFile, []byte("file content"), 0o644); err != nil {
		t.Fatalf("failed to write data file: %v", err)
	}

	// Create a task file with all three expansion types
	taskFile := filepath.Join(tasksDir, "test-expander.md")
	taskContent := fmt.Sprintf(`---
task_name: test-expander
---
# Test Expander

Parameter: ${component}
Command: !`+"`echo hello`"+`
Path: @%s
Combined: ${component} !`+"`echo world`"+`
`, dataFile)
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program with parameters
	output := runTool(t, "-C", tmpDir, "-p", "component=auth", "test-expander")

	// Check parameter expansion
	if !strings.Contains(output, "Parameter: auth") {
		t.Errorf("parameter expansion failed. Output:\n%s", output)
	}

	// Check command expansion
	if !strings.Contains(output, "Command: hello") {
		t.Errorf("command expansion failed. Output:\n%s", output)
	}

	// Check path expansion
	if !strings.Contains(output, "Path: file content") {
		t.Errorf("path expansion failed. Output:\n%s", output)
	}

	// Check combined expansion
	if !strings.Contains(output, "Combined: auth world") {
		t.Errorf("combined expansion failed. Output:\n%s", output)
	}
}

func TestExpanderSecurityIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a file that contains expansion syntax (should not be re-expanded)
	dataFile := filepath.Join(tmpDir, "injection.txt")
	if err := os.WriteFile(dataFile, []byte("${injected} and !`echo hacked`"), 0o644); err != nil {
		t.Fatalf("failed to write data file: %v", err)
	}

	// Create a task file that tests security (no re-expansion)
	taskFile := filepath.Join(tasksDir, "test-security.md")
	taskContent := fmt.Sprintf(`---
task_name: test-security
---
# Test Security

File content: @%s
Param with command: ${evil}
Command with param: !`+"`echo '${secret}'`"+`
`, dataFile)
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program with parameters that contain expansion syntax
	output := runTool(t, "-C", tmpDir, "-p", "evil=!`echo INJECTED`", "-p", "secret=TOPSECRET", "test-security")

	// Split output into lines to separate stderr logs from stdout prompt
	lines := strings.Split(output, "\n")
	var promptLines []string
	inPrompt := false
	for _, line := range lines {
		// Look for the start of the frontmatter (which marks beginning of stdout)
		if strings.HasPrefix(line, "---") {
			inPrompt = true
		}
		if inPrompt {
			promptLines = append(promptLines, line)
		}
	}
	promptOutput := strings.Join(promptLines, "\n")

	// Check that file content with expansion syntax is NOT re-expanded
	if !strings.Contains(promptOutput, "File content: ${injected} and !`echo hacked`") {
		t.Errorf("file content was re-expanded (security issue). Output:\n%s", output)
	}

	// Check that parameter value with command syntax is NOT executed
	if !strings.Contains(promptOutput, "Param with command: !`echo INJECTED`") {
		t.Errorf("parameter with command syntax was executed (security issue). Output:\n%s", output)
	}

	// Check that command output with parameter syntax is NOT re-expanded
	if !strings.Contains(promptOutput, "Command with param: ${secret}") {
		t.Errorf("command output was re-expanded (security issue). Output:\n%s", output)
	}

	// Verify that sensitive data is NOT in the prompt output (only check stdout, not stderr logs)
	// The parameter value should only appear in logging (stderr), not in the actual prompt content
	if strings.Contains(promptOutput, "TOPSECRET") {
		t.Errorf("parameter was re-expanded from command output (security issue). Prompt output:\n%s", promptOutput)
	}
	// Check that the literal command syntax is preserved (not executed)
	// The word "hacked" appears in the literal text, so we check for the full context
	if !strings.Contains(output, "!`echo hacked`") {
		t.Errorf("file content was re-expanded (security issue). Output:\n%s", output)
	}
}

func TestMdcFileSupport(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a .mdc rule file
	ruleFile := filepath.Join(dirs.rulesDir, "custom.mdc")
	ruleContent := `---
---
# Custom Rules

This is a .mdc file.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write .mdc rule file: %v", err)
	}

	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that .mdc file content is present
	if !strings.Contains(output, "# Custom Rules") {
		t.Errorf(".mdc file content not found in stdout")
	}
}

func TestMdcFileWithBootstrap(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a .mdc rule file
	ruleFile := filepath.Join(dirs.rulesDir, "custom.mdc")
	ruleContent := `---
---
# Custom Rules

This is a .mdc file with bootstrap.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write .mdc rule file: %v", err)
	}

	// Create a bootstrap file for the .mdc file (custom.mdc -> custom-bootstrap)
	bootstrapFile := filepath.Join(dirs.rulesDir, "custom-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Running custom bootstrap"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0o755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that bootstrap ran and content is present
	if !strings.Contains(output, "Running custom bootstrap") {
		t.Errorf("custom bootstrap output not found in stdout")
	}
	if !strings.Contains(output, "# Custom Rules") {
		t.Errorf(".mdc file content not found in stdout")
	}
}

func TestBootstrapWithoutExecutePermission(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file
	ruleFile := filepath.Join(dirs.rulesDir, "setup.md")
	ruleContent := `---
---
# Development Setup

This is a setup guide.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file WITHOUT execute permission (0644 instead of 0755)
	// This simulates a bootstrap file that was checked out from git on Windows
	// or otherwise doesn't have the executable bit set
	bootstrapFile := filepath.Join(dirs.rulesDir, "setup-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Bootstrap executed successfully"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0o644); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Verify the file is not executable initially
	fileInfo, err := os.Stat(bootstrapFile)
	if err != nil {
		t.Fatalf("failed to stat bootstrap file: %v", err)
	}
	if fileInfo.Mode()&0o111 != 0 {
		t.Fatalf("bootstrap file should not be executable initially, but has mode: %v", fileInfo.Mode())
	}

	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run the program - this should chmod +x the bootstrap file before running it
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that bootstrap output appears (proving it ran successfully)
	if !strings.Contains(output, "Bootstrap executed successfully") {
		t.Errorf("bootstrap output not found in stdout, meaning it didn't run successfully")
	}

	// Check that rule content is present
	if !strings.Contains(output, "# Development Setup") {
		t.Errorf("rule content not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}

	// Verify the bootstrap file is now executable
	fileInfo, err = os.Stat(bootstrapFile)
	if err != nil {
		t.Fatalf("failed to stat bootstrap file after run: %v", err)
	}
	if fileInfo.Mode()&0o111 == 0 {
		t.Errorf("bootstrap file should be executable after run, but has mode: %v", fileInfo.Mode())
	}
}

func TestOpenCodeRulesSupport(t *testing.T) {
	tmpDir := t.TempDir()
	openCodeAgentDir := filepath.Join(tmpDir, ".opencode", "agent")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(openCodeAgentDir, 0o755); err != nil {
		t.Fatalf("failed to create opencode agent dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create an agent rule file in .opencode/agent
	agentFile := filepath.Join(openCodeAgentDir, "docs.md")
	agentContent := `# Documentation Agent

This agent helps with documentation.
`
	if err := os.WriteFile(agentFile, []byte(agentContent), 0o644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-opencode.md")
	taskContent := `---
task_name: test-opencode
---
# Test OpenCode Task

This is a test task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program
	output := runTool(t, "-C", tmpDir, "test-opencode")

	// Check that agent rule content is present
	if !strings.Contains(output, "# Documentation Agent") {
		t.Errorf("OpenCode agent rule content not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(output, "# Test OpenCode Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestOpenCodeCommandTaskSupport(t *testing.T) {
	tmpDir := t.TempDir()
	// Tasks must be in .agents/tasks directory
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a task file in the correct location
	taskFile := filepath.Join(tasksDir, "fix-bug.md")
	taskContent := `---
task_name: fix-bug
---
# Fix Bug Task

This is a task for fixing bugs.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program
	output := runTool(t, "-C", tmpDir, "fix-bug")

	// Check that task content is present
	if !strings.Contains(output, "# Fix Bug Task") {
		t.Errorf("task content not found in stdout")
	}
	if !strings.Contains(output, "This is a task for fixing bugs.") {
		t.Errorf("task description not found in stdout")
	}
}

func TestTaskSelectionByFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a task file - task name is based on filename now
	taskFile := filepath.Join(tasksDir, "my-special-task.md")
	taskContent := `---
---
# My Special Task

This task name is based on the filename.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program with task name matching the filename
	output := runTool(t, "-C", tmpDir, "my-special-task")

	// Check that task content is present
	if !strings.Contains(output, "# My Special Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestTaskWithoutTaskNameUsesFilename(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a file WITHOUT task_name in frontmatter - should use filename
	taskFile := filepath.Join(tasksDir, "my-task.md")
	taskContent := `---
description: A task without task_name
---
# My Task

This file uses the filename as task_name.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Run the program - should succeed using filename as task name
	output := runTool(t, "-C", tmpDir, "my-task")

	// Check that task content is present
	if !strings.Contains(output, "# My Task") {
		t.Errorf("task content not found in stdout")
	}
	if !strings.Contains(output, "This file uses the filename as task_name.") {
		t.Errorf("task description not found in stdout")
	}
}

func TestTaskSelectionWithSelectors(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create two task files with different filenames but same base name and different environments
	taskFile1 := filepath.Join(tasksDir, "deploy-staging.md")
	taskContent1 := `---
environment: staging
---
# Deploy to Staging

Deploy to the staging environment.
`
	if err := os.WriteFile(taskFile1, []byte(taskContent1), 0o644); err != nil {
		t.Fatalf("failed to write staging task file: %v", err)
	}

	taskFile2 := filepath.Join(tasksDir, "deploy-production.md")
	taskContent2 := `---
environment: production
---
# Deploy to Production

Deploy to the production environment.
`
	if err := os.WriteFile(taskFile2, []byte(taskContent2), 0o644); err != nil {
		t.Fatalf("failed to write production task file: %v", err)
	}

	// Run the program with selector for staging - use the staging task filename
	output := runTool(t, "-C", tmpDir, "-s", "environment=staging", "deploy-staging")

	// Check that staging task content is present
	if !strings.Contains(output, "# Deploy to Staging") {
		t.Errorf("staging task content not found in stdout")
	}
	if strings.Contains(output, "# Deploy to Production") {
		t.Errorf("production task content should not be in stdout when selecting staging")
	}

	// Run the program with selector for production - use the production task filename
	output = runTool(t, "-C", tmpDir, "-s", "environment=production", "deploy-production")

	// Check that production task content is present
	if !strings.Contains(output, "# Deploy to Production") {
		t.Errorf("production task content not found in stdout")
	}
	if strings.Contains(output, "# Deploy to Staging") {
		t.Errorf("staging task content should not be in stdout when selecting production")
	}
}

func TestResumeMode(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file that should be included in normal mode
	ruleFile := filepath.Join(dirs.rulesDir, "coding-standards.md")
	ruleContent := `---
---
# Coding Standards

These are the coding standards for the project.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap script for the rule file to verify it doesn't run in resume mode
	ruleBootstrapFile := filepath.Join(dirs.rulesDir, "coding-standards-bootstrap")
	ruleBootstrapContent := `#!/bin/bash
echo "RULE_BOOTSTRAP_RAN" >&2
`
	if err := os.WriteFile(ruleBootstrapFile, []byte(ruleBootstrapContent), 0o755); err != nil {
		t.Fatalf("failed to write rule bootstrap file: %v", err)
	}

	// Create a normal task file (without resume field)
	normalTaskFile := filepath.Join(dirs.tasksDir, "fix-bug.md")
	normalTaskContent := `---
---
# Fix Bug (Initial)

This is the initial task prompt for fixing a bug.
`
	if err := os.WriteFile(normalTaskFile, []byte(normalTaskContent), 0o644); err != nil {
		t.Fatalf("failed to write normal task file: %v", err)
	}

	// Create a resume task file (with resume: true)
	resumeTaskFile := filepath.Join(dirs.tasksDir, "fix-bug-resume.md")
	resumeTaskContent := `---
resume: true
---
# Fix Bug (Resume)

This is the resume task prompt for continuing the bug fix.
`
	if err := os.WriteFile(resumeTaskFile, []byte(resumeTaskContent), 0o644); err != nil {
		t.Fatalf("failed to write resume task file: %v", err)
	}

	// Test 1: Run in normal mode (without resume selector, or with -s resume=false)
	// Capture stderr to verify bootstrap scripts DO run in normal mode
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	cmd := exec.Command("go", "run", wd, "-C", dirs.tmpDir, "-s", "resume=false", "fix-bug")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run binary in normal mode: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	stderrOutput := stderr.String()

	// In normal mode, rules should be included
	if !strings.Contains(output, "# Coding Standards") {
		t.Errorf("normal mode: rule content not found in stdout")
	}

	// In normal mode, bootstrap scripts SHOULD run
	if !strings.Contains(stderrOutput, "RULE_BOOTSTRAP_RAN") {
		t.Errorf("normal mode: rule bootstrap script should run (stderr: %s)", stderrOutput)
	}

	// In normal mode, should use the normal task (not resume task)
	if !strings.Contains(output, "# Fix Bug (Initial)") {
		t.Errorf("normal mode: normal task content not found in stdout")
	}
	if strings.Contains(output, "# Fix Bug (Resume)") {
		t.Errorf("normal mode: resume task content should not be in stdout")
	}

	// Test 2: Run in resume mode (with -s resume=true selector)
	// Capture stdout and stderr separately to verify bootstrap scripts don't run
	cmd = exec.Command("go", "run", wd, "-C", dirs.tmpDir, "-s", "resume=true", "fix-bug-resume")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("failed to run binary in resume mode: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	output = stdout.String()
	stderrOutput = stderr.String()

	// In resume mode, rules should NOT be included
	if strings.Contains(output, "# Coding Standards") {
		t.Errorf("resume mode: rule content should not be in stdout")
	}

	// In resume mode, bootstrap scripts should NOT run
	if strings.Contains(stderrOutput, "RULE_BOOTSTRAP_RAN") {
		t.Errorf("resume mode: rule bootstrap script should not run (found in stderr: %s)", stderrOutput)
	}

	// In resume mode, should use the resume task
	if !strings.Contains(output, "# Fix Bug (Resume)") {
		t.Errorf("resume mode: resume task content not found in stdout")
	}
	if strings.Contains(output, "# Fix Bug (Initial)") {
		t.Errorf("resume mode: normal task content should not be in stdout")
	}

	// Test 3: Run in resume mode (with -r flag)
	cmd = exec.Command("go", "run", wd, "-C", dirs.tmpDir, "-r", "fix-bug-resume")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("failed to run binary in resume mode with -r flag: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	output = stdout.String()
	stderrOutput = stderr.String()

	// In resume mode with -r flag, rules should NOT be included
	if strings.Contains(output, "# Coding Standards") {
		t.Errorf("resume mode (-r flag): rule content should not be in stdout")
	}

	// In resume mode with -r flag, bootstrap scripts should NOT run
	if strings.Contains(stderrOutput, "RULE_BOOTSTRAP_RAN") {
		t.Errorf("resume mode (-r flag): rule bootstrap script should not run (found in stderr: %s)", stderrOutput)
	}

	// In resume mode with -r flag, should use the resume task
	if !strings.Contains(output, "# Fix Bug (Resume)") {
		t.Errorf("resume mode (-r flag): resume task content not found in stdout")
	}
	if strings.Contains(output, "# Fix Bug (Initial)") {
		t.Errorf("resume mode (-r flag): normal task content should not be in stdout")
	}
}

func TestRemoteRuleFromHTTP(t *testing.T) {
	// Create a remote directory structure to serve
	remoteDir := t.TempDir()
	rulesDir := filepath.Join(remoteDir, ".agents", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create remote rules dir: %v", err)
	}

	// Create a remote rule file
	remoteRuleFile := filepath.Join(rulesDir, "remote-rule.md")
	remoteRuleContent := `---
---
# Remote Rule

This is a rule loaded from a remote directory.
`
	if err := os.WriteFile(remoteRuleFile, []byte(remoteRuleContent), 0o644); err != nil {
		t.Fatalf("failed to write remote rule file: %v", err)
	}

	// Create a temporary directory structure for local task
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	createStandardTask(t, tasksDir, "test-task")

	// Run the program with remote directory (using file:// URL)
	remoteURL := "file://" + remoteDir
	output := runTool(t, "-C", tmpDir, "-d", remoteURL, "test-task")

	// Check that remote rule content is present
	if !strings.Contains(output, "# Remote Rule") {
		t.Errorf("remote rule content not found in stdout")
	}
	if !strings.Contains(output, "This is a rule loaded from a remote directory") {
		t.Errorf("remote rule description not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestPrintTaskFrontmatter(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file
	ruleFile := filepath.Join(dirs.rulesDir, "test-rule.md")
	ruleContent := `---
language: go
---
# Test Rule

This is a test rule.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a task file with frontmatter
	taskFile := filepath.Join(dirs.tasksDir, "test-task.md")
	taskContent := `---
task_name: test-task
author: tester
version: 1.0
---
# Test Task

This is a test task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Test that frontmatter is always printed
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	lines := strings.Split(output, "\n")

	// Find the first non-log line (skip lines starting with "time=")
	var firstContentLine string
	for _, line := range lines {
		if !strings.HasPrefix(line, "time=") {
			firstContentLine = line
			break
		}
	}

	// First content line should be frontmatter delimiter
	if firstContentLine != "---" {
		t.Errorf("expected first content line to be '---', got %q", firstContentLine)
	}

	// Should contain task frontmatter fields
	if !strings.Contains(output, "task_name: test-task") {
		t.Errorf("task frontmatter field 'task_name' not found in output")
	}
	if !strings.Contains(output, "author: tester") {
		t.Errorf("task frontmatter field 'author' not found in output")
	}
	if !strings.Contains(output, "version: 1.0") {
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
	if !strings.Contains(output, "# Test Rule") {
		t.Errorf("rule content not found in output")
	}

	// Task content should appear after rules
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in output")
	}

	// Verify order: frontmatter should come before rules, rules before task content
	frontmatterIdx := strings.Index(output, "task_name: test-task")
	ruleIdx := strings.Index(output, "# Test Rule")
	taskIdx := strings.Index(output, "# Test Task")

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
	if strings.Contains(output, "language: go") {
		t.Errorf("rule frontmatter should not be printed in output")
	}
}

func TestTaskBootstrapFromFile(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a simple task file
	taskFile := filepath.Join(dirs.tasksDir, "test-task.md")
	taskContent := `---
task_name: test-task
---
# Test Task

This is a test task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Note: Tasks no longer have bootstrap scripts - only rules do

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that task content is present
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
}

func TestTaskBootstrapFileNotRequired(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a task file WITHOUT a bootstrap
	taskFile := filepath.Join(dirs.tasksDir, "no-bootstrap-task.md")
	taskContent := `---
task_name: no-bootstrap-task
---
# Task Without Bootstrap

This task has no bootstrap script.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program - should succeed without a bootstrap file
	output := runTool(t, "-C", dirs.tmpDir, "no-bootstrap-task")

	// Check that task content is present
	if !strings.Contains(output, "# Task Without Bootstrap") {
		t.Errorf("task content not found in stdout")
	}
}

func TestTaskBootstrapWithRuleBootstrap(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file with bootstrap
	ruleFile := filepath.Join(dirs.rulesDir, "setup.md")
	ruleContent := `---
---
# Setup Rule

Setup instructions.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	ruleBootstrapFile := filepath.Join(dirs.rulesDir, "setup-bootstrap")
	ruleBootstrapContent := `#!/bin/bash
echo "Running rule bootstrap"
`
	if err := os.WriteFile(ruleBootstrapFile, []byte(ruleBootstrapContent), 0o755); err != nil {
		t.Fatalf("failed to write rule bootstrap file: %v", err)
	}

	// Create a task file (tasks no longer have bootstrap scripts)
	taskFile := filepath.Join(dirs.tasksDir, "deploy-task.md")
	taskContent := `---
task_name: deploy-task
---
# Deploy Task

Deploy instructions.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "deploy-task")

	// Check that rule bootstrap ran (rules still have bootstrap scripts)
	if !strings.Contains(output, "Running rule bootstrap") {
		t.Errorf("rule bootstrap output not found in stdout")
	}

	// Check that both rule and task contents are present
	if !strings.Contains(output, "# Setup Rule") {
		t.Errorf("rule content not found in stdout")
	}
	if !strings.Contains(output, "# Deploy Task") {
		t.Errorf("task content not found in stdout")
	}

	// Verify the order: rule bootstrap -> rule content -> task content
	ruleBootstrapIdx := strings.Index(output, "Running rule bootstrap")
	ruleContentIdx := strings.Index(output, "# Setup Rule")
	taskContentIdx := strings.Index(output, "# Deploy Task")

	if ruleBootstrapIdx > ruleContentIdx {
		t.Errorf("rule bootstrap should run before rule content")
	}
	if ruleContentIdx > taskContentIdx {
		t.Errorf("rule content should appear before task content")
	}
}

func TestManifestFile(t *testing.T) {
	// Create main project directory
	mainDir := t.TempDir()
	mainRulesDir := filepath.Join(mainDir, ".agents", "rules")
	mainTasksDir := filepath.Join(mainDir, ".agents", "tasks")

	if err := os.MkdirAll(mainRulesDir, 0o755); err != nil {
		t.Fatalf("failed to create main rules dir: %v", err)
	}
	if err := os.MkdirAll(mainTasksDir, 0o755); err != nil {
		t.Fatalf("failed to create main tasks dir: %v", err)
	}

	// Create a task file in the main directory
	taskFile := filepath.Join(mainTasksDir, "test-task.md")
	taskContent := `---
task_name: test-task
---
# Test Task

This is a test task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Create a rule file in the main directory (should be included)
	mainRuleFile := filepath.Join(mainRulesDir, "main-rule.md")
	mainRuleContent := `---
---
# Main Rule

This rule is in the main project.
`
	if err := os.WriteFile(mainRuleFile, []byte(mainRuleContent), 0o644); err != nil {
		t.Fatalf("failed to write main rule file: %v", err)
	}

	// Create a remote directory with rules
	remoteDir := t.TempDir()
	remoteRulesDir := filepath.Join(remoteDir, ".agents", "rules")
	if err := os.MkdirAll(remoteRulesDir, 0o755); err != nil {
		t.Fatalf("failed to create remote rules dir: %v", err)
	}

	remoteRuleFile := filepath.Join(remoteRulesDir, "remote-rule.md")
	remoteRuleContent := `---
---
# Remote Rule

This rule is from a remote directory.
`
	if err := os.WriteFile(remoteRuleFile, []byte(remoteRuleContent), 0o644); err != nil {
		t.Fatalf("failed to write remote rule file: %v", err)
	}

	// Create a manifest file that references the remote directory
	manifestFile := filepath.Join(t.TempDir(), "manifest.txt")
	manifestContent := fmt.Sprintf("file://%s\n", remoteDir)
	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("failed to write manifest file: %v", err)
	}

	// Run the tool with the manifest file
	output := runTool(t, "-C", mainDir, "-m", "file://"+manifestFile, "test-task")

	// Check that the main rule is included
	if !strings.Contains(output, "# Main Rule") {
		t.Errorf("main rule not found in stdout. Output:\n%s", output)
	}

	// Check that the remote rule from the manifest is included
	if !strings.Contains(output, "# Remote Rule") {
		t.Errorf("remote rule from manifest not found in stdout. Output:\n%s", output)
	}

	// Check that the task is included
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task not found in stdout. Output:\n%s", output)
	}
}

// TestSingleExpansion verifies that content is expanded only once in the full flow
func TestSingleExpansion(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a task that uses a parameter with expansion syntax
	taskFile := filepath.Join(dirs.tasksDir, "test-expand.md")
	taskContent := `Task with parameter: ${param1}

And a value that looks like expansion syntax but should not be expanded: ${"nested"}`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to create task file: %v", err)
	}

	// Run with param1 set to a value that contains expansion syntax
	output := runTool(t, "-C", dirs.tmpDir, "-p", "param1=!`echo hello`", "test-expand")

	// The param1 should be replaced with the literal string "!`echo hello`"
	// It should NOT be expanded again (that would execute the command)
	if !strings.Contains(output, "!`echo hello`") {
		t.Errorf("Expected param1 to be replaced with literal value, got: %s", output)
	}

	// Verify "hello" is not in output (which would indicate the command was executed)
	// Note: there may be other "hello" strings, so check for the specific context
	if strings.Contains(output, "Task with parameter: hello") {
		t.Errorf("Parameter value was re-expanded (command was executed), got: %s", output)
	}
}

// TestCommandExpansionOnce verifies that command files are expanded only once
func TestCommandExpansionOnce(t *testing.T) {
	dirs := setupTestDirs(t)
	commandsDir := filepath.Join(dirs.tmpDir, ".agents", "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}

	// Create a command file with a parameter
	commandFile := filepath.Join(commandsDir, "test-cmd.md")
	commandContent := `Command param: ${cmd_param}`
	if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
		t.Fatalf("failed to create command file: %v", err)
	}

	// Create a task that calls the command with a param containing expansion syntax
	taskFile := filepath.Join(dirs.tasksDir, "test-cmd-task.md")
	taskContent := `/test-cmd cmd_param="!` + "`echo injected`" + `"`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to create task file: %v", err)
	}

	// Run the task
	output := runTool(t, "-C", dirs.tmpDir, "test-cmd-task")

	// The command parameter should be replaced with the literal string "!`echo injected`"
	// It should NOT be expanded again (that would execute the command)
	if !strings.Contains(output, "!`echo injected`") {
		t.Errorf("Expected command param to be replaced with literal value, got: %s", output)
	}

	// Verify "injected" is not in output (which would indicate the command was executed)
	if strings.Contains(output, "Command param: injected") {
		t.Errorf("Command parameter value was re-expanded (command was executed), got: %s", output)
	}
}

func TestWriteRulesOption(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file
	ruleFile := filepath.Join(dirs.rulesDir, "test-rule.md")
	ruleContent := `---
language: go
---
# Test Rule

This is a test rule that should be written to the user rules path.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(dirs.tasksDir, "test-task.md")
	taskContent := `---
task_name: test-task
---
# Test Task

This is the task prompt.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Create a temporary home directory for this test
	tmpHome := t.TempDir()

	// Run with -w flag and -a copilot
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	cmd := exec.Command("go", "run", wd, "-C", dirs.tmpDir, "-a", "copilot", "-w", "test-task")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Build a clean environment that explicitly sets GOMODCACHE outside tmpDir
	// to avoid permission issues during cleanup
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		gomodcache = filepath.Join(os.Getenv("HOME"), "go", "pkg", "mod")
	}
	cmd.Env = append(os.Environ(),
		"HOME="+tmpHome,
		"GOMODCACHE="+gomodcache,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run binary: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	stderrOutput := stderr.String()

	// Verify that the rules were NOT printed to stdout
	if strings.Contains(output, "# Test Rule") {
		t.Errorf("rules should not be in stdout when using -w flag")
	}

	// Verify that the task IS printed to stdout
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}
	if !strings.Contains(output, "This is the task prompt.") {
		t.Errorf("task description not found in stdout")
	}

	// Verify that rules were written to the user rules path
	expectedRulesPath := filepath.Join(tmpHome, ".github", "agents", "AGENTS.md")
	rulesFileContent, err := os.ReadFile(expectedRulesPath)
	if err != nil {
		t.Fatalf("failed to read rules file at %s: %v", expectedRulesPath, err)
	}

	rulesStr := string(rulesFileContent)
	if !strings.Contains(rulesStr, "# Test Rule") {
		t.Errorf("rules file does not contain rule content")
	}
	if !strings.Contains(rulesStr, "This is a test rule that should be written to the user rules path.") {
		t.Errorf("rules file does not contain rule description")
	}

	// Verify that the logger reported where rules were written
	if !strings.Contains(stderrOutput, "Rules written") {
		t.Errorf("stderr should contain 'Rules written' message")
	}
}

func TestWriteRulesOptionWithoutAgent(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a simple task file without agent field
	createStandardTask(t, dirs.tasksDir, "test-task")

	// Run with -w flag but WITHOUT -a flag and task has no agent field (should fail)
	output, err := runToolWithError("-C", dirs.tmpDir, "-w", "test-task")
	if err == nil {
		t.Errorf("expected error when using -w without agent (from task or -a flag), but command succeeded")
	}

	// Verify error message
	if !strings.Contains(output, "-w flag requires an agent") {
		t.Errorf("expected error message about requiring an agent, got: %s", output)
	}
}

func TestWriteRulesOptionWithResumeMode(t *testing.T) {
	dirs := setupTestDirs(t)

	// Create a rule file
	ruleFile := filepath.Join(dirs.rulesDir, "test-rule.md")
	ruleContent := `---
language: go
---
# Test Rule

This is a test rule that should NOT be written in resume mode.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a resume task file
	taskFile := filepath.Join(dirs.tasksDir, "test-task-resume.md")
	taskContent := `---
resume: true
---
# Test Task Resume

This is the task prompt for resume mode.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Create a temporary home directory for this test
	tmpHome := t.TempDir()

	// Run with -w flag, -r flag (resume mode), and -a copilot
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	cmd := exec.Command("go", "run", wd, "-C", dirs.tmpDir, "-a", "copilot", "-w", "-r", "test-task-resume")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Build a clean environment that explicitly sets GOMODCACHE outside tmpDir
	// to avoid permission issues during cleanup
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		gomodcache = filepath.Join(os.Getenv("HOME"), "go", "pkg", "mod")
	}
	cmd.Env = append(os.Environ(),
		"HOME="+tmpHome,
		"GOMODCACHE="+gomodcache,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run binary: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	stderrOutput := stderr.String()

	// Verify that the rules were NOT printed to stdout
	if strings.Contains(output, "# Test Rule") {
		t.Errorf("rules should not be in stdout when using -w flag with resume mode")
	}

	// Verify that the task IS printed to stdout
	if !strings.Contains(output, "# Test Task Resume") {
		t.Errorf("task content not found in stdout")
	}
	if !strings.Contains(output, "This is the task prompt for resume mode.") {
		t.Errorf("task description not found in stdout")
	}

	// Verify that NO rules file was created in resume mode
	expectedRulesPath := filepath.Join(tmpHome, ".github", "agents", "AGENTS.md")
	if _, err := os.Stat(expectedRulesPath); err == nil {
		t.Errorf("rules file should NOT be created in resume mode with -w flag, but found at %s", expectedRulesPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("unexpected error checking for rules file: %v", err)
	}

	// Verify that the logger did NOT report writing rules
	if strings.Contains(stderrOutput, "Rules written") {
		t.Errorf("stderr should NOT contain 'Rules written' message in resume mode")
	}
}

// TestLocalDirectoryNotDeleted verifies that local directories passed via -d flag
// are not deleted after the command completes.
func TestLocalDirectoryNotDeleted(t *testing.T) {
	// Create a local directory with a rule file and a marker file
	localDir := t.TempDir()
	rulesDir := filepath.Join(localDir, ".agents", "rules")

	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "local-rule.md")
	ruleContent := `---
language: go
---
# Local Rule

This is a rule from a local directory.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a marker file to verify the directory is not deleted
	markerFile := filepath.Join(localDir, "marker.txt")
	if err := os.WriteFile(markerFile, []byte("marker"), 0o644); err != nil {
		t.Fatalf("failed to write marker file: %v", err)
	}

	// Create a temporary directory for the task
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	createStandardTask(t, tasksDir, "test-task")

	// Run the program with local directory using file:// URL
	localURL := "file://" + localDir
	output := runTool(t, "-C", tmpDir, "-d", localURL, "test-task")

	// Check that local rule content is present
	if !strings.Contains(output, "# Local Rule") {
		t.Errorf("local rule content not found in stdout")
	}
	if !strings.Contains(output, "This is a rule from a local directory") {
		t.Errorf("local rule description not found in stdout")
	}

	// Verify the marker file still exists (directory was not deleted)
	if _, err := os.Stat(markerFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("marker file was deleted, indicating local directory was deleted")
		} else {
			t.Fatalf("unexpected error checking marker file: %v", err)
		}
	}

	// Verify the rule file still exists
	if _, err := os.Stat(ruleFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("rule file was deleted, indicating local directory was deleted")
		} else {
			t.Fatalf("unexpected error checking rule file: %v", err)
		}
	}
}

// TestLocalDirectoryWithoutProtocol verifies that local directories passed
// without the file:// protocol are not deleted.
func TestLocalDirectoryWithoutProtocol(t *testing.T) {
	// Create a local directory with a rule file and a marker file
	localDir := t.TempDir()
	rulesDir := filepath.Join(localDir, ".agents", "rules")

	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "local-rule.md")
	ruleContent := `---
language: go
---
# Local Rule

This is a rule from a local directory without protocol.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0o644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a marker file to verify the directory is not deleted
	markerFile := filepath.Join(localDir, "marker.txt")
	if err := os.WriteFile(markerFile, []byte("marker"), 0o644); err != nil {
		t.Fatalf("failed to write marker file: %v", err)
	}

	// Create a temporary directory for the task
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	createStandardTask(t, tasksDir, "test-task")

	// Run the program with local directory using absolute path (no protocol)
	output := runTool(t, "-C", tmpDir, "-d", localDir, "test-task")

	// Check that local rule content is present
	if !strings.Contains(output, "# Local Rule") {
		t.Errorf("local rule content not found in stdout")
	}
	if !strings.Contains(output, "This is a rule from a local directory without protocol") {
		t.Errorf("local rule description not found in stdout")
	}

	// Verify the marker file still exists (directory was not deleted)
	if _, err := os.Stat(markerFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("marker file was deleted, indicating local directory was deleted")
		} else {
			t.Fatalf("unexpected error checking marker file: %v", err)
		}
	}

	// Verify the rule file still exists
	if _, err := os.Stat(ruleFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("rule file was deleted, indicating local directory was deleted")
		} else {
			t.Fatalf("unexpected error checking rule file: %v", err)
		}
	}
}

// TestTaskWithEmptyContent verifies that tasks with only frontmatter
// and empty or whitespace-only content are handled gracefully.
func TestTaskWithEmptyContent(t *testing.T) {
	tests := []struct {
		name        string
		taskName    string
		taskContent string
	}{
		{
			name:     "empty content",
			taskName: "empty-task",
			taskContent: `---
task_name: empty-task
---
`,
		},
		{
			name:     "single newline",
			taskName: "newline-task",
			taskContent: `---
task_name: newline-task
---

`,
		},
		{
			name:     "multiple newlines",
			taskName: "newlines-task",
			taskContent: `---
task_name: newlines-task
---


`,
		},
		{
			name:     "whitespace only",
			taskName: "whitespace-task",
			taskContent: `---
task_name: whitespace-task
---
   
	
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs := setupTestDirs(t)

			// Create task file with empty or whitespace content
			// Use the task name in the filename
			taskFile := filepath.Join(dirs.tasksDir, tt.taskName+".md")
			if err := os.WriteFile(taskFile, []byte(tt.taskContent), 0o644); err != nil {
				t.Fatalf("failed to write task file: %v", err)
			}

			// Run the program - should not error
			output := runTool(t, "-C", dirs.tmpDir, tt.taskName)

			// The output should contain the frontmatter but not fail
			// (exact output format may vary, but it should succeed)
			if output == "" {
				t.Errorf("expected some output, got empty string")
			}
		})
	}
}
