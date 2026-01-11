package markdown

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
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				TaskName: "test-task",
			},
			want: "task_name: test-task\n",
		},
		{
			name: "task with all fields",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "full-task"},
				},
				TaskName:   "full-task",
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
				TaskName:  "polyglot-task",
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
				TaskName: "test-task",
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
				TaskName:  "test-task",
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
				TaskName:  "test-task",
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
				TaskName:   "full-task",
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
				t.Errorf("Content.TaskName = %q, want %q", gotTaskName, wantTaskName)
			}
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

func TestCommandFrontMatter_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    CommandFrontMatter
		wantErr bool
	}{
		{
			name: "minimal command",
			yaml: "command_name: test-command\n",
			want: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"command_name": "test-command"},
				},
				CommandName: "test-command",
			},
		},
		{
			name: "command with expand false",
			yaml: `command_name: no-expand-command
expand: false
`,
			want: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"command_name": "no-expand-command"},
				},
				CommandName:  "no-expand-command",
				ExpandParams: boolPtr(false),
			},
		},
		{
			name: "command with selectors",
			yaml: `command_name: db-command
selectors:
  database: postgres
`,
			want: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"command_name": "db-command"},
				},
				CommandName: "db-command",
				Selectors: map[string]any{
					"database": "postgres",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CommandFrontMatter
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if got.CommandName != tt.want.CommandName {
				t.Errorf("CommandName = %q, want %q", got.CommandName, tt.want.CommandName)
			}
			if (got.ExpandParams == nil) != (tt.want.ExpandParams == nil) {
				t.Errorf("ExpandParams presence mismatch: got %v, want %v", got.ExpandParams, tt.want.ExpandParams)
			} else if got.ExpandParams != nil && *got.ExpandParams != *tt.want.ExpandParams {
				t.Errorf("ExpandParams = %v, want %v", *got.ExpandParams, *tt.want.ExpandParams)
			}
		})
	}
}

func TestSkillFrontMatter_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    SkillFrontMatter
		wantErr bool
	}{
		{
			name: "skill with name field",
			yaml: `name: data-analysis
description: Analyze datasets
`,
			want: SkillFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"name": "data-analysis"},
				},
				Name:        "data-analysis",
				Description: "Analyze datasets",
			},
		},
		{
			name: "skill with skill_name alias",
			yaml: `skill_name: data-analysis
name: data-analysis
description: Analyze datasets
`,
			want: SkillFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"name": "data-analysis"},
				},
				Name:        "data-analysis",
				SkillName:   "data-analysis",
				Description: "Analyze datasets",
			},
		},
		{
			name: "skill with metadata",
			yaml: `name: pdf-processor
description: Process PDF files
license: MIT
compatibility: Python 3.8+
`,
			want: SkillFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"name": "pdf-processor"},
				},
				Name:          "pdf-processor",
				Description:   "Process PDF files",
				License:       "MIT",
				Compatibility: "Python 3.8+",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SkillFrontMatter
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.SkillName != tt.want.SkillName {
				t.Errorf("SkillName = %q, want %q", got.SkillName, tt.want.SkillName)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.License != tt.want.License {
				t.Errorf("License = %q, want %q", got.License, tt.want.License)
			}
			if got.Compatibility != tt.want.Compatibility {
				t.Errorf("Compatibility = %q, want %q", got.Compatibility, tt.want.Compatibility)
			}
		})
	}
}

// boolPtr returns a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}
