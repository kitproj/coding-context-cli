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
