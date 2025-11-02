package main

import (
	"os"
	"path/filepath"
)

// initAgentRules initializes and returns the agent rules map
func initAgentRules() (map[Agent][]RulePath, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	agentRules := make(map[Agent][]RulePath)

	// Default agent - normalized storage for all rules
	agentRules[Default] = []RulePath{
		NewRulePath(".agents/rules", ".agents/rules"),
		NewRulePath(filepath.Join(homeDir, ".agents", "rules"), filepath.Join(homeDir, ".agents", "rules")),
		NewRulePath(filepath.Join(homeDir, ".agents", "AGENTS.md"), filepath.Join(homeDir, ".agents", "AGENTS.md")),
		NewRulePath("/etc/agents/rules", "/etc/agents/rules"),
	}
	// Add ancestor paths for Default
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		agentRules[Default] = append(agentRules[Default], NewRulePath(ancestorPath, ancestorPath))
	}

	// Claude - Hierarchical Concatenation
	agentRules[Claude] = []RulePath{
		NewRulePath("CLAUDE.local.md", ".agents/rules/local.md"),
		NewRulePath(filepath.Join(homeDir, ".claude", "CLAUDE.md"), filepath.Join(homeDir, ".agents", "rules", "CLAUDE.md")),
	}
	// Add ancestor paths for Claude
	for _, ancestorPath := range expandAncestorPaths("CLAUDE.md") {
		agentRules[Claude] = append(agentRules[Claude], NewRulePath(ancestorPath, "AGENTS.md"))
	}

	// Gemini CLI - Hierarchical Concatenation + Simple System Prompt
	agentRules[Gemini] = []RulePath{
		NewRulePath(".gemini/styleguide.md", ".agents/rules/gemini-styleguide.md"),
		NewRulePath(filepath.Join(homeDir, ".gemini", "GEMINI.md"), filepath.Join(homeDir, ".agents", "rules", "GEMINI.md")),
	}
	// Add ancestor paths for Gemini
	for _, ancestorPath := range expandAncestorPaths("GEMINI.md") {
		agentRules[Gemini] = append(agentRules[Gemini], NewRulePath(ancestorPath, "AGENTS.md"))
	}

	// Codex CLI - Hierarchical Concatenation
	agentRules[Codex] = []RulePath{
		NewRulePath(filepath.Join(homeDir, ".codex", "AGENTS.md"), filepath.Join(homeDir, ".agents", "AGENTS.md")),
	}
	// Add ancestor paths for Codex
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		agentRules[Codex] = append(agentRules[Codex], NewRulePath(ancestorPath, "AGENTS.md"))
	}

	// Cursor - Declarative Context Injection + Simple System Prompt
	agentRules[Cursor] = []RulePath{
		NewRulePath(".cursor/rules/", ".agents/rules/cursor"),
	}
	// Add ancestor paths for Cursor
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		agentRules[Cursor] = append(agentRules[Cursor], NewRulePath(ancestorPath, "AGENTS.md"))
	}

	// GitHub Copilot - Simple System Prompt + Hierarchical Concatenation + Agent Definition
	agentRules[Copilot] = []RulePath{
		NewRulePath(".github/agents/", ".agents/rules/copilot-agents"),
		NewRulePath(".github/copilot-instructions.md", ".agents/rules/copilot-instructions.md"),
	}
	// Add ancestor paths for Copilot
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		agentRules[Copilot] = append(agentRules[Copilot], NewRulePath(ancestorPath, "AGENTS.md"))
	}

	// Augment CLI - Declarative Context Injection + Compatibility
	agentRules[Augment] = []RulePath{
		NewRulePath(".augment/rules/", ".agents/rules/augment"),
		NewRulePath(".augment/guidelines.md", ".agents/rules/augment-guidelines.md"),
	}
	// Add ancestor paths for Augment (CLAUDE.md and AGENTS.md)
	for _, ancestorPath := range expandAncestorPaths("CLAUDE.md") {
		agentRules[Augment] = append(agentRules[Augment], NewRulePath(ancestorPath, "AGENTS.md"))
	}
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		agentRules[Augment] = append(agentRules[Augment], NewRulePath(ancestorPath, "AGENTS.md"))
	}

	// Windsurf (Codeium) - Declarative Context Injection
	agentRules[Windsurf] = []RulePath{
		NewRulePath(".windsurf/rules/", ".agents/rules/windsurf"),
	}

	// Goose - Compatibility (External Standard)
	agentRules[Goose] = []RulePath{}
	// Add ancestor paths for Goose
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		agentRules[Goose] = append(agentRules[Goose], NewRulePath(ancestorPath, "AGENTS.md"))
	}

	return agentRules, nil
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
