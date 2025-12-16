package codingcontext

import (
	"testing"
)

// TestParseTask_LenientParsing tests the lenient parameter parsing features
// added to support flexible quote types and escape sequences
func TestParseTask_LenientParsing(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, task Task)
	}{
		{
			name:    "single-quoted argument value",
			input:   "/command 'single quoted value'\n",
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
				// The parser includes quotes in the token
				expectedValue := `'single quoted value'`
				if cmd.Arguments[0].Value != expectedValue {
					t.Errorf("expected argument %q, got %q", expectedValue, cmd.Arguments[0].Value)
				}
				// After stripQuotes, it should be unquoted
				params := cmd.Params()
				if params["1"] != "single quoted value" {
					t.Errorf("expected params[1] = %q, got %q", "single quoted value", params["1"])
				}
			},
		},
		{
			name:    "double-quoted argument value",
			input:   `/command "double quoted value"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				if len(task) != 1 {
					t.Fatalf("expected 1 block, got %d", len(task))
				}
				cmd := task[0].SlashCommand
				if cmd == nil {
					t.Fatal("expected slash command block")
				}
				params := cmd.Params()
				if params["1"] != "double quoted value" {
					t.Errorf("expected params[1] = %q, got %q", "double quoted value", params["1"])
				}
			},
		},
		{
			name:    "single-quoted named parameter",
			input:   "/command key='value'\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				if params["key"] != "value" {
					t.Errorf("expected params[key] = %q, got %q", "value", params["key"])
				}
			},
		},
		{
			name:    "double-quoted named parameter",
			input:   `/command key="value"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				if params["key"] != "value" {
					t.Errorf("expected params[key] = %q, got %q", "value", params["key"])
				}
			},
		},
		{
			name:    "escape sequence: newline",
			input:   `/command "line1\nline2"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := "line1\nline2"
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: tab",
			input:   `/command "col1\tcol2"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := "col1\tcol2"
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: carriage return",
			input:   `/command "line1\rline2"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := "line1\rline2"
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: backslash",
			input:   `/command "path\\to\\file"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := `path\to\file`
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: escaped double quote",
			input:   `/command "say \"hello\""` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := `say "hello"`
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: escaped single quote",
			input:   "/command 'say \\'hello\\''\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := `say 'hello'`
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: Unicode \\uXXXX",
			input:   `/command "hello\u0020world"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := "hello world" // \u0020 is space
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: hex \\xHH",
			input:   `/command "A\x42C"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := "ABC" // \x42 is 'B'
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "escape sequence: octal \\OOO",
			input:   `/command "\101\102\103"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := "ABC" // \101=A, \102=B, \103=C in octal
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "mixed escape sequences",
			input:   `/command "line1\nline2\ttabbed\rreturned\\backslash"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				expected := "line1\nline2\ttabbed\rreturned\\backslash"
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "invalid escape sequence treated as literal",
			input:   `/command "\z\q"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				// Unknown escapes keep the character after backslash
				expected := "zq"
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "multiple arguments with different quote types",
			input:   "/command \"double\" 'single' unquoted\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				if len(cmd.Arguments) != 3 {
					t.Fatalf("expected 3 arguments, got %d", len(cmd.Arguments))
				}
				params := cmd.Params()
				if params["1"] != "double" {
					t.Errorf("expected params[1] = %q, got %q", "double", params["1"])
				}
				if params["2"] != "single" {
					t.Errorf("expected params[2] = %q, got %q", "single", params["2"])
				}
				if params["3"] != "unquoted" {
					t.Errorf("expected params[3] = %q, got %q", "unquoted", params["3"])
				}
			},
		},
		{
			name:    "named parameters with different quote types",
			input:   "/command k1=\"v1\" k2='v2' k3=v3\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				if params["k1"] != "v1" {
					t.Errorf("expected params[k1] = %q, got %q", "v1", params["k1"])
				}
				if params["k2"] != "v2" {
					t.Errorf("expected params[k2] = %q, got %q", "v2", params["k2"])
				}
				if params["k3"] != "v3" {
					t.Errorf("expected params[k3] = %q, got %q", "v3", params["k3"])
				}
			},
		},
		{
			name:    "UTF-8 characters in values",
			input:   "/command \"こんにちは\" emoji=\"🚀\"\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				if params["1"] != "こんにちは" {
					t.Errorf("expected params[1] = %q, got %q", "こんにちは", params["1"])
				}
				if params["emoji"] != "🚀" {
					t.Errorf("expected params[emoji] = %q, got %q", "🚀", params["emoji"])
				}
			},
		},
		{
			name:    "incomplete Unicode escape handled gracefully",
			input:   `/command "\u00a"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				// Incomplete escape should be kept as-is
				expected := "\\u00a"
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
				}
			},
		},
		{
			name:    "incomplete hex escape handled gracefully",
			input:   `/command "\x4"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				// Incomplete escape should be kept as-is
				expected := "\\x4"
				if params["1"] != expected {
					t.Errorf("expected params[1] = %q, got %q", expected, params["1"])
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

// TestStripQuotes tests the stripQuotes function with various input types
func TestStripQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double-quoted string",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "single-quoted string",
			input:    `'hello'`,
			expected: "hello",
		},
		{
			name:     "unquoted string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "double-quoted with escaped quote",
			input:    `"say \"hello\""`,
			expected: `say "hello"`,
		},
		{
			name:     "single-quoted with escaped quote",
			input:    `'say \'hello\''`,
			expected: `say 'hello'`,
		},
		{
			name:     "double-quoted with newline escape",
			input:    `"line1\nline2"`,
			expected: "line1\nline2",
		},
		{
			name:     "double-quoted with tab escape",
			input:    `"col1\tcol2"`,
			expected: "col1\tcol2",
		},
		{
			name:     "double-quoted with Unicode escape",
			input:    `"hello\u0020world"`,
			expected: "hello world",
		},
		{
			name:     "double-quoted with hex escape",
			input:    `"A\x42C"`,
			expected: "ABC",
		},
		{
			name:     "double-quoted with octal escape",
			input:    `"\101\102\103"`,
			expected: "ABC",
		},
		{
			name:     "empty double-quoted string",
			input:    `""`,
			expected: "",
		},
		{
			name:     "empty single-quoted string",
			input:    `''`,
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "a",
		},
		{
			name:     "no escape processing for unquoted",
			input:    `hello\nworld`,
			expected: `hello\nworld`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("stripQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestProcessEscapeSequences tests the processEscapeSequences function
func TestProcessEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no escapes",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "newline",
			input:    `line1\nline2`,
			expected: "line1\nline2",
		},
		{
			name:     "tab",
			input:    `col1\tcol2`,
			expected: "col1\tcol2",
		},
		{
			name:     "carriage return",
			input:    `line1\rline2`,
			expected: "line1\rline2",
		},
		{
			name:     "backslash",
			input:    `path\\to\\file`,
			expected: `path\to\file`,
		},
		{
			name:     "double quote",
			input:    `say \"hello\"`,
			expected: `say "hello"`,
		},
		{
			name:     "single quote",
			input:    `say \'hello\'`,
			expected: `say 'hello'`,
		},
		{
			name:     "Unicode escape",
			input:    `\u0048\u0065\u006c\u006c\u006f`, // "Hello"
			expected: "Hello",
		},
		{
			name:     "hex escape",
			input:    `\x48\x65\x6c\x6c\x6f`, // "Hello"
			expected: "Hello",
		},
		{
			name:     "octal escape",
			input:    `\110\145\154\154\157`, // "Hello"
			expected: "Hello",
		},
		{
			name:     "mixed escapes",
			input:    `\n\t\r\\\"\'\u0020\x20\40`,
			expected: "\n\t\r\\\"' \x20 ",
		},
		{
			name:     "unknown escape",
			input:    `\z\q`,
			expected: "zq",
		},
		{
			name:     "incomplete Unicode escape",
			input:    `\u00a`,
			expected: `\u00a`,
		},
		{
			name:     "incomplete hex escape",
			input:    `\x4`,
			expected: `\x4`,
		},
		{
			name:     "backslash at end",
			input:    `hello\`,
			expected: `hello\`,
		},
		{
			name:     "octal with non-octal digits",
			input:    `\7\8\9`,
			expected: "\x0789", // \7 is octal 7, \8 and \9 are treated as unknown escapes and output as '8' and '9'
		},
		{
			name:     "short octal sequences",
			input:    `\7\77`,
			expected: "\x07?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processEscapeSequences(tt.input)
			if result != tt.expected {
				t.Errorf("processEscapeSequences(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
