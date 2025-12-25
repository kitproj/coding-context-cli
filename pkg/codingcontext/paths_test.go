package codingcontext

import (
	"path/filepath"
	"testing"
)

func TestRulePaths(t *testing.T) {
	tests := []struct {
		name           string
		dir            string
		home           bool
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "non-home directory includes all agent paths",
			dir:  "/project",
			home: false,
			wantContains: []string{
				filepath.Join("/project", ".agents", "rules"),
				filepath.Join("/project", ".cursor", "rules"),
				filepath.Join("/project", ".cursorrules"),
				filepath.Join("/project", ".claude"),
			},
			wantNotContain: []string{},
		},
		{
			name: "home directory includes only home agents",
			dir:  "/home/user",
			home: true,
			wantContains: []string{
				filepath.Join("/home/user", ".agents", "rules"),
				filepath.Join("/home/user", ".claude"),
				filepath.Join("/home/user", ".codex"),
			},
			wantNotContain: []string{
				filepath.Join("/home/user", ".cursor", "rules"),
				filepath.Join("/home/user", ".windsurf", "rules"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := rulePaths(tt.dir, tt.home)

			// Check that expected paths are present
			for _, want := range tt.wantContains {
				found := false
				for _, path := range paths {
					if path == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected path %q not found in rulePaths", want)
				}
			}

			// Check that unwanted paths are not present
			for _, unwant := range tt.wantNotContain {
				for _, path := range paths {
					if path == unwant {
						t.Errorf("Unexpected path %q found in rulePaths", unwant)
					}
				}
			}
		})
	}
}

func TestTaskSearchPaths(t *testing.T) {
	dir := "/project"
	paths := taskSearchPaths(dir)

	// Should contain at least the .agents/tasks path
	expectedPath := filepath.Join(dir, ".agents", "tasks")
	found := false
	for _, path := range paths {
		if path == expectedPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected path %q not found in taskSearchPaths", expectedPath)
	}
}

func TestCommandSearchPaths(t *testing.T) {
	dir := "/project"
	paths := commandSearchPaths(dir)

	// Should contain at least the .agents/commands path
	expectedPaths := []string{
		filepath.Join(dir, ".agents", "commands"),
		filepath.Join(dir, ".cursor", "commands"),
		filepath.Join(dir, ".opencode", "command"),
	}

	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected path %q not found in commandSearchPaths", expected)
		}
	}
}

func TestSkillSearchPaths(t *testing.T) {
	dir := "/project"
	paths := skillSearchPaths(dir)

	// Should contain at least the .agents/skills path
	expectedPath := filepath.Join(dir, ".agents", "skills")
	found := false
	for _, path := range paths {
		if path == expectedPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected path %q not found in skillSearchPaths", expectedPath)
	}
}

func TestPathsUseAgentsPaths(t *testing.T) {
	// Verify that all path functions are using the agentsPaths configuration
	// by checking that they return paths for all configured agents

	dir := "/test"
	
	// Get paths from functions
	rulePaths := rulePaths(dir, false)
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
