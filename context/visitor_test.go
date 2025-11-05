package context

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CustomVisitor is a test visitor that collects rule information
type CustomVisitor struct {
	VisitedRules []*Rule
	stderr       *bytes.Buffer
}

func (v *CustomVisitor) VisitRule(ctx context.Context, rule *Rule) error {
	v.VisitedRules = append(v.VisitedRules, rule)
	return nil
}

func TestAssembler_WithCustomVisitor(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create rule files
	rule1File := filepath.Join(rulesDir, "rule1.md")
	rule1Content := `---
language: go
---
# Rule 1

This is rule 1.
`
	if err := os.WriteFile(rule1File, []byte(rule1Content), 0644); err != nil {
		t.Fatalf("failed to write rule file 1: %v", err)
	}

	rule2File := filepath.Join(rulesDir, "rule2.md")
	rule2Content := `---
language: python
---
# Rule 2

This is rule 2.
`
	if err := os.WriteFile(rule2File, []byte(rule2Content), 0644); err != nil {
		t.Fatalf("failed to write rule file 2: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Create custom visitor
	var stderr bytes.Buffer
	customVisitor := &CustomVisitor{
		VisitedRules: make([]*Rule, 0),
		stderr:       &stderr,
	}

	// Test assembling context with custom visitor
	var stdout bytes.Buffer
	assembler := NewAssembler(Config{
		WorkDir:   tmpDir,
		TaskName:  "test-task",
		Params:    make(ParamMap),
		Selectors: make(SelectorMap),
		Stdout:    &stdout,
		Stderr:    &stderr,
		Visitor:   customVisitor,
	})

	ctx := context.Background()
	if err := assembler.Assemble(ctx); err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	// Verify that the custom visitor was called
	if len(customVisitor.VisitedRules) != 2 {
		t.Errorf("expected 2 rules to be visited, got %d", len(customVisitor.VisitedRules))
	}

	// Verify rule information
	for _, rule := range customVisitor.VisitedRules {
		if rule.Path == "" {
			t.Errorf("rule path should not be empty")
		}
		if rule.Content == "" {
			t.Errorf("rule content should not be empty")
		}
		if rule.Tokens == 0 {
			t.Errorf("rule tokens should not be zero")
		}
		if rule.Frontmatter == nil {
			t.Errorf("rule frontmatter should not be nil")
		}
	}

	// Verify that rules were visited in order
	foundRule1 := false
	foundRule2 := false
	for _, rule := range customVisitor.VisitedRules {
		if strings.Contains(rule.Content, "Rule 1") {
			foundRule1 = true
		}
		if strings.Contains(rule.Content, "Rule 2") {
			foundRule2 = true
		}
	}
	if !foundRule1 {
		t.Errorf("rule1 was not visited")
	}
	if !foundRule2 {
		t.Errorf("rule2 was not visited")
	}
}

func TestAssembler_WithDefaultVisitor(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create a rule file
	ruleFile := filepath.Join(rulesDir, "setup.md")
	ruleContent := `# Development Setup

This is a setup guide.
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("failed to write rule file: %v", err)
	}

	// Create a task file
	taskFile := filepath.Join(tasksDir, "test-task.md")
	taskContent := `# Test Task

Please help with this task.
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Test with default visitor (should write to stdout)
	var stdout, stderr bytes.Buffer
	assembler := NewAssembler(Config{
		WorkDir:   tmpDir,
		TaskName:  "test-task",
		Params:    make(ParamMap),
		Selectors: make(SelectorMap),
		Stdout:    &stdout,
		Stderr:    &stderr,
		// Visitor not specified, should use default
	})

	ctx := context.Background()
	if err := assembler.Assemble(ctx); err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	// Check that rule content is present in stdout (default behavior)
	outputStr := stdout.String()
	if !strings.Contains(outputStr, "# Development Setup") {
		t.Errorf("rule content not found in stdout with default visitor")
	}

	// Check that task content is present
	if !strings.Contains(outputStr, "# Test Task") {
		t.Errorf("task content not found in stdout")
	}

	// Check stderr for progress messages
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Including rule file") {
		t.Errorf("progress messages not found in stderr with default visitor")
	}
}
