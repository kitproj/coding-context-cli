package markdown

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCommandFrontMatter_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		command CommandFrontMatter
		want    string
	}{
		{
			name:    "minimal command",
			command: CommandFrontMatter{},
			want:    "{}\n",
		},
		{
			name: "command with standard id, name, description",
			command: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Standard Command",
					Description: "This is a standard command with metadata",
				},
			},
			want: "{}\n",
		},
		{
			name: "command with expand false",
			command: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "No Expand Command",
					Description: "Command with expansion disabled",
				},
				ExpandParams: func() *bool {
					b := false
					return &b
				}(),
			},
			want: "expand: false\n",
		},
		{
			name: "command with selectors",
			command: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Selector Command",
					Description: "Command with selectors",
				},
				Selectors: map[string]any{
					"database": "postgres",
					"feature":  "auth",
				},
			},
			want: "selectors:\n    database: postgres\n    feature: auth\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yaml.Marshal(&tt.command)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %q, want %q", string(got), tt.want)
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
			name: "command with standard id, name, description",
			yaml: `id: urn:agents:command:named
name: Named Command
description: A command with standard fields
`,
			want: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Named Command",
					Description: "A command with standard fields",
					Content:     map[string]any{"id": "urn:agents:command:named"},
				},
			},
		},
		{
			name: "command with expand false",
			yaml: `id: urn:agents:command:no-expand
name: No Expand
description: No expansion
expand: false
`,
			want: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "No Expand",
					Description: "No expansion",
					Content:     map[string]any{"id": "urn:agents:command:no-expand"},
				},
				ExpandParams: nil,
			},
		},
		{
			name: "command with selectors",
			yaml: `id: urn:agents:command:selector
name: Selector Command
description: Has selectors
selectors:
  database: postgres
  feature: auth
`,
			want: CommandFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "Selector Command",
					Description: "Has selectors",
					Content:     map[string]any{"id": "urn:agents:command:selector"},
				},
				Selectors: map[string]any{
					"database": "postgres",
					"feature":  "auth",
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

			// Compare fields individually
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if tt.want.ExpandParams != nil {
				if (got.ExpandParams == nil) != (tt.want.ExpandParams == nil) {
					t.Errorf("ExpandParams nil mismatch: got %v, want %v", got.ExpandParams == nil, tt.want.ExpandParams == nil)
				} else if got.ExpandParams != nil && tt.want.ExpandParams != nil {
					if *got.ExpandParams != *tt.want.ExpandParams {
						t.Errorf("ExpandParams = %v, want %v", *got.ExpandParams, *tt.want.ExpandParams)
					}
				}
			}
		})
	}
}
