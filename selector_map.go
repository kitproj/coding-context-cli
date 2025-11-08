package main

import (
	"fmt"
	"strings"
)

// selectorOperator represents the type of selector operation
type selectorOperator string

const (
	selectorEquals      selectorOperator = "="  // exact match
	selectorIncludes    selectorOperator = ":=" // value in array or equals scalar
	selectorNotEquals   selectorOperator = "!=" // not equal
	selectorNotIncludes selectorOperator = "!:" // not in array
)

// selector represents a single selector with an operator and value
type selector struct {
	operator selectorOperator
	value    string
}

// selectorMap stores selectors by key
type selectorMap map[string]selector

func (s *selectorMap) String() string {
	return fmt.Sprint(*s)
}

func (s *selectorMap) Set(value string) error {
	// Parse selector format: key<operator>value
	// Supported operators: =, :=, !=, !:
	// Check for operators in order of length (longest first to avoid mismatching)
	var key string
	var op selectorOperator
	var val string

	if idx := strings.Index(value, ":="); idx != -1 {
		key = strings.TrimSpace(value[:idx])
		val = strings.TrimSpace(value[idx+2:])
		op = selectorIncludes
	} else if idx := strings.Index(value, "!="); idx != -1 {
		key = strings.TrimSpace(value[:idx])
		val = strings.TrimSpace(value[idx+2:])
		op = selectorNotEquals
	} else if idx := strings.Index(value, "!:"); idx != -1 {
		key = strings.TrimSpace(value[:idx])
		val = strings.TrimSpace(value[idx+2:])
		op = selectorNotIncludes
	} else if idx := strings.Index(value, "="); idx != -1 {
		key = strings.TrimSpace(value[:idx])
		val = strings.TrimSpace(value[idx+1:])
		op = selectorEquals
	} else {
		return fmt.Errorf("invalid selector format: %s (expected format: key<op>value where <op> is =, :=, !=, or !:)", value)
	}

	if key == "" {
		return fmt.Errorf("invalid selector format: %s (key cannot be empty)", value)
	}

	if *s == nil {
		*s = make(selectorMap)
	}
	(*s)[key] = selector{operator: op, value: val}
	return nil
}

// matchesIncludes returns true if the frontmatter matches all selectors
// If a key doesn't exist in frontmatter, it's allowed for positive operators (=, :=)
// but disallowed for negative operators (!=, !:)
func (includes *selectorMap) matchesIncludes(frontmatter frontMatter) bool {
	for key, sel := range *includes {
		fmValue, exists := frontmatter[key]

		switch sel.operator {
		case selectorEquals:
			// Exact match required if key exists
			if exists && fmt.Sprint(fmValue) != sel.value {
				return false
			}
			// If key doesn't exist, allow it

		case selectorIncludes:
			// Check if value is in array or equals scalar
			if exists {
				if !valueIncludes(fmValue, sel.value) {
					return false
				}
			}
			// If key doesn't exist, allow it

		case selectorNotEquals:
			// Value must not equal if key exists
			if exists && fmt.Sprint(fmValue) == sel.value {
				return false
			}
			// If key doesn't exist, it's not equal, so it matches

		case selectorNotIncludes:
			// Value must not be in array
			if exists && valueIncludes(fmValue, sel.value) {
				return false
			}
			// If key doesn't exist, value is not included, so it matches
		}
	}
	return true
}

// valueIncludes checks if a value is included in a frontmatter value
// Returns true if:
// - fmValue is a scalar and equals the target value
// - fmValue is an array and contains the target value
func valueIncludes(fmValue any, targetValue string) bool {
	// Check if fmValue is an array
	switch v := fmValue.(type) {
	case []any:
		// Check if any element in the array matches
		for _, item := range v {
			if fmt.Sprint(item) == targetValue {
				return true
			}
		}
		return false
	default:
		// Scalar value - check for exact match
		return fmt.Sprint(fmValue) == targetValue
	}
}
