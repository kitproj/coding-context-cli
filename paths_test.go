package main

import (
	"path/filepath"
	"testing"
)

// TestAllTaskSearchPaths verifies that all expected task search paths are included
func TestAllTaskSearchPaths(t *testing.T) {
	homeDir := "/home/testuser"
	paths := allTaskSearchPaths(homeDir)

	expectedPaths := []string{
		filepath.Join(".agents", "tasks"),
		filepath.Join(".cursor", "commands"),
		filepath.Join(".opencode", "command"),
		filepath.Join(homeDir, ".agents", "tasks"),
		filepath.Join(homeDir, ".cursor", "commands"),
		filepath.Join(homeDir, ".opencode", "command"),
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
	}

	// Convert to map for easier checking
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	for _, expected := range expectedPaths {
		if !pathMap[expected] {
			t.Errorf("Missing expected path: %s", expected)
		}
	}
}

// TestAllRulePaths verifies that all expected rule paths are included
func TestAllRulePaths(t *testing.T) {
	homeDir := "/home/testuser"
	paths := allRulePaths(homeDir)

	// Paths that must be present
	requiredPaths := []string{
		// Local project paths
		".agents/rules",
		".cursor/rules",
		".augment/rules",
		".windsurf/rules",
		".opencode/agent",
		".opencode/command",
		".opencode/rules",

		// User-level paths
		filepath.Join(homeDir, ".agents", "rules"),
		filepath.Join(homeDir, ".cursor", "rules"),
		filepath.Join(homeDir, ".augment", "rules"),
		filepath.Join(homeDir, ".windsurf", "rules"),
		filepath.Join(homeDir, ".opencode", "agent"),
		filepath.Join(homeDir, ".opencode", "command"),
		filepath.Join(homeDir, ".opencode", "rules"),
	}

	// Convert to map for easier checking
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	for _, required := range requiredPaths {
		if !pathMap[required] {
			t.Errorf("Missing required path: %s", required)
		}
	}
}

// TestDownloadedRulePaths verifies that all expected downloaded rule paths are included
func TestDownloadedRulePaths(t *testing.T) {
	dir := "/tmp/downloaded"
	paths := downloadedRulePaths(dir)

	requiredPaths := []string{
		filepath.Join(dir, ".agents", "rules"),
		filepath.Join(dir, ".cursor", "rules"),
		filepath.Join(dir, ".augment", "rules"),
		filepath.Join(dir, ".windsurf", "rules"),
		filepath.Join(dir, ".opencode", "agent"),
		filepath.Join(dir, ".opencode", "command"),
		filepath.Join(dir, ".opencode", "rules"),
	}

	// Convert to map for easier checking
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	for _, required := range requiredPaths {
		if !pathMap[required] {
			t.Errorf("Missing required path: %s", required)
		}
	}
}

// TestDownloadedTaskSearchPaths verifies that all expected downloaded task paths are included
func TestDownloadedTaskSearchPaths(t *testing.T) {
	dir := "/tmp/downloaded"
	paths := downloadedTaskSearchPaths(dir)

	expectedPaths := []string{
		filepath.Join(dir, ".agents", "tasks"),
		filepath.Join(dir, ".cursor", "commands"),
		filepath.Join(dir, ".opencode", "command"),
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
	}

	// Convert to map for easier checking
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	for _, expected := range expectedPaths {
		if !pathMap[expected] {
			t.Errorf("Missing expected path: %s", expected)
		}
	}
}

// TestOpenCodePathsPresent verifies all OpenCode.ai paths are included
func TestOpenCodePathsPresent(t *testing.T) {
	homeDir := "/home/testuser"

	t.Run("local_opencode_paths", func(t *testing.T) {
		paths := allRulePaths(homeDir)
		pathMap := make(map[string]bool)
		for _, p := range paths {
			pathMap[p] = true
		}

		opencodePaths := []string{
			".opencode/agent",
			".opencode/command",
			".opencode/rules",
		}

		for _, path := range opencodePaths {
			if !pathMap[path] {
				t.Errorf("Missing OpenCode.ai local path: %s", path)
			}
		}
	})

	t.Run("user_opencode_paths", func(t *testing.T) {
		paths := allRulePaths(homeDir)
		pathMap := make(map[string]bool)
		for _, p := range paths {
			pathMap[p] = true
		}

		opencodePaths := []string{
			filepath.Join(homeDir, ".opencode", "agent"),
			filepath.Join(homeDir, ".opencode", "command"),
			filepath.Join(homeDir, ".opencode", "rules"),
		}

		for _, path := range opencodePaths {
			if !pathMap[path] {
				t.Errorf("Missing OpenCode.ai user path: %s", path)
			}
		}
	})

	t.Run("downloaded_opencode_paths", func(t *testing.T) {
		dir := "/tmp/test"
		paths := downloadedRulePaths(dir)
		pathMap := make(map[string]bool)
		for _, p := range paths {
			pathMap[p] = true
		}

		opencodePaths := []string{
			filepath.Join(dir, ".opencode", "agent"),
			filepath.Join(dir, ".opencode", "command"),
			filepath.Join(dir, ".opencode", "rules"),
		}

		for _, path := range opencodePaths {
			if !pathMap[path] {
				t.Errorf("Missing OpenCode.ai downloaded path: %s", path)
			}
		}
	})

	t.Run("opencode_task_paths", func(t *testing.T) {
		paths := allTaskSearchPaths(homeDir)
		pathMap := make(map[string]bool)
		for _, p := range paths {
			pathMap[p] = true
		}

		opencodePaths := []string{
			filepath.Join(".opencode", "command"),
			filepath.Join(homeDir, ".opencode", "command"),
		}

		for _, path := range opencodePaths {
			if !pathMap[path] {
				t.Errorf("Missing OpenCode.ai task path: %s", path)
			}
		}
	})
}

// TestUserLevelPathsPresent verifies all user-level paths for various tools are included
func TestUserLevelPathsPresent(t *testing.T) {
	homeDir := "/home/testuser"
	paths := allRulePaths(homeDir)
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	userPaths := []string{
		filepath.Join(homeDir, ".cursor", "rules"),
		filepath.Join(homeDir, ".augment", "rules"),
		filepath.Join(homeDir, ".windsurf", "rules"),
	}

	for _, path := range userPaths {
		if !pathMap[path] {
			t.Errorf("Missing user-level path: %s", path)
		}
	}
}

// TestUserLevelTaskPathsPresent verifies user-level task paths are included
func TestUserLevelTaskPathsPresent(t *testing.T) {
	homeDir := "/home/testuser"
	paths := allTaskSearchPaths(homeDir)
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	userTaskPaths := []string{
		filepath.Join(homeDir, ".cursor", "commands"),
		filepath.Join(homeDir, ".opencode", "command"),
	}

	for _, path := range userTaskPaths {
		if !pathMap[path] {
			t.Errorf("Missing user-level task path: %s", path)
		}
	}
}
