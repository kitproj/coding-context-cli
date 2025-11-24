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

    // Get home directory for default search paths
    homeDir, _ := os.UserHomeDir()
    
    // Create context with all options
    ctx := codingcontext.New(
        codingcontext.WithWorkDir("."),
        codingcontext.WithParams(codingcontext.Params{
            "issue_number": "123",
        }),
        codingcontext.WithSelectors(selectors),
        codingcontext.WithSearchPaths(codingcontext.DefaultSearchPaths(".", homeDir)),
        codingcontext.WithPath("https://github.com/org/repo//path/to/rules"),
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

#### `SearchPath`

Represents a single search location with its associated subpaths:
- `BasePath string` - Base directory path to search
- `RulesSubPaths []string` - Relative subpaths within BasePath where rule files can be found
- `TaskSubPaths []string` - Relative subpaths within BasePath where task files can be found

### Functions

#### `New(opts ...Option) *Context`

Creates a new Context with the given options.

**Options:**
- `WithWorkDir(dir string)` - Set the working directory
- `WithParams(params Params)` - Set parameters
- `WithSelectors(selectors Selectors)` - Set selectors for filtering
- `WithSearchPaths(searchPaths []SearchPath)` - Set search paths to use. Typically called with `DefaultSearchPaths(baseDir, homeDir)` to enable default local path searching
- `WithPath(path string)` - Add a single path (local or remote) to be downloaded/copied and searched. Can be called multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, file://, etc.)
- `WithLogger(logger *slog.Logger)` - Set logger
- `WithResume(resume bool)` - Enable resume mode, which skips rule discovery and bootstrap scripts
- `WithAgent(agent Agent)` - Set the target agent, which excludes that agent's own rules

#### `(*Context) Run(ctx context.Context, taskName string) (*Result, error)`

Executes the context assembly for the given task name and returns the assembled result structure with rule and task markdown files (including frontmatter and content).

#### `ParseMarkdownFile(path string, frontmatter any) (string, error)`

Parses a markdown file into frontmatter and content.

#### `DefaultSearchPaths(baseDir, homeDir string) []SearchPath`

Returns the search paths for default local paths (baseDir and homeDir). Each `SearchPath` represents one base path with its associated rule and task subpaths.

#### `PathSearchPaths(dir string) []SearchPath`

Returns the search paths for a given directory path (used for both local and remote paths after download). Uses the same standard subpaths as downloaded directories.

## See Also

- [Main CLI Tool](https://github.com/kitproj/coding-context-cli)
- [Documentation](https://kitproj.github.io/coding-context-cli/)
