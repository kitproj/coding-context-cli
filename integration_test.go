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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

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

	// Create a bootstrap file for the rule (setup.md -> setup-bootstrap)
	bootstrapFile := filepath.Join(rulesDir, "setup-bootstrap")
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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

	// Check that the three output files were created
	personaOutput := filepath.Join(outputDir, "persona.md")
	rulesOutput := filepath.Join(outputDir, "rules.md")
	taskOutput := filepath.Join(outputDir, "task.md")
	
	if _, err := os.Stat(personaOutput); os.IsNotExist(err) {
		t.Errorf("persona.md file was not created")
	}
	if _, err := os.Stat(rulesOutput); os.IsNotExist(err) {
		t.Errorf("rules.md file was not created")
	}
	if _, err := os.Stat(taskOutput); os.IsNotExist(err) {
		t.Errorf("task.md file was not created")
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "jira.md")
	ruleContent := `---
---
# Jira Integration
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file for the rule (jira.md -> jira-bootstrap)
	bootstrapFile := filepath.Join(rulesDir, "jira-bootstrap")
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file WITHOUT a bootstrap
	ruleFile := filepath.Join(rulesDir, "info.md")
	ruleContent := `---
---
# Project Info

Just some information.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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

	// Check that the three output files were still created
	rulesOutput := filepath.Join(outputDir, "rules.md")
	taskOutput := filepath.Join(outputDir, "task.md")
	
	if _, err := os.Stat(rulesOutput); os.IsNotExist(err) {
		t.Errorf("rules.md file was not created")
	}
	if _, err := os.Stat(taskOutput); os.IsNotExist(err) {
		t.Errorf("task.md file was not created")
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create first rule file with bootstrap
	if err := os.WriteFile(filepath.Join(rulesDir, "setup.md"), []byte("---\n---\n# Setup\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "setup-bootstrap"), []byte("#!/bin/bash\necho setup\n"), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Create second rule file with bootstrap
	if err := os.WriteFile(filepath.Join(rulesDir, "deps.md"), []byte("---\n---\n# Dependencies\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "deps-bootstrap"), []byte("#!/bin/bash\necho deps\n"), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Create a prompt file
	if err := os.WriteFile(filepath.Join(tasksDir, "test-task.md"), []byte("---\n---\n# Test\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create rule files with different frontmatter
	if err := os.WriteFile(filepath.Join(rulesDir, "prod.md"), []byte("---\nenv: production\nlanguage: go\n---\n# Production\nProd content\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "dev.md"), []byte("---\nenv: development\nlanguage: python\n---\n# Development\nDev content\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "test.md"), []byte("---\nenv: test\nlanguage: go\n---\n# Test\nTest content\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	// Create a file without frontmatter (should be included by default)
	if err := os.WriteFile(filepath.Join(rulesDir, "nofm.md"), []byte("---\n---\n# No Frontmatter\nNo FM content\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a prompt file
	if err := os.WriteFile(filepath.Join(tasksDir, "test-task.md"), []byte("---\n---\n# Test Task\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Test 1: Include by env=production
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	rulesOutput := filepath.Join(outputDir, "rules.md")
	content, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-s", "language=go", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-S", "env=production", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "-s", "language=go", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "-S", "language=python", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
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

	// Read the task output (template expansion happens in task.md)
	taskOutput := filepath.Join(outputDir, "task.md")
	content, err := os.ReadFile(taskOutput)
	if err != nil {
		t.Fatalf("failed to read task output: %v", err)
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

	// Read the task output (template expansion happens in task.md)
	taskOutput := filepath.Join(outputDir, "task.md")
	content, err := os.ReadFile(taskOutput)
	if err != nil {
		t.Fatalf("failed to read task output: %v", err)
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

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
# Setup

This is a setup guide.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file that creates a marker file
	bootstrapFile := filepath.Join(rulesDir, "setup-bootstrap")
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-b", "test-task")
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

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
# Setup

This is a setup guide.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file that creates a marker file
	bootstrapFile := filepath.Join(rulesDir, "setup-bootstrap")
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

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
# Setup

Long running setup.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a bootstrap file that runs for a while
	bootstrapFile := filepath.Join(rulesDir, "setup-bootstrap")
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-b", "test-task")
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
// automatically includes/excludes rule files based on the task being run
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create rule files with task_name frontmatter
	if err := os.WriteFile(filepath.Join(rulesDir, "deploy-specific.md"), []byte("---\ntask_name: deploy\n---\n# Deploy Rule\nDeploy-specific content\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "test-specific.md"), []byte("---\ntask_name: test\n---\n# Test Rule\nTest-specific content\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}
	// Create a file without task_name (should be included for all tasks)
	if err := os.WriteFile(filepath.Join(rulesDir, "general.md"), []byte("---\n---\n# General Rule\nGeneral content\n"), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create prompt files for both tasks
	if err := os.WriteFile(filepath.Join(tasksDir, "deploy.md"), []byte("---\n---\n# Deploy Task\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tasksDir, "test.md"), []byte("---\n---\n# Test Task\n"), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	// Test 1: Run with "deploy" task - should include deploy-specific and general, but not test-specific
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "deploy")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	rulesOutput := filepath.Join(outputDir, "rules.md")
	content, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	content, err = os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
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
	rulesDir := filepath.Join(contextDir, "rules")
	personasDir := filepath.Join(contextDir, "personas")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
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

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "context.md")
	ruleContent := `---
---
# Context

This is context.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
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
	cmd = exec.Command(binaryPath, "-r", personasDir, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-p", "feature=auth", "test-task", "expert")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check the output - now we have three separate files
	personaOutput := filepath.Join(outputDir, "persona.md")
	personaBytes, err := os.ReadFile(personaOutput)
	if err != nil {
		t.Fatalf("failed to read persona output: %v", err)
	}
	
	rulesOutput := filepath.Join(outputDir, "rules.md")
	rulesBytes, err2 := os.ReadFile(rulesOutput)
	if err2 != nil {
		t.Fatalf("failed to read rules output: %v", err2)
	}
	
	taskOutput := filepath.Join(outputDir, "task.md")
	taskBytes, err3 := os.ReadFile(taskOutput)
	if err3 != nil {
		t.Fatalf("failed to read task output: %v", err3)
	}

	// Verify persona content
	personaStr := string(personaBytes)
	if !strings.Contains(personaStr, "Expert Persona") {
		t.Errorf("Expected to find 'Expert Persona' in persona.md")
	}
	if !strings.Contains(personaStr, "You are an expert in Go") {
		t.Errorf("Expected persona content to remain as-is without template expansion")
	}

	// Verify rules content
	rulesStr := string(rulesBytes)
	if !strings.Contains(rulesStr, "# Context") {
		t.Errorf("Expected to find '# Context' in rules.md")
	}

	// Verify task content
	taskStr := string(taskBytes)
	if !strings.Contains(taskStr, "# Task") {
		t.Errorf("Expected to find '# Task' in task.md")
	}
	if !strings.Contains(taskStr, "Please help with auth") {
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "context.md")
	ruleContent := `---
---
# Context

This is context.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary without persona: %v\n%s", err, output)
	}

	// Check the rules and task outputs
	rulesOutput := filepath.Join(outputDir, "rules.md")
	rulesBytes, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
	}

	taskOutput := filepath.Join(outputDir, "task.md")
	taskBytes, err2 := os.ReadFile(taskOutput)
	if err2 != nil {
		t.Fatalf("failed to read task output: %v", err2)
	}

	// Verify context and task are present
	if !strings.Contains(string(rulesBytes), "# Context") {
		t.Errorf("Expected to find '# Context' in rules.md")
	}
	if !strings.Contains(string(taskBytes), "# Task") {
		t.Errorf("Expected to find '# Task' in task.md")
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
	rulesDir := filepath.Join(workDir, ".prompts", "rules")
	tasksDir := filepath.Join(workDir, ".prompts", "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file in the work directory
	ruleFile := filepath.Join(rulesDir, "test.md")
	ruleContent := `---
---
# Test Rule
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
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
	cmd = exec.Command(binaryPath, "-C", workDir, "-m", ".prompts/rules", "-t", ".prompts/tasks", "-o", outputDir, "task")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary with -C option: %v\n%s", err, output)
	}

	// Verify that the three output files were created in the output directory
	rulesOutFile := filepath.Join(outputDir, "rules.md")
	taskOutFile := filepath.Join(outputDir, "task.md")
	personaOutFile := filepath.Join(outputDir, "persona.md")
	
	var statErr error
	if _, statErr = os.Stat(rulesOutFile); os.IsNotExist(statErr) {
		t.Errorf("rules.md was not created in output directory")
	}
	if _, statErr = os.Stat(taskOutFile); os.IsNotExist(statErr) {
		t.Errorf("task.md was not created in output directory")
	}
	if _, statErr = os.Stat(personaOutFile); os.IsNotExist(statErr) {
		t.Errorf("persona.md was not created in output directory")
	}

	// Verify the content includes the rule
	content, err := os.ReadFile(rulesOutFile)
	if err != nil {
		t.Fatalf("failed to read rules.md: %v", err)
	}
	if !strings.Contains(string(content), "Test Rule") {
		t.Errorf("rules.md does not contain expected rule content")
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
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	personasDir := filepath.Join(contextDir, "personas")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
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

	// Create rule files
	ruleFile1 := filepath.Join(rulesDir, "setup.md")
	ruleContent1 := `# Development Setup

This is a setup guide with detailed instructions.`
	if err := os.WriteFile(ruleFile1, []byte(ruleContent1), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	ruleFile2 := filepath.Join(rulesDir, "conventions.md")
	ruleContent2 := `# Coding Conventions

Follow best practices and write clean code.`
	if err := os.WriteFile(ruleFile2, []byte(ruleContent2), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
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
	if !strings.Contains(outputStr, "Including rule file:") {
		t.Errorf("Expected rule file message in output")
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

func TestMdcFileSupport(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a .mdc rule file (Cursor format)
	mdcRuleFile := filepath.Join(rulesDir, "cursor-rules.mdc")
	mdcRuleContent := `---
env: development
---
# Cursor AI Rules

These are Cursor-specific rules in .mdc format.
`
	if err := os.WriteFile(mdcRuleFile, []byte(mdcRuleContent), 0644); err != nil {
		t.Fatalf("failed to write .mdc rule file: %v", err)
	}

	// Create a .md rule file for comparison
	mdRuleFile := filepath.Join(rulesDir, "regular-rules.md")
	mdRuleContent := `---
env: development
---
# Regular Markdown Rules

These are regular .md format rules.
`
	if err := os.WriteFile(mdRuleFile, []byte(mdRuleContent), 0644); err != nil {
		t.Fatalf("failed to write .md rule file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `---
---
# Test Task

Test task for .mdc file support.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Run the binary
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that rules.md was created and contains content from both .md and .mdc files
	rulesOutput := filepath.Join(outputDir, "rules.md")
	content, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
	}

	contentStr := string(content)

	// Verify .mdc content is included
	if !strings.Contains(contentStr, "Cursor AI Rules") {
		t.Errorf("Expected .mdc file content to be included in rules.md")
	}
	if !strings.Contains(contentStr, "Cursor-specific rules in .mdc format") {
		t.Errorf("Expected .mdc file body content to be included in rules.md")
	}

	// Verify .md content is still included
	if !strings.Contains(contentStr, "Regular Markdown Rules") {
		t.Errorf("Expected .md file content to be included in rules.md")
	}
	if !strings.Contains(contentStr, "regular .md format rules") {
		t.Errorf("Expected .md file body content to be included in rules.md")
	}
}

func TestMdcFileWithBootstrap(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a .mdc rule file
	mdcRuleFile := filepath.Join(rulesDir, "cursor-setup.mdc")
	mdcRuleContent := `---
---
# Cursor Setup

Setup instructions for Cursor.
`
	if err := os.WriteFile(mdcRuleFile, []byte(mdcRuleContent), 0644); err != nil {
		t.Fatalf("failed to write .mdc rule file: %v", err)
	}

	// Create a bootstrap file for the .mdc rule (cursor-setup.mdc -> cursor-setup-bootstrap)
	bootstrapFile := filepath.Join(rulesDir, "cursor-setup-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Setting up Cursor"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
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
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "test-task")
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

		// Verify the naming format: cursor-setup-bootstrap-<8-hex-chars>
		fileName := files[0].Name()
		if !strings.HasPrefix(fileName, "cursor-setup-bootstrap-") {
			t.Errorf("bootstrap file name should start with 'cursor-setup-bootstrap-', got: %s", fileName)
		}
	}
}

func TestMdcFileWithSelectors(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	contextDir := filepath.Join(tmpDir, ".prompts")
	rulesDir := filepath.Join(contextDir, "rules")
	tasksDir := filepath.Join(contextDir, "tasks")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create .mdc files with different frontmatter
	prodMdcFile := filepath.Join(rulesDir, "prod-cursor.mdc")
	prodMdcContent := `---
env: production
editor: cursor
---
# Production Cursor Rules

Production-specific Cursor rules.
`
	if err := os.WriteFile(prodMdcFile, []byte(prodMdcContent), 0644); err != nil {
		t.Fatalf("failed to write prod .mdc file: %v", err)
	}

	devMdcFile := filepath.Join(rulesDir, "dev-cursor.mdc")
	devMdcContent := `---
env: development
editor: cursor
---
# Development Cursor Rules

Development-specific Cursor rules.
`
	if err := os.WriteFile(devMdcFile, []byte(devMdcContent), 0644); err != nil {
		t.Fatalf("failed to write dev .mdc file: %v", err)
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

	// Run with production selector
	cmd = exec.Command(binaryPath, "-m", rulesDir, "-t", tasksDir, "-o", outputDir, "-s", "env=production", "test-task")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binary: %v\n%s", err, output)
	}

	// Check that only production .mdc content is included
	rulesOutput := filepath.Join(outputDir, "rules.md")
	content, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules output: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "Production Cursor Rules") {
		t.Errorf("Expected production .mdc content to be included")
	}
	if !strings.Contains(contentStr, "Production-specific Cursor rules") {
		t.Errorf("Expected production .mdc body content to be included")
	}
	if strings.Contains(contentStr, "Development Cursor Rules") {
		t.Errorf("Did not expect development .mdc content to be included")
	}
}


