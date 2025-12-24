package codingcontext

import "path/filepath"

// DownloadedRulePaths returns the search paths for rule files in downloaded directories
func rulePaths(dir string, home bool) []string {
	if home {
		return []string{
			// user
			filepath.Join(dir, ".agents", "rules"),
			filepath.Join(dir, ".claude", "CLAUDE.md"),
			filepath.Join(dir, ".codex", "AGENTS.md"),
			filepath.Join(dir, ".gemini", "GEMINI.md"),
			filepath.Join(dir, ".opencode", "rules"),
		}
	}
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
		filepath.Join(dir, "CLAUDE.local.md"),
		filepath.Join(dir, "GEMINI.md"),
		filepath.Join(dir, ".cursorrules"),
		filepath.Join(dir, ".windsurfrules"),
	}
}

// taskSearchPaths returns the search paths for task files in a directory
func taskSearchPaths(dir string) []string {
	return []string{
		filepath.Join(dir, ".agents", "tasks"),
	}
}

// commandSearchPaths returns the search paths for command files in a directory
func commandSearchPaths(dir string) []string {
	return []string{
		filepath.Join(dir, ".agents", "commands"),
		filepath.Join(dir, ".cursor", "commands"),
		filepath.Join(dir, ".opencode", "command"),
	}
}

// skillSearchPaths returns the search paths for skill directories in a directory
func skillSearchPaths(dir string) []string {
	return []string{
		filepath.Join(dir, ".agents", "skills"),
	}
}
