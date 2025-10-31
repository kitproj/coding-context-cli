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
		wantOp  selectorType
		wantErr bool
	}{
		{
			name:    "valid equals selector",
			value:   "env=production",
			wantKey: "env",
			wantVal: "production",
			wantOp:  selectorEquals,
			wantErr: false,
		},
		{
			name:    "valid not equals selector",
			value:   "env!=test",
			wantKey: "env",
			wantVal: "test",
			wantOp:  selectorNotEquals,
			wantErr: false,
		},
		{
			name:    "equals with spaces",
			value:   "env = production",
			wantKey: "env",
			wantVal: "production",
			wantOp:  selectorEquals,
			wantErr: false,
		},
		{
			name:    "not equals with spaces",
			value:   "env != test",
			wantKey: "env",
			wantVal: "test",
			wantOp:  selectorNotEquals,
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
			var s selectorMap
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
				if s[0].key != tt.wantKey {
					t.Errorf("Set() key = %q, want %q", s[0].key, tt.wantKey)
				}
				if s[0].value != tt.wantVal {
					t.Errorf("Set() value = %q, want %q", s[0].value, tt.wantVal)
				}
				if s[0].op != tt.wantOp {
					t.Errorf("Set() op = %v, want %v", s[0].op, tt.wantOp)
				}
			}
		})
	}
}

func TestSelectorMap_SetMultiple(t *testing.T) {
	var s selectorMap
	if err := s.Set("env=production"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := s.Set("language!=python"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if len(s) != 2 {
		t.Errorf("Set() resulted in %d selectors, want 2", len(s))
	}
}

func TestSelectorMap_Matches(t *testing.T) {
	tests := []struct {
		name        string
		selectors   []string
		frontmatter map[string]string
		wantMatch   bool
	}{
		{
			name:        "single equals - match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "single equals - no match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"env": "development"},
			wantMatch:   false,
		},
		{
			name:        "single equals - key missing",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{"language": "go"},
			wantMatch:   false,
		},
		{
			name:        "single not equals - match (different value)",
			selectors:   []string{"env!=production"},
			frontmatter: map[string]string{"env": "development"},
			wantMatch:   true,
		},
		{
			name:        "single not equals - no match (same value)",
			selectors:   []string{"env!=production"},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   false,
		},
		{
			name:        "single not equals - match (key missing)",
			selectors:   []string{"env!=production"},
			frontmatter: map[string]string{"language": "go"},
			wantMatch:   true,
		},
		{
			name:        "multiple selectors - all match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "go"},
			wantMatch:   true,
		},
		{
			name:        "multiple selectors - one doesn't match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: map[string]string{"env": "production", "language": "python"},
			wantMatch:   false,
		},
		{
			name:        "mixed operators - all match",
			selectors:   []string{"env=production", "language!=python"},
			frontmatter: map[string]string{"env": "production", "language": "go"},
			wantMatch:   true,
		},
		{
			name:        "mixed operators - one doesn't match",
			selectors:   []string{"env=production", "language!=python"},
			frontmatter: map[string]string{"env": "production", "language": "python"},
			wantMatch:   false,
		},
		{
			name:        "empty selectors - always match",
			selectors:   []string{},
			frontmatter: map[string]string{"env": "production"},
			wantMatch:   true,
		},
		{
			name:        "empty frontmatter - equals doesn't match",
			selectors:   []string{"env=production"},
			frontmatter: map[string]string{},
			wantMatch:   false,
		},
		{
			name:        "empty frontmatter - not equals matches",
			selectors:   []string{"env!=production"},
			frontmatter: map[string]string{},
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s selectorMap
			for _, sel := range tt.selectors {
				if err := s.Set(sel); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			if got := s.matches(tt.frontmatter); got != tt.wantMatch {
				t.Errorf("matches() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSelectorMap_String(t *testing.T) {
	var s selectorMap
	s.Set("env=production")
	s.Set("language!=python")
	
	str := s.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}
