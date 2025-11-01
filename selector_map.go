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
func (includes *selectorMap) matchesIncludes(frontmatter map[string]string) bool {
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

// matchesExcludes returns true if the frontmatter doesn't match any exclude selectors
// If a key doesn't exist in frontmatter, it's allowed
func (excludes *selectorMap) matchesExcludes(frontmatter map[string]string) bool {
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

// explainIncludes returns whether the frontmatter matches include selectors and an explanation
func (includes *selectorMap) explainIncludes(frontmatter map[string]string) (bool, string) {
	if len(*includes) == 0 {
		return true, "no include selectors specified"
	}
	
	var matched []string
	var missing []string
	var mismatched []string
	
	for key, value := range *includes {
		fmValue, exists := frontmatter[key]
		if !exists {
			missing = append(missing, fmt.Sprintf("%s (key not in frontmatter)", key))
		} else if fmValue == value {
			matched = append(matched, fmt.Sprintf("%s=%s", key, value))
		} else {
			mismatched = append(mismatched, fmt.Sprintf("%s=%s (has %s=%s)", key, value, key, fmValue))
		}
	}
	
	// If any mismatched, file is excluded
	if len(mismatched) > 0 {
		return false, fmt.Sprintf("does not match include selector(s): %s", strings.Join(mismatched, ", "))
	}
	
	// Build explanation
	var parts []string
	if len(matched) > 0 {
		parts = append(parts, fmt.Sprintf("matches %s", strings.Join(matched, ", ")))
	}
	if len(missing) > 0 {
		parts = append(parts, fmt.Sprintf("allows missing %s", strings.Join(missing, ", ")))
	}
	
	return true, strings.Join(parts, "; ")
}

// explainExcludes returns whether the frontmatter passes exclude selectors and an explanation
func (excludes *selectorMap) explainExcludes(frontmatter map[string]string) (bool, string) {
	if len(*excludes) == 0 {
		return true, "no exclude selectors specified"
	}
	
	var excluded []string
	var notMatched []string
	var missing []string
	
	for key, value := range *excludes {
		fmValue, exists := frontmatter[key]
		if !exists {
			missing = append(missing, fmt.Sprintf("%s (key not in frontmatter)", key))
		} else if fmValue == value {
			excluded = append(excluded, fmt.Sprintf("%s=%s", key, value))
		} else {
			notMatched = append(notMatched, fmt.Sprintf("%s!=%s (has %s=%s)", key, value, key, fmValue))
		}
	}
	
	// If any matched exclude selector, file is excluded
	if len(excluded) > 0 {
		return false, fmt.Sprintf("matches exclude selector(s): %s", strings.Join(excluded, ", "))
	}
	
	// Build explanation
	var parts []string
	if len(notMatched) > 0 {
		parts = append(parts, fmt.Sprintf("does not match exclude %s", strings.Join(notMatched, ", ")))
	}
	if len(missing) > 0 {
		parts = append(parts, fmt.Sprintf("allows missing %s", strings.Join(missing, ", ")))
	}
	
	return true, strings.Join(parts, "; ")
}
