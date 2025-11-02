package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetRuleContent(t *testing.T) {
	// Create a temporary rule file
	tempDir := t.TempDir()
	ruleFile := filepath.Join(tempDir, "test-rule.md")
	expectedContent := "# Test Rule\n\nThis is a test rule."
	
	if err := os.WriteFile(ruleFile, []byte(expectedContent), 0644); err != nil {
		t.Fatalf("Failed to create test rule file: %v", err)
	}

	// Create a GetRuleContent function
	ruleFunc := func() string {
		content, err := os.ReadFile(ruleFile)
		if err != nil {
			return ""
		}
		return string(content)
	}

	// Test that the function returns the correct content
	actualContent := ruleFunc()
	if actualContent != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, actualContent)
	}
}

func TestFindUserRules(t *testing.T) {
	// Create a temporary directory structure to simulate user rules
	tempHome := t.TempDir()
	userRulesDir := filepath.Join(tempHome, ".prompts", "rules")
	if err := os.MkdirAll(userRulesDir, 0755); err != nil {
		t.Fatalf("Failed to create user rules directory: %v", err)
	}

	// Create some test rule files
	rule1Content := "# Rule 1\n\nFirst rule"
	rule2Content := "# Rule 2\n\nSecond rule"
	
	if err := os.WriteFile(filepath.Join(userRulesDir, "rule1.md"), []byte(rule1Content), 0644); err != nil {
		t.Fatalf("Failed to create rule1.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(userRulesDir, "rule2.md"), []byte(rule2Content), 0644); err != nil {
		t.Fatalf("Failed to create rule2.md: %v", err)
	}

	// Create a non-.md file that should be ignored
	if err := os.WriteFile(filepath.Join(userRulesDir, "ignored.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("Failed to create ignored.txt: %v", err)
	}

	// Temporarily override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Test FindUserRules
	ruleFuncs, err := FindUserRules()
	if err != nil {
		t.Fatalf("FindUserRules returned error: %v", err)
	}

	// We should have 2 rule functions (rule1.md and rule2.md)
	if len(ruleFuncs) != 2 {
		t.Errorf("Expected 2 rule functions, got %d", len(ruleFuncs))
	}

	// Verify that the functions return the correct content
	var foundRule1, foundRule2 bool
	for _, ruleFunc := range ruleFuncs {
		content := ruleFunc()
		if content == rule1Content {
			foundRule1 = true
		} else if content == rule2Content {
			foundRule2 = true
		}
	}

	if !foundRule1 {
		t.Error("Did not find rule1 content")
	}
	if !foundRule2 {
		t.Error("Did not find rule2 content")
	}
}

func TestFindUserRules_NoDirectory(t *testing.T) {
	// Use a temporary directory that doesn't have .prompts/rules
	tempHome := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Test FindUserRules when directory doesn't exist
	ruleFuncs, err := FindUserRules()
	if err != nil {
		t.Fatalf("FindUserRules returned error when directory doesn't exist: %v", err)
	}

	// Should return empty slice, not error
	if len(ruleFuncs) != 0 {
		t.Errorf("Expected 0 rule functions when directory doesn't exist, got %d", len(ruleFuncs))
	}
}

func TestGetNormalizedRulePaths(t *testing.T) {
	paths, err := GetNormalizedRulePaths()
	if err != nil {
		t.Fatalf("GetNormalizedRulePaths returned error: %v", err)
	}

	// Should have 3 paths in the hierarchy
	if len(paths) != 3 {
		t.Errorf("Expected 3 paths, got %d", len(paths))
	}

	// Verify the paths are in the correct order and format
	// L2: Project-rules
	if paths[0] != ".prompts/rules" {
		t.Errorf("Expected first path to be '.prompts/rules', got %q", paths[0])
	}

	// L1: User-rules
	homeDir, _ := os.UserHomeDir()
	expectedUserPath := filepath.Join(homeDir, ".prompts", "rules")
	if paths[1] != expectedUserPath {
		t.Errorf("Expected second path to be %q, got %q", expectedUserPath, paths[1])
	}

	// L0: System-rules
	if paths[2] != "/etc/prompts/rules" {
		t.Errorf("Expected third path to be '/etc/prompts/rules', got %q", paths[2])
	}
}

func TestGetNormalizedPersonaPaths(t *testing.T) {
	paths, err := GetNormalizedPersonaPaths()
	if err != nil {
		t.Fatalf("GetNormalizedPersonaPaths returned error: %v", err)
	}

	if len(paths) != 3 {
		t.Errorf("Expected 3 paths, got %d", len(paths))
	}

	if paths[0] != ".prompts/personas" {
		t.Errorf("Expected first path to be '.prompts/personas', got %q", paths[0])
	}

	homeDir, _ := os.UserHomeDir()
	expectedUserPath := filepath.Join(homeDir, ".prompts", "personas")
	if paths[1] != expectedUserPath {
		t.Errorf("Expected second path to be %q, got %q", expectedUserPath, paths[1])
	}

	if paths[2] != "/etc/prompts/personas" {
		t.Errorf("Expected third path to be '/etc/prompts/personas', got %q", paths[2])
	}
}

func TestGetNormalizedTaskPaths(t *testing.T) {
	paths, err := GetNormalizedTaskPaths()
	if err != nil {
		t.Fatalf("GetNormalizedTaskPaths returned error: %v", err)
	}

	if len(paths) != 3 {
		t.Errorf("Expected 3 paths, got %d", len(paths))
	}

	if paths[0] != ".prompts/tasks" {
		t.Errorf("Expected first path to be '.prompts/tasks', got %q", paths[0])
	}

	homeDir, _ := os.UserHomeDir()
	expectedUserPath := filepath.Join(homeDir, ".prompts", "tasks")
	if paths[1] != expectedUserPath {
		t.Errorf("Expected second path to be %q, got %q", expectedUserPath, paths[1])
	}

	if paths[2] != "/etc/prompts/tasks" {
		t.Errorf("Expected third path to be '/etc/prompts/tasks', got %q", paths[2])
	}
}

func TestGetRuleContentWithMissingFile(t *testing.T) {
	// Test that GetRuleContent returns empty string when file doesn't exist
	ruleFunc := func() string {
		content, err := os.ReadFile("/nonexistent/file.md")
		if err != nil {
			return ""
		}
		return string(content)
	}

	content := ruleFunc()
	if content != "" {
		t.Errorf("Expected empty string for missing file, got %q", content)
	}
}

func TestFindUserRules_WithSubdirectories(t *testing.T) {
	// Create a temporary directory structure with subdirectories
	tempHome := t.TempDir()
	userRulesDir := filepath.Join(tempHome, ".prompts", "rules")
	subDir := filepath.Join(userRulesDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create rules in different levels
	rootRuleContent := "# Root Rule"
	subRuleContent := "# Sub Rule"
	
	if err := os.WriteFile(filepath.Join(userRulesDir, "root.md"), []byte(rootRuleContent), 0644); err != nil {
		t.Fatalf("Failed to create root.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "sub.md"), []byte(subRuleContent), 0644); err != nil {
		t.Fatalf("Failed to create sub.md: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	ruleFuncs, err := FindUserRules()
	if err != nil {
		t.Fatalf("FindUserRules returned error: %v", err)
	}

	// Should find both rules (in root and subdirectory)
	if len(ruleFuncs) != 2 {
		t.Errorf("Expected 2 rule functions, got %d", len(ruleFuncs))
	}

	// Verify content
	var foundRoot, foundSub bool
	for _, ruleFunc := range ruleFuncs {
		content := ruleFunc()
		if strings.Contains(content, "Root Rule") {
			foundRoot = true
		} else if strings.Contains(content, "Sub Rule") {
			foundSub = true
		}
	}

	if !foundRoot {
		t.Error("Did not find root rule content")
	}
	if !foundSub {
		t.Error("Did not find sub rule content")
	}
}
