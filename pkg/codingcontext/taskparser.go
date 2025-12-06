package codingcontext

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Input represents the entire parsed task input, consisting of multiple blocks
type Input struct {
	Blocks []*Block `@@*`
}

// Block represents either a slash command or text content
type Block struct {
	SlashCommand *SlashCommand `  @@`
	Text         *Text         `| @@`
}

// SlashCommand represents a command starting with "/" at the beginning of a line
type SlashCommand struct {
	Name      string      `CmdStart @Term`
	Arguments []*Argument `( Whitespace @@ )* Whitespace? Newline`
}

// Argument represents either a named (key=value) or positional (value) argument
type Argument struct {
	Key   *string `( @Term Assign )?`
	Value *Value  `@@`
}

// Value represents either a quoted string or a term
type Value struct {
	String *string `  @String`
	Term   *string `| @Term`
}

// Text represents multi-line text content (anything that's not a slash command)
type Text struct {
	Content []string `( @Term | @String | @Slash | @Assign | @Whitespace | @Newline )+`
}

// taskLexer defines the lexical grammar for parsing tasks
var taskLexer = lexer.MustStateful(lexer.Rules{
	"Root": {
		{Name: "CmdStart", Pattern: `(?m)^/`, Action: nil},
		{Name: "Slash", Pattern: `/`, Action: nil},
		{Name: "Assign", Pattern: `=`, Action: nil},
		{Name: "String", Pattern: `"(?:[^"\\]|\\.)*"`, Action: nil},
		{Name: "Term", Pattern: `[^ \t\n\r/"=]+`, Action: nil},
		{Name: "Whitespace", Pattern: `[ \t]+`, Action: nil},
		{Name: "Newline", Pattern: `[\n\r]+`, Action: nil},
	},
})

// TaskParser is the global parser instance for parsing task inputs
var TaskParser = participle.MustBuild[Input](
	participle.Lexer(taskLexer),
	participle.Unquote("String"),
	participle.UseLookahead(2),
)

// ParseTask parses a task string into an Input structure
func ParseTask(input string) (*Input, error) {
	return TaskParser.ParseString("", input)
}
