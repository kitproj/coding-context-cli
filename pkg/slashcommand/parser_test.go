package slashcommand

import (
	"reflect"
	"testing"
)

func TestParseSlashCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		wantTask    string
		wantParams  map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:       "simple command without parameters",
			command:    "/fix-bug",
			wantTask:   "fix-bug",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "command with single parameter",
			command:    `/fix-bug issue_number="123"`,
			wantTask:   "fix-bug",
			wantParams: map[string]string{"issue_number": "123"},
			wantErr:    false,
		},
		{
			name:       "command with multiple parameters",
			command:    `/implement-feature feature_name="User Login" priority="high"`,
			wantTask:   "implement-feature",
			wantParams: map[string]string{"feature_name": "User Login", "priority": "high"},
			wantErr:    false,
		},
		{
			name:       "command with quoted value containing spaces",
			command:    `/code-review pr_title="Fix authentication bug in login flow"`,
			wantTask:   "code-review",
			wantParams: map[string]string{"pr_title": "Fix authentication bug in login flow"},
			wantErr:    false,
		},
		{
			name:       "command with extra whitespace",
			command:    `  /fix-bug   issue_number="123"   `,
			wantTask:   "fix-bug",
			wantParams: map[string]string{"issue_number": "123"},
			wantErr:    false,
		},
		{
			name:       "command with extra whitespace around parameters",
			command:    `/fix-bug  issue_number="123"   title="Bug fix"  `,
			wantTask:   "fix-bug",
			wantParams: map[string]string{"issue_number": "123", "title": "Bug fix"},
			wantErr:    false,
		},
		{
			name:        "missing leading slash",
			command:     "fix-bug",
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "must start with '/'",
		},
		{
			name:        "empty command",
			command:     "/",
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "empty string",
			command:     "",
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "must start with '/'",
		},
		{
			name:        "parameter without quotes",
			command:     "/fix-bug issue_number=123",
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "must be quoted",
		},
		{
			name:        "parameter with single quotes (should fail)",
			command:     "/fix-bug issue_number='123'",
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "must be quoted",
		},
		{
			name:        "unclosed quote",
			command:     `/fix-bug issue_number="123`,
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "unclosed quote",
		},
		{
			name:        "invalid parameter format (no equals sign)",
			command:     `/fix-bug invalid_param`,
			wantTask:    "",
			wantParams:  nil,
			wantErr:     true,
			errContains: "invalid parameter format",
		},
		{
			name:       "task name with hyphens",
			command:    "/implement-new-feature",
			wantTask:   "implement-new-feature",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "task name with underscores",
			command:    "/fix_critical_bug",
			wantTask:   "fix_critical_bug",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "empty parameter value",
			command:    `/fix-bug issue_number=""`,
			wantTask:   "fix-bug",
			wantParams: map[string]string{"issue_number": ""},
			wantErr:    false,
		},
		{
			name:       "parameter with special characters",
			command:    `/deploy url="https://example.com/api/v1"`,
			wantTask:   "deploy",
			wantParams: map[string]string{"url": "https://example.com/api/v1"},
			wantErr:    false,
		},
		{
			name:       "parameter with numbers",
			command:    `/review pr_number="12345"`,
			wantTask:   "review",
			wantParams: map[string]string{"pr_number": "12345"},
			wantErr:    false,
		},
		{
			name:       "multiple parameters with various spacing",
			command:    `/task a="1"  b="2"   c="3"`,
			wantTask:   "task",
			wantParams: map[string]string{"a": "1", "b": "2", "c": "3"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, gotParams, err := ParseSlashCommand(tt.command)

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
