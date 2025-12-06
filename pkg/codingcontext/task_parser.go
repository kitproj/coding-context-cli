package codingcontext

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Input is the top-level wrapper for parsing
type Input struct {
	Blocks []Block `parser:"@@*"`
}

// Task represents a parsed task, which is a sequence of blocks
type Task []Block

// Block represents either a slash command or text content
type Block struct {
	SlashCommand *SlashCommand `parser:"@@"`
	Text         *Text         `parser:"| @@"`
}

// SlashCommand represents a command starting with "/" that ends with a newline or EOF
// The newline is optional to handle EOF, but when present, prevents matching inline slashes
type SlashCommand struct {
	Name      string     `parser:"Slash @Term"`
	Arguments []Argument `parser:"(Whitespace @@)* Whitespace? Newline?"`
}

// Argument represents either a named (key=value) or positional argument
type Argument struct {
	Key   string `parser:"(@Term Assign)?"`
	Value string `parser:"(@String | @Term)"`
}

// Text represents a block of text
// It can span multiple lines, consuming line content and newlines
// But it will stop before a newline that's followed by a slash (potential command)
type Text struct {
	Lines []TextLine `parser:"@@+"`
}

// TextLine is a single line of text content (not starting with a slash)
// It matches tokens until the end of the line
type TextLine struct {
	NonSlashStart []string `parser:"(@Term | @String | @Assign | @Whitespace)"`           // First token can't be Slash
	RestOfLine    []string `parser:"(@Term | @String | @Slash | @Assign | @Whitespace)*"` // Rest can include Slash
	NewlineOpt    string   `parser:"@Newline?"`
}

// Content returns the text content with all lines concatenated
func (t *Text) Content() string {
	var sb strings.Builder
	for _, line := range t.Lines {
		for _, tok := range line.NonSlashStart {
			sb.WriteString(tok)
		}
		for _, tok := range line.RestOfLine {
			sb.WriteString(tok)
		}
		sb.WriteString(line.NewlineOpt)
	}
	return sb.String()
}

// Define the lexer using participle's lexer.MustSimple
var taskLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Slash", Pattern: `/`},                // Any "/"
	{Name: "Assign", Pattern: `=`},               // "="
	{Name: "String", Pattern: `"(?:\\.|[^"])*"`}, // Quoted strings with escapes
	{Name: "Whitespace", Pattern: `[ \t]+`},      // Spaces and tabs (horizontal only)
	{Name: "Newline", Pattern: `[\n\r]+`},        // Newlines
	{Name: "Term", Pattern: `[^ \t\n\r/"=]+`},    // Any char except space, newline, /, ", =
})

var parser = participle.MustBuild[Input](
	participle.Lexer(taskLexer),
	participle.UseLookahead(4), // Use lookahead to help distinguish Text from SlashCommand patterns
)

// ParseTask parses a task string into a Task structure
func ParseTask(text string) (Task, error) {
	input, err := parser.ParseString("", text)
	if err != nil {
		return nil, err
	}
	return Task(input.Blocks), nil
}

// String returns the original text representation of a task
func (t Task) String() string {
	var sb strings.Builder
	for _, block := range t {
		sb.WriteString(block.String())
	}
	return sb.String()
}

// String returns the original text representation of a block
func (b Block) String() string {
	if b.SlashCommand != nil {
		return b.SlashCommand.String()
	}
	if b.Text != nil {
		return b.Text.String()
	}
	return ""
}

// String returns the original text representation of a slash command
func (s SlashCommand) String() string {
	var sb strings.Builder
	sb.WriteString("/")
	sb.WriteString(s.Name)
	for _, arg := range s.Arguments {
		sb.WriteString(" ")
		sb.WriteString(arg.String())
	}
	sb.WriteString("\n")
	return sb.String()
}

// String returns the original text representation of an argument
func (a Argument) String() string {
	if a.Key != "" {
		return a.Key + "=" + a.Value
	}
	return a.Value
}

// String returns the original text representation of text
func (t Text) String() string {
	return t.Content()
}
