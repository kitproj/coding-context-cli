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
			s := make(selectorMap)
			for _, sel := range tt.selectors {
				if err := s.Set(sel); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			if got := s.matchesExcludes(tt.frontmatter); got != tt.wantMatch {
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

func TestSelectorMap_ExplainIncludes(t *testing.T) {
	tests := []struct {
		name              string
		selectors         []string
		frontmatter       map[string]string
		wantMatch         bool
		wantExplanation   string
		checkExplanation  bool
	}{
		{
			name:              "single include - match",
			selectors:         []string{"env=production"},
			frontmatter:       map[string]string{"env": "production"},
			wantMatch:         true,
			wantExplanation:   "matches env=production",
			checkExplanation:  true,
		},
		{
			name:              "single include - no match",
			selectors:         []string{"env=production"},
			frontmatter:       map[string]string{"env": "development"},
			wantMatch:         false,
			wantExplanation:   "does not match include selector(s): env=production (has env=development)",
			checkExplanation:  true,
		},
		{
			name:              "single include - key missing (allowed)",
			selectors:         []string{"env=production"},
			frontmatter:       map[string]string{"language": "go"},
			wantMatch:         true,
			wantExplanation:   "allows missing env (key not in frontmatter)",
			checkExplanation:  true,
		},
		{
			name:              "no selectors",
			selectors:         []string{},
			frontmatter:       map[string]string{"env": "production"},
			wantMatch:         true,
			wantExplanation:   "no include selectors specified",
			checkExplanation:  true,
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

			gotMatch, gotExplanation := s.explainIncludes(tt.frontmatter)
			if gotMatch != tt.wantMatch {
				t.Errorf("explainIncludes() match = %v, want %v", gotMatch, tt.wantMatch)
			}
			if tt.checkExplanation && gotExplanation != tt.wantExplanation {
				t.Errorf("explainIncludes() explanation = %q, want %q", gotExplanation, tt.wantExplanation)
			}
		})
	}
}

func TestSelectorMap_ExplainExcludes(t *testing.T) {
	tests := []struct {
		name              string
		selectors         []string
		frontmatter       map[string]string
		wantMatch         bool
		wantExplanation   string
		checkExplanation  bool
	}{
		{
			name:              "single exclude - doesn't match (allowed)",
			selectors:         []string{"env=production"},
			frontmatter:       map[string]string{"env": "development"},
			wantMatch:         true,
			wantExplanation:   "does not match exclude env!=production (has env=development)",
			checkExplanation:  true,
		},
		{
			name:              "single exclude - matches (excluded)",
			selectors:         []string{"env=production"},
			frontmatter:       map[string]string{"env": "production"},
			wantMatch:         false,
			wantExplanation:   "matches exclude selector(s): env=production",
			checkExplanation:  true,
		},
		{
			name:              "single exclude - key missing (allowed)",
			selectors:         []string{"env=production"},
			frontmatter:       map[string]string{"language": "go"},
			wantMatch:         true,
			wantExplanation:   "allows missing env (key not in frontmatter)",
			checkExplanation:  true,
		},
		{
			name:              "no selectors",
			selectors:         []string{},
			frontmatter:       map[string]string{"env": "production"},
			wantMatch:         true,
			wantExplanation:   "no exclude selectors specified",
			checkExplanation:  true,
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

			gotMatch, gotExplanation := s.explainExcludes(tt.frontmatter)
			if gotMatch != tt.wantMatch {
				t.Errorf("explainExcludes() match = %v, want %v", gotMatch, tt.wantMatch)
			}
			if tt.checkExplanation && gotExplanation != tt.wantExplanation {
				t.Errorf("explainExcludes() explanation = %q, want %q", gotExplanation, tt.wantExplanation)
			}
		})
	}
}
