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
			expression: "env == 'production'",
			wantErr:    false,
		},
		{
			name:       "valid AND expression",
			expression: "env == 'production' && language == 'Go'",
			wantErr:    false,
		},
		{
			name:       "valid OR expression",
			expression: "language == 'Go' || language == 'Python'",
			wantErr:    false,
		},
		{
			name:       "valid nested field access",
			expression: "stage == 'implementation'",
			wantErr:    false,
		},
		{
			name:       "empty expression defaults to true",
			expression: "",
			wantErr:    false,
		},
		{
			name:       "invalid syntax",
			expression: "env ==",
			wantErr:    true,
		},
		{
			name:       "non-boolean expression",
			expression: "env",
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

			if !tt.wantErr {
				// Empty expression should be converted to "true"
				expectedExpr := tt.expression
				if expectedExpr == "" {
					expectedExpr = "true"
				}
				if s.expression != expectedExpr {
					t.Errorf("Set() s.expression = %q, want %q", s.expression, expectedExpr)
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
			expression:  "env == 'production'",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "simple equality - no match",
			expression:  "env == 'production'",
			frontmatter: frontMatter{"env": "development"},
			wantMatch:   false,
		},
		{
			name:        "simple equality - missing field (allowed)",
			expression:  "env == 'production'",
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true, // Missing fields match for backward compatibility
		},
		{
			name:        "AND expression - all match",
			expression:  "env == 'production' && language == 'Go'",
			frontmatter: frontMatter{"env": "production", "language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "AND expression - one doesn't match",
			expression:  "env == 'production' && language == 'Go'",
			frontmatter: frontMatter{"env": "production", "language": "Python"},
			wantMatch:   false,
		},
		{
			name:        "AND expression - one field missing (allowed)",
			expression:  "env == 'production' && language == 'Go'",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true, // Missing fields match for backward compatibility
		},
		{
			name:        "OR expression - first matches",
			expression:  "language == 'Go' || language == 'Python'",
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "OR expression - second matches",
			expression:  "language == 'Go' || language == 'Python'",
			frontmatter: frontMatter{"language": "Python"},
			wantMatch:   true,
		},
		{
			name:        "OR expression - neither matches",
			expression:  "language == 'Go' || language == 'Python'",
			frontmatter: frontMatter{"language": "JavaScript"},
			wantMatch:   false,
		},
		{
			name:        "empty expression - always match (defaults to true)",
			expression:  "",
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "task_name check - match",
			expression:  "task_name == 'deploy'",
			frontmatter: frontMatter{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "task_name check - no match",
			expression:  "task_name == 'deploy'",
			frontmatter: frontMatter{"task_name": "test"},
			wantMatch:   false,
		},
		{
			name:        "boolean value - match",
			expression:  "is_active == true",
			frontmatter: frontMatter{"is_active": true},
			wantMatch:   true,
		},
		{
			name:        "boolean value - no match",
			expression:  "is_active == true",
			frontmatter: frontMatter{"is_active": false},
			wantMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := selector{}
			if err := s.Set(tt.expression); err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			if got := s.matchesIncludes(tt.frontmatter); got != tt.wantMatch {
				t.Errorf("matchesIncludes() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSelector_String(t *testing.T) {
	s := selector{}
	s.Set("env == 'production'")

	str := s.String()
	if str != "env == 'production'" {
		t.Errorf("String() = %q, want %q", str, "env == 'production'")
	}
}
