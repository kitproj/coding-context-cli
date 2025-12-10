package codingcontext

// CommandFrontMatter represents the frontmatter fields for command files
type CommandFrontMatter struct {
	BaseFrontMatter `yaml:",inline"`

	// ExpandParams controls whether parameter expansion should occur
	// Defaults to true if not specified
	ExpandParams *bool `yaml:"expand_params,omitempty" json:"expand_params,omitempty"`
}
