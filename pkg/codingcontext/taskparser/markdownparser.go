package taskparser

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	goldmark "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// bodyOffset returns the byte offset where body content starts, after an optional
// YAML frontmatter block delimited by "---". Returns 0 when no frontmatter is present.
// This mirrors the contentStartOffset logic in the markdown package to avoid a circular import.
func bodyOffset(source []byte) int {
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

// codeRange represents a byte range [start, stop) in the source that should not be
// parsed for slash commands (e.g., fenced code blocks, indented code blocks, HTML blocks).
type codeRange struct {
	start, stop int
}

// collectCodeRanges walks the goldmark AST and returns the byte ranges of all code
// sections (fenced code blocks, indented code blocks, HTML blocks) in source order.
// These ranges cover only the content lines (not fence delimiter lines), which is
// sufficient because delimiter lines start with characters (“ ` “, `<`) that the
// grammar already treats as plain text.
func collectCodeRanges(doc ast.Node) ([]codeRange, error) {
	var ranges []codeRange

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindFencedCodeBlock, ast.KindCodeBlock, ast.KindHTMLBlock:
			lines := n.Lines()
			if lines.Len() > 0 {
				first := lines.At(0)
				last := lines.At(lines.Len() - 1)
				ranges = append(ranges, codeRange{first.Start, last.Stop})
			}

			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk AST: %w", err)
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].start < ranges[j].start
	})

	return ranges, nil
}

// splitAndParse splits content into alternating text/code sections based on the
// provided code ranges, then:
//   - For text sections: runs the grammar parser to detect slash commands.
//   - For code sections: wraps content in a raw Text block (no command detection).
//
// Each code range's stop position is extended to include any immediately following
// newline characters. This prevents the next text segment from starting with a bare
// newline immediately before a slash, which would cause the grammar parser to fail
// (it cannot parse a Text block that has only leading newlines before a slash).
func splitAndParse(content string, codeRanges []codeRange) (Task, error) {
	var allBlocks []Block

	pos := 0

	for i, cr := range codeRanges {
		if pos < cr.start {
			blocks, err := parseGrammar(content[pos:cr.start])
			if err != nil {
				return nil, err
			}

			allBlocks = append(allBlocks, blocks...)
		}

		// Extend stop past any trailing newlines so the next text segment never
		// starts with bare newlines before a slash command.
		stop := trailingNewlineEnd(content, cr.stop)

		// Never extend into the next code range.
		if i+1 < len(codeRanges) && stop > codeRanges[i+1].start {
			stop = codeRanges[i+1].start
		}

		if cr.start < stop {
			allBlocks = append(allBlocks, rawTextBlock(content[cr.start:stop]))
		}

		pos = stop
	}

	if pos < len(content) {
		blocks, err := parseGrammar(content[pos:])
		if err != nil {
			return nil, err
		}

		allBlocks = append(allBlocks, blocks...)
	}

	return Task(allBlocks), nil
}

// trailingNewlineEnd advances pos past any consecutive newline/carriage-return bytes.
func trailingNewlineEnd(content string, pos int) int {
	for pos < len(content) && (content[pos] == '\n' || content[pos] == '\r') {
		pos++
	}

	return pos
}

// parseGrammar runs the participle grammar parser on a plain text segment.
// It returns nil blocks (not an error) for whitespace-only input.
func parseGrammar(content string) ([]Block, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	input, err := parser().ParseString("", content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	return input.Blocks, nil
}

// rawTextBlock wraps a raw string as a Text block without any slash command parsing.
// Content() and String() on the returned block both return the original string exactly.
func rawTextBlock(content string) Block {
	return Block{
		Text: &Text{
			Lines: []TextLine{
				{RestOfLine: []string{content}},
			},
		},
	}
}

// Extension is a goldmark extension that parses task structure during the markdown parse.
// Include it in a goldmark instance and use GetTask to retrieve the parsed Task after parsing.
//
// Example:
//
//	pctx := parser.NewContext()
//	goldmark.New(goldmark.WithExtensions(taskparser.Extension)).Parser().
//	    Parse(text.NewReader(source), parser.WithContext(pctx))
//	task, err := taskparser.GetTask(pctx)
//
//nolint:gochecknoglobals // goldmark.WithExtensions expects a package-level extender
var Extension goldmark.Extender = &taskExtension{}

type taskExtension struct{}

func (e *taskExtension) Extend(m goldmark.Markdown) {
	const taskTransformerPriority = 100
	m.Parser().AddOptions(gparser.WithASTTransformers(
		util.Prioritized(&taskTransformer{}, taskTransformerPriority),
	))
}

// contextKey stores task parse results in a goldmark parser.Context.
//
//nolint:gochecknoglobals // parser context keys are conventionally package-level
var contextKey = gparser.NewContextKey()

type taskParseResult struct {
	task Task
	err  error
}

// GetTask retrieves the parsed Task from a goldmark parser.Context after a parse
// that included Extension. Returns (nil, nil) if Extension was not used.
func GetTask(pc gparser.Context) (Task, error) {
	v := pc.Get(contextKey)
	if v == nil {
		return nil, nil
	}

	r, ok := v.(*taskParseResult)
	if !ok {
		return nil, nil
	}

	return r.task, r.err
}

// taskTransformer implements parser.ASTTransformer. It runs after goldmark has built
// the document AST and extracts task structure (text vs. slash commands), skipping
// slash command detection inside code blocks, indented code, and HTML blocks.
type taskTransformer struct{}

func (t *taskTransformer) Transform(node *ast.Document, reader text.Reader, pc gparser.Context) {
	source := reader.Source()
	offset := bodyOffset(source)
	content := string(source[offset:])

	if strings.TrimSpace(content) == "" {
		pc.Set(contextKey, &taskParseResult{})

		return
	}

	ranges, err := collectCodeRanges(node)
	if err != nil {
		pc.Set(contextKey, &taskParseResult{err: err})

		return
	}

	// Adjust code ranges to be relative to content (body) rather than the full source.
	// When a goldmark parse includes frontmatter (e.g. via goldmark-meta), code block byte
	// positions in the AST are relative to the full source. We subtract the frontmatter
	// offset so that ranges align with the body-only content string passed to splitAndParse.
	adjusted := make([]codeRange, 0, len(ranges))
	for _, r := range ranges {
		adjStart := r.start - offset

		adjStop := r.stop - offset
		if adjStop <= 0 {
			continue // entirely within frontmatter
		}

		if adjStart < 0 {
			adjStart = 0
		}

		adjusted = append(adjusted, codeRange{adjStart, adjStop})
	}

	task, parseErr := splitAndParse(content, adjusted)
	pc.Set(contextKey, &taskParseResult{task: task, err: parseErr})
}

// parseMarkdownAware parses task content while skipping slash command detection inside
// code blocks (fenced code, indented code, HTML blocks) by running the Extension
// during a single goldmark parse pass.
func parseMarkdownAware(content string) (Task, error) {
	source := []byte(content)
	pctx := gparser.NewContext()
	goldmark.New(goldmark.WithExtensions(Extension)).Parser().
		Parse(text.NewReader(source), gparser.WithContext(pctx))

	return GetTask(pctx)
}
