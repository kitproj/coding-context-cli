package codingcontext

import (
	"strings"
	"testing"
)

func TestParseTaskPrompt(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		wantHasSlashCommand  bool
		wantFirstCommandName string
		wantParams           map[string]string
		wantAllTextContains  []string
	}{
		{
			name:                 "simple slash command",
			input:                "/fix-bug 123\n",
			wantHasSlashCommand:  true,
			wantFirstCommandName: "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": "123",
				"1":         "123",
			},
			wantAllTextContains: []string{"/fix-bug 123"},
		},
		{
			name:                 "plain text without slash command",
			input:                "Just some plain text\n",
			wantHasSlashCommand:  false,
			wantFirstCommandName: "",
			wantParams:           nil,
			wantAllTextContains:  []string{"Just some plain text"},
		},
		{
			name: "text followed by slash command",
			input: `Please fix the bug
/fix-bug 123
`,
			wantHasSlashCommand:  true,
			wantFirstCommandName: "fix-bug",
			wantParams: map[string]string{
				"ARGUMENTS": "123",
				"1":         "123",
			},
			wantAllTextContains: []string{"Please fix the bug", "/fix-bug 123"},
		},
		{
			name:                 "slash command with named parameter",
			input:                "/deploy env=production\n",
			wantHasSlashCommand:  true,
			wantFirstCommandName: "deploy",
			wantParams: map[string]string{
				"ARGUMENTS": "env=production",
				"1":         "env=production",
				"env":       "production",
			},
			wantAllTextContains: []string{"/deploy env=production"},
		},
		{
			name:                 "slash command with quoted value",
			input:                "/implement feature=\"Add auth\"\n",
			wantHasSlashCommand:  true,
			wantFirstCommandName: "implement",
			wantParams: map[string]string{
				"ARGUMENTS": "feature=\"Add auth\"",
				"1":         "feature=\"Add auth\"",
				"feature":   "Add auth",
			},
			wantAllTextContains: []string{"/implement", "Add auth"},
		},
		{
			name: "multiple slash commands - should extract first",
			input: `/init project
/config name=myapp
`,
			wantHasSlashCommand:  true,
			wantFirstCommandName: "init",
			wantParams: map[string]string{
				"ARGUMENTS": "project",
				"1":         "project",
			},
			wantAllTextContains: []string{"/init project", "/config name=myapp"},
		},
		{
			name:                 "mixed positional and named parameters",
			input:                "/task arg1 key=value arg2\n",
			wantHasSlashCommand:  true,
			wantFirstCommandName: "task",
			wantParams: map[string]string{
				"ARGUMENTS": "arg1 key=value arg2",
				"1":         "arg1",
				"2":         "key=value",
				"3":         "arg2",
				"key":       "value",
			},
			wantAllTextContains: []string{"/task arg1 key=value arg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTaskPrompt(tt.input)
			if err != nil {
				t.Fatalf("ParseTaskPrompt() error = %v", err)
			}

			if result.HasSlashCommand != tt.wantHasSlashCommand {
				t.Errorf("HasSlashCommand = %v, want %v", result.HasSlashCommand, tt.wantHasSlashCommand)
			}

			if result.FirstCommandName != tt.wantFirstCommandName {
				t.Errorf("FirstCommandName = %q, want %q", result.FirstCommandName, tt.wantFirstCommandName)
			}

			// Check parameters
			if tt.wantParams != nil {
				if result.FirstCommandParams == nil {
					t.Fatalf("FirstCommandParams is nil, want %v", tt.wantParams)
				}
				for key, wantValue := range tt.wantParams {
					gotValue, ok := result.FirstCommandParams[key]
					if !ok {
						t.Errorf("FirstCommandParams[%q] not found", key)
						continue
					}
					if gotValue != wantValue {
						t.Errorf("FirstCommandParams[%q] = %q, want %q", key, gotValue, wantValue)
					}
				}
				// Check that we don't have extra keys
				for key := range result.FirstCommandParams {
					if _, ok := tt.wantParams[key]; !ok {
						t.Errorf("FirstCommandParams has unexpected key %q = %q", key, result.FirstCommandParams[key])
					}
				}
			} else if result.FirstCommandParams != nil {
				t.Errorf("FirstCommandParams = %v, want nil", result.FirstCommandParams)
			}

			// Check that AllText contains expected substrings
			for _, substr := range tt.wantAllTextContains {
				if !strings.Contains(result.AllText, substr) {
					t.Errorf("AllText = %q, want to contain %q", result.AllText, substr)
				}
			}
		})
	}
}
