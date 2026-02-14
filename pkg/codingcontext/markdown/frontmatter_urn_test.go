package markdown

import (
	"encoding/json"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestURN_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantURN string
		wantErr bool
	}{
		{
			name:    "empty URN (optional)",
			yaml:    "urn: \"\"\n",
			wantURN: "",
			wantErr: false,
		},
		{
			name:    "valid simple URN",
			yaml:    "urn: urn:example:task-123\n",
			wantURN: "urn:example:task-123",
			wantErr: false,
		},
		{
			name:    "valid URN with namespace",
			yaml:    "urn: urn:namespace:resource\n",
			wantURN: "urn:namespace:resource",
			wantErr: false,
		},
		{
			name:    "valid URN with ISBN",
			yaml:    "urn: urn:isbn:0451450523\n",
			wantURN: "urn:isbn:0451450523",
			wantErr: false,
		},
		{
			name:    "valid URN with IETF RFC",
			yaml:    "urn: urn:ietf:rfc:2648\n",
			wantURN: "urn:ietf:rfc:2648",
			wantErr: false,
		},
		{
			name:    "valid URN with complex NSS",
			yaml:    "urn: urn:example:a:b:c:d\n",
			wantURN: "urn:example:a:b:c:d",
			wantErr: false,
		},
		{
			name:    "invalid - not a URN",
			yaml:    "urn: not-a-urn\n",
			wantErr: true,
		},
		{
			name:    "invalid - missing NID and NSS",
			yaml:    "urn: \"urn:\"\n",
			wantErr: true,
		},
		{
			name:    "invalid - missing NSS",
			yaml:    "urn: \"urn:example:\"\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				URN URN `yaml:"urn,omitempty"`
			}
			var got testStruct
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.URN.String() != tt.wantURN {
				t.Errorf("UnmarshalYAML() URN = %q, want %q", got.URN.String(), tt.wantURN)
			}
		})
	}
}

func TestTaskFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantURN string
		wantErr bool
	}{
		{
			name: "task with valid URN",
			yaml: `task_name: test-task
id: task-123
urn: urn:example:task-123
`,
			wantURN: "urn:example:task-123",
			wantErr: false,
		},
		{
			name: "task without URN",
			yaml: `task_name: test-task
id: task-456
`,
			wantURN: "",
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

			if got.URN.String() != tt.wantURN {
				t.Errorf("URN = %q, want %q", got.URN.String(), tt.wantURN)
			}
		})
	}
}

func TestCommandFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantURN string
		wantErr bool
	}{
		{
			name: "command with valid URN",
			yaml: `id: cmd-123
urn: urn:example:command-123
`,
			wantURN: "urn:example:command-123",
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

			if got.URN.String() != tt.wantURN {
				t.Errorf("URN = %q, want %q", got.URN.String(), tt.wantURN)
			}
		})
	}
}

func TestRuleFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantURN string
		wantErr bool
	}{
		{
			name: "rule with valid URN",
			yaml: `id: rule-123
urn: urn:example:rule-123
`,
			wantURN: "urn:example:rule-123",
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

			if got.URN.String() != tt.wantURN {
				t.Errorf("URN = %q, want %q", got.URN.String(), tt.wantURN)
			}
		})
	}
}

func TestSkillFrontMatter_URN_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantURN string
		wantErr bool
	}{
		{
			name: "skill with valid URN",
			yaml: `name: test-skill
description: A test skill
urn: urn:example:skill-123
`,
			wantURN: "urn:example:skill-123",
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

			if got.URN.String() != tt.wantURN {
				t.Errorf("URN = %q, want %q", got.URN.String(), tt.wantURN)
			}
		})
	}
}

func TestTaskFrontMatter_URN_JSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantURN string
		wantErr bool
	}{
		{
			name: "valid URN via JSON",
			json: `{
"id": "task-123",
"urn": "urn:example:task-123"
}`,
			wantURN: "urn:example:task-123",
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
				return
			}
			if !tt.wantErr && got.URN.String() != tt.wantURN {
				t.Errorf("JSON Unmarshal() URN = %q, want %q", got.URN.String(), tt.wantURN)
			}
		})
	}
}
