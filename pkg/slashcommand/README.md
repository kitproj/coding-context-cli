# slashcommand

Package `slashcommand` provides a parser for slash commands commonly used in AI coding assistants.

## Overview

This package parses slash commands using bash-like argument parsing. The parser can find slash commands anywhere in the input text, not just at the beginning:
```
/task-name arg1 "arg 2" arg3
```

The parser extracts:
- **Task name**: The command identifier (without the leading `/`)
- **Arguments**: Positional arguments accessed via `$ARGUMENTS`, `$1`, `$2`, `$3`, etc.
- **Found status**: Boolean indicating whether a slash command was found

Arguments are parsed like bash:
- Quoted arguments (single or double quotes) can contain spaces
- Quotes are removed from parsed arguments
- Escape sequences are supported in double quotes (`\"`)
- Arguments are extracted until end of line

## Installation

```bash
go get github.com/kitproj/coding-context-cli/pkg/slashcommand
```

## Usage

```go
import "github.com/kitproj/coding-context-cli/pkg/slashcommand"

// Parse a simple command
found, taskName, params, err := slashcommand.ParseSlashCommand("/fix-bug")
// found: true
// taskName: "fix-bug"
// params: map[]

// Parse a command with arguments
found, taskName, params, err := slashcommand.ParseSlashCommand("/fix-bug 123")
// found: true
// taskName: "fix-bug"
// params: map["ARGUMENTS": "123", "1": "123"]

// Parse a command with quoted arguments
found, taskName, params, err := slashcommand.ParseSlashCommand(`/code-review "Fix login bug" high`)
// found: true
// taskName: "code-review"
// params: map["ARGUMENTS": "\"Fix login bug\" high", "1": "Fix login bug", "2": "high"]

// Command found in middle of text
found, taskName, params, err := slashcommand.ParseSlashCommand("Please /deploy production now")
// found: true
// taskName: "deploy"
// params: map["ARGUMENTS": "production now", "1": "production", "2": "now"]

// No command found
found, taskName, params, err := slashcommand.ParseSlashCommand("No command here")
// found: false
// taskName: ""
// params: nil
```

## Command Format

### Basic Structure
```
/task-name arg1 "arg 2" arg3 ...
```

### Argument Parsing Rules
1. Slash commands can appear **anywhere** in the input text
2. Task name comes immediately after the `/` (no spaces)
3. Arguments are extracted until end of line (newline stops argument collection)
4. Arguments can be quoted with single (`'`) or double (`"`) quotes
5. Quoted arguments can contain spaces
6. Quotes are removed from parsed arguments
7. Double quotes support escape sequences: `\"`
8. Single quotes preserve everything literally (no escapes)
9. Text before the `/` is ignored (prefix lost)
10. Text after a newline is ignored (suffix lost)

### Returned Parameters
The `params` map contains:
- `ARGUMENTS`: The full argument string (with quotes preserved)
- `1`, `2`, `3`, etc.: Individual positional arguments (with quotes removed)

### Valid Examples
```
/fix-bug                           # No arguments
/fix-bug 123                       # Single argument: $1 = "123"
/deploy staging v1.2.3             # Two arguments: $1 = "staging", $2 = "v1.2.3"
/code-review "PR #42"              # Quoted argument: $1 = "PR #42"
/echo 'He said "hello"'            # Single quotes preserve quotes: $1 = "He said \"hello\""
/echo "He said \"hello\""          # Escaped quotes in double quotes: $1 = "He said \"hello\""
Please /fix-bug 123 today          # Command in middle: task = "fix-bug", $1 = "123", $2 = "today"
Text /deploy prod\nNext line       # Arguments stop at newline: task = "deploy", $1 = "prod"
```

### Cases with No Command Found
```
fix-bug                    # Missing leading /: found = false
No command here            # No slash: found = false
```

## Error Handling

The parser returns errors only for malformed commands (e.g., unclosed quotes). If no slash command is found, the function returns `found=false` without an error.

```go
// No command found - not an error
found, _, _, err := slashcommand.ParseSlashCommand("No command here")
// found: false, err: nil

// Unclosed quote - returns error
found, _, _, err := slashcommand.ParseSlashCommand(`/fix-bug "unclosed`)
// found: false, err: "unclosed quote in arguments"
```

## API

### ParseSlashCommand

```go
func ParseSlashCommand(command string) (found bool, taskName string, params map[string]string, err error)
```

Parses a slash command string and extracts the task name and arguments. The function searches for a slash command anywhere in the input text.

**Parameters:**
- `command` (string): The text that may contain a slash command

**Returns:**
- `found` (bool): True if a slash command was found, false otherwise
- `taskName` (string): The task name without the leading `/`
- `params` (map[string]string): Contains `ARGUMENTS` (full arg string) and `1`, `2`, `3`, etc. (positional args)
- `err` (error): Error if the command format is invalid (e.g., unclosed quotes)

## Testing

The package includes comprehensive tests covering:
- Commands without arguments
- Commands with single and multiple arguments
- Quoted arguments (both single and double quotes)
- Escaped quotes
- Empty quoted arguments
- Commands embedded in text (prefix/suffix text)
- Commands with newlines
- Edge cases and error conditions

Run tests with:
```bash
go test -v ./pkg/slashcommand
```

## License

This package is part of the [coding-context-cli](https://github.com/kitproj/coding-context-cli) project and is licensed under the MIT License.
