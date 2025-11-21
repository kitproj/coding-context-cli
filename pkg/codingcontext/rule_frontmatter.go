package codingcontext

// RuleFrontMatter represents the standard frontmatter fields for rule files
type RuleFrontMatter struct {
	FrontMatter `yaml:",inline"`

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
