package codingcontext

import (
	"encoding/json"
)

// TaskFrontMatter represents the standard frontmatter fields for task files
type TaskFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// Agent specifies the default agent if not specified via -a flag
	// This is not used for selecting tasks or rules, only as a default
	Agent string `yaml:"agent,omitempty" json:"agent,omitempty"`

	// Languages specifies the programming language(s) for filtering rules
	// Array of languages for OR logic (e.g., ["go", "python"])
	Languages []string `yaml:"languages,omitempty" json:"languages,omitempty"`

	// Model specifies the AI model identifier
	// Does not filter rules, metadata only
	Model string `yaml:"model,omitempty" json:"model,omitempty"`

	// SingleShot indicates whether the task runs once or multiple times
	// Does not filter rules, metadata only
	SingleShot bool `yaml:"single_shot,omitempty" json:"single_shot,omitempty"`

	// Timeout specifies the task timeout in time.Duration format (e.g., "10m", "1h")
	// Does not filter rules, metadata only
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// MCPServers lists the MCP servers required for this task
	// Does not filter rules, metadata only
	MCPServers []MCPServerConfig `yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`

	// Resume indicates if this task should be resumed
	Resume bool `yaml:"resume,omitempty" json:"resume,omitempty"`

	// Selectors contains additional custom selectors for filtering rules
	Selectors map[string]any `yaml:"selectors,omitempty" json:"selectors,omitempty"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map
func (t *TaskFrontMatter) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary type to avoid infinite recursion
	type Alias TaskFrontMatter
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &t.Content); err != nil {
		return err
	}

	return nil
}
