package codingcontext

import (
	"path/filepath"
	"testing"
)

func TestParseAgent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Agent
		wantErr bool
	}{
		{
			name:    "valid - cursor",
			input:   "cursor",
			want:    AgentCursor,
			wantErr: false,
		},
		{
			name:    "valid - opencode",
			input:   "opencode",
			want:    AgentOpenCode,
			wantErr: false,
		},
		{
			name:    "valid - copilot",
			input:   "copilot",
			want:    AgentCopilot,
			wantErr: false,
		},
		{
			name:    "valid - claude",
			input:   "claude",
			want:    AgentClaude,
			wantErr: false,
		},
		{
			name:    "valid - gemini",
			input:   "gemini",
			want:    AgentGemini,
			wantErr: false,
		},
		{
			name:    "valid - augment",
			input:   "augment",
			want:    AgentAugment,
			wantErr: false,
		},
		{
			name:    "valid - windsurf",
			input:   "windsurf",
			want:    AgentWindsurf,
			wantErr: false,
		},
		{
			name:    "valid - codex",
			input:   "codex",
			want:    AgentCodex,
			wantErr: false,
		},
		{
			name:    "uppercase should fail",
			input:   "CURSOR",
			want:    "",
			wantErr: true,
		},
		{
			name:    "mixed case should fail",
			input:   "OpenCode",
			want:    "",
			wantErr: true,
		},
		{
			name:    "with spaces should fail",
			input:   "  cursor  ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid agent",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_MatchesPath(t *testing.T) {
	tests := []struct {
		name      string
		agent     Agent
		path      string
		wantMatch bool
	}{
		{
			name:      "cursor matches .cursor/rules",
			agent:     AgentCursor,
			path:      ".cursor/rules/example.md",
			wantMatch: true,
		},
		{
			name:      "cursor matches .cursorrules",
			agent:     AgentCursor,
			path:      ".cursorrules",
			wantMatch: true,
		},
		{
			name:      "cursor does not match .agents/rules",
			agent:     AgentCursor,
			path:      ".agents/rules/example.md",
			wantMatch: false,
		},
		{
			name:      "opencode matches .opencode/agent",
			agent:     AgentOpenCode,
			path:      ".opencode/agent/rule.md",
			wantMatch: true,
		},
		{
			name:      "opencode matches .opencode/command",
			agent:     AgentOpenCode,
			path:      ".opencode/command/task.md",
			wantMatch: true,
		},
		{
			name:      "copilot matches instructions",
			agent:     AgentCopilot,
			path:      ".github/copilot-instructions.md",
			wantMatch: true,
		},
		{
			name:      "copilot matches agents dir",
			agent:     AgentCopilot,
			path:      ".github/agents/rule.md",
			wantMatch: true,
		},
		{
			name:      "claude matches CLAUDE.md",
			agent:     AgentClaude,
			path:      "CLAUDE.md",
			wantMatch: true,
		},
		{
			name:      "claude matches CLAUDE.local.md",
			agent:     AgentClaude,
			path:      "CLAUDE.local.md",
			wantMatch: true,
		},
		{
			name:      "claude matches .claude dir",
			agent:     AgentClaude,
			path:      ".claude/CLAUDE.md",
			wantMatch: true,
		},
		{
			name:      "gemini matches GEMINI.md",
			agent:     AgentGemini,
			path:      "GEMINI.md",
			wantMatch: true,
		},
		{
			name:      "gemini matches .gemini/styleguide.md",
			agent:     AgentGemini,
			path:      ".gemini/styleguide.md",
			wantMatch: true,
		},
		{
			name:      "augment matches .augment/rules",
			agent:     AgentAugment,
			path:      ".augment/rules/example.md",
			wantMatch: true,
		},
		{
			name:      "windsurf matches .windsurf/rules",
			agent:     AgentWindsurf,
			path:      ".windsurf/rules/example.md",
			wantMatch: true,
		},
		{
			name:      "windsurf matches .windsurfrules",
			agent:     AgentWindsurf,
			path:      ".windsurfrules",
			wantMatch: true,
		},
		{
			name:      "codex matches .codex dir",
			agent:     AgentCodex,
			path:      ".codex/AGENTS.md",
			wantMatch: true,
		},
		{
			name:      "codex matches AGENTS.md",
			agent:     AgentCodex,
			path:      "AGENTS.md",
			wantMatch: true,
		},
		{
			name:      "absolute path matching",
			agent:     AgentCursor,
			path:      "/home/user/project/.cursor/rules/example.md",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Normalize path for testing
			normalizedPath := filepath.FromSlash(tt.path)

			if got := tt.agent.MatchesPath(normalizedPath); got != tt.wantMatch {
				t.Errorf("Agent.MatchesPath(%q) = %v, want %v", tt.path, got, tt.wantMatch)
			}
		})
	}
}

func TestTargetAgent_Set(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantAgent Agent
		wantErr   bool
	}{
		{
			name:      "valid - cursor",
			value:     "cursor",
			wantAgent: AgentCursor,
			wantErr:   false,
		},
		{
			name:      "valid - opencode",
			value:     "opencode",
			wantAgent: AgentOpenCode,
			wantErr:   false,
		},
		{
			name:      "valid - copilot",
			value:     "copilot",
			wantAgent: AgentCopilot,
			wantErr:   false,
		},
		{
			name:    "uppercase should fail",
			value:   "CURSOR",
			wantErr: true,
		},
		{
			name:    "with spaces should fail",
			value:   "  cursor  ",
			wantErr: true,
		},
		{
			name:    "invalid agent",
			value:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ta TargetAgent
			err := ta.Set(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if ta.Agent() == nil {
					t.Errorf("Set() agent is nil, want %v", tt.wantAgent)
				} else if *ta.Agent() != tt.wantAgent {
					t.Errorf("Set() agent = %v, want %v", *ta.Agent(), tt.wantAgent)
				}
			}
		})
	}
}

func TestTargetAgent_ShouldExcludePath(t *testing.T) {
	tests := []struct {
		name        string
		targetAgent string
		path        string
		wantExclude bool
	}{
		{
			name:        "target cursor - do not exclude opencode rules",
			targetAgent: "cursor",
			path:        ".opencode/agent/rule.md",
			wantExclude: false,
		},
		{
			name:        "target cursor - do not exclude copilot rules",
			targetAgent: "cursor",
			path:        ".github/copilot-instructions.md",
			wantExclude: false,
		},
		{
			name:        "target cursor - exclude cursor rules (cursor will read its own)",
			targetAgent: "cursor",
			path:        ".cursor/rules/example.md",
			wantExclude: true,
		},
		{
			name:        "target cursor - do not exclude generic rules",
			targetAgent: "cursor",
			path:        ".agents/rules/example.md",
			wantExclude: false,
		},
		{
			name:        "target opencode - do not exclude cursor rules",
			targetAgent: "opencode",
			path:        ".cursor/rules/example.md",
			wantExclude: false,
		},
		{
			name:        "target opencode - exclude opencode rules (opencode will read its own)",
			targetAgent: "opencode",
			path:        ".opencode/agent/rule.md",
			wantExclude: true,
		},
		{
			name:        "no target agent - do not exclude anything",
			targetAgent: "",
			path:        ".cursor/rules/example.md",
			wantExclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ta TargetAgent
			if tt.targetAgent != "" {
				if err := ta.Set(tt.targetAgent); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			// Normalize the path for testing
			normalizedPath := filepath.FromSlash(tt.path)

			if got := ta.ShouldExcludePath(normalizedPath); got != tt.wantExclude {
				t.Errorf("ShouldExcludePath(%q) = %v, want %v", tt.path, got, tt.wantExclude)
			}
		})
	}
}

func TestTargetAgent_String(t *testing.T) {
	var ta TargetAgent
	ta.Set("cursor")

	str := ta.String()
	if str != "cursor" {
		t.Errorf("String() = %q, want %q", str, "cursor")
	}

	// Test nil agent
	var emptyTA TargetAgent
	if emptyTA.String() != "" {
		t.Errorf("String() on empty agent = %q, want empty string", emptyTA.String())
	}
}

func TestAllAgents(t *testing.T) {
	agents := AllAgents()

	if len(agents) != 8 {
		t.Errorf("AllAgents() returned %d agents, want 8", len(agents))
	}

	// Verify all expected agents are present
	expected := map[Agent]bool{
		AgentCursor:   true,
		AgentOpenCode: true,
		AgentCopilot:  true,
		AgentClaude:   true,
		AgentGemini:   true,
		AgentAugment:  true,
		AgentWindsurf: true,
		AgentCodex:    true,
	}

	for _, agent := range agents {
		if !expected[agent] {
			t.Errorf("AllAgents() returned unexpected agent: %v", agent)
		}
	}
}
