package taskparser

import (
	"strings"
)

// ParseTask parses a task string into a structured Task representation.
// A task consists of alternating blocks of text content and slash commands
// (commands starting with /). The parser distinguishes between inline
// slashes (part of text) and command slashes (at the start of a line).
//
// Task Structure:
//
//	A Task is a sequence of Block elements, where each block is either:
//	  - A Text block: regular text content
//	  - A SlashCommand block: a command starting with / at line start
//	    (optional whitespace before / is allowed)
//
// Command Detection:
//
//   - Command: A / at the start of a line (after newline or at start of input)
//     starts a command. Optional whitespace (spaces or tabs) before the /
//     is allowed and preserved.
//
//   - Text: A / not at the start of a line, or preceded by non-whitespace
//     characters, is treated as regular text
//
//   - Line boundaries: Newlines separate commands from text
//
//     Examples:
//     "/fix-bug"              // Command
//     "  /deploy"            // Command (whitespace before slash allowed)
//     "Some text /fix"       // Text (slash not at line start)
//     "text/deploy"          // Text (non-whitespace before slash prevents command)
//
// Command Arguments:
//
//	Commands can have:
//	  - Positional arguments: values without keys
//	  - Named arguments: key=value pairs
//	  - Quoted arguments: both single and double quotes supported
//	  - Mixed arguments: positional and named can be interleaved
//
//	Examples:
//	  "/fix-bug 123 urgent"                    // Positional only
//	  "/deploy env=\"production\""             // Named only
//	  "/task arg1 key=\"value\" arg2"         // Mixed
//	  "/deploy arg1 env=\"production\" arg2"  // Positional before/after named
//
//	Note: The argument parsing in ParseTask uses a simpler grammar than
//	ParseParams. For more advanced parsing (commas, escape sequences, etc.),
//	use the Params() method on SlashCommand.
//
// Text Content:
//
//	Text blocks:
//	  - Preserve whitespace: all whitespace, including indentation
//	  - Can span multiple lines
//	  - No special processing: stored as-is
//
//	Examples:
//	  "Line 1\n  Indented line 2\nLine 3"
//
// Task Composition:
//
//	Tasks can contain:
//	  - Only text: "This is a task with only text."
//	  - Only commands: "/command1\n/command2\n"
//	  - Mixed: "Introduction text\n/command1 arg1\nMiddle text\n/command2\n"
//
// SlashCommand.Params() Method:
//
//	The SlashCommand type provides a Params() method that converts command
//	arguments to a Params map using ParseParams. This provides:
//	  - Full ParseParams feature support (commas, escape sequences, etc.)
//	  - Consistent API with ParseParams
//	  - More permissive parsing than the initial grammar
//
//	Example:
//	  task, _ := ParseTask("/deploy env=\"production\" region=\"us-east-1\"\n")
//	  cmd := task[0].SlashCommand
//	  params := cmd.Params()
//	  // params["env"] = []string{"production"}
//	  // params["region"] = []string{"us-east-1"}
//
// Error Conditions:
//
//	The function rarely returns errors due to the permissive grammar.
//	Potential errors include malformed input that cannot be tokenized.
//
// Examples:
//
//	// Simple text task
//	task, _ := ParseTask("This is a simple text block.")
//	// task[0].Text.Content() == "This is a simple text block."
//
//	// Simple command
//	task, _ := ParseTask("/fix-bug\n")
//	// task[0].SlashCommand.Name == "fix-bug"
//
//	// Command with whitespace before slash (allowed)
//	task, _ := ParseTask("  /deploy env=\"production\"\n")
//	// task[0].SlashCommand.Name == "deploy"
//
//	// Command with positional arguments
//	task, _ := ParseTask("/fix-bug 123 urgent\n")
//	// task[0].SlashCommand.Arguments[0].Value == "123"
//
//	// Mixed content
//	task, _ := ParseTask("Introduction text\n/fix-bug 123\nSome text after")
//	// len(task) == 3 (text, command, text)
func ParseTask(text string) (Task, error) {
	// Handle empty or whitespace-only content gracefully
	// TrimSpace returns empty string for whitespace-only input
	if strings.TrimSpace(text) == "" {
		return Task{}, nil
	}

	input, err := parser().ParseString("", text)
	if err != nil {
		return nil, err
	}
	return Task(input.Blocks), nil
}

// Params converts the slash command's arguments into a parameter map using ParseParams.
// This provides a more permissive parser that supports commas, single quotes, and other features.
// Returns a map with:
// - "ARGUMENTS": positional arguments (values without keys)
// - named parameters: key-value pairs from key="value" or key='value' arguments
func (s *SlashCommand) Params() Params {
	// Reconstruct the arguments string from the parsed Arguments
	var argStrings []string
	for _, arg := range s.Arguments {
		if arg.Key != "" {
			// Named parameter: key="value" or key='value'
			argStrings = append(argStrings, arg.Key+"="+arg.Value)
		} else {
			// Positional parameter
			argStrings = append(argStrings, arg.Value)
		}
	}

	// Join arguments with spaces (preserving the original format)
	argsString := strings.Join(argStrings, " ")

	// Use ParseParams to parse the arguments string
	// This is more permissive and handles commas, single quotes, etc.
	params, err := ParseParams(argsString)
	if err != nil {
		// If parsing fails, return empty params
		// This should rarely happen since ParseParams handles the same format
		// that was parsed by the grammar, but we handle it gracefully
		return make(Params)
	}

	return params
}

// Content returns the text content with all lines concatenated
func (t *Text) Content() string {
	var sb strings.Builder
	// Write leading newlines first
	for _, nl := range t.LeadingNewlines {
		sb.WriteString(nl)
	}
	// Then write all the lines
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
	sb.WriteString(s.LeadingWhitespace)
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
