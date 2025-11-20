package codingcontext

import (
	"fmt"
	"strings"
)

// Params is a map of parameter key-value pairs for template substitution
type Params map[string]string

// String implements the fmt.Stringer interface for Params
func (p *Params) String() string {
	return fmt.Sprint(*p)
}

// Set implements the flag.Value interface for Params
func (p *Params) Set(value string) error {
	kv := strings.SplitN(value, "=", 2)
	if len(kv) != 2 {
		return fmt.Errorf("invalid parameter format: %s", value)
	}
	if *p == nil {
		*p = make(map[string]string)
	}
	(*p)[kv[0]] = kv[1]
	return nil
}

// ParseParams parses a string in the format 'key1="value1",key2="value2"' and returns a Params map.
// Values must be quoted with double quotes. If the string is not in the correct format, it returns an error.
func ParseParams(value string) (Params, error) {
	p := make(Params)
	if value == "" {
		return p, nil
	}

	// Parse comma-separated key="value" pairs
	// We need to handle quoted values properly
	pairs := parseKeyValuePairs(value)
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s", pair)
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		// Remove quotes from value
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		} else {
			return nil, fmt.Errorf("value must be quoted: %s", pair)
		}

		p[key] = val
	}
	return p, nil
}

// parseKeyValuePairs splits a string by commas, respecting quoted values
func parseKeyValuePairs(s string) []string {
	var pairs []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if ch == '"' {
			inQuotes = !inQuotes
			current.WriteByte(ch)
		} else if ch == ',' && !inQuotes {
			// End of a pair
			if current.Len() > 0 {
				pairs = append(pairs, strings.TrimSpace(current.String()))
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	// Add the last pair
	if current.Len() > 0 {
		pairs = append(pairs, strings.TrimSpace(current.String()))
	}

	return pairs
}
