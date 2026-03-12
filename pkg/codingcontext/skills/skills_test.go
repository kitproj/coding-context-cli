package skills

import (
	"testing"
)

func availableSkillsCases() []struct {
	name    string
	skills  AvailableSkills
	want    string
	wantErr bool
} {
	return []struct {
		name    string
		skills  AvailableSkills
		want    string
		wantErr bool
	}{
		{
			name:    "empty skills",
			skills:  AvailableSkills{Skills: []Skill{}},
			want:    "<available_skills></available_skills>",
			wantErr: false,
		},
		{
			name: "single skill",
			skills: AvailableSkills{Skills: []Skill{
				{Name: "test-skill", Description: "A test skill", Location: "/path/to/skill/SKILL.md"},
			}},
			want: "<available_skills>\n  <skill>\n    <name>test-skill</name>\n" +
				"    <description>A test skill</description>\n" +
				"    <location>/path/to/skill/SKILL.md</location>\n  </skill>\n</available_skills>",
		},
		{
			name: "multiple skills",
			skills: AvailableSkills{Skills: []Skill{
				{Name: "skill-one", Description: "First skill", Location: "/path/to/skill-one/SKILL.md"},
				{Name: "skill-two", Description: "Second skill", Location: "/path/to/skill-two/SKILL.md"},
			}},
			want: "<available_skills>\n  <skill>\n    <name>skill-one</name>\n" +
				"    <description>First skill</description>\n" +
				"    <location>/path/to/skill-one/SKILL.md</location>\n  </skill>\n" +
				"  <skill>\n    <name>skill-two</name>\n    <description>Second skill</description>\n" +
				"    <location>/path/to/skill-two/SKILL.md</location>\n  </skill>\n</available_skills>",
		},
		{
			name: "skill with special XML characters",
			skills: AvailableSkills{Skills: []Skill{
				{Name: "special-chars", Description: "Test <tag> & \"quotes\" 'apostrophes'",
					Location: "/path/to/skill/SKILL.md"},
			}},
			want: "<available_skills>\n  <skill>\n    <name>special-chars</name>\n" +
				"    <description>Test &lt;tag&gt; &amp; &#34;quotes&#34; &#39;apostrophes&#39;</description>\n" +
				"    <location>/path/to/skill/SKILL.md</location>\n  </skill>\n</available_skills>",
		},
	}
}

func TestAvailableSkills_AsXML(t *testing.T) {
	t.Parallel()

	for _, tt := range availableSkillsCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()
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
