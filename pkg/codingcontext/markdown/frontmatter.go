// Package markdown provides parsing and structs for markdown frontmatter.
package markdown

import (
	"encoding/json"
	"fmt"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

// BaseFrontMatter represents parsed YAML frontmatter from markdown files.
type BaseFrontMatter struct {
	// Name is the skill identifier
	// Must be 1-64 characters, lowercase alphanumeric and hyphens only
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Description explains what the prompt does and when to use it
	// Must be 1-1024 characters
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Content captures any frontmatter fields not explicitly declared in the struct.
	// With yaml:",inline", goccy/go-yaml populates this map with all unknown keys,
	// while known fields on the embedding struct (e.g. TaskNames, License) are set
	// directly on those fields. This ensures outer-struct fields are not shadowed.
	Content map[string]any `json:"-" yaml:",inline"`
}

// TaskFrontMatter represents the standard frontmatter fields for task files.
type TaskFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// Agent specifies the default agent if not specified via -a flag
	// This is not used for selecting tasks or rules, only as a default
	Agent string `json:"agent,omitempty" yaml:"agent,omitempty"`

	// Languages specifies the programming language(s) for filtering rules
	// Array of languages for OR logic (e.g., ["go", "python"])
	Languages []string `json:"languages,omitempty" yaml:"languages,omitempty"`

	// Model specifies the AI model identifier
	// Does not filter rules, metadata only
	Model string `json:"model,omitempty" yaml:"model,omitempty"`

	// SingleShot indicates whether the task runs once or multiple times
	// Does not filter rules, metadata only
	SingleShot bool `json:"single_shot,omitempty" yaml:"single_shot,omitempty"`

	// Timeout specifies the task timeout in time.Duration format (e.g., "10m", "1h")
	// Does not filter rules, metadata only
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// Resume indicates if this task should be resumed
	Resume bool `json:"resume,omitempty" yaml:"resume,omitempty"`

	// Selectors contains additional custom selectors for filtering rules
	Selectors map[string]any `json:"selectors,omitempty" yaml:"selectors,omitempty"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `json:"expand,omitempty" yaml:"expand,omitempty"`

	// IncludeUnmatched controls whether rules/skills that don't explicitly match
	// any active selector are included by default. Defaults to true (current behaviour).
	// Set to false to require an explicit selector match (strict/opt-in mode).
	IncludeUnmatched *bool `json:"include_unmatched,omitempty" yaml:"include_unmatched,omitempty"`
}

// populateContent unmarshals raw JSON into the inline Content map.
// Called by each concrete frontmatter type's UnmarshalJSON after the
// typed fields have been populated via the alias trick.
func (b *BaseFrontMatter) populateContent(data []byte) error {
	return json.Unmarshal(data, &b.Content)
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map.
func (t *TaskFrontMatter) UnmarshalJSON(data []byte) error {
	type Alias TaskFrontMatter
	aux := &struct{ *Alias }{Alias: (*Alias)(t)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal task frontmatter: %w", err)
	}
	return t.populateContent(data)
}

// CommandFrontMatter represents the frontmatter fields for command files.
// Previously this was an empty placeholder struct, but now supports the expand field
// to control parameter expansion behavior in command content.
type CommandFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `json:"expand,omitempty" yaml:"expand,omitempty"`

	// Selectors contains additional custom selectors for filtering rules
	// When a command is used in a task, its selectors are combined with task selectors
	Selectors map[string]any `json:"selectors,omitempty" yaml:"selectors,omitempty"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map.
func (c *CommandFrontMatter) UnmarshalJSON(data []byte) error {
	type Alias CommandFrontMatter
	aux := &struct{ *Alias }{Alias: (*Alias)(c)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal command frontmatter: %w", err)
	}
	return c.populateContent(data)
}

// RuleFrontMatter represents the standard frontmatter fields for rule files.
type RuleFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// TaskNames specifies which task(s) this rule applies to
	// Array of task names for OR logic
	TaskNames []string `json:"task_names,omitempty" yaml:"task_names,omitempty"`

	// Languages specifies which programming language(s) this rule applies to
	// Array of languages for OR logic (e.g., ["go", "python"])
	Languages []string `json:"languages,omitempty" yaml:"languages,omitempty"`

	// Agent specifies which AI agent this rule is intended for
	Agent string `json:"agent,omitempty" yaml:"agent,omitempty"`

	// MCPServer specifies a single MCP server configuration
	// Metadata only, does not filter
	MCPServer mcp.MCPServerConfig `json:"mcp_server,omitzero" yaml:"mcp_server,omitzero"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `json:"expand,omitempty" yaml:"expand,omitempty"`

	// Bootstrap contains a shell script to execute before including the rule
	// This is preferred over file-based bootstrap scripts
	Bootstrap string `json:"bootstrap,omitempty" yaml:"bootstrap,omitempty"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map.
func (r *RuleFrontMatter) UnmarshalJSON(data []byte) error {
	type Alias RuleFrontMatter
	aux := &struct{ *Alias }{Alias: (*Alias)(r)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal rule frontmatter: %w", err)
	}
	return r.populateContent(data)
}

// SkillFrontMatter represents the standard frontmatter fields for skill files.
type SkillFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// License specifies the license applied to the skill (optional)
	License string `json:"license,omitempty" yaml:"license,omitempty"`

	// Compatibility indicates environment requirements (optional)
	// Max 500 characters
	Compatibility string `json:"compatibility,omitempty" yaml:"compatibility,omitempty"`

	// Metadata contains arbitrary key-value pairs (optional)
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// AllowedTools is a space-delimited list of pre-approved tools (optional, experimental)
	AllowedTools string `json:"allowed_tools,omitempty" yaml:"allowed_tools,omitempty"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map.
func (s *SkillFrontMatter) UnmarshalJSON(data []byte) error {
	type Alias SkillFrontMatter
	aux := &struct{ *Alias }{Alias: (*Alias)(s)}
	if err := json.Unmarshal(data, aux); err != nil {
		return fmt.Errorf("failed to unmarshal skill frontmatter: %w", err)
	}
	return s.populateContent(data)
}
