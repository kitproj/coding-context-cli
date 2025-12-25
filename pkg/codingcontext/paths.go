package codingcontext

import "path/filepath"

// rulePaths returns the search paths for rule files in a directory.
// It collects rule paths from all agents in the agentsPaths configuration.
// If home is true, only returns paths for agents that are user-level (home directory).
func rulePaths(dir string, home bool) []string {
	var paths []string

	// Define which agents should be included for home directory
	homeAgents := map[Agent]bool{
		Agent(""):     true, // generic .agents
		AgentClaude:   true,
		AgentCodex:    true,
		AgentGemini:   true,
		AgentOpenCode: true,
	}

	// Iterate through all configured agents
	for agent, config := range agentsPaths {
		// Skip non-home agents if we're in home directory mode
		if home && !homeAgents[agent] {
			continue
		}

		// Add each rule path for this agent
		for _, rulePath := range config.RulesPaths {
			paths = append(paths, filepath.Join(dir, rulePath))
		}
	}

	return paths
}

// taskSearchPaths returns the search paths for task files in a directory.
// It collects task paths from all agents in the agentsPaths configuration.
func taskSearchPaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range agentsPaths {
		if config.TasksPath != "" {
			paths = append(paths, filepath.Join(dir, config.TasksPath))
		}
	}

	return paths
}

// commandSearchPaths returns the search paths for command files in a directory.
// It collects command paths from all agents in the agentsPaths configuration.
func commandSearchPaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range agentsPaths {
		if config.CommandsPath != "" {
			paths = append(paths, filepath.Join(dir, config.CommandsPath))
		}
	}

	return paths
}

// skillSearchPaths returns the search paths for skill directories in a directory.
// It collects skill paths from all agents in the agentsPaths configuration.
func skillSearchPaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range agentsPaths {
		if config.SkillsPath != "" {
			paths = append(paths, filepath.Join(dir, config.SkillsPath))
		}
	}

	return paths
}
