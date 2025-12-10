package codingcontext

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

// expander.go implements a text expansion system that supports three types of expansions:
//
// 1. Parameter Expansion: ${name}
//    - Expands to the value of the parameter with the given name
//    - If the parameter is not found, logs a warning and leaves the placeholder unexpanded
//    - Parameter names must match: [a-zA-Z_][a-zA-Z0-9_.-]*
//
// 2. Command Expansion: !`command`
//    - Executes the command in a shell and expands to its output
//    - If the command fails, logs a warning but still substitutes the output (which may be empty)
//    - Output is trimmed of leading/trailing whitespace
//
// 3. File Expansion: @path
//    - Reads the file at the given path and expands to its contents
//    - Supports escaped spaces in paths: @path\ with\ spaces.txt
//    - If the file is not found, logs a warning and leaves the placeholder unexpanded
//    - Path continues until unescaped whitespace or another expansion marker
//
// The parser uses a manual scanning approach rather than participle's lexer to handle
// the complex requirements of literal text mixed with multiple expansion types.
//
// Grammar (EBNF):
//
// (* Top-level construct: A string containing zero or more segments *)
// String          = { Segment } ;
//
// (* A segment is either raw text or one of the expansion types *)
// Segment         = Expansion | Literal ;
//
// (* The three supported expansion types *)
// Expansion       = ParameterExp | CommandExp | FileExp ;
//
// (* 1. Parameter Expansion: ${...} *)
// ParameterExp    = "${" , ParamName , "}" ;
// ParamName       = Identifier ;
//
// (* 2. Command Expansion: !`command` *)
// CommandExp      = "!" , BacktickString ;
// BacktickString  = "`" , { CommandChar } , "`" ;
//
// (* 3. File Expansion: @path *)
// (* The path continues until it hits an unescaped whitespace or End of String *)
// FileExp         = "@" , FilePath ;
// FilePath        = PathChar , { PathChar } ;
//
// (* Low-level definitions *)
// Literal         = LiteralChar , { LiteralChar } ;
//
// (* Identifier for parameters (e.g. valid variable names) *)
// Identifier      = Letter , { Letter | Digit | "_" | "-" | "." } ;
//
// (* Command contents: anything except the backtick delimiter *)
// CommandChar     = ? all characters except "`" ? ;
//
// (* Path logic: Allow normal path chars OR an escaped space *)
// PathChar        = NormalPathChar | EscapedSpace ;
// NormalPathChar  = ? any character except whitespace, "$", "!", "@" ? ;
// EscapedSpace    = "\" , " " ;
//
// (* Literal text: anything that doesn't trigger an expansion start *)
// LiteralChar     = ? all characters except "$", "!", "@" when followed by expansion markers ? ;
//
// (* Basic character classes *)
// Letter          = "a"..."z" | "A"..."Z" ;
// Digit           = "0"..."9" ;

// Segment represents a part of the string - either literal text or an expansion
type Segment interface {
	Expand(params map[string]string, logger *slog.Logger) string
}

// LiteralSegment represents literal text
type LiteralSegment struct {
	Text string
}

// ParameterSegment represents parameter expansion: ${...}
type ParameterSegment struct {
	Name string
}

// CommandSegment represents command expansion: !`command`
type CommandSegment struct {
	Command string
}

// FileSegment represents file expansion: @path
type FileSegment struct {
	Path string
}

// Expand methods for each segment type

func (l *LiteralSegment) Expand(params map[string]string, logger *slog.Logger) string {
	return l.Text
}

func (p *ParameterSegment) Expand(params map[string]string, logger *slog.Logger) string {
	if val, ok := params[p.Name]; ok {
		return val
	}
	// Parameter not found - log and leave unexpanded
	logger.Warn("parameter not found, leaving unexpanded", "param", p.Name)
	return fmt.Sprintf("${%s}", p.Name)
}

func (c *CommandSegment) Expand(params map[string]string, logger *slog.Logger) string {
	cmd := exec.Command("sh", "-c", c.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Command failed - log error code but still substitute output
		logger.Warn("command expansion failed", "command", c.Command, "error", err)
	}
	// Return the output (may be empty if command failed)
	return strings.TrimSpace(string(output))
}

func (f *FileSegment) Expand(params map[string]string, logger *slog.Logger) string {
	// Unescape spaces in the path
	path := strings.ReplaceAll(f.Path, `\ `, ` `)

	content, err := os.ReadFile(path)
	if err != nil {
		// File not found - log and leave unsubstituted
		logger.Warn("file not found, leaving unexpanded", "path", path, "error", err)
		return fmt.Sprintf("@%s", f.Path)
	}
	return string(content)
}

// parseString parses an input string into segments
func parseString(input string) ([]Segment, error) {
	var segments []Segment
	i := 0

	for i < len(input) {
		// Try to match an expansion
		if i < len(input)-1 && input[i] == '$' && input[i+1] == '{' {
			// Parameter expansion: ${...}
			seg, consumed, err := parseParameter(input[i:])
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
			i += consumed
		} else if i < len(input)-1 && input[i] == '!' && input[i+1] == '`' {
			// Command expansion: !`...`
			seg, consumed, err := parseCommand(input[i:])
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
			i += consumed
		} else if input[i] == '@' {
			// File expansion: @path
			seg, consumed := parseFile(input[i:])
			segments = append(segments, seg)
			i += consumed
		} else {
			// Literal text - consume until we hit an expansion start
			start := i
			for i < len(input) {
				// Check if we're at the start of an expansion
				if input[i] == '$' && i+1 < len(input) && input[i+1] == '{' {
					break
				}
				if input[i] == '!' && i+1 < len(input) && input[i+1] == '`' {
					break
				}
				if input[i] == '@' {
					// Only break if @ is followed by a valid path char
					if i+1 < len(input) && (unicode.IsLetter(rune(input[i+1])) || input[i+1] == '/' || input[i+1] == '.') {
						break
					}
				}
				i++
			}
			if i > start {
				segments = append(segments, &LiteralSegment{Text: input[start:i]})
			}
		}
	}

	return segments, nil
}

// parseParameter parses a parameter expansion ${name}
func parseParameter(input string) (*ParameterSegment, int, error) {
	// input starts with ${
	if len(input) < 3 || input[0] != '$' || input[1] != '{' {
		return nil, 0, fmt.Errorf("invalid parameter expansion")
	}

	i := 2 // Skip ${
	start := i

	// Parse identifier: [a-zA-Z_][a-zA-Z0-9_.-]*
	if i >= len(input) || !(unicode.IsLetter(rune(input[i])) || input[i] == '_') {
		return nil, 0, fmt.Errorf("invalid parameter name")
	}

	i++
	for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) ||
		input[i] == '_' || input[i] == '-' || input[i] == '.') {
		i++
	}

	name := input[start:i]

	// Expect closing }
	if i >= len(input) || input[i] != '}' {
		return nil, 0, fmt.Errorf("unclosed parameter expansion")
	}

	return &ParameterSegment{Name: name}, i + 1, nil
}

// parseCommand parses a command expansion !`command`
func parseCommand(input string) (*CommandSegment, int, error) {
	// input starts with !`
	if len(input) < 3 || input[0] != '!' || input[1] != '`' {
		return nil, 0, fmt.Errorf("invalid command expansion")
	}

	i := 2 // Skip !`
	start := i

	// Find closing backtick
	for i < len(input) && input[i] != '`' {
		i++
	}

	if i >= len(input) {
		return nil, 0, fmt.Errorf("unclosed command expansion")
	}

	command := input[start:i]
	return &CommandSegment{Command: command}, i + 1, nil
}

// parseFile parses a file expansion @path
func parseFile(input string) (*FileSegment, int) {
	// input starts with @
	i := 1 // Skip @
	start := i

	// Parse path: continues until unescaped whitespace or end of string
	for i < len(input) {
		if input[i] == '\\' && i+1 < len(input) && input[i+1] == ' ' {
			// Escaped space - include it
			i += 2
		} else if unicode.IsSpace(rune(input[i])) {
			// Unescaped whitespace - stop
			break
		} else if input[i] == '$' || input[i] == '!' || input[i] == '@' {
			// Stop before another expansion
			break
		} else {
			i++
		}
	}

	path := input[start:i]
	return &FileSegment{Path: path}, i
}

// ExpandString is a convenience function to parse and expand a string
func ExpandString(input string, params map[string]string, logger *slog.Logger) (string, error) {
	segments, err := parseString(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse string: %w", err)
	}

	var result strings.Builder
	for _, segment := range segments {
		result.WriteString(segment.Expand(params, logger))
	}

	return result.String(), nil
}
