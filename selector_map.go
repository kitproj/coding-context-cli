package main

import (
	"fmt"
	"strings"
)

type selector struct {
	key   string
	value string
}

type selectorMap []selector

func (s *selectorMap) String() string {
	return fmt.Sprint(*s)
}

func (s *selectorMap) Set(value string) error {
	// Parse key=value format
	if strings.Contains(value, "=") {
		kv := strings.SplitN(value, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid selector format: %s", value)
		}
		*s = append(*s, selector{
			key:   strings.TrimSpace(kv[0]),
			value: strings.TrimSpace(kv[1]),
		})
		return nil
	}

	return fmt.Errorf("invalid selector format: %s (must contain =)", value)
}

// matchesIncludes returns true if the frontmatter matches all include selectors
// If a key doesn't exist in frontmatter, it's allowed
func (includes *selectorMap) matchesIncludes(frontmatter map[string]string) bool {
	for _, sel := range *includes {
		fmValue, exists := frontmatter[sel.key]
		// If key exists, it must match the value
		if exists && fmValue != sel.value {
			return false
		}
		// If key doesn't exist, allow it
	}
	return true
}

// matchesExcludes returns true if the frontmatter doesn't match any exclude selectors
// If a key doesn't exist in frontmatter, it's allowed
func (excludes *selectorMap) matchesExcludes(frontmatter map[string]string) bool {
	for _, sel := range *excludes {
		fmValue, exists := frontmatter[sel.key]
		// If key exists and matches the value, exclude it
		if exists && fmValue == sel.value {
			return false
		}
		// If key doesn't exist, allow it
	}
	return true
}
