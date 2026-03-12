package markdown

import (
	"encoding/json"
	"testing"

	yaml "github.com/goccy/go-yaml"
)

func TestSkillFrontMatter_Marshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		skill SkillFrontMatter
		want  string
	}{
		{
			name:  "minimal skill",
			skill: SkillFrontMatter{},
			want:  "{}\n",
		},
		{
			name: "skill with name and description",
			skill: SkillFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "my-skill",
					Description: "Does something useful",
				},
			},
			want: "name: my-skill\ndescription: Does something useful\n",
		},
		{
			name: "skill with all fields",
			skill: SkillFrontMatter{
				BaseFrontMatter: BaseFrontMatter{
					Name:        "full-skill",
					Description: "A complete skill",
				},
				License:       "MIT",
				Compatibility: "go>=1.21",
				AllowedTools:  "Bash Read Write",
				Metadata: map[string]string{
					"version": "1.0.0",
				},
			},
			want: "name: full-skill\ndescription: A complete skill\nlicense: MIT\ncompatibility: go>=1.21\n" +
				"metadata:\n  version: 1.0.0\nallowed_tools: Bash Read Write\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(&tt.skill)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("Marshal() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func validateSkillEmptyYAML(t *testing.T, fm SkillFrontMatter) {
	t.Helper()

	if fm.Name != "" {
		t.Errorf("Name = %q, want empty", fm.Name)
	}
}

func validateSkillNameDescLicense(t *testing.T, fm SkillFrontMatter) {
	t.Helper()

	if fm.Name != "my-skill" {
		t.Errorf("Name = %q, want my-skill", fm.Name)
	}

	if fm.Description != "Does something useful" {
		t.Errorf("Description = %q, want 'Does something useful'", fm.Description)
	}

	if fm.License != "MIT" {
		t.Errorf("License = %q, want MIT", fm.License)
	}
}

func validateSkillCompatAllowedTools(t *testing.T, fm SkillFrontMatter) {
	t.Helper()

	if fm.Compatibility != "go>=1.21" {
		t.Errorf("Compatibility = %q, want go>=1.21", fm.Compatibility)
	}

	if fm.AllowedTools != "Bash Read" {
		t.Errorf("AllowedTools = %q, want 'Bash Read'", fm.AllowedTools)
	}
}

func validateSkillMetadata(t *testing.T, fm SkillFrontMatter) {
	t.Helper()

	if fm.Metadata["version"] != "2.0" {
		t.Errorf("Metadata[version] = %q, want 2.0", fm.Metadata["version"])
	}

	if fm.Metadata["author"] != "team" {
		t.Errorf("Metadata[author] = %q, want team", fm.Metadata["author"])
	}
}

func validateSkillExtraFields(t *testing.T, fm SkillFrontMatter) {
	t.Helper()

	if fm.Name != "extra-skill" {
		t.Errorf("Name = %q, want extra-skill", fm.Name)
	}

	if fm.Content == nil {
		t.Fatal("Content should not be nil")
	}

	if fm.Content["unknown-field"] == nil {
		t.Error("Content[unknown-field] should be set")
	}
}

func TestSkillFrontMatter_Unmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yamlStr  string
		wantErr  bool
		validate func(t *testing.T, fm SkillFrontMatter)
	}{
		{name: "empty YAML", yamlStr: "{}\n", validate: validateSkillEmptyYAML},
		{
			name:     "skill with name, description, license",
			yamlStr:  "name: my-skill\ndescription: Does something useful\nlicense: MIT\n",
			validate: validateSkillNameDescLicense,
		},
		{
			name:     "skill with compatibility and allowed_tools",
			yamlStr:  "name: compat-skill\ncompatibility: go>=1.21\nallowed_tools: Bash Read\n",
			validate: validateSkillCompatAllowedTools,
		},
		{
			name:     "skill with metadata map",
			yamlStr:  "name: meta-skill\nmetadata:\n  version: \"2.0\"\n  author: team\n",
			validate: validateSkillMetadata,
		},
		{
			name:     "extra fields captured in Content map",
			yamlStr:  "name: extra-skill\nunknown-field: some-value\n",
			validate: validateSkillExtraFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fm SkillFrontMatter

			err := yaml.Unmarshal([]byte(tt.yamlStr), &fm)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, fm)
			}
		})
	}
}

func TestSkillFrontMatter_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, fm SkillFrontMatter)
	}{
		{
			name:  "empty JSON",
			input: `{}`,
			validate: func(t *testing.T, fm SkillFrontMatter) {
				t.Helper()

				if fm.Content == nil {
					t.Error("Content should be non-nil empty map for {}")
				}
			},
		},
		{
			name:  "typed fields parsed correctly",
			input: `{"name": "my-skill", "license": "MIT", "compatibility": "go>=1.21", "allowed_tools": "Bash"}`,
			validate: func(t *testing.T, fm SkillFrontMatter) {
				t.Helper()

				if fm.Name != "my-skill" {
					t.Errorf("Name = %q, want my-skill", fm.Name)
				}

				if fm.License != "MIT" {
					t.Errorf("License = %q, want MIT", fm.License)
				}

				if fm.Compatibility != "go>=1.21" {
					t.Errorf("Compatibility = %q, want go>=1.21", fm.Compatibility)
				}

				if fm.AllowedTools != "Bash" {
					t.Errorf("AllowedTools = %q, want Bash", fm.AllowedTools)
				}
			},
		},
		{
			name:  "extra fields populate Content map",
			input: `{"name": "extra-skill", "custom-key": "custom-val"}`,
			validate: func(t *testing.T, fm SkillFrontMatter) {
				t.Helper()

				if fm.Name != "extra-skill" {
					t.Errorf("Name = %q, want extra-skill", fm.Name)
				}

				if fm.Content == nil {
					t.Fatal("Content should not be nil")
				}

				if v, ok := fm.Content["custom-key"]; !ok || v != "custom-val" {
					t.Errorf("Content[custom-key] = %v, want custom-val", v)
				}
			},
		},
		{
			name:    "invalid JSON returns error",
			input:   `{bad json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fm SkillFrontMatter

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
