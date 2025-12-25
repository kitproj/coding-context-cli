package codingcontext

// agentPathsConfig describes the search paths for a specific agent.
// This is the internal configuration structure used by the agentsPaths map.
type agentPathsConfig struct {
	rulesPaths   []string // Paths to search for rule files
	skillsPath   string   // Path to search for skill directories
	commandsPath string   // Path to search for command files
	tasksPath    string   // Path to search for task files
}

// agentsPaths maps each agent to its specific search paths.
// Empty string agent ("") represents the generic .agents directory structure.
// If a path is empty, it is not defined for that agent.
var agentsPaths = map[Agent]agentPathsConfig{
	// Generic .agents directory structure (empty agent name)
	Agent(""): {
		rulesPaths:   []string{".agents/rules"},
		skillsPath:   ".agents/skills",
		commandsPath: ".agents/commands",
		tasksPath:    ".agents/tasks",
	},
	// Cursor agent paths
	AgentCursor: {
		rulesPaths:   []string{".cursor/rules", ".cursorrules"},
		skillsPath:   ".cursor/skills",
		commandsPath: ".cursor/commands",
		// No tasks path defined for Cursor
	},
	// OpenCode agent paths
	AgentOpenCode: {
		rulesPaths:   []string{".opencode/agent", ".opencode/rules"},
		skillsPath:   ".opencode/skills",
		commandsPath: ".opencode/command",
		// No tasks path defined for OpenCode
	},
	// Copilot agent paths
	AgentCopilot: {
		rulesPaths: []string{".github/copilot-instructions.md", ".github/agents"},
		skillsPath: ".github/skills",
		// No commands or tasks paths defined for Copilot
	},
	// Claude agent paths
	AgentClaude: {
		rulesPaths: []string{".claude", "CLAUDE.md", "CLAUDE.local.md"},
		skillsPath: ".claude/skills",
		// No commands or tasks paths defined for Claude
	},
	// Gemini agent paths
	AgentGemini: {
		rulesPaths: []string{".gemini/styleguide.md", ".gemini", "GEMINI.md"},
		skillsPath: ".gemini/skills",
		// No commands or tasks paths defined for Gemini
	},
	// Augment agent paths
	AgentAugment: {
		rulesPaths: []string{".augment/rules", ".augment/guidelines.md"},
		skillsPath: ".augment/skills",
		// No commands or tasks paths defined for Augment
	},
	// Windsurf agent paths
	AgentWindsurf: {
		rulesPaths: []string{".windsurf/rules", ".windsurfrules"},
		skillsPath: ".windsurf/skills",
		// No commands or tasks paths defined for Windsurf
	},
	// Codex agent paths
	AgentCodex: {
		rulesPaths: []string{".codex", "AGENTS.md"},
		skillsPath: ".codex/skills",
		// No commands or tasks paths defined for Codex
	},
}
