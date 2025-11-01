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
				if s[tt.wantKey] != tt.wantVal {
					t.Errorf("Set() s[%q] = %q, want %q", tt.wantKey, s[tt.wantKey], tt.wantVal)
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
}

func TestSelectorMap_MatchesIncludes(t *testing.T) {
	tests := []struct {
		name        string
		selectors   []string
		frontmatter map[string]string
		builtins    map[string]string
		wantMatch   bool
	}{
		{
			name:        "single include - match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "single include - no match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "development"},
			builtins:    map[string]string{},
			wantMatch:   false,
		},
		{
			name:        "single include - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"language": "go"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "multiple includes - all match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "go"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "multiple includes - one doesn't match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "python"},
			builtins:    map[string]string{},
			wantMatch:   false,
		},
		{
			name:        "multiple includes - one key missing (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "empty includes - always match",
			selectors:   []string{},
			frontmatter: map[string]string{"env": "production"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "builtin task_name - match",
			selectors:   []string{},
			frontmatter: map[string]string{"task_name": "deploy"},
			builtins:    map[string]string{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "builtin task_name - no match",
			selectors:   []string{},
			frontmatter: map[string]string{"task_name": "test"},
			builtins:    map[string]string{"task_name": "deploy"},
			wantMatch:   false,
		},
		{
			name:        "builtin task_name - key missing (allowed)",
			selectors:   []string{},
			frontmatter: map[string]string{"env": "production"},
			builtins:    map[string]string{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "builtin task_name with selector - both match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production", "task_name": "deploy"},
			builtins:    map[string]string{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "builtin task_name with selector - selector matches, task doesn't",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production", "task_name": "test"},
			builtins:    map[string]string{"task_name": "deploy"},
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

			if got := s.matchesIncludes(tt.frontmatter, tt.builtins); got != tt.wantMatch {
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
		builtins    map[string]string
		wantMatch   bool
	}{
		{
			name:        "single exclude - doesn't match (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "development"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "single exclude - matches (excluded)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production"},
			builtins:    map[string]string{},
			wantMatch:   false,
		},
		{
			name:        "single exclude - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"language": "go"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "multiple excludes - none match (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "development", "language": "python"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "multiple excludes - one matches (excluded)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "python"},
			builtins:    map[string]string{},
			wantMatch:   false,
		},
		{
			name:        "multiple excludes - one key missing (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "development"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "empty excludes - always match",
			selectors:   []string{},
			frontmatter: map[string]string{"env": "production"},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{},
			builtins:    map[string]string{},
			wantMatch:   true,
		},
		{
			name:        "builtin task_name - doesn't affect excludes (allowed)",
			selectors:   []string{},
			frontmatter: map[string]string{"task_name": "deploy"},
			builtins:    map[string]string{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "builtin task_name with different value - doesn't affect excludes (allowed)",
			selectors:   []string{},
			frontmatter: map[string]string{"task_name": "test"},
			builtins:    map[string]string{"task_name": "deploy"},
			wantMatch:   true,
		},
		{
			name:        "builtin and explicit exclude - only explicit exclude matters",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production", "task_name": "deploy"},
			builtins:    map[string]string{"task_name": "deploy"},
			wantMatch:   false,
		},
		{
			name:        "builtin without explicit exclude - allowed",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "development", "task_name": "deploy"},
			builtins:    map[string]string{"task_name": "deploy"},
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

			if got := s.matchesExcludes(tt.frontmatter, tt.builtins); got != tt.wantMatch {
				t.Errorf("matchesExcludes() = %v, want %v", got, tt.wantMatch)
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
