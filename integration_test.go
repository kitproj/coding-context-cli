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
	memoryFile := filepath.Join(memoriesDir, "setup.md")
	memoryContent := `---
---
# Development Setup

This is a setup guide.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a bootstrap file for the memory (setup.md -> setup-bootstrap)
	bootstrapFile := filepath.Join(memoriesDir, "setup-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Running bootstrap"
npm install
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task

Please help with this task.
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

	// Check that the bootstrap.d directory was created
	bootstrapDDir := filepath.Join(outputDir, "bootstrap.d")
	if _, err := os.Stat(bootstrapDDir); os.IsNotExist(err) {
		t.Errorf("bootstrap.d directory was not created")
	}

	// Check that a bootstrap file exists in bootstrap.d
	files, err := os.ReadDir(bootstrapDDir)
	if err != nil {
		t.Fatalf("failed to read bootstrap.d dir: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 bootstrap file, got %d", len(files))
	}

	// Check that the bootstrap file has the correct content
	if len(files) > 0 {
		bootstrapPath := filepath.Join(bootstrapDDir, files[0].Name())
		content, err := os.ReadFile(bootstrapPath)
		if err != nil {
			t.Fatalf("failed to read bootstrap file: %v", err)
		}
		if string(content) != bootstrapContent {
			t.Errorf("bootstrap content mismatch:\ngot: %q\nwant: %q", string(content), bootstrapContent)
		}
	}

	// Check that the prompt.md file was created
	promptOutput := filepath.Join(outputDir, "prompt.md")
	if _, err := os.Stat(promptOutput); os.IsNotExist(err) {
		t.Errorf("prompt.md file was not created")
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

	// Create a memory file WITHOUT a bootstrap
	memoryFile := filepath.Join(memoriesDir, "info.md")
	memoryContent := `---
---
# Project Info

Just some information.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a prompt file
	promptFile := filepath.Join(tasksDir, "test-task.md")
	promptContent := `---
---
# Test Task

Please help with this task.
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

	// Check that the bootstrap.d directory was created but is empty
	bootstrapDDir := filepath.Join(outputDir, "bootstrap.d")
	files, err := os.ReadDir(bootstrapDDir)
	if err != nil {
		t.Fatalf("failed to read bootstrap.d dir: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 bootstrap files, got %d", len(files))
	}

	// Check that the prompt.md file was still created
	promptOutput := filepath.Join(outputDir, "prompt.md")
	if _, err := os.Stat(promptOutput); os.IsNotExist(err) {
		t.Errorf("prompt.md file was not created")
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

	// Create first memory file with bootstrap
	if err := os.WriteFile(filepath.Join(memoriesDir, "setup.md"), []byte("---\n---\n# Setup\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memoriesDir, "setup-bootstrap"), []byte("#!/bin/bash\necho setup\n"), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Create second memory file with bootstrap
	if err := os.WriteFile(filepath.Join(memoriesDir, "deps.md"), []byte("---\n---\n# Dependencies\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memoriesDir, "deps-bootstrap"), []byte("#!/bin/bash\necho deps\n"), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Create a prompt file
	if err := os.WriteFile(filepath.Join(tasksDir, "test-task.md"), []byte("---\n---\n# Test\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that both bootstrap files exist in bootstrap.d
	bootstrapDDir := filepath.Join(outputDir, "bootstrap.d")
	files, err := os.ReadDir(bootstrapDDir)
	if err != nil {
		t.Fatalf("failed to read bootstrap.d dir: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 bootstrap files, got %d", len(files))
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

	// Create memory files with different frontmatter
	if err := os.WriteFile(filepath.Join(memoriesDir, "prod.md"), []byte("---\nenv: production\nlanguage: go\n---\n# Production\nProd content\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memoriesDir, "dev.md"), []byte("---\nenv: development\nlanguage: python\n---\n# Development\nDev content\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memoriesDir, "test.md"), []byte("---\nenv: test\nlanguage: go\n---\n# Test\nTest content\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	// Create a file without frontmatter (should be included by default)
	if err := os.WriteFile(filepath.Join(memoriesDir, "nofm.md"), []byte("---\n---\n# No Frontmatter\nNo FM content\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a prompt file
	if err := os.WriteFile(filepath.Join(tasksDir, "test-task.md"), []byte("---\n---\n# Test Task\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Test 1: Include by env=production
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "-s", "env=production", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "Prod content") {
		t.Errorf("Expected production content in output")
	}
	if strings.Contains(contentStr, "Dev content") {
		t.Errorf("Did not expect development content in output")
	}
	if strings.Contains(contentStr, "Test content") {
		t.Errorf("Did not expect test content in output")
	}
	// File without env key should be included (key missing is allowed)
	if !strings.Contains(contentStr, "No FM content") {
		t.Errorf("Expected no frontmatter content in output (missing key should be allowed)")
	}

	// Clean output for next test
	os.RemoveAll(outputDir)

	// Test 2: Include by language=go (should include prod and test, and nofm)
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "-s", "language=go", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}
	contentStr = string(content)
	if !strings.Contains(contentStr, "Prod content") {
		t.Errorf("Expected production content in output")
	}
	if strings.Contains(contentStr, "Dev content") {
		t.Errorf("Did not expect development content in output")
	}
	if !strings.Contains(contentStr, "Test content") {
		t.Errorf("Expected test content in output")
	}
	if !strings.Contains(contentStr, "No FM content") {
		t.Errorf("Expected no frontmatter content in output (missing key should be allowed)")
	}

	// Clean output for next test
	os.RemoveAll(outputDir)

	// Test 3: Exclude by env=production (should include dev and test, and nofm)
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "-S", "env=production", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}
	contentStr = string(content)
	if strings.Contains(contentStr, "Prod content") {
		t.Errorf("Did not expect production content in output")
	}
	if !strings.Contains(contentStr, "Dev content") {
		t.Errorf("Expected development content in output")
	}
	if !strings.Contains(contentStr, "Test content") {
		t.Errorf("Expected test content in output")
	}
	if !strings.Contains(contentStr, "No FM content") {
		t.Errorf("Expected no frontmatter content in output (missing key should be allowed)")
	}

	// Clean output for next test
	os.RemoveAll(outputDir)

	// Test 4: Multiple includes env=production language=go (should include only prod and nofm)
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "-s", "env=production", "-s", "language=go", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}
	contentStr = string(content)
	if !strings.Contains(contentStr, "Prod content") {
		t.Errorf("Expected production content in output")
	}
	if strings.Contains(contentStr, "Dev content") {
		t.Errorf("Did not expect development content in output")
	}
	if strings.Contains(contentStr, "Test content") {
		t.Errorf("Did not expect test content in output")
	}
	if !strings.Contains(contentStr, "No FM content") {
		t.Errorf("Expected no frontmatter content in output (missing key should be allowed)")
	}

	// Clean output for next test
	os.RemoveAll(outputDir)

	// Test 5: Mix of include and exclude -s env=production -S language=python (should include only prod with go)
	cmd = exec.Command(binaryPath, "-d", contextDir, "-o", outputDir, "-s", "env=production", "-S", "language=python", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}
	contentStr = string(content)
	if !strings.Contains(contentStr, "Prod content") {
		t.Errorf("Expected production content in output")
	}
	if strings.Contains(contentStr, "Dev content") {
		t.Errorf("Did not expect development content in output")
	}
	if strings.Contains(contentStr, "Test content") {
		t.Errorf("Did not expect test content in output")
	}
	if !strings.Contains(contentStr, "No FM content") {
		t.Errorf("Expected no frontmatter content in output (missing keys should be allowed)")
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
	contextDir := filepath.Join(tmpDir, ".prompts")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a prompt file with os.Expand style templates
	promptFile := filepath.Join(tasksDir, "test-expand.md")
	promptContent := `---
---
# Test Task: ${taskName}

Please implement ${feature} using ${language}.

The project is for $company.
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary with parameters
	cmd = exec.Command(binaryPath, 
		"-d", contextDir, 
		"-o", outputDir,
		"-p", "taskName=AddAuth",
		"-p", "feature=Authentication",
		"-p", "language=Go",
		"-p", "company=Acme Corp",
		"test-expand")
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

	// Verify substitutions
	if !strings.Contains(contentStr, "Test Task: AddAuth") {
		t.Errorf("Expected 'Test Task: AddAuth' in output, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "Please implement Authentication using Go") {
		t.Errorf("Expected 'Please implement Authentication using Go' in output, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "The project is for Acme Corp") {
		t.Errorf("Expected 'The project is for Acme Corp' in output, got:\n%s", contentStr)
	}
}

func TestTemplateExpansionWithMissingParams(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a prompt file with variables that won't be provided
	promptFile := filepath.Join(tasksDir, "test-missing.md")
	promptContent := `---
---
# Task: ${providedVar}

Missing var: ${missingVar}
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary with only one parameter
	cmd = exec.Command(binaryPath, 
		"-d", contextDir, 
		"-o", outputDir,
		"-p", "providedVar=ProvidedValue",
		"test-missing")
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

	// Verify provided variable is substituted
	if !strings.Contains(contentStr, "Task: ProvidedValue") {
		t.Errorf("Expected 'Task: ProvidedValue' in output, got:\n%s", contentStr)
	}
	
	// Verify missing variable is replaced with empty string
	if strings.Contains(contentStr, "${missingVar}") {
		t.Errorf("Expected ${missingVar} to be replaced with empty string, got:\n%s", contentStr)
	}
}

func TestDirectoryNotFound(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	
	// Try to run with a non-existent directory
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")
	cmd = exec.Command(binaryPath, "-d", nonExistentDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	
	// Should fail with appropriate error
	if err == nil {
		t.Fatalf("expected error when task not found, but succeeded")
	}
	
	outputStr := string(output)
	
	// Verify the error message includes the searched paths
	if !strings.Contains(outputStr, "prompt file not found for task: test-task") {
		t.Errorf("Expected error message about task not found, got:\n%s", outputStr)
	}
	
	// Verify the error message includes the non-existent directory that was searched
	if !strings.Contains(outputStr, nonExistentDir) {
		t.Errorf("Expected error message to include searched directory %s, got:\n%s", nonExistentDir, outputStr)
	}
	
	// Verify the error message includes "searched:"
	if !strings.Contains(outputStr, "searched:") {
		t.Errorf("Expected error message to include list of searched directories, got:\n%s", outputStr)
	}
}
