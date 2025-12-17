package codingcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.Len(t, task, 1)
				cmd := task[0].SlashCommand
				require.NotNil(t, cmd)
				require.Len(t, cmd.Arguments, 1)
				// The parser includes quotes in the token
				assert.Equal(t, `'single quoted value'`, cmd.Arguments[0].Value)
				// After stripQuotes, it should be unquoted
				params := cmd.Params()
				assert.Equal(t, "single quoted value", params["1"])
			},
		},
		{
			name:    "double-quoted argument value",
			input:   `/command "double quoted value"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				require.Len(t, task, 1)
				cmd := task[0].SlashCommand
				require.NotNil(t, cmd)
				params := cmd.Params()
				assert.Equal(t, "double quoted value", params["1"])
			},
		},
		{
			name:    "single-quoted named parameter",
			input:   "/command key='value'\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "value", params["key"])
			},
		},
		{
			name:    "double-quoted named parameter",
			input:   `/command key="value"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "value", params["key"])
			},
		},
		{
			name:    "escape sequence: newline",
			input:   `/command "line1\nline2"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "line1\nline2", params["1"])
			},
		},
		{
			name:    "escape sequence: tab",
			input:   `/command "col1\tcol2"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "col1\tcol2", params["1"])
			},
		},
		{
			name:    "escape sequence: carriage return",
			input:   `/command "line1\rline2"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "line1\rline2", params["1"])
			},
		},
		{
			name:    "escape sequence: backslash",
			input:   `/command "path\\to\\file"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, `path\to\file`, params["1"])
			},
		},
		{
			name:    "escape sequence: escaped double quote",
			input:   `/command "say \"hello\""` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, `say "hello"`, params["1"])
			},
		},
		{
			name:    "escape sequence: escaped single quote",
			input:   "/command 'say \\'hello\\''\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, `say 'hello'`, params["1"])
			},
		},
		{
			name:    "escape sequence: Unicode \\uXXXX",
			input:   `/command "hello\u0020world"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "hello world", params["1"]) // \u0020 is space
			},
		},
		{
			name:    "escape sequence: hex \\xHH",
			input:   `/command "A\x42C"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "ABC", params["1"]) // \x42 is 'B'
			},
		},
		{
			name:    "escape sequence: octal \\OOO",
			input:   `/command "\101\102\103"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "ABC", params["1"]) // \101=A, \102=B, \103=C in octal
			},
		},
		{
			name:    "mixed escape sequences",
			input:   `/command "line1\nline2\ttabbed\rreturned\\backslash"` + "\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "line1\nline2\ttabbed\rreturned\\backslash", params["1"])
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
				assert.Equal(t, "zq", params["1"])
			},
		},
		{
			name:    "multiple arguments with different quote types",
			input:   "/command \"double\" 'single' unquoted\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				require.Len(t, cmd.Arguments, 3)
				params := cmd.Params()
				assert.Equal(t, "double", params["1"])
				assert.Equal(t, "single", params["2"])
				assert.Equal(t, "unquoted", params["3"])
			},
		},
		{
			name:    "named parameters with different quote types",
			input:   "/command k1=\"v1\" k2='v2' k3=v3\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "v1", params["k1"])
				assert.Equal(t, "v2", params["k2"])
				assert.Equal(t, "v3", params["k3"])
			},
		},
		{
			name:    "UTF-8 characters in values",
			input:   "/command \"こんにちは\" emoji=\"🚀\"\n",
			wantErr: false,
			check: func(t *testing.T, task Task) {
				cmd := task[0].SlashCommand
				params := cmd.Params()
				assert.Equal(t, "こんにちは", params["1"])
				assert.Equal(t, "🚀", params["emoji"])
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
				assert.Equal(t, "\\u00a", params["1"])
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
				assert.Equal(t, "\\x4", params["1"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := ParseTask(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
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
			assert.Equal(t, tt.expected, result)
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
			expected: "\a89", // \7 is octal 7 (ASCII bell \a), \8 and \9 are treated as unknown escapes and output as '8' and '9'
		},
		{
			name:     "short octal sequences",
			input:    `\7\77`,
			expected: "\a?", // \7 is octal 7 (ASCII bell \a), \77 is octal 77 (ASCII '?')
		},
		{
			name:     "octal values above 127",
			input:    `\200\377`,
			expected: "\x80\xff", // \200 is 128, \377 is 255
		},
		{
			name:     "octal high byte values",
			input:    `\240\300\350`,
			expected: "\xa0\xc0\xe8", // \240 is 160, \300 is 192, \350 is 232
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processEscapeSequences(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
