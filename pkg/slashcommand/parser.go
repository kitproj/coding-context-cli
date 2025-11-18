package slashcommand

import (
	"fmt"
	"strings"
)

// ParseSlashCommand parses a slash command string and extracts the task name and parameters.
// The expected format is: /task-name param1="value1" param2="value2"
//
// Examples:
//   - "/fix-bug issue_number=\"123\""
//   - "/implement-feature feature_name=\"User Login\" priority=\"high\""
//   - "/code-review"
//
// Returns:
//   - taskName: the task name (without the leading slash)
//   - params: a map of parameter key-value pairs
//   - err: an error if the command format is invalid
func ParseSlashCommand(command string) (taskName string, params map[string]string, err error) {
	command = strings.TrimSpace(command)

	// Check if command starts with '/'
	if !strings.HasPrefix(command, "/") {
		return "", nil, fmt.Errorf("slash command must start with '/'")
	}

	// Remove leading slash
	command = command[1:]

	if command == "" {
		return "", nil, fmt.Errorf("slash command cannot be empty")
	}

	// Split into task name and parameters
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("slash command cannot be empty")
	}

	taskName = parts[0]
	params = make(map[string]string)

	// If there are no parameters, return early
	if len(parts) == 1 {
		return taskName, params, nil
	}

	// Parse parameters from the remaining parts
	// We need to handle quoted values that may contain spaces
	paramString := strings.TrimSpace(command[len(taskName):])

	if paramString == "" {
		return taskName, params, nil
	}

	// Parse key="value" pairs, handling quoted values with spaces
	pairs, err := parseKeyValuePairs(paramString)
	if err != nil {
		return "", nil, err
	}

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return "", nil, fmt.Errorf("invalid parameter format: %s", pair)
		}

		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		// Remove quotes from value
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		} else {
			return "", nil, fmt.Errorf("parameter value must be quoted: %s", pair)
		}

		params[key] = val
	}

	return taskName, params, nil
}

// parseKeyValuePairs splits a string into key="value" pairs, respecting quoted values
func parseKeyValuePairs(s string) ([]string, error) {
	var pairs []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			inQuotes = !inQuotes
			current.WriteByte(ch)
		} else if (ch == ' ' || ch == '\t') && !inQuotes {
			// End of a pair
			if current.Len() > 0 {
				pair := strings.TrimSpace(current.String())
				if pair != "" {
					pairs = append(pairs, pair)
				}
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	// Add the last pair
	if current.Len() > 0 {
		pair := strings.TrimSpace(current.String())
		if pair != "" {
			pairs = append(pairs, pair)
		}
	}

	// Check for unclosed quotes
	if inQuotes {
		return nil, fmt.Errorf("unclosed quote in parameters")
	}

	return pairs, nil
}
