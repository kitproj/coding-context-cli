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
// builtins is a map of built-in filter values (e.g., task_name)
func (includes *selectorMap) matchesIncludes(frontmatter map[string]string, builtins map[string]string) bool {
	for key, value := range *includes {
		fmValue, exists := frontmatter[key]
		// If key exists, it must match the value
		if exists && fmValue != value {
			return false
		}
		// If key doesn't exist, allow it
	}
	
	// Check built-in filters (e.g., task_name)
	// Built-in filters are automatically applied based on the context
	for key, builtinValue := range builtins {
		fmValue, exists := frontmatter[key]
		// If the built-in key exists in frontmatter, it must match the built-in value
		if exists && fmValue != builtinValue {
			return false
		}
		// If key doesn't exist in frontmatter, allow it
	}
	
	return true
}

// matchesExcludes returns true if the frontmatter doesn't match any exclude selectors
// If a key doesn't exist in frontmatter, it's allowed
// builtins is a map of built-in filter values (e.g., task_name) - not used for excludes
func (excludes *selectorMap) matchesExcludes(frontmatter map[string]string, builtins map[string]string) bool {
	for key, value := range *excludes {
		fmValue, exists := frontmatter[key]
		// If key exists and matches the value, exclude it
		if exists && fmValue == value {
			return false
		}
		// If key doesn't exist, allow it
	}
	
	// Built-in filters do not affect excludes - they only affect includes
	// This allows built-in filters to act as automatic include filters only
	
	return true
}
