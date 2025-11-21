package slashcommand

import (
	"reflect"
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
			name:      "command with single-quoted argument containing spaces",
			command:   `/code-review 'Fix authentication bug'`,
			wantFound: true,
			wantTask:  "code-review",
			wantParams: map[string]string{
				"ARGUMENTS": `'Fix authentication bug'`,
				"1":         "Fix authentication bug",
			},
			wantErr: false,
		},
		{
			name:      "command with mixed quoted and unquoted arguments",
			command:   `/deploy "staging server" v1.2.3 --force`,
			wantFound: true,
			wantTask:  "deploy",
			wantParams: map[string]string{
				"ARGUMENTS": `"staging server" v1.2.3 --force`,
				"1":         "staging server",
				"2":         "v1.2.3",
				"3":         "--force",
			},
			wantErr: false,
		},
		{
			name:      "command with extra whitespace",
			command:   `  /fix-bug   123   "high priority"  `,
			wantFound: true,
			wantTask:  "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": `123   "high priority"`,
				"1":         "123",
				"2":         "high priority",
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
			name:        "unclosed single quote",
			command:     `/fix-bug 'unclosed`,
			wantFound:   false,
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "unclosed quote",
		},
		{
			name:       "task name with hyphens",
			command:    "/implement-new-feature",
			wantFound:  true,
			wantTask:   "implement-new-feature",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "task name with underscores",
			command:    "/fix_critical_bug",
			wantFound:  true,
			wantTask:   "fix_critical_bug",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:      "empty quoted argument",
			command:   `/fix-bug ""`,
			wantFound: true,
			wantTask:  "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": `""`,
				"1":         "",
			},
			wantErr: false,
		},
		{
			name:      "argument with special characters",
			command:   `/deploy https://example.com/api/v1`,
			wantFound: true,
			wantTask:  "deploy",
			wantParams: map[string]string{
				"ARGUMENTS": "https://example.com/api/v1",
				"1":         "https://example.com/api/v1",
			},
			wantErr: false,
		},
		{
			name:      "argument with numbers",
			command:   `/review 12345`,
			wantFound: true,
			wantTask:  "review",
			wantParams: map[string]string{
				"ARGUMENTS": "12345",
				"1":         "12345",
			},
			wantErr: false,
		},
		{
			name:      "multiple arguments with various spacing",
			command:   `/task a  b   c`,
			wantFound: true,
			wantTask:  "task",
			wantParams: map[string]string{
				"ARGUMENTS": "a  b   c",
				"1":         "a",
				"2":         "b",
				"3":         "c",
			},
			wantErr: false,
		},
		{
			name:      "escaped quote in double quotes",
			command:   `/echo "He said \"hello\""`,
			wantFound: true,
			wantTask:  "echo",
			wantParams: map[string]string{
				"ARGUMENTS": `"He said \"hello\""`,
				"1":         `He said "hello"`,
			},
			wantErr: false,
		},
		{
			name:      "single quotes preserve everything",
			command:   `/echo 'He said "hello"'`,
			wantFound: true,
			wantTask:  "echo",
			wantParams: map[string]string{
				"ARGUMENTS": `'He said "hello"'`,
				"1":         `He said "hello"`,
			},
			wantErr: false,
		},
		// New test cases for finding slash commands anywhere in the string
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
			name:      "command with prefix text",
			command:   "Hey can you /deploy production",
			wantFound: true,
			wantTask:  "deploy",
			wantParams: map[string]string{
				"ARGUMENTS": "production",
				"1":         "production",
			},
			wantErr: false,
		},
		{
			name:      "command with text before and on same line",
			command:   "Some text /code-review \"PR #42\" high more text",
			wantFound: true,
			wantTask:  "code-review",
			wantParams: map[string]string{
				"ARGUMENTS": `"PR #42" high more text`,
				"1":         "PR #42",
				"2":         "high",
				"3":         "more",
				"4":         "text",
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
			name:      "command in middle without arguments",
			command:   "Some prefix /task-name and suffix",
			wantFound: true,
			wantTask:  "task-name",
			wantParams: map[string]string{
				"ARGUMENTS": "and suffix",
				"1":         "and",
				"2":         "suffix",
			},
			wantErr: false,
		},
		{
			name:       "command at end without args",
			command:    "Please run /deploy",
			wantFound:  true,
			wantTask:   "deploy",
			wantParams: map[string]string{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, gotParams, gotFound, err := ParseSlashCommand(tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSlashCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseSlashCommand() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if gotFound != tt.wantFound {
				t.Errorf("ParseSlashCommand() gotFound = %v, want %v", gotFound, tt.wantFound)
			}

			if gotTask != tt.wantTask {
				t.Errorf("ParseSlashCommand() gotTask = %v, want %v", gotTask, tt.wantTask)
			}

			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("ParseSlashCommand() gotParams = %v, want %v", gotParams, tt.wantParams)
			}
		})
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
