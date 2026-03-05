package codingcontext

import (
	"path/filepath"
	"slices"
	"testing"
)

const (
	testProjectDir = "/project"
	testNamespace  = "myteam"
)

func TestRulePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dir          string
		wantContains []string
	}{
		{
			name: "directory includes all agent paths",
			dir:  testProjectDir,
			wantContains: []string{
				filepath.Join(testProjectDir, ".agents", "rules"),
				filepath.Join(testProjectDir, ".cursor", "rules"),
				filepath.Join(testProjectDir, ".cursorrules"),
				filepath.Join(testProjectDir, ".claude"),
				filepath.Join(testProjectDir, ".codex"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			paths := rulePaths(tt.dir)

			// Check that expected paths are present
			for _, want := range tt.wantContains {
				if !slices.Contains(paths, want) {
					t.Errorf("Expected path %q not found in rulePaths", want)
				}
			}
		})
	}
}

func TestTaskSearchPaths(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	paths := taskSearchPaths(dir)

	// Should contain at least the .agents/tasks path
	expectedPath := filepath.Join(dir, ".agents", "tasks")
	if !slices.Contains(paths, expectedPath) {
		t.Errorf("Expected path %q not found in taskSearchPaths", expectedPath)
	}
}

func TestCommandSearchPaths(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	paths := commandSearchPaths(dir)

	// Should contain at least the .agents/commands path
	expectedPaths := []string{
		filepath.Join(dir, ".agents", "commands"),
		filepath.Join(dir, ".cursor", "commands"),
		filepath.Join(dir, ".opencode", "command"),
	}

	for _, expected := range expectedPaths {
		if !slices.Contains(paths, expected) {
			t.Errorf("Expected path %q not found in commandSearchPaths", expected)
		}
	}
}

func TestSkillSearchPaths(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	paths := skillSearchPaths(dir)

	// Should contain at least the .agents/skills path
	expectedPath := filepath.Join(dir, ".agents", "skills")
	if !slices.Contains(paths, expectedPath) {
		t.Errorf("Expected path %q not found in skillSearchPaths", expectedPath)
	}
}

func TestPathsUseAgentsPaths(t *testing.T) {
	t.Parallel()
	// Verify that all path functions are using the agentsPaths configuration
	// by checking that they return paths for all configured agents
	dir := "/test"

	// Get paths from functions
	rulePaths := rulePaths(dir)
	taskPaths := taskSearchPaths(dir)
	commandPaths := commandSearchPaths(dir)
	skillPaths := skillSearchPaths(dir)

	// Verify rulePaths contains paths from multiple agents
	if len(rulePaths) < 5 {
		t.Errorf("rulePaths should contain paths from multiple agents, got %d paths", len(rulePaths))
	}

	// Verify taskPaths is not empty
	if len(taskPaths) == 0 {
		t.Error("taskSearchPaths should return at least one path")
	}

	// Verify commandPaths is not empty
	if len(commandPaths) == 0 {
		t.Error("commandSearchPaths should return at least one path")
	}

	// Verify skillPaths is not empty
	if len(skillPaths) == 0 {
		t.Error("skillSearchPaths should return at least one path")
	}
}
