package context

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssembler_Assemble(t *testing.T) {
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

	// Test assembling context
	var stdout, stderr bytes.Buffer
	assembler := NewAssembler(Config{
		WorkDir:   tmpDir,
		TaskName:  "test-task",
		Params:    make(ParamMap),
		Selectors: make(SelectorMap),
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	ctx := context.Background()
	if err := assembler.Assemble(ctx); err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	// Check that rule content is present
	outputStr := stdout.String()
	if !strings.Contains(outputStr, "# Development Setup") {
		t.Errorf("rule content not found in stdout")
	}

	// Check that task content is present
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}

	// Check stderr for progress messages
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Including rule file") {
		t.Errorf("progress messages not found in stderr")
	}
	if !strings.Contains(stderrStr, "Total estimated tokens") {
		t.Errorf("total token count not found in stderr")
	}
}

func TestAssembler_AssembleWithParams(t *testing.T) {
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

	// Test assembling context with parameters
	var stdout, stderr bytes.Buffer
	params := make(ParamMap)
	params["component"] = "auth"
	params["issue"] = "login bug"

	assembler := NewAssembler(Config{
		WorkDir:   tmpDir,
		TaskName:  "test-task",
		Params:    params,
		Selectors: make(SelectorMap),
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	ctx := context.Background()
	if err := assembler.Assemble(ctx); err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	// Check that template variables were expanded
	outputStr := stdout.String()
	if !strings.Contains(outputStr, "Please work on auth and fix login bug.") {
		t.Errorf("template variables were not expanded correctly. Output:\n%s", outputStr)
	}
}

func TestAssembler_AssembleWithSelectors(t *testing.T) {
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

	// Test assembling context with selector filtering for Python
	var stdout, stderr bytes.Buffer
	selectors := make(SelectorMap)
	selectors["language"] = "python"

	assembler := NewAssembler(Config{
		WorkDir:   tmpDir,
		TaskName:  "test-task",
		Params:    make(ParamMap),
		Selectors: selectors,
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	ctx := context.Background()
	if err := assembler.Assemble(ctx); err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	// Check that only Python guidelines are included
	outputStr := stdout.String()
	if !strings.Contains(outputStr, "# Python Guidelines") {
		t.Errorf("Python guidelines not found in stdout")
	}
	if strings.Contains(outputStr, "# Go Guidelines") {
		t.Errorf("Go guidelines should not be in stdout when filtering for Python")
	}
}

func TestAssembler_TaskNotFound(t *testing.T) {
	// Create a temporary directory without tasks
	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	assembler := NewAssembler(Config{
		WorkDir:   tmpDir,
		TaskName:  "nonexistent-task",
		Params:    make(ParamMap),
		Selectors: make(SelectorMap),
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	ctx := context.Background()
	err := assembler.Assemble(ctx)
	if err == nil {
		t.Fatalf("expected error for nonexistent task, got nil")
	}

	if !strings.Contains(err.Error(), "prompt file not found") {
		t.Errorf("expected 'prompt file not found' error, got: %v", err)
	}
}

func TestNewAssembler_DefaultValues(t *testing.T) {
	// Test that NewAssembler sets default values correctly
	config := Config{
		WorkDir:  ".",
		TaskName: "test",
	}

	assembler := NewAssembler(config)

	if assembler.config.Stdout != os.Stdout {
		t.Errorf("expected Stdout to default to os.Stdout")
	}
	if assembler.config.Stderr != os.Stderr {
		t.Errorf("expected Stderr to default to os.Stderr")
	}
	if assembler.config.Params == nil {
		t.Errorf("expected Params to be initialized")
	}
	if assembler.config.Selectors == nil {
		t.Errorf("expected Selectors to be initialized")
	}
}
