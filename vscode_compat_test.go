package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   map[string]string
		expected string
	}{
		{
			name:     "simple variable",
			input:    "Hello ${name}",
			params:   map[string]string{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "input variable",
			input:    "Feature: ${input:featureName}",
			params:   map[string]string{"featureName": "Authentication"},
			expected: "Feature: Authentication",
		},
		{
			name:     "input variable with placeholder",
			input:    "Feature: ${input:featureName:Enter feature name}",
			params:   map[string]string{"featureName": "UserProfile"},
			expected: "Feature: UserProfile",
		},
		{
			name:     "multiple variables",
			input:    "Task ${taskName} in ${language}",
			params:   map[string]string{"taskName": "MyTask", "language": "Go"},
			expected: "Task MyTask in Go",
		},
		{
			name:     "workspace variable (ignored)",
			input:    "Path: ${workspaceFolder}/src",
			params:   map[string]string{},
			expected: "Path: ${workspaceFolder}/src",
		},
		{
			name:     "workspaceFolderBasename variable (ignored)",
			input:    "Basename: ${workspaceFolderBasename}",
			params:   map[string]string{},
			expected: "Basename: ${workspaceFolderBasename}",
		},
		{
			name:     "file variable (ignored)",
			input:    "File: ${file}",
			params:   map[string]string{},
			expected: "File: ${file}",
		},
		{
			name:     "fileBasename variable (ignored)",
			input:    "File: ${fileBasename}",
			params:   map[string]string{},
			expected: "File: ${fileBasename}",
		},
		{
			name:     "selection variable (ignored)",
			input:    "Selected: ${selection}",
			params:   map[string]string{},
			expected: "Selected: ${selection}",
		},
		{
			name:     "user variable starting with 'file' (substituted)",
			input:    "Type: ${fileType}",
			params:   map[string]string{"fileType": "markdown"},
			expected: "Type: markdown",
		},
		{
			name:     "user variable starting with 'workspace' (substituted)",
			input:    "Config: ${workspaceConfig}",
			params:   map[string]string{"workspaceConfig": "config.json"},
			expected: "Config: config.json",
		},
		{
			name:     "mixed variables",
			input:    "Task ${taskName} in ${workspaceFolder} with ${language}",
			params:   map[string]string{"taskName": "MyTask", "language": "Go"},
			expected: "Task MyTask in ${workspaceFolder} with Go",
		},
		{
			name:     "no variables",
			input:    "Plain text without variables",
			params:   map[string]string{},
			expected: "Plain text without variables",
		},
		{
			name:     "missing parameter",
			input:    "Task ${taskName}",
			params:   map[string]string{},
			expected: "Task ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteVariables(tt.input, tt.params)
			if result != tt.expected {
				t.Errorf("substituteVariables() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPromptMdExtension(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".github", "prompts")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a prompt file with .md extension
	promptFile := filepath.Join(tasksDir, "test-vscode-task.md")
	promptContent := `---
description: 'Test VS Code prompt'
mode: 'ask'
---
# Test Task

This is a VS Code style prompt file.
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "test-vscode-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that the prompt.md file was created
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	// Check content includes the prompt
	contentStr := string(content)
	if !strings.Contains(contentStr, "This is a VS Code style prompt file.") {
		t.Errorf("Expected prompt content not found in output")
	}
}

func TestVSCodeVariableConversion(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".github", "prompts")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a prompt file with VS Code variable syntax
	promptFile := filepath.Join(tasksDir, "vscode-vars.md")
	promptContent := `---
description: 'Test variable conversion'
---
# Task: ${input:taskName}

Implement ${input:featureName:Enter feature name} in ${language}.
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary with parameters
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir,
		"-p", "taskName=MyTask",
		"-p", "featureName=Auth",
		"-p", "language=Go",
		"vscode-vars")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that the prompt.md file was created
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	// Check that variables were substituted
	contentStr := string(content)
	if !strings.Contains(contentStr, "Task: MyTask") {
		t.Errorf("Expected 'Task: MyTask' in output, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "Implement Auth in Go") {
		t.Errorf("Expected 'Implement Auth in Go' in output, got: %s", contentStr)
	}
}

func TestGitHubPromptsDirectory(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure using .github/prompts
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".github", "prompts")
	tasksDir := filepath.Join(contextDir, "tasks")
	memoriesDir := filepath.Join(contextDir, "memories")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}

	// Create a memory file
	memoryFile := filepath.Join(memoriesDir, "context.md")
	memoryContent := `---
---
# Project Context

This is from .github/prompts directory.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "github-task.md")
	promptContent := `---
---
# GitHub Task

This uses the .github/prompts location.
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary - it should find files in .github/prompts by default
	cmd = exec.Command(binaryPath, "-o", outputDir, "github-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that the prompt.md file was created and includes both memory and task
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "This is from .github/prompts directory") {
		t.Errorf("Expected memory content from .github/prompts")
	}
	if !strings.Contains(contentStr, "This uses the .github/prompts location") {
		t.Errorf("Expected task content from .github/prompts")
	}
}

func TestFindPromptFile(t *testing.T) {
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	tests := []struct {
		name          string
		createFile    string
		taskName      string
		shouldFind    bool
		expectedFile  string
	}{
		{
			name:         "finds .md",
			createFile:   "test.md",
			taskName:     "test",
			shouldFind:   true,
			expectedFile: "test.md",
		},
		{
			name:       "not found",
			createFile: "",
			taskName:   "nonexistent",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up tasks directory
			os.RemoveAll(tasksDir)
			os.MkdirAll(tasksDir, 0755)

			// Create file if needed
			if tt.createFile != "" {
				filePath := filepath.Join(tasksDir, tt.createFile)
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			// Test findPromptFile
			found, err := findPromptFile(tmpDir, tt.taskName)
			if tt.shouldFind {
				if err != nil {
					t.Errorf("expected to find file, got error: %v", err)
				}
				expectedPath := filepath.Join(tmpDir, "tasks", tt.expectedFile)
				if found != expectedPath {
					t.Errorf("expected %s, got %s", expectedPath, found)
				}
			} else {
				if err == nil {
					t.Errorf("expected error, but file was found: %s", found)
				}
			}
		})
	}
}


