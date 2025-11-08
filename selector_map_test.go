package main

import (
	"testing"
)

func TestSelector_Set(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{
			name:       "valid simple equality",
			expression: "frontmatter.env == 'production'",
			wantErr:    false,
		},
		{
			name:       "valid AND expression",
			expression: "frontmatter.env == 'production' && frontmatter.language == 'Go'",
			wantErr:    false,
		},
		{
			name:       "valid OR expression",
			expression: "frontmatter.language == 'Go' || frontmatter.language == 'Python'",
			wantErr:    false,
		},
		{
			name:       "valid nested field access",
			expression: "frontmatter.stage == 'implementation'",
			wantErr:    false,
		},
		{
			name:       "empty expression",
			expression: "",
			wantErr:    false,
		},
		{
			name:       "invalid syntax",
			expression: "frontmatter.env ==",
			wantErr:    true,
		},
		{
			name:       "non-boolean expression",
			expression: "frontmatter.env",
			wantErr:    true,
		},
		{
			name:       "invalid field reference",
			expression: "invalid_var == 'test'",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := selector{}
			err := s.Set(tt.expression)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.expression != "" {
				if s.expression != tt.expression {
					t.Errorf("Set() s.expression = %q, want %q", s.expression, tt.expression)
				}
				if s.program == nil {
					t.Errorf("Set() s.program is nil, expected non-nil")
				}
			}
		})
	}
}

func TestSelector_MatchesIncludes(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		frontmatter frontMatter
		wantMatch   bool
	}{
		{
			name:        "simple equality - match",
			expression:  "frontmatter.env == 'production'",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "simple equality - no match",
			expression:  "frontmatter.env == 'production'",
			frontmatter: frontMatter{"env": "development"},
			wantMatch:   false,
		},
		{
			name:        "simple equality - missing field (allowed)",
			expression:  "frontmatter.env == 'production'",
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true, // Missing fields match for backward compatibility
		},
		{
			name:        "AND expression - all match",
			expression:  "frontmatter.env == 'production' && frontmatter.language == 'Go'",
			frontmatter: frontMatter{"env": "production", "language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "AND expression - one doesn't match",
			expression:  "frontmatter.env == 'production' && frontmatter.language == 'Go'",
			frontmatter: frontMatter{"env": "production", "language": "Python"},
			wantMatch:   false,
		},
		{
			name:        "AND expression - one field missing (allowed)",
			expression:  "frontmatter.env == 'production' && frontmatter.language == 'Go'",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true, // Missing fields match for backward compatibility
		},
		{
			name:        "OR expression - first matches",
			expression:  "frontmatter.language == 'Go' || frontmatter.language == 'Python'",
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "OR expression - second matches",
			expression:  "frontmatter.language == 'Go' || frontmatter.language == 'Python'",
			frontmatter: frontMatter{"language": "Python"},
			wantMatch:   true,
		},
		{
			name:        "OR expression - neither matches",
			expression:  "frontmatter.language == 'Go' || frontmatter.language == 'Python'",
			frontmatter: frontMatter{"language": "JavaScript"},
			wantMatch:   false,
		},
		{
			name:        "empty expression - always match",
			expression:  "",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "task_name check - match",
			expression:  "frontmatter.task_name == 'deploy'",
			frontmatter: frontMatter{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "task_name check - no match",
			expression:  "frontmatter.task_name == 'deploy'",
			frontmatter: frontMatter{"task_name": "test"},
			wantMatch:   false,
		},
		{
			name:        "boolean value - match",
			expression:  "frontmatter.is_active == true",
			frontmatter: frontMatter{"is_active": true},
			wantMatch:   true,
		},
		{
			name:        "boolean value - no match",
			expression:  "frontmatter.is_active == true",
			frontmatter: frontMatter{"is_active": false},
			wantMatch:   false,
		},
		{
			name:        "has() function - field exists",
			expression:  "has(frontmatter.env)",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "has() function - field missing",
			expression:  "has(frontmatter.env)",
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   false,
		},
		{
			name:        "complex expression with has()",
			expression:  "has(frontmatter.language) && frontmatter.language == 'Go'",
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "complex expression - field missing, has() prevents error",
			expression:  "has(frontmatter.language) && frontmatter.language == 'Go'",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   false, // has() returns false, so AND is false (no error thrown)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := selector{}
			if tt.expression != "" {
				if err := s.Set(tt.expression); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			if got := s.matchesIncludes(tt.frontmatter); got != tt.wantMatch {
				t.Errorf("matchesIncludes() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSelector_String(t *testing.T) {
	s := selector{}
	s.Set("frontmatter.env == 'production'")

	str := s.String()
	if str != "frontmatter.env == 'production'" {
		t.Errorf("String() = %q, want %q", str, "frontmatter.env == 'production'")
	}
}
