package codingcontext

// agentPathsConfig describes the search paths for a specific agent
type agentPathsConfig struct {
	RulesPaths   []string // Paths to search for rule files
	SkillsPath   string   // Path to search for skill directories
	CommandsPath string   // Path to search for command files
	TasksPath    string   // Path to search for task files
}

// AgentsPaths provides access to agent-specific search paths
type AgentsPaths struct {
	agent Agent
}

// RulesPaths returns the rules paths for the agent
func (ap AgentsPaths) RulesPaths() []string {
	if paths, exists := agentsPaths[ap.agent]; exists {
		return paths.RulesPaths
	}
	return nil
}

// SkillsPath returns the skills path for the agent
func (ap AgentsPaths) SkillsPath() string {
	if paths, exists := agentsPaths[ap.agent]; exists {
		return paths.SkillsPath
	}
	return ""
}

// CommandsPath returns the commands path for the agent
func (ap AgentsPaths) CommandsPath() string {
	if paths, exists := agentsPaths[ap.agent]; exists {
		return paths.CommandsPath
	}
	return ""
}

// TasksPath returns the tasks path for the agent
func (ap AgentsPaths) TasksPath() string {
	if paths, exists := agentsPaths[ap.agent]; exists {
		return paths.TasksPath
	}
	return ""
}

// Paths returns an AgentsPaths instance for accessing the agent's paths
func (a Agent) Paths() AgentsPaths {
	return AgentsPaths{agent: a}
}

// agentsPaths maps each agent to its specific search paths.
// Empty string agent ("") represents the generic .agents directory structure.
// If a path is empty, it is not defined for that agent.
var agentsPaths = map[Agent]agentPathsConfig{
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
		SkillsPath:   ".cursor/skills",
		CommandsPath: ".cursor/commands",
		// No tasks path defined for Cursor
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
