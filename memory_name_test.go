package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemoryNameExclusion(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "memory-name-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two subdirectories with memory files with the same basename
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	memory1Content := `---
---
This is the first memory.`
	memory1Path := filepath.Join(dir1, "memory.md")
	if err := os.WriteFile(memory1Path, []byte(memory1Content), 0644); err != nil {
		t.Fatalf("Failed to write memory.md in dir1: %v", err)
	}

	memory2Content := `---
---
This is the second memory.`
	memory2Path := filepath.Join(dir2, "memory.md")
	if err := os.WriteFile(memory2Path, []byte(memory2Content), 0644); err != nil {
		t.Fatalf("Failed to write memory.md in dir2: %v", err)
	}

	// Create a dummy task file
	taskContent := "This is the task."
	taskPath := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(taskPath, []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to write task.md: %v", err)
	}

	// Set up the arguments for the run function
	args := []string{"task"}
	memories = []string{dir1, dir2}
	tasks = []string{tmpDir}
	outputDir = tmpDir
	params = make(paramMap)
	includes = make(selectorMap)
	excludes = make(selectorMap)
	runBootstrap = false
	workDir = tmpDir

	// Run the application
	if err := run(context.Background(), args); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	// Check the output
	promptBytes, err := os.ReadFile(filepath.Join(tmpDir, "prompt.md"))
	if err != nil {
		t.Fatalf("Failed to read prompt.md: %v", err)
	}
	prompt := string(promptBytes)

	// We expect only one of the memories to be included
	hasFirst := strings.Contains(prompt, "This is the first memory.")
	hasSecond := strings.Contains(prompt, "This is the second memory.")
	if hasFirst == hasSecond {
		t.Errorf("Expected only one memory to be included, but got: %s", prompt)
	}
}
