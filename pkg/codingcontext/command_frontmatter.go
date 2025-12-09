package codingcontext

// CommandFrontMatter is an empty struct used as a placeholder for commands.
// Commands don't have frontmatter, but we use this type to maintain consistency
// in the getMarkdown API instead of passing nil.
type CommandFrontMatter struct{}
