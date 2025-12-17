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

// stripQuotes removes surrounding quotes from a string if present and processes escape sequences.
// Supports both single (') and double (") quotes.
// Processes escape sequences: \n, \t, \r, \\, \", \', \uXXXX (Unicode), \xHH (hex), \OOO (octal)
// Unknown escape sequences (e.g., \z) preserve only the character after the backslash.
// Incomplete escape sequences (e.g., \u00a, \x4) are preserved literally including the backslash.
func stripQuotes(s string) string {
	// Check if the string is quoted
	if len(s) < 2 {
		return s
	}

	quoteChar := byte(0)
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		quoteChar = s[0]
		s = s[1 : len(s)-1] // Remove surrounding quotes
	}

	// If not quoted, return as-is (no escape processing for unquoted values in slash commands)
	if quoteChar == 0 {
		return s
	}

	// Process escape sequences
	return processEscapeSequences(s)
}

// processEscapeSequences decodes escape sequences in a string.
// Supports: \n, \t, \r, \\, \", \', \uXXXX (Unicode), \xHH (hex), \OOO (octal)
// Unknown escape sequences (e.g., \z) preserve only the character after the backslash.
// Incomplete escape sequences (e.g., \u00a, \x4) are preserved literally including the backslash.
func processEscapeSequences(s string) string {
	if !strings.Contains(s, "\\") {
		return s // Fast path: no escapes
	}

	var result strings.Builder
	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i == len(s)-1 {
			result.WriteByte(s[i])
			continue
		}

		// We have a backslash and there's at least one more character
		i++ // Move past backslash
		switch s[i] {
		case 'n':
			result.WriteByte('\n')
		case 't':
			result.WriteByte('\t')
		case 'r':
			result.WriteByte('\r')
		case '\\':
			result.WriteByte('\\')
		case '"':
			result.WriteByte('"')
		case '\'':
			result.WriteByte('\'')
		case 'u':
			// Unicode escape: \uXXXX
			if i+5 <= len(s) {
				hexStr := s[i+1 : i+5]
				if val, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
					result.WriteRune(rune(val))
					i += 4
				} else {
					// Invalid Unicode escape, keep as-is
					result.WriteString("\\u")
				}
			} else {
				// Incomplete Unicode escape
				result.WriteString("\\u")
			}
		case 'x':
			// Hex escape: \xHH
			if i+3 <= len(s) {
				hexStr := s[i+1 : i+3]
				if val, err := strconv.ParseInt(hexStr, 16, 8); err == nil {
					result.WriteByte(byte(val))
					i += 2
				} else {
					// Invalid hex escape, keep as-is
					result.WriteString("\\x")
				}
			} else {
				// Incomplete hex escape
				result.WriteString("\\x")
			}
		default:
			// Check for octal escape: \OOO (up to 3 digits, 0-7)
			if s[i] >= '0' && s[i] <= '7' {
				octalStart := i
				octalEnd := i + 1
				// Read up to 2 more octal digits
				for octalEnd-octalStart < 3 && octalEnd < len(s) && s[octalEnd] >= '0' && s[octalEnd] <= '7' {
					octalEnd++
				}
				octalStr := s[octalStart:octalEnd]
				if val, err := strconv.ParseInt(octalStr, 8, 16); err == nil {
					result.WriteByte(byte(val))
					i = octalEnd - 1
				} else {
					// Invalid octal, keep as-is
					result.WriteByte('\\')
					result.WriteByte(s[i])
				}
			} else {
				// Unknown escape sequence, keep the character after backslash
				result.WriteByte(s[i])
			}
		}
	}

	return result.String()
}

// Argument represents either a named (key=value) or positional argument
type Argument struct {
	Key   string `parser:"(@Term Assign)?"`
	Value string `parser:"(@DQString | @SQString | @Term)"`
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
	NonSlashStart []string `parser:"(@Term | @DQString | @SQString | @Assign | @Whitespace)"`           // First token can't be Slash
	RestOfLine    []string `parser:"(@Term | @DQString | @SQString | @Slash | @Assign | @Whitespace)*"` // Rest can include Slash
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
	{Name: "Slash", Pattern: `/`},                  // Any "/"
	{Name: "Assign", Pattern: `=`},                 // "="
	{Name: "DQString", Pattern: `"(?:\\.|[^"])*"`}, // Double-quoted strings with escapes
	{Name: "SQString", Pattern: `'(?:\\.|[^'])*'`}, // Single-quoted strings with escapes
	{Name: "Whitespace", Pattern: `[ \t]+`},        // Spaces and tabs (horizontal only)
	{Name: "Newline", Pattern: `[\n\r]+`},          // Newlines
	{Name: "Term", Pattern: `[^ \t\n\r/"=']+`},     // Any char except space, newline, /, ", ', =
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
