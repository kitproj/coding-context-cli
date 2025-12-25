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
			paths, exists := agentsPaths[tt.agent]
			if !exists {
				t.Errorf("Agent %q not found in agentsPaths", tt.agent)
				return
			}

			// Check that at least one path is defined
			hasAnyPath := len(paths.rulesPaths) > 0 ||
				paths.skillsPath != "" ||
				paths.commandsPath != "" ||
				paths.tasksPath != ""

			if !hasAnyPath {
				t.Errorf("Agent %q has no paths defined", tt.agent)
			}
		})
	}
}

func TestAgentPaths_EmptyAgentHasAllPaths(t *testing.T) {
	paths, exists := agentsPaths[Agent("")]
	if !exists {
		t.Fatal("Empty agent not found in agentsPaths")
	}

	if len(paths.rulesPaths) == 0 {
		t.Error("Empty agent should have rulesPaths defined")
	}
	if paths.skillsPath == "" {
		t.Error("Empty agent should have skillsPath defined")
	}
	if paths.commandsPath == "" {
		t.Error("Empty agent should have commandsPath defined")
	}
	if paths.tasksPath == "" {
		t.Error("Empty agent should have tasksPath defined")
	}
}

func TestAgentPaths_RulesPathsNotEmpty(t *testing.T) {
	// Every agent should have at least one rules path
	for agent, paths := range agentsPaths {
		if len(paths.rulesPaths) == 0 {
			t.Errorf("Agent %q should have at least one rulesPaths entry", agent)
		}
	}
}

func TestAgentPaths_NoAbsolutePaths(t *testing.T) {
	// All paths should be relative (not absolute)
	for agent, paths := range agentsPaths {
		for _, rulePath := range paths.rulesPaths {
			if len(rulePath) > 0 && rulePath[0] == '/' {
				t.Errorf("Agent %q rulesPaths contains absolute path: %q", agent, rulePath)
			}
		}
		if len(paths.skillsPath) > 0 && paths.skillsPath[0] == '/' {
			t.Errorf("Agent %q skillsPath is absolute: %q", agent, paths.skillsPath)
		}
		if len(paths.commandsPath) > 0 && paths.commandsPath[0] == '/' {
			t.Errorf("Agent %q commandsPath is absolute: %q", agent, paths.commandsPath)
		}
		if len(paths.tasksPath) > 0 && paths.tasksPath[0] == '/' {
			t.Errorf("Agent %q tasksPath is absolute: %q", agent, paths.tasksPath)
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
			wantSkillsPath: ".cursor/skills",
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
			paths, exists := agentsPaths[tt.agent]
			if !exists {
				t.Fatalf("Agent %q not found in agentsPaths", tt.agent)
			}

			gotRulesPaths := paths.rulesPaths
			if len(gotRulesPaths) != len(tt.wantRulesPaths) {
				t.Errorf("rulesPaths length = %d, want %d", len(gotRulesPaths), len(tt.wantRulesPaths))
			}
			for i, want := range tt.wantRulesPaths {
				if i < len(gotRulesPaths) && gotRulesPaths[i] != want {
					t.Errorf("rulesPaths[%d] = %q, want %q", i, gotRulesPaths[i], want)
				}
			}

			if got := paths.skillsPath; got != tt.wantSkillsPath {
				t.Errorf("skillsPath = %q, want %q", got, tt.wantSkillsPath)
			}
		})
	}
}
