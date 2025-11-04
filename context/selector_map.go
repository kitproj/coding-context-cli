package context

import (
	"fmt"
	"strings"
)

// SelectorMap reuses ParamMap for parsing key=value pairs
type SelectorMap ParamMap

func (s *SelectorMap) String() string {
	return (*ParamMap)(s).String()
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
	// Trim spaces from both key and value for selectors
	(*s)[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	return nil
}

// MatchesIncludes returns true if the frontmatter matches all include selectors
// If a key doesn't exist in frontmatter, it's allowed
func (includes *SelectorMap) MatchesIncludes(frontmatter map[string]string) bool {
	for key, value := range *includes {
		fmValue, exists := frontmatter[key]
		// If key exists, it must match the value
		if exists && fmValue != value {
			return false
		}
		// If key doesn't exist, allow it
	}
	return true
}

// MatchesExcludes returns true if the frontmatter doesn't match any exclude selectors
// If a key doesn't exist in frontmatter, it's allowed
func (excludes *SelectorMap) MatchesExcludes(frontmatter map[string]string) bool {
	for key, value := range *excludes {
		fmValue, exists := frontmatter[key]
		// If key exists and matches the value, exclude it
		if exists && fmValue == value {
			return false
		}
		// If key doesn't exist, allow it
	}
	return true
}
