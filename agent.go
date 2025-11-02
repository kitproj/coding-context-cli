package main

import (
	"os"
	"path/filepath"
)

// Agent represents a coding agent/tool
type Agent string

const (
	Claude       Agent = "Claude"
	Gemini       Agent = "Gemini"
	Cursor       Agent = "Cursor"
	Copilot      Agent = "Copilot"
	Codex        Agent = "Codex"
	Augment      Agent = "Augment"
	Windsurf     Agent = "Windsurf"
	Goose        Agent = "Goose"
	ContinueDev  Agent = "ContinueDev"
)

// RuleLevel represents the priority level of rules
type RuleLevel int

const (
	ProjectLevel  RuleLevel = 0 // Most important
	AncestorLevel RuleLevel = 1 // Next most important
	UserLevel     RuleLevel = 2
	SystemLevel   RuleLevel = 3 // Least important
)

// RulePath represents a path to rules with its level
type RulePath struct {
	Path  string
	Level RuleLevel
}

// agentRules maps each agent to its rule paths
// This will be populated on startup based on cwd
var agentRules map[Agent][]RulePath

// initAgentRules initializes the agent rules based on current working directory
func initAgentRules() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	agentRules = make(map[Agent][]RulePath)

	// Claude - Hierarchical Concatenation
	agentRules[Claude] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, "CLAUDE.local.md"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, "CLAUDE.md"), Level: ProjectLevel},
		
		// Ancestor Rules (will be expanded to search up the hierarchy)
		// User Rules
		{Path: filepath.Join(homeDir, ".claude", "CLAUDE.md"), Level: UserLevel},
	}

	// Gemini CLI - Hierarchical Concatenation + Simple System Prompt
	agentRules[Gemini] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, ".gemini", "styleguide.md"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, "GEMINI.md"), Level: ProjectLevel},
		
		// User Rules
		{Path: filepath.Join(homeDir, ".gemini", "GEMINI.md"), Level: UserLevel},
	}

	// Codex CLI - Hierarchical Concatenation
	agentRules[Codex] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, "AGENTS.md"), Level: ProjectLevel},
		
		// User Rules
		{Path: filepath.Join(homeDir, ".codex", "AGENTS.md"), Level: UserLevel},
	}

	// Cursor - Declarative Context Injection + Simple System Prompt
	agentRules[Cursor] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, ".cursor", "rules/"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, "AGENTS.md"), Level: ProjectLevel},
	}

	// GitHub Copilot - Simple System Prompt + Hierarchical Concatenation + Agent Definition
	agentRules[Copilot] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, ".github", "agents/"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, ".github", "copilot-instructions.md"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, "AGENTS.md"), Level: ProjectLevel},
	}

	// Augment CLI - Declarative Context Injection + Compatibility
	agentRules[Augment] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, ".augment", "rules/"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, ".augment", "guidelines.md"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, "CLAUDE.md"), Level: ProjectLevel},
		{Path: filepath.Join(cwd, "AGENTS.md"), Level: ProjectLevel},
	}

	// Windsurf (Codeium) - Declarative Context Injection
	agentRules[Windsurf] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, ".windsurf", "rules/"), Level: ProjectLevel},
	}

	// Goose - Compatibility (External Standard)
	agentRules[Goose] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, "AGENTS.md"), Level: ProjectLevel},
	}

	// Continue.dev
	agentRules[ContinueDev] = []RulePath{
		// Project Rules
		{Path: filepath.Join(cwd, ".continuerules"), Level: ProjectLevel},
	}

	return nil
}

// expandAncestorPaths expands ancestor-level paths to search up the directory hierarchy
func expandAncestorPaths(paths []RulePath) []RulePath {
	expanded := make([]RulePath, 0, len(paths))
	
	for _, rp := range paths {
		if rp.Level == AncestorLevel {
			// Search up the directory tree
			cwd, err := os.Getwd()
			if err != nil {
				continue
			}
			
			// Get the filename from the path
			filename := filepath.Base(rp.Path)
			
			// Search from cwd up to root
			dir := cwd
			for {
				ancestorPath := filepath.Join(dir, filename)
				expanded = append(expanded, RulePath{
					Path:  ancestorPath,
					Level: AncestorLevel,
				})
				
				parent := filepath.Dir(dir)
				if parent == dir {
					// Reached root
					break
				}
				dir = parent
			}
		} else {
			expanded = append(expanded, rp)
		}
	}
	
	return expanded
}
