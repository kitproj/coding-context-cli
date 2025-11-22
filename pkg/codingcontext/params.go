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

// ParseParams parses a string containing key=value pairs separated by spaces.
// Values must be quoted with double quotes, and quotes can be escaped.
// Unquoted values are treated as an error.
// Examples:
//   - `key1="value1" key2="value2"`
//   - `key1="value with spaces" key2="value2"`
//   - `key1="value with \"escaped\" quotes"`
func ParseParams(s string) (Params, error) {
	params := make(Params)
	if s == "" {
		return params, nil
	}

	s = strings.TrimSpace(s)
	var i int
	for i < len(s) {
		// Skip whitespace
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) {
			break
		}

		// Find the key (until '=')
		keyStart := i
		for i < len(s) && s[i] != '=' {
			i++
		}
		if i >= len(s) {
			break
		}
		key := strings.TrimSpace(s[keyStart:i])
		if key == "" {
			i++
			continue
		}

		// Skip '='
		i++

		// Skip whitespace after '='
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) {
			return nil, fmt.Errorf("missing quoted value for key %q", key)
		}

		// Values must be quoted
		if s[i] != '"' {
			return nil, fmt.Errorf("unquoted value for key %q: values must be double-quoted", key)
		}

		// Parse the double-quoted value
		var value strings.Builder
		i++ // skip opening quote
		quoted := false
		for i < len(s) {
			if s[i] == '\\' && i+1 < len(s) && s[i+1] == '"' {
				value.WriteByte('"')
				i += 2
			} else if s[i] == '"' {
				i++ // skip closing quote
				quoted = true
				break
			} else {
				value.WriteByte(s[i])
				i++
			}
		}

		if !quoted {
			return nil, fmt.Errorf("unclosed quote for key %q", key)
		}

		params[key] = value.String()
	}

	return params, nil
}
