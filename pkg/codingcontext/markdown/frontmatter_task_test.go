package markdown

import (
	"testing"

	"gopkg.in/yaml.v3"
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
			want: "{}\n",
		},
		{
			name: "task with standard id, name, description",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Standard Test Task",
					Description: "This is a test task with standard fields",
					Content:     map[string]any{"task_name": "standard-task"},
				},
			},
			want: "{}\n",
		},
		{
			name: "task with all fields",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Full Task",
					Description: "A task with all fields",
					Content:     map[string]any{"task_name": "full-task"},
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
			want: "agent: cursor\nlanguages:\n    - go\nmodel: gpt-4\nsingle_shot: true\ntimeout: 10m\nselectors:\n    stage: implementation\n",
		},
		{
			name: "task with multiple languages",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "polyglot-task"},
				},
				Languages: []string{"go", "python", "javascript"},
			},
			want: "languages:\n    - go\n    - python\n    - javascript\n",
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
			name: "task with standard id, name, description",
			yaml: `task_name: standard-task
id: urn:agents:task:standard-task
name: Standard Task
description: This is a standard task
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Standard Task",
					Description: "This is a standard task",
					Content:     map[string]any{"task_name": "standard-task", "id": "urn:agents:task:standard-task"},
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
id: urn:agents:task:full-task
name: Full Task
description: A complete task
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
					Name:        "Full Task",
					Description: "A complete task",
					Content: map[string]any{
						"task_name": "full-task",
						"id":        "urn:agents:task:full-task",
					},
				},
				Agent:      "",
				Languages:  []string{"go"},
				Model:      "",
				SingleShot: false,
				Timeout:    "",
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
			gotTaskName, _ := got.Content["task_name"].(string)
			wantTaskName, _ := tt.want.Content["task_name"].(string)
			if gotTaskName != wantTaskName {
				t.Errorf("TaskName = %q, want %q", gotTaskName, wantTaskName)
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
