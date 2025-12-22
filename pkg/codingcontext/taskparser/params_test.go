package taskparser_test

import (
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTaskParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected taskparser.Params
	}{
		{
			name:     "empty string",
			input:    "",
			expected: taskparser.Params{},
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: taskparser.Params{},
		},
		{
			name:  "single pair",
			input: "key=value",
			expected: taskparser.Params{
				"key": {"value"},
			},
		},
		{
			name:  "comma separated pairs",
			input: "key=value,foo=bar",
			expected: taskparser.Params{
				"key": {"value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "space separated pairs",
			input: "key=value foo=bar",
			expected: taskparser.Params{
				"key": {"value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "wrapped single quotes",
			input: "key=\"'value'\"",
			expected: taskparser.Params{
				"key": {"'value'"},
			},
		},
		{
			name:  "wrapped single quotes",
			input: "key='\"value\"'",
			expected: taskparser.Params{
				"key": {`"value"`},
			},
		},
		{
			name:  "mixed separators",
			input: "key1=value1, key2=value2 key3=value3",
			expected: taskparser.Params{
				"key1": {"value1"},
				"key2": {"value2"},
				"key3": {"value3"},
			},
		},
		{
			name:  "trailing comma",
			input: "key=value,",
			expected: taskparser.Params{
				"key": {"value"},
			},
		},
		{
			name:  "multiple spaces",
			input: "key=value   foo=bar",
			expected: taskparser.Params{
				"key": {"value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "whitespace around equals",
			input: "key = value, foo = bar",
			expected: taskparser.Params{
				"key": {"value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "non-breaking spaces trimmed from unquoted values",
			input: "key=\u00a0value\u00a0, foo=\u00a0bar\u00a0",
			expected: taskparser.Params{
				"key": {"value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "non-breaking spaces trimmed from quoted values",
			input: "key=\"\u00a0value\u00a0\", foo='\u00a0bar\u00a0'",
			expected: taskparser.Params{
				"key": {"value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "unicode escape sequence in quoted value",
			input: `key="foo\u00a0", foo='bar\u00a0'`,
			expected: taskparser.Params{
				"key": {"foo"},
				"foo": {"bar"},
			},
		},
		{
			name:  "unicode escape sequence in unquoted value",
			input: `key=foo\u00a0, foo=bar\u00a0`,
			expected: taskparser.Params{
				"key": {"foo"},
				"foo": {"bar"},
			},
		},
		{
			name:  "unicode escape sequence with regular characters",
			input: `key="test\u00a0value", foo=hello\u0020world`,
			expected: taskparser.Params{
				"key": {"test\u00a0value"},
				"foo": {"hello world"},
			},
		},
		{
			name:  "quoted values with double quotes",
			input: `key="string value", foo="bar"`,
			expected: taskparser.Params{
				"key": {"string value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "quoted values with single quotes",
			input: `key='string value', foo='bar'`,
			expected: taskparser.Params{
				"key": {"string value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "quotes with spaces around",
			input: `key = "value" , foo = "bar"`,
			expected: taskparser.Params{
				"key": {"value"},
				"foo": {"bar"},
			},
		},
		{
			name:  "escaped quotes in double quoted string",
			input: `key="bar\"baz\""`,
			expected: taskparser.Params{
				"key": {`bar"baz"`},
			},
		},
		{
			name:  "escaped quotes in single quoted string",
			input: `key='bar\'baz\''`,
			expected: taskparser.Params{
				"key": {`bar'baz'`},
			},
		},
		{
			name:  "multiple escape sequences",
			input: `text="line1\nline2\ttabbed\rreturned\\backslash"`,
			expected: taskparser.Params{
				"text": {"line1\nline2\ttabbed\rreturned\\backslash"},
			},
		},
		{
			name:  "unquoted values with spaces become positional arguments",
			input: "key=value with spaces",
			expected: taskparser.Params{
				"key":                   {"value"},
				taskparser.ArgumentsKey: {"with", "spaces"},
			},
		},
		{
			name:  "quoted values with spaces",
			input: `key="value with spaces"`,
			expected: taskparser.Params{
				"key": {"value with spaces"},
			},
		},
		{
			name:  "value with comma in quotes",
			input: `key="value,with,commas"`,
			expected: taskparser.Params{
				"key": {"value,with,commas"},
			},
		},
		{
			name:  "complex example from user",
			input: `key="string value", foo="bar\"baz\"", multiline="line1\nline2"`,
			expected: taskparser.Params{
				"key":       {"string value"},
				"foo":       {`bar"baz"`},
				"multiline": {"line1\nline2"},
			},
		},
		{
			name:  "hex escape sequence",
			input: `key="\x41\x42"`,
			expected: taskparser.Params{
				"key": {"AB"},
			},
		},
		{
			name:  "octal escape sequence",
			input: `key="\101\102"`,
			expected: taskparser.Params{
				"key": {"AB"},
			},
		},
		{
			name:  "octal escape with different lengths",
			input: `key="\7\77\177"`,
			expected: taskparser.Params{
				"key": {"\x07?\x7f"},
			},
		},
		{
			name:  "non-octal digit escape",
			input: `key="\8\9\89"`,
			expected: taskparser.Params{
				"key": {"8989"},
			},
		},
		{
			name:  "non-octal escape characters",
			input: `key="\a\z\A\!"`,
			expected: taskparser.Params{
				"key": {"azA!"},
			},
		},
		{
			name:  "mixed octal and non-octal escapes",
			input: `key="\101\8\102\9"`,
			expected: taskparser.Params{
				"key": {"A8B9"},
			},
		},
		{
			name:  "octal escape boundary",
			input: `key="\08\09"`,
			expected: taskparser.Params{
				"key": {"\x008\x009"},
			},
		},
		{
			name:  "duplicate keys",
			input: "key=value1 key=value2, key=value3",
			expected: taskparser.Params{
				"key": {"value1", "value2", "value3"},
			},
		},
		{
			name:  "multiple duplicate keys",
			input: "key1=value1, key2=value2, key1=value3, key2=value4",
			expected: taskparser.Params{
				"key1": {"value1", "value3"},
				"key2": {"value2", "value4"},
			},
		},
		{
			name:  "case folding",
			input: "Key=value, FOO=bar, KeyName=value1, keyName=value2, KEYNAME=value3",
			expected: taskparser.Params{
				"key":     {"value"},
				"foo":     {"bar"},
				"keyname": {"value1", "value2", "value3"},
			},
		},
		{
			name:  "UTF-8 characters",
			input: "ключ=значение, key=こんにちは, 键=值, emoji=🚀",
			expected: taskparser.Params{
				"ключ":  {"значение"},
				"key":   {"こんにちは"},
				"键":     {"值"},
				"emoji": {"🚀"},
			},
		},
		{
			name:  "UTF-8 with Unicode whitespace",
			input: "ключ1=значение1\u2003ключ2=значение2",
			expected: taskparser.Params{
				"ключ1": {"значение1"},
				"ключ2": {"значение2"},
			},
		},
		{
			name:  "quoted UTF-8 value with spaces",
			input: `ключ="значение с пробелами"`,
			expected: taskparser.Params{
				"ключ": {"значение с пробелами"},
			},
		},
		{
			name:  "mixed UTF-8 and ASCII",
			input: "key=value ключ=значение foo=bar",
			expected: taskparser.Params{
				"key":  {"value"},
				"ключ": {"значение"},
				"foo":  {"bar"},
			},
		},
		{
			name:  "UTF-8 value containing equals sign",
			input: `key="значение=со=равно"`,
			expected: taskparser.Params{
				"key": {"значение=со=равно"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := taskparser.ParseParams(tt.input)
			require.NoError(t, err)

			// Check positional arguments using Arguments() accessor
			expectedArgs, hasArgs := tt.expected[taskparser.ArgumentsKey]
			actualArgs := result.Arguments()
			if hasArgs {
				assert.Equal(t, expectedArgs, actualArgs, "positional arguments mismatch")
			} else {
				assert.Empty(t, actualArgs, "expected no positional arguments")
			}

			// Check named parameters
			for key, expectedValues := range tt.expected {
				if key != taskparser.ArgumentsKey {
					assert.Equal(t, expectedValues, result.Values(key), "values mismatch for key %q", key)
				}
			}

			// Verify no unexpected keys
			for key := range result {
				if key != taskparser.ArgumentsKey && tt.expected[key] == nil {
					t.Errorf("unexpected key in result: %q", key)
				}
			}
		})
	}
}

func TestParams_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		key      string
		expected string
	}{
		{
			name:     "single value",
			input:    "key=value",
			key:      "key",
			expected: "value",
		},
		{
			name:     "multiple values returns first",
			input:    "key=value1 key=value2 key=value3",
			key:      "key",
			expected: "value1",
		},
		{
			name:     "non-existent key returns empty",
			input:    "key=value",
			key:      "nonexistent",
			expected: "",
		},
		{
			name:     "empty params returns empty",
			input:    "",
			key:      "key",
			expected: "",
		},
		{
			name:     "case insensitive lookup",
			input:    "Key=value, KeyName=value2",
			key:      "KEY",
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := taskparser.ParseParams(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, params.Value(tt.key))
		})
	}
}

func TestParams_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		key      string
		expected []string
	}{
		{
			name:     "single value",
			input:    "key=value",
			key:      "key",
			expected: []string{"value"},
		},
		{
			name:     "multiple values returns all",
			input:    "key=value1 key=value2 key=value3",
			key:      "key",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "non-existent key returns nil",
			input:    "key=value",
			key:      "nonexistent",
			expected: nil,
		},
		{
			name:     "empty params returns nil",
			input:    "",
			key:      "key",
			expected: nil,
		},
		{
			name:     "multiple keys",
			input:    "key1=value1 key2=value2 key1=value3",
			key:      "key1",
			expected: []string{"value1", "value3"},
		},
		{
			name:     "case insensitive lookup",
			input:    "Key=value1 Key=value2, KeyName=value3 keyName=value4",
			key:      "KEY",
			expected: []string{"value1", "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := taskparser.ParseParams(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, params.Values(tt.key))
		})
	}
}

func TestParams_NilSafety(t *testing.T) {
	t.Parallel()

	var params taskparser.Params

	assert.Equal(t, "", params.Value("key"))
	assert.Nil(t, params.Values("key"))
}

func TestParse_EmptyValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		key      string
		expected []string
	}{
		{
			name:     "empty unquoted value returns empty slice",
			input:    "key=",
			key:      "key",
			expected: []string{},
		},
		{
			name:     "empty unquoted value with trailing space returns empty slice",
			input:    "key= ",
			key:      "key",
			expected: []string{},
		},
		{
			name:     "explicitly quoted empty value returns slice with empty string",
			input:    `key=""`,
			key:      "key",
			expected: []string{""},
		},
		{
			name:     "empty value before comma returns empty slice",
			input:    "key=,foo=bar",
			key:      "key",
			expected: []string{},
		},
		{
			name:     "empty value with trailing comma at end returns empty slice",
			input:    "key=,",
			key:      "key",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := taskparser.ParseParams(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, params.Values(tt.key))
		})
	}
}

func TestParse_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty key",
			input: "=value",
		},
		{
			name:  "empty key, second value",
			input: "key=value =value",
		},
		{
			name:  "incorrect quotes",
			input: "key='value\"",
		},
		{
			name:  "incorrect quotes, double quote first",
			input: "key=\"value'",
		},
		{
			name:  "incorrect quotes, multiple",
			input: "key=\"value', key2=\"value2'",
		},
		{
			name:  "empty key, second value, comma",
			input: "key=value, =value",
		},
		{
			name:  "unclosed double quote",
			input: `key="unclosed`,
		},
		{
			name:  "unclosed single quote",
			input: `key='unclosed`,
		},
		{
			name:  "incomplete hex escape - missing one digit",
			input: `key="\x4"`,
		},
		{
			name:  "incomplete hex escape - missing both digits",
			input: `key="\x"`,
		},
		{
			name:  "invalid hex escape",
			input: `key="\xGH"`,
		},
		{
			name:  "incomplete hex escape - quote immediately after x with extra char",
			input: `key="\x"X"`,
		},
		{
			name:  "incomplete hex escape - invalid char followed by quote",
			input: `key="\xG"`,
		},
		{
			name:  "invalid hex escape - valid first digit, invalid second digit",
			input: `key="\x4G"`,
		},
		{
			name:  "incomplete unicode escape - missing one digit",
			input: `key="\u00a"`,
		},
		{
			name:  "incomplete unicode escape - missing all digits",
			input: `key="\u"`,
		},
		{
			name:  "invalid unicode escape",
			input: `key="\u00GH"`,
		},
		{
			name:  "incomplete unicode escape in unquoted value",
			input: `key=foo\u00a`,
		},
		{
			name:  "invalid unicode escape in unquoted value",
			input: `key=foo\u00GH`,
		},
		{
			name:  "wrapped quotes - malformed single quote wrapping",
			input: `key='\"'value\"'`,
		},
		{
			name:  "unclosed trailing quote",
			input: `key=value=with=equals"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := taskparser.ParseParams(tt.input)
			require.Error(t, err)
		})
	}
}

func TestParseParams_PositionalArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected taskparser.Params
	}{
		{
			name:  "single positional argument",
			input: "value",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"value"},
			},
		},
		{
			name:  "multiple positional arguments",
			input: "value1 value2 value3",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"value1", "value2", "value3"},
			},
		},
		{
			name:  "positional arguments with commas",
			input: "value1, value2, value3",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"value1", "value2", "value3"},
			},
		},
		{
			name:  "positional argument with spaces becomes multiple arguments",
			input: "value with spaces",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"value", "with", "spaces"},
			},
		},
		{
			name:  "quoted positional argument",
			input: `"quoted value"`,
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"quoted value"},
			},
		},
		{
			name:  "single quoted positional argument",
			input: `'quoted value'`,
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"quoted value"},
			},
		},
		{
			name:  "mixed positional and named arguments",
			input: "positional1 key=value positional2",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"positional1", "positional2"},
				"key":                   {"value"},
			},
		},
		{
			name:  "positional before named",
			input: "arg1 arg2 key=value",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "arg2"},
				"key":                   {"value"},
			},
		},
		{
			name:  "positional after named",
			input: "key=value arg1 arg2",
			expected: taskparser.Params{
				"key":                   {"value"},
				taskparser.ArgumentsKey: {"arg1", "arg2"},
			},
		},
		{
			name:  "positional between named",
			input: "key1=value1 arg1 key2=value2",
			expected: taskparser.Params{
				"key1":                  {"value1"},
				taskparser.ArgumentsKey: {"arg1"},
				"key2":                  {"value2"},
			},
		},
		{
			name:  "multiple positional with named",
			input: "arg1 key1=value1 arg2 arg3 key2=value2 arg4",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "arg2", "arg3", "arg4"},
				"key1":                  {"value1"},
				"key2":                  {"value2"},
			},
		},
		{
			name:  "positional with quoted value containing spaces",
			input: `arg1 "quoted arg with spaces" arg3`,
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "quoted arg with spaces", "arg3"},
			},
		},
		{
			name:  "positional with empty quoted value",
			input: `arg1 "" arg3`,
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "", "arg3"},
			},
		},
		{
			name:  "positional arguments with UTF-8",
			input: "значение1 значение2",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"значение1", "значение2"},
			},
		},
		{
			name:  "positional with escape sequences",
			input: `arg1 "line1\nline2" arg3`,
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "line1\nline2", "arg3"},
			},
		},
		{
			name:  "positional arguments separated by commas",
			input: "arg1,arg2,arg3",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "arg2", "arg3"},
			},
		},
		{
			name:  "positional with trailing comma",
			input: "arg1,arg2,",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "arg2"},
			},
		},
		{
			name:  "positional with leading comma",
			input: ",arg1,arg2",
			expected: taskparser.Params{
				taskparser.ArgumentsKey: {"arg1", "arg2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := taskparser.ParseParams(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
