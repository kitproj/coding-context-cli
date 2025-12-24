package markdown

import (
	"encoding/json"
	"fmt"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

// BaseFrontMatter represents parsed YAML frontmatter from markdown files
type BaseFrontMatter struct {
	Content map[string]any `json:"-" yaml:",inline"`
}

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

	// Resume indicates if this task should be resumed
	Resume bool `yaml:"resume,omitempty" json:"resume,omitempty"`

	// Selectors contains additional custom selectors for filtering rules
	Selectors map[string]any `yaml:"selectors,omitempty" json:"selectors,omitempty"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `yaml:"expand,omitempty" json:"expand,omitempty"`
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
		return fmt.Errorf("failed to unmarshal task frontmatter: %w", err)
	}

	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &t.BaseFrontMatter.Content); err != nil {
		return fmt.Errorf("failed to unmarshal task frontmatter content: %w", err)
	}

	return nil
}

// CommandFrontMatter represents the frontmatter fields for command files.
// Previously this was an empty placeholder struct, but now supports the expand field
// to control parameter expansion behavior in command content.
type CommandFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `yaml:"expand,omitempty" json:"expand,omitempty"`

	// Selectors contains additional custom selectors for filtering rules
	// When a command is used in a task, its selectors are combined with task selectors
	Selectors map[string]any `yaml:"selectors,omitempty" json:"selectors,omitempty"`
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
		return fmt.Errorf("failed to unmarshal command frontmatter: %w", err)
	}

	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &c.BaseFrontMatter.Content); err != nil {
		return fmt.Errorf("failed to unmarshal command frontmatter content: %w", err)
	}

	return nil
}

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

	// MCPServer specifies a single MCP server configuration
	// Metadata only, does not filter
	MCPServer mcp.MCPServerConfig `yaml:"mcp_server,omitempty" json:"mcp_server,omitempty"`

	// RuleName is an optional identifier for the rule file
	RuleName string `yaml:"rule_name,omitempty" json:"rule_name,omitempty"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `yaml:"expand,omitempty" json:"expand,omitempty"`
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
		return fmt.Errorf("failed to unmarshal rule frontmatter: %w", err)
	}

	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &r.BaseFrontMatter.Content); err != nil {
		return fmt.Errorf("failed to unmarshal rule frontmatter content: %w", err)
	}

	return nil
}

// SkillFrontMatter represents the frontmatter fields for skill files
type SkillFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// SkillName is an optional identifier for the skill
	SkillName string `yaml:"skill_name,omitempty" json:"skill_name,omitempty"`

	// TaskNames specifies which task(s) this skill applies to
	// Array of task names for OR logic
	TaskNames []string `yaml:"task_names,omitempty" json:"task_names,omitempty"`

	// Languages specifies which programming language(s) this skill applies to
	// Array of languages for OR logic (e.g., ["go", "python"])
	Languages []string `yaml:"languages,omitempty" json:"languages,omitempty"`

	// Agent specifies which AI agent this skill is intended for
	Agent string `yaml:"agent,omitempty" json:"agent,omitempty"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `yaml:"expand,omitempty" json:"expand,omitempty"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map
func (s *SkillFrontMatter) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary type to avoid infinite recursion
	type Alias SkillFrontMatter
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal skill frontmatter: %w", err)
	}

	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &s.BaseFrontMatter.Content); err != nil {
		return fmt.Errorf("failed to unmarshal skill frontmatter content: %w", err)
	}

	return nil
}
