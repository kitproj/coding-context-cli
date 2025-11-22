package codingcontext

import (
	"strings"
	"testing"
)

func TestParams_Set(t *testing.T) {
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
			name:    "key=value with equals in value",
			value:   "key=value=with=equals",
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
			name:    "invalid format - no equals",
			value:   "keyvalue",
			wantErr: true,
		},
		{
			name:    "invalid format - only key",
			value:   "key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Params{}
			err := p.Set(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Params.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if p[tt.wantKey] != tt.wantVal {
					t.Errorf("Params[%q] = %q, want %q", tt.wantKey, p[tt.wantKey], tt.wantVal)
				}
			}
		})
	}
}

func TestParams_String(t *testing.T) {
	p := Params{
		"key1": "value1",
		"key2": "value2",
	}
	s := p.String()
	if s == "" {
		t.Error("Params.String() returned empty string")
	}
}

func TestParams_SetMultiple(t *testing.T) {
	p := Params{}

	if err := p.Set("key1=value1"); err != nil {
		t.Fatalf("Params.Set() failed: %v", err)
	}
	if err := p.Set("key2=value2"); err != nil {
		t.Fatalf("Params.Set() failed: %v", err)
	}

	if len(p) != 2 {
		t.Errorf("Params length = %d, want 2", len(p))
	}
	if p["key1"] != "value1" {
		t.Errorf("Params[key1] = %q, want %q", p["key1"], "value1")
	}
	if p["key2"] != "value2" {
		t.Errorf("Params[key2] = %q, want %q", p["key2"], "value2")
	}
}

func TestParseParams(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Params
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty string",
			input:     "",
			expected:  Params{},
			wantError: false,
		},
		{
			name:      "single quoted key=value",
			input:     `key="value"`,
			expected:  Params{"key": "value"},
			wantError: false,
		},
		{
			name:      "multiple quoted key=value pairs",
			input:     `key1="value1" key2="value2" key3="value3"`,
			expected:  Params{"key1": "value1", "key2": "value2", "key3": "value3"},
			wantError: false,
		},
		{
			name:      "double-quoted value with spaces",
			input:     `key1="value with spaces" key2="value2"`,
			expected:  Params{"key1": "value with spaces", "key2": "value2"},
			wantError: false,
		},
		{
			name:      "escaped double quotes",
			input:     `key1="value with \"escaped\" quotes"`,
			expected:  Params{"key1": `value with "escaped" quotes`},
			wantError: false,
		},
		{
			name:      "value with equals sign in quotes",
			input:     `key1="value=with=equals" key2="normal"`,
			expected:  Params{"key1": "value=with=equals", "key2": "normal"},
			wantError: false,
		},
		{
			name:      "empty quoted value",
			input:     `key1="" key2="value2"`,
			expected:  Params{"key1": "", "key2": "value2"},
			wantError: false,
		},
		{
			name:      "whitespace around equals",
			input:     `key1 = "value1" key2="value2"`,
			expected:  Params{"key1": "value1", "key2": "value2"},
			wantError: false,
		},
		{
			name:      "quoted value with spaces and equals",
			input:     `key1="value with spaces and = signs"`,
			expected:  Params{"key1": "value with spaces and = signs"},
			wantError: false,
		},
		{
			name:      "unquoted value - error",
			input:     `key1=value1`,
			wantError: true,
			errorMsg:  "unquoted value",
		},
		{
			name:      "mixed quoted and unquoted - error",
			input:     `key1="quoted value" key2=unquoted`,
			wantError: true,
			errorMsg:  "unquoted value",
		},
		{
			name:      "unclosed quote - error",
			input:     `key1="value with spaces`,
			wantError: true,
			errorMsg:  "unclosed quote",
		},
		{
			name:      "missing value after equals - error",
			input:     `key1= key2="value2"`,
			wantError: true,
			errorMsg:  "unquoted value",
		},
		{
			name:      "single quote not supported - error",
			input:     `key1='value'`,
			wantError: true,
			errorMsg:  "unquoted value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseParams(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("ParseParams() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				if err != nil && tt.errorMsg != "" {
					if !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("ParseParams() error = %v, want error containing %q", err, tt.errorMsg)
					}
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ParseParams() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("ParseParams()[%q] = %q, want %q", k, result[k], v)
				}
			}
		})
	}
}
