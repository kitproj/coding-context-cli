package codingcontext

// AgentPaths describes the search paths for a specific agent
type AgentPaths struct {
	RulesPaths   []string // Paths to search for rule files
	SkillsPath   string   // Path to search for skill directories
	CommandsPath string   // Path to search for command files
	TasksPath    string   // Path to search for task files
}

// AgentsPaths maps each agent to its specific search paths.
// Empty string agent ("") represents the generic .agents directory structure.
// If a path is empty, it is not defined for that agent.
var AgentsPaths = map[Agent]AgentPaths{
	// Generic .agents directory structure (empty agent name)
	Agent(""): {
		RulesPaths:   []string{".agents/rules"},
		SkillsPath:   ".agents/skills",
		CommandsPath: ".agents/commands",
		TasksPath:    ".agents/tasks",
	},
	// Cursor agent paths
	AgentCursor: {
		RulesPaths:   []string{".cursor/rules", ".cursorrules"},
		CommandsPath: ".cursor/commands",
		// No skills or tasks paths defined for Cursor
	},
	// OpenCode agent paths
	AgentOpenCode: {
		RulesPaths:   []string{".opencode/agent", ".opencode/rules"},
		CommandsPath: ".opencode/command",
		// No skills or tasks paths defined for OpenCode
	},
	// Copilot agent paths
	AgentCopilot: {
		RulesPaths: []string{".github/copilot-instructions.md", ".github/agents"},
		// No skills, commands, or tasks paths defined for Copilot
	},
	// Claude agent paths
	AgentClaude: {
		RulesPaths: []string{".claude", "CLAUDE.md", "CLAUDE.local.md"},
		// No skills, commands, or tasks paths defined for Claude
	},
	// Gemini agent paths
	AgentGemini: {
		RulesPaths: []string{".gemini/styleguide.md", ".gemini", "GEMINI.md"},
		// No skills, commands, or tasks paths defined for Gemini
	},
	// Augment agent paths
	AgentAugment: {
		RulesPaths: []string{".augment/rules", ".augment/guidelines.md"},
		// No skills, commands, or tasks paths defined for Augment
	},
	// Windsurf agent paths
	AgentWindsurf: {
		RulesPaths: []string{".windsurf/rules", ".windsurfrules"},
		// No skills, commands, or tasks paths defined for Windsurf
	},
	// Codex agent paths
	AgentCodex: {
		RulesPaths: []string{".codex", "AGENTS.md"},
		// No skills, commands, or tasks paths defined for Codex
	},
}
