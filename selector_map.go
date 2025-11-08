package main

import (
	"fmt"
	"strings"
)

// selectorMap stores multiple values per key for inclusive selection
// When the same key is specified multiple times (e.g., -s language=Go -s language=Typescript),
// a rule matches if its frontmatter value matches ANY of the specified values (OR logic).
// Different keys use AND logic (e.g., -s language=Go -s stage=implementation).
type selectorMap map[string][]string

func (s *selectorMap) String() string {
	if *s == nil {
		return ""
	}
	var parts []string
	for key, values := range *s {
		for _, value := range values {
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return strings.Join(parts, " ")
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
	val := strings.TrimSpace(kv[1])

	// Append value to the list for this key (supports multiple values per key)
	(*s)[key] = append((*s)[key], val)
	return nil
}

// matchesIncludes returns true if the frontmatter matches the include selectors
// - For each key in selectors, if that key exists in frontmatter, it must match at least one of the values (OR logic within same key)
// - All keys in selectors must be satisfied (AND logic across different keys)
// - If a selector key doesn't exist in frontmatter, it's allowed (matches)
// - Frontmatter values can be scalars or arrays. For arrays, any element matching any selector value is a match.
func (includes *selectorMap) matchesIncludes(frontmatter frontMatter) bool {
	for key, values := range *includes {
		fmValue, exists := frontmatter[key]

		// If key doesn't exist in frontmatter, allow it
		if !exists {
			continue
		}

		// Check if frontmatter value matches ANY of the selector values
		matched := false

		// Handle both scalar and array values in frontmatter
		switch v := fmValue.(type) {
		case []any:
			// Frontmatter value is an array - check if any element matches any selector value
			for _, elem := range v {
				elemStr := fmt.Sprint(elem)
				for _, value := range values {
					if elemStr == value {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
		default:
			// Frontmatter value is a scalar - match as string
			fmValueStr := fmt.Sprint(fmValue)
			for _, value := range values {
				if fmValueStr == value {
					matched = true
					break
				}
			}
		}

		// If none of the values matched, this selector key fails
		if !matched {
			return false
		}
	}
	return true
}
