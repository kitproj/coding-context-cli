package codingcontext

import (
	"fmt"
	"strconv"
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
//   - Both single and double quotes are supported for positional arguments
//   - Quotes are removed from the parsed arguments
//   - Arguments are extracted until end of line
//
// Named parameters:
//   - Named parameters use key="value" format with mandatory double quotes
//   - Named parameters are also counted as positional arguments (retaining their original form)
//   - The named parameter key must not be a reserved key (ARGUMENTS) or a numeric key
//
// Examples:
//   - "/fix-bug 123" -> taskName: "fix-bug", params: {"ARGUMENTS": "123", "1": "123"}, found: true
//   - "/code-review \"PR #42\" high" -> taskName: "code-review", params: {"ARGUMENTS": "\"PR #42\" high", "1": "PR #42", "2": "high"}, found: true
//   - "/fix-bug issue=\"PROJ-123\"" -> taskName: "fix-bug", params: {"ARGUMENTS": "issue=\"PROJ-123\"", "1": "issue=\"PROJ-123\"", "issue": "PROJ-123"}, found: true
//   - "/task arg1 key=\"val\" arg2" -> taskName: "task", params: {"ARGUMENTS": "arg1 key=\"val\" arg2", "1": "arg1", "2": "key=\"val\"", "3": "arg2", "key": "val"}, found: true
//   - "no command here" -> taskName: "", params: nil, found: false
//
// Returns:
//   - taskName: the task name (without the leading slash)
//   - params: a map containing:
//   - "ARGUMENTS": the full argument string (with quotes preserved)
//   - "1", "2", "3", etc.: all arguments in order (with quotes removed), including named parameters in their original form
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
	allArgs, namedParams, err := parseBashArgsWithNamed(argsString)
	if err != nil {
		return "", nil, false, err
	}

	// Add all arguments as positional parameters $1, $2, $3, etc.
	// Named parameters are also included in positional numbering
	for i, arg := range allArgs {
		params[fmt.Sprintf("%d", i+1)] = arg
	}

	// Add named parameters (excluding reserved and numeric keys)
	for key, value := range namedParams {
		// Skip reserved key ARGUMENTS and numeric keys
		if key == "ARGUMENTS" {
			continue
		}
		if _, err := strconv.Atoi(key); err == nil {
			continue
		}
		params[key] = value
	}

	return taskName, params, true, nil
}

// parseBashArgsWithNamed parses a string into all arguments and named parameters.
// All arguments (including named parameters) are returned in allArgs for positional numbering.
// Named parameters must use key="value" format with mandatory double quotes.
// Returns all arguments, named parameters map, and any error.
func parseBashArgsWithNamed(s string) ([]string, map[string]string, error) {
	var allArgs []string
	namedParams := make(map[string]string)

	var current strings.Builder
	var rawArg strings.Builder // Tracks the raw argument including quotes
	inQuotes := false
	quoteChar := byte(0)
	escaped := false
	justClosedQuotes := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			current.WriteByte(ch)
			rawArg.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' && inQuotes && quoteChar == '"' {
			// Only recognize escape in double quotes
			escaped = true
			rawArg.WriteByte(ch)
			continue
		}

		if (ch == '"' || ch == '\'') && !inQuotes {
			// Start of quoted string
			inQuotes = true
			quoteChar = ch
			justClosedQuotes = false
			rawArg.WriteByte(ch)
		} else if ch == quoteChar && inQuotes {
			// End of quoted string - mark that we just closed quotes
			inQuotes = false
			quoteChar = 0
			justClosedQuotes = true
			rawArg.WriteByte(ch)
		} else if (ch == ' ' || ch == '\t') && !inQuotes {
			// Whitespace outside quotes - end of argument
			if current.Len() > 0 || justClosedQuotes {
				arg := current.String()
				rawArgStr := rawArg.String()

				// Check if this is a named parameter with mandatory double quotes
				if key, value, isNamed := parseNamedParamWithQuotes(rawArgStr); isNamed {
					allArgs = append(allArgs, rawArgStr)
					namedParams[key] = value
				} else {
					allArgs = append(allArgs, arg)
				}
				current.Reset()
				rawArg.Reset()
				justClosedQuotes = false
			}
		} else {
			// Regular character
			current.WriteByte(ch)
			rawArg.WriteByte(ch)
			justClosedQuotes = false
		}
	}

	// Add the last argument
	if current.Len() > 0 || justClosedQuotes {
		arg := current.String()
		rawArgStr := rawArg.String()

		// Check if this is a named parameter with mandatory double quotes
		if key, value, isNamed := parseNamedParamWithQuotes(rawArgStr); isNamed {
			allArgs = append(allArgs, rawArgStr)
			namedParams[key] = value
		} else {
			allArgs = append(allArgs, arg)
		}
	}

	// Check for unclosed quotes
	if inQuotes {
		return nil, nil, fmt.Errorf("unclosed quote in arguments")
	}

	return allArgs, namedParams, nil
}

// parseNamedParamWithQuotes checks if an argument is a named parameter in key="value" format.
// Double quotes are mandatory for the value portion.
// Returns the key, value (with quotes stripped), and whether it was a valid named parameter.
// Key must be non-empty and cannot contain spaces or tabs.
func parseNamedParamWithQuotes(rawArg string) (key string, value string, isNamed bool) {
	// Find the equals sign
	eqIdx := strings.Index(rawArg, "=")
	if eqIdx == -1 {
		return "", "", false
	}

	key = rawArg[:eqIdx]
	// Key must be a valid identifier (non-empty, no spaces or tabs)
	if key == "" || strings.ContainsAny(key, " \t") {
		return "", "", false
	}

	// The value portion (after '=')
	valuePart := rawArg[eqIdx+1:]

	// Value must start with double quote (mandatory)
	if len(valuePart) < 2 || valuePart[0] != '"' {
		return "", "", false
	}

	// Value must end with double quote
	if valuePart[len(valuePart)-1] != '"' {
		return "", "", false
	}

	// Extract the value between quotes and handle escaped quotes
	quotedValue := valuePart[1 : len(valuePart)-1]
	var unescaped strings.Builder
	for i := 0; i < len(quotedValue); i++ {
		if quotedValue[i] == '\\' && i+1 < len(quotedValue) && quotedValue[i+1] == '"' {
			unescaped.WriteByte('"')
			i++ // Skip the escaped quote
		} else {
			unescaped.WriteByte(quotedValue[i])
		}
	}

	return key, unescaped.String(), true
}
