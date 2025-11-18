# codingcontext - Reusable Library

This package provides the core functionality for assembling dynamic context for AI coding agents. It can be used programmatically in Go applications.

## Usage Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

func main() {
    // Create a new context with options
    params := make(codingcontext.ParamMap)
    params.Set("issue_number=123")
    
    includes := make(codingcontext.SelectorMap)
    includes.Set("language=Go")
    
    ctx := codingcontext.New(
        codingcontext.WithWorkDir("."),
        codingcontext.WithParams(params),
        codingcontext.WithIncludes(includes),
        codingcontext.WithOutput(os.Stdout),
        codingcontext.WithLogOutput(os.Stderr),
    )
    
    // Run the context assembly for a specific task
    if err := ctx.Run(context.Background(), "fix-bug"); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Get the total estimated tokens
    fmt.Fprintf(os.Stderr, "Total tokens: %d\n", ctx.TotalTokens())
}
```

## Public API

### Types

- `Context` - Main context builder
- `FrontMatter` - Map representing YAML frontmatter
- `ParamMap` - Map for parameter substitution
- `SelectorMap` - Map for filtering rules

### Functions

- `New(opts ...Option) *Context` - Create a new context
- `ParseMarkdownFile(path string, frontmatter any) (string, error)` - Parse markdown with frontmatter
- `EstimateTokens(text string) int` - Estimate token count
- `DownloadRemoteDirectory(ctx context.Context, src string) (string, error)` - Download remote directory

### Options

- `WithWorkDir(dir string)` - Set working directory
- `WithResume(resume bool)` - Enable resume mode
- `WithParams(params ParamMap)` - Set parameters
- `WithIncludes(includes SelectorMap)` - Set selectors
- `WithRemotePaths(paths []string)` - Set remote paths
- `WithEmitTaskFrontmatter(emit bool)` - Enable frontmatter emission
- `WithOutput(w io.Writer)` - Set output writer
- `WithLogOutput(w io.Writer)` - Set log output writer
- `WithCmdRunner(runner func(cmd *exec.Cmd) error)` - Set custom command runner (for testing)

### Methods

- `Run(ctx context.Context, taskName string) error` - Execute context assembly
- `TotalTokens() int` - Get total estimated tokens
