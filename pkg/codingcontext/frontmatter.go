package codingcontext

// FrontMatter represents parsed YAML frontmatter from markdown files
type FrontMatter struct {
	Content map[string]any `json:"-" yaml:",inline"`
}

// NewFrontMatter creates a new FrontMatter with an initialized Content map
func NewFrontMatter() FrontMatter {
	return FrontMatter{Content: make(map[string]any)}
}
