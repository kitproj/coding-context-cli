package codingcontext

import (
	"reflect"
	"testing"
)

// Helper functions to create pointers for test expectations
func strPtr(s string) *string { return &s }

func termValue(s string) *Value {
	return &Value{Term: strPtr(s)}
}

func stringValue(s string) *Value {
	return &Value{String: strPtr(s)}
}

func positionalArg(value string) *Argument {
	return &Argument{Value: termValue(value)}
}

func positionalStringArg(value string) *Argument {
	return &Argument{Value: stringValue(value)}
}

func namedArg(key, value string) *Argument {
	return &Argument{Key: strPtr(key), Value: termValue(value)}
}

func namedStringArg(key, value string) *Argument {
	return &Argument{Key: strPtr(key), Value: stringValue(value)}
}

func TestParseTask_SlashCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Input
		wantErr bool
	}{
		{
			name:  "simple command without arguments",
			input: "/fix-bug\n",
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{Name: "fix-bug", Arguments: nil}},
				},
			},
		},
		{
			name:  "command with single positional argument",
			input: "/fix-bug 123\n",
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{
						Name:      "fix-bug",
						Arguments: []*Argument{positionalArg("123")},
					}},
				},
			},
		},
		{
			name:  "command with multiple positional arguments",
			input: "/implement-feature login high urgent\n",
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{
						Name: "implement-feature",
						Arguments: []*Argument{
							positionalArg("login"),
							positionalArg("high"),
							positionalArg("urgent"),
						},
					}},
				},
			},
		},
		{
			name:  "command with quoted argument",
			input: "/code-review \"Fix authentication bug\"\n",
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{
						Name:      "code-review",
						Arguments: []*Argument{positionalStringArg("Fix authentication bug")},
					}},
				},
			},
		},
		{
			name:  "command with named parameter",
			input: "/fix-bug issue=PROJ-123\n",
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{
						Name:      "fix-bug",
						Arguments: []*Argument{namedArg("issue", "PROJ-123")},
					}},
				},
			},
		},
		{
			name:  "command with named parameter with quoted value",
			input: "/implement feature=\"Add user auth\"\n",
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{
						Name:      "implement",
						Arguments: []*Argument{namedStringArg("feature", "Add user auth")},
					}},
				},
			},
		},
		{
			name:  "command with mixed positional and named arguments",
			input: "/task arg1 key=value arg2\n",
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{
						Name: "task",
						Arguments: []*Argument{
							positionalArg("arg1"),
							namedArg("key", "value"),
							positionalArg("arg2"),
						},
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTask() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseTask_Text(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Input
		wantErr bool
	}{
		{
			name:  "simple text",
			input: "This is a simple task\n",
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"This", " ", "is", " ", "a", " ", "simple", " ", "task", "\n"}}},
				},
			},
		},
		{
			name:  "multiline text",
			input: "Line 1\nLine 2\nLine 3\n",
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"Line", " ", "1", "\n", "Line", " ", "2", "\n", "Line", " ", "3", "\n"}}},
				},
			},
		},
		{
			name:  "text with slash not at line start",
			input: "Check the file path\n",
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"Check", " ", "the", " ", "file", " ", "path", "\n"}}},
				},
			},
		},
		{
			name:  "text with equals sign",
			input: "x = y + z\n",
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"x", " ", "=", " ", "y", " ", "+", " ", "z", "\n"}}},
				},
			},
		},
		{
			name:  "text with quoted string",
			input: "The message is \"Hello World\"\n",
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"The", " ", "message", " ", "is", " ", "Hello World", "\n"}}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTask() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseTask_MixedBlocks(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Input
		wantErr bool
	}{
		{
			name: "text followed by command",
			input: `Please fix the bug
/fix-bug 123
`,
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"Please", " ", "fix", " ", "the", " ", "bug", "\n"}}},
					{SlashCommand: &SlashCommand{Name: "fix-bug", Arguments: []*Argument{positionalArg("123")}}},
				},
			},
		},
		{
			name: "command followed by text",
			input: `/fix-bug 123
This is additional context
`,
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{Name: "fix-bug", Arguments: []*Argument{positionalArg("123")}}},
					{Text: &Text{Content: []string{"This", " ", "is", " ", "additional", " ", "context", "\n"}}},
				},
			},
		},
		{
			name: "multiple commands with text between",
			input: `/init project
Set up the environment
/config name=myapp
Additional settings here
`,
			want: &Input{
				Blocks: []*Block{
					{SlashCommand: &SlashCommand{Name: "init", Arguments: []*Argument{positionalArg("project")}}},
					{Text: &Text{Content: []string{"Set", " ", "up", " ", "the", " ", "environment", "\n"}}},
					{SlashCommand: &SlashCommand{Name: "config", Arguments: []*Argument{namedArg("name", "myapp")}}},
					{Text: &Text{Content: []string{"Additional", " ", "settings", " ", "here", "\n"}}},
				},
			},
		},
		{
			name:    "command at start with no trailing newline",
			input:   `/task arg`,
			wantErr: true, // Commands require a newline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTask() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseTask_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Input
		wantErr bool
	}{
		{
			name:  "empty input",
			input: "",
			want: &Input{
				Blocks: nil,
			},
		},
		{
			name:  "just newlines",
			input: "\n\n\n",
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"\n\n\n"}}},
				},
			},
		},
		{
			name:  "just whitespace",
			input: "   \t  \n",
			want: &Input{
				Blocks: []*Block{
					{Text: &Text{Content: []string{"   \t  ", "\n"}}},
				},
			},
		},
		{
			name:    "slash at start with no command name",
			input:   "/\n",
			wantErr: true, // Should error as there's no command name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTask() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
