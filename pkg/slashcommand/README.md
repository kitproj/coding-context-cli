# slashcommand

Package `slashcommand` provides a parser for slash commands commonly used in AI coding assistants.

## Overview

This package parses slash commands in the format:
```
/task-name param1="value1" param2="value2"
```

The parser extracts:
- **Task name**: The command identifier (without the leading `/`)
- **Parameters**: Key-value pairs where values must be quoted

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

// Parse a command with parameters
taskName, params, err := slashcommand.ParseSlashCommand(`/fix-bug issue_number="123"`)
// taskName: "fix-bug"
// params: map["issue_number": "123"]

// Parse a command with multiple parameters
taskName, params, err := slashcommand.ParseSlashCommand(`/implement-feature feature_name="User Login" priority="high"`)
// taskName: "implement-feature"
// params: map["feature_name": "User Login", "priority": "high"]
```

## Command Format

### Basic Structure
```
/task-name param1="value1" param2="value2" ...
```

### Rules
1. Commands **must** start with `/`
2. Task name comes immediately after the `/` (no spaces)
3. Parameter values **must** be quoted with double quotes (`"`)
4. Parameters are separated by whitespace
5. Values can contain spaces when quoted

### Valid Examples
```
/fix-bug
/fix-bug issue_number="123"
/implement-feature feature_name="User Login"
/code-review pr_title="Fix bug in authentication flow"
/deploy environment="production" version="v1.2.3"
```

### Invalid Examples
```
fix-bug                    # Missing leading /
/fix-bug issue_number=123  # Value not quoted
/fix-bug issue_number='123' # Single quotes not allowed
/                          # Empty command
```

## Error Handling

The parser returns descriptive errors for invalid commands:

```go
_, _, err := slashcommand.ParseSlashCommand("fix-bug")
// Error: slash command must start with '/'

_, _, err := slashcommand.ParseSlashCommand("/")
// Error: slash command cannot be empty

_, _, err := slashcommand.ParseSlashCommand("/fix-bug issue_number=123")
// Error: parameter value must be quoted: issue_number=123

_, _, err := slashcommand.ParseSlashCommand(`/fix-bug issue_number="unclosed`)
// Error: unclosed quote in parameters
```

## API

### ParseSlashCommand

```go
func ParseSlashCommand(command string) (taskName string, params map[string]string, err error)
```

Parses a slash command string and extracts the task name and parameters.

**Parameters:**
- `command` (string): The slash command to parse

**Returns:**
- `taskName` (string): The task name without the leading `/`
- `params` (map[string]string): Parameter key-value pairs
- `err` (error): Error if the command format is invalid

## Testing

The package includes comprehensive tests covering:
- Simple commands without parameters
- Commands with single and multiple parameters
- Quoted values with spaces
- Edge cases and error conditions
- Parameter validation

Run tests with:
```bash
go test -v ./pkg/slashcommand
```

## License

This package is part of the [coding-context-cli](https://github.com/kitproj/coding-context-cli) project and is licensed under the MIT License.
