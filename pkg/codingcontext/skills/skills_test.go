package skills

import (
	"testing"
)

func TestAvailableSkills_AsXML(t *testing.T) {
	tests := []struct {
		name    string
		skills  AvailableSkills
		want    string
		wantErr bool
	}{
		{
			name: "empty skills",
			skills: AvailableSkills{
				Skills: []Skill{},
			},
			want:    "",
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
    <description>Test &lt;tag&gt; &amp; &#34;quotes&#34; &#39;apostrophes&#39;</description>
  </skill>
</available_skills>`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.skills.AsXML()
			if (err != nil) != tt.wantErr {
				t.Errorf("AvailableSkills.AsXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AvailableSkills.AsXML() mismatch\nGot:\n%s\n\nWant:\n%s", got, tt.want)
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
