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
		{
			name:       "command with leading period and spaces",
			command:    ".   /taskname",
			wantFound:  true,
			wantTask:   "taskname",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "command with leading period and more spaces",
			command:    ".    /taskname",
			wantFound:  true,
			wantTask:   "taskname",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:      "command with leading period, spaces and arguments",
			command:   ".   /fix-bug PROJ-123",
			wantFound: true,
			wantTask:  "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": "PROJ-123",
				"1":         "PROJ-123",
			},
			wantErr: false,
		},
		{
			name:       "command with leading period, spaces, and newline",
			command:    ".    /taskname\n",
			wantFound:  true,
			wantTask:   "taskname",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		// Named parameter tests
		{
			name:      "command with single named parameter",
			command:   `/fix-bug issue="PROJ-123"`,
			wantFound: true,
			wantTask:  "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": `issue="PROJ-123"`,
				"1":         `issue="PROJ-123"`,
				"issue":     "PROJ-123",
			},
			wantErr: false,
		},
		{
			name:      "command with multiple named parameters",
			command:   `/deploy env="production" version="1.2.3"`,
			wantFound: true,
			wantTask:  "deploy",
			wantParams: map[string]string{
				"ARGUMENTS": `env="production" version="1.2.3"`,
				"1":         `env="production"`,
				"2":         `version="1.2.3"`,
				"env":       "production",
				"version":   "1.2.3",
			},
			wantErr: false,
		},
		{
			name:      "command with mixed positional and named parameters",
			command:   `/task arg1 key="value" arg2`,
			wantFound: true,
			wantTask:  "task",
			wantParams: map[string]string{
				"ARGUMENTS": `arg1 key="value" arg2`,
				"1":         "arg1",
				"2":         `key="value"`,
				"3":         "arg2",
				"key":       "value",
			},
			wantErr: false,
		},
		{
			name:      "named parameter with spaces in value",
			command:   `/implement feature="Add user authentication"`,
			wantFound: true,
			wantTask:  "implement",
			wantParams: map[string]string{
				"ARGUMENTS": `feature="Add user authentication"`,
				"1":         `feature="Add user authentication"`,
				"feature":   "Add user authentication",
			},
			wantErr: false,
		},
		{
			name:      "named parameter with escaped quotes in value",
			command:   `/log message="User said \"hello\""`,
			wantFound: true,
			wantTask:  "log",
			wantParams: map[string]string{
				"ARGUMENTS": `message="User said \"hello\""`,
				"1":         `message="User said \"hello\""`,
				"message":   `User said "hello"`,
			},
			wantErr: false,
		},
		{
			name:      "positional before and after named parameter",
			command:   `/task before key="middle" after`,
			wantFound: true,
			wantTask:  "task",
			wantParams: map[string]string{
				"ARGUMENTS": `before key="middle" after`,
				"1":         "before",
				"2":         `key="middle"`,
				"3":         "after",
				"key":       "middle",
			},
			wantErr: false,
		},
		{
			name:      "multiple named parameters with different types of values",
			command:   `/config host="localhost" port="8080" debug="true"`,
			wantFound: true,
			wantTask:  "config",
			wantParams: map[string]string{
				"ARGUMENTS": `host="localhost" port="8080" debug="true"`,
				"1":         `host="localhost"`,
				"2":         `port="8080"`,
				"3":         `debug="true"`,
				"host":      "localhost",
				"port":      "8080",
				"debug":     "true",
			},
			wantErr: false,
		},
		{
			name:      "named parameter with empty value",
			command:   `/task key=""`,
			wantFound: true,
			wantTask:  "task",
			wantParams: map[string]string{
				"ARGUMENTS": `key=""`,
				"1":         `key=""`,
				"key":       "",
			},
			wantErr: false,
		},
		{
			name:      "named parameter with equals sign in value",
			command:   `/run equation="x=y+z"`,
			wantFound: true,
			wantTask:  "run",
			wantParams: map[string]string{
				"ARGUMENTS": `equation="x=y+z"`,
				"1":         `equation="x=y+z"`,
				"equation":  "x=y+z",
			},
			wantErr: false,
		},
		// Edge case tests for named parameters
		{
			name:      "numeric key in named parameter is ignored",
			command:   `/task arg1 1="override"`,
			wantFound: true,
			wantTask:  "task",
			wantParams: map[string]string{
				"ARGUMENTS": `arg1 1="override"`,
				"1":         "arg1",
				"2":         `1="override"`,
			},
			wantErr: false,
		},
		{
			name:      "ARGUMENTS key in named parameter is ignored",
			command:   `/task arg1 ARGUMENTS="custom"`,
			wantFound: true,
			wantTask:  "task",
			wantParams: map[string]string{
				"ARGUMENTS": `arg1 ARGUMENTS="custom"`,
				"1":         "arg1",
				"2":         `ARGUMENTS="custom"`,
			},
			wantErr: false,
		},
		{
			name:      "duplicate named parameter keys - last value wins",
			command:   `/task key="first" key="second"`,
			wantFound: true,
			wantTask:  "task",
			wantParams: map[string]string{
				"ARGUMENTS": `key="first" key="second"`,
				"1":         `key="first"`,
				"2":         `key="second"`,
				"key":       "second",
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
