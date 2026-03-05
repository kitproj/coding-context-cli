package markdown

import (
	"encoding/json"
	"testing"

	yaml "github.com/goccy/go-yaml"
)

func TestTaskFrontMatter_Marshal(t *testing.T) {
	t.Parallel()

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
			name: "task with standard id, name, description",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Standard Test Task",
					Description: "This is a test task with standard fields",
					Content:     map[string]any{"task_name": "standard-task"},
				},
			},
			want: "name: Standard Test Task\ndescription: This is a test task with standard fields\ntask_name: standard-task\n",
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
			want: "name: Full Task\ndescription: A task with all fields\ntask_name: full-task\nagent: cursor\n" +
				"languages:\n- go\nmodel: gpt-4\nsingle_shot: true\ntimeout: 10m\nselectors:\n  stage: implementation\n",
		},
		{
			name: "task with multiple languages",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "polyglot-task"},
				},
				Languages: []string{"go", "python", "javascript"},
			},
			want: "task_name: polyglot-task\nlanguages:\n- go\n- python\n- javascript\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

func taskFrontMatterUnmarshalCases() []struct {
	name    string
	yaml    string
	want    TaskFrontMatter
	wantErr bool
} {
	return []struct {
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
			name: "task with include_unmatched false",
			yaml: `task_name: test-task
include_unmatched: false
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				IncludeUnmatched: func() *bool { b := false; return &b }(),
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
}

func TestTaskFrontMatter_Unmarshal(t *testing.T) {
	t.Parallel()

	for _, tt := range taskFrontMatterUnmarshalCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got TaskFrontMatter

			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			assertTaskFrontMatter(t, got, tt.want, err, tt.wantErr)
		})
	}
}

func assertTaskFrontMatter(t *testing.T, got, want TaskFrontMatter, err error, wantErr bool) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Fatalf("Unmarshal() error = %v, wantErr %v", err, wantErr)
	}

	if err != nil {
		return
	}

	gotTaskName, _ := got.Content["task_name"].(string)
	wantTaskName, _ := want.Content["task_name"].(string)

	if gotTaskName != wantTaskName {
		t.Errorf("TaskName = %q, want %q", gotTaskName, wantTaskName)
	}

	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.Description != want.Description {
		t.Errorf("Description = %q, want %q", got.Description, want.Description)
	}

	if got.Agent != want.Agent {
		t.Errorf("Agent = %q, want %q", got.Agent, want.Agent)
	}

	if got.Model != want.Model {
		t.Errorf("Model = %q, want %q", got.Model, want.Model)
	}

	if got.SingleShot != want.SingleShot {
		t.Errorf("SingleShot = %v, want %v", got.SingleShot, want.SingleShot)
	}

	if got.Timeout != want.Timeout {
		t.Errorf("Timeout = %q, want %q", got.Timeout, want.Timeout)
	}

	switch {
	case got.IncludeUnmatched == nil && want.IncludeUnmatched == nil:
		// both unset — ok
	case got.IncludeUnmatched == nil || want.IncludeUnmatched == nil:
		t.Errorf("IncludeUnmatched = %v, want %v", got.IncludeUnmatched, want.IncludeUnmatched)
	case *got.IncludeUnmatched != *want.IncludeUnmatched:
		t.Errorf("IncludeUnmatched = %v, want %v", *got.IncludeUnmatched, *want.IncludeUnmatched)
	}
}

func validateTaskEmptyJSON(t *testing.T, fm TaskFrontMatter) {
	t.Helper()

	if fm.Agent != "" {
		t.Errorf("Agent = %q, want empty", fm.Agent)
	}

	if fm.Content == nil {
		t.Error("Content should be non-nil empty map for {}")
	}
}

func validateTaskStandardFields(t *testing.T, fm TaskFrontMatter) {
	t.Helper()

	if fm.Agent != "cursor" {
		t.Errorf("Agent = %q, want cursor", fm.Agent)
	}

	if len(fm.Languages) != 2 || fm.Languages[0] != "go" || fm.Languages[1] != "python" {
		t.Errorf("Languages = %v, want [go python]", fm.Languages)
	}

	if fm.Model != "gpt-4" {
		t.Errorf("Model = %q, want gpt-4", fm.Model)
	}

	if !fm.SingleShot {
		t.Error("SingleShot should be true")
	}
}

func validateTaskExtraFields(t *testing.T, fm TaskFrontMatter) {
	t.Helper()

	if fm.Agent != "cursor" {
		t.Errorf("Agent = %q, want cursor", fm.Agent)
	}

	if fm.Content == nil {
		t.Fatal("Content should not be nil")
	}

	if v, ok := fm.Content["custom-field"]; !ok || v != "custom-value" {
		t.Errorf("Content[custom-field] = %v, want custom-value", v)
	}

	if _, ok := fm.Content["agent"]; !ok {
		t.Error("Content should also contain typed fields like agent")
	}
}

//nolint:dupl // table-driven test structure mirrors RuleFrontMatter test
//nolint:dupl // Table-driven test structure is similar to RuleFrontMatter but uses different types.
func TestTaskFrontMatter_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, fm TaskFrontMatter)
	}{
		{name: "empty JSON", input: `{}`, validate: validateTaskEmptyJSON},
		{
			name:     "standard typed fields",
			input:    `{"agent": "cursor", "languages": ["go", "python"], "model": "gpt-4", "single_shot": true}`,
			validate: validateTaskStandardFields,
		},
		{
			name:     "extra fields populate Content map",
			input:    `{"agent": "cursor", "custom-field": "custom-value", "priority": 42}`,
			validate: validateTaskExtraFields,
		},
		{name: "invalid JSON returns error", input: `{invalid json}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fm TaskFrontMatter

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
