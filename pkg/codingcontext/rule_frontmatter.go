package codingcontext

// RuleFrontMatter represents the standard frontmatter fields for rule files
type RuleFrontMatter struct {
	FrontMatter `yaml:",inline"`

	// TaskNames specifies which task(s) this rule applies to
	// Array of task names for OR logic
	TaskNames []string `yaml:"task_names,omitempty"`

	// Languages specifies which programming language(s) this rule applies to
	// Array of languages for OR logic (e.g., ["go", "python"])
	Languages []string `yaml:"languages,omitempty"`

	// Agent specifies which AI agent this rule is intended for
	Agent string `yaml:"agent,omitempty"`

	// MCPServers lists the MCP servers that need to be running for this rule
	// Metadata only, does not filter
	MCPServers []MCPServerConfig `yaml:"mcp_servers,omitempty"`

	// RuleName is an optional identifier for the rule file
	RuleName string `yaml:"rule_name,omitempty"`
}
