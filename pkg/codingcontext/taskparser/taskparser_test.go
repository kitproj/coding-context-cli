package taskparser

import (
	"strings"
	"testing"
)

func TestParseTask(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, task Task)
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 0 {
					t.Errorf("expected empty task, got %d blocks", len(task))
				}
			},
		},
		{
			name:    "simple text block",
			input:   "This is a simple text block.",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 1 {
					t.Fatalf("expected 1 block, got %d", len(task))
				}
				if task[0].Text == nil {
					t.Fatal("expected text block")
				}
				if task[0].Text.Content() != "This is a simple text block." {
					t.Errorf("expected 'This is a simple text block.', got %q", task[0].Text.Content())
				}
			},
		},
		{
			name:    "simple slash command without arguments",
			input:   "/fix-bug\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
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
			},
		},
		{
			name:    "slash command with positional arguments",
			input:   "/fix-bug 123 urgent\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
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
			},
		},
		{
			name:    "slash command with quoted argument",
			input:   "/fix-bug \"Fix authentication bug\"\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
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
				// The parser captures the quotes as part of the String token
				expectedValue := `"Fix authentication bug"`
				if cmd.Arguments[0].Value != expectedValue {
					t.Errorf("expected argument %q, got %q", expectedValue, cmd.Arguments[0].Value)
				}
			},
		},
		{
			name:    "slash command with named argument",
			input:   "/deploy env=\"production\"\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
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
				expectedValue := `"production"`
				if cmd.Arguments[0].Value != expectedValue {
					t.Errorf("expected value %q, got %q", expectedValue, cmd.Arguments[0].Value)
				}
			},
		},
		{
			name:    "slash command with mixed positional and named arguments",
			input:   "/task arg1 key=\"value\" arg2\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
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
					t.Errorf("expected named arg key='key', value='\"value\"', got key=%q, value=%q", cmd.Arguments[1].Key, cmd.Arguments[1].Value)
				}
				if cmd.Arguments[2].Key != "" || cmd.Arguments[2].Value != "arg2" {
					t.Errorf("expected positional arg 'arg2', got key=%q, value=%q", cmd.Arguments[2].Key, cmd.Arguments[2].Value)
				}
			},
		},
		{
			name:    "text block followed by slash command",
			input:   "Some text here\n/fix-bug 123\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 2 {
					t.Fatalf("expected 2 blocks, got %d", len(task))
				}
				if task[0].Text == nil {
					t.Fatal("expected first block to be text")
				}
				if task[1].SlashCommand == nil {
					t.Fatal("expected second block to be slash command")
				}
			},
		},
		{
			name:    "slash command followed by text block",
			input:   "/fix-bug 123\nSome text after command",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 2 {
					t.Fatalf("expected 2 blocks, got %d", len(task))
				}
				if task[0].SlashCommand == nil {
					t.Fatal("expected first block to be slash command")
				}
				if task[1].Text == nil {
					t.Fatal("expected second block to be text")
				}
			},
		},
		{
			name:    "multiple slash commands",
			input:   "/command1 arg1\n/command2 arg2\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 2 {
					t.Fatalf("expected 2 blocks, got %d", len(task))
				}
				if task[0].SlashCommand == nil || task[0].SlashCommand.Name != "command1" {
					t.Fatal("expected first block to be command1")
				}
				if task[1].SlashCommand == nil || task[1].SlashCommand.Name != "command2" {
					t.Fatal("expected second block to be command2")
				}
			},
		},
		{
			name:    "text with inline slash (not at line start)",
			input:   "This is text with a /slash in the middle.",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 1 {
					t.Fatalf("expected 1 block, got %d", len(task))
				}
				if task[0].Text == nil {
					t.Fatal("expected text block")
				}
				// The inline slash should be part of the text
				if !strings.Contains(task[0].Text.Content(), "/slash") {
					t.Errorf("expected text to contain '/slash', got %q", task[0].Text.Content())
				}
			},
		},
		{
			name:    "non-whitespace before slash prevents command",
			input:   "text/deploy env=\"production\"\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 1 {
					t.Fatalf("expected 1 block, got %d", len(task))
				}
				if task[0].Text == nil {
					t.Fatal("expected text block, not command")
				}
				// The slash should be part of the text, not a command
				if task[0].SlashCommand != nil {
					t.Fatal("expected no slash command when non-whitespace precedes slash")
				}
				if !strings.Contains(task[0].Text.Content(), "/deploy") {
					t.Errorf("expected text to contain '/deploy', got %q", task[0].Text.Content())
				}
			},
		},
		{
			name:    "text with equals sign",
			input:   "This is text with key=value pairs.",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 1 {
					t.Fatalf("expected 1 block, got %d", len(task))
				}
				if task[0].Text == nil {
					t.Fatal("expected text block")
				}
			},
		},
		{
			name:    "multiline text preserves whitespace",
			input:   "Line 1\n  Indented line 2\nLine 3",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 1 {
					t.Fatalf("expected 1 block, got %d", len(task))
				}
				if task[0].Text == nil {
					t.Fatal("expected text block")
				}
				// Check that whitespace is preserved
				expected := "Line 1\n  Indented line 2\nLine 3"
				if task[0].Text.Content() != expected {
					t.Errorf("expected %q, got %q", expected, task[0].Text.Content())
				}
			},
		},
		{
			name:    "complex mixed content",
			input:   "Introduction text\n/command1 arg1 key=\"value\"\nMiddle text\n/command2\nEnding text",
			wantErr: false,
			check: func(t *testing.T, task Task) {
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestSlashCommand_Params(t *testing.T) {
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
			task, err := ParseTask(tt.input)
			if err != nil {
				t.Fatalf("ParseTask() error = %v", err)
			}

			// Check for SlashCommand at the expected index
			if len(task) <= tt.commandIndex {
				t.Fatalf("expected at least %d blocks, got %d", tt.commandIndex+1, len(task))
			}
			if task[tt.commandIndex].SlashCommand == nil {
				t.Fatalf("expected slash command block at index %d", tt.commandIndex)
			}

			cmd := task[tt.commandIndex].SlashCommand
			if cmd.Name != tt.expectedName {
				t.Errorf("expected command name %q, got %q", tt.expectedName, cmd.Name)
			}

			// Use Params() to validate expectations
			params := cmd.Params()

			// Validate positional arguments
			actualArgs := params.Arguments()
			if len(tt.expectedArgs) != len(actualArgs) {
				t.Errorf("expected %d positional arguments, got %d: expected=%v, got=%v",
					len(tt.expectedArgs), len(actualArgs), tt.expectedArgs, actualArgs)
			} else {
				for i, expected := range tt.expectedArgs {
					if i < len(actualArgs) && actualArgs[i] != expected {
						t.Errorf("positional arg[%d]: expected %q, got %q", i, expected, actualArgs[i])
					}
				}
			}

			// Validate named parameters
			for key, expectedValues := range tt.expectedParams {
				if key == ArgumentsKey {
					continue // Already validated above
				}
				actualValues := params.Values(key)
				if len(expectedValues) != len(actualValues) {
					t.Errorf("key %q: expected %d values, got %d: expected=%v, got=%v",
						key, len(expectedValues), len(actualValues), expectedValues, actualValues)
				} else {
					for i, expected := range expectedValues {
						if i < len(actualValues) && actualValues[i] != expected {
							t.Errorf("key %q[%d]: expected %q, got %q", key, i, expected, actualValues[i])
						}
					}
				}
			}

			// Verify no unexpected keys (except ArgumentsKey which we handle separately)
			for key := range params {
				if key != ArgumentsKey {
					if _, exists := tt.expectedParams[key]; !exists {
						t.Errorf("unexpected key in params: %q", key)
					}
				}
			}
		})
	}
}
