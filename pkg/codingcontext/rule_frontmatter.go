package codingcontext

import (
	"encoding/json"
)

// RuleFrontMatter represents the standard frontmatter fields for rule files
type RuleFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// TaskNames specifies which task(s) this rule applies to
	// Array of task names for OR logic
	TaskNames []string `yaml:"task_names,omitempty" json:"task_names,omitempty"`

	// Languages specifies which programming language(s) this rule applies to
	// Array of languages for OR logic (e.g., ["go", "python"])
	Languages []string `yaml:"languages,omitempty" json:"languages,omitempty"`

	// Agent specifies which AI agent this rule is intended for
	Agent string `yaml:"agent,omitempty" json:"agent,omitempty"`

	// MCPServers lists the MCP servers that need to be running for this rule
	// Metadata only, does not filter
	MCPServers []MCPServerConfig `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`

	// RuleName is an optional identifier for the rule file
	RuleName string `yaml:"rule_name,omitempty" json:"rule_name,omitempty"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map
func (r *RuleFrontMatter) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary type to avoid infinite recursion
	type Alias RuleFrontMatter
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	
	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &r.BaseFrontMatter.Content); err != nil {
		return err
	}
	
	return nil
}
