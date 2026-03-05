package taskparser

import (
	"strings"
	"testing"
)

type parseTaskCase struct {
	name    string
	input   string
	wantErr bool
	check   func(t *testing.T, task Task)
}

func checkEmptyTask(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 0 {
		t.Errorf("expected empty task, got %d blocks", len(task))
	}
}

func checkSimpleTextBlock(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	if task[0].Text == nil {
		t.Fatal("expected text block")
	}

	if task[0].Text.Content() != "This is a simple text block." {
		t.Errorf("expected 'This is a simple text block.', got %q", task[0].Text.Content())
	}
}

func checkSimpleSlashNoArgs(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	if task[0].SlashCommand == nil {
		t.Fatal("expected slash command block")
	}

	if task[0].SlashCommand.Name != "fix-bug" {
		t.Errorf("expected name 'fix-bug', got %q", task[0].SlashCommand.Name)
	}

	if len(task[0].SlashCommand.Arguments) != 0 {
		t.Errorf("expected no arguments, got %d", len(task[0].SlashCommand.Arguments))
	}
}

func checkSlashWithTwoArgs(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	cmd := task[0].SlashCommand
	if cmd == nil {
		t.Fatal("expected slash command block")
	}

	if cmd.Name != "fix-bug" {
		t.Errorf("expected name 'fix-bug', got %q", cmd.Name)
	}

	if len(cmd.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(cmd.Arguments))
	}

	if cmd.Arguments[0].Value != "123" {
		t.Errorf("expected first arg '123', got %q", cmd.Arguments[0].Value)
	}

	if cmd.Arguments[1].Value != "urgent" {
		t.Errorf("expected second arg 'urgent', got %q", cmd.Arguments[1].Value)
	}
}

func checkSlashWithQuotedArg(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	cmd := task[0].SlashCommand
	if cmd == nil {
		t.Fatal("expected slash command block")
	}

	if len(cmd.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(cmd.Arguments))
	}

	expectedValue := `"Fix authentication bug"`
	if cmd.Arguments[0].Value != expectedValue {
		t.Errorf("expected argument %q, got %q", expectedValue, cmd.Arguments[0].Value)
	}
}

func checkSlashWithNamedArg(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	cmd := task[0].SlashCommand
	if cmd == nil {
		t.Fatal("expected slash command block")
	}

	if len(cmd.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(cmd.Arguments))
	}

	if cmd.Arguments[0].Key != "env" {
		t.Errorf("expected key 'env', got %q", cmd.Arguments[0].Key)
	}

	if cmd.Arguments[0].Value != `"production"` {
		t.Errorf("expected value %q, got %q", `"production"`, cmd.Arguments[0].Value)
	}
}

func checkSlashMixedArgs(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	cmd := task[0].SlashCommand
	if cmd == nil {
		t.Fatal("expected slash command block")
	}

	if len(cmd.Arguments) != 3 {
		t.Fatalf("expected 3 arguments, got %d", len(cmd.Arguments))
	}

	if cmd.Arguments[0].Key != "" || cmd.Arguments[0].Value != "arg1" {
		t.Errorf("expected positional arg 'arg1', got key=%q, value=%q", cmd.Arguments[0].Key, cmd.Arguments[0].Value)
	}

	if cmd.Arguments[1].Key != "key" || cmd.Arguments[1].Value != `"value"` {
		t.Errorf("expected named arg key='key', value='\"value\"', got key=%q, value=%q",
			cmd.Arguments[1].Key, cmd.Arguments[1].Value)
	}

	if cmd.Arguments[2].Key != "" || cmd.Arguments[2].Value != "arg2" {
		t.Errorf("expected positional arg 'arg2', got key=%q, value=%q", cmd.Arguments[2].Key, cmd.Arguments[2].Value)
	}
}

func checkTextThenSlash(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(task))
	}

	if task[0].Text == nil {
		t.Fatal("expected first block to be text")
	}

	if task[1].SlashCommand == nil {
		t.Fatal("expected second block to be slash command")
	}
}

func checkSlashThenText(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(task))
	}

	if task[0].SlashCommand == nil {
		t.Fatal("expected first block to be slash command")
	}

	if task[1].Text == nil {
		t.Fatal("expected second block to be text")
	}
}

func checkTwoSlashCommands(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(task))
	}

	if task[0].SlashCommand == nil || task[0].SlashCommand.Name != "command1" {
		t.Fatal("expected first block to be command1")
	}

	if task[1].SlashCommand == nil || task[1].SlashCommand.Name != "command2" {
		t.Fatal("expected second block to be command2")
	}
}

func checkTextWithInlineSlash(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	if task[0].Text == nil {
		t.Fatal("expected text block")
	}

	if !strings.Contains(task[0].Text.Content(), "/slash") {
		t.Errorf("expected text to contain '/slash', got %q", task[0].Text.Content())
	}
}

func checkNonWhitespaceBeforeSlash(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	if task[0].Text == nil {
		t.Fatal("expected text block, not command")
	}

	if task[0].SlashCommand != nil {
		t.Fatal("expected no slash command when non-whitespace precedes slash")
	}

	if !strings.Contains(task[0].Text.Content(), "/deploy") {
		t.Errorf("expected text to contain '/deploy', got %q", task[0].Text.Content())
	}
}

func checkSingleTextBlockOnly(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	if task[0].Text == nil {
		t.Fatal("expected text block")
	}
}

func checkMultilineTextContent(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 1 {
		t.Fatalf("expected 1 block, got %d", len(task))
	}

	if task[0].Text == nil {
		t.Fatal("expected text block")
	}

	expected := "Line 1\n  Indented line 2\nLine 3"
	if task[0].Text.Content() != expected {
		t.Errorf("expected %q, got %q", expected, task[0].Text.Content())
	}
}

func checkComplexMixedContent(t *testing.T, task Task) {
	t.Helper()

	if len(task) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(task))
	}

	if task[0].Text == nil {
		t.Fatal("expected block 0 to be text")
	}

	if task[1].SlashCommand == nil || task[1].SlashCommand.Name != "command1" {
		t.Fatal("expected block 1 to be command1")
	}

	if task[2].Text == nil {
		t.Fatal("expected block 2 to be text")
	}

	if task[3].SlashCommand == nil || task[3].SlashCommand.Name != "command2" {
		t.Fatal("expected block 3 to be command2")
	}

	if task[4].Text == nil {
		t.Fatal("expected block 4 to be text")
	}
}

// ==================== Markdown-aware parsing tests ====================

func checkNoSlashCommands(t *testing.T, task Task) {
	t.Helper()

	for _, b := range task {
		if b.SlashCommand != nil {
			t.Errorf("expected no slash commands, but found /%s", b.SlashCommand.Name)
		}
	}
}

func checkRoundTrip(input string) func(t *testing.T, task Task) {
	return func(t *testing.T, task Task) {
		t.Helper()

		got := task.String()
		if got != input {
			t.Errorf("round-trip failed:\ninput: %q\ngot:   %q", input, got)
		}
	}
}

func checkCommandsOutsideCodeBlock(names ...string) func(t *testing.T, task Task) {
	return func(t *testing.T, task Task) {
		t.Helper()

		var got []string

		for _, b := range task {
			if b.SlashCommand != nil {
				got = append(got, b.SlashCommand.Name)
			}
		}

		if len(got) != len(names) {
			t.Fatalf("expected commands %v, got %v", names, got)
		}

		for i, name := range names {
			if got[i] != name {
				t.Errorf("command[%d]: expected %q, got %q", i, name, got[i])
			}
		}
	}
}

func markdownAwareTestCases() []parseTaskCase {
	return []parseTaskCase{
		{
			// Slash on its own line inside a fenced code block must NOT be treated as a command.
			name:  "fenced code block: slash inside not a command",
			input: "Some text\n```\n/not-a-command\n```\n\n/actual-command\n",
			check: checkCommandsOutsideCodeBlock("actual-command"),
		},
		{
			// Round-trip: original content is reconstructed exactly from the task blocks.
			name:  "fenced code block: round-trip preserves content",
			input: "Some text\n```\n/not-a-command\n```\n\n/actual-command\n",
			check: checkRoundTrip("Some text\n```\n/not-a-command\n```\n\n/actual-command\n"),
		},
		{
			// Language-tagged fenced code block.
			name:  "fenced code block with language tag",
			input: "# Heading\n\n```bash\n/inside-code\necho hello\n```\n\n/real-command\n",
			check: checkCommandsOutsideCodeBlock("real-command"),
		},
		{
			// Indented code block (4 spaces): content lines start with spaces then /,
			// which would otherwise match the slash command grammar.
			name:  "indented code block: slash inside not a command",
			input: "Text before\n\n    /indented-command\n    more code\n\n/real-command\n",
			check: checkCommandsOutsideCodeBlock("real-command"),
		},
		{
			// Indented code block round-trip.
			name:  "indented code block: round-trip preserves content",
			input: "Text before\n\n    /indented-command\n    more code\n\n/real-command\n",
			check: checkRoundTrip("Text before\n\n    /indented-command\n    more code\n\n/real-command\n"),
		},
		{
			// Multiple fenced code blocks with slash commands between them.
			name:  "multiple fenced code blocks with commands between",
			input: "```\n/inside-first\n```\n/between-blocks\n```\n/inside-second\n```\n",
			check: checkCommandsOutsideCodeBlock("between-blocks"),
		},
		{
			// Fenced code block at the very start of the content.
			name:  "fenced code block at start",
			input: "```\n/code-only\n```\n",
			check: checkNoSlashCommands,
		},
		{
			// Fenced code block at the very end (no trailing real commands).
			name:  "fenced code block at end, no commands after",
			input: "/real-command\n```\n/code-only\n```\n",
			check: checkCommandsOutsideCodeBlock("real-command"),
		},
		{
			// Empty fenced code block (no content lines).
			name:  "empty fenced code block",
			input: "```\n```\n/real-command\n",
			check: checkCommandsOutsideCodeBlock("real-command"),
		},
		{
			// Inline code (backtick) with slash: already safe because the line starts with `
			// but we verify that it does not create a false command.
			name:  "inline code span with slash is not a command",
			input: "Use `\\/path` or `/other` in the text.\n",
			check: checkNoSlashCommands,
		},
		{
			// HTML block: slash inside HTML block should not be a command.
			name:  "HTML block: slash inside not a command",
			input: "<!-- /not-a-command -->\n\n/real-command\n",
			check: checkCommandsOutsideCodeBlock("real-command"),
		},
		{
			// Block quote: the grammar already treats > at line start as text,
			// but verify the behavior is correct.
			name:  "block quote with slash is not a command",
			input: "> /quoted-line\n\n/real-command\n",
			check: checkCommandsOutsideCodeBlock("real-command"),
		},
		{
			// Content with no code blocks behaves the same as before.
			name:  "no code blocks: slash commands still detected normally",
			input: "Some intro text\n/deploy env=\"prod\"\nSome outro text\n",
			check: checkCommandsOutsideCodeBlock("deploy"),
		},
		{
			// Fenced code block whose content contains a command-like line with arguments.
			name:  "fenced code block with command-like line including arguments",
			input: "```sh\n/deploy env=\"production\" region=\"us-east-1\"\n```\n\n/real env=\"staging\"\n",
			check: checkCommandsOutsideCodeBlock("real"),
		},
		{
			// Nested code blocks are not valid markdown, but indented fenced blocks
			// inside a list item should still be protected.
			name:  "fenced code block inside list item",
			input: "- item 1\n- item 2\n  ```\n  /code-in-list\n  ```\n\n/real-command\n",
			check: checkCommandsOutsideCodeBlock("real-command"),
		},
		{
			// Plain text without code blocks: existing parser behavior is unchanged.
			name:  "plain text no code blocks",
			input: "This is plain text without any code blocks.",
			check: func(t *testing.T, task Task) {
				t.Helper()

				if len(task) != 1 || task[0].Text == nil {
					t.Fatal("expected a single text block")
				}

				if task[0].Text.Content() != "This is plain text without any code blocks." {
					t.Errorf("unexpected content: %q", task[0].Text.Content())
				}
			},
		},
	}
}

func TestParseTask_MarkdownAware(t *testing.T) {
	t.Parallel()

	tests := markdownAwareTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, task)
			}
		})
	}
}

func TestParseTask(t *testing.T) {
	t.Parallel()

	tests := []parseTaskCase{
		{name: "empty string", input: "", check: checkEmptyTask},
		{name: "single newline", input: "\n", check: checkEmptyTask},
		{name: "multiple newlines", input: "\n\n\n", check: checkEmptyTask},
		{name: "whitespace only", input: "   \t  \n  \n", check: checkEmptyTask},
		{name: "simple text block", input: "This is a simple text block.", check: checkSimpleTextBlock},
		{name: "simple slash command without arguments", input: "/fix-bug\n", check: checkSimpleSlashNoArgs},
		{name: "slash command with positional arguments", input: "/fix-bug 123 urgent\n", check: checkSlashWithTwoArgs},
		{name: "slash command with quoted argument",
			input: "/fix-bug \"Fix authentication bug\"\n", check: checkSlashWithQuotedArg},
		{name: "slash command with named argument",
			input: "/deploy env=\"production\"\n", check: checkSlashWithNamedArg},
		{name: "slash command with mixed positional and named arguments",
			input: "/task arg1 key=\"value\" arg2\n", check: checkSlashMixedArgs},
		{name: "text block followed by slash command",
			input: "Some text here\n/fix-bug 123\n", check: checkTextThenSlash},
		{name: "slash command followed by text block",
			input: "/fix-bug 123\nSome text after command", check: checkSlashThenText},
		{name: "multiple slash commands",
			input: "/command1 arg1\n/command2 arg2\n", check: checkTwoSlashCommands},
		{name: "text with inline slash (not at line start)",
			input: "This is text with a /slash in the middle.", check: checkTextWithInlineSlash},
		{name: "non-whitespace before slash prevents command",
			input: "text/deploy env=\"production\"\n", check: checkNonWhitespaceBeforeSlash},
		{name: "text with equals sign",
			input: "This is text with key=value pairs.", check: checkSingleTextBlockOnly},
		{name: "multiline text preserves whitespace",
			input: "Line 1\n  Indented line 2\nLine 3", check: checkMultilineTextContent},
		{name: "complex mixed content",
			input: "Introduction text\n/command1 arg1 key=\"value\"\nMiddle text\n/command2\nEnding text",
			check: checkComplexMixedContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, task)
			}
		})
	}
}

func TestTask_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple text",
			input: "This is text.",
		},
		{
			name:  "slash command",
			input: "/command arg1 arg2\n",
		},
		{
			name:  "mixed content",
			input: "Text before\n/command arg\nText after",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task, err := ParseTask(tt.input)
			if err != nil {
				t.Fatalf("ParseTask() error = %v", err)
			}

			result := task.String()
			// The string representation should closely match the input
			// Note: exact match may not be possible due to whitespace normalization
			if result == "" && tt.input != "" {
				t.Errorf("Task.String() returned empty string, expected non-empty")
			}
		})
	}
}

//nolint:funlen
func TestSlashCommand_Params(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		commandIndex   int // Which block contains the slash command (0-based)
		expectedName   string
		expectedParams Params
		expectedArgs   []string // Positional arguments
	}{
		{
			name:         "simple named parameters",
			input:        "/deploy env=\"production\" region=\"us-east-1\" version=1.2.3\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				"env":     {"production"},
				"region":  {"us-east-1"},
				"version": {"1.2.3"},
			},
			expectedArgs: nil,
		},
		{
			name:         "whitespace before initial slash",
			input:        "  /deploy env=\"production\" region=\"us-east-1\" version=1.2.3\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				"env":     {"production"},
				"region":  {"us-east-1"},
				"version": {"1.2.3"},
			},
			expectedArgs: nil,
		},
		{
			name:         "text before slash command",
			input:        "Some introduction text\n/deploy env=\"production\"\n",
			commandIndex: 1,
			expectedName: "deploy",
			expectedParams: Params{
				"env": {"production"},
			},
			expectedArgs: nil,
		},
		{
			name:         "positional arguments only",
			input:        "/fix-bug 123 urgent\n",
			commandIndex: 0,
			expectedName: "fix-bug",
			expectedParams: Params{
				ArgumentsKey: {"123", "urgent"},
			},
			expectedArgs: []string{"123", "urgent"},
		},
		{
			name:         "mixed positional and named arguments",
			input:        "/task arg1 key=\"value\" arg2 env=\"prod\"\n",
			commandIndex: 0,
			expectedName: "task",
			expectedParams: Params{
				ArgumentsKey: {"arg1", "arg2"},
				"key":        {"value"},
				"env":        {"prod"},
			},
			expectedArgs: []string{"arg1", "arg2"},
		},
		{
			name:         "positional before named",
			input:        "/deploy arg1 arg2 env=\"production\"\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				ArgumentsKey: {"arg1", "arg2"},
				"env":        {"production"},
			},
			expectedArgs: []string{"arg1", "arg2"},
		},
		{
			name:         "positional after named",
			input:        "/deploy env=\"production\" arg1 arg2\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				ArgumentsKey: {"arg1", "arg2"},
				"env":        {"production"},
			},
			expectedArgs: []string{"arg1", "arg2"},
		},
		{
			name:         "positional between named",
			input:        "/deploy env=\"production\" arg1 region=\"us-east-1\"\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				ArgumentsKey: {"arg1"},
				"env":        {"production"},
				"region":     {"us-east-1"},
			},
			expectedArgs: []string{"arg1"},
		},
		{
			name:         "quoted positional arguments",
			input:        "/deploy \"quoted arg\" 'single quoted' normal\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				ArgumentsKey: {"quoted arg", "single quoted", "normal"},
			},
			expectedArgs: []string{"quoted arg", "single quoted", "normal"},
		},
		{
			name:         "multiple named parameters with spaces",
			input:        "/deploy env=\"production\" region=\"us-east-1\" version=1.2.3\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				"env":     {"production"},
				"region":  {"us-east-1"},
				"version": {"1.2.3"},
			},
			expectedArgs: nil,
		},
		{
			name:         "text before and after slash command",
			input:        "Before text\n/deploy env=\"production\" arg1\nAfter text",
			commandIndex: 1,
			expectedName: "deploy",
			expectedParams: Params{
				ArgumentsKey: {"arg1"},
				"env":        {"production"},
			},
			expectedArgs: []string{"arg1"},
		},
		{
			name:           "no arguments",
			input:          "/deploy\n",
			commandIndex:   0,
			expectedName:   "deploy",
			expectedParams: Params{},
			expectedArgs:   nil,
		},
		{
			name:         "single quoted named parameter",
			input:        "/deploy env='production'\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				"env": {"production"},
			},
			expectedArgs: nil,
		},
		{
			name:         "complex mixed arguments",
			input:        "/deploy arg1 env=\"production\" arg2 region=\"us-east-1\" arg3 version=1.2.3\n",
			commandIndex: 0,
			expectedName: "deploy",
			expectedParams: Params{
				ArgumentsKey: {"arg1", "arg2", "arg3"},
				"env":        {"production"},
				"region":     {"us-east-1"},
				"version":    {"1.2.3"},
			},
			expectedArgs: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:         "multiple slash commands - test second command",
			input:        "/command1 arg1\n/deploy env=\"production\" arg1 region=\"us-east-1\"\n/command3 arg3\n",
			commandIndex: 1,
			expectedName: "deploy",
			expectedParams: Params{
				ArgumentsKey: {"arg1"},
				"env":        {"production"},
				"region":     {"us-east-1"},
			},
			expectedArgs: []string{"arg1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task, err := ParseTask(tt.input)
			if err != nil {
				t.Fatalf("ParseTask() error = %v", err)
			}

			verifySlashCommandParams(t, task, tt.commandIndex, tt.expectedName, tt.expectedParams, tt.expectedArgs)
		})
	}
}

//nolint:cyclop
func verifySlashCommandParams(t *testing.T, task Task, idx int, name string, expected Params, expectedArgs []string) {
	t.Helper()

	if len(task) <= idx {
		t.Fatalf("expected at least %d blocks, got %d", idx+1, len(task))
	}

	if task[idx].SlashCommand == nil {
		t.Fatalf("expected slash command block at index %d", idx)
	}

	cmd := task[idx].SlashCommand
	if cmd.Name != name {
		t.Errorf("expected command name %q, got %q", name, cmd.Name)
	}

	params := cmd.Params()
	actualArgs := params.Arguments()

	if len(expectedArgs) != len(actualArgs) {
		t.Errorf("expected %d positional arguments, got %d: expected=%v, got=%v",
			len(expectedArgs), len(actualArgs), expectedArgs, actualArgs)
	} else {
		for i, exp := range expectedArgs {
			if i < len(actualArgs) && actualArgs[i] != exp {
				t.Errorf("positional arg[%d]: expected %q, got %q", i, exp, actualArgs[i])
			}
		}
	}

	for key, expectedValues := range expected {
		if key == ArgumentsKey {
			continue
		}

		actualValues := params.Values(key)
		if len(expectedValues) != len(actualValues) {
			t.Errorf("key %q: expected %d values, got %d: expected=%v, got=%v",
				key, len(expectedValues), len(actualValues), expectedValues, actualValues)
		} else {
			for i, exp := range expectedValues {
				if i < len(actualValues) && actualValues[i] != exp {
					t.Errorf("key %q[%d]: expected %q, got %q", key, i, exp, actualValues[i])
				}
			}
		}
	}

	for key := range params {
		if key != ArgumentsKey {
			if _, exists := expected[key]; !exists {
				t.Errorf("unexpected key in params: %q", key)
			}
		}
	}
}
