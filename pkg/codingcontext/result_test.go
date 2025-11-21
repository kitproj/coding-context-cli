package codingcontext

import (
	"testing"
)

func TestResult_ParseTaskFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter FrontMatter
		target      any
		wantErr     bool
		validate    func(t *testing.T, target any)
	}{
		{
			name: "parse into struct with basic fields",
			frontmatter: FrontMatter{
				"task_name": "fix-bug",
				"resume":    false,
				"priority":  "high",
			},
			target: &struct {
				TaskName string `yaml:"task_name"`
				Resume   bool   `yaml:"resume"`
				Priority string `yaml:"priority"`
			}{},
			wantErr: false,
			validate: func(t *testing.T, target any) {
				meta := target.(*struct {
					TaskName string `yaml:"task_name"`
					Resume   bool   `yaml:"resume"`
					Priority string `yaml:"priority"`
				})
				if meta.TaskName != "fix-bug" {
					t.Errorf("TaskName = %q, want %q", meta.TaskName, "fix-bug")
				}
				if meta.Resume != false {
					t.Errorf("Resume = %v, want %v", meta.Resume, false)
				}
				if meta.Priority != "high" {
					t.Errorf("Priority = %q, want %q", meta.Priority, "high")
				}
			},
		},
		{
			name: "parse with nested selectors",
			frontmatter: FrontMatter{
				"task_name": "implement-feature",
				"selectors": map[string]any{
					"language": "Go",
					"stage":    "implementation",
				},
			},
			target: &struct {
				TaskName  string         `yaml:"task_name"`
				Selectors map[string]any `yaml:"selectors"`
			}{},
			wantErr: false,
			validate: func(t *testing.T, target any) {
				meta := target.(*struct {
					TaskName  string         `yaml:"task_name"`
					Selectors map[string]any `yaml:"selectors"`
				})
				if meta.TaskName != "implement-feature" {
					t.Errorf("TaskName = %q, want %q", meta.TaskName, "implement-feature")
				}
				if len(meta.Selectors) != 2 {
					t.Errorf("Selectors length = %d, want 2", len(meta.Selectors))
				}
				if meta.Selectors["language"] != "Go" {
					t.Errorf("Selectors[language] = %v, want Go", meta.Selectors["language"])
				}
			},
		},
		{
			name: "parse with array values",
			frontmatter: FrontMatter{
				"task_name": "test-code",
				"languages": []any{"Go", "Python", "JavaScript"},
			},
			target: &struct {
				TaskName  string   `yaml:"task_name"`
				Languages []string `yaml:"languages"`
			}{},
			wantErr: false,
			validate: func(t *testing.T, target any) {
				meta := target.(*struct {
					TaskName  string   `yaml:"task_name"`
					Languages []string `yaml:"languages"`
				})
				if meta.TaskName != "test-code" {
					t.Errorf("TaskName = %q, want %q", meta.TaskName, "test-code")
				}
				if len(meta.Languages) != 3 {
					t.Errorf("Languages length = %d, want 3", len(meta.Languages))
				}
				if meta.Languages[0] != "Go" {
					t.Errorf("Languages[0] = %q, want Go", meta.Languages[0])
				}
			},
		},
		{
			name: "parse with optional fields",
			frontmatter: FrontMatter{
				"task_name": "deploy",
			},
			target: &struct {
				TaskName    string `yaml:"task_name"`
				Environment string `yaml:"environment"`
				Priority    string `yaml:"priority"`
			}{},
			wantErr: false,
			validate: func(t *testing.T, target any) {
				meta := target.(*struct {
					TaskName    string `yaml:"task_name"`
					Environment string `yaml:"environment"`
					Priority    string `yaml:"priority"`
				})
				if meta.TaskName != "deploy" {
					t.Errorf("TaskName = %q, want %q", meta.TaskName, "deploy")
				}
				// Optional fields should be zero values
				if meta.Environment != "" {
					t.Errorf("Environment = %q, want empty string", meta.Environment)
				}
				if meta.Priority != "" {
					t.Errorf("Priority = %q, want empty string", meta.Priority)
				}
			},
		},
		{
			name:        "nil frontmatter returns error",
			frontmatter: nil,
			target: &struct {
				TaskName string `yaml:"task_name"`
			}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &Result{
				Task: Markdown{
					Path:        "/test/task.md",
					FrontMatter: tt.frontmatter,
					Content:     "Test content",
				},
			}

			err := result.ParseTaskFrontmatter(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTaskFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, tt.target)
			}
		})
	}
}

func TestMarkdown_BootstrapPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "md file",
			path: "/path/to/task.md",
			want: "/path/to/task-bootstrap",
		},
		{
			name: "mdc file",
			path: "/path/to/rule.mdc",
			want: "/path/to/rule-bootstrap",
		},
		{
			name: "empty path",
			path: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Markdown{Path: tt.path}
			got := m.BootstrapPath()
			if got != tt.want {
				t.Errorf("BootstrapPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
