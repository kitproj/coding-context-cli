package codingcontext

// FrontMatter represents parsed YAML frontmatter from markdown files
type FrontMatter struct {
	Content map[string]any `json:"-" yaml:",inline"`
}
