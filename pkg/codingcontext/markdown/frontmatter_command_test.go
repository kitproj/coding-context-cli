package markdown

import (
	"encoding/json"
	"testing"

	yaml "github.com/goccy/go-yaml"
)

func TestCommandFrontMatter_Marshal(t *testing.T) {
	t.Parallel()

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
			want: "name: Standard Command\ndescription: This is a standard command with metadata\n",
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
			want: "name: No Expand Command\ndescription: Command with expansion disabled\nexpand: false\n",
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
			want: "name: Selector Command\ndescription: Command with selectors\nselectors:\n  database: postgres\n  " +
				"feature: auth\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

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
			t.Parallel()

			var got CommandFrontMatter

			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			assertCommandFrontMatter(t, got, tt.want, err, tt.wantErr)
		})
	}
}

func assertCommandFrontMatter(t *testing.T, got, want CommandFrontMatter, err error, wantErr bool) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Fatalf("Unmarshal() error = %v, wantErr %v", err, wantErr)
	}

	if err != nil {
		return
	}

	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.Description != want.Description {
		t.Errorf("Description = %q, want %q", got.Description, want.Description)
	}

	assertExpandParams(t, got.ExpandParams, want.ExpandParams)
}

func assertExpandParams(t *testing.T, got, want *bool) {
	t.Helper()

	if want == nil {
		return
	}

	if (got == nil) != (want == nil) {
		t.Errorf("ExpandParams nil mismatch: got %v, want %v", got == nil, want == nil)

		return
	}

	if got != nil && *got != *want {
		t.Errorf("ExpandParams = %v, want %v", *got, *want)
	}
}

func TestCommandFrontMatter_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	expandFalse := false
	expandTrue := true

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, fm CommandFrontMatter)
	}{
		{
			name:  "empty JSON",
			input: `{}`,
			validate: func(t *testing.T, fm CommandFrontMatter) {
				t.Helper()

				if fm.Content == nil {
					t.Error("Content should be non-nil empty map for {}")
				}

				if fm.ExpandParams != nil {
					t.Errorf("ExpandParams should be nil, got %v", *fm.ExpandParams)
				}
			},
		},
		{
			name:  "expand false",
			input: `{"expand": false, "name": "no-expand-cmd"}`,
			validate: func(t *testing.T, fm CommandFrontMatter) {
				t.Helper()
				assertExpandParams(t, fm.ExpandParams, &expandFalse)

				if fm.Name != "no-expand-cmd" {
					t.Errorf("Name = %q, want no-expand-cmd", fm.Name)
				}
			},
		},
		{
			name:  "expand true with selectors",
			input: `{"expand": true, "selectors": {"env": "prod"}}`,
			validate: func(t *testing.T, fm CommandFrontMatter) {
				t.Helper()
				assertExpandParams(t, fm.ExpandParams, &expandTrue)

				if v, ok := fm.Selectors["env"]; !ok || v != "prod" {
					t.Errorf("Selectors[env] = %v, want prod", v)
				}
			},
		},
		{
			name:  "extra fields populate Content map",
			input: `{"name": "my-cmd", "extra-field": "extra-value"}`,
			validate: func(t *testing.T, fm CommandFrontMatter) {
				t.Helper()

				if fm.Name != "my-cmd" {
					t.Errorf("Name = %q, want my-cmd", fm.Name)
				}

				if fm.Content == nil {
					t.Fatal("Content should not be nil")
				}

				if v, ok := fm.Content["extra-field"]; !ok || v != "extra-value" {
					t.Errorf("Content[extra-field] = %v, want extra-value", v)
				}
			},
		},
		{
			name:    "invalid JSON returns error",
			input:   `{bad`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fm CommandFrontMatter

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
