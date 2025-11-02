package main

import (
	"os"
	"path/filepath"
)

// a map from agents to their rule paths by level
var agentRules map[Agent]map[RuleLevel][]string

// initAgentRules initializes the agent rules based on current working directory
func initAgentRules() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	agentRules = make(map[Agent]map[RuleLevel][]string)

	// Default agent - for .prompts/rules directories
	agentRules[Default] = map[RuleLevel][]string{
		ProjectLevel: {
			".prompts/rules",
		},
		UserLevel: {
			filepath.Join(homeDir, ".config", "prompts", "rules"),
		},
		SystemLevel: {
			"/var/local/prompts/rules",
		},
	}

	// Claude - Hierarchical Concatenation
	agentRules[Claude] = map[RuleLevel][]string{
		ProjectLevel: {
			"CLAUDE.local.md",
		},
		AncestorLevel: {
			"CLAUDE.md",
		},
		UserLevel: {
			filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		},
	}

	// Gemini CLI - Hierarchical Concatenation + Simple System Prompt
	agentRules[Gemini] = map[RuleLevel][]string{
		ProjectLevel: {
			".gemini/styleguide.md",
		},
		AncestorLevel: {
			"GEMINI.md",
		},
		UserLevel: {
			filepath.Join(homeDir, ".gemini", "GEMINI.md"),
		},
	}

	// Codex CLI - Hierarchical Concatenation
	agentRules[Codex] = map[RuleLevel][]string{
		AncestorLevel: {
			"AGENTS.md",
		},
		UserLevel: {
			filepath.Join(homeDir, ".codex", "AGENTS.md"),
		},
	}

	// Cursor - Declarative Context Injection + Simple System Prompt
	agentRules[Cursor] = map[RuleLevel][]string{
		ProjectLevel: {
			".cursor/rules/",
		},
		AncestorLevel: {
			"AGENTS.md",
		},
	}

	// GitHub Copilot - Simple System Prompt + Hierarchical Concatenation + Agent Definition
	agentRules[Copilot] = map[RuleLevel][]string{
		ProjectLevel: {
			".github/agents/",
			".github/copilot-instructions.md",
		},
		AncestorLevel: {
			"AGENTS.md",
		},
	}

	// Augment CLI - Declarative Context Injection + Compatibility
	agentRules[Augment] = map[RuleLevel][]string{
		ProjectLevel: {
			".augment/rules/",
			".augment/guidelines.md",
		},
		AncestorLevel: {
			"CLAUDE.md",
			"AGENTS.md",
		},
	}

	// Windsurf (Codeium) - Declarative Context Injection
	agentRules[Windsurf] = map[RuleLevel][]string{
		ProjectLevel: {
			".windsurf/rules/",
		},
	}

	// Goose - Compatibility (External Standard)
	agentRules[Goose] = map[RuleLevel][]string{
		AncestorLevel: {
			"AGENTS.md",
		},
	}

	// Expand ancestor paths for all agents
	for agent, levels := range agentRules {
		if ancestorPaths, ok := levels[AncestorLevel]; ok {
			expanded := expandAncestorPaths(ancestorPaths)
			agentRules[agent][AncestorLevel] = expanded
		}
	}

	return nil
}

// expandAncestorPaths expands ancestor-level paths to search up the directory hierarchy
func expandAncestorPaths(paths []string) []string {
	expanded := make([]string, 0)
	
	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get cwd, return paths as-is
		return paths
	}
	
	for _, filename := range paths {
		// Search from cwd up to root
		dir := cwd
		for {
			ancestorPath := filepath.Join(dir, filename)
			expanded = append(expanded, ancestorPath)
			
			parent := filepath.Dir(dir)
			if parent == dir {
				// Reached root
				break
			}
			dir = parent
		}
	}
	
	return expanded
}
