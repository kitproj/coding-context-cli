package codingcontext

import "github.com/goccy/go-yaml"

// TaskFrontMatter represents the standard frontmatter fields for task files
type TaskFrontMatter struct {
	FrontMatter `yaml:",inline"`

	// TaskName is the unique identifier for the task (required)
	TaskName string `yaml:"task_name"`

	// Agent specifies the default agent if not specified via -a flag
	// This is not used for selecting tasks or rules, only as a default
	Agent string `yaml:"agent,omitempty"`

	// Languages specifies the programming language(s) for filtering rules
	// Array of languages for OR logic (e.g., ["go", "python"])
	Languages []string `yaml:"languages,omitempty"`

	// Model specifies the AI model identifier
	// Does not filter rules, metadata only
	Model string `yaml:"model,omitempty"`

	// SingleShot indicates whether the task runs once or multiple times
	// Does not filter rules, metadata only
	SingleShot bool `yaml:"single_shot,omitempty"`

	// Timeout specifies the task timeout in time.Duration format (e.g., "10m", "1h")
	// Does not filter rules, metadata only
	Timeout string `yaml:"timeout,omitempty"`

	// MCPServers lists the MCP servers required for this task
	// Does not filter rules, metadata only
	MCPServers []MCPServerConfig `yaml:"mcp_servers,omitempty"`

	// Resume indicates if this task should be resumed
	Resume bool `yaml:"resume,omitempty"`

	// Selectors contains additional custom selectors for filtering rules
	Selectors map[string]any `yaml:"selectors,omitempty"`
}

// SyncToContent ensures all typed fields are also present in the Content map
func (t *TaskFrontMatter) SyncToContent() error {
	// Marshal the whole struct to YAML, then unmarshal back to get the Content map
	yamlBytes, err := yaml.Marshal(t)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(yamlBytes, &t.FrontMatter)
}
