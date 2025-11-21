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

// AllAgents returns all supported agents
func AllAgents() []Agent {
	agents := make([]Agent, 0, len(agentPathPatterns))
	for agent := range agentPathPatterns {
		agents = append(agents, agent)
	}
	return agents
}

// ParseAgent parses a string into an Agent type
func ParseAgent(s string) (Agent, error) {
	normalized := Agent(strings.ToLower(strings.TrimSpace(s)))

	// Check if agent exists in the path patterns map
	if _, exists := agentPathPatterns[normalized]; exists {
		return normalized, nil
	}

	// Build list of supported agents for error message
	supported := make([]string, 0, len(agentPathPatterns))
	for agent := range agentPathPatterns {
		supported = append(supported, agent.String())
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
	},
}

// AgentExcludes stores which agents to exclude rules from
type AgentExcludes map[Agent]bool

// String implements the fmt.Stringer interface for AgentExcludes
func (e *AgentExcludes) String() string {
	if *e == nil {
		return ""
	}
	var names []string
	for agent := range *e {
		names = append(names, agent.String())
	}
	return strings.Join(names, ",")
}

// Set implements the flag.Value interface for AgentExcludes
func (e *AgentExcludes) Set(value string) error {
	if *e == nil {
		*e = make(AgentExcludes)
	}

	agent, err := ParseAgent(value)
	if err != nil {
		return err
	}

	(*e)[agent] = true
	return nil
}

// ShouldExcludePath returns true if the given path should be excluded
func (e *AgentExcludes) ShouldExcludePath(path string) bool {
	if *e == nil || len(*e) == 0 {
		return false
	}

	// Check if any excluded agent matches this path
	for agent := range *e {
		if agent.MatchesPath(path) {
			return true
		}
	}

	return false
}
