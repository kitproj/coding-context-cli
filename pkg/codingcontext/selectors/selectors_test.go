package selectors

import (
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
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
			s := make(Selectors)
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
				if !s.GetValue(tt.wantKey, tt.wantVal) {
					t.Errorf("Set() s[%q] does not contain value %q", tt.wantKey, tt.wantVal)
				}
			}
		})
	}
}

func TestSelectorMap_SetMultiple(t *testing.T) {
	s := make(Selectors)
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
		name           string
		selectors      []string
		setupSelectors func(s Selectors) // Optional function to set up array selectors directly
		frontmatter    markdown.BaseFrontMatter
		wantMatch      bool
	}{
		{
			name:        "single selector - match",
			selectors:   []string{"env=production"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "production"}},
			wantMatch:   true,
		},
		{
			name:        "single selector - no match",
			selectors:   []string{"env=production"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "development"}},
			wantMatch:   false,
		},
		{
			name:        "single selector - key missing (allowed)",
			selectors:   []string{"env=production"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"language": "go"}},
			wantMatch:   true,
		},
		{
			name:        "multiple selectors - all match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "production", "language": "go"}},
			wantMatch:   true,
		},
		{
			name:        "multiple selectors - one doesn't match",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "production", "language": "python"}},
			wantMatch:   false,
		},
		{
			name:        "multiple selectors - one key missing (allowed)",
			selectors:   []string{"env=production", "language=go"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "production"}},
			wantMatch:   true,
		},
		{
			name:        "empty selectors - always match",
			selectors:   []string{},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "production"}},
			wantMatch:   true,
		},
		{
			name:        "boolean value conversion - match",
			selectors:   []string{"is_active=true"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"is_active": true}},
			wantMatch:   true,
		},
		{
			name:        "array selector - match",
			selectors:   []string{},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"rule_name": "rule2"}},
			wantMatch:   true,
			setupSelectors: func(s Selectors) {
				s.SetValue("rule_name", "rule1")
				s.SetValue("rule_name", "rule2")
				s.SetValue("rule_name", "rule3")
			},
		},
		{
			name:        "array selector - no match",
			selectors:   []string{},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"rule_name": "rule4"}},
			wantMatch:   false,
			setupSelectors: func(s Selectors) {
				s.SetValue("rule_name", "rule1")
				s.SetValue("rule_name", "rule2")
				s.SetValue("rule_name", "rule3")
			},
		},
		{
			name:        "array selector - key missing (allowed)",
			selectors:   []string{},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "prod"}},
			wantMatch:   true,
			setupSelectors: func(s Selectors) {
				s.SetValue("rule_name", "rule1")
				s.SetValue("rule_name", "rule2")
			},
		},
		{
			name:        "mixed selectors - array and string both match",
			selectors:   []string{"env=prod"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "prod", "rule_name": "rule1"}},
			wantMatch:   true,
			setupSelectors: func(s Selectors) {
				s.SetValue("rule_name", "rule1")
				s.SetValue("rule_name", "rule2")
			},
		},
		{
			name:        "mixed selectors - string doesn't match",
			selectors:   []string{"env=dev"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "prod", "rule_name": "rule1"}},
			wantMatch:   false,
			setupSelectors: func(s Selectors) {
				s.SetValue("rule_name", "rule1")
				s.SetValue("rule_name", "rule2")
			},
		},
		{
			name:        "multiple array selectors - both match",
			selectors:   []string{},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"rule_name": "rule1", "language": "go"}},
			wantMatch:   true,
			setupSelectors: func(s Selectors) {
				s.SetValue("rule_name", "rule1")
				s.SetValue("rule_name", "rule2")
				s.SetValue("language", "go")
				s.SetValue("language", "python")
			},
		},
		{
			name:        "multiple array selectors - one doesn't match",
			selectors:   []string{},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"rule_name": "rule1", "language": "java"}},
			wantMatch:   false,
			setupSelectors: func(s Selectors) {
				s.SetValue("rule_name", "rule1")
				s.SetValue("rule_name", "rule2")
				s.SetValue("language", "go")
				s.SetValue("language", "python")
			},
		},
		{
			name:        "OR logic - same key multiple values matches",
			selectors:   []string{"env=prod", "env=dev"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "dev"}},
			wantMatch:   true,
		},
		{
			name:        "OR logic - same key multiple values no match",
			selectors:   []string{"env=prod", "env=dev"},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "staging"}},
			wantMatch:   false,
		},
		{
			name:        "empty value selector - key exists in frontmatter (no match)",
			selectors:   []string{"env="},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"env": "production"}},
			wantMatch:   false,
		},
		{
			name:        "empty value selector - key missing in frontmatter (match)",
			selectors:   []string{"env="},
			frontmatter: markdown.BaseFrontMatter{Content: map[string]any{"language": "go"}},
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := make(Selectors)
			for _, sel := range tt.selectors {
				if err := s.Set(sel); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			// Set up array selectors if provided
			if tt.setupSelectors != nil {
				tt.setupSelectors(s)
			}

			if got := s.MatchesIncludes(tt.frontmatter); got != tt.wantMatch {
				t.Errorf("MatchesIncludes() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestSelectorMap_String(t *testing.T) {
	s := make(Selectors)
	s.Set("env=production")
	s.Set("language=go")

	str := s.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}
