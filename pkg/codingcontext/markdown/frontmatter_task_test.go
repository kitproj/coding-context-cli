package markdown

import (
	"fmt"
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
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
			},
			want: "task_name: test-task\n",
		},
		{
			name: "task with all fields",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "full-task"},
				},
				Agent:      "cursor",
				Languages:  []string{"go"},
				Model:      "gpt-4",
				SingleShot: true,
				Timeout:    "10m",
				Resume:     false,
				Selectors: map[string]any{
					"stage": "implementation",
				},
			},
			want: `task_name: full-task
agent: cursor
languages:
- go
model: gpt-4
single_shot: true
timeout: 10m
selectors:
  stage: implementation
`,
		},
		{
			name: "task with multiple languages",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "polyglot-task"},
				},
				Languages: []string{"go", "python", "javascript"},
			},
			want: `task_name: polyglot-task
languages:
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
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
			},
		},
		{
			name: "task with single language",
			yaml: `task_name: test-task
languages:
  - go
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				Languages: []string{"go"},
			},
		},
		{
			name: "task with multiple languages",
			yaml: `task_name: test-task
languages:
  - go
  - python
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				Languages: []string{"go", "python"},
			},
		},
		{
			name: "full task",
			yaml: `task_name: full-task
agent: cursor
languages:
  - go
model: gpt-4
single_shot: true
timeout: 10m
selectors:
  stage: implementation
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "full-task"},
				},
				Agent:      "cursor",
				Languages:  []string{"go"},
				Model:      "gpt-4",
				SingleShot: true,
				Timeout:    "10m",
				Selectors: map[string]any{
					"stage": "implementation",
				},
			},
		},
		{
			name: "task with selectors and custom fields",
			yaml: `---
single_shot: true
collect_and_push: false
selectors:
    tool: [""]
---

Say hello.
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{
						"single_shot":      true,
						"collect_and_push": false,
						"selectors":        map[string]any{"tool": []any{""}},
					},
				},
				SingleShot: true,
				Selectors: map[string]any{
					"tool": []any{""},
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
			gotTaskName, _ := got.Content["task_name"].(string)
			wantTaskName, _ := tt.want.Content["task_name"].(string)
			if gotTaskName != wantTaskName {
				t.Errorf("TaskName = %q, want %q", gotTaskName, wantTaskName)
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

			// Check custom Content fields if present in want
			for key, wantVal := range tt.want.Content {
				if key == "task_name" {
					continue // Already checked above
				}
				gotVal, exists := got.Content[key]
				if !exists {
					t.Errorf("Content[%q] missing, want %v", key, wantVal)
				} else if fmt.Sprintf("%v", gotVal) != fmt.Sprintf("%v", wantVal) {
					t.Errorf("Content[%q] = %v, want %v", key, gotVal, wantVal)
				}
			}

			// Check Selectors if present in want
			if len(tt.want.Selectors) > 0 {
				if len(got.Selectors) != len(tt.want.Selectors) {
					t.Errorf("Selectors length = %d, want %d", len(got.Selectors), len(tt.want.Selectors))
				}
				for key, wantVal := range tt.want.Selectors {
					gotVal, exists := got.Selectors[key]
					if !exists {
						t.Errorf("Selectors[%q] missing, want %v", key, wantVal)
					} else if fmt.Sprintf("%v", gotVal) != fmt.Sprintf("%v", wantVal) {
						t.Errorf("Selectors[%q] = %v, want %v", key, gotVal, wantVal)
					}
				}
			}
		})
	}
}
