package main

import (
	"testing"
)

func TestSelectorMap_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantKey string
		wantOp  selectorOperator
		wantVal string
		wantErr bool
	}{
		{
			name:    "valid selector - equals",
			value:   "env=production",
			wantKey: "env",
			wantOp:  selectorEquals,
			wantVal: "production",
			wantErr: false,
		},
		{
			name:    "valid selector - includes",
			value:   "language:=Go",
			wantKey: "language",
			wantOp:  selectorIncludes,
			wantVal: "Go",
			wantErr: false,
		},
		{
			name:    "valid selector - not equals",
			value:   "env!=staging",
			wantKey: "env",
			wantOp:  selectorNotEquals,
			wantVal: "staging",
			wantErr: false,
		},
		{
			name:    "valid selector - not includes",
			value:   "language!:Python",
			wantKey: "language",
			wantOp:  selectorNotIncludes,
			wantVal: "Python",
			wantErr: false,
		},
		{
			name:    "selector with spaces",
			value:   "env = production",
			wantKey: "env",
			wantOp:  selectorEquals,
			wantVal: "production",
			wantErr: false,
		},
		{
			name:    "selector with spaces - includes",
			value:   "language := Go",
			wantKey: "language",
			wantOp:  selectorIncludes,
			wantVal: "Go",
			wantErr: false,
		},
		{
			name:    "invalid format - no operator",
			value:   "env",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make(selectorMap)
			err := s.Set(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(s) != 1 {
					t.Errorf("Set() resulted in %d selectors, want 1", len(s))
					return
				}
				sel := s[tt.wantKey]
				if sel.operator != tt.wantOp {
					t.Errorf("Set() operator = %q, want %q", sel.operator, tt.wantOp)
				}
				if sel.value != tt.wantVal {
					t.Errorf("Set() value = %q, want %q", sel.value, tt.wantVal)
				}
			}
		})
	}
}

func TestSelectorMap_SetMultiple(t *testing.T) {
	s := make(selectorMap)
	if err := s.Set("env=production"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := s.Set("language:=go"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if len(s) != 2 {
		t.Errorf("Set() resulted in %d selectors, want 2", len(s))
	}
}

func TestSelectorMap_MatchesIncludes(t *testing.T) {
	tests := []struct {
		name        string
		selectors   []string
		frontmatter frontMatter
		wantMatch   bool
	}{
		// Test equals operator
		{
			name:        "equals - match",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "equals - no match",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{"env": "development"},
			wantMatch:   false,
		},
		{
			name:        "equals - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{"language": "go"},
			wantMatch:   true,
		},
		{
			name:        "equals - boolean value match",
			selectors:   []string{"is_active=true"},
			frontmatter: frontMatter{"is_active": true},
			wantMatch:   true,
		},

		// Test includes operator with scalar
		{
			name:        "includes - scalar match",
			selectors:   []string{"language:=Go"},
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "includes - scalar no match",
			selectors:   []string{"language:=Go"},
			frontmatter: frontMatter{"language": "Python"},
			wantMatch:   false,
		},

		// Test includes operator with array
		{
			name:        "includes - array contains value",
			selectors:   []string{"language:=Go"},
			frontmatter: frontMatter{"language": []any{"Go", "TypeScript"}},
			wantMatch:   true,
		},
		{
			name:        "includes - array does not contain value",
			selectors:   []string{"language:=Python"},
			frontmatter: frontMatter{"language": []any{"Go", "TypeScript"}},
			wantMatch:   false,
		},
		{
			name:        "includes - array with single value",
			selectors:   []string{"language:=Go"},
			frontmatter: frontMatter{"language": []any{"Go"}},
			wantMatch:   true,
		},
		{
			name:        "includes - key missing (allowed)",
			selectors:   []string{"language:=Go"},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},

		// Test not equals operator
		{
			name:        "not equals - different value",
			selectors:   []string{"env!=staging"},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "not equals - same value (no match)",
			selectors:   []string{"env!=staging"},
			frontmatter: frontMatter{"env": "staging"},
			wantMatch:   false,
		},
		{
			name:        "not equals - key missing (allowed)",
			selectors:   []string{"env!=staging"},
			frontmatter: frontMatter{"language": "go"},
			wantMatch:   true,
		},

		// Test not includes operator with scalar
		{
			name:        "not includes - scalar different",
			selectors:   []string{"language!:Python"},
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "not includes - scalar same (no match)",
			selectors:   []string{"language!:Python"},
			frontmatter: frontMatter{"language": "Python"},
			wantMatch:   false,
		},

		// Test not includes operator with array
		{
			name:        "not includes - array does not contain",
			selectors:   []string{"language!:Python"},
			frontmatter: frontMatter{"language": []any{"Go", "TypeScript"}},
			wantMatch:   true,
		},
		{
			name:        "not includes - array contains (no match)",
			selectors:   []string{"language!:Go"},
			frontmatter: frontMatter{"language": []any{"Go", "TypeScript"}},
			wantMatch:   false,
		},
		{
			name:        "not includes - key missing (allowed)",
			selectors:   []string{"language!:Python"},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},

		// Test multiple selectors
		{
			name:        "multiple - all match",
			selectors:   []string{"env=production", "language:=Go"},
			frontmatter: frontMatter{"env": "production", "language": []any{"Go", "TypeScript"}},
			wantMatch:   true,
		},
		{
			name:        "multiple - one doesn't match",
			selectors:   []string{"env=production", "language:=Python"},
			frontmatter: frontMatter{"env": "production", "language": []any{"Go", "TypeScript"}},
			wantMatch:   false,
		},
		{
			name:        "multiple - mixed operators",
			selectors:   []string{"env=production", "language:=Go", "stage!=testing"},
			frontmatter: frontMatter{"env": "production", "language": []any{"Go", "TypeScript"}, "stage": "implementation"},
			wantMatch:   true,
		},

		// Edge cases
		{
			name:        "empty selectors - always match",
			selectors:   []string{},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - positive selectors (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - negative selectors (allowed)",
			selectors:   []string{"env!=staging"},
			frontmatter: frontMatter{},
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make(selectorMap)
			for _, sel := range tt.selectors {
				if err := s.Set(sel); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			if got := s.matchesIncludes(tt.frontmatter); got != tt.wantMatch {
				t.Errorf("matchesIncludes() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSelectorMap_String(t *testing.T) {
	s := make(selectorMap)
	s.Set("env=production")
	s.Set("language:=go")

	str := s.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}
