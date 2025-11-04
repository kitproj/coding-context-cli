package main

import (
	"github.com/kitproj/coding-context-cli/internal/parser"
)

// parseMarkdownFile parses the file into frontmatter and content
func parseMarkdownFile(path string, frontmatter any) (string, error) {
	return parser.ParseMarkdownFile(path, frontmatter)
}
