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

    // Run a task and get the result
    result, err := ctx.Run(context.Background(), "my-task")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Access the assembled context
    for _, rule := range result.Rules {
        fmt.Println(rule.Content)
    }
    fmt.Println(result.Task.Content)
}
}
```

### Advanced Example

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
        codingcontext.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
    )

    // Run the task and get the result
    result, err := ctx.Run(context.Background(), "implement-feature")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Process the result
    fmt.Printf("Task: %s\n", result.Task.Content)
    fmt.Printf("Rules found: %d\n", len(result.Rules))
    
    // Access task metadata
    if taskName, ok := result.Task.FrontMatter["task_name"]; ok {
        fmt.Printf("Task name from frontmatter: %s\n", taskName)
    }
}
```

## API Reference

### Types

#### `Context`

The main type for assembling context.

#### `Result`

Result holds the assembled context from running a task:
- `Rules []Markdown` - List of included rule files
- `Task Markdown` - Task file with frontmatter and content

#### `Markdown`

Represents a markdown file with frontmatter and content:
- `Path string` - Path to the markdown file
- `FrontMatter FrontMatter` - Parsed YAML frontmatter
- `Content string` - Expanded content of the markdown
- `Tokens int` - Estimated token count

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
- `WithEmitTaskFrontmatter(emit bool)` - Enable task frontmatter inclusion in result
- `WithLogger(logger *slog.Logger)` - Set logger

#### `(*Context) Run(ctx context.Context, taskName string) (*Result, error)`

Executes the context assembly for the given task name and returns the assembled result structure with rule and task markdown files (including frontmatter and content).

#### `ParseMarkdownFile(path string, frontmatter any) (string, error)`

Parses a markdown file into frontmatter and content.

#### `AllTaskSearchPaths(baseDir, homeDir string) []string`

Returns the standard search paths for task files. `baseDir` is the working directory to resolve relative paths from.

#### `AllRulePaths(baseDir, homeDir string) []string`

Returns the standard search paths for rule files. `baseDir` is the working directory to resolve relative paths from.

## See Also

- [Main CLI Tool](https://github.com/kitproj/coding-context-cli)
- [Documentation](https://kitproj.github.io/coding-context-cli/)
