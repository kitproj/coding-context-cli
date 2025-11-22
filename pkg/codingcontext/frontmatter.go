package codingcontext

// BaseFrontMatter represents parsed YAML frontmatter from markdown files
type BaseFrontMatter struct {
	Content map[string]any `json:"-" yaml:",inline"`
}
