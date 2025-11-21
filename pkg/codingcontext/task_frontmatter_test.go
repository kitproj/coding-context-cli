package codingcontext

import (
	"testing"

	"github.com/goccy/go-yaml"
)

func TestTaskFrontMatter_Marshal(t *testing.T) {
	tests := []struct {
		name string
		task TaskFrontMatter
		want string
	}{
		{
			name: "minimal task",
			task: TaskFrontMatter{
				TaskName: "test-task",
			},
			want: "task_name: test-task\n",
		},
		{
			name: "task with all fields",
			task: TaskFrontMatter{
				TaskName:   "full-task",
				Agent:      "cursor",
				Language:   "go",
				Model:      "gpt-4",
				SingleShot: true,
				Timeout:    "10m",
				MCPServers: []string{"filesystem", "git"},
				Resume:     false,
				Selectors: map[string]any{
					"stage": "implementation",
				},
			},
			want: `task_name: full-task
agent: cursor
language: go
model: gpt-4
single_shot: true
timeout: 10m
mcp_servers:
- filesystem
- git
selectors:
  stage: implementation
`,
		},
		{
			name: "task with language array",
			task: TaskFrontMatter{
				TaskName: "polyglot-task",
				Language: []string{"go", "python", "javascript"},
			},
			want: `task_name: polyglot-task
language:
- go
- python
- javascript
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yaml.Marshal(&tt.task)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestTaskFrontMatter_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    TaskFrontMatter
		wantErr bool
	}{
		{
			name: "minimal task",
			yaml: "task_name: test-task\n",
			want: TaskFrontMatter{
				TaskName: "test-task",
			},
		},
		{
			name: "task with string language",
			yaml: `task_name: test-task
language: go
`,
			want: TaskFrontMatter{
				TaskName: "test-task",
				Language: "go",
			},
		},
		{
			name: "task with language array",
			yaml: `task_name: test-task
language:
  - go
  - python
`,
			want: TaskFrontMatter{
				TaskName: "test-task",
				Language: []any{"go", "python"},
			},
		},
		{
			name: "full task",
			yaml: `task_name: full-task
agent: cursor
language: go
model: gpt-4
single_shot: true
timeout: 10m
mcp_servers:
  - filesystem
  - git
selectors:
  stage: implementation
`,
			want: TaskFrontMatter{
				TaskName:   "full-task",
				Agent:      "cursor",
				Language:   "go",
				Model:      "gpt-4",
				SingleShot: true,
				Timeout:    "10m",
				MCPServers: []string{"filesystem", "git"},
				Selectors: map[string]any{
					"stage": "implementation",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TaskFrontMatter
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// Compare fields individually for better error messages
			if got.TaskName != tt.want.TaskName {
				t.Errorf("TaskName = %q, want %q", got.TaskName, tt.want.TaskName)
			}
			if got.Agent != tt.want.Agent {
				t.Errorf("Agent = %q, want %q", got.Agent, tt.want.Agent)
			}
			if got.Model != tt.want.Model {
				t.Errorf("Model = %q, want %q", got.Model, tt.want.Model)
			}
			if got.SingleShot != tt.want.SingleShot {
				t.Errorf("SingleShot = %v, want %v", got.SingleShot, tt.want.SingleShot)
			}
			if got.Timeout != tt.want.Timeout {
				t.Errorf("Timeout = %q, want %q", got.Timeout, tt.want.Timeout)
			}
		})
	}
}
