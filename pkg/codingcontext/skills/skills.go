// Package skills provides types and serialization for agent skills.
package skills

import (
	"encoding/xml"
	"fmt"
)

// Skill represents a discovered skill with its metadata.
type Skill struct {
	XMLName     xml.Name `xml:"skill"`
	Name        string   `xml:"name"`
	Description string   `xml:"description"`
	Location    string   `xml:"location"` // Absolute path to the SKILL.md file
	// Plugin is the skill's "plugin" frontmatter value, if any. It is not
	// part of the prompt XML.
	Plugin string `xml:"-"`
	// SearchPath is the local search path the skill was discovered under. It
	// is not part of the prompt XML.
	SearchPath string `xml:"-"`
}

// AvailableSkills represents a collection of discovered skills.
type AvailableSkills struct {
	XMLName xml.Name `xml:"available_skills"`
	Skills  []Skill  `xml:"skill"`
}

// AsXML returns the XML representation of available skills.
func (a AvailableSkills) AsXML() (string, error) {
	// Use xml.MarshalIndent to properly encode the XML with indentation
	xmlBytes, err := xml.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal skills to XML: %w", err)
	}

	return string(xmlBytes), nil
}
