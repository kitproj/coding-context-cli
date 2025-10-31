package main

import (
	"fmt"
	"strings"
)

type selectorType int

const (
	selectorEquals selectorType = iota
	selectorNotEquals
)

type selector struct {
	key   string
	value string
	op    selectorType
}

type selectorMap []selector

func (s *selectorMap) String() string {
	return fmt.Sprint(*s)
}

func (s *selectorMap) Set(value string) error {
	// Check for != operator first (it's longer)
	if strings.Contains(value, "!=") {
		kv := strings.SplitN(value, "!=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid selector format: %s", value)
		}
		*s = append(*s, selector{
			key:   strings.TrimSpace(kv[0]),
			value: strings.TrimSpace(kv[1]),
			op:    selectorNotEquals,
		})
		return nil
	}

	// Check for = operator
	if strings.Contains(value, "=") {
		kv := strings.SplitN(value, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid selector format: %s", value)
		}
		*s = append(*s, selector{
			key:   strings.TrimSpace(kv[0]),
			value: strings.TrimSpace(kv[1]),
			op:    selectorEquals,
		})
		return nil
	}

	return fmt.Errorf("invalid selector format: %s (must contain = or !=)", value)
}

// matches returns true if the frontmatter matches all selectors
func (s *selectorMap) matches(frontmatter map[string]string) bool {
	for _, sel := range *s {
		fmValue, exists := frontmatter[sel.key]
		
		switch sel.op {
		case selectorEquals:
			// For equals, the key must exist and match the value
			if !exists || fmValue != sel.value {
				return false
			}
		case selectorNotEquals:
			// For not equals, if key exists it must not match the value
			if exists && fmValue == sel.value {
				return false
			}
		}
	}
	return true
}
