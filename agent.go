package main

import (
	"os"
	"path/filepath"
)

// initImportRules initializes and returns the import rules map (source -> target)
// Import reads from agent-specific locations and writes to Default agent locations
func initImportRules() (map[Agent][]RulePath, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	importRules := make(map[Agent][]RulePath)

	// Default agent - normalized storage for all rules (no import from itself)
	importRules[Default] = []RulePath{}

	// Claude - Hierarchical Concatenation
	importRules[Claude] = []RulePath{
		NewRulePath("CLAUDE.local.md", ".agents/rules/local.md"),
		NewRulePath(filepath.Join(homeDir, ".claude", "CLAUDE.md"), filepath.Join(homeDir, ".agents", "rules", "CLAUDE.md")),
	}
	// Add ancestor paths for Claude
	for _, ancestorPath := range expandAncestorPaths("CLAUDE.md") {
		importRules[Claude] = append(importRules[Claude], NewRulePath(ancestorPath, ".agents/rules/CLAUDE.md"))
	}

	// Gemini CLI - Hierarchical Concatenation + Simple System Prompt
	importRules[Gemini] = []RulePath{
		NewRulePath(".gemini/styleguide.md", ".agents/rules/gemini-styleguide.md"),
		NewRulePath(filepath.Join(homeDir, ".gemini", "GEMINI.md"), filepath.Join(homeDir, ".agents", "rules", "GEMINI.md")),
	}
	// Add ancestor paths for Gemini
	for _, ancestorPath := range expandAncestorPaths("GEMINI.md") {
		importRules[Gemini] = append(importRules[Gemini], NewRulePath(ancestorPath, ".agents/rules/GEMINI.md"))
	}

	// Codex CLI - Hierarchical Concatenation
	importRules[Codex] = []RulePath{
		NewRulePath(filepath.Join(homeDir, ".codex", "AGENTS.md"), filepath.Join(homeDir, ".agents", "rules", "CODEX.md")),
	}
	// Add ancestor paths for Codex
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		importRules[Codex] = append(importRules[Codex], NewRulePath(ancestorPath, ".agents/rules/CODEX.md"))
	}

	// Cursor - Declarative Context Injection + Simple System Prompt
	importRules[Cursor] = []RulePath{
		NewRulePath(".cursor/rules/", ".agents/rules/cursor"),
	}
	// Add ancestor paths for Cursor
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		importRules[Cursor] = append(importRules[Cursor], NewRulePath(ancestorPath, ".agents/rules/cursor.md"))
	}

	// GitHub Copilot - Simple System Prompt + Hierarchical Concatenation + Agent Definition
	importRules[Copilot] = []RulePath{
		NewRulePath(".github/agents/", ".agents/rules/copilot-agents"),
		NewRulePath(".github/copilot-instructions.md", ".agents/rules/copilot-instructions.md"),
	}
	// Add ancestor paths for Copilot
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		importRules[Copilot] = append(importRules[Copilot], NewRulePath(ancestorPath, ".agents/rules/copilot.md"))
	}

	// Augment CLI - Declarative Context Injection + Compatibility
	importRules[Augment] = []RulePath{
		NewRulePath(".augment/rules/", ".agents/rules/augment"),
		NewRulePath(".augment/guidelines.md", ".agents/rules/augment-guidelines.md"),
	}
	// Add ancestor paths for Augment (CLAUDE.md and AGENTS.md)
	for _, ancestorPath := range expandAncestorPaths("CLAUDE.md") {
		importRules[Augment] = append(importRules[Augment], NewRulePath(ancestorPath, ".agents/rules/augment-CLAUDE.md"))
	}
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		importRules[Augment] = append(importRules[Augment], NewRulePath(ancestorPath, ".agents/rules/augment.md"))
	}

	// Windsurf (Codeium) - Declarative Context Injection
	importRules[Windsurf] = []RulePath{
		NewRulePath(".windsurf/rules/", ".agents/rules/windsurf"),
	}

	// Goose - Compatibility (External Standard)
	importRules[Goose] = []RulePath{}
	// Add ancestor paths for Goose
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		importRules[Goose] = append(importRules[Goose], NewRulePath(ancestorPath, ".agents/rules/goose.md"))
	}

	return importRules, nil
}

// initExportRules initializes and returns the export rules map (source -> target)
// Export reads from Default agent locations and writes to agent-specific locations
func initExportRules() (map[Agent][]RulePath, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	exportRules := make(map[Agent][]RulePath)

	// Default agent - normalized storage (source for exports)
	exportRules[Default] = []RulePath{
		NewRulePath(".agents/rules", ".agents/rules"),
		NewRulePath(filepath.Join(homeDir, ".agents", "rules"), filepath.Join(homeDir, ".agents", "rules")),
		NewRulePath("/etc/agents/rules", "/etc/agents/rules"),
	}

	// Claude - Hierarchical Concatenation
	exportRules[Claude] = []RulePath{
		NewRulePath(".agents/rules/local.md", "CLAUDE.local.md"),
		NewRulePath(filepath.Join(homeDir, ".agents", "rules", "CLAUDE.md"), filepath.Join(homeDir, ".claude", "CLAUDE.md")),
	}
	// Add ancestor paths for Claude
	for _, ancestorPath := range expandAncestorPaths("CLAUDE.md") {
		exportRules[Claude] = append(exportRules[Claude], NewRulePath(".agents/rules/CLAUDE.md", ancestorPath))
	}

	// Gemini CLI - Hierarchical Concatenation + Simple System Prompt
	exportRules[Gemini] = []RulePath{
		NewRulePath(".agents/rules/gemini-styleguide.md", ".gemini/styleguide.md"),
		NewRulePath(filepath.Join(homeDir, ".agents", "rules", "GEMINI.md"), filepath.Join(homeDir, ".gemini", "GEMINI.md")),
	}
	// Add ancestor paths for Gemini
	for _, ancestorPath := range expandAncestorPaths("GEMINI.md") {
		exportRules[Gemini] = append(exportRules[Gemini], NewRulePath(".agents/rules/GEMINI.md", ancestorPath))
	}

	// Codex CLI - Hierarchical Concatenation
	exportRules[Codex] = []RulePath{
		NewRulePath(filepath.Join(homeDir, ".agents", "rules", "CODEX.md"), filepath.Join(homeDir, ".codex", "AGENTS.md")),
	}
	// Add ancestor paths for Codex
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		exportRules[Codex] = append(exportRules[Codex], NewRulePath(".agents/rules/CODEX.md", ancestorPath))
	}

	// Cursor - Declarative Context Injection + Simple System Prompt
	exportRules[Cursor] = []RulePath{
		NewRulePath(".agents/rules/cursor", ".cursor/rules/"),
	}
	// Add ancestor paths for Cursor
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		exportRules[Cursor] = append(exportRules[Cursor], NewRulePath(".agents/rules/cursor.md", ancestorPath))
	}

	// GitHub Copilot - Simple System Prompt + Hierarchical Concatenation + Agent Definition
	exportRules[Copilot] = []RulePath{
		NewRulePath(".agents/rules/copilot-agents", ".github/agents/"),
		NewRulePath(".agents/rules/copilot-instructions.md", ".github/copilot-instructions.md"),
	}
	// Add ancestor paths for Copilot
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		exportRules[Copilot] = append(exportRules[Copilot], NewRulePath(".agents/rules/copilot.md", ancestorPath))
	}

	// Augment CLI - Declarative Context Injection + Compatibility
	exportRules[Augment] = []RulePath{
		NewRulePath(".agents/rules/augment", ".augment/rules/"),
		NewRulePath(".agents/rules/augment-guidelines.md", ".augment/guidelines.md"),
	}
	// Add ancestor paths for Augment (CLAUDE.md and AGENTS.md)
	for _, ancestorPath := range expandAncestorPaths("CLAUDE.md") {
		exportRules[Augment] = append(exportRules[Augment], NewRulePath(".agents/rules/augment-CLAUDE.md", ancestorPath))
	}
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		exportRules[Augment] = append(exportRules[Augment], NewRulePath(".agents/rules/augment.md", ancestorPath))
	}

	// Windsurf (Codeium) - Declarative Context Injection
	exportRules[Windsurf] = []RulePath{
		NewRulePath(".agents/rules/windsurf", ".windsurf/rules/"),
	}

	// Goose - Compatibility (External Standard)
	exportRules[Goose] = []RulePath{}
	// Add ancestor paths for Goose
	for _, ancestorPath := range expandAncestorPaths("AGENTS.md") {
		exportRules[Goose] = append(exportRules[Goose], NewRulePath(".agents/rules/goose.md", ancestorPath))
	}

	return exportRules, nil
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
