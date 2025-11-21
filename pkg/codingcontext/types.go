package codingcontext

// TaskFrontMatter represents the standard frontmatter fields for task files
type TaskFrontMatter struct {
	// TaskName is the unique identifier for the task (required)
	TaskName string `yaml:"task_name"`

	// Agent specifies the target agent (e.g., "cursor", "copilot")
	// When set, excludes the agent's own rules (same as -a flag)
	Agent string `yaml:"agent,omitempty"`

	// Language specifies the programming language(s) for filtering rules
	// Can be a string or array for OR logic (e.g., "go" or ["go", "python"])
	Language any `yaml:"language,omitempty"`

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
	MCPServers []string `yaml:"mcp_servers,omitempty"`

	// Resume indicates if this task should be resumed
	Resume bool `yaml:"resume,omitempty"`

	// Selectors contains additional custom selectors for filtering rules
	Selectors map[string]any `yaml:"selectors,omitempty"`
}

// RuleFrontMatter represents the standard frontmatter fields for rule files
type RuleFrontMatter struct {
	// TaskName specifies which task(s) this rule applies to
	// Can be a string or array for OR logic
	TaskName any `yaml:"task_name,omitempty"`

	// Language specifies which programming language(s) this rule applies to
	// Can be a string or array for OR logic (e.g., "go" or ["go", "python"])
	Language any `yaml:"language,omitempty"`

	// Agent specifies which AI agent this rule is intended for
	Agent string `yaml:"agent,omitempty"`

	// MCPServers lists the MCP servers that need to be running for this rule
	// Metadata only, does not filter
	MCPServers []string `yaml:"mcp_servers,omitempty"`

	// RuleName is an optional identifier for the rule file
	RuleName string `yaml:"rule_name,omitempty"`
}
