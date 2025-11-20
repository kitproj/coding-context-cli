package codingcontext

import "path/filepath"

// AllTaskSearchPaths returns the standard search paths for task files
// baseDir is the working directory to resolve relative paths from
func AllTaskSearchPaths(baseDir, homeDir string) []string {
	return []string{
		filepath.Join(baseDir, ".agents", "tasks"),
		filepath.Join(baseDir, ".cursor", "commands"),
		filepath.Join(baseDir, ".opencode", "command"),
		filepath.Join(homeDir, ".agents", "tasks"),
	}
}

// AllRulePaths returns the standard search paths for rule files
// baseDir is the working directory to resolve relative paths from
func AllRulePaths(baseDir, homeDir string) []string {
	return []string{
		filepath.Join(baseDir, "CLAUDE.local.md"),

		filepath.Join(baseDir, ".agents", "rules"),
		filepath.Join(baseDir, ".cursor", "rules"),
		filepath.Join(baseDir, ".augment", "rules"),
		filepath.Join(baseDir, ".windsurf", "rules"),
		filepath.Join(baseDir, ".opencode", "agent"),

		filepath.Join(baseDir, ".github", "copilot-instructions.md"),
		filepath.Join(baseDir, ".gemini", "styleguide.md"),
		filepath.Join(baseDir, ".github", "agents"),
		filepath.Join(baseDir, ".augment", "guidelines.md"),

		filepath.Join(baseDir, "AGENTS.md"),
		filepath.Join(baseDir, "CLAUDE.md"),
		filepath.Join(baseDir, "GEMINI.md"),

		filepath.Join(baseDir, ".cursorrules"),
		filepath.Join(baseDir, ".windsurfrules"),

		// ancestors
		filepath.Join(baseDir, "..", "AGENTS.md"),
		filepath.Join(baseDir, "..", "CLAUDE.md"),
		filepath.Join(baseDir, "..", "GEMINI.md"),

		filepath.Join(baseDir, "..", "..", "AGENTS.md"),
		filepath.Join(baseDir, "..", "..", "CLAUDE.md"),
		filepath.Join(baseDir, "..", "..", "GEMINI.md"),

		// user
		filepath.Join(homeDir, ".agents", "rules"),
		filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		filepath.Join(homeDir, ".codex", "AGENTS.md"),
		filepath.Join(homeDir, ".gemini", "GEMINI.md"),
		filepath.Join(homeDir, ".opencode", "rules"),
	}
}

// DownloadedRulePaths returns the search paths for rule files in downloaded directories
func DownloadedRulePaths(dir string) []string {
	return []string{
		filepath.Join(dir, ".agents", "rules"),
		filepath.Join(dir, ".cursor", "rules"),
		filepath.Join(dir, ".augment", "rules"),
		filepath.Join(dir, ".windsurf", "rules"),
		filepath.Join(dir, ".opencode", "agent"),
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

// DownloadedTaskSearchPaths returns the search paths for task files in downloaded directories
func DownloadedTaskSearchPaths(dir string) []string {
	return []string{
		filepath.Join(dir, ".agents", "tasks"),
		filepath.Join(dir, ".cursor", "commands"),
		filepath.Join(dir, ".opencode", "command"),
	}
}
