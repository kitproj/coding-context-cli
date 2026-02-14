package markdown

import (
	"encoding/json"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestValidateURN(t *testing.T) {
	tests := []struct {
		name    string
		urn     string
		wantErr bool
	}{
		{
			name:    "empty URN (optional)",
			urn:     "",
			wantErr: false,
		},
		{
			name:    "valid simple URN",
			urn:     "urn:example:task-123",
			wantErr: false,
		},
		{
			name:    "valid URN with namespace",
			urn:     "urn:namespace:resource",
			wantErr: false,
		},
		{
			name:    "valid URN with ISBN",
			urn:     "urn:isbn:0451450523",
			wantErr: false,
		},
		{
			name:    "valid URN with IETF RFC",
			urn:     "urn:ietf:rfc:2648",
			wantErr: false,
		},
		{
			name:    "valid URN with complex NSS",
			urn:     "urn:example:a:b:c:d",
			wantErr: false,
		},
		{
			name:    "invalid - not a URN",
			urn:     "not-a-urn",
			wantErr: true,
		},
		{
			name:    "invalid - missing NID and NSS",
			urn:     "urn:",
			wantErr: true,
		},
		{
			name:    "invalid - missing NSS",
			urn:     "urn:example:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURN(tt.urn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskFrontMatter_URN_Marshal(t *testing.T) {
	tests := []struct {
		name string
		task TaskFrontMatter
		want string
	}{
		{
			name: "task with URN",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				ID:          "task-123",
				Name:        "Test Task",
				Description: "A test task",
				URN:         "urn:example:task-123",
			},
			want: `task_name: test-task
id: task-123
name: Test Task
description: A test task
urn: urn:example:task-123
`,
		},
		{
			name: "task without URN",
			task: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				ID: "task-456",
			},
			want: `task_name: test-task
id: task-456
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

func TestTaskFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    TaskFrontMatter
		wantErr bool
	}{
		{
			name: "task with valid URN",
			yaml: `task_name: test-task
id: task-123
urn: urn:example:task-123
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				ID:  "task-123",
				URN: "urn:example:task-123",
			},
			wantErr: false,
		},
		{
			name: "task without URN",
			yaml: `task_name: test-task
id: task-456
`,
			want: TaskFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Content: map[string]any{"task_name": "test-task"},
				},
				ID: "task-456",
			},
			wantErr: false,
		},
		{
			name: "task with invalid URN",
			yaml: `task_name: test-task
id: task-789
urn: not-a-valid-urn
`,
			wantErr: true,
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

			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.URN != tt.want.URN {
				t.Errorf("URN = %q, want %q", got.URN, tt.want.URN)
			}
		})
	}
}

func TestCommandFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    CommandFrontMatter
		wantErr bool
	}{
		{
			name: "command with valid URN",
			yaml: `id: cmd-123
urn: urn:example:command-123
`,
			want: CommandFrontMatter{
				ID:  "cmd-123",
				URN: "urn:example:command-123",
			},
			wantErr: false,
		},
		{
			name: "command with invalid URN",
			yaml: `id: cmd-456
urn: invalid-urn
`,
			wantErr: true,
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

			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.URN != tt.want.URN {
				t.Errorf("URN = %q, want %q", got.URN, tt.want.URN)
			}
		})
	}
}

func TestRuleFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    RuleFrontMatter
		wantErr bool
	}{
		{
			name: "rule with valid URN",
			yaml: `id: rule-123
urn: urn:example:rule-123
`,
			want: RuleFrontMatter{
				ID:  "rule-123",
				URN: "urn:example:rule-123",
			},
			wantErr: false,
		},
		{
			name: "rule with invalid URN",
			yaml: `id: rule-456
urn: bad-urn
`,
			wantErr: true,
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

			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.URN != tt.want.URN {
				t.Errorf("URN = %q, want %q", got.URN, tt.want.URN)
			}
		})
	}
}

func TestSkillFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    SkillFrontMatter
		wantErr bool
	}{
		{
			name: "skill with valid URN",
			yaml: `name: test-skill
description: A test skill
urn: urn:example:skill-123
`,
			want: SkillFrontMatter{
				Name:        "test-skill",
				Description: "A test skill",
				URN:         "urn:example:skill-123",
			},
			wantErr: false,
		},
		{
			name: "skill with invalid URN",
			yaml: `name: test-skill
description: A test skill
urn: invalid
`,
			wantErr: true,
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
			if got.URN != tt.want.URN {
				t.Errorf("URN = %q, want %q", got.URN, tt.want.URN)
			}
		})
	}
}

func TestTaskFrontMatter_URN_JSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid URN via JSON",
			json: `{
				"id": "task-123",
				"urn": "urn:example:task-123"
			}`,
			wantErr: false,
		},
		{
			name: "invalid URN via JSON",
			json: `{
				"id": "task-456",
				"urn": "not-valid"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TaskFrontMatter
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
