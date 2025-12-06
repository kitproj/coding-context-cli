package codingcontext

import (
	"strings"
	"testing"
)

func TestParseTask_SlashCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantBlocks  int
		validateCmd func(t *testing.T, cmd *SlashCommand)
	}{
		{
			name:       "simple command without arguments",
			input:      "/fix-bug\n",
			wantErr:    false,
			wantBlocks: 1,
			validateCmd: func(t *testing.T, cmd *SlashCommand) {
				if cmd == nil {
					t.Fatal("expected SlashCommand, got nil")
				}
				if cmd.Name != "fix-bug" {
					t.Errorf("Name = %q, want %q", cmd.Name, "fix-bug")
				}
				if len(cmd.Arguments) != 0 {
					t.Errorf("Arguments length = %d, want 0", len(cmd.Arguments))
				}
			},
		},
		{
			name:       "command with single positional argument",
			input:      "/fix-bug 123\n",
			wantErr:    false,
			wantBlocks: 1,
			validateCmd: func(t *testing.T, cmd *SlashCommand) {
				if cmd == nil {
					t.Fatal("expected SlashCommand, got nil")
				}
				if cmd.Name != "fix-bug" {
					t.Errorf("Name = %q, want %q", cmd.Name, "fix-bug")
				}
				if len(cmd.Arguments) != 1 {
					t.Fatalf("Arguments length = %d, want 1", len(cmd.Arguments))
				}
				if cmd.Arguments[0].Key != nil {
					t.Errorf("Arguments[0].Key = %v, want nil", cmd.Arguments[0].Key)
				}
				if cmd.Arguments[0].Value == nil || cmd.Arguments[0].Value.Term == nil {
					t.Fatal("Arguments[0].Value.Term is nil")
				}
				if *cmd.Arguments[0].Value.Term != "123" {
					t.Errorf("Arguments[0].Value.Term = %q, want %q", *cmd.Arguments[0].Value.Term, "123")
				}
			},
		},
		{
			name:       "command with multiple positional arguments",
			input:      "/implement-feature login high urgent\n",
			wantErr:    false,
			wantBlocks: 1,
			validateCmd: func(t *testing.T, cmd *SlashCommand) {
				if cmd == nil {
					t.Fatal("expected SlashCommand, got nil")
				}
				if cmd.Name != "implement-feature" {
					t.Errorf("Name = %q, want %q", cmd.Name, "implement-feature")
				}
				if len(cmd.Arguments) != 3 {
					t.Fatalf("Arguments length = %d, want 3", len(cmd.Arguments))
				}
				wantArgs := []string{"login", "high", "urgent"}
				for i, want := range wantArgs {
					if cmd.Arguments[i].Value == nil || cmd.Arguments[i].Value.Term == nil {
						t.Fatalf("Arguments[%d].Value.Term is nil", i)
					}
					if *cmd.Arguments[i].Value.Term != want {
						t.Errorf("Arguments[%d].Value.Term = %q, want %q", i, *cmd.Arguments[i].Value.Term, want)
					}
				}
			},
		},
		{
			name:       "command with quoted argument",
			input:      "/code-review \"Fix authentication bug\"\n",
			wantErr:    false,
			wantBlocks: 1,
			validateCmd: func(t *testing.T, cmd *SlashCommand) {
				if cmd == nil {
					t.Fatal("expected SlashCommand, got nil")
				}
				if cmd.Name != "code-review" {
					t.Errorf("Name = %q, want %q", cmd.Name, "code-review")
				}
				if len(cmd.Arguments) != 1 {
					t.Fatalf("Arguments length = %d, want 1", len(cmd.Arguments))
				}
				if cmd.Arguments[0].Value == nil || cmd.Arguments[0].Value.String == nil {
					t.Fatal("Arguments[0].Value.String is nil")
				}
				if *cmd.Arguments[0].Value.String != "Fix authentication bug" {
					t.Errorf("Arguments[0].Value.String = %q, want %q", *cmd.Arguments[0].Value.String, "Fix authentication bug")
				}
			},
		},
		{
			name:       "command with named parameter",
			input:      "/fix-bug issue=PROJ-123\n",
			wantErr:    false,
			wantBlocks: 1,
			validateCmd: func(t *testing.T, cmd *SlashCommand) {
				if cmd == nil {
					t.Fatal("expected SlashCommand, got nil")
				}
				if cmd.Name != "fix-bug" {
					t.Errorf("Name = %q, want %q", cmd.Name, "fix-bug")
				}
				if len(cmd.Arguments) != 1 {
					t.Fatalf("Arguments length = %d, want 1", len(cmd.Arguments))
				}
				if cmd.Arguments[0].Key == nil {
					t.Fatal("Arguments[0].Key is nil")
				}
				if *cmd.Arguments[0].Key != "issue" {
					t.Errorf("Arguments[0].Key = %q, want %q", *cmd.Arguments[0].Key, "issue")
				}
				if cmd.Arguments[0].Value == nil || cmd.Arguments[0].Value.Term == nil {
					t.Fatal("Arguments[0].Value.Term is nil")
				}
				if *cmd.Arguments[0].Value.Term != "PROJ-123" {
					t.Errorf("Arguments[0].Value.Term = %q, want %q", *cmd.Arguments[0].Value.Term, "PROJ-123")
				}
			},
		},
		{
			name:       "command with named parameter with quoted value",
			input:      "/implement feature=\"Add user auth\"\n",
			wantErr:    false,
			wantBlocks: 1,
			validateCmd: func(t *testing.T, cmd *SlashCommand) {
				if cmd == nil {
					t.Fatal("expected SlashCommand, got nil")
				}
				if cmd.Name != "implement" {
					t.Errorf("Name = %q, want %q", cmd.Name, "implement")
				}
				if len(cmd.Arguments) != 1 {
					t.Fatalf("Arguments length = %d, want 1", len(cmd.Arguments))
				}
				if cmd.Arguments[0].Key == nil {
					t.Fatal("Arguments[0].Key is nil")
				}
				if *cmd.Arguments[0].Key != "feature" {
					t.Errorf("Arguments[0].Key = %q, want %q", *cmd.Arguments[0].Key, "feature")
				}
				if cmd.Arguments[0].Value == nil || cmd.Arguments[0].Value.String == nil {
					t.Fatal("Arguments[0].Value.String is nil")
				}
				if *cmd.Arguments[0].Value.String != "Add user auth" {
					t.Errorf("Arguments[0].Value.String = %q, want %q", *cmd.Arguments[0].Value.String, "Add user auth")
				}
			},
		},
		{
			name:       "command with mixed positional and named arguments",
			input:      "/task arg1 key=value arg2\n",
			wantErr:    false,
			wantBlocks: 1,
			validateCmd: func(t *testing.T, cmd *SlashCommand) {
				if cmd == nil {
					t.Fatal("expected SlashCommand, got nil")
				}
				if cmd.Name != "task" {
					t.Errorf("Name = %q, want %q", cmd.Name, "task")
				}
				if len(cmd.Arguments) != 3 {
					t.Fatalf("Arguments length = %d, want 3", len(cmd.Arguments))
				}
				// arg1 - positional
				if cmd.Arguments[0].Key != nil {
					t.Errorf("Arguments[0].Key = %v, want nil", cmd.Arguments[0].Key)
				}
				if cmd.Arguments[0].Value == nil || cmd.Arguments[0].Value.Term == nil {
					t.Fatal("Arguments[0].Value.Term is nil")
				}
				if *cmd.Arguments[0].Value.Term != "arg1" {
					t.Errorf("Arguments[0].Value.Term = %q, want %q", *cmd.Arguments[0].Value.Term, "arg1")
				}
				// key=value - named
				if cmd.Arguments[1].Key == nil {
					t.Fatal("Arguments[1].Key is nil")
				}
				if *cmd.Arguments[1].Key != "key" {
					t.Errorf("Arguments[1].Key = %q, want %q", *cmd.Arguments[1].Key, "key")
				}
				if cmd.Arguments[1].Value == nil || cmd.Arguments[1].Value.Term == nil {
					t.Fatal("Arguments[1].Value.Term is nil")
				}
				if *cmd.Arguments[1].Value.Term != "value" {
					t.Errorf("Arguments[1].Value.Term = %q, want %q", *cmd.Arguments[1].Value.Term, "value")
				}
				// arg2 - positional
				if cmd.Arguments[2].Key != nil {
					t.Errorf("Arguments[2].Key = %v, want nil", cmd.Arguments[2].Key)
				}
				if cmd.Arguments[2].Value == nil || cmd.Arguments[2].Value.Term == nil {
					t.Fatal("Arguments[2].Value.Term is nil")
				}
				if *cmd.Arguments[2].Value.Term != "arg2" {
					t.Errorf("Arguments[2].Value.Term = %q, want %q", *cmd.Arguments[2].Value.Term, "arg2")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if len(result.Blocks) != tt.wantBlocks {
				t.Fatalf("Blocks length = %d, want %d", len(result.Blocks), tt.wantBlocks)
			}
			if result.Blocks[0].SlashCommand == nil {
				t.Fatal("expected SlashCommand in first block, got nil")
			}
			tt.validateCmd(t, result.Blocks[0].SlashCommand)
		})
	}
}

func TestParseTask_Text(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantBlocks int
		validate   func(t *testing.T, result *Input)
	}{
		{
			name:       "simple text",
			input:      "This is a simple task\n",
			wantErr:    false,
			wantBlocks: 1,
			validate: func(t *testing.T, result *Input) {
				if result.Blocks[0].Text == nil {
					t.Fatal("expected Text block, got nil")
				}
				text := strings.Join(result.Blocks[0].Text.Content, "")
				if !strings.Contains(text, "This") || !strings.Contains(text, "simple") || !strings.Contains(text, "task") {
					t.Errorf("Text content = %q, want to contain 'This', 'simple', 'task'", text)
				}
			},
		},
		{
			name:       "multiline text",
			input:      "Line 1\nLine 2\nLine 3\n",
			wantErr:    false,
			wantBlocks: 1,
			validate: func(t *testing.T, result *Input) {
				if result.Blocks[0].Text == nil {
					t.Fatal("expected Text block, got nil")
				}
				text := strings.Join(result.Blocks[0].Text.Content, "")
				if !strings.Contains(text, "Line") {
					t.Errorf("Text content = %q, want to contain 'Line'", text)
				}
			},
		},
		{
			name:       "text with slash not at line start",
			input:      "Check the file path\n",
			wantErr:    false,
			wantBlocks: 1,
			validate: func(t *testing.T, result *Input) {
				if result.Blocks[0].Text == nil {
					t.Fatal("expected Text block, got nil")
				}
				text := strings.Join(result.Blocks[0].Text.Content, "")
				if !strings.Contains(text, "file") || !strings.Contains(text, "path") {
					t.Errorf("Text content = %q, want to contain 'file' and 'path'", text)
				}
			},
		},
		{
			name:       "text with equals sign",
			input:      "x = y + z\n",
			wantErr:    false,
			wantBlocks: 1,
			validate: func(t *testing.T, result *Input) {
				if result.Blocks[0].Text == nil {
					t.Fatal("expected Text block, got nil")
				}
				text := strings.Join(result.Blocks[0].Text.Content, "")
				if !strings.Contains(text, "=") {
					t.Errorf("Text content = %q, want to contain '='", text)
				}
			},
		},
		{
			name:       "text with quoted string",
			input:      "The message is \"Hello World\"\n",
			wantErr:    false,
			wantBlocks: 1,
			validate: func(t *testing.T, result *Input) {
				if result.Blocks[0].Text == nil {
					t.Fatal("expected Text block, got nil")
				}
				// Should have captured the string token
				found := false
				for _, content := range result.Blocks[0].Text.Content {
					if strings.Contains(content, "Hello World") {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Text content = %v, want to contain 'Hello World'", result.Blocks[0].Text.Content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if len(result.Blocks) != tt.wantBlocks {
				t.Fatalf("Blocks length = %d, want %d", len(result.Blocks), tt.wantBlocks)
			}
			tt.validate(t, result)
		})
	}
}

func TestParseTask_MixedBlocks(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantBlocks int
		validate   func(t *testing.T, result *Input)
	}{
		{
			name: "text followed by command",
			input: `Please fix the bug
/fix-bug 123
`,
			wantErr:    false,
			wantBlocks: 2,
			validate: func(t *testing.T, result *Input) {
				// First block should be text
				if result.Blocks[0].Text == nil {
					t.Fatal("expected Text in first block, got nil")
				}
				text := strings.Join(result.Blocks[0].Text.Content, "")
				if !strings.Contains(text, "Please") {
					t.Errorf("First block text = %q, want to contain 'Please'", text)
				}
				// Second block should be command
				if result.Blocks[1].SlashCommand == nil {
					t.Fatal("expected SlashCommand in second block, got nil")
				}
				if result.Blocks[1].SlashCommand.Name != "fix-bug" {
					t.Errorf("SlashCommand.Name = %q, want %q", result.Blocks[1].SlashCommand.Name, "fix-bug")
				}
			},
		},
		{
			name: "command followed by text",
			input: `/fix-bug 123
This is additional context
`,
			wantErr:    false,
			wantBlocks: 2,
			validate: func(t *testing.T, result *Input) {
				// First block should be command
				if result.Blocks[0].SlashCommand == nil {
					t.Fatal("expected SlashCommand in first block, got nil")
				}
				if result.Blocks[0].SlashCommand.Name != "fix-bug" {
					t.Errorf("SlashCommand.Name = %q, want %q", result.Blocks[0].SlashCommand.Name, "fix-bug")
				}
				// Second block should be text
				if result.Blocks[1].Text == nil {
					t.Fatal("expected Text in second block, got nil")
				}
				text := strings.Join(result.Blocks[1].Text.Content, "")
				if !strings.Contains(text, "additional") {
					t.Errorf("Second block text = %q, want to contain 'additional'", text)
				}
			},
		},
		{
			name: "multiple commands with text between",
			input: `/init project
Set up the environment
/config name=myapp
Additional settings here
`,
			wantErr:    false,
			wantBlocks: 4,
			validate: func(t *testing.T, result *Input) {
				// Block 0: command
				if result.Blocks[0].SlashCommand == nil {
					t.Fatal("expected SlashCommand in block 0, got nil")
				}
				if result.Blocks[0].SlashCommand.Name != "init" {
					t.Errorf("Block 0 SlashCommand.Name = %q, want %q", result.Blocks[0].SlashCommand.Name, "init")
				}
				// Block 1: text
				if result.Blocks[1].Text == nil {
					t.Fatal("expected Text in block 1, got nil")
				}
				// Block 2: command
				if result.Blocks[2].SlashCommand == nil {
					t.Fatal("expected SlashCommand in block 2, got nil")
				}
				if result.Blocks[2].SlashCommand.Name != "config" {
					t.Errorf("Block 2 SlashCommand.Name = %q, want %q", result.Blocks[2].SlashCommand.Name, "config")
				}
				// Block 3: text
				if result.Blocks[3].Text == nil {
					t.Fatal("expected Text in block 3, got nil")
				}
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
			result, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if len(result.Blocks) != tt.wantBlocks {
				t.Fatalf("Blocks length = %d, want %d", len(result.Blocks), tt.wantBlocks)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestParseTask_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "just newlines",
			input:   "\n\n\n",
			wantErr: false,
		},
		{
			name:    "just whitespace",
			input:   "   \t  \n",
			wantErr: false,
		},
		{
			name:    "slash at start with no command name",
			input:   "/\n",
			wantErr: true, // Should error as there's no command name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
