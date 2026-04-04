package selectors

import (
	"strings"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
)

func TestSelectorMap_Set(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

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
	t.Parallel()

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

type matchesIncludesCase struct {
	name             string
	selectors        []string
	setupSelectors   func(s Selectors)
	frontmatter      markdown.BaseFrontMatter
	excludeUnmatched bool // when true, unmatched rules are excluded; default false = include by default
	wantMatch        bool
}

func fm(content map[string]any) markdown.BaseFrontMatter {
	return markdown.BaseFrontMatter{Content: content}
}

func setupRuleNames(names ...string) func(s Selectors) {
	return func(s Selectors) {
		for _, n := range names {
			s.SetValue("rule_name", n)
		}
	}
}

func matchesIncludesCases() []matchesIncludesCase {
	return []matchesIncludesCase{
		{name: "single selector - match", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"env": "production"}), wantMatch: true},
		{name: "single selector - no match", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"env": "development"}), wantMatch: false},
		{name: "single selector - key missing (allowed)", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"language": "go"}), wantMatch: true},
		{name: "multiple selectors - all match", selectors: []string{"env=production", "language=go"},
			frontmatter: fm(map[string]any{"env": "production", "language": "go"}), wantMatch: true},
		{name: "multiple selectors - one doesn't match", selectors: []string{"env=production", "language=go"},
			frontmatter: fm(map[string]any{"env": "production", "language": "python"}), wantMatch: false},
		{name: "multiple selectors - one key missing (allowed)", selectors: []string{"env=production", "language=go"},
			frontmatter: fm(map[string]any{"env": "production"}), wantMatch: true},
		{name: "empty selectors - always match", selectors: []string{},
			frontmatter: fm(map[string]any{"env": "production"}), wantMatch: true},
		{name: "boolean value conversion - match", selectors: []string{"is_active=true"},
			frontmatter: fm(map[string]any{"is_active": true}), wantMatch: true},
		{name: "array selector - match", selectors: []string{},
			frontmatter: fm(map[string]any{"rule_name": "rule2"}), wantMatch: true,
			setupSelectors: setupRuleNames("rule1", "rule2", "rule3")},
		{name: "array selector - no match", selectors: []string{},
			frontmatter: fm(map[string]any{"rule_name": "rule4"}), wantMatch: false,
			setupSelectors: setupRuleNames("rule1", "rule2", "rule3")},
		{name: "array selector - key missing (allowed)", selectors: []string{},
			frontmatter: fm(map[string]any{"env": "prod"}), wantMatch: true,
			setupSelectors: setupRuleNames("rule1", "rule2")},
		{name: "mixed selectors - array and string both match", selectors: []string{"env=prod"},
			frontmatter: fm(map[string]any{"env": "prod", "rule_name": "rule1"}), wantMatch: true,
			setupSelectors: setupRuleNames("rule1", "rule2")},
		{name: "mixed selectors - string doesn't match", selectors: []string{"env=dev"},
			frontmatter: fm(map[string]any{"env": "prod", "rule_name": "rule1"}), wantMatch: false,
			setupSelectors: setupRuleNames("rule1", "rule2")},
		{name: "multiple array selectors - both match", selectors: []string{},
			frontmatter: fm(map[string]any{"rule_name": "rule1", "language": "go"}), wantMatch: true,
			setupSelectors: func(s Selectors) {
				setupRuleNames("rule1", "rule2")(s)
				s.SetValue("language", "go")
				s.SetValue("language", "python")
			}},
		{name: "multiple array selectors - one doesn't match", selectors: []string{},
			frontmatter: fm(map[string]any{"rule_name": "rule1", "language": "java"}), wantMatch: false,
			setupSelectors: func(s Selectors) {
				setupRuleNames("rule1", "rule2")(s)
				s.SetValue("language", "go")
				s.SetValue("language", "python")
			}},
		{name: "OR logic - same key multiple values matches", selectors: []string{"env=prod", "env=dev"},
			frontmatter: fm(map[string]any{"env": "dev"}), wantMatch: true},
		{name: "OR logic - same key multiple values no match", selectors: []string{"env=prod", "env=dev"},
			frontmatter: fm(map[string]any{"env": "staging"}), wantMatch: false},
		{name: "empty value selector - key exists in frontmatter (no match)", selectors: []string{"env="},
			frontmatter: fm(map[string]any{"env": "production"}), wantMatch: false},
		{name: "empty value selector - key missing in frontmatter (match)", selectors: []string{"env="},
			frontmatter: fm(map[string]any{"language": "go"}), wantMatch: true},
		// Array frontmatter values (YAML arrays like `languages: [nodejs]` are parsed as []interface{})
		{name: "array frontmatter value - single element match", selectors: []string{},
			frontmatter: fm(map[string]any{"languages": []interface{}{"nodejs"}}), wantMatch: true,
			setupSelectors: func(s Selectors) {
				s.SetValue("languages", "nodejs")
				s.SetValue("languages", "python")
			}},
		{name: "array frontmatter value - single element no match", selectors: []string{},
			frontmatter: fm(map[string]any{"languages": []interface{}{"java"}}), wantMatch: false,
			setupSelectors: func(s Selectors) {
				s.SetValue("languages", "nodejs")
				s.SetValue("languages", "python")
			}},
		{name: "array frontmatter value - multi element match", selectors: []string{},
			frontmatter: fm(map[string]any{"languages": []interface{}{"go", "python"}}), wantMatch: true,
			setupSelectors: func(s Selectors) {
				s.SetValue("languages", "nodejs")
				s.SetValue("languages", "python")
			}},
		{name: "array frontmatter value - multi element no match", selectors: []string{},
			frontmatter: fm(map[string]any{"languages": []interface{}{"java", "rust"}}), wantMatch: false,
			setupSelectors: func(s Selectors) {
				s.SetValue("languages", "nodejs")
				s.SetValue("languages", "python")
			}},
		{name: "array frontmatter value - with string selector", selectors: []string{"languages=nodejs"},
			frontmatter: fm(map[string]any{"languages": []interface{}{"nodejs"}}), wantMatch: true},
		// excludeUnmatched=true cases
		{name: "exclude by default - key missing", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"language": "go"}), excludeUnmatched: true, wantMatch: false},
		{name: "exclude by default - explicit match still included", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"env": "production"}), excludeUnmatched: true, wantMatch: true},
		{name: "exclude by default - no active selectors (early return)", selectors: []string{},
			frontmatter: fm(map[string]any{"language": "go"}), excludeUnmatched: true, wantMatch: true},
	}
}

func runMatchesIncludes(t *testing.T, tt matchesIncludesCase) {
	t.Helper()

	s := make(Selectors)

	for _, sel := range tt.selectors {
		if err := s.Set(sel); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	if tt.setupSelectors != nil {
		tt.setupSelectors(s)
	}

	gotMatch, gotReason := s.MatchesIncludes(tt.frontmatter, !tt.excludeUnmatched)
	if gotMatch != tt.wantMatch {
		t.Errorf("MatchesIncludes() = %v, want %v (reason: %s)", gotMatch, tt.wantMatch, gotReason)
	}
}

func TestSelectorMap_MatchesIncludes(t *testing.T) {
	t.Parallel()

	for _, tt := range matchesIncludesCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runMatchesIncludes(t, tt)
		})
	}
}

func TestSelectorMap_String(t *testing.T) {
	t.Parallel()

	s := make(Selectors)
	if err := s.Set("env=production"); err != nil {
		t.Fatal(err)
	}

	if err := s.Set("language=go"); err != nil {
		t.Fatal(err)
	}

	str := s.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

func checkMultipleSelectorsReason(t *testing.T, reason string) {
	t.Helper()

	if !strings.Contains(reason, "matched selectors:") {
		t.Errorf("Expected reason to contain 'matched selectors:', got %q", reason)
	}

	if !strings.Contains(reason, "env=production") || !strings.Contains(reason, "language=go") {
		t.Errorf("Expected reason to contain both selectors, got %q", reason)
	}
}

func checkArraySelectorNoMatchReason(t *testing.T, reason string) {
	t.Helper()

	if !strings.Contains(reason, "selectors did not match:") {
		t.Errorf("Expected reason to start with 'selectors did not match:', got %q", reason)
	}

	if !strings.Contains(reason, "rule_name=rule4") {
		t.Errorf("Expected reason to contain 'rule_name=rule4', got %q", reason)
	}

	if !strings.Contains(reason, "rule1") || !strings.Contains(reason, "rule2") || !strings.Contains(reason, "rule3") {
		t.Errorf("Expected reason to contain all expected values, got %q", reason)
	}
}

type matchesIncludesReasonCase struct {
	name             string
	selectors        []string
	setupSelectors   func(s Selectors)
	frontmatter      markdown.BaseFrontMatter
	excludeUnmatched bool // when true, unmatched rules are excluded; default false = include by default
	wantMatch        bool
	wantReason       string
	checkReason      func(t *testing.T, reason string)
}

func matchesIncludesReasonCases() []matchesIncludesReasonCase {
	return []matchesIncludesReasonCase{
		{name: "single selector - match", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"env": "production"}), wantMatch: true,
			wantReason: "matched selectors: env=production"},
		{name: "single selector - no match", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"env": "development"}), wantMatch: false,
			wantReason: "selectors did not match: env=development (expected env=production)"},
		{name: "single selector - key missing (allowed)", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"language": "go"}), wantMatch: true,
			wantReason: "no selectors specified (included by default)"},
		{name: "multiple selectors - all match", selectors: []string{"env=production", "language=go"},
			frontmatter: fm(map[string]any{"env": "production", "language": "go"}), wantMatch: true,
			checkReason: checkMultipleSelectorsReason},
		{name: "multiple selectors - one doesn't match", selectors: []string{"env=production", "language=go"},
			frontmatter: fm(map[string]any{"env": "production", "language": "python"}), wantMatch: false,
			wantReason: "selectors did not match: language=python (expected language=go)"},
		{name: "empty selectors", selectors: []string{},
			frontmatter: fm(map[string]any{"env": "production"}), wantMatch: true, wantReason: ""},
		{name: "array selector - match", selectors: []string{},
			frontmatter: fm(map[string]any{"rule_name": "rule2"}), wantMatch: true,
			wantReason:     "matched selectors: rule_name=rule2",
			setupSelectors: setupRuleNames("rule1", "rule2", "rule3")},
		{name: "array selector - no match", selectors: []string{},
			frontmatter: fm(map[string]any{"rule_name": "rule4"}), wantMatch: false,
			setupSelectors: setupRuleNames("rule1", "rule2", "rule3"),
			checkReason:    checkArraySelectorNoMatchReason},
		{name: "boolean value conversion", selectors: []string{"is_active=true"},
			frontmatter: fm(map[string]any{"is_active": true}), wantMatch: true,
			wantReason: "matched selectors: is_active=true"},
		{name: "exclude by default - key missing", selectors: []string{"env=production"},
			frontmatter: fm(map[string]any{"language": "go"}), excludeUnmatched: true,
			wantMatch: false, wantReason: "excluded by default (no matching selectors)"},
	}
}

func TestSelectorMap_MatchesIncludesReasons(t *testing.T) {
	t.Parallel()

	for _, tt := range matchesIncludesReasonCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := make(Selectors)

			for _, sel := range tt.selectors {
				if err := s.Set(sel); err != nil {
					t.Fatalf("Set() error = %v", err)
				}
			}

			if tt.setupSelectors != nil {
				tt.setupSelectors(s)
			}

			gotMatch, gotReason := s.MatchesIncludes(tt.frontmatter, !tt.excludeUnmatched)

			if gotMatch != tt.wantMatch {
				t.Errorf("MatchesIncludes() match = %v, want %v (reason: %s)", gotMatch, tt.wantMatch, gotReason)
			}

			if tt.checkReason != nil {
				tt.checkReason(t, gotReason)
			} else if gotReason != tt.wantReason {
				t.Errorf("MatchesIncludes() reason = %q, want %q", gotReason, tt.wantReason)
			}
		})
	}
}
