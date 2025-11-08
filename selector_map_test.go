package main

import (
	"testing"
)

func TestSelectorMap_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			name:    "valid selector",
			value:   "env=production",
			wantKey: "env",
			wantVal: "production",
			wantErr: false,
		},
		{
			name:    "selector with spaces",
			value:   "env = production",
			wantKey: "env",
			wantVal: "production",
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
				values := s[tt.wantKey]
				if len(values) != 1 || values[0] != tt.wantVal {
					t.Errorf("Set() s[%q] = %v, want [%q]", tt.wantKey, values, tt.wantVal)
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
	if err := s.Set("language=go"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if len(s) != 2 {
		t.Errorf("Set() resulted in %d selectors, want 2", len(s))
	}
	
	// Test setting the same key multiple times (inclusive selection)
	if err := s.Set("language=typescript"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	
	// Should still have 2 keys, but language should have 2 values
	if len(s) != 2 {
		t.Errorf("Set() resulted in %d keys, want 2", len(s))
	}
	
	languageValues := s["language"]
	if len(languageValues) != 2 {
		t.Errorf("language has %d values, want 2", len(languageValues))
	}
	
	if languageValues[0] != "go" || languageValues[1] != "typescript" {
		t.Errorf("language values = %v, want [go, typescript]", languageValues)
	}
}

func TestSelectorMap_MatchesIncludes(t *testing.T) {
	tests := []struct {
		name        string
		selectors   []string
		frontmatter frontMatter
		wantMatch   bool
	}{
		{
			name:        "single include - match",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "single include - no match",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{"env": "development"},
			wantMatch:   false,
		},
		{
			name:        "single include - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{"language": "go"},
			wantMatch:   true,
		},
		{
			name:        "multiple includes - all match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: frontMatter{"env": "production", "language": "go"},
			wantMatch:   true,
		},
		{
			name:        "multiple includes - one doesn't match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: frontMatter{"env": "production", "language": "python"},
			wantMatch:   false,
		},
		{
			name:        "multiple includes - one key missing (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "empty includes - always match",
			selectors:   []string{},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: frontMatter{},
			wantMatch:   true,
		},
		{
			name:        "task_name include - match",
			selectors:   []string{"task_name=deploy"},
			frontmatter: frontMatter{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "task_name include - no match",
			selectors:   []string{"task_name=deploy"},
			frontmatter: frontMatter{"task_name": "test"},
			wantMatch:   false,
		},
		{
			name:        "task_name include - key missing (allowed)",
			selectors:   []string{"task_name=deploy"},
			frontmatter: frontMatter{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "include a boolean value - match",
			selectors:   []string{"is_active=true"},
			frontmatter: frontMatter{"is_active": true},
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
	s.Set("language=go")

	str := s.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

// TestSelectorMap_InclusiveSelection tests the new inclusive selection behavior
// where multiple values for the same key use OR logic
func TestSelectorMap_InclusiveSelection(t *testing.T) {
	tests := []struct {
		name        string
		selectors   []string
		frontmatter frontMatter
		wantMatch   bool
	}{
		{
			name:        "multiple values for same key - match first",
			selectors:   []string{"language=Go", "language=Typescript"},
			frontmatter: frontMatter{"language": "Go"},
			wantMatch:   true,
		},
		{
			name:        "multiple values for same key - match second",
			selectors:   []string{"language=Go", "language=Typescript"},
			frontmatter: frontMatter{"language": "Typescript"},
			wantMatch:   true,
		},
		{
			name:        "multiple values for same key - no match",
			selectors:   []string{"language=Go", "language=Typescript"},
			frontmatter: frontMatter{"language": "Python"},
			wantMatch:   false,
		},
		{
			name:        "multiple values for same key - key missing (allowed)",
			selectors:   []string{"language=Go", "language=Typescript"},
			frontmatter: frontMatter{"stage": "implementation"},
			wantMatch:   true,
		},
		{
			name:        "multiple keys with multiple values - all match",
			selectors:   []string{"language=Go", "language=Typescript", "stage=implementation", "stage=testing"},
			frontmatter: frontMatter{"language": "Go", "stage": "implementation"},
			wantMatch:   true,
		},
		{
			name:        "multiple keys with multiple values - language match, stage match second",
			selectors:   []string{"language=Go", "language=Typescript", "stage=implementation", "stage=testing"},
			frontmatter: frontMatter{"language": "Typescript", "stage": "testing"},
			wantMatch:   true,
		},
		{
			name:        "multiple keys with multiple values - language no match",
			selectors:   []string{"language=Go", "language=Typescript", "stage=implementation"},
			frontmatter: frontMatter{"language": "Python", "stage": "implementation"},
			wantMatch:   false,
		},
		{
			name:        "multiple keys with multiple values - stage no match",
			selectors:   []string{"language=Go", "language=Typescript", "stage=implementation"},
			frontmatter: frontMatter{"language": "Go", "stage": "planning"},
			wantMatch:   false,
		},
		{
			name:        "three values for same key - match middle",
			selectors:   []string{"language=Go", "language=Typescript", "language=Python"},
			frontmatter: frontMatter{"language": "Typescript"},
			wantMatch:   true,
		},
		{
			name:        "mixed single and multiple values - all match",
			selectors:   []string{"language=Go", "language=Typescript", "env=production"},
			frontmatter: frontMatter{"language": "Go", "env": "production"},
			wantMatch:   true,
		},
		{
			name:        "mixed single and multiple values - single key no match",
			selectors:   []string{"language=Go", "language=Typescript", "env=production"},
			frontmatter: frontMatter{"language": "Go", "env": "staging"},
			wantMatch:   false,
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
				t.Errorf("matchesIncludes() = %v, want %v\nSelectors: %v\nFrontmatter: %v", got, tt.wantMatch, s, tt.frontmatter)
			}
		})
	}
}
