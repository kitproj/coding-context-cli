package codingcontext

import (
	"fmt"
	"strings"
)

// parseSlashCommand parses a slash command string and extracts the task name and parameters.
// It searches for a slash command anywhere in the input string, not just at the beginning.
// The expected format is: /task-name arg1 "arg 2" arg3 key="value"
//
// The function will find the slash command even if it's embedded in other text. For example:
//   - "Please /fix-bug 123 today" -> taskName: "fix-bug", params: {"ARGUMENTS": "123 today", "1": "123", "2": "today"}, found: true
//   - "Some text /code-review" -> taskName: "code-review", params: {}, found: true
//
// Arguments are parsed like Bash:
//   - Quoted arguments can contain spaces
//   - Both single and double quotes are supported
//   - Quotes are removed from the parsed arguments
//   - Arguments are extracted until end of line
//
// Named parameters:
//   - Named parameters use key="value" format (double quotes required)
//   - Named parameters can be mixed with positional arguments
//   - Named parameters do not count toward positional numbering
//
// Examples:
//   - "/fix-bug 123" -> taskName: "fix-bug", params: {"ARGUMENTS": "123", "1": "123"}, found: true
//   - "/code-review \"PR #42\" high" -> taskName: "code-review", params: {"ARGUMENTS": "\"PR #42\" high", "1": "PR #42", "2": "high"}, found: true
//   - "/fix-bug issue=\"PROJ-123\"" -> taskName: "fix-bug", params: {"ARGUMENTS": "issue=\"PROJ-123\"", "issue": "PROJ-123"}, found: true
//   - "/task arg1 key=\"val\" arg2" -> taskName: "task", params: {"ARGUMENTS": "arg1 key=\"val\" arg2", "1": "arg1", "2": "arg2", "key": "val"}, found: true
//   - "no command here" -> taskName: "", params: nil, found: false
//
// Returns:
//   - taskName: the task name (without the leading slash)
//   - params: a map containing:
//   - "ARGUMENTS": the full argument string (with quotes preserved)
//   - "1", "2", "3", etc.: positional arguments (with quotes removed)
//   - "key": named parameter value (with quotes removed)
//   - found: true if a slash command was found, false otherwise
//   - err: an error if the command format is invalid (e.g., unclosed quotes)
func parseSlashCommand(command string) (taskName string, params map[string]string, found bool, err error) {
	// Find the slash command anywhere in the string
	slashIdx := strings.Index(command, "/")
	if slashIdx == -1 {
		return "", nil, false, nil
	}

	// Extract from the slash onwards
	command = command[slashIdx+1:]

	if command == "" {
		return "", nil, false, nil
	}

	// Find the task name (first word after the slash)
	// Task name ends at first whitespace or newline
	endIdx := strings.IndexAny(command, " \t\n\r")
	if endIdx == -1 {
		// No arguments, just the task name (rest of the string)
		return command, make(map[string]string), true, nil
	}

	taskName = command[:endIdx]

	// Extract arguments until end of line
	restOfString := command[endIdx:]
	newlineIdx := strings.IndexAny(restOfString, "\n\r")
	var argsString string
	if newlineIdx == -1 {
		// No newline, use everything
		argsString = strings.TrimSpace(restOfString)
	} else {
		// Only use up to the newline
		argsString = strings.TrimSpace(restOfString[:newlineIdx])
	}

	params = make(map[string]string)

	// Store the full argument string (with quotes preserved)
	if argsString != "" {
		params["ARGUMENTS"] = argsString
	}

	// If there are no arguments, return early
	if argsString == "" {
		return taskName, params, true, nil
	}

	// Parse arguments using bash-like parsing, handling both positional and named parameters
	args, namedParams, err := parseBashArgsWithNamed(argsString)
	if err != nil {
		return "", nil, false, err
	}

	// Add positional arguments as $1, $2, $3, etc.
	for i, arg := range args {
		params[fmt.Sprintf("%d", i+1)] = arg
	}

	// Add named parameters
	for key, value := range namedParams {
		params[key] = value
	}

	return taskName, params, true, nil
}

// parseBashArgsWithNamed parses a string into positional arguments and named parameters.
// Named parameters have the format key="value" (double quotes required).
// Returns positional arguments, named parameters, and any error.
func parseBashArgsWithNamed(s string) ([]string, map[string]string, error) {
	var positionalArgs []string
	namedParams := make(map[string]string)

	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	escaped := false
	justClosedQuotes := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' && inQuotes && quoteChar == '"' {
			// Only recognize escape in double quotes
			escaped = true
			continue
		}

		if (ch == '"' || ch == '\'') && !inQuotes {
			// Start of quoted string
			inQuotes = true
			quoteChar = ch
			justClosedQuotes = false
		} else if ch == quoteChar && inQuotes {
			// End of quoted string - mark that we just closed quotes
			inQuotes = false
			quoteChar = 0
			justClosedQuotes = true
		} else if (ch == ' ' || ch == '\t') && !inQuotes {
			// Whitespace outside quotes - end of argument
			if current.Len() > 0 || justClosedQuotes {
				arg := current.String()
				// Check if this is a named parameter (key="value" format)
				if key, value, isNamed := parseNamedParam(arg); isNamed {
					namedParams[key] = value
				} else {
					positionalArgs = append(positionalArgs, arg)
				}
				current.Reset()
				justClosedQuotes = false
			}
		} else {
			// Regular character
			current.WriteByte(ch)
			justClosedQuotes = false
		}
	}

	// Add the last argument
	if current.Len() > 0 || justClosedQuotes {
		arg := current.String()
		// Check if this is a named parameter (key="value" format)
		if key, value, isNamed := parseNamedParam(arg); isNamed {
			namedParams[key] = value
		} else {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Check for unclosed quotes
	if inQuotes {
		return nil, nil, fmt.Errorf("unclosed quote in arguments")
	}

	return positionalArgs, namedParams, nil
}

// parseNamedParam checks if a string is a named parameter in key=value format.
// The value must have been quoted (quotes are already stripped by the caller).
// Returns the key, value, and whether it was a named parameter.
func parseNamedParam(arg string) (key string, value string, isNamed bool) {
	// Find the equals sign
	eqIdx := strings.Index(arg, "=")
	if eqIdx == -1 {
		return "", "", false
	}

	key = arg[:eqIdx]
	// Key must be a valid identifier (non-empty, no spaces)
	if key == "" || strings.ContainsAny(key, " \t") {
		return "", "", false
	}

	value = arg[eqIdx+1:]
	return key, value, true
}
