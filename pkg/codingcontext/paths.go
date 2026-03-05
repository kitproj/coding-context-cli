package codingcontext

import "path/filepath"

// namespacedTaskSearchPaths returns task search paths for the given namespace.
// Namespace task dir is searched first; global task dirs follow as fallback.
func namespacedTaskSearchPaths(dir, namespace string) []string {
	var paths []string
	if namespace != "" {
		paths = append(paths, filepath.Join(dir, ".agents/namespaces", namespace, "tasks"))
	}

	paths = append(paths, taskSearchPaths(dir)...)

	return paths
}

// namespacedRuleSearchPaths returns rule search paths for the given namespace.
// Namespace rule dir is prepended so namespace rules appear first in context output.
// Global rule dirs always follow (both are included).
func namespacedRuleSearchPaths(dir, namespace string) []string {
	var paths []string
	if namespace != "" {
		paths = append(paths, filepath.Join(dir, ".agents/namespaces", namespace, "rules"))
	}

	paths = append(paths, rulePaths(dir)...)

	return paths
}

// namespacedCommandSearchPaths returns command search paths for the given namespace.
// Namespace command dir is searched first; the first match wins (namespace overrides global).
func namespacedCommandSearchPaths(dir, namespace string) []string {
	var paths []string
	if namespace != "" {
		paths = append(paths, filepath.Join(dir, ".agents/namespaces", namespace, "commands"))
	}

	paths = append(paths, commandSearchPaths(dir)...)

	return paths
}

// namespacedSkillSearchPaths returns skill search paths for the given namespace.
// Namespace skill dir is listed first so namespace skills appear earlier in discovery.
func namespacedSkillSearchPaths(dir, namespace string) []string {
	var paths []string
	if namespace != "" {
		paths = append(paths, filepath.Join(dir, ".agents/namespaces", namespace, "skills"))
	}

	paths = append(paths, skillSearchPaths(dir)...)

	return paths
}

// rulePaths returns the search paths for rule files in a directory.
// It collects rule paths from all agents in the agents paths configuration.
func rulePaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range getAgentsPaths() {
		// Add each rule path for this agent
		for _, rulePath := range config.rulesPaths {
			paths = append(paths, filepath.Join(dir, rulePath))
		}
	}

	return paths
}

// taskSearchPaths returns the search paths for task files in a directory.
// It collects task paths from all agents in the agents paths configuration.
func taskSearchPaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range getAgentsPaths() {
		if config.tasksPath != "" {
			paths = append(paths, filepath.Join(dir, config.tasksPath))
		}
	}

	return paths
}

// commandSearchPaths returns the search paths for command files in a directory.
// It collects command paths from all agents in the agents paths configuration.
func commandSearchPaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range getAgentsPaths() {
		if config.commandsPath != "" {
			paths = append(paths, filepath.Join(dir, config.commandsPath))
		}
	}

	return paths
}

// skillSearchPaths returns the search paths for skill directories in a directory.
// It collects skill paths from all agents in the agents paths configuration.
func skillSearchPaths(dir string) []string {
	var paths []string

	// Iterate through all configured agents
	for _, config := range getAgentsPaths() {
		if config.skillsPath != "" {
			paths = append(paths, filepath.Join(dir, config.skillsPath))
		}
	}

	return paths
}
