package main

import (
	"fmt"
	"slices"
	"strings"
)

// selectorMap stores selector key-value pairs where values are always string slices
// Multiple values for the same key use OR logic (match any value in the slice)
type selectorMap map[string][]string

func (s *selectorMap) String() string {
	if *s == nil {
		return "{}"
	}
	var parts []string
	for k, v := range *s {
		if len(v) == 1 {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v[0]))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

func (s *selectorMap) Set(value string) error {
	// Parse key=value format with trimming
	kv := strings.SplitN(value, "=", 2)
	if len(kv) != 2 {
		return fmt.Errorf("invalid selector format: %s", value)
	}
	if *s == nil {
		*s = make(selectorMap)
	}
	key := strings.TrimSpace(kv[0])
	newValue := strings.TrimSpace(kv[1])

	// If key already exists, append to slice for OR logic
	if existingValues, exists := (*s)[key]; exists {
		// Check if new value is already in the slice
		if !slices.Contains(existingValues, newValue) {
			(*s)[key] = append(existingValues, newValue)
		}
	} else {
		// Key doesn't exist, store as single-element slice
		(*s)[key] = []string{newValue}
	}
	return nil
}

// matchesIncludes returns true if the frontmatter matches all include selectors
// If a key doesn't exist in frontmatter, it's allowed
// Multiple values for the same key use OR logic (matches if frontmatter value is in the slice)
func (includes *selectorMap) matchesIncludes(frontmatter frontMatter) bool {
	for key, values := range *includes {
		fmValue, exists := frontmatter[key]
		if !exists {
			// If key doesn't exist in frontmatter, allow it
			continue
		}

		// Check if frontmatter value matches any element in the slice (OR logic)
		fmStr := fmt.Sprint(fmValue)
		if !slices.Contains(values, fmStr) {
			return false
		}
	}
	return true
}
