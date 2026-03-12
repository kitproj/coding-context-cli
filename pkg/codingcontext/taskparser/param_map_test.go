package taskparser_test

import (
	"strings"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParams_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			name:    "valid key=value",
			value:   "key=value",
			wantKey: "key",
			wantVal: "value",
			wantErr: false,
		},
		{
			name:    "key=value with equals in value (requires quotes)",
			value:   `key="value=with=equals"`,
			wantKey: "key",
			wantVal: "value=with=equals",
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "key=",
			wantKey: "key",
			wantVal: "",
			wantErr: false,
		},
		{
			name:    "positional argument - no equals",
			value:   "keyvalue",
			wantVal: "keyvalue",
			wantErr: false,
		},
		{
			name:    "positional argument - only key",
			value:   "key",
			wantVal: "key",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := taskparser.Params{}
			err := p.Set(tt.value)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.wantKey != "" {
					// Named parameter
					assert.Equal(t, tt.wantVal, p.Value(tt.wantKey))
				} else {
					// Positional argument
					args := p.Arguments()
					require.NotEmpty(t, args, "expected positional arguments")
					assert.Equal(t, tt.wantVal, args[0])
				}
			}
		})
	}
}

func TestParams_String(t *testing.T) {
	t.Parallel()

	p := taskparser.Params{
		"key1": []string{"value1"},
		"key2": []string{"value2"},
	}

	s := p.String()
	if s == "" {
		t.Error("Params.String() returned empty string")
	}
}

func TestParams_SetMultiple(t *testing.T) {
	t.Parallel()

	p, err := taskparser.ParseParams("key1=value1, key2=value2")
	require.NoError(t, err)
	assert.Len(t, p, 2)
	assert.Equal(t, "value1", p.Value("key1"))
	assert.Equal(t, "value2", p.Value("key2"))
}

func parseParamsCases() []struct {
	name      string
	input     string
	expected  taskparser.Params
	wantError bool
	errorMsg  string
} {
	return []struct {
		name      string
		input     string
		expected  taskparser.Params
		wantError bool
		errorMsg  string
	}{
		{name: "empty string", input: "", expected: taskparser.Params{}},
		{name: "single quoted key=value", input: `key="value"`,
			expected: taskparser.Params{"key": []string{"value"}}},
		{name: "multiple quoted key=value pairs", input: `key1="value1" key2="value2" key3="value3"`,
			expected: taskparser.Params{"key1": {"value1"}, "key2": {"value2"}, "key3": {"value3"}}},
		{name: "double-quoted value with spaces", input: `key1="value with spaces" key2="value2"`,
			expected: taskparser.Params{"key1": {"value with spaces"}, "key2": {"value2"}}},
		{name: "escaped double quotes", input: `key1="value with \"escaped\" quotes"`,
			expected: taskparser.Params{"key1": {`value with "escaped" quotes`}}},
		{name: "value with equals sign in quotes", input: `key1="value=with=equals" key2="normal"`,
			expected: taskparser.Params{"key1": {"value=with=equals"}, "key2": {"normal"}}},
		{name: "empty quoted value", input: `key1="" key2="value2"`,
			expected: taskparser.Params{"key1": {""}, "key2": {"value2"}}},
		{name: "whitespace around equals", input: `key1 = "value1" key2="value2"`,
			expected: taskparser.Params{"key1": {"value1"}, "key2": {"value2"}}},
		{name: "quoted value with spaces and equals", input: `key1="value with spaces and = signs"`,
			expected: taskparser.Params{"key1": {"value with spaces and = signs"}}},
		{name: "unquoted value - error", input: `key1=value1`,
			expected: taskparser.Params{"key1": {"value1"}}},
		{name: "mixed quoted and unquoted", input: `key1="quoted value" key2=unquoted`,
			expected: taskparser.Params{"key1": {"quoted value"}, "key2": {"unquoted"}}},
		{name: "unclosed quote - error", input: `key1="value with spaces`,
			wantError: true, errorMsg: "unclosed quote"},
		{name: "missing value after equals with comma separator", input: `key1=, key2="value2"`,
			expected: taskparser.Params{"key1": {}, "key2": {"value2"}}},
		{name: "single quotes", input: `key1='value'`,
			expected: taskparser.Params{"key1": {"value"}}},
	}
}

func TestParseParams(t *testing.T) {
	t.Parallel()

	for _, tt := range parseParamsCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := taskparser.ParseParams(tt.input)

			if tt.wantError {
				require.Error(t, err)

				if tt.errorMsg != "" {
					if !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("ParseParams() error = %v, want error containing %q", err, tt.errorMsg)
					}
				}

				return
			}

			require.NoError(t, err)
			assert.Len(t, result, len(tt.expected))

			for k, v := range tt.expected {
				assert.Equal(t, v, result.Values(k))
			}
		})
	}
}
