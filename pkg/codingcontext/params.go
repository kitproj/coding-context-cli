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
// Values can be quoted with double quotes, and quotes can be escaped.
// Examples:
//   - "key1=value1 key2=value2"
//   - `key1="value with spaces" key2=value2`
//   - `key1="value with \"escaped\" quotes"`
func ParseParams(s string) Params {
	params := make(Params)
	if s == "" {
		return params
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
			params[key] = ""
			break
		}

		// Parse the value
		var value strings.Builder
		if s[i] == '"' {
			// Double-quoted value
			i++ // skip opening quote
			for i < len(s) {
				if s[i] == '\\' && i+1 < len(s) && s[i+1] == '"' {
					value.WriteByte('"')
					i += 2
				} else if s[i] == '"' {
					i++ // skip closing quote
					break
				} else {
					value.WriteByte(s[i])
					i++
				}
			}
		} else {
			// Check if we're at the start of a new key-value pair (look for '=' ahead)
			// This handles cases like "key1= key2=value2"
			j := i
			for j < len(s) && s[j] != '=' && s[j] != ' ' && s[j] != '\t' {
				j++
			}
			isNewKeyValuePair := j < len(s) && s[j] == '=' && j > i

			if isNewKeyValuePair {
				// Empty value, next token is a new key-value pair
				params[key] = ""
				continue
			}

			// Unquoted value (until space or end of string)
			valueStart := i
			for i < len(s) && s[i] != ' ' && s[i] != '\t' {
				i++
			}
			value.WriteString(s[valueStart:i])
		}

		params[key] = value.String()
	}

	return params
}
