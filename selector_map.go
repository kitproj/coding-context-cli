package main

import (
	"fmt"
	"strings"
)

// selectorMap reuses paramMap for parsing key=value pairs
type selectorMap paramMap

func (s *selectorMap) String() string {
	return (*paramMap)(s).String()
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
	// Trim spaces from both key and value for selectors
	(*s)[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	return nil
}

// matchesIncludes returns true if the frontmatter matches all include selectors
// If a key doesn't exist in frontmatter, it's allowed
// params contains both user-provided parameters and built-in values like task_name
func (includes *selectorMap) matchesIncludes(frontmatter map[string]string, params paramMap) bool {
	for key, value := range *includes {
		fmValue, exists := frontmatter[key]
		// If key exists, it must match the value
		if exists && fmValue != value {
			return false
		}
		// If key doesn't exist, allow it
	}
	
	// Check params for automatic filtering (e.g., task_name)
	// If frontmatter has a key that exists in params, it must match
	for key, paramValue := range params {
		fmValue, exists := frontmatter[key]
		// If the key exists in frontmatter, it must match the param value
		if exists && fmValue != paramValue {
			return false
		}
		// If key doesn't exist in frontmatter, allow it
	}
	
	return true
}

// matchesExcludes returns true if the frontmatter doesn't match any exclude selectors
// If a key doesn't exist in frontmatter, it's allowed
// params is not used for excludes - only explicit exclude selectors matter
func (excludes *selectorMap) matchesExcludes(frontmatter map[string]string, params paramMap) bool {
	for key, value := range *excludes {
		fmValue, exists := frontmatter[key]
		// If key exists and matches the value, exclude it
		if exists && fmValue == value {
			return false
		}
		// If key doesn't exist, allow it
	}
	
	// params are not used for excludes - they only affect includes
	// This allows task_name and other params to act as automatic include filters only
	
	return true
}
