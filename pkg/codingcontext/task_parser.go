package codingcontext

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// TaskBlocks represents a parsed task as a sequence of blocks
type TaskBlocks struct {
	Blocks []Block `parser:"@@*"`
}

// Task is a slice of blocks for convenience
type Task []Block

// Block represents either a slash command or text content
type Block struct {
	SlashCommand *SlashCommand `parser:"@@"`
	Text         *Text         `parser:"| @@"`
}

// SlashCommand represents a command starting with "/" at the beginning of a line
type SlashCommand struct {
	Name      string     `parser:"CmdStart @Term"`
	Arguments []Argument `parser:"(Whitespace @@)*"`
}

// Argument represents either a named (key=value) or positional argument
type Argument struct {
	Key   string `parser:"(@Term Assign)?"`
	Value string `parser:"(@String | @Term)"`
}

// TextToken represents a single token in text content
type TextToken struct {
	Token string `parser:"@Term | @String | @Slash | @Assign | @Whitespace | @Newline"`
}

// Text represents non-command text content
type Text struct {
	Tokens []TextToken `parser:"@@+"`
}

// Content returns the concatenated text content
func (t *Text) Content() string {
	var result strings.Builder
	for _, token := range t.Tokens {
		result.WriteString(token.Token)
	}
	return result.String()
}

// Define the lexer with the tokens from the grammar
var taskLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "CmdStart", Pattern: `(?m)^/`},      // Matches "/" only at the start of a line
	{Name: "Slash", Pattern: `/`},              // Matches "/" elsewhere
	{Name: "Assign", Pattern: `=`},             // Matches "="
	{Name: "String", Pattern: `"(?:\\.|[^"])*"`}, // Matches quoted strings with escape sequences
	{Name: "Term", Pattern: `[^ \t\n\r/"=]+`},  // Any char except space, newline, /, ", =
	{Name: "Whitespace", Pattern: `[ \t]+`},    // Spaces and tabs (horizontal only)
	{Name: "Newline", Pattern: `[\n\r]+`},      // Newlines
})

// taskParser is the participle parser instance
var taskParser = participle.MustBuild[TaskBlocks](
	participle.Lexer(taskLexer),
	participle.Unquote("String"),
)

// ParseTask parses a task string into a Task structure
func ParseTask(text string) (Task, error) {
	taskBlocks, err := taskParser.ParseString("", text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}
	return taskBlocks.Blocks, nil
}

// taskToPrompt converts a Task into a prompt string with parameter substitution
func taskToPrompt(task Task, params map[string]string) string {
	var result strings.Builder

	for _, block := range task {
		if block.SlashCommand != nil {
			// Slash commands are removed from the final prompt
			// Their parameters were extracted during parsing
			continue
		}

		if block.Text != nil {
			// Expand parameters in text blocks
			expanded := expandParamsInText(block.Text.Content(), params)
			result.WriteString(expanded)
		}
	}

	return result.String()
}

// expandParamsInText expands ${param} placeholders in text
func expandParamsInText(text string, params map[string]string) string {
	return os.Expand(text, func(key string) string {
		if val, ok := params[key]; ok {
			return val
		}
		// If parameter doesn't exist, keep the original placeholder
		return fmt.Sprintf("${%s}", key)
	})
}
