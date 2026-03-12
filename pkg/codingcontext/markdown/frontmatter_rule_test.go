package markdown

import (
	"encoding/json"
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

const agentCursor = "cursor"

func validateRuleEmptyJSON(t *testing.T, fm RuleFrontMatter) {
	t.Helper()

	if fm.Content == nil {
		t.Error("Content should be non-nil empty map for {}")
	}
}

func validateRuleStandardFields(t *testing.T, fm RuleFrontMatter) {
	t.Helper()

	if len(fm.TaskNames) != 1 || fm.TaskNames[0] != "fix-bug" {
		t.Errorf("TaskNames = %v, want [fix-bug]", fm.TaskNames)
	}

	if len(fm.Languages) != 1 || fm.Languages[0] != "go" {
		t.Errorf("Languages = %v, want [go]", fm.Languages)
	}

	if fm.Agent != agentCursor {
		t.Errorf("Agent = %q, want %s", fm.Agent, agentCursor)
	}

	if fm.Bootstrap == "" {
		t.Error("Bootstrap should be set")
	}
}

func validateRuleExtraFields(t *testing.T, fm RuleFrontMatter) {
	t.Helper()

	if fm.Agent != "copilot" {
		t.Errorf("Agent = %q, want copilot", fm.Agent)
	}

	if fm.Content == nil {
		t.Fatal("Content should not be nil")
	}

	if v, ok := fm.Content["custom-key"]; !ok || v != "custom-val" {
		t.Errorf("Content[custom-key] = %v, want custom-val", v)
	}
}

func TestRuleFrontMatter_Marshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule RuleFrontMatter
		want string
	}{
		{
			name: "minimal rule",
			rule: RuleFrontMatter{},
			want: "{}\n",
		},
		{
			name: "rule with standard id, name, description",
			rule: RuleFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Standard Rule",
					Description: "This is a standard rule with metadata",
				},
			},
			want: "name: Standard Rule\ndescription: This is a standard rule with metadata\n",
		},
		{
			name: "rule with task_names",
			rule: RuleFrontMatter{
				TaskNames: []string{"implement-feature"},
				Languages: []string{"go"},
			},
			want: "task_names:\n- implement-feature\nlanguages:\n- go\n",
		},
		{
			name: "rule with multiple task_names",
			rule: RuleFrontMatter{
				TaskNames: []string{"fix-bug", "implement-feature"},
				Languages: []string{"go"},
				Agent:     "cursor",
			},
			want: "task_names:\n- fix-bug\n- implement-feature\nlanguages:\n- go\nagent: cursor\n",
		},
		{
			name: "rule with all fields",
			rule: RuleFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Complete Rule",
					Description: "A rule with all fields",
				},
				TaskNames: []string{"test-task"},
				Languages: []string{"go", "python"},
				Agent:     "copilot",
				MCPServer: mcp.MCPServerConfig{
					Type:    mcp.TransportTypeStdio,
					Command: "database-server",
					Args:    []string{"--port", "5432"},
				},
			},
			want: "name: Complete Rule\ndescription: A rule with all fields\ntask_names:\n- test-task\n" +
				"languages:\n- go\n- python\nagent: copilot\nmcp_server:\n  type: stdio\n  command: database-server\n" +
				"  args:\n  - --port\n  - \"5432\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(&tt.rule)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("Marshal() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestRuleFrontMatter_Unmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		want    RuleFrontMatter
		wantErr bool
	}{
		{
			name: "rule with standard id, name, description",
			yaml: `id: urn:agents:rule:named
name: Named Rule
description: A rule with standard fields
`,
			want: RuleFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Named Rule",
					Description: "A rule with standard fields",
					Content:     map[string]any{"id": "urn:agents:rule:named"},
				},
			},
		},
		{
			name: "rule with task_names and languages",
			yaml: `task_names:
  - implement-feature
languages:
  - go
agent: cursor
`,
			want: RuleFrontMatter{
				TaskNames: []string{"implement-feature"},
				Languages: []string{"go"},
				Agent:     "cursor",
			},
		},
		{
			name: "rule with multiple task_names",
			yaml: `task_names:
  - fix-bug
  - implement-feature
languages:
  - go
`,
			want: RuleFrontMatter{
				TaskNames: []string{"fix-bug", "implement-feature"},
				Languages: []string{"go"},
			},
		},
		{
			name: "rule with multiple languages",
			yaml: `languages:
  - go
  - python
  - javascript
`,
			want: RuleFrontMatter{
				Languages: []string{"go", "python", "javascript"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got RuleFrontMatter

			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			// Compare fields individually
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}

			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}

			if got.Agent != tt.want.Agent {
				t.Errorf("Agent = %q, want %q", got.Agent, tt.want.Agent)
			}
		})
	}
}

//nolint:dupl // table-driven test structure mirrors TaskFrontMatter test
//nolint:dupl // Table-driven test structure is similar to TaskFrontMatter but uses different types.
func TestRuleFrontMatter_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, fm RuleFrontMatter)
	}{
		{name: "empty JSON", input: `{}`, validate: validateRuleEmptyJSON},
		{
			name:     "standard typed fields",
			input:    `{"task_names": ["fix-bug"], "languages": ["go"], "agent": "cursor", "bootstrap": "#!/bin/bash\necho hi"}`,
			validate: validateRuleStandardFields,
		},
		{
			name:     "extra fields populate Content map",
			input:    `{"agent": "copilot", "custom-key": "custom-val"}`,
			validate: validateRuleExtraFields,
		},
		{name: "invalid JSON returns error", input: `{bad json`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fm RuleFrontMatter

			err := json.Unmarshal([]byte(tt.input), &fm)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, fm)
			}
		})
	}
}
