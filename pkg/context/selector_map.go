package context

import (
	"fmt"
	"strings"
)

// SelectorMap stores selector key-value pairs where values are stored in inner maps
// Multiple values for the same key use OR logic (match any value in the inner map)
// Each value can be represented exactly once per key
type SelectorMap map[string]map[string]bool

func (s *SelectorMap) String() string {
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

func (s *SelectorMap) Set(value string) error {
	// Parse key=value format with trimming
	kv := strings.SplitN(value, "=", 2)
	if len(kv) != 2 {
		return fmt.Errorf("invalid selector format: %s", value)
	}
	if *s == nil {
		*s = make(SelectorMap)
	}
	key := strings.TrimSpace(kv[0])
	newValue := strings.TrimSpace(kv[1])

	// If value is empty, set the key to an empty map only if it's currently unset
	if newValue == "" {
		if _, exists := (*s)[key]; !exists {
			(*s)[key] = make(map[string]bool)
		}
		return nil
	} else {
		s.SetValue(key, newValue)
	}

	return nil
}

// SetValue sets a value in the inner map for the given key.
// If the key doesn't exist, it creates a new inner map.
// Each value can be represented exactly once per key.
func (s *SelectorMap) SetValue(key, value string) {
	if *s == nil {
		*s = make(SelectorMap)
	}
	if (*s)[key] == nil {
		(*s)[key] = make(map[string]bool)
	}
	(*s)[key][value] = true
}

// GetValue returns true if the given value exists in the inner map for the given key.
// Returns false if the key doesn't exist or the value is not present.
func (s *SelectorMap) GetValue(key, value string) bool {
	if *s == nil {
		return false
	}
	innerMap, exists := (*s)[key]
	if !exists {
		return false
	}
	return innerMap[value]
}

// MatchesIncludes returns true if the frontmatter matches all include selectors
// If a key doesn't exist in frontmatter, it's allowed
// Multiple values for the same key use OR logic (matches if frontmatter value is in the inner map)
func (includes *SelectorMap) MatchesIncludes(frontmatter FrontMatter) bool {
	for key, values := range *includes {
		fmValue, exists := frontmatter[key]
		if !exists {
			// If key doesn't exist in frontmatter, allow it
			continue
		}

		// Check if frontmatter value matches any element in the inner map (OR logic)
		fmStr := fmt.Sprint(fmValue)
		if !values[fmStr] {
			return false
		}
	}
	return true
}
