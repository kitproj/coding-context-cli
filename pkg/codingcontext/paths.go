package codingcontext

import "path/filepath"

// SearchPath represents a single search location with its associated subpaths
type SearchPath struct {
	BasePath      string
	RulesSubPaths []string
	TaskSubPaths  []string
}

// TaskSearchDirs returns the full paths for task search directories
// by joining BasePath with each TaskSubPath
func (sp SearchPath) TaskSearchDirs() []string {
	dirs := make([]string, 0, len(sp.TaskSubPaths))
	for _, subPath := range sp.TaskSubPaths {
		dirs = append(dirs, filepath.Join(sp.BasePath, subPath))
	}
	return dirs
}

// RulesSearchDirs returns the full paths for rule search directories
// by joining BasePath with each RulesSubPath
func (sp SearchPath) RulesSearchDirs() []string {
	dirs := make([]string, 0, len(sp.RulesSubPaths))
	for _, subPath := range sp.RulesSubPaths {
		dirs = append(dirs, filepath.Join(sp.BasePath, subPath))
	}
	return dirs
}

// DefaultSearchPaths returns the search paths for default local paths (baseDir and homeDir)
func DefaultSearchPaths(baseDir, homeDir string) []SearchPath {
	return []SearchPath{
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
		// homeDir search paths
		{
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
		},
	}
}

// NewSearchPathWithDefaults creates a SearchPath with default subpaths for a given base path
// (uses the same default subpaths as PathSearchPaths)
func NewSearchPathWithDefaults(basePath string) SearchPath {
	return SearchPath{
		BasePath: basePath,
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
	}
}

// PathSearchPaths returns the search paths for a given directory path
// (used for both local and remote paths after download)
func PathSearchPaths(dir string) []SearchPath {
	return []SearchPath{
		NewSearchPathWithDefaults(dir),
	}
}
