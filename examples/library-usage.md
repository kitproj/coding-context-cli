# Library Usage Example

This directory contains an example of how to use the coding-context library programmatically.

## Example Program

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

func main() {
    // Set up parameters for variable substitution
    params := make(codingcontext.ParamMap)
    params.Set("issue_number=123")
    params.Set("branch=feature/new-feature")
    
    // Set up selectors to filter rules
    includes := make(codingcontext.SelectorMap)
    includes.Set("language=Go")
    includes.Set("stage=implementation")
    
    // Create a new context with options
    ctx := codingcontext.New(
        codingcontext.WithWorkDir("."),
        codingcontext.WithParams(params),
        codingcontext.WithIncludes(includes),
        codingcontext.WithOutput(os.Stdout),
        codingcontext.WithLogOutput(os.Stderr),
    )
    
    // Run the context assembly for a specific task
    if err := ctx.Run(context.Background(), "implement-feature"); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Access the total estimated tokens
    fmt.Fprintf(os.Stderr, "Total tokens: %d\n", ctx.TotalTokens())
}
```

## Running the Example

1. Make sure you have a `.agents/tasks` directory with task files
2. Optionally create rule files in locations like `CLAUDE.md`, `.agents/rules/`, etc.
3. Run the program:

```bash
go run example.go
```

## Advanced Usage

### Resume Mode

```go
ctx := codingcontext.New(
    codingcontext.WithResume(true),  // Skip rules, only output task
    // ... other options
)
```

### Remote Rules

```go
ctx := codingcontext.New(
    codingcontext.WithRemotePaths([]string{
        "git::https://github.com/company/shared-rules.git",
        "https://cdn.example.com/coding-standards",
    }),
    // ... other options
)
```

### Custom Command Runner (for testing)

```go
mockRunner := func(cmd *exec.Cmd) error {
    // Mock command execution for testing
    return nil
}

ctx := codingcontext.New(
    codingcontext.WithCmdRunner(mockRunner),
    // ... other options
)
```

## API Reference

See [pkg/codingcontext/README.md](../pkg/codingcontext/README.md) for complete API documentation.
