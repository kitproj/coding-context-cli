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

	// Default agent - normalized storage for all rules
	agentRules[Default] = map[RuleLevel][]string{
		ProjectLevel: {
			".agents/rules",
		},
		AncestorLevel: expandAncestorPaths("AGENTS.md"),
		UserLevel: {
			filepath.Join(homeDir, ".agents", "rules"),
			filepath.Join(homeDir, ".agents", "AGENTS.md"),
		},
		SystemLevel: {
			"/etc/agents/rules",
		},
	}

	// Claude - Hierarchical Concatenation
	agentRules[Claude] = map[RuleLevel][]string{
		ProjectLevel: {
			"CLAUDE.local.md",
		},
		AncestorLevel: expandAncestorPaths("CLAUDE.md"),
		UserLevel: {
			filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		},
	}

	// Gemini CLI - Hierarchical Concatenation + Simple System Prompt
	agentRules[Gemini] = map[RuleLevel][]string{
		ProjectLevel: {
			".gemini/styleguide.md",
		},
		AncestorLevel: expandAncestorPaths("GEMINI.md"),
		UserLevel: {
			filepath.Join(homeDir, ".gemini", "GEMINI.md"),
		},
	}

	// Codex CLI - Hierarchical Concatenation
	agentRules[Codex] = map[RuleLevel][]string{
		AncestorLevel: expandAncestorPaths("AGENTS.md"),
		UserLevel: {
			filepath.Join(homeDir, ".codex", "AGENTS.md"),
		},
	}

	// Cursor - Declarative Context Injection + Simple System Prompt
	agentRules[Cursor] = map[RuleLevel][]string{
		ProjectLevel: {
			".cursor/rules/",
		},
		AncestorLevel: expandAncestorPaths("AGENTS.md"),
	}

	// GitHub Copilot - Simple System Prompt + Hierarchical Concatenation + Agent Definition
	agentRules[Copilot] = map[RuleLevel][]string{
		ProjectLevel: {
			".github/agents/",
			".github/copilot-instructions.md",
		},
		AncestorLevel: expandAncestorPaths("AGENTS.md"),
	}

	// Augment CLI - Declarative Context Injection + Compatibility
	agentRules[Augment] = map[RuleLevel][]string{
		ProjectLevel: {
			".augment/rules/",
			".augment/guidelines.md",
		},
		AncestorLevel: append(expandAncestorPaths("CLAUDE.md"), expandAncestorPaths("AGENTS.md")...),
	}

	// Windsurf (Codeium) - Declarative Context Injection
	agentRules[Windsurf] = map[RuleLevel][]string{
		ProjectLevel: {
			".windsurf/rules/",
		},
	}

	// Goose - Compatibility (External Standard)
	agentRules[Goose] = map[RuleLevel][]string{
		AncestorLevel: expandAncestorPaths("AGENTS.md"),
	}

	return nil
}

// expandAncestorPaths expands ancestor-level paths to search up the directory hierarchy
func expandAncestorPaths(filename string) []string {
	expanded := make([]string, 0)
	
	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get cwd, return filename as-is
		return []string{filename}
	}
	
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
	
	return expanded
}
