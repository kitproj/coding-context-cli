package skills

import (
	"testing"
)

func TestAvailableSkills_String(t *testing.T) {
	tests := []struct {
		name   string
		skills AvailableSkills
		want   string
	}{
		{
			name: "empty skills",
			skills: AvailableSkills{
				Skills: []Skill{},
			},
			want: "",
		},
		{
			name: "single skill",
			skills: AvailableSkills{
				Skills: []Skill{
					{
						Name:        "test-skill",
						Description: "A test skill",
						Location:    "/path/to/skill/SKILL.md",
					},
				},
			},
			want: `<available_skills>
  <skill>
    <name>test-skill</name>
    <description>A test skill</description>
  </skill>
</available_skills>`,
		},
		{
			name: "multiple skills",
			skills: AvailableSkills{
				Skills: []Skill{
					{
						Name:        "skill-one",
						Description: "First skill",
						Location:    "/path/to/skill-one/SKILL.md",
					},
					{
						Name:        "skill-two",
						Description: "Second skill",
						Location:    "/path/to/skill-two/SKILL.md",
					},
				},
			},
			want: `<available_skills>
  <skill>
    <name>skill-one</name>
    <description>First skill</description>
  </skill>
  <skill>
    <name>skill-two</name>
    <description>Second skill</description>
  </skill>
</available_skills>`,
		},
		{
			name: "skill with special XML characters",
			skills: AvailableSkills{
				Skills: []Skill{
					{
						Name:        "special-chars",
						Description: "Test <tag> & \"quotes\" 'apostrophes'",
						Location:    "/path/to/skill/SKILL.md",
					},
				},
			},
			want: `<available_skills>
  <skill>
    <name>special-chars</name>
    <description>Test &lt;tag&gt; &amp; &quot;quotes&quot; &apos;apostrophes&apos;</description>
  </skill>
</available_skills>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.skills.String()
			if got != tt.want {
				t.Errorf("AvailableSkills.String() mismatch\nGot:\n%s\n\nWant:\n%s", got, tt.want)
			}
		})
	}
}

func TestXMLEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special characters",
			input: "normal text",
			want:  "normal text",
		},
		{
			name:  "ampersand",
			input: "A & B",
			want:  "A &amp; B",
		},
		{
			name:  "less than and greater than",
			input: "<tag>",
			want:  "&lt;tag&gt;",
		},
		{
			name:  "quotes",
			input: `"quoted" text`,
			want:  "&quot;quoted&quot; text",
		},
		{
			name:  "apostrophes",
			input: "it's here",
			want:  "it&apos;s here",
		},
		{
			name:  "multiple special characters",
			input: `<tag attr="value" & 'single'>`,
			want:  "&lt;tag attr=&quot;value&quot; &amp; &apos;single&apos;&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xmlEscape(tt.input)
			if got != tt.want {
				t.Errorf("xmlEscape() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSkillsNotNil(t *testing.T) {
	// Ensure we can create skills without panicking
	skill := Skill{
		Name:        "test",
		Description: "test description",
		Location:    "/path/to/skill",
	}

	if skill.Name != "test" {
		t.Errorf("Expected name 'test', got %q", skill.Name)
	}
}
