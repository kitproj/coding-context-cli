package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportCommand(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create AGENTS.md for Codex agent
	agentsFile := filepath.Join(tmpDir, "AGENTS.md")
	agentsContent := `---
env: test
---
# Test Agents

This is a test agents file.
`
	if err := os.WriteFile(agentsFile, []byte(agentsContent), 0644); err != nil {
		t.Fatalf("failed to write AGENTS.md: %v", err)
	}

	// Run the import command
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-o", outputDir, "import", "Codex")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run import command: %v\n%s", err, output)
	}

	// Check output contains the file
	outputStr := string(output)
	if !strings.Contains(outputStr, "Including rule file:") {
		t.Errorf("Expected 'Including rule file:' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "AGENTS.md") {
		t.Errorf("Expected 'AGENTS.md' in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "level 1") {
		t.Errorf("Expected 'level 1' (AncestorLevel) in output, got: %s", outputStr)
	}

	// Check that rules.md was created
	rulesOutput := filepath.Join(outputDir, "rules.md")
	if _, err := os.Stat(rulesOutput); os.IsNotExist(err) {
		t.Errorf("rules.md file was not created")
	}

	// Check content of rules.md
	content, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules.md: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "# Test Agents") {
		t.Errorf("Expected '# Test Agents' in rules.md content")
	}
	if !strings.Contains(contentStr, "This is a test agents file.") {
		t.Errorf("Expected agents file content in rules.md")
	}

	// Check that bootstrap and bootstrap.d were created
	bootstrapFile := filepath.Join(outputDir, "bootstrap")
	if _, err := os.Stat(bootstrapFile); os.IsNotExist(err) {
		t.Errorf("bootstrap file was not created")
	}
	bootstrapDir := filepath.Join(outputDir, "bootstrap.d")
	if _, err := os.Stat(bootstrapDir); os.IsNotExist(err) {
		t.Errorf("bootstrap.d directory was not created")
	}
}

func TestImportWithBootstrap(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create CLAUDE.md for Claude agent
	claudeFile := filepath.Join(tmpDir, "CLAUDE.md")
	claudeContent := `# Claude Rules

Setup instructions for Claude.
`
	if err := os.WriteFile(claudeFile, []byte(claudeContent), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}

	// Create a bootstrap file for CLAUDE.md
	bootstrapFile := filepath.Join(tmpDir, "CLAUDE-bootstrap")
	bootstrapContent := `#!/bin/bash
echo "Setting up Claude"
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// Run the import command for Claude
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-o", outputDir, "import", "Claude")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run import command: %v\n%s", err, output)
	}

	// Check that bootstrap.d contains the bootstrap file
	bootstrapDDir := filepath.Join(outputDir, "bootstrap.d")
	files, err := os.ReadDir(bootstrapDDir)
	if err != nil {
		t.Fatalf("failed to read bootstrap.d dir: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 bootstrap file, got %d", len(files))
	}

	// Check that the bootstrap file has correct content
	if len(files) > 0 {
		bootstrapPath := filepath.Join(bootstrapDDir, files[0].Name())
		content, err := os.ReadFile(bootstrapPath)
		if err != nil {
			t.Fatalf("failed to read bootstrap file: %v", err)
		}
		if string(content) != bootstrapContent {
			t.Errorf("bootstrap content mismatch:\ngot: %q\nwant: %q", string(content), bootstrapContent)
		}

		// Verify the naming format: CLAUDE-bootstrap-<8-hex-chars>
		fileName := files[0].Name()
		if !strings.HasPrefix(fileName, "CLAUDE-bootstrap-") {
			t.Errorf("bootstrap file name should start with 'CLAUDE-bootstrap-', got: %s", fileName)
		}
	}
}

func TestImportUnknownAgent(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Run the import command with unknown agent
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-o", outputDir, "import", "UnknownAgent")
	output, err := cmd.CombinedOutput()

	// Should error
	if err == nil {
		t.Errorf("Expected error for unknown agent, but command succeeded")
	}

	// Check error message
	if !strings.Contains(string(output), "unknown agent") {
		t.Errorf("Expected 'unknown agent' error message, got: %s", string(output))
	}
}

func TestImportCursorWithDirectory(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create .cursor/rules directory
	cursorRulesDir := filepath.Join(tmpDir, ".cursor", "rules")
	if err := os.MkdirAll(cursorRulesDir, 0755); err != nil {
		t.Fatalf("failed to create .cursor/rules dir: %v", err)
	}

	// Create rule files in .cursor/rules
	rule1 := filepath.Join(cursorRulesDir, "rule1.md")
	rule1Content := `# Cursor Rule 1

First cursor rule.
`
	if err := os.WriteFile(rule1, []byte(rule1Content), 0644); err != nil {
		t.Fatalf("failed to write rule1.md: %v", err)
	}

	rule2 := filepath.Join(cursorRulesDir, "rule2.mdc")
	rule2Content := `# Cursor Rule 2

Second cursor rule in .mdc format.
`
	if err := os.WriteFile(rule2, []byte(rule2Content), 0644); err != nil {
		t.Fatalf("failed to write rule2.mdc: %v", err)
	}

	// Run the import command for Cursor
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-o", outputDir, "import", "Cursor")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run import command: %v\n%s", err, output)
	}

	// Check that rules.md contains both files
	rulesOutput := filepath.Join(outputDir, "rules.md")
	content, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules.md: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "# Cursor Rule 1") {
		t.Errorf("Expected '# Cursor Rule 1' in rules.md content")
	}
	if !strings.Contains(contentStr, "# Cursor Rule 2") {
		t.Errorf("Expected '# Cursor Rule 2' in rules.md content")
	}
	if !strings.Contains(contentStr, "First cursor rule") {
		t.Errorf("Expected first rule content in rules.md")
	}
	if !strings.Contains(contentStr, "Second cursor rule in .mdc format") {
		t.Errorf("Expected second rule content (.mdc) in rules.md")
	}
}

func TestBootstrapCommand(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create AGENTS.md
	agentsFile := filepath.Join(tmpDir, "AGENTS.md")
	agentsContent := `# Test

Test content.
`
	if err := os.WriteFile(agentsFile, []byte(agentsContent), 0644); err != nil {
		t.Fatalf("failed to write AGENTS.md: %v", err)
	}

	// Create a bootstrap file
	bootstrapFile := filepath.Join(tmpDir, "AGENTS-bootstrap")
	markerFile := filepath.Join(outputDir, "bootstrap-ran.txt")
	bootstrapContent := `#!/bin/bash
echo "Bootstrap executed" > ` + markerFile + `
`
	if err := os.WriteFile(bootstrapFile, []byte(bootstrapContent), 0755); err != nil {
		t.Fatalf("failed to write bootstrap file: %v", err)
	}

	// First run import to create bootstrap files
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-o", outputDir, "import", "Codex")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run import command: %v\n%s", err, output)
	}

	// Then run bootstrap command
	cmd = exec.Command(binaryPath, "-C", tmpDir, "-o", outputDir, "bootstrap")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run bootstrap command: %v\n%s", err, output)
	}

	// Check that the marker file was created
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

func TestCommandWithoutArgs(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Run without any command
	cmd = exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()

	// Should error
	if err == nil {
		t.Errorf("Expected error when running without command")
	}

	// Check that usage is displayed
	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage:") {
		t.Errorf("Expected usage message in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "import <agent>") {
		t.Errorf("Expected 'import <agent>' in usage message, got: %s", outputStr)
	}
}

func TestImportWithoutAgent(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	tmpDir := t.TempDir()

	// Run import without agent name
	cmd = exec.Command(binaryPath, "-C", tmpDir, "import")
	output, err := cmd.CombinedOutput()

	// Should error
	if err == nil {
		t.Errorf("Expected error when running import without agent name")
	}

	// Check error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "usage:") {
		t.Errorf("Expected usage error message, got: %s", outputStr)
	}
}

func TestImportWithAncestorPaths(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Create a directory hierarchy with AGENTS.md at different levels
	tmpDir := t.TempDir()
	rootAgents := filepath.Join(tmpDir, "AGENTS.md")
	sub1Dir := filepath.Join(tmpDir, "sub1")
	sub1Agents := filepath.Join(sub1Dir, "AGENTS.md")
	sub2Dir := filepath.Join(sub1Dir, "sub2")
	outputDir := filepath.Join(sub2Dir, "output")

	// Create directories
	if err := os.MkdirAll(sub2Dir, 0755); err != nil {
		t.Fatalf("failed to create directory structure: %v", err)
	}

	// Create AGENTS.md at root level
	rootContent := `# Root Level Rules

This is from the root.
`
	if err := os.WriteFile(rootAgents, []byte(rootContent), 0644); err != nil {
		t.Fatalf("failed to write root AGENTS.md: %v", err)
	}

	// Create AGENTS.md at sub1 level
	sub1Content := `# Sub1 Level Rules

This is from sub1.
`
	if err := os.WriteFile(sub1Agents, []byte(sub1Content), 0644); err != nil {
		t.Fatalf("failed to write sub1 AGENTS.md: %v", err)
	}

	// Run import from sub2 directory (should find both sub1 and root AGENTS.md)
	cmd = exec.Command(binaryPath, "-C", sub2Dir, "-o", outputDir, "import", "Codex")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run import command: %v\n%s", err, output)
	}

	// Check output mentions both files
	outputStr := string(output)
	if !strings.Contains(outputStr, "sub1/AGENTS.md") {
		t.Errorf("Expected sub1/AGENTS.md in output, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, tmpDir+"/AGENTS.md") {
		t.Errorf("Expected root AGENTS.md in output, got: %s", outputStr)
	}

	// Check that rules.md contains both files' content
	rulesOutput := filepath.Join(outputDir, "rules.md")
	content, err := os.ReadFile(rulesOutput)
	if err != nil {
		t.Fatalf("failed to read rules.md: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "# Sub1 Level Rules") {
		t.Errorf("Expected sub1 content in rules.md")
	}
	if !strings.Contains(contentStr, "This is from sub1.") {
		t.Errorf("Expected sub1 text in rules.md")
	}
	if !strings.Contains(contentStr, "# Root Level Rules") {
		t.Errorf("Expected root content in rules.md")
	}
	if !strings.Contains(contentStr, "This is from the root.") {
		t.Errorf("Expected root text in rules.md")
	}

	// Verify files are in correct order (closer files first, based on level)
	sub1Pos := strings.Index(contentStr, "# Sub1 Level Rules")
	rootPos := strings.Index(contentStr, "# Root Level Rules")
	if sub1Pos > rootPos {
		t.Errorf("Expected sub1 content before root content (closer to cwd should be first)")
	}
}

func TestMultipleAgents(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "coding-context")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		agent         string
		setupFiles    map[string]string
		expectedFiles []string
	}{
		{
			name:  "Claude",
			agent: "Claude",
			setupFiles: map[string]string{
				"CLAUDE.local.md": "# Claude Local\n",
				"CLAUDE.md":       "# Claude Global\n",
			},
			expectedFiles: []string{"CLAUDE.local.md", "CLAUDE.md"},
		},
		{
			name:  "Gemini",
			agent: "Gemini",
			setupFiles: map[string]string{
				".gemini/styleguide.md": "# Gemini Styleguide\n",
				"GEMINI.md":             "# Gemini Rules\n",
			},
			expectedFiles: []string{".gemini/styleguide.md", "GEMINI.md"},
		},
		{
			name:  "Cursor",
			agent: "Cursor",
			setupFiles: map[string]string{
				".cursor/rules/rule1.md":  "# Cursor Rule 1\n",
				".cursor/rules/rule2.mdc": "# Cursor Rule 2\n",
			},
			expectedFiles: []string{".cursor/rules/rule1.md", ".cursor/rules/rule2.mdc"},
		},
		{
			name:  "Copilot",
			agent: "Copilot",
			setupFiles: map[string]string{
				".github/copilot-instructions.md": "# Copilot Instructions\n",
				"AGENTS.md":                       "# Agents\n",
			},
			expectedFiles: []string{".github/copilot-instructions.md", "AGENTS.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subdirectory for this test
			agentDir := filepath.Join(tmpDir, tt.name)
			outputDir := filepath.Join(agentDir, "output")

			// Setup files
			for path, content := range tt.setupFiles {
				fullPath := filepath.Join(agentDir, path)
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create directory %s: %v", dir, err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("failed to write file %s: %v", path, err)
				}
			}

			// Run import
			cmd := exec.Command(binaryPath, "-C", agentDir, "-o", outputDir, "import", tt.agent)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to run import for %s: %v\n%s", tt.agent, err, output)
			}

			// Check that rules.md was created and is not empty
			rulesOutput := filepath.Join(outputDir, "rules.md")
			content, err := os.ReadFile(rulesOutput)
			if err != nil {
				t.Fatalf("failed to read rules.md: %v", err)
			}
			if len(content) == 0 {
				t.Errorf("rules.md is empty for agent %s", tt.agent)
			}

			// Check that expected files are mentioned in output
			outputStr := string(output)
			for _, expectedFile := range tt.expectedFiles {
				if !strings.Contains(outputStr, expectedFile) {
					t.Errorf("Expected %s to be mentioned in output for agent %s, got: %s", expectedFile, tt.agent, outputStr)
				}
			}
		})
	}
}

func TestBootstrapCommandWithoutImport(t *testing.T) {
// Build the binary
binaryPath := filepath.Join(t.TempDir(), "coding-context")
cmd := exec.Command("go", "build", "-o", binaryPath, ".")
if output, err := cmd.CombinedOutput(); err != nil {
t.Fatalf("failed to build binary: %v\n%s", err, output)
}

tmpDir := t.TempDir()
outputDir := filepath.Join(tmpDir, "output")

// Try to run bootstrap without importing first
cmd = exec.Command(binaryPath, "-C", tmpDir, "-o", outputDir, "bootstrap")
output, err := cmd.CombinedOutput()

// Should error
if err == nil {
t.Errorf("Expected error when running bootstrap without import")
}

// Check error message
outputStr := string(output)
if !strings.Contains(outputStr, "bootstrap file not found") {
t.Errorf("Expected 'bootstrap file not found' error message, got: %s", outputStr)
}
if !strings.Contains(outputStr, "Run 'import' command first") {
t.Errorf("Expected 'Run 'import' command first' in error message, got: %s", outputStr)
}
}
