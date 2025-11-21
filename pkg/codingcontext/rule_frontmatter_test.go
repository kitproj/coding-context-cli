package codingcontext

import (
	"testing"

	"github.com/goccy/go-yaml"
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
			name: "rule with string task_name",
			rule: RuleFrontMatter{
				TaskName: "implement-feature",
				Language: "go",
			},
			want: `task_name: implement-feature
language: go
`,
		},
		{
			name: "rule with array task_name",
			rule: RuleFrontMatter{
				TaskName: []string{"fix-bug", "implement-feature"},
				Language: "go",
				Agent:    "cursor",
			},
			want: `task_name:
- fix-bug
- implement-feature
language: go
agent: cursor
`,
		},
		{
			name: "rule with all fields",
			rule: RuleFrontMatter{
				TaskName:   "test-task",
				Language:   []string{"go", "python"},
				Agent:      "copilot",
				MCPServers: []string{"database"},
				RuleName:   "test-rule",
			},
			want: `task_name: test-task
language:
- go
- python
agent: copilot
mcp_servers:
- database
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
			name: "rule with string task_name and language",
			yaml: `task_name: implement-feature
language: go
agent: cursor
`,
			want: RuleFrontMatter{
				TaskName: "implement-feature",
				Language: "go",
				Agent:    "cursor",
			},
		},
		{
			name: "rule with array task_name",
			yaml: `task_name:
  - fix-bug
  - implement-feature
language: go
`,
			want: RuleFrontMatter{
				TaskName: []any{"fix-bug", "implement-feature"},
				Language: "go",
			},
		},
		{
			name: "rule with array language",
			yaml: `language:
  - go
  - python
  - javascript
`,
			want: RuleFrontMatter{
				Language: []any{"go", "python", "javascript"},
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
