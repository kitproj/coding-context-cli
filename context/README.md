# Context Package

The `context` package provides a library for dynamically assembling context for AI coding agents. This package can be embedded in other Go applications to programmatically collect and assemble context from various rule files and task prompts.

## Overview

The context package extracts the core functionality from the `coding-context-cli` tool, making it reusable as a library. It allows you to:

- Assemble context from rule files and task prompts
- Filter rules based on frontmatter metadata
- Substitute parameters in task prompts
- Run bootstrap scripts before processing rules
- Estimate token counts for the assembled context

## Installation

```bash
go get github.com/kitproj/coding-context-cli/context
```

## Quick Start

Here's a simple example of using the context package:

```go
package main

import (
	"context"
	"fmt"
	"os"

	ctxlib "github.com/kitproj/coding-context-cli/context"
)

func main() {
	// Create parameters for substitution
	params := make(ctxlib.ParamMap)
	params["component"] = "auth"
	params["issue"] = "login bug"

	// Create selectors for filtering rules
	selectors := make(ctxlib.SelectorMap)
	selectors["language"] = "go"

	// Configure the assembler
	config := ctxlib.Config{
		WorkDir:   ".",
		TaskName:  "fix-bug",
		Params:    params,
		Selectors: selectors,
	}

	// Create the assembler
	assembler := ctxlib.NewAssembler(config)

	// Assemble the context
	ctx := context.Background()
	if err := assembler.Assemble(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

## Core Types

### Config

`Config` holds the configuration for context assembly:

```go
type Config struct {
	// WorkDir is the working directory to use
	WorkDir string
	
	// TaskName is the name of the task to execute
	TaskName string
	
	// Params are parameters for substitution in task prompts
	Params ParamMap
	
	// Selectors are frontmatter selectors for filtering rules
	Selectors SelectorMap
	
	// Stdout is where assembled context is written (defaults to os.Stdout)
	Stdout io.Writer
	
	// Stderr is where progress messages are written (defaults to os.Stderr)
	Stderr io.Writer
}
```

### Assembler

`Assembler` assembles context from rule and task files:

```go
// Create a new assembler
assembler := context.NewAssembler(config)

// Assemble the context
err := assembler.Assemble(ctx)
```

### ParamMap

`ParamMap` represents a map of parameters for substitution in task prompts:

```go
params := make(context.ParamMap)
params["key"] = "value"

// Or use the Set method (useful for flag parsing)
params.Set("key=value")
```

### SelectorMap

`SelectorMap` is used for filtering rules based on frontmatter metadata:

```go
selectors := make(context.SelectorMap)
selectors["language"] = "go"
selectors["environment"] = "production"

// Or use the Set method (useful for flag parsing)
selectors.Set("language=go")
```

## Utility Functions

### ParseMarkdownFile

Parse a markdown file with YAML frontmatter:

```go
var frontmatter map[string]string
content, err := context.ParseMarkdownFile("path/to/file.md", &frontmatter)
```

### EstimateTokens

Estimate the number of LLM tokens in text:

```go
tokens := context.EstimateTokens("This is some text")
```

## Advanced Usage

### Custom Output Writers

You can redirect the output to custom writers:

```go
var stdout, stderr bytes.Buffer

config := context.Config{
	WorkDir:   ".",
	TaskName:  "my-task",
	Params:    make(context.ParamMap),
	Selectors: make(context.SelectorMap),
	Stdout:    &stdout,  // Assembled context goes here
	Stderr:    &stderr,  // Progress messages go here
}

assembler := context.NewAssembler(config)
err := assembler.Assemble(ctx)

// Now you can process the output
contextContent := stdout.String()
progressMessages := stderr.String()
```

### Context Cancellation

The `Assemble` method respects context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

assembler := context.NewAssembler(config)
err := assembler.Assemble(ctx)
```

## File Search Paths

The assembler searches for task and rule files in predefined locations:

**Tasks:**
- `./.agents/tasks/<task-name>.md`
- `~/.agents/tasks/<task-name>.md`
- `/etc/agents/tasks/<task-name>.md`

**Rules:**
The tool searches for various files and directories, including:
- `CLAUDE.local.md`
- `.agents/rules`, `.cursor/rules`, `.augment/rules`, `.windsurf/rules`
- `.opencode/agent`, `.opencode/command`
- `.github/copilot-instructions.md`, `.gemini/styleguide.md`
- `AGENTS.md`, `CLAUDE.md`, `GEMINI.md` (and in parent directories)
- User-specific rules in `~/.agents/rules`, `~/.claude/CLAUDE.md`, etc.
- System-wide rules in `/etc/agents/rules`, `/etc/opencode/rules`

## Bootstrap Scripts

Bootstrap scripts are executed before processing rule files. If a rule file `setup.md` exists, the assembler will look for `setup-bootstrap` and execute it if found. This is useful for environment setup or tool installation.

## Testing

The package includes comprehensive tests. Run them with:

```bash
go test github.com/kitproj/coding-context-cli/context
```

## License

This package is part of the coding-context-cli project and follows the same license.
