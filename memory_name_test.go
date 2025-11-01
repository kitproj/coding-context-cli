package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemoryNameExclusion(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := ioutil.TempDir("", "memory-name-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two memory files with the same name in their frontmatter
	name: MyMemory
---
This is the first memory.`
	memory1Path := filepath.Join(tmpDir, "memory1.md")
	if err := ioutil.WriteFile(memory1Path, []byte(memory1Content), 0644); err != nil {
		t.Fatalf("Failed to write memory1.md: %v", err)
	}

	memory2Content := `---
name: MyMemory
---
This is the second memory.`
	memory2Path := filepath.Join(tmpDir, "memory2.md")
	if err := ioutil.WriteFile(memory2Path, []byte(memory2Content), 0644); err != nil {
		t.Fatalf("Failed to write memory2.md: %v", err)
	}

	// Create a dummy task file
	taskContent := "This is the task."
	taskPath := filepath.Join(tmpDir, "task.md")
	if err := ioutil.WriteFile(taskPath, []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to write task.md: %v", err)
	}

	// Set up the arguments for the run function
	args := []string{"task"}
	memories = []string{tmpDir}
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
	promptBytes, err := ioutil.ReadFile(filepath.Join(tmpDir, "prompt.md"))
	if err != nil {
		t.Fatalf("Failed to read prompt.md: %v", err)
	}
	prompt := string(promptBytes)

	// We expect only one of the memories to be included
	if !(strings.Contains(prompt, "This is the first memory.") && !strings.Contains(prompt, "This is the second memory.")) &&
		!(!strings.Contains(prompt, "This is the first memory.") && strings.Contains(prompt, "This is the second memory.")) {
		t.Errorf("Expected only one memory to be included, but got: %s", prompt)
	}
}
