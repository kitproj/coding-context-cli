# codingcontext

Go package for dynamically assembling context for AI coding agents.

## Installation

```bash
go get github.com/kitproj/coding-context-cli/pkg/codingcontext
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

func main() {
    // Create a new context with options
    ctx := codingcontext.New(
        codingcontext.WithWorkDir("."),
        codingcontext.WithParams(codingcontext.Params{
            "issue_number": "123",
            "feature":      "authentication",
        }),
        codingcontext.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
    )

    // Run a task
    if err := ctx.Run(context.Background(), "my-task"); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Advanced Example

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

func main() {
    // Create a buffer to capture output
    var output bytes.Buffer

    // Create selectors for filtering rules
    selectors := make(codingcontext.Selectors)
    selectors.SetValue("language", "go")
    selectors.SetValue("stage", "implementation")

    // Create context with all options
    ctx := codingcontext.New(
        codingcontext.WithWorkDir("."),
        codingcontext.WithResume(false),
        codingcontext.WithParams(codingcontext.Params{
            "issue_number": "123",
        }),
        codingcontext.WithSelectors(selectors),
        codingcontext.WithRemotePaths([]string{
            "https://github.com/org/repo//path/to/rules",
        }),
        codingcontext.WithEmitTaskFrontmatter(true),
        codingcontext.WithOutput(&output),
        codingcontext.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
    )

    // Run the task
    if err := ctx.Run(context.Background(), "implement-feature"); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Use the generated context
    fmt.Println(output.String())
}
```

## API Reference

### Types

#### `Context`

The main type for assembling context.

#### `Params`

Map of parameter key-value pairs for template substitution.

#### `Selectors`

Map structure for filtering rules based on frontmatter metadata.

#### `FrontMatter`

Map representing parsed YAML frontmatter from markdown files.

### Functions

#### `New(opts ...Option) *Context`

Creates a new Context with the given options.

**Options:**
- `WithWorkDir(dir string)` - Set the working directory
- `WithResume(resume bool)` - Enable resume mode
- `WithParams(params Params)` - Set parameters
- `WithSelectors(selectors Selectors)` - Set selectors for filtering
- `WithRemotePaths(paths []string)` - Set remote directories to download
- `WithEmitTaskFrontmatter(emit bool)` - Enable task frontmatter emission
- `WithOutput(w io.Writer)` - Set output writer
- `WithLogger(logger *slog.Logger)` - Set logger
- `WithCmdRunner(runner func(*exec.Cmd) error)` - Set custom command runner

#### `(*Context) Run(ctx context.Context, taskName string) error`

Executes the context assembly for the given task name.

#### `ParseMarkdownFile(path string, frontmatter any) (string, error)`

Parses a markdown file into frontmatter and content.

#### `EstimateTokens(text string) int`

Estimates the number of LLM tokens in the given text.

#### `AllTaskSearchPaths(homeDir string) []string`

Returns the standard search paths for task files.

#### `AllRulePaths(homeDir string) []string`

Returns the standard search paths for rule files.

#### `DownloadRemoteDirectory(ctx context.Context, src string) (string, error)`

Downloads a remote directory using go-getter.

## See Also

- [Main CLI Tool](https://github.com/kitproj/coding-context-cli)
- [Documentation](https://kitproj.github.io/coding-context-cli/)
