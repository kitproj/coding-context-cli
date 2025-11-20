# slashcommand

Package `slashcommand` provides a parser for slash commands commonly used in AI coding assistants.

## Overview

This package parses slash commands using bash-like argument parsing:
```
/task-name arg1 "arg 2" arg3
```

The parser extracts:
- **Task name**: The command identifier (without the leading `/`)
- **Arguments**: Positional arguments accessed via `$ARGUMENTS`, `$1`, `$2`, `$3`, etc.

Arguments are parsed like bash:
- Quoted arguments (single or double quotes) can contain spaces
- Quotes are removed from parsed arguments
- Escape sequences are supported in double quotes (`\"`)

## Installation

```bash
go get github.com/kitproj/coding-context-cli/pkg/slashcommand
```

## Usage

```go
import "github.com/kitproj/coding-context-cli/pkg/slashcommand"

// Parse a simple command
taskName, params, err := slashcommand.ParseSlashCommand("/fix-bug")
// taskName: "fix-bug"
// params: map[]

// Parse a command with arguments
taskName, params, err := slashcommand.ParseSlashCommand("/fix-bug 123")
// taskName: "fix-bug"
// params: map["ARGUMENTS": "123", "1": "123"]

// Parse a command with quoted arguments
taskName, params, err := slashcommand.ParseSlashCommand(`/code-review "Fix login bug" high`)
// taskName: "code-review"
// params: map["ARGUMENTS": "\"Fix login bug\" high", "1": "Fix login bug", "2": "high"]
```

## Command Format

### Basic Structure
```
/task-name arg1 "arg 2" arg3 ...
```

### Argument Parsing Rules
1. Commands **must** start with `/`
2. Task name comes immediately after the `/` (no spaces)
3. Arguments can be quoted with single (`'`) or double (`"`) quotes
4. Quoted arguments can contain spaces
5. Quotes are removed from parsed arguments
6. Double quotes support escape sequences: `\"`
7. Single quotes preserve everything literally (no escapes)

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
```

### Invalid Examples
```
fix-bug                    # Missing leading /
/                          # Empty command
/fix-bug "unclosed         # Unclosed quote
```

## Error Handling

The parser returns descriptive errors for invalid commands:

```go
_, _, err := slashcommand.ParseSlashCommand("fix-bug")
// Error: slash command must start with '/'

_, _, err := slashcommand.ParseSlashCommand("/")
// Error: slash command cannot be empty

_, _, err := slashcommand.ParseSlashCommand(`/fix-bug "unclosed`)
// Error: unclosed quote in arguments
```

## API

### ParseSlashCommand

```go
func ParseSlashCommand(command string) (taskName string, params map[string]string, err error)
```

Parses a slash command string and extracts the task name and arguments.

**Parameters:**
- `command` (string): The slash command to parse

**Returns:**
- `taskName` (string): The task name without the leading `/`
- `params` (map[string]string): Contains `ARGUMENTS` (full arg string) and `1`, `2`, `3`, etc. (positional args)
- `err` (error): Error if the command format is invalid

## Testing

The package includes comprehensive tests covering:
- Commands without arguments
- Commands with single and multiple arguments
- Quoted arguments (both single and double quotes)
- Escaped quotes
- Empty quoted arguments
- Edge cases and error conditions

Run tests with:
```bash
go test -v ./pkg/slashcommand
```

## License

This package is part of the [coding-context-cli](https://github.com/kitproj/coding-context-cli) project and is licensed under the MIT License.
