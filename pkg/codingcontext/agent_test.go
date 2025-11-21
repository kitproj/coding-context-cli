package codingcontext

import (
	"path/filepath"
	"strings"
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
			name:    "uppercase normalized",
			input:   "CURSOR",
			want:    AgentCursor,
			wantErr: false,
		},
		{
			name:    "mixed case normalized",
			input:   "OpenCode",
			want:    AgentOpenCode,
			wantErr: false,
		},
		{
			name:    "with spaces trimmed",
			input:   "  cursor  ",
			want:    AgentCursor,
			wantErr: false,
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

func TestAgentExcludes_Set(t *testing.T) {
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
			name:      "uppercase normalized",
			value:     "CURSOR",
			wantAgent: AgentCursor,
			wantErr:   false,
		},
		{
			name:      "with spaces",
			value:     "  cursor  ",
			wantAgent: AgentCursor,
			wantErr:   false,
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
			e := make(AgentExcludes)
			err := e.Set(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !e[tt.wantAgent] {
					t.Errorf("Set() did not add agent %v to exclusions", tt.wantAgent)
				}
			}
		})
	}
}

func TestAgentExcludes_SetMultiple(t *testing.T) {
	e := make(AgentExcludes)
	if err := e.Set("cursor"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := e.Set("opencode"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if len(e) != 2 {
		t.Errorf("Set() resulted in %d exclusions, want 2", len(e))
	}
	if !e[AgentCursor] {
		t.Error("cursor should be excluded")
	}
	if !e[AgentOpenCode] {
		t.Error("opencode should be excluded")
	}
}

func TestAgentExcludes_ShouldExcludePath(t *testing.T) {
	tests := []struct {
		name        string
		excludes    []string
		path        string
		wantExclude bool
	}{
		{
			name:        "exclude cursor rules",
			excludes:    []string{"cursor"},
			path:        ".cursor/rules/example.md",
			wantExclude: true,
		},
		{
			name:        "exclude cursorrules file",
			excludes:    []string{"cursor"},
			path:        ".cursorrules",
			wantExclude: true,
		},
		{
			name:        "exclude opencode agent",
			excludes:    []string{"opencode"},
			path:        ".opencode/agent/rule.md",
			wantExclude: true,
		},
		{
			name:        "exclude copilot instructions",
			excludes:    []string{"copilot"},
			path:        ".github/copilot-instructions.md",
			wantExclude: true,
		},
		{
			name:        "exclude copilot agents directory",
			excludes:    []string{"copilot"},
			path:        ".github/agents/rule.md",
			wantExclude: true,
		},
		{
			name:        "do not exclude agents rules when excluding cursor",
			excludes:    []string{"cursor"},
			path:        ".agents/rules/example.md",
			wantExclude: false,
		},
		{
			name:        "do not exclude AGENTS.md when excluding cursor",
			excludes:    []string{"cursor"},
			path:        "AGENTS.md",
			wantExclude: false,
		},
		{
			name:        "exclude with absolute path",
			excludes:    []string{"cursor"},
			path:        "/home/user/project/.cursor/rules/example.md",
			wantExclude: true,
		},
		{
			name:        "multiple exclusions - match first",
			excludes:    []string{"cursor", "opencode"},
			path:        ".cursor/rules/example.md",
			wantExclude: true,
		},
		{
			name:        "multiple exclusions - match second",
			excludes:    []string{"cursor", "opencode"},
			path:        ".opencode/agent/rule.md",
			wantExclude: true,
		},
		{
			name:        "multiple exclusions - match none",
			excludes:    []string{"cursor", "opencode"},
			path:        ".agents/rules/example.md",
			wantExclude: false,
		},
		{
			name:        "no exclusions - do not exclude",
			excludes:    []string{},
			path:        ".cursor/rules/example.md",
			wantExclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := make(AgentExcludes)
			for _, exclude := range tt.excludes {
				if err := e.Set(exclude); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			// Normalize the path for testing
			normalizedPath := filepath.FromSlash(tt.path)

			if got := e.ShouldExcludePath(normalizedPath); got != tt.wantExclude {
				t.Errorf("ShouldExcludePath(%q) = %v, want %v", tt.path, got, tt.wantExclude)
			}
		})
	}
}

func TestAgentExcludes_String(t *testing.T) {
	e := make(AgentExcludes)
	e.Set("cursor")
	e.Set("opencode")

	str := e.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// Should contain both agent names (order might vary)
	if !strings.Contains(str, "cursor") || !strings.Contains(str, "opencode") {
		t.Errorf("String() = %q, want to contain both 'cursor' and 'opencode'", str)
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
