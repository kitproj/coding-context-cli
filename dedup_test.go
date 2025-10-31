package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSymlinkDeduplication(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	memoriesDir := filepath.Join(contextDir, "memories")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a memory file
	memoryFile := filepath.Join(memoriesDir, "original.md")
	memoryContent := `---
---
# Original Memory

This is the original content.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a symlink to the memory file
	symlinkFile := filepath.Join(memoriesDir, "symlink.md")
	if err := os.Symlink(memoryFile, symlinkFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Read the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	// The content should appear only once, not twice
	contentStr := string(content)
	occurrences := strings.Count(contentStr, "This is the original content.")
	if occurrences != 1 {
		t.Errorf("Expected content to appear once, but found %d occurrences. Content:\n%s", occurrences, contentStr)
	}
}

func TestContentHashDeduplication(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	memoriesDir := filepath.Join(contextDir, "memories")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create original memory file
	memoryContent := `---
---
# Setup Instructions

Follow these steps to set up the project.
`
	memoryFile1 := filepath.Join(memoriesDir, "setup.md")
	if err := os.WriteFile(memoryFile1, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file 1: %v", err)
	}

	// Create a copy with the same content (duplicate)
	memoryFile2 := filepath.Join(memoriesDir, "setup-copy.md")
	if err := os.WriteFile(memoryFile2, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file 2: %v", err)
	}

	// Create a different memory file
	differentContent := `---
---
# Different Instructions

This is different content.
`
	memoryFile3 := filepath.Join(memoriesDir, "different.md")
	if err := os.WriteFile(memoryFile3, []byte(differentContent), 0644); err != nil {
		t.Fatalf("failed to write memory file 3: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Read the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// "Follow these steps" should appear only once (not twice from the duplicate)
	setupOccurrences := strings.Count(contentStr, "Follow these steps to set up the project.")
	if setupOccurrences != 1 {
		t.Errorf("Expected setup content to appear once, but found %d occurrences", setupOccurrences)
	}

	// "This is different content" should appear once
	differentOccurrences := strings.Count(contentStr, "This is different content.")
	if differentOccurrences != 1 {
		t.Errorf("Expected different content to appear once, but found %d occurrences", differentOccurrences)
	}
}

func TestMixedDuplicates(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	memoriesDir := filepath.Join(contextDir, "memories")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create original memory file
	originalContent := `---
---
# Original Memory

This content is unique and should appear once.
`
	originalFile := filepath.Join(memoriesDir, "original.md")
	if err := os.WriteFile(originalFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write original file: %v", err)
	}

	// Create a copy
	copyFile := filepath.Join(memoriesDir, "copy.md")
	if err := os.WriteFile(copyFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write copy file: %v", err)
	}

	// Create a symlink to the original
	symlinkFile := filepath.Join(memoriesDir, "symlink.md")
	if err := os.Symlink(originalFile, symlinkFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Create a unique file
	uniqueContent := `---
---
# Unique Memory

This is unique content.
`
	uniqueFile := filepath.Join(memoriesDir, "unique.md")
	if err := os.WriteFile(uniqueFile, []byte(uniqueContent), 0644); err != nil {
		t.Fatalf("failed to write unique file: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Read the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Original content should appear only once despite being in 3 files (original, copy, symlink)
	originalOccurrences := strings.Count(contentStr, "This content is unique and should appear once.")
	if originalOccurrences != 1 {
		t.Errorf("Expected original content to appear once, but found %d occurrences. Content:\n%s", originalOccurrences, contentStr)
	}

	// Unique content should appear once
	uniqueOccurrences := strings.Count(contentStr, "This is unique content.")
	if uniqueOccurrences != 1 {
		t.Errorf("Expected unique content to appear once, but found %d occurrences", uniqueOccurrences)
	}
}

func TestDeduplicationAcrossMultipleDirectories(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir1 := filepath.Join(tmpDir, "context1")
	contextDir2 := filepath.Join(tmpDir, "context2")
	memoriesDir1 := filepath.Join(contextDir1, "memories")
	memoriesDir2 := filepath.Join(contextDir2, "memories")
	tasksDir := filepath.Join(contextDir1, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir1, 0755); err != nil {
		t.Fatalf("failed to create memories dir 1: %v", err)
	}
	if err := os.MkdirAll(memoriesDir2, 0755); err != nil {
		t.Fatalf("failed to create memories dir 2: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create the same content in both directories
	sharedContent := `---
---
# Shared Configuration

This content exists in multiple directories.
`
	file1 := filepath.Join(memoriesDir1, "shared.md")
	if err := os.WriteFile(file1, []byte(sharedContent), 0644); err != nil {
		t.Fatalf("failed to write file 1: %v", err)
	}

	file2 := filepath.Join(memoriesDir2, "shared.md")
	if err := os.WriteFile(file2, []byte(sharedContent), 0644); err != nil {
		t.Fatalf("failed to write file 2: %v", err)
	}

	// Create unique content in each directory
	unique1 := `---
---
# Context 1 Specific

Only in context 1.
`
	file3 := filepath.Join(memoriesDir1, "unique1.md")
	if err := os.WriteFile(file3, []byte(unique1), 0644); err != nil {
		t.Fatalf("failed to write file 3: %v", err)
	}

	unique2 := `---
---
# Context 2 Specific

Only in context 2.
`
	file4 := filepath.Join(memoriesDir2, "unique2.md")
	if err := os.WriteFile(file4, []byte(unique2), 0644); err != nil {
		t.Fatalf("failed to write file 4: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary with both directories
	cmd = exec.Command(binaryPath, "-d", contextDir1, "-d", contextDir2, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Read the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Shared content should appear only once despite being in both directories
	sharedOccurrences := strings.Count(contentStr, "This content exists in multiple directories.")
	if sharedOccurrences != 1 {
		t.Errorf("Expected shared content to appear once, but found %d occurrences", sharedOccurrences)
	}

	// Unique content from context 1 should appear once
	unique1Occurrences := strings.Count(contentStr, "Only in context 1.")
	if unique1Occurrences != 1 {
		t.Errorf("Expected context 1 unique content to appear once, but found %d occurrences", unique1Occurrences)
	}

	// Unique content from context 2 should appear once
	unique2Occurrences := strings.Count(contentStr, "Only in context 2.")
	if unique2Occurrences != 1 {
		t.Errorf("Expected context 2 unique content to appear once, but found %d occurrences", unique2Occurrences)
	}
}

func TestDifferentFrontmatterNotDeduplicated(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	memoriesDir := filepath.Join(contextDir, "memories")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create two files with the same body content but different frontmatter
	file1Content := `---
env: production
---
# Common Content

This is the body content.
`
	file1 := filepath.Join(memoriesDir, "prod.md")
	if err := os.WriteFile(file1, []byte(file1Content), 0644); err != nil {
		t.Fatalf("failed to write file 1: %v", err)
	}

	file2Content := `---
env: development
---
# Common Content

This is the body content.
`
	file2 := filepath.Join(memoriesDir, "dev.md")
	if err := os.WriteFile(file2, []byte(file2Content), 0644); err != nil {
		t.Fatalf("failed to write file 2: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Read the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// The body content should appear TWICE since the files have different frontmatter
	// (and therefore different raw content)
	bodyOccurrences := strings.Count(contentStr, "This is the body content.")
	if bodyOccurrences != 2 {
		t.Errorf("Expected body content to appear twice (different frontmatter = different files), but found %d occurrences. Content:\n%s", bodyOccurrences, contentStr)
	}
}

func TestSimilarityBasedDeduplication(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	memoriesDir := filepath.Join(contextDir, "memories")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create an original file
	originalContent := `---
---
# Development Setup Guide

This project requires the following setup steps:

1. Install Node.js version 18 or higher
2. Run npm install to install dependencies
3. Create a .env file with your configuration
4. Run npm test to verify everything works

Make sure to follow these steps carefully.
`
	originalFile := filepath.Join(memoriesDir, "setup-original.md")
	if err := os.WriteFile(originalFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write original file: %v", err)
	}

	// Create a very similar file (minor typo/rewording - should be caught as similar)
	similarContent := `---
---
# Development Setup Guide

This project requires the following setup steps:

1. Install Node.js version 18 or higher
2. Run npm install to install dependencies  
3. Create a .env file with your configuration
4. Run npm test to verify everything works

Make sure to follow these instructions carefully.
`
	similarFile := filepath.Join(memoriesDir, "setup-similar.md")
	if err := os.WriteFile(similarFile, []byte(similarContent), 0644); err != nil {
		t.Fatalf("failed to write similar file: %v", err)
	}

	// Create a different file (should NOT be caught as similar)
	differentContent := `---
---
# Coding Standards

Please follow these coding standards:

- Use 2 spaces for indentation
- Write descriptive variable names
- Add comments for complex logic
`
	differentFile := filepath.Join(memoriesDir, "standards.md")
	if err := os.WriteFile(differentFile, []byte(differentContent), 0644); err != nil {
		t.Fatalf("failed to write different file: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Read the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// The similar file should be deduplicated (only one setup guide should appear)
	setupOccurrences := strings.Count(contentStr, "Development Setup Guide")
	if setupOccurrences != 1 {
		t.Errorf("Expected setup guide to appear once (similar file should be deduplicated), but found %d occurrences", setupOccurrences)
	}

	// The different file should still be included
	if !strings.Contains(contentStr, "Coding Standards") {
		t.Errorf("Expected coding standards to be included (different content)")
	}

	// Verify that "follow these steps carefully" or "follow these instructions carefully" appears only once
	stepsCount := strings.Count(contentStr, "steps carefully")
	instructionsCount := strings.Count(contentStr, "instructions carefully")
	totalOccurrences := stepsCount + instructionsCount
	if totalOccurrences != 1 {
		t.Errorf("Expected either 'steps carefully' or 'instructions carefully' to appear once, but found %d total occurrences", totalOccurrences)
	}
}
