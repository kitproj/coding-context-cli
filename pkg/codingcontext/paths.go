package codingcontext

// SearchPath represents a single search location with its associated subpaths
type SearchPath struct {
	BasePath      string
	RulesSubPaths []string
	TaskSubPaths  []string
}

// DefaultSearchPaths returns the search paths for default local paths (baseDir and homeDir)
func DefaultSearchPaths(baseDir, homeDir string) []SearchPath {
	searchPaths := []SearchPath{
		// baseDir search paths
		{
			BasePath: baseDir,
			RulesSubPaths: []string{
				"CLAUDE.local.md",
				".agents/rules",
				".cursor/rules",
				".augment/rules",
				".windsurf/rules",
				".opencode/agent",
				".github/copilot-instructions.md",
				".gemini/styleguide.md",
				".github/agents",
				".augment/guidelines.md",
				"AGENTS.md",
				"CLAUDE.md",
				"GEMINI.md",
				".cursorrules",
				".windsurfrules",
				"../AGENTS.md",
				"../CLAUDE.md",
				"../GEMINI.md",
				"../../AGENTS.md",
				"../../CLAUDE.md",
				"../../GEMINI.md",
			},
			TaskSubPaths: []string{
				".agents/tasks",
				".cursor/commands",
				".opencode/command",
			},
		},
	}

	// Only add homeDir search paths if homeDir is not empty
	if homeDir != "" {
		searchPaths = append(searchPaths, SearchPath{
			BasePath: homeDir,
			RulesSubPaths: []string{
				".agents/rules",
				".claude/CLAUDE.md",
				".codex/AGENTS.md",
				".gemini/GEMINI.md",
				".opencode/rules",
			},
			TaskSubPaths: []string{
				".agents/tasks",
			},
		})
	}

	return searchPaths
}

// PathSearchPaths returns the search paths for a given directory path
// (used for both local and remote paths after download)
func PathSearchPaths(dir string) []SearchPath {
	return []SearchPath{
		{
			BasePath: dir,
			RulesSubPaths: []string{
				".agents/rules",
				".cursor/rules",
				".augment/rules",
				".windsurf/rules",
				".opencode/agent",
				".github/copilot-instructions.md",
				".gemini/styleguide.md",
				".github/agents",
				".augment/guidelines.md",
				"AGENTS.md",
				"CLAUDE.md",
				"GEMINI.md",
				".cursorrules",
				".windsurfrules",
			},
			TaskSubPaths: []string{
				".agents/tasks",
				".cursor/commands",
				".opencode/command",
			},
		},
	}
}
