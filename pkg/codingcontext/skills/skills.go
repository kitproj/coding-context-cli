package skills

import (
	"encoding/xml"
	"strings"
)

// Skill represents a discovered skill with its metadata
type Skill struct {
	XMLName     xml.Name `xml:"skill"`
	Name        string   `xml:"name"`
	Description string   `xml:"description"`
	Location    string   `xml:"-"` // Absolute path to the SKILL.md file
}

// AvailableSkills represents a collection of discovered skills
type AvailableSkills struct {
	XMLName xml.Name `xml:"available_skills"`
	Skills  []Skill  `xml:"skill"`
}

// String returns the XML representation of available skills
func (a AvailableSkills) String() string {
	if len(a.Skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<available_skills>\n")
	for _, skill := range a.Skills {
		sb.WriteString("  <skill>\n")
		sb.WriteString("    <name>")
		sb.WriteString(xmlEscape(skill.Name))
		sb.WriteString("</name>\n")
		sb.WriteString("    <description>")
		sb.WriteString(xmlEscape(skill.Description))
		sb.WriteString("</description>\n")
		sb.WriteString("  </skill>\n")
	}
	sb.WriteString("</available_skills>")
	return sb.String()
}

// xmlEscape escapes special XML characters
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
