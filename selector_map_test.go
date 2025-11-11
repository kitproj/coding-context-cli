package main

import (
	"testing"

	"github.com/kitproj/coding-context-cli/context"
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
			s := make(context.SelectorMap)
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
				if s[tt.wantKey] != tt.wantVal {
					t.Errorf("Set() s[%q] = %q, want %q", tt.wantKey, s[tt.wantKey], tt.wantVal)
				}
			}
		})
	}
}

func TestSelectorMap_SetMultiple(t *testing.T) {
	s := make(context.SelectorMap)
	if err := s.Set("env=production"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := s.Set("language=go"); err != nil {
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
		frontmatter map[string]string
		wantMatch   bool
	}{
		{
			name:        "single include - match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "single include - no match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "development"},
			wantMatch:   false,
		},
		{
			name:        "single include - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"language": "go"},
			wantMatch:   true,
		},
		{
			name:        "multiple includes - all match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "go"},
			wantMatch:   true,
		},
		{
			name:        "multiple includes - one doesn't match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "python"},
			wantMatch:   false,
		},
		{
			name:        "multiple includes - one key missing (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "empty includes - always match",
			selectors:   []string{},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "task_name include - match",
			selectors:   []string{"task_name=deploy"},
			frontmatter: map[string]string{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "task_name include - no match",
			selectors:   []string{"task_name=deploy"},
			frontmatter: map[string]string{"task_name": "test"},
			wantMatch:   false,
		},
		{
			name:        "task_name include - key missing (allowed)",
			selectors:   []string{"task_name=deploy"},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make(context.SelectorMap)
			for _, sel := range tt.selectors {
				if err := s.Set(sel); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			if got := (*context.SelectorMap)(&s).MatchesIncludes(tt.frontmatter); got != tt.wantMatch {
				t.Errorf("matchesIncludes() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSelectorMap_MatchesExcludes(t *testing.T) {
	tests := []struct {
		name        string
		selectors   []string
		frontmatter map[string]string
		wantMatch   bool
	}{
		{
			name:        "single exclude - doesn't match (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "development"},
			wantMatch:   true,
		},
		{
			name:        "single exclude - matches (excluded)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   false,
		},
		{
			name:        "single exclude - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"language": "go"},
			wantMatch:   true,
		},
		{
			name:        "multiple excludes - none match (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "development", "language": "python"},
			wantMatch:   true,
		},
		{
			name:        "multiple excludes - one matches (excluded)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "python"},
			wantMatch:   false,
		},
		{
			name:        "multiple excludes - one key missing (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "development"},
			wantMatch:   true,
		},
		{
			name:        "empty excludes - always match",
			selectors:   []string{},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{},
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make(context.SelectorMap)
			for _, sel := range tt.selectors {
				if err := s.Set(sel); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			if got := (*context.SelectorMap)(&s).MatchesExcludes(tt.frontmatter); got != tt.wantMatch {
				t.Errorf("matchesExcludes() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSelectorMap_String(t *testing.T) {
	s := make(context.SelectorMap)
	s.Set("env=production")
	s.Set("language=go")

	str := s.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}
