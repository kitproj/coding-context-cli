package main

import (
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
	openCodeCommandDir := filepath.Join(tmpDir, ".opencode", "command")

	if err := os.MkdirAll(openCodeCommandDir, 0o755); err != nil {
		t.Fatalf("failed to create opencode command dir: %v", err)
	}

	// Create a task file in .opencode/command
	taskFile := filepath.Join(openCodeCommandDir, "fix-bug.md")
	taskContent := `---
task_name: fix-bug
---
# Fix Bug Command

This is an OpenCode command task for fixing bugs.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program
	output := runTool(t, "-C", tmpDir, "fix-bug")

	// Check that task content is present
	if !strings.Contains(output, "# Fix Bug Command") {
		t.Errorf("OpenCode command task content not found in stdout")
	}
	if !strings.Contains(output, "This is an OpenCode command task for fixing bugs.") {
		t.Errorf("OpenCode command task description not found in stdout")
	}
}

func TestTaskSelectionByFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
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
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the program with task name matching the task_name frontmatter, not filename
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

func TestMultipleTasksWithSameNameError(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
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
	if err := os.WriteFile(taskFile1, []byte(taskContent1), 0o644); err != nil {
		t.Fatalf("failed to write task file 1: %v", err)
	}

	taskFile2 := filepath.Join(tasksDir, "file2.md")
	taskContent2 := `---
task_name: duplicate-task
---
# Task File 2

This is the second file.
`
	if err := os.WriteFile(taskFile2, []byte(taskContent2), 0o644); err != nil {
		t.Fatalf("failed to write task file 2: %v", err)
	}

	// Run the program - should fail with an error about duplicate task names
	output, err := runToolWithError("-C", tmpDir, "duplicate-task")
	if err == nil {
		t.Fatalf("expected program to fail with duplicate task names, but it succeeded")
	}

	// Check that error message mentions multiple task files
	if !strings.Contains(output, "multiple task files found") {
		t.Errorf("expected error about multiple task files, got: %s", output)
	}
}

func TestTaskSelectionWithSelectors(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
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
	if err := os.WriteFile(taskFile1, []byte(taskContent1), 0o644); err != nil {
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
	if err := os.WriteFile(taskFile2, []byte(taskContent2), 0o644); err != nil {
		t.Fatalf("failed to write production task file: %v", err)
	}

	// Run the program with selector for staging
	output := runTool(t, "-C", tmpDir, "-s", "environment=staging", "deploy")

	// Check that staging task content is present
	if !strings.Contains(output, "# Deploy to Staging") {
		t.Errorf("staging task content not found in stdout")
	}
	if strings.Contains(output, "# Deploy to Production") {
		t.Errorf("production task content should not be in stdout when selecting staging")
	}

	// Run the program with selector for production
	output = runTool(t, "-C", tmpDir, "-s", "environment=production", "deploy")

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

	// Create a normal task file (with resume: false)
	normalTaskFile := filepath.Join(dirs.tasksDir, "fix-bug-initial.md")
	normalTaskContent := `---
task_name: fix-bug
resume: false
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
task_name: fix-bug
resume: true
---
# Fix Bug (Resume)

This is the resume task prompt for continuing the bug fix.
`
	if err := os.WriteFile(resumeTaskFile, []byte(resumeTaskContent), 0o644); err != nil {
		t.Fatalf("failed to write resume task file: %v", err)
	}

	// Test 1: Run in normal mode (with -s resume=false to select non-resume task)
	output := runTool(t, "-C", dirs.tmpDir, "-s", "resume=false", "fix-bug")

	// In normal mode, rules should be included
	if !strings.Contains(output, "# Coding Standards") {
		t.Errorf("normal mode: rule content not found in stdout")
	}

	// In normal mode, should use the normal task (not resume task)
	if !strings.Contains(output, "# Fix Bug (Initial)") {
		t.Errorf("normal mode: normal task content not found in stdout")
	}
	if strings.Contains(output, "# Fix Bug (Resume)") {
		t.Errorf("normal mode: resume task content should not be in stdout")
	}

	// Test 2: Run in resume mode (with -r flag)
	output = runTool(t, "-C", dirs.tmpDir, "-r", "fix-bug")

	// In resume mode, rules should NOT be included
	if strings.Contains(output, "# Coding Standards") {
		t.Errorf("resume mode: rule content should not be in stdout")
	}

	// In resume mode, should use the resume task
	if !strings.Contains(output, "# Fix Bug (Resume)") {
		t.Errorf("resume mode: resume task content not found in stdout")
	}
	if strings.Contains(output, "# Fix Bug (Initial)") {
		t.Errorf("resume mode: normal task content should not be in stdout")
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

	// Test without -t flag (should not print frontmatter)
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Should not contain frontmatter delimiters in the main output
	lines := strings.Split(output, "\n")
	if len(lines) > 0 && lines[0] == "---" {
		t.Errorf("frontmatter should not be printed without -t flag")
	}
	// Task content should be present
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in output without -t flag")
	}
	// Rule content should be present
	if !strings.Contains(output, "# Test Rule") {
		t.Errorf("rule content not found in output without -t flag")
	}

	// Test with -t flag (should print frontmatter)
	output = runTool(t, "-C", dirs.tmpDir, "-t", "test-task")

	lines = strings.Split(output, "\n")

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
		t.Errorf("rule content not found in output with -t flag")
	}

	// Task content should appear after rules
	if !strings.Contains(output, "# Test Task") {
		t.Errorf("task content not found in output with -t flag")
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

	// Create a bootstrap file for the task (test-task.md -> test-task-bootstrap)
	bootstrapFile := filepath.Join(dirs.tasksDir, "test-task-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Running task bootstrap"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0o755); err != nil {
		t.Fatalf("failed to write task bootstrap file: %v", err)
	}

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "test-task")

	// Check that bootstrap output appears before task content
	bootstrapIdx := strings.Index(output, "Running task bootstrap")
	taskIdx := strings.Index(output, "# Test Task")

	if bootstrapIdx == -1 {
		t.Errorf("task bootstrap output not found in stdout")
	}
	if taskIdx == -1 {
		t.Errorf("task content not found in stdout")
	}
	if bootstrapIdx != -1 && taskIdx != -1 && bootstrapIdx > taskIdx {
		t.Errorf("task bootstrap output should appear before task content")
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

	// Create a task file with bootstrap
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

	taskBootstrapFile := filepath.Join(dirs.tasksDir, "deploy-task-bootstrap")
	taskBootstrapContent := `#!/bin/bash
echo "Running task bootstrap"
`
	if err := os.WriteFile(taskBootstrapFile, []byte(taskBootstrapContent), 0o755); err != nil {
		t.Fatalf("failed to write task bootstrap file: %v", err)
	}

	// Run the program
	output := runTool(t, "-C", dirs.tmpDir, "deploy-task")

	// Check that both bootstrap scripts ran
	if !strings.Contains(output, "Running rule bootstrap") {
		t.Errorf("rule bootstrap output not found in stdout")
	}
	if !strings.Contains(output, "Running task bootstrap") {
		t.Errorf("task bootstrap output not found in stdout")
	}

	// Check that both rule and task contents are present
	if !strings.Contains(output, "# Setup Rule") {
		t.Errorf("rule content not found in stdout")
	}
	if !strings.Contains(output, "# Deploy Task") {
		t.Errorf("task content not found in stdout")
	}

	// Verify the order: rule bootstrap -> rule content -> task bootstrap -> task content
	ruleBootstrapIdx := strings.Index(output, "Running rule bootstrap")
	ruleContentIdx := strings.Index(output, "# Setup Rule")
	taskBootstrapIdx := strings.Index(output, "Running task bootstrap")
	taskContentIdx := strings.Index(output, "# Deploy Task")

	if ruleBootstrapIdx > ruleContentIdx {
		t.Errorf("rule bootstrap should run before rule content")
	}
	if ruleContentIdx > taskBootstrapIdx {
		t.Errorf("rule content should appear before task bootstrap")
	}
	if taskBootstrapIdx > taskContentIdx {
		t.Errorf("task bootstrap should run before task content")
	}
}
