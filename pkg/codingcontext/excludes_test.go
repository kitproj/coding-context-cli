package codingcontext

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIExcludes_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantCLI string
		wantErr bool
	}{
		{
			name:    "valid CLI name - cursor",
			value:   "cursor",
			wantCLI: "cursor",
			wantErr: false,
		},
		{
			name:    "valid CLI name - opencode",
			value:   "opencode",
			wantCLI: "opencode",
			wantErr: false,
		},
		{
			name:    "valid CLI name - copilot",
			value:   "copilot",
			wantCLI: "copilot",
			wantErr: false,
		},
		{
			name:    "valid CLI name with spaces",
			value:   "  cursor  ",
			wantCLI: "cursor",
			wantErr: false,
		},
		{
			name:    "uppercase should be normalized",
			value:   "CURSOR",
			wantCLI: "cursor",
			wantErr: false,
		},
		{
			name:    "mixed case should be normalized",
			value:   "OpenCode",
			wantCLI: "opencode",
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := make(CLIExcludes)
			err := e.Set(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !e[tt.wantCLI] {
					t.Errorf("Set() did not add CLI %q to exclusions", tt.wantCLI)
				}
			}
		})
	}
}

func TestCLIExcludes_SetMultiple(t *testing.T) {
	e := make(CLIExcludes)
	if err := e.Set("cursor"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := e.Set("opencode"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if len(e) != 2 {
		t.Errorf("Set() resulted in %d exclusions, want 2", len(e))
	}
	if !e["cursor"] {
		t.Error("cursor should be excluded")
	}
	if !e["opencode"] {
		t.Error("opencode should be excluded")
	}
}

func TestCLIExcludes_ShouldExcludePath(t *testing.T) {
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
			name:        "exclude opencode command",
			excludes:    []string{"opencode"},
			path:        ".opencode/command/task.md",
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
			name:        "exclude claude file",
			excludes:    []string{"claude"},
			path:        "CLAUDE.md",
			wantExclude: true,
		},
		{
			name:        "exclude claude local file",
			excludes:    []string{"claude"},
			path:        "CLAUDE.local.md",
			wantExclude: true,
		},
		{
			name:        "exclude claude directory",
			excludes:    []string{"claude"},
			path:        ".claude/CLAUDE.md",
			wantExclude: true,
		},
		{
			name:        "exclude gemini file",
			excludes:    []string{"gemini"},
			path:        "GEMINI.md",
			wantExclude: true,
		},
		{
			name:        "exclude gemini styleguide",
			excludes:    []string{"gemini"},
			path:        ".gemini/styleguide.md",
			wantExclude: true,
		},
		{
			name:        "exclude augment rules",
			excludes:    []string{"augment"},
			path:        ".augment/rules/example.md",
			wantExclude: true,
		},
		{
			name:        "exclude augment guidelines",
			excludes:    []string{"augment"},
			path:        ".augment/guidelines.md",
			wantExclude: true,
		},
		{
			name:        "exclude windsurf rules",
			excludes:    []string{"windsurf"},
			path:        ".windsurf/rules/example.md",
			wantExclude: true,
		},
		{
			name:        "exclude windsurfrules file",
			excludes:    []string{"windsurf"},
			path:        ".windsurfrules",
			wantExclude: true,
		},
		{
			name:        "exclude codex directory",
			excludes:    []string{"codex"},
			path:        ".codex/AGENTS.md",
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
			e := make(CLIExcludes)
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

func TestCLIExcludes_String(t *testing.T) {
	e := make(CLIExcludes)
	e.Set("cursor")
	e.Set("opencode")

	str := e.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// Should contain both CLI names (order might vary)
	if !contains(str, "cursor") || !contains(str, "opencode") {
		t.Errorf("String() = %q, want to contain both 'cursor' and 'opencode'", str)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
