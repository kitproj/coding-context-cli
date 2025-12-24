package codingcontext

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createSkill creates a skill file (SKILL.md) in the .agents/skills/<skillName> directory
func createSkill(t *testing.T, dir, skillName, frontmatter, content string) {
	t.Helper()
	skillDir := filepath.Join(dir, ".agents", "skills", skillName)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill directory: %v", err)
	}

	var fileContent string
	if frontmatter != "" {
		fileContent = "---\n" + frontmatter + "\n---\n" + content
	} else {
		fileContent = content
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("failed to create skill file: %v", err)
	}
}

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name      string
		skillName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid simple name",
			skillName: "pdf-processing",
			wantErr:   false,
		},
		{
			name:      "valid name with numbers",
			skillName: "data-analysis-v2",
			wantErr:   false,
		},
		{
			name:      "valid single word",
			skillName: "testing",
			wantErr:   false,
		},
		{
			name:      "empty name",
			skillName: "",
			wantErr:   true,
			errMsg:    "must be 1-64 characters",
		},
		{
			name:      "too long name",
			skillName: "this-is-a-very-long-skill-name-that-exceeds-sixty-four-characters-limit",
			wantErr:   true,
			errMsg:    "must be 1-64 characters",
		},
		{
			name:      "starts with hyphen",
			skillName: "-invalid",
			wantErr:   true,
			errMsg:    "cannot start or end with a hyphen",
		},
		{
			name:      "ends with hyphen",
			skillName: "invalid-",
			wantErr:   true,
			errMsg:    "cannot start or end with a hyphen",
		},
		{
			name:      "consecutive hyphens",
			skillName: "invalid--name",
			wantErr:   true,
			errMsg:    "cannot contain consecutive hyphens",
		},
		{
			name:      "uppercase letters",
			skillName: "Invalid-Name",
			wantErr:   true,
			errMsg:    "can only contain lowercase letters, numbers, and hyphens",
		},
		{
			name:      "special characters",
			skillName: "invalid_name",
			wantErr:   true,
			errMsg:    "can only contain lowercase letters, numbers, and hyphens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSkillName(tt.skillName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSkillName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateSkillName() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateSkillDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid description",
			description: "This is a valid skill description.",
			wantErr:     false,
		},
		{
			name:        "empty description",
			description: "",
			wantErr:     true,
			errMsg:      "must be 1-1024 characters",
		},
		{
			name:        "too long description",
			description: strings.Repeat("a", 1025),
			wantErr:     true,
			errMsg:      "must be 1-1024 characters",
		},
		{
			name:        "max length description",
			description: strings.Repeat("a", 1024),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSkillDescription(tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSkillDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateSkillDescription() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestSkillDiscovery(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid skill
	createSkill(t, tmpDir, "pdf-processing",
		"name: pdf-processing\ndescription: Process PDF files with various operations",
		"# PDF Processing\n\nInstructions for processing PDFs.")

	// Create task
	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task content")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if len(result.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(result.Skills))
	}

	if len(result.Skills) > 0 {
		skill := result.Skills[0]
		if skill.FrontMatter.Name != "pdf-processing" {
			t.Errorf("expected skill name 'pdf-processing', got %q", skill.FrontMatter.Name)
		}
		if skill.FrontMatter.Description != "Process PDF files with various operations" {
			t.Errorf("expected specific description, got %q", skill.FrontMatter.Description)
		}
	}
}

func TestSkillNameMustMatchDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a skill where name doesn't match directory
	createSkill(t, tmpDir, "pdf-processing",
		"name: wrong-name\ndescription: Process PDF files",
		"# PDF Processing")

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	_, err := cc.Run(ctx, "test-task")
	if err == nil {
		t.Fatalf("expected error when skill name doesn't match directory, got nil")
	}

	if !strings.Contains(err.Error(), "must match parent directory name") {
		t.Errorf("expected error about directory name mismatch, got: %v", err)
	}
}

func TestSkillWithOver100Tokens(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a skill with > 100 tokens (400+ characters = 100+ tokens)
	// This should be allowed (no warning, no error)
	longContent := strings.Repeat("This is a sentence to make the content longer. ", 20)
	createSkill(t, tmpDir, "large-skill",
		"name: large-skill\ndescription: A skill with many tokens",
		longContent)

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if len(result.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(result.Skills))
	}

	if len(result.Skills) > 0 && result.Skills[0].Tokens <= 100 {
		t.Errorf("expected skill with > 100 tokens, got %d", result.Skills[0].Tokens)
	}
}

func TestSkillWithOver2000Tokens(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a skill with many tokens (8000+ characters = 2000+ tokens)
	// This should be allowed - no token limit errors
	longContent := strings.Repeat("This is a sentence to make the content much longer. ", 200)
	createSkill(t, tmpDir, "huge-skill",
		"name: huge-skill\ndescription: A skill with many tokens but valid name and description",
		longContent)

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() should not fail for large skills, got error: %v", err)
	}

	if len(result.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(result.Skills))
	}

	if len(result.Skills) > 0 && result.Skills[0].Tokens <= 2000 {
		t.Errorf("expected skill with > 2000 tokens, got %d", result.Skills[0].Tokens)
	}
}

func TestSkillWithOptionalFields(t *testing.T) {
	tmpDir := t.TempDir()

	createSkill(t, tmpDir, "advanced-skill",
		`name: advanced-skill
description: An advanced skill with all fields
license: MIT
compatibility: Requires Python 3.8+
metadata:
  author: test-author
  version: "1.0"
allowed-tools: bash git docker`,
		"# Advanced Skill\n\nDetailed instructions.")

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if len(result.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Skills))
	}

	skill := result.Skills[0]
	if skill.FrontMatter.License != "MIT" {
		t.Errorf("expected license 'MIT', got %q", skill.FrontMatter.License)
	}
	if skill.FrontMatter.Compatibility != "Requires Python 3.8+" {
		t.Errorf("expected specific compatibility, got %q", skill.FrontMatter.Compatibility)
	}
	if skill.FrontMatter.AllowedTools != "bash git docker" {
		t.Errorf("expected specific allowed-tools, got %q", skill.FrontMatter.AllowedTools)
	}
	if len(skill.FrontMatter.Metadata) != 2 {
		t.Errorf("expected 2 metadata entries, got %d", len(skill.FrontMatter.Metadata))
	}
	if skill.FrontMatter.Metadata["author"] != "test-author" {
		t.Errorf("expected metadata author 'test-author', got %q", skill.FrontMatter.Metadata["author"])
	}
}

func TestSkillCompatibilityTooLong(t *testing.T) {
	tmpDir := t.TempDir()

	// Create skill with compatibility > 500 characters
	longCompat := strings.Repeat("a", 501)
	createSkill(t, tmpDir, "invalid-skill",
		"name: invalid-skill\ndescription: A skill with long compatibility\ncompatibility: "+longCompat,
		"Content")

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	_, err := cc.Run(ctx, "test-task")
	if err == nil {
		t.Fatalf("expected error for compatibility over 500 chars, got nil")
	}

	if !strings.Contains(err.Error(), "compatibility field") && !strings.Contains(err.Error(), "exceeds 500 characters") {
		t.Errorf("expected error about compatibility length, got: %v", err)
	}
}

func TestSkillsInResumeMode(t *testing.T) {
	tmpDir := t.TempDir()

	createSkill(t, tmpDir, "test-skill",
		"name: test-skill\ndescription: A test skill",
		"Content")

	createTask(t, tmpDir, "test-task", "task_name: test-task\nresume: true", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://"+tmpDir),
		WithResume(true),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// In resume mode, skills should be skipped (like rules)
	if len(result.Skills) != 0 {
		t.Errorf("expected 0 skills in resume mode, got %d", len(result.Skills))
	}
}

func TestMultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()

	createSkill(t, tmpDir, "skill-one",
		"name: skill-one\ndescription: First skill",
		"First skill content")

	createSkill(t, tmpDir, "skill-two",
		"name: skill-two\ndescription: Second skill",
		"Second skill content")

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if len(result.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(result.Skills))
	}

	// Verify both skills are present
	skillNames := make(map[string]bool)
	for _, skill := range result.Skills {
		skillNames[skill.FrontMatter.Name] = true
	}

	if !skillNames["skill-one"] || !skillNames["skill-two"] {
		t.Errorf("expected both skill-one and skill-two, got: %v", skillNames)
	}
}

func TestSkillSelectorsFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create skill with language selector
	createSkill(t, tmpDir, "go-skill",
		"name: go-skill\ndescription: Go language skill\nlanguage: go",
		"Go skill content")

	// Create skill with different language
	createSkill(t, tmpDir, "python-skill",
		"name: python-skill\ndescription: Python language skill\nlanguage: python",
		"Python skill content")

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()

	// Test with go selector
	includes := make(map[string]map[string]bool)
	includes["language"] = map[string]bool{"go": true}
	cc := New(
		WithSearchPaths("file://"+tmpDir),
		WithSelectors(includes),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if len(result.Skills) != 1 {
		t.Errorf("expected 1 skill with go selector, got %d", len(result.Skills))
	}

	if len(result.Skills) > 0 && result.Skills[0].FrontMatter.Name != "go-skill" {
		t.Errorf("expected go-skill, got %s", result.Skills[0].FrontMatter.Name)
	}
}

func TestSkillOnlyInSKILLmdFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-SKILL.md file in skills directory
	skillDir := filepath.Join(tmpDir, ".agents", "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill directory: %v", err)
	}

	wrongName := filepath.Join(skillDir, "README.md")
	content := "---\nname: test-skill\ndescription: Test\n---\nContent"
	if err := os.WriteFile(wrongName, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task")

	ctx := context.Background()
	cc := New(
		WithSearchPaths("file://" + tmpDir),
	)

	result, err := cc.Run(ctx, "test-task")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Should not find any skills since the file is not named SKILL.md
	if len(result.Skills) != 0 {
		t.Errorf("expected 0 skills (file not named SKILL.md), got %d", len(result.Skills))
	}
}
