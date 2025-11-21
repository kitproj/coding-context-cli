package codingcontext

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseSlashCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		wantFound   bool
		wantTask    string
		wantParams  map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:       "simple command without parameters",
			command:    "/fix-bug",
			wantFound:  true,
			wantTask:   "fix-bug",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:      "command with single unquoted argument",
			command:   "/fix-bug 123",
			wantFound: true,
			wantTask:  "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": "123",
				"1":         "123",
			},
			wantErr: false,
		},
		{
			name:      "command with multiple unquoted arguments",
			command:   "/implement-feature login high urgent",
			wantFound: true,
			wantTask:  "implement-feature",
			wantParams: map[string]string{
				"ARGUMENTS": "login high urgent",
				"1":         "login",
				"2":         "high",
				"3":         "urgent",
			},
			wantErr: false,
		},
		{
			name:      "command with double-quoted argument containing spaces",
			command:   `/code-review "Fix authentication bug in login flow"`,
			wantFound: true,
			wantTask:  "code-review",
			wantParams: map[string]string{
				"ARGUMENTS": `"Fix authentication bug in login flow"`,
				"1":         "Fix authentication bug in login flow",
			},
			wantErr: false,
		},
		{
			name:       "missing leading slash",
			command:    "fix-bug",
			wantFound:  false,
			wantTask:   "",
			wantParams: nil,
			wantErr:    false,
		},
		{
			name:       "empty command",
			command:    "/",
			wantFound:  false,
			wantTask:   "",
			wantParams: nil,
			wantErr:    false,
		},
		{
			name:       "empty string",
			command:    "",
			wantFound:  false,
			wantTask:   "",
			wantParams: nil,
			wantErr:    false,
		},
		{
			name:        "unclosed double quote",
			command:     `/fix-bug "unclosed`,
			wantFound:   false,
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "unclosed quote",
		},
		{
			name:      "command in middle of string",
			command:   "Please /fix-bug 123 today",
			wantFound: true,
			wantTask:  "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": "123 today",
				"1":         "123",
				"2":         "today",
			},
			wantErr: false,
		},
		{
			name:      "command followed by newline",
			command:   "Text before /fix-bug 123\nText after on next line",
			wantFound: true,
			wantTask:  "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": "123",
				"1":         "123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, gotParams, gotFound, err := parseSlashCommand(tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSlashCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseSlashCommand() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if gotFound != tt.wantFound {
				t.Errorf("parseSlashCommand() gotFound = %v, want %v", gotFound, tt.wantFound)
			}

			if gotTask != tt.wantTask {
				t.Errorf("parseSlashCommand() gotTask = %v, want %v", gotTask, tt.wantTask)
			}

			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("parseSlashCommand() gotParams = %v, want %v", gotParams, tt.wantParams)
			}
		})
	}
}
