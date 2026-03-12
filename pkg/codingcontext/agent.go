package codingcontext

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ErrUnknownAgent is returned when parsing an unknown or unsupported agent name.
var ErrUnknownAgent = errors.New("unknown agent")

// Agent represents an AI coding agent.
type Agent string

// Supported agents.
const (
	AgentCursor   Agent = "cursor"
	AgentOpenCode Agent = "opencode"
	AgentCopilot  Agent = "copilot"
	AgentClaude   Agent = "claude"
	AgentGemini   Agent = "gemini"
	AgentAugment  Agent = "augment"
	AgentWindsurf Agent = "windsurf"
	AgentCodex    Agent = "codex"
)

// ParseAgent parses a string into an Agent type.
func ParseAgent(s string) (Agent, error) {
	agent := Agent(s)

	patterns := getAgentPathPatterns()

	// Check if agent exists in the path patterns map
	if _, exists := patterns[agent]; exists {
		return agent, nil
	}

	// Build list of supported agents for error message
	supported := make([]string, 0, len(patterns))
	for a := range patterns {
		supported = append(supported, a.String())
	}

	return "", fmt.Errorf("%w: %s (supported: %s)", ErrUnknownAgent, s, strings.Join(supported, ", "))
}

// String returns the string representation of the agent.
func (a *Agent) String() string {
	if a == nil {
		return ""
	}

	return string(*a)
}

// PathPatterns returns the path patterns associated with this agent.
func (a *Agent) PathPatterns() []string {
	if a == nil {
		return nil
	}

	return getAgentPathPatterns()[*a]
}

// MatchesPath returns true if the given path matches any of the agent's patterns.
func (a *Agent) MatchesPath(path string) bool {
	if a == nil {
		return false
	}

	normalizedPath := filepath.ToSlash(path)
	patterns := a.PathPatterns()

	for _, pattern := range patterns {
		if strings.Contains(normalizedPath, pattern) {
			return true
		}
	}

	return false
}

// getAgentPathPatterns returns the map of agents to their associated path patterns.
func getAgentPathPatterns() map[Agent][]string {
	return map[Agent][]string{
		AgentCursor: {
			".cursor/",
			".cursorrules",
		},
		AgentOpenCode: {
			".opencode/",
		},
		AgentCopilot: {
			".github/copilot-instructions.md",
			".github/agents/",
		},
		AgentClaude: {
			".claude/",
			"CLAUDE.md",
			"CLAUDE.local.md",
		},
		AgentGemini: {
			".gemini/",
			"GEMINI.md",
		},
		AgentAugment: {
			".augment/",
		},
		AgentWindsurf: {
			".windsurf/",
			".windsurfrules",
		},
		AgentCodex: {
			".codex/",
			"AGENTS.md",
		},
	}
}

// Set implements the flag.Value interface for Agent.
func (a *Agent) Set(value string) error {
	agent, err := ParseAgent(value)
	if err != nil {
		return fmt.Errorf("failed to set agent value %q: %w", value, err)
	}

	*a = agent

	return nil
}

// ShouldExcludePath returns true if the given path should be excluded based on this agent
// Empty agent means no exclusion.
func (a *Agent) ShouldExcludePath(path string) bool {
	if a == nil || *a == "" {
		return false
	}

	// Exclude paths from ONLY this agent
	// The agent will read its own rules, so we don't need to include them
	// But we might want rules from other agents or generic rules
	return a.MatchesPath(path)
}

// IsSet returns true if an agent has been specified (non-empty).
func (a *Agent) IsSet() bool {
	return a != nil && *a != ""
}

// getAgentUserRulePaths returns the map of each agent to its primary user rules path (relative to home directory).
func getAgentUserRulePaths() map[Agent]string {
	return map[Agent]string{
		AgentCursor:   filepath.Join(".cursor", "rules", "AGENTS.md"),
		AgentOpenCode: filepath.Join(".opencode", "rules", "AGENTS.md"),
		AgentCopilot:  filepath.Join(".github", "agents", "AGENTS.md"),
		AgentClaude:   filepath.Join(".claude", "CLAUDE.md"),
		AgentGemini:   filepath.Join(".gemini", "GEMINI.md"),
		AgentAugment:  filepath.Join(".augment", "rules", "AGENTS.md"),
		AgentWindsurf: filepath.Join(".windsurf", "rules", "AGENTS.md"),
		AgentCodex:    filepath.Join(".codex", "AGENTS.md"),
	}
}

// UserRulePath returns the primary user-level rules path for this agent relative to home directory.
// Returns an empty string if the agent is not set.
// The path is relative and should be joined with the home directory.
func (a *Agent) UserRulePath() string {
	if a == nil || !a.IsSet() {
		return ""
	}

	return getAgentUserRulePaths()[*a]
}
