package codingcontext

import (
	"testing"
)

func TestAgentPaths_Structure(t *testing.T) {
	tests := []struct {
		name  string
		agent Agent
	}{
		{
			name:  "empty agent (generic .agents)",
			agent: Agent(""),
		},
		{
			name:  "cursor agent",
			agent: AgentCursor,
		},
		{
			name:  "opencode agent",
			agent: AgentOpenCode,
		},
		{
			name:  "copilot agent",
			agent: AgentCopilot,
		},
		{
			name:  "claude agent",
			agent: AgentClaude,
		},
		{
			name:  "gemini agent",
			agent: AgentGemini,
		},
		{
			name:  "augment agent",
			agent: AgentAugment,
		},
		{
			name:  "windsurf agent",
			agent: AgentWindsurf,
		},
		{
			name:  "codex agent",
			agent: AgentCodex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.agent.Paths()

			// Check that at least one path is defined
			hasAnyPath := len(paths.RulesPaths()) > 0 ||
				paths.SkillsPath() != "" ||
				paths.CommandsPath() != "" ||
				paths.TasksPath() != ""

			if !hasAnyPath {
				t.Errorf("Agent %q has no paths defined", tt.agent)
			}
		})
	}
}

func TestAgentPaths_EmptyAgentHasAllPaths(t *testing.T) {
	paths := Agent("").Paths()

	if len(paths.RulesPaths()) == 0 {
		t.Error("Empty agent should have RulesPaths defined")
	}
	if paths.SkillsPath() == "" {
		t.Error("Empty agent should have SkillsPath defined")
	}
	if paths.CommandsPath() == "" {
		t.Error("Empty agent should have CommandsPath defined")
	}
	if paths.TasksPath() == "" {
		t.Error("Empty agent should have TasksPath defined")
	}
}

func TestAgentPaths_RulesPathsNotEmpty(t *testing.T) {
	// Every agent should have at least one rules path
	for agent := range agentsPaths {
		paths := agent.Paths()
		if len(paths.RulesPaths()) == 0 {
			t.Errorf("Agent %q should have at least one RulesPaths entry", agent)
		}
	}
}

func TestAgentPaths_NoAbsolutePaths(t *testing.T) {
	// All paths should be relative (not absolute)
	for agent := range agentsPaths {
		paths := agent.Paths()
		for _, rulePath := range paths.RulesPaths() {
			if len(rulePath) > 0 && rulePath[0] == '/' {
				t.Errorf("Agent %q RulesPaths contains absolute path: %q", agent, rulePath)
			}
		}
		if len(paths.SkillsPath()) > 0 && paths.SkillsPath()[0] == '/' {
			t.Errorf("Agent %q SkillsPath is absolute: %q", agent, paths.SkillsPath())
		}
		if len(paths.CommandsPath()) > 0 && paths.CommandsPath()[0] == '/' {
			t.Errorf("Agent %q CommandsPath is absolute: %q", agent, paths.CommandsPath())
		}
		if len(paths.TasksPath()) > 0 && paths.TasksPath()[0] == '/' {
			t.Errorf("Agent %q TasksPath is absolute: %q", agent, paths.TasksPath())
		}
	}
}

func TestAgentPaths_Count(t *testing.T) {
	// Should have 9 entries: 1 empty agent + 8 named agents
	expectedCount := 9
	if len(agentsPaths) != expectedCount {
		t.Errorf("agentsPaths should have %d entries, got %d", expectedCount, len(agentsPaths))
	}
}

func TestAgent_Paths(t *testing.T) {
	tests := []struct {
		name           string
		agent          Agent
		wantRulesPaths []string
		wantSkillsPath string
	}{
		{
			name:           "cursor agent",
			agent:          AgentCursor,
			wantRulesPaths: []string{".cursor/rules", ".cursorrules"},
			wantSkillsPath: "",
		},
		{
			name:           "empty agent",
			agent:          Agent(""),
			wantRulesPaths: []string{".agents/rules"},
			wantSkillsPath: ".agents/skills",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.agent.Paths()

			gotRulesPaths := paths.RulesPaths()
			if len(gotRulesPaths) != len(tt.wantRulesPaths) {
				t.Errorf("RulesPaths() length = %d, want %d", len(gotRulesPaths), len(tt.wantRulesPaths))
			}
			for i, want := range tt.wantRulesPaths {
				if i < len(gotRulesPaths) && gotRulesPaths[i] != want {
					t.Errorf("RulesPaths()[%d] = %q, want %q", i, gotRulesPaths[i], want)
				}
			}

			if got := paths.SkillsPath(); got != tt.wantSkillsPath {
				t.Errorf("SkillsPath() = %q, want %q", got, tt.wantSkillsPath)
			}
		})
	}
}
