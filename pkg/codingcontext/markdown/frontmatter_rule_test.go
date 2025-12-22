package markdown

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
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
			name: "rule with task_names",
			rule: RuleFrontMatter{
				TaskNames: []string{"implement-feature"},
				Languages: []string{"go"},
			},
			want: `task_names:
- implement-feature
languages:
- go
`,
		},
		{
			name: "rule with multiple task_names",
			rule: RuleFrontMatter{
				TaskNames: []string{"fix-bug", "implement-feature"},
				Languages: []string{"go"},
				Agent:     "cursor",
			},
			want: `task_names:
- fix-bug
- implement-feature
languages:
- go
agent: cursor
`,
		},
		{
			name: "rule with all fields",
			rule: RuleFrontMatter{
				TaskNames: []string{"test-task"},
				Languages: []string{"go", "python"},
				Agent:     "copilot",
				MCPServer: mcp.MCPServerConfig{
					Type:    mcp.TransportTypeStdio,
					Command: "database-server",
					Args:    []string{"--port", "5432"},
				},
				RuleName: "test-rule",
			},
			want: `task_names:
- test-task
languages:
- go
- python
agent: copilot
mcp_server:
  type: stdio
  command: database-server
  args:
  - --port
  - "5432"
rule_name: test-rule
`,
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
			var got RuleFrontMatter
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// Compare fields individually
			if got.Agent != tt.want.Agent {
				t.Errorf("Agent = %q, want %q", got.Agent, tt.want.Agent)
			}
			if got.RuleName != tt.want.RuleName {
				t.Errorf("RuleName = %q, want %q", got.RuleName, tt.want.RuleName)
			}
		})
	}
}
