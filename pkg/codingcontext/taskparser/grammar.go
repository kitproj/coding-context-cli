package taskparser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Task represents a parsed task, which is a sequence of blocks
type Task []Block

// Input is the top-level wrapper for parsing
type Input struct {
	Blocks []Block `parser:"@@*"`
}

// Block represents either a slash command or text content
type Block struct {
	SlashCommand *SlashCommand `parser:"@@"`
	Text         *Text         `parser:"| @@"`
}

// SlashCommand represents a command starting with "/" that ends with a newline or EOF
// The newline is optional to handle EOF, but when present, prevents matching inline slashes
// Leading whitespace is optional to allow indented commands
type SlashCommand struct {
	LeadingWhitespace string     `parser:"Whitespace?"`
	Name              string     `parser:"Slash @Term"`
	Arguments         []Argument `parser:"(Whitespace @@)* Whitespace? Newline?"`
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
	LeadingNewlines []string   `parser:"@Newline*"` // Leading newlines before any content (empty lines at the start)
	Lines           []TextLine `parser:"@@+"`       // At least one line with actual content
}

// TextLine is a single line of text content (not starting with a slash)
// It matches tokens until the end of the line
type TextLine struct {
	NonSlashStart []string `parser:"(@Term | @String | @Assign | @Whitespace)"`           // First token can't be Slash
	RestOfLine    []string `parser:"(@Term | @String | @Slash | @Assign | @Whitespace)*"` // Rest can include Slash
	NewlineOpt    string   `parser:"@Newline?"`
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

func parser() *participle.Parser[Input] {
	return participle.MustBuild[Input](
		participle.Lexer(taskLexer),
		participle.UseLookahead(4), // Use lookahead to help distinguish Text from SlashCommand patterns
	)
}

// ========== PARAMS GRAMMAR ==========

// ParamsInput is the top-level structure for parsing parameters
// It now directly parses into named parameters and positional arguments
type ParamsInput struct {
	Items []ParamsItem `parser:"@@*"`
}

// ParamsItem represents either a named parameter or a positional argument
type ParamsItem struct {
	Pos        lexer.Position
	Separator  *Separator  `parser:"@@"`   // Whitespace or comma
	Named      *NamedParam `parser:"| @@"` // key=value
	Positional *Value      `parser:"| @@"` // standalone value
}

// Separator represents whitespace or comma separators
type Separator struct {
	Pos lexer.Position
	Val string `parser:"(@Whitespace | @Comma)"`
}

// NamedParam represents a key=value pair
type NamedParam struct {
	Pos             lexer.Position
	Key             string  `parser:"@Token"`
	PreEqualsSpace  *string `parser:"Whitespace?"`
	Equals          string  `parser:"@Assign"`
	PostEqualsSpace *string `parser:"Whitespace?"`
	Value           *Value  `parser:"@@?"` // Optional to handle empty values like key=
}

// Value represents a parsed value (quoted or unquoted)
// The Raw field captures the entire token including quotes and escapes
type Value struct {
	Pos lexer.Position
	Raw string `parser:"(@QuotedDouble | @QuotedSingle | @Token)"`
}

// paramsLexer defines the lexer for parsing parameters
// Using a simpler approach: capture entire quoted strings and unquoted tokens
var paramsLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Whitespace", Pattern: `[\s\p{Z}]+`}, // Match ASCII and Unicode whitespace
	{Name: "Comma", Pattern: `,`},
	{Name: "Assign", Pattern: `=`},
	// Quoted strings - capture entire string including quotes
	// Handles escaped quotes inside
	{Name: "QuotedDouble", Pattern: `"(?:\\.|[^"\\])*"`},
	{Name: "QuotedSingle", Pattern: `'(?:\\.|[^'\\])*'`},
	// Unquoted token - matches any sequence of non-delimiter characters
	// Can contain escape sequences
	{Name: "Token", Pattern: `(?:\\.|[^\s\p{Z},="'\\])+`},
})

func paramsParser() *participle.Parser[ParamsInput] {
	return participle.MustBuild[ParamsInput](
		participle.Lexer(paramsLexer),
		participle.UseLookahead(3), // Lookahead to distinguish key=value from positional
	)
}
