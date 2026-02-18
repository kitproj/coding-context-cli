package markdown

import (
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
	"gopkg.in/yaml.v3"
)

func TestRuleFrontMatter_Marshal(t *testing.T) {
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
					URN:         mustParseURN("urn:agents:rule:standard"),
					Name:        "Standard Rule",
					Description: "This is a standard rule with metadata",
				},
			},
			want: "{}\n",
		},
		{
			name: "rule with task_names",
			rule: RuleFrontMatter{
				TaskNames: []string{"implement-feature"},
				Languages: []string{"go"},
			},
			want: "task_names:\n    - implement-feature\nlanguages:\n    - go\n",
		},
		{
			name: "rule with multiple task_names",
			rule: RuleFrontMatter{
				TaskNames: []string{"fix-bug", "implement-feature"},
				Languages: []string{"go"},
				Agent:     "cursor",
			},
			want: "task_names:\n    - fix-bug\n    - implement-feature\nlanguages:\n    - go\nagent: cursor\n",
		},
		{
			name: "rule with all fields",
			rule: RuleFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					URN:         mustParseURN("urn:agents:rule:all-fields"),
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
			want: "task_names:\n    - test-task\nlanguages:\n    - go\n    - python\nagent: copilot\nmcp_server:\n    type: stdio\n    command: database-server\n    args:\n        - --port\n        - \"5432\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
					URN:         mustParseURN("urn:agents:rule:named"),
					Name:        "Named Rule",
					Description: "A rule with standard fields",
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
				Agent:     "",
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
			var got RuleFrontMatter
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// Compare fields individually
			if !urnEqual(got.URN, tt.want.URN) {
				t.Errorf("URN = %q, want %q", urnString(got.URN), urnString(tt.want.URN))
			}
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
