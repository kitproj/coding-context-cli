package codingcontext

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileReferenceIntegration(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create a sample source file
	srcDir := filepath.Join(tmpDir, "src", "components")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	buttonContent := `function Button({ onClick, children }) {
  return <button onClick={onClick}>{children}</button>;
}

export default Button;`

	buttonPath := filepath.Join(srcDir, "Button.tsx")
	if err := os.WriteFile(buttonPath, []byte(buttonContent), 0o644); err != nil {
		t.Fatalf("Failed to create Button.tsx: %v", err)
	}

	// Create .agents directory structure
	agentsDir := filepath.Join(tmpDir, ".agents")
	tasksDir := filepath.Join(agentsDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("Failed to create tasks directory: %v", err)
	}

	// Create a task file that references the Button.tsx file
	taskContent := `---
task_name: review-component
---

# Review Component

Review the component in ${file:src/components/Button.tsx}.
Check for performance issues and suggest improvements.`

	taskPath := filepath.Join(tasksDir, "review-component.md")
	if err := os.WriteFile(taskPath, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("Failed to create task file: %v", err)
	}

	// Create the context and run it
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	ctx := context.Background()
	result, err := cc.Run(ctx, "review-component")
	if err != nil {
		t.Fatalf("Failed to run context: %v", err)
	}

	// Verify the result contains the file content
	if !strings.Contains(result.Task.Content, "function Button") {
		t.Errorf("Expected task content to include Button component code, got: %s", result.Task.Content)
	}

	if !strings.Contains(result.Task.Content, "File: src/components/Button.tsx") {
		t.Errorf("Expected task content to include file header, got: %s", result.Task.Content)
	}

	if !strings.Contains(result.Task.Content, "Review the component in") {
		t.Errorf("Expected task content to include original text, got: %s", result.Task.Content)
	}

	if !strings.Contains(result.Task.Content, "```") {
		t.Errorf("Expected task content to include code fence, got: %s", result.Task.Content)
	}

	// Verify that the original task description is preserved
	if !strings.Contains(result.Task.Content, "Check for performance issues") {
		t.Errorf("Expected task content to preserve original text after file reference, got: %s", result.Task.Content)
	}
}

func TestFileReferenceWithMultipleFiles(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create multiple source files
	if err := os.MkdirAll(filepath.Join(tmpDir, "src"), 0o755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	file1Content := "const API_URL = 'https://api.example.com';"
	file2Content := "export function fetchData() { /* implementation */ }"

	if err := os.WriteFile(filepath.Join(tmpDir, "src", "config.ts"), []byte(file1Content), 0o644); err != nil {
		t.Fatalf("Failed to create config.ts: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "src", "api.ts"), []byte(file2Content), 0o644); err != nil {
		t.Fatalf("Failed to create api.ts: %v", err)
	}

	// Create .agents directory structure
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("Failed to create tasks directory: %v", err)
	}

	// Create a task file that references multiple files
	taskContent := `---
task_name: review-api
---

# Review API Configuration

Compare ${file:src/config.ts} and ${file:src/api.ts} for consistency.`

	taskPath := filepath.Join(tasksDir, "review-api.md")
	if err := os.WriteFile(taskPath, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("Failed to create task file: %v", err)
	}

	// Create the context and run it
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	ctx := context.Background()
	result, err := cc.Run(ctx, "review-api")
	if err != nil {
		t.Fatalf("Failed to run context: %v", err)
	}

	// Verify both files are included
	if !strings.Contains(result.Task.Content, "API_URL") {
		t.Errorf("Expected task content to include config.ts content")
	}

	if !strings.Contains(result.Task.Content, "fetchData") {
		t.Errorf("Expected task content to include api.ts content")
	}

	if !strings.Contains(result.Task.Content, "File: src/config.ts") {
		t.Errorf("Expected task content to include config.ts header")
	}

	if !strings.Contains(result.Task.Content, "File: src/api.ts") {
		t.Errorf("Expected task content to include api.ts header")
	}
}

func TestFileReferenceNotFound(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create .agents directory structure
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("Failed to create tasks directory: %v", err)
	}

	// Create a task file that references a non-existent file
	taskContent := `---
task_name: review-missing
---

# Review Missing File

Review ${file:nonexistent.txt} for issues.`

	taskPath := filepath.Join(tasksDir, "review-missing.md")
	if err := os.WriteFile(taskPath, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("Failed to create task file: %v", err)
	}

	// Create the context and run it
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	ctx := context.Background()
	result, err := cc.Run(ctx, "review-missing")

	// Should succeed but the file reference should remain unexpanded
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The placeholder should remain in the content since the file wasn't found
	if !strings.Contains(result.Task.Content, "${file:nonexistent.txt}") {
		t.Errorf("Expected unexpanded placeholder to remain in content, got: %s", result.Task.Content)
	}
}
