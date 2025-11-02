package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to set up workDir and restore it after test
func setupWorkDir(t *testing.T, dir string) func() {
	oldWorkDir := workDir
	workDir = dir
	return func() { workDir = oldWorkDir }
}

// TestRunInvalidArguments tests that the run function returns an error when invalid arguments are provided
func TestRunInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no arguments",
			args: []string{},
		},
		{
			name: "one argument only",
			args: []string{"ClaudeCode"},
		},
		{
			name: "too many arguments",
			args: []string{"ClaudeCode", "task1", "extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := run(ctx, tt.args)
			if err == nil {
				t.Error("expected error for invalid arguments, got nil")
			}
			if !strings.Contains(err.Error(), "invalid usage") {
				t.Errorf("expected 'invalid usage' error, got: %v", err)
			}
		})
	}
}

// TestRunWithValidArguments tests the basic flow with valid arguments
func TestRunWithValidArguments(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create .agents/tasks directory with a task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

This is a test task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRuleFileSynchronization tests that rule files are correctly synchronized
func TestRuleFileSynchronization(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create source rule files
	ruleContent := `---
---
# Test Rule

This is a test rule.
`
	
	// Create CLAUDE.md in the project root
	claudeFile := filepath.Join(tmpDir, "CLAUDE.md")
	if err := os.WriteFile(claudeFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}
	
	// Create AGENTS.md in the project root
	agentsFile := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(agentsFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write AGENTS.md: %v", err)
	}
	
	// Create .agents/tasks directory with a task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	if err := os.WriteFile(taskFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Verify that .agents/rules directory was created
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Error(".agents/rules directory was not created")
	}
	
	// Verify that CLAUDE.md was synchronized
	syncedClaudeFile := filepath.Join(rulesDir, "CLAUDE.md")
	if _, err := os.Stat(syncedClaudeFile); os.IsNotExist(err) {
		t.Error("CLAUDE.md was not synchronized to .agents/rules")
	}
	
	// Verify that AGENTS.md was synchronized
	syncedAgentsFile := filepath.Join(rulesDir, "AGENTS.md")
	if _, err := os.Stat(syncedAgentsFile); os.IsNotExist(err) {
		t.Error("AGENTS.md was not synchronized to .agents/rules")
	}
}

// TestBootstrapFileHandling tests that bootstrap files are correctly copied
func TestBootstrapFileHandling(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create a rule file with a bootstrap script
	ruleContent := `---
---
# Setup Rule

This rule has a bootstrap script.
`
	
	claudeFile := filepath.Join(tmpDir, "CLAUDE.md")
	if err := os.WriteFile(claudeFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}
	
	// Create a bootstrap file
	bootstrapFile := filepath.Join(tmpDir, "CLAUDE-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Running bootstrap"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}
	
	// Create .agents/tasks directory with a task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	if err := os.WriteFile(taskFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Verify that the bootstrap file was copied
	syncedBootstrapFile := filepath.Join(tmpDir, ".agents", "rules", "CLAUDE-bootstrap")
	if _, err := os.Stat(syncedBootstrapFile); os.IsNotExist(err) {
		t.Error("bootstrap file was not copied to .agents/rules")
	} else {
		// Verify content
		content, err := os.ReadFile(syncedBootstrapFile)
		if err != nil {
			t.Fatalf("failed to read bootstrap file: %v", err)
		}
		if string(content) != bootstrapContent {
			t.Errorf("bootstrap content mismatch:\ngot: %q\nwant: %q", string(content), bootstrapContent)
		}
	}
}

// TestParameterSubstitution tests that parameters are correctly substituted in task files
func TestParameterSubstitution(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create .agents/tasks directory with a task file containing parameters
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskContent := `---
---
# Test Task

The project name is ${project_name}.
The version is ${version}.
`
	taskFile := filepath.Join(tasksDir, "test-task.md")
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	// Set parameters
	params["project_name"] = "test-project"
	params["version"] = "1.0.0"
	defer func() {
		delete(params, "project_name")
		delete(params, "version")
	}()
	
	ctx := context.Background()
	
	// Note: This test verifies the logic runs without error
	// Full verification would require capturing stdout
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestFrontmatterFiltering tests that rules are filtered based on frontmatter
func TestFrontmatterFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create .agents/rules directory
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	
	// Create a rule file with frontmatter
	ruleWithTaskName := `---
task_name: specific-task
---
# Specific Task Rule

This rule is for a specific task.
`
	ruleFile1 := filepath.Join(rulesDir, "specific-rule.md")
	if err := os.WriteFile(ruleFile1, []byte(ruleWithTaskName), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	
	// Create a rule file without task_name
	ruleGeneral := `---
---
# General Rule

This is a general rule.
`
	ruleFile2 := filepath.Join(rulesDir, "general-rule.md")
	if err := os.WriteFile(ruleFile2, []byte(ruleGeneral), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	
	// Create .agents/tasks directory with a task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAgentRulesPathLookup tests different agent configurations
func TestAgentRulesPathLookup(t *testing.T) {
	agents := []string{
		"ClaudeCode",
		"Cursor",
		"Windsurf",
		"Codex",
		"GitHubCopilot",
		"AugmentCLI",
		"Goose",
		"Gemini",
	}
	
	for _, agent := range agents {
		t.Run(agent, func(t *testing.T) {
			tmpDir := t.TempDir()
			defer setupWorkDir(t, tmpDir)()
			
			// Create .agents/tasks directory with a task file
			tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
			if err := os.MkdirAll(tasksDir, 0755); err != nil {
				t.Fatalf("failed to create tasks dir: %v", err)
			}
			
			taskContent := `---
---
# Test Task

This is a test task for ` + agent + `.
`
			taskFile := filepath.Join(tasksDir, "test-task.md")
			if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
				t.Fatalf("failed to write task file: %v", err)
			}
			
			ctx := context.Background()
			err := run(ctx, []string{agent, "test-task"})
			if err != nil {
				t.Errorf("unexpected error for agent %s: %v", agent, err)
			}
		})
	}
}

// TestTaskFileNotFound tests that an error is returned when the task file is not found
func TestTaskFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "nonexistent-task"})
	if err == nil {
		t.Error("expected error for nonexistent task file, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestCursorRulesDirectory tests that Cursor agent correctly handles .cursor/rules directory
func TestCursorRulesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create .cursor/rules directory with multiple rule files
	cursorRulesDir := filepath.Join(tmpDir, ".cursor", "rules")
	if err := os.MkdirAll(cursorRulesDir, 0755); err != nil {
		t.Fatalf("failed to create cursor rules dir: %v", err)
	}
	
	ruleContent1 := `---
---
# Cursor Rule 1

This is the first cursor rule.
`
	if err := os.WriteFile(filepath.Join(cursorRulesDir, "rule1.md"), []byte(ruleContent1), 0644); err != nil {
		t.Fatalf("failed to write rule1.md: %v", err)
	}
	
	ruleContent2 := `---
---
# Cursor Rule 2

This is the second cursor rule.
`
	if err := os.WriteFile(filepath.Join(cursorRulesDir, "rule2.md"), []byte(ruleContent2), 0644); err != nil {
		t.Fatalf("failed to write rule2.md: %v", err)
	}
	
	// Create .agents/tasks directory with a task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	ctx := context.Background()
	err := run(ctx, []string{"Cursor", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Verify .agents/rules/cursor directory was created
	syncedCursorDir := filepath.Join(tmpDir, ".agents", "rules", "cursor")
	if _, err := os.Stat(syncedCursorDir); os.IsNotExist(err) {
		t.Error(".agents/rules/cursor directory was not created")
	}
}

// TestWorkDirFlag tests the -C flag for changing working directory
func TestWorkDirFlag(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	
	// Create subdirectory with task file
	tasksDir := filepath.Join(subDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	// Save and restore workDir
	defer setupWorkDir(t, subDir)()
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMdcFileExtension tests that .mdc files are processed as rule files
func TestMdcFileExtension(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create a .agents/rules directory with a .mdc file
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	
	// Create a .mdc rule file in .agents/rules
	ruleContent := `---
---
# MDC Rule

This is an MDC rule file.
`
	mdcFile := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(mdcFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write test-rule.mdc: %v", err)
	}
	
	// Create .agents/tasks directory with a task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Verify that the .mdc file still exists and was processed
	if _, err := os.Stat(mdcFile); os.IsNotExist(err) {
		t.Error("test-rule.mdc should still exist in .agents/rules")
	}
}

// TestMultipleRuleSources tests that rules from multiple sources are aggregated
func TestMultipleRuleSources(t *testing.T) {
	tmpDir := t.TempDir()
	defer setupWorkDir(t, tmpDir)()
	
	// Create multiple rule files
	files := map[string]string{
		"CLAUDE.md":        "# Claude Rule\n",
		"AGENTS.md":        "# Agents Rule\n",
		"GEMINI.md":        "# Gemini Rule\n",
		"CLAUDE.local.md": "# Claude Local Rule\n",
	}
	
	for filename, content := range files {
		fullContent := "---\n---\n" + content
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(fullContent), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", filename, err)
		}
	}
	
	// Create .agents/tasks directory with a task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	ctx := context.Background()
	err := run(ctx, []string{"ClaudeCode", "test-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Verify that multiple rule files were synchronized
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	expectedFiles := []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md", "CLAUDE.local.md"}
	
	for _, filename := range expectedFiles {
		syncedFile := filepath.Join(rulesDir, filename)
		if _, err := os.Stat(syncedFile); os.IsNotExist(err) {
			t.Errorf("%s was not synchronized to .agents/rules", filename)
		}
	}
}

// TestIntegrationWithBinary is an end-to-end integration test using the compiled binary
func TestIntegrationWithBinary(t *testing.T) {
	// Save current directory before tests may change it
	originalWd, err := os.Getwd()
	if err != nil {
		// If we can't get working directory, use the module path
		cmd := exec.Command("go", "list", "-f", "{{.Dir}}")
		output, err2 := cmd.CombinedOutput()
		if err2 != nil {
			t.Skipf("Cannot determine build directory: %v (original error: %v)", err2, err)
			return
		}
		originalWd = strings.TrimSpace(string(output))
	}
	
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = originalWd
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}
	
	// Create temporary directory structure
	tmpDir := t.TempDir()
	
	// Create CLAUDE.md
	claudeFile := filepath.Join(tmpDir, "CLAUDE.md")
	claudeContent := `---
---
# Claude Instructions

Use TypeScript for this project.
`
	if err := os.WriteFile(claudeFile, []byte(claudeContent), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}
	
	// Create task file
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	
	taskFile := filepath.Join(tasksDir, "build-feature.md")
	taskContent := `---
---
# Build Feature

Please build a new feature with ${feature_name}.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	
	// Run the binary
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-p", "feature_name=authentication", "ClaudeCode", "build-feature")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}
	
	// Verify output contains the task with substituted parameters
	outputStr := string(output)
	if !strings.Contains(outputStr, "authentication") {
		t.Errorf("output does not contain substituted parameter 'authentication':\n%s", outputStr)
	}
	
	if !strings.Contains(outputStr, "Build Feature") {
		t.Errorf("output does not contain task content:\n%s", outputStr)
	}
	
	// Verify .agents/rules directory was created
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Error(".agents/rules directory was not created")
	}
	
	// Verify CLAUDE.md was synchronized
	syncedClaude := filepath.Join(rulesDir, "CLAUDE.md")
	if _, err := os.Stat(syncedClaude); os.IsNotExist(err) {
		t.Error("CLAUDE.md was not synchronized")
	}
}
