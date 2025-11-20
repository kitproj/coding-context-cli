package slashcommand

import (
	"fmt"
	"strings"
)

// ParseSlashCommand parses a slash command string and extracts the task name and parameters.
// The expected format is: /task-name arg1 "arg 2" arg3
//
// Arguments are parsed like Bash:
//   - Quoted arguments can contain spaces
//   - Both single and double quotes are supported
//   - Quotes are removed from the parsed arguments
//
// Examples:
//   - "/fix-bug 123" -> taskName: "fix-bug", params: {"ARGUMENTS": "123", "1": "123"}
//   - "/code-review \"PR #42\" high" -> taskName: "code-review", params: {"ARGUMENTS": "\"PR #42\" high", "1": "PR #42", "2": "high"}
//
// Returns:
//   - taskName: the task name (without the leading slash)
//   - params: a map containing:
//   - "ARGUMENTS": the full argument string (with quotes preserved)
//   - "1", "2", "3", etc.: positional arguments (with quotes removed)
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

	// Find the task name (first word)
	spaceIdx := strings.IndexAny(command, " \t")
	if spaceIdx == -1 {
		// No arguments, just the task name
		return command, make(map[string]string), nil
	}

	taskName = command[:spaceIdx]
	argsString := strings.TrimSpace(command[spaceIdx:])

	params = make(map[string]string)

	// Store the full argument string (with quotes preserved)
	params["ARGUMENTS"] = argsString

	// If there are no arguments, return early
	if argsString == "" {
		return taskName, params, nil
	}

	// Parse positional arguments using bash-like parsing
	args, err := parseBashArgs(argsString)
	if err != nil {
		return "", nil, err
	}

	// Add positional arguments as $1, $2, $3, etc.
	for i, arg := range args {
		params[fmt.Sprintf("%d", i+1)] = arg
	}

	return taskName, params, nil
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
