package codingcontext

import (
	"strconv"
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

// Block represents either a slash command, shell command, or text content
type Block struct {
	SlashCommand *SlashCommand `parser:"@@"`
	ShellCommand *ShellCommand `parser:"| @@"`
	Text         *Text         `parser:"| @@"`
}

// SlashCommand represents a command starting with "/" that ends with a newline or EOF
// The newline is optional to handle EOF, but when present, prevents matching inline slashes
type SlashCommand struct {
	Name      string     `parser:"Slash @Term"`
	Arguments []Argument `parser:"(Whitespace @@)* Whitespace? Newline?"`
}

// ShellCommand represents a shell command starting with "!`" and ending with "`"
// The command output will be injected into the prompt
type ShellCommand struct {
	FullCommand string `parser:"@ShellCommand"`
}

// Params converts the slash command's arguments into a parameter map
// Returns a map with:
// - "ARGUMENTS": space-separated string of all arguments
// - "1", "2", etc.: positional parameters (1-indexed)
// - named parameters: key-value pairs from key="value" arguments
func (s *SlashCommand) Params() map[string]string {
	params := make(map[string]string)

	// Build the ARGUMENTS string from all arguments
	if len(s.Arguments) > 0 {
		var argStrings []string
		for _, arg := range s.Arguments {
			if arg.Key != "" {
				// Named parameter: key="value"
				argStrings = append(argStrings, arg.Key+"="+arg.Value)
			} else {
				// Positional parameter
				argStrings = append(argStrings, arg.Value)
			}
		}
		params["ARGUMENTS"] = strings.Join(argStrings, " ")
	}

	// Add positional and named parameters
	for i, arg := range s.Arguments {
		// Positional parameter (1-indexed)
		posKey := strconv.Itoa(i + 1)
		if arg.Key != "" {
			// This is a named parameter - store as key="value" for positional
			params[posKey] = arg.Key + "=" + arg.Value
			// Also store the value under the key name (strip quotes if present)
			params[arg.Key] = stripQuotes(arg.Value)
		} else {
			// Pure positional parameter (strip quotes if present)
			params[posKey] = stripQuotes(arg.Value)
		}
	}

	return params
}

// stripQuotes removes surrounding double quotes from a string if present.
// Single quotes are not supported as the grammar only allows double-quoted strings.
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		// Remove quotes and handle escaped quotes inside
		unquoted := s[1 : len(s)-1]
		return strings.ReplaceAll(unquoted, `\"`, `"`)
	}
	return s
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

// TextLine is a single line of text content (not starting with a slash or shell command)
// It matches tokens until the end of the line, or just a newline by itself
type TextLine struct {
	JustNewline   string   `parser:"@Newline"`                                            // A line that's just a newline
	NonSlashStart []string `parser:"| (@Term | @String | @Assign | @Whitespace)"`         // First token can't be Slash or ShellCommand
	RestOfLine    []string `parser:"(@Term | @String | @Slash | @Assign | @Whitespace)*"` // Rest can include Slash
	NewlineOpt    string   `parser:"@Newline?"`                                           // Optional newline at end
}

// Content returns the text content with all lines concatenated
func (t *Text) Content() string {
	var sb strings.Builder
	for _, line := range t.Lines {
		if line.JustNewline != "" {
			sb.WriteString(line.JustNewline)
		} else {
			for _, tok := range line.NonSlashStart {
				sb.WriteString(tok)
			}
			for _, tok := range line.RestOfLine {
				sb.WriteString(tok)
			}
			sb.WriteString(line.NewlineOpt)
		}
	}
	return sb.String()
}

// Define the lexer using participle's lexer.MustSimple
var taskLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "ShellCommand", Pattern: "!`[^`]+`"},   // Shell command: !`command`
	{Name: "Slash", Pattern: `/`},                 // Any "/"
	{Name: "Assign", Pattern: `=`},                // "="
	{Name: "String", Pattern: `"(?:\\.|[^"])*"`},  // Quoted strings with escapes
	{Name: "Whitespace", Pattern: `[ \t]+`},       // Spaces and tabs (horizontal only)
	{Name: "Newline", Pattern: `[\n\r]+`},         // Newlines
	{Name: "Term", Pattern: "[^ \\t\\n\\r/\"=]+"}, // Any char except space, newline, /, ", =
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
	if b.ShellCommand != nil {
		return b.ShellCommand.String()
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

// Command extracts the command string from the full shell command token
func (s *ShellCommand) Command() string {
	// Strip the !` prefix and ` suffix
	if len(s.FullCommand) > 3 && s.FullCommand[:2] == "!`" && s.FullCommand[len(s.FullCommand)-1] == '`' {
		return s.FullCommand[2 : len(s.FullCommand)-1]
	}
	return s.FullCommand
}

// String returns the original text representation of a shell command
func (s ShellCommand) String() string {
	return s.FullCommand
}

// String returns the original text representation of text
func (t Text) String() string {
	return t.Content()
}
