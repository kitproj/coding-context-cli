package markdown

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "github.com/goccy/go-yaml"
	goldmark "github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/tokencount"
)

// ParseError is a markdown parsing error with file, line, and column position.
type ParseError struct {
	File    string
	Line    int // 1-indexed; 0 means unknown
	Column  int // 1-indexed; 0 means unknown
	Message string
}

func (e *ParseError) Error() string {
	switch {
	case e.Line > 0 && e.Column > 0:
		return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	case e.Line > 0:
		return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
	default:
		return fmt.Sprintf("%s: %s", e.File, e.Message)
	}
}

// Markdown represents a markdown file with frontmatter and content.
type Markdown[T any] struct {
	FrontMatter T               // Parsed YAML frontmatter
	Content     string          // Markdown body, excluding frontmatter
	Structure   ast.Node        // Document AST from goldmark parse
	Task        taskparser.Task // Parsed task structure (slash commands and text blocks)
	Tokens      int             // Estimated token count
}

// FromContent creates a Markdown from processed content string (e.g. after
// parameter expansion). The content is parsed to produce the Structure AST.
func FromContent[T any](frontMatter T, content string) Markdown[T] {
	source := []byte(content)
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	return Markdown[T]{
		FrontMatter: frontMatter,
		Content:     content,
		Structure:   doc,
		Tokens:      tokencount.EstimateTokens(content),
	}
}

// TaskMarkdown is a Markdown with TaskFrontMatter.
type TaskMarkdown = Markdown[TaskFrontMatter]

// RuleMarkdown is a Markdown with RuleFrontMatter.
type RuleMarkdown = Markdown[RuleFrontMatter]

// ParseMarkdownFile parses a markdown file into frontmatter and content using goldmark.
// Errors include file path and, where available, line and column position.
func ParseMarkdownFile[T any](path string, frontMatter *T) (Markdown[T], error) {
	cleanPath := filepath.Clean(path)

	source, err := os.ReadFile(cleanPath)
	if err != nil {
		return Markdown[T]{}, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	// Parse with goldmark+meta+taskparser in a single pass: meta extracts frontmatter,
	// taskparser.Extension captures task structure (slash commands) from the body.
	pctx := parser.NewContext()
	doc := goldmark.New(goldmark.WithExtensions(meta.Meta, taskparser.Extension)).Parser().
		Parse(text.NewReader(source), parser.WithContext(pctx))

	// Get frontmatter map from goldmark-meta (parsed during goldmark parse).
	metaData, yamlErr := meta.TryGet(pctx)
	if yamlErr != nil {
		line := yamlMessageLine(yamlErr.Error())
		// Offset by 1 to account for the opening "---" delimiter line.
		if line > 0 {
			line++
		}

		return Markdown[T]{}, &ParseError{
			File:    path,
			Line:    line,
			Message: fmt.Sprintf("failed to parse YAML frontmatter: %v", yamlErr),
		}
	}

	if len(metaData) > 0 {
		// Marshal map to YAML and unmarshal into typed struct to reproject onto T.
		yamlBytes, err := yaml.Marshal(metaData)
		if err != nil {
			return Markdown[T]{}, &ParseError{
				File:    path,
				Message: fmt.Sprintf("failed to marshal frontmatter: %v", err),
			}
		}

		if err := yaml.Unmarshal(yamlBytes, frontMatter); err != nil {
			line, col := yamlErrorPosition(err)
			if line > 0 {
				line++ // offset for opening "---" delimiter
			}

			return Markdown[T]{}, &ParseError{
				File:    path,
				Line:    line,
				Column:  col,
				Message: fmt.Sprintf("failed to parse YAML frontmatter: %v", err),
			}
		}
	}

	content := string(source[contentStartOffset(source):])

	task, _ := taskparser.GetTask(pctx)

	return Markdown[T]{
		FrontMatter: *frontMatter,
		Content:     content,
		Structure:   doc,
		Task:        task,
		Tokens:      tokencount.EstimateTokens(content),
	}, nil
}

// contentStartOffset returns the byte offset at which document content begins,
// after the frontmatter block delimited by "---". Returns 0 when no frontmatter
// is present.
func contentStartOffset(source []byte) int {
	const sep = "---\n"
	if !bytes.HasPrefix(source, []byte(sep)) {
		return 0
	}

	pos := len(sep)
	for pos < len(source) {
		next := bytes.IndexByte(source[pos:], '\n')
		if next < 0 {
			break
		}

		lineEnd := pos + next + 1

		line := bytes.TrimRight(source[pos:lineEnd], "\r\n")
		if bytes.Equal(line, []byte("---")) {
			return lineEnd
		}

		pos = lineEnd
	}

	return 0
}

// yamlMessageLine parses a line number from a yaml.v2-style error message.
// yaml.v2 errors look like "yaml: line N: message".
func yamlMessageLine(msg string) int {
	var line int
	if n, _ := fmt.Sscanf(msg, "yaml: line %d:", &line); n == 1 && line > 0 {
		return line
	}

	return 0
}

// yamlErrorPosition extracts line and column from a goccy/go-yaml error.
// goccy formats errors as "[line:col] message\n<source context>".
func yamlErrorPosition(err error) (int, int) {
	const maxLinesForPosition = 2
	// goccy/go-yaml formats errors as "[line:col] message\n..."
	firstLine := strings.SplitN(err.Error(), "\n", maxLinesForPosition)[0]

	var l, c int
	if n, _ := fmt.Sscanf(firstLine, "[%d:%d]", &l, &c); n == 2 && l > 0 {
		return l, c
	}
	// Fallback: yaml.v2-style "yaml: line N: message" (from goldmark-meta errors)
	return yamlMessageLine(firstLine), 0
}
