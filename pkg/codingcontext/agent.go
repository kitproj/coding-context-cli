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

// TargetAgent represents the agent being used, which excludes other agents' rules
type TargetAgent struct {
	agent *Agent
}

// String implements the fmt.Stringer interface for TargetAgent
func (t *TargetAgent) String() string {
	if t.agent == nil {
		return ""
	}
	return t.agent.String()
}

// Set implements the flag.Value interface for TargetAgent
func (t *TargetAgent) Set(value string) error {
	agent, err := ParseAgent(value)
	if err != nil {
		return err
	}

	t.agent = &agent
	return nil
}

// ShouldExcludePath returns true if the given path should be excluded based on target agent
func (t *TargetAgent) ShouldExcludePath(path string) bool {
	if t.agent == nil {
		return false
	}

	// Exclude paths from ALL agents (including the target agent)
	// The target agent will use generic rules, which can filter themselves
	// with the agent selector in frontmatter
	for agent := range agentPathPatterns {
		if agent.MatchesPath(path) {
			return true
		}
	}

	return false
}

// Agent returns the target agent, or nil if not set
func (t *TargetAgent) Agent() *Agent {
	return t.agent
}
