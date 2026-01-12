package markdown

import (
	"testing"

	"github.com/goccy/go-yaml"
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
				ID:          "cmd-123",
				Name:        "Standard Command",
				Description: "This is a standard command with metadata",
			},
			want: `id: cmd-123
name: Standard Command
description: This is a standard command with metadata
`,
		},
		{
			name: "command with expand false",
			command: CommandFrontMatter{
				ID:          "cmd-456",
				Name:        "No Expand Command",
				Description: "Command with expansion disabled",
				ExpandParams: func() *bool {
					b := false
					return &b
				}(),
			},
			want: `id: cmd-456
name: No Expand Command
description: Command with expansion disabled
expand: false
`,
		},
		{
			name: "command with selectors",
			command: CommandFrontMatter{
				ID:          "cmd-789",
				Name:        "Selector Command",
				Description: "Command with selectors",
				Selectors: map[string]any{
					"database": "postgres",
					"feature":  "auth",
				},
			},
			want: `id: cmd-789
name: Selector Command
description: Command with selectors
selectors:
  database: postgres
  feature: auth
`,
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
			yaml: `id: cmd-abc
name: Named Command
description: A command with standard fields
`,
			want: CommandFrontMatter{
				ID:          "cmd-abc",
				Name:        "Named Command",
				Description: "A command with standard fields",
			},
		},
		{
			name: "command with expand false",
			yaml: `id: cmd-def
name: No Expand
description: No expansion
expand: false
`,
			want: CommandFrontMatter{
				ID:          "cmd-def",
				Name:        "No Expand",
				Description: "No expansion",
				ExpandParams: func() *bool {
					b := false
					return &b
				}(),
			},
		},
		{
			name: "command with selectors",
			yaml: `id: cmd-ghi
name: Selector Command
description: Has selectors
selectors:
  database: postgres
  feature: auth
`,
			want: CommandFrontMatter{
				ID:          "cmd-ghi",
				Name:        "Selector Command",
				Description: "Has selectors",
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
			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if (got.ExpandParams == nil) != (tt.want.ExpandParams == nil) {
				t.Errorf("ExpandParams nil mismatch: got %v, want %v", got.ExpandParams == nil, tt.want.ExpandParams == nil)
			} else if got.ExpandParams != nil && tt.want.ExpandParams != nil {
				if *got.ExpandParams != *tt.want.ExpandParams {
					t.Errorf("ExpandParams = %v, want %v", *got.ExpandParams, *tt.want.ExpandParams)
				}
			}
		})
	}
}
