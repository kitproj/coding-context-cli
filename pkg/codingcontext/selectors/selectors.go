// Package selectors provides selector parsing and matching for rule/skill frontmatter.
package selectors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
)

// ErrInvalidSelectorFormat is returned when a selector string is not in key=value format.
var ErrInvalidSelectorFormat = errors.New("invalid selector format")

// Selectors stores selector key-value pairs where values are stored in inner maps
// Multiple values for the same key use OR logic (match any value in the inner map)
// Each value can be represented exactly once per key.
type Selectors map[string]map[string]bool

// String implements the fmt.Stringer interface for Selectors.
func (s *Selectors) String() string {
	if *s == nil {
		return "{}"
	}

	var parts []string

	for k, v := range *s {
		values := make([]string, 0, len(v))
		for val := range v {
			values = append(values, val)
		}

		if len(values) == 1 {
			parts = append(parts, fmt.Sprintf("%s=%s", k, values[0]))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%v", k, values))
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

// Set implements the flag.Value interface for Selectors.
func (s *Selectors) Set(value string) error {
	const keyValueParts = 2
	// Parse key=value format with trimming
	kv := strings.SplitN(value, "=", keyValueParts)
	if len(kv) != keyValueParts {
		return fmt.Errorf("%w: %s", ErrInvalidSelectorFormat, value)
	}

	if *s == nil {
		*s = make(Selectors)
	}

	key := strings.TrimSpace(kv[0])
	newValue := strings.TrimSpace(kv[1])

	// If value is empty, set the key to an empty map only if it's currently unset
	if newValue == "" {
		if _, exists := (*s)[key]; !exists {
			(*s)[key] = make(map[string]bool)
		}

		return nil
	}

	s.SetValue(key, newValue)

	return nil
}

// SetValue sets a value in the inner map for the given key.
// If the key doesn't exist, it creates a new inner map.
// Each value can be represented exactly once per key.
func (s *Selectors) SetValue(key, value string) {
	if *s == nil {
		*s = make(Selectors)
	}

	if (*s)[key] == nil {
		(*s)[key] = make(map[string]bool)
	}

	(*s)[key][value] = true
}

// GetValue returns true if the given value exists in the inner map for the given key.
// Returns false if the key doesn't exist or the value is not present.
func (s *Selectors) GetValue(key, value string) bool {
	if *s == nil {
		return false
	}

	innerMap, exists := (*s)[key]
	if !exists {
		return false
	}

	return innerMap[value]
}

// MatchesIncludes returns whether the frontmatter matches all include selectors,
// along with a human-readable reason explaining the result.
// If a key doesn't exist in frontmatter, it's allowed when includeByDefault is true (the default).
// When includeByDefault is false, rules/skills that produce no explicit selector match are excluded.
// Multiple values for the same key use OR logic (matches if frontmatter value is in the inner map).
// This enables combining CLI selectors (-s flag) with task frontmatter selectors:
// both are added to the same Selectors map, creating an OR condition for rules to match.
//
// Returns:
//   - bool: true if all selectors match, false otherwise
//   - string: reason explaining why (matched selectors or mismatch details)
func (s *Selectors) MatchesIncludes(frontmatter markdown.BaseFrontMatter, includeByDefault bool) (bool, string) {
	if len(*s) == 0 {
		return true, ""
	}

	var (
		matchedSelectors []string
		noMatchReasons   []string
	)

	for key, values := range *s {
		fmValue, exists := frontmatter.Content[key]
		if !exists {
			// If key doesn't exist in frontmatter, allow it
			continue
		}

		fmStr := fmt.Sprint(fmValue)
		if values[fmStr] {
			// This selector matched
			matchedSelectors = append(matchedSelectors, fmt.Sprintf("%s=%s", key, fmStr))
		} else {
			// This selector didn't match
			var expectedValues []string
			for val := range values {
				expectedValues = append(expectedValues, val)
			}

			if len(expectedValues) == 1 {
				noMatchReasons = append(noMatchReasons, fmt.Sprintf("%s=%s (expected %s=%s)", key, fmStr, key, expectedValues[0]))
			} else {
				noMatchReasons = append(noMatchReasons,
					fmt.Sprintf("%s=%s (expected %s in [%s])", key, fmStr, key, strings.Join(expectedValues, ", ")))
			}
		}
	}

	// If any selector didn't match, return false with the mismatch reasons
	if len(noMatchReasons) > 0 {
		return false, "selectors did not match: " + strings.Join(noMatchReasons, ", ")
	}

	// All selectors matched
	if len(matchedSelectors) > 0 {
		return true, "matched selectors: " + strings.Join(matchedSelectors, ", ")
	}

	// No explicit selector match
	if !includeByDefault {
		return false, "excluded by default (no matching selectors)"
	}

	return true, "no selectors specified (included by default)"
}
