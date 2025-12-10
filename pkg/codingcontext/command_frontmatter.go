package codingcontext

import (
	"encoding/json"
)

// CommandFrontMatter represents the frontmatter fields for command files
type CommandFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `yaml:"expand_params,omitempty" json:"expand_params,omitempty"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map
func (c *CommandFrontMatter) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary type to avoid infinite recursion
	type Alias CommandFrontMatter
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &c.BaseFrontMatter.Content); err != nil {
		return err
	}

	return nil
}
