package codingcontext

import (
	"fmt"
	"strings"
)

// parseSlashCommand parses a slash command string and extracts the task name and parameters.
// It searches for a slash command anywhere in the input string, not just at the beginning.
// The expected format is: /task-name arg1 "arg 2" arg3
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
// Examples:
//   - "/fix-bug 123" -> taskName: "fix-bug", params: {"ARGUMENTS": "123", "1": "123"}, found: true
//   - "/code-review \"PR #42\" high" -> taskName: "code-review", params: {"ARGUMENTS": "\"PR #42\" high", "1": "PR #42", "2": "high"}, found: true
//   - "no command here" -> taskName: "", params: nil, found: false
//
// Returns:
//   - taskName: the task name (without the leading slash)
//   - params: a map containing:
//   - "ARGUMENTS": the full argument string (with quotes preserved)
//   - "1", "2", "3", etc.: positional arguments (with quotes removed)
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

	// Parse positional arguments using bash-like parsing
	args, err := parseBashArgs(argsString)
	if err != nil {
		return "", nil, false, err
	}

	// Add positional arguments as $1, $2, $3, etc.
	for i, arg := range args {
		params[fmt.Sprintf("%d", i+1)] = arg
	}

	return taskName, params, true, nil
}

// parseBashArgs parses a string into arguments like bash does, respecting quoted values
func parseBashArgs(s string) ([]string, error) {
	var args []string
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
				args = append(args, current.String())
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
		args = append(args, current.String())
	}

	// Check for unclosed quotes
	if inQuotes {
		return nil, fmt.Errorf("unclosed quote in arguments")
	}

	return args, nil
}
