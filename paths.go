package main

import "path/filepath"

func allTaskSearchPaths(homeDir string) []string {
	return []string{
		filepath.Join(".agents", "tasks"),
		filepath.Join(".cursor", "commands"),
		filepath.Join(".opencode", "command"),
		filepath.Join(homeDir, ".agents", "tasks"),
		filepath.Join(homeDir, ".cursor", "commands"),
		filepath.Join(homeDir, ".opencode", "command"),
	}
}

func allRulePaths(homeDir string) []string {
	return []string{
		"CLAUDE.local.md",

		".agents/rules",
		".cursor/rules",
		".augment/rules",
		".windsurf/rules",
		".opencode/agent",
		".opencode/command",
		".opencode/rules",

		".github/copilot-instructions.md",
		".gemini/styleguide.md",
		".github/agents",
		".augment/guidelines.md",

		"AGENTS.md",
		"CLAUDE.md",
		"GEMINI.md",

		".cursorrules",
		".windsurfrules",

		// ancestors
		"../AGENTS.md",
		"../CLAUDE.md",
		"../GEMINI.md",

		"../../AGENTS.md",
		"../../CLAUDE.md",
		"../../GEMINI.md",

		// user
		filepath.Join(homeDir, ".agents", "rules"),
		filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		filepath.Join(homeDir, ".cursor", "rules"),
		filepath.Join(homeDir, ".augment", "rules"),
		filepath.Join(homeDir, ".windsurf", "rules"),
		filepath.Join(homeDir, ".opencode", "agent"),
		filepath.Join(homeDir, ".opencode", "command"),
		filepath.Join(homeDir, ".opencode", "rules"),
		filepath.Join(homeDir, ".codex", "AGENTS.md"),
		filepath.Join(homeDir, ".gemini", "GEMINI.md"),
	}
}

func downloadedRulePaths(dir string) []string {
	return []string{
		filepath.Join(dir, ".agents", "rules"),
		filepath.Join(dir, ".cursor", "rules"),
		filepath.Join(dir, ".augment", "rules"),
		filepath.Join(dir, ".windsurf", "rules"),
		filepath.Join(dir, ".opencode", "agent"),
		filepath.Join(dir, ".opencode", "command"),
		filepath.Join(dir, ".opencode", "rules"),
		filepath.Join(dir, ".github", "copilot-instructions.md"),
		filepath.Join(dir, ".gemini", "styleguide.md"),
		filepath.Join(dir, ".github", "agents"),
		filepath.Join(dir, ".augment", "guidelines.md"),
		filepath.Join(dir, "AGENTS.md"),
		filepath.Join(dir, "CLAUDE.md"),
		filepath.Join(dir, "GEMINI.md"),
		filepath.Join(dir, ".cursorrules"),
		filepath.Join(dir, ".windsurfrules"),
	}
}

func downloadedTaskSearchPaths(dir string) []string {
	return []string{
		filepath.Join(dir, ".agents", "tasks"),
		filepath.Join(dir, ".cursor", "commands"),
		filepath.Join(dir, ".opencode", "command"),
	}
}
