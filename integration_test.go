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
