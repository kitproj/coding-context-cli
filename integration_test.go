package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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

func TestBootstrapFileNaming(t *testing.T) {
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
	memoryFile := filepath.Join(memoriesDir, "jira.md")
	memoryContent := `---
---
# Jira Integration
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a bootstrap file for the memory (jira.md -> jira-bootstrap)
	bootstrapFile := filepath.Join(memoriesDir, "jira-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Setting up Jira"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that the bootstrap file has the correct naming format
	bootstrapDDir := filepath.Join(outputDir, "bootstrap.d")
	files, err := os.ReadDir(bootstrapDDir)
	if err != nil {
		t.Fatalf("failed to read bootstrap.d dir: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 bootstrap file, got %d", len(files))
	}

	// Verify the naming format: jira-bootstrap-<8-hex-chars>
	if len(files) > 0 {
		fileName := files[0].Name()
		// Should start with "jira-bootstrap-"
		if !strings.HasPrefix(fileName, "jira-bootstrap-") {
			t.Errorf("bootstrap file name should start with 'jira-bootstrap-', got: %s", fileName)
		}
		// Should have exactly 8 hex characters after the prefix
		suffix := strings.TrimPrefix(fileName, "jira-bootstrap-")
		if len(suffix) != 8 {
			t.Errorf("bootstrap file name should have 8 hex characters after prefix, got %d: %s", len(suffix), fileName)
		}
		// Verify all characters in suffix are hex
		for _, c := range suffix {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("bootstrap file name suffix should only contain hex characters, got: %s", fileName)
				break
			}
		}
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "test-task")
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-s", "language=go", "test-task")
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-S", "env=production", "test-task")
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "-s", "language=go", "test-task")
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
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "-S", "language=python", "test-task")
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
		"-t", tasksDir,
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
		"-t", tasksDir,
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
	if !strings.Contains(contentStr, "${missingVar}") {
		t.Errorf("Expected ${missingVar} to not be replaced, got:\n%s", contentStr)
	}
}

func TestBootstrapFlag(t *testing.T) {
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
# Setup

This is a setup guide.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a bootstrap file that creates a marker file
	bootstrapFile := filepath.Join(memoriesDir, "setup-bootstrap")
	markerFile := filepath.Join(outputDir, "bootstrap-ran.txt")
	bootstrapContent := `#!/bin/bash
echo "Bootstrap executed" > ` + markerFile + `
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

	// Run the binary WITH the -b flag
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-b", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that the marker file was created (proving the bootstrap ran)
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Errorf("marker file was not created, bootstrap script did not run")
	}

	// Verify the marker file content
	content, err := os.ReadFile(markerFile)
	if err != nil {
		t.Fatalf("failed to read marker file: %v", err)
	}
	expectedContent := "Bootstrap executed\n"
	if string(content) != expectedContent {
		t.Errorf("marker file content mismatch:\ngot: %q\nwant: %q", string(content), expectedContent)
	}
}

func TestBootstrapFlagNotSet(t *testing.T) {
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
# Setup

This is a setup guide.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a bootstrap file that creates a marker file
	bootstrapFile := filepath.Join(memoriesDir, "setup-bootstrap")
	markerFile := filepath.Join(outputDir, "bootstrap-ran.txt")
	bootstrapContent := `#!/bin/bash
echo "Bootstrap executed" > ` + markerFile + `
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

	// Run the binary WITHOUT the -b flag
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that the marker file was NOT created (bootstrap should not run)
	if _, err := os.Stat(markerFile); !os.IsNotExist(err) {
		t.Errorf("marker file was created, but bootstrap should not have run without -b flag")
	}
}

func TestBootstrapCancellation(t *testing.T) {
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
# Setup

Long running setup.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a bootstrap file that runs for a while
	bootstrapFile := filepath.Join(memoriesDir, "setup-bootstrap")
	bootstrapContent := `#!/bin/bash
for i in {1..30}; do
  echo "Running $i"
  sleep 1
done
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
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

	// Run the binary WITH the -b flag and send interrupt signal
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-b", "test-task")
	cmd.Dir = tmpDir

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start command: %v", err)
	}

	// Give it a moment to start the bootstrap script
	time.Sleep(2 * time.Second)

	// Send interrupt signal
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("failed to send interrupt signal: %v", err)
	}

	// Wait for the process to finish
	err := cmd.Wait()

	// The process should exit due to the signal
	// Check that it didn't complete successfully (which would mean it ran all 30 iterations)
	if err == nil {
		t.Error("expected command to be interrupted, but it completed successfully")
	}
}

// TestTaskNameBuiltinFilter verifies that the task_name built-in filter
// automatically includes/excludes memory files based on the task being run
func TestTaskNameBuiltinFilter(t *testing.T) {
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

	// Create memory files with task_name frontmatter
	if err := os.WriteFile(filepath.Join(memoriesDir, "deploy-specific.md"), []byte("---\ntask_name: deploy\n---\n# Deploy Memory\nDeploy-specific content\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memoriesDir, "test-specific.md"), []byte("---\ntask_name: test\n---\n# Test Memory\nTest-specific content\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}
	// Create a file without task_name (should be included for all tasks)
	if err := os.WriteFile(filepath.Join(memoriesDir, "general.md"), []byte("---\n---\n# General Memory\nGeneral content\n"), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create prompt files for both tasks
	if err := os.WriteFile(filepath.Join(tasksDir, "deploy.md"), []byte("---\n---\n# Deploy Task\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tasksDir, "test.md"), []byte("---\n---\n# Test Task\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Test 1: Run with "deploy" task - should include deploy-specific and general, but not test-specific
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "deploy")
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
	if !strings.Contains(contentStr, "Deploy-specific content") {
		t.Errorf("Expected deploy-specific content in output for deploy task")
	}
	if strings.Contains(contentStr, "Test-specific content") {
		t.Errorf("Did not expect test-specific content in output for deploy task")
	}
	if !strings.Contains(contentStr, "General content") {
		t.Errorf("Expected general content in output (no task_name key should be allowed)")
	}

	// Clean output for next test
	os.RemoveAll(outputDir)

	// Test 2: Run with "test" task - should include test-specific and general, but not deploy-specific
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}
	contentStr = string(content)
	if strings.Contains(contentStr, "Deploy-specific content") {
		t.Errorf("Did not expect deploy-specific content in output for test task")
	}
	if !strings.Contains(contentStr, "Test-specific content") {
		t.Errorf("Expected test-specific content in output for test task")
	}
	if !strings.Contains(contentStr, "General content") {
		t.Errorf("Expected general content in output (no task_name key should be allowed)")
	}
}

func TestPersonaBasic(t *testing.T) {
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
	personasDir := filepath.Join(contextDir, "personas")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(personasDir, 0755); err != nil {
		t.Fatalf("failed to create personas dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a persona file (without template variables since personas don't expand them)
	personaFile := filepath.Join(personasDir, "expert.md")
	personaContent := `---
---
# Expert Persona

You are an expert in Go.
`
	if err := os.WriteFile(personaFile, []byte(personaContent), 0644); err != nil {
		t.Fatalf("failed to write persona file: %v", err)
	}

	// Create a memory file
	memoryFile := filepath.Join(memoriesDir, "context.md")
	memoryContent := `---
---
# Context

This is context.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Task

Please help with ${feature}.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run with persona (persona is now a positional argument after task name)
	cmd = exec.Command(binaryPath, "-r", personasDir, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-p", "feature=auth", "test-task", "expert")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Verify persona appears first
	expertIdx := strings.Index(contentStr, "Expert Persona")
	contextIdx := strings.Index(contentStr, "# Context")
	taskIdx := strings.Index(contentStr, "# Task")

	if expertIdx == -1 {
		t.Errorf("Expected to find 'Expert Persona' in output")
	}
	if contextIdx == -1 {
		t.Errorf("Expected to find '# Context' in output")
	}
	if taskIdx == -1 {
		t.Errorf("Expected to find '# Task' in output")
	}

	// Verify order: persona -> context -> task
	if expertIdx > contextIdx {
		t.Errorf("Persona should appear before context. Persona at %d, Context at %d", expertIdx, contextIdx)
	}
	if contextIdx > taskIdx {
		t.Errorf("Context should appear before task. Context at %d, Task at %d", contextIdx, taskIdx)
	}

	// Verify persona content is not expanded (no template substitution)
	if !strings.Contains(contentStr, "You are an expert in Go") {
		t.Errorf("Expected persona content to remain as-is without template expansion")
	}
	// Verify task template substitution still works
	if !strings.Contains(contentStr, "Please help with auth") {
		t.Errorf("Expected task template to be expanded with feature=auth")
	}
}

func TestPersonaOptional(t *testing.T) {
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
	memoryFile := filepath.Join(memoriesDir, "context.md")
	memoryContent := `---
---
# Context

This is context.
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Task

Please help.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run WITHOUT persona (should still work)
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary without persona: %v\n%s", err, output)
	}

	// Check the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Verify context and task are present
	if !strings.Contains(contentStr, "# Context") {
		t.Errorf("Expected to find '# Context' in output")
	}
	if !strings.Contains(contentStr, "# Task") {
		t.Errorf("Expected to find '# Task' in output")
	}
}

func TestPersonaNotFound(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	personasDir := filepath.Join(contextDir, "personas")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(personasDir, 0755); err != nil {
		t.Fatalf("failed to create personas dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a task file (but no persona file)
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Task

Please help.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run with non-existent persona (should fail) - persona is now a positional argument
	cmd = exec.Command(binaryPath, "-r", personasDir, "-t", tasksDir, "-o", outputDir, "test-task", "nonexistent")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	// Should error
	if err == nil {
		t.Errorf("Expected error when persona file not found, but command succeeded")
	}

	// Check error message
	if !strings.Contains(string(output), "persona file not found") {
		t.Errorf("Expected 'persona file not found' error message, got: %s", string(output))
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
	workDir := filepath.Join(tmpDir, "work")
	memoriesDir := filepath.Join(workDir, ".prompts", "memories")
	tasksDir := filepath.Join(workDir, ".prompts", "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a memory file in the work directory
	memoryFile := filepath.Join(memoriesDir, "test.md")
	memoryContent := `---
---
# Test Memory
`
	if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "task.md")
	taskContent := `---
---
# Test Task
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary with -C option to change to work directory
	cmd = exec.Command(binaryPath, "-C", workDir, "-m", ".prompts/memories", "-t", ".prompts/tasks", "-o", outputDir, "task")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary with -C option: %v\n%s", err, output)
	}

	// Verify that prompt.md was created in the output directory
	promptFile := filepath.Join(outputDir, "prompt.md")
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		t.Errorf("prompt.md was not created in output directory")
	}

	// Verify the content includes the memory
	content, err := os.ReadFile(promptFile)
	if err != nil {
		t.Fatalf("failed to read prompt.md: %v", err)
	}
	if !strings.Contains(string(content), "Test Memory") {
		t.Errorf("prompt.md does not contain expected memory content")
	}
}

func TestTokenCounting(t *testing.T) {
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
	personasDir := filepath.Join(contextDir, "personas")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(memoriesDir, 0755); err != nil {
		t.Fatalf("failed to create memories dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	if err := os.MkdirAll(personasDir, 0755); err != nil {
		t.Fatalf("failed to create personas dir: %v", err)
	}

	// Create a persona file
	personaFile := filepath.Join(personasDir, "expert.md")
	personaContent := `# Expert Developer

You are an expert developer.`
	if err := os.WriteFile(personaFile, []byte(personaContent), 0644); err != nil {
		t.Fatalf("failed to write persona file: %v", err)
	}

	// Create memory files
	memoryFile1 := filepath.Join(memoriesDir, "setup.md")
	memoryContent1 := `# Development Setup

This is a setup guide with detailed instructions.`
	if err := os.WriteFile(memoryFile1, []byte(memoryContent1), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	memoryFile2 := filepath.Join(memoriesDir, "conventions.md")
	memoryContent2 := `# Coding Conventions

Follow best practices and write clean code.`
	if err := os.WriteFile(memoryFile2, []byte(memoryContent2), 0644); err != nil {
		t.Fatalf("failed to write memory file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `# Test Task

Complete this task with high quality.`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary with persona
	cmd = exec.Command(binaryPath, "-o", outputDir, "test-task", "expert")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	outputStr := string(output)

	// Verify token counts are printed for each file
	if !strings.Contains(outputStr, "Using persona file:") {
		t.Errorf("Expected persona file message in output")
	}
	if !strings.Contains(outputStr, "tokens)") {
		t.Errorf("Expected token count in output")
	}
	if !strings.Contains(outputStr, "Including memory file:") {
		t.Errorf("Expected memory file message in output")
	}
	if !strings.Contains(outputStr, "Using task file:") {
		t.Errorf("Expected task file message in output")
	}
	if !strings.Contains(outputStr, "Total estimated tokens:") {
		t.Errorf("Expected total token count in output")
	}

	// Verify the total is printed at the end (after all file processing)
	lines := strings.Split(outputStr, "\n")
	var totalLine string
	for _, line := range lines {
		if strings.Contains(line, "Total estimated tokens:") {
			totalLine = line
		}
	}
	if totalLine == "" {
		t.Fatalf("Total token count line not found in output: %s", outputStr)
	}

	// The total should be greater than 0
	if !strings.Contains(totalLine, "Total estimated tokens:") {
		t.Errorf("Expected 'Total estimated tokens:' in output, got: %s", totalLine)
	}
}

func TestMemoryDeduplication(t *testing.T) {
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

	// Create base memory file (coding-standards.md)
	baseMemory := filepath.Join(memoriesDir, "coding-standards.md")
	baseContent := `---
name: CodingStandards
---
# General Coding Standards

Use clean code principles.
`
	if err := os.WriteFile(baseMemory, []byte(baseContent), 0644); err != nil {
		t.Fatalf("failed to write base memory file: %v", err)
	}

	// Create specialized memory file that replaces the base (go-coding-standards.md)
	specializedMemory := filepath.Join(memoriesDir, "go-coding-standards.md")
	specializedContent := `---
name: GoCodingStandards
replaces: CodingStandards
---
# Go Coding Standards

Use clean code principles in Go.
Follow effective Go guidelines.
`
	if err := os.WriteFile(specializedMemory, []byte(specializedContent), 0644); err != nil {
		t.Fatalf("failed to write specialized memory file: %v", err)
	}

	// Create another unrelated memory file
	otherMemory := filepath.Join(memoriesDir, "project-info.md")
	otherContent := `---
---
# Project Info

This is a Go project.
`
	if err := os.WriteFile(otherMemory, []byte(otherContent), 0644); err != nil {
		t.Fatalf("failed to write other memory file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Verify the specialized memory is included
	if !strings.Contains(contentStr, "Go Coding Standards") {
		t.Errorf("Expected specialized memory (go-coding-standards.md) to be included")
	}

	// Verify the base memory is NOT included (replaced)
	if strings.Contains(contentStr, "General Coding Standards") {
		t.Errorf("Did not expect base memory (coding-standards.md) to be included - it should be replaced")
	}

	// Verify the unrelated memory is included
	if !strings.Contains(contentStr, "Project Info") {
		t.Errorf("Expected unrelated memory (project-info.md) to be included")
	}

	// Verify the output message indicates replacement
	outputStr := string(output)
	if !strings.Contains(outputStr, "Excluding memory file (replaced by another memory)") {
		t.Errorf("Expected message about excluded file due to replacement")
	}
}

func TestMemoryDeduplicationMultipleReplacements(t *testing.T) {
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

	// Create first base memory file
	base1 := filepath.Join(memoriesDir, "base1.md")
	if err := os.WriteFile(base1, []byte("---\nname: Base1\n---\n# Base 1\n"), 0644); err != nil {
		t.Fatalf("failed to write base1 memory file: %v", err)
	}

	// Create second base memory file
	base2 := filepath.Join(memoriesDir, "base2.md")
	if err := os.WriteFile(base2, []byte("---\nname: Base2\n---\n# Base 2\n"), 0644); err != nil {
		t.Fatalf("failed to write base2 memory file: %v", err)
	}

	// Create specialized memory that replaces both
	specialized := filepath.Join(memoriesDir, "specialized.md")
	specializedContent := `---
name: SpecializedMemory
replaces: Base1, Base2
---
# Specialized Memory

Replaces both base1 and base2.
`
	if err := os.WriteFile(specialized, []byte(specializedContent), 0644); err != nil {
		t.Fatalf("failed to write specialized memory file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	if err := os.WriteFile(taskFile, []byte("---\n---\n# Test\n"), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Verify the specialized memory is included
	if !strings.Contains(contentStr, "Specialized Memory") {
		t.Errorf("Expected specialized memory to be included")
	}

	// Verify both base memories are NOT included
	if strings.Contains(contentStr, "# Base 1") {
		t.Errorf("Did not expect base1.md to be included - it should be replaced")
	}
	if strings.Contains(contentStr, "# Base 2") {
		t.Errorf("Did not expect base2.md to be included - it should be replaced")
	}
}

func TestMemoryDeduplicationWithSelectors(t *testing.T) {
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

	// Create base memory for production
	baseProd := filepath.Join(memoriesDir, "base-prod.md")
	baseProdContent := `---
name: BaseProdConfig
env: production
---
# Base Production Config
`
	if err := os.WriteFile(baseProd, []byte(baseProdContent), 0644); err != nil {
		t.Fatalf("failed to write base prod memory: %v", err)
	}

	// Create specialized memory that replaces base, also for production
	specProd := filepath.Join(memoriesDir, "spec-prod.md")
	specProdContent := `---
name: SpecializedProdConfig
env: production
replaces: BaseProdConfig
---
# Specialized Production Config
`
	if err := os.WriteFile(specProd, []byte(specProdContent), 0644); err != nil {
		t.Fatalf("failed to write specialized prod memory: %v", err)
	}

	// Create development memory (should be filtered out)
	dev := filepath.Join(memoriesDir, "dev.md")
	devContent := `---
env: development
---
# Development Config
`
	if err := os.WriteFile(dev, []byte(devContent), 0644); err != nil {
		t.Fatalf("failed to write dev memory: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	if err := os.WriteFile(taskFile, []byte("---\n---\n# Test\n"), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run with production selector
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Verify only specialized production config is included
	if !strings.Contains(contentStr, "Specialized Production Config") {
		t.Errorf("Expected specialized production memory to be included")
	}
	if strings.Contains(contentStr, "Base Production Config") {
		t.Errorf("Did not expect base production memory - should be replaced")
	}
	if strings.Contains(contentStr, "Development Config") {
		t.Errorf("Did not expect development memory - should be filtered by selector")
	}
}

func TestMemoryDeduplicationNoReplacement(t *testing.T) {
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

	// Create memory files without any replacement
	mem1 := filepath.Join(memoriesDir, "mem1.md")
	if err := os.WriteFile(mem1, []byte("---\n---\n# Memory 1\nContent 1\n"), 0644); err != nil {
		t.Fatalf("failed to write mem1: %v", err)
	}

	mem2 := filepath.Join(memoriesDir, "mem2.md")
	if err := os.WriteFile(mem2, []byte("---\n---\n# Memory 2\nContent 2\n"), 0644); err != nil {
		t.Fatalf("failed to write mem2: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	if err := os.WriteFile(taskFile, []byte("---\n---\n# Test\n"), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-m", memoriesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check the output
	promptOutput := filepath.Join(outputDir, "prompt.md")
	content, err := os.ReadFile(promptOutput)
	if err != nil {
		t.Fatalf("failed to read prompt output: %v", err)
	}

	contentStr := string(content)

	// Verify both memories are included (no replacement)
	if !strings.Contains(contentStr, "Memory 1") {
		t.Errorf("Expected Memory 1 to be included")
	}
	if !strings.Contains(contentStr, "Memory 2") {
		t.Errorf("Expected Memory 2 to be included")
	}
}
