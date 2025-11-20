package main

import (
	"context"
	"strings"
	"testing"
)

func TestProcessShellCommands(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:  "no shell commands",
			input: "This is plain text\nwith multiple lines",
			want:  "This is plain text\nwith multiple lines",
		},
		{
			name:  "single shell command - echo",
			input: "Before\n!`echo hello`\nAfter",
			want:  "Before\nhello\nAfter",
		},
		{
			name:  "multiple shell commands",
			input: "First:\n!`echo one`\nSecond:\n!`echo two`\nDone",
			want:  "First:\none\nSecond:\ntwo\nDone",
		},
		{
			name:  "shell command with pipe",
			input: "Result:\n!`echo 'hello world' | tr a-z A-Z`",
			want:  "Result:\nHELLO WORLD",
		},
		{
			name:  "shell command at start of content",
			input: "!`echo first line`\nSecond line",
			want:  "first line\nSecond line",
		},
		{
			name:  "shell command at end of content",
			input: "First line\n!`echo last line`",
			want:  "First line\nlast line",
		},
		{
			name:  "command with trailing spaces",
			input: "!`echo test`   \n",
			want:  "test",
		},
		{
			name:  "multiline output from command",
			input: "Lines:\n!`printf 'line1\\nline2\\nline3'`\nAfter",
			want:  "Lines:\nline1\nline2\nline3\nAfter",
		},
		{
			name:  "inline command not on own line ignored",
			input: "This !`echo inline` should not be processed",
			want:  "This !`echo inline` should not be processed",
		},
		{
			name:  "command without backticks ignored",
			input: "!echo test\n",
			want:  "!echo test\n",
		},
		{
			name:  "empty command",
			input: "!``\n",
			want:  "",
		},
		{
			name:        "command that fails",
			input:       "!`exit 1`\n",
			wantErr:     true,
			errContains: "shell command errors",
		},
		{
			name:  "command with special characters in output",
			input: "!`echo 'special: $var & < > \"quotes\"'`",
			want:  "special: $var & < > \"quotes\"",
		},
		{
			name:  "pwd command",
			input: "Directory:\n!`basename $(pwd)`",
			want:  "Directory:\ncoding-context-cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := processShellCommands(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("processShellCommands() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("processShellCommands() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("processShellCommands() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("processShellCommands() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcessShellCommandsContext(t *testing.T) {
	// Test that context cancellation stops command execution
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	input := "!`sleep 10`\n"
	_, err := processShellCommands(ctx, input)

	if err == nil {
		t.Error("processShellCommands() expected error with cancelled context but got none")
	}
}
