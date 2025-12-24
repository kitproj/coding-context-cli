package codingcontext

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Agent represents an AI coding agent
type Agent string

// Supported agents
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

// ParseAgent parses a string into an Agent type
func ParseAgent(s string) (Agent, error) {
	agent := Agent(s)

	// Check if agent exists in the path patterns map
	if _, exists := agentPathPatterns[agent]; exists {
		return agent, nil
	}

	// Build list of supported agents for error message
	supported := make([]string, 0, len(agentPathPatterns))
	for a := range agentPathPatterns {
		supported = append(supported, a.String())
	}
	return "", fmt.Errorf("unknown agent: %s (supported: %s)", s, strings.Join(supported, ", "))
}

// String returns the string representation of the agent
func (a Agent) String() string {
	return string(a)
}

// PathPatterns returns the path patterns associated with this agent
func (a Agent) PathPatterns() []string {
	return agentPathPatterns[a]
}

// MatchesPath returns true if the given path matches any of the agent's patterns
func (a Agent) MatchesPath(path string) bool {
	normalizedPath := filepath.ToSlash(path)
	patterns := a.PathPatterns()

	for _, pattern := range patterns {
		if strings.Contains(normalizedPath, pattern) {
			return true
		}
	}

	return false
}

// agentPathPatterns maps agents to their associated path patterns
var agentPathPatterns = map[Agent][]string{
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

// Set implements the flag.Value interface for Agent
func (a *Agent) Set(value string) error {
	agent, err := ParseAgent(value)
	if err != nil {
		return fmt.Errorf("failed to set agent value: %w", err)
	}

	*a = agent
	return nil
}

// ShouldExcludePath returns true if the given path should be excluded based on this agent
// Empty agent means no exclusion
func (a Agent) ShouldExcludePath(path string) bool {
	if a == "" {
		return false
	}

	// Exclude paths from ONLY this agent
	// The agent will read its own rules, so we don't need to include them
	// But we might want rules from other agents or generic rules
	return a.MatchesPath(path)
}

// IsSet returns true if an agent has been specified (non-empty)
func (a Agent) IsSet() bool {
	return a != ""
}

// UserRulePath returns the primary user-level rules path for this agent relative to home directory.
// Returns an empty string if the agent is not set.
// The path is relative and should be joined with the home directory.
func (a Agent) UserRulePath() string {
	if !a.IsSet() {
		return ""
	}

	// Map each agent to its primary user rules path (relative to home directory)
	// All paths are files (not directories)
	switch a {
	case AgentCursor:
		return filepath.Join(".cursor", "rules", "AGENTS.md")
	case AgentOpenCode:
		return filepath.Join(".opencode", "rules", "AGENTS.md")
	case AgentCopilot:
		return filepath.Join(".github", "agents", "AGENTS.md")
	case AgentClaude:
		return filepath.Join(".claude", "CLAUDE.md")
	case AgentGemini:
		return filepath.Join(".gemini", "GEMINI.md")
	case AgentAugment:
		return filepath.Join(".augment", "rules", "AGENTS.md")
	case AgentWindsurf:
		return filepath.Join(".windsurf", "rules", "AGENTS.md")
	case AgentCodex:
		return filepath.Join(".codex", "AGENTS.md")
	default:
		return ""
	}
}
