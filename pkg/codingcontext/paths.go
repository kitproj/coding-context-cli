package codingcontext

import "path/filepath"

// rulePaths returns the search paths for rule files in a directory.
// It collects rule paths from all agents in the agentsPaths configuration.
func rulePaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range agentsPaths {
		// Add each rule path for this agent
		for _, rulePath := range config.rulesPaths {
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
		if config.tasksPath != "" {
			paths = append(paths, filepath.Join(dir, config.tasksPath))
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
		if config.commandsPath != "" {
			paths = append(paths, filepath.Join(dir, config.commandsPath))
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
		if config.skillsPath != "" {
			paths = append(paths, filepath.Join(dir, config.skillsPath))
		}
	}

	return paths
}
