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
        codingcontext.WithSearchPaths("file://.", "file://"+os.Getenv("HOME")),
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
        codingcontext.WithSearchPaths(
            "file://.",
            "git::https://github.com/org/repo//path/to/rules",
        ),
        codingcontext.WithParams(codingcontext.Params{
            "issue_number": "123",
        }),
        codingcontext.WithSelectors(selectors),
        codingcontext.WithAgent(codingcontext.AgentCursor),
        codingcontext.WithResume(false),
        codingcontext.WithUserPrompt("Additional context or instructions"),
        codingcontext.WithManifestURL("https://example.com/manifest.txt"),
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
    fmt.Printf("Total tokens: %d\n", result.Tokens)
    fmt.Printf("Agent: %s\n", result.Agent)
    
    // Access task metadata using typed fields
    if len(result.Task.FrontMatter.Languages) > 0 {
        fmt.Printf("Languages: %v\n", result.Task.FrontMatter.Languages)
    }
    
    // You can also access any frontmatter field via the Content map
    if customField, ok := result.Task.FrontMatter.Content["custom_field"]; ok {
        fmt.Printf("Custom field: %v\n", customField)
    }
    
    // Access MCP server configurations
    mcpServers := result.MCPServers()
    for name, config := range mcpServers {
        fmt.Printf("MCP Server %s: %s\n", name, config.Command)
    }
}
```

## API Reference

### Types

#### `Context`

The main type for assembling context.

#### `Result`

Result holds the assembled context from running a task:
- `Rules []Markdown[RuleFrontMatter]` - List of included rule files
- `Task Markdown[TaskFrontMatter]` - Task file with frontmatter and content
- `Tokens int` - Total estimated token count
- `Agent Agent` - The agent used (from task frontmatter or option)

**Methods:**
- `MCPServers() MCPServerConfigs` - Returns all MCP server configurations from rules and task

#### `Markdown[T]`

Represents a markdown file with frontmatter and content (generic type):
- `FrontMatter T` - Parsed YAML frontmatter (type depends on usage)
- `Content string` - Expanded content of the markdown
- `Tokens int` - Estimated token count

Type aliases:
- `TaskMarkdown` = `Markdown[TaskFrontMatter]`
- `RuleMarkdown` = `Markdown[RuleFrontMatter]`

#### `TaskFrontMatter`

Frontmatter structure for task files with fields:
- `Agent string` - Default agent if not specified via option
- `Languages []string` - Programming languages for filtering rules
- `Model string` - AI model identifier (metadata only)
- `SingleShot bool` - Whether task runs once or multiple times (metadata only)
- `Timeout string` - Task timeout in time.Duration format (metadata only)
- `MCPServers MCPServerConfigs` - MCP server configurations (metadata only)
- `Resume bool` - Whether this task should be resumed
- `Selectors map[string]any` - Additional custom selectors for filtering rules
- `ExpandParams *bool` - Controls parameter expansion (defaults to true)
- `Content map[string]any` - All frontmatter fields as map (from `BaseFrontMatter`)

#### `RuleFrontMatter`

Frontmatter structure for rule files with fields:
- `TaskNames []string` - Which task(s) this rule applies to
- `Languages []string` - Which programming language(s) this rule applies to
- `Agent string` - Which AI agent this rule is intended for
- `MCPServers MCPServerConfigs` - MCP server configurations (metadata only)
- `RuleName string` - Optional identifier for the rule file
- `ExpandParams *bool` - Controls parameter expansion (defaults to true)
- `Content map[string]any` - All frontmatter fields as map (from `BaseFrontMatter`)

#### `CommandFrontMatter`

Frontmatter structure for command files with fields:
- `ExpandParams *bool` - Controls parameter expansion (defaults to true)
- `Content map[string]any` - All frontmatter fields as map (from `BaseFrontMatter`)

#### `BaseFrontMatter`

Base frontmatter structure that other frontmatter types embed:
- `Content map[string]any` - All frontmatter fields as a map for selector matching

#### `Agent`

Type representing an AI coding agent (string type).

**Constants:**
- `AgentCursor` - Cursor AI (cursor.sh)
- `AgentOpenCode` - OpenCode.ai agent
- `AgentCopilot` - GitHub Copilot
- `AgentClaude` - Anthropic Claude AI
- `AgentGemini` - Google Gemini AI
- `AgentAugment` - Augment Code assistant
- `AgentWindsurf` - Codeium Windsurf
- `AgentCodex` - Codex AI agent

**Methods:**
- `String() string` - Returns string representation
- `PathPatterns() []string` - Returns path patterns for this agent
- `MatchesPath(path string) bool` - Checks if path matches agent patterns
- `ShouldExcludePath(path string) bool` - Returns true if path should be excluded
- `IsSet() bool` - Returns true if agent is set (non-empty)
- `UserRulePath() string` - Returns user-level rules path for agent

#### `MCPServerConfig`

Configuration for MCP (Model Context Protocol) servers:
- `Type TransportType` - Connection protocol ("stdio", "sse", "http")
- `Command string` - Executable to run (for stdio type)
- `Args []string` - Command arguments
- `Env map[string]string` - Environment variables
- `URL string` - Endpoint URL (for http/sse types)
- `Headers map[string]string` - Custom HTTP headers

#### `MCPServerConfigs`

Type alias: `map[string]MCPServerConfig` - Maps server names to configurations

#### `TransportType`

Type representing MCP transport protocol (string type):

**Constants:**
- `TransportTypeStdio` - Local process communication
- `TransportTypeSSE` - Server-Sent Events (remote)
- `TransportTypeHTTP` - Standard HTTP/POST

#### `Params`

Map of parameter key-value pairs for template substitution: `map[string]string`

**Methods:**
- `String() string` - Returns string representation
- `Set(value string) error` - Parses and sets key=value pair (implements flag.Value)

#### `Selectors`

Map structure for filtering rules based on frontmatter metadata: `map[string]map[string]bool`

**Methods:**
- `String() string` - Returns string representation
- `Set(value string) error` - Parses and sets key=value pair (implements flag.Value)
- `SetValue(key, value string)` - Sets a value for a key
- `GetValue(key, value string) bool` - Checks if value exists for key
- `MatchesIncludes(frontmatter BaseFrontMatter) bool` - Tests if frontmatter matches selectors

#### Task Parser Types

Types for parsing task content with slash commands:

- `Task` - Slice of `Block` elements representing parsed task content
- `Block` - Contains either `Text` or `SlashCommand`
- `SlashCommand` - Parsed slash command with name and arguments
- `Text` - Text content (slice of `TextLine`)
- `TextLine` - Single line of text content
- `Input` - Top-level wrapper type for parsing
- `Argument` - Slash command argument (can be positional or named key=value)

**Methods:**
- `(*SlashCommand) Params() map[string]string` - Returns parsed parameters as map
- `(*Text) Content() string` - Returns text content as string
- Various `String()` methods for formatting each type

### Constants

#### `FreeTextTaskName`

Constant: `"free-text"` - Task name used for free-text prompts

#### `FreeTextParamName`

Constant: `"text"` - Parameter name for text content in free-text tasks

### Functions

#### `New(opts ...Option) *Context`

Creates a new Context with the given options.

**Options:**
- `WithSearchPaths(paths ...string)` - Add search paths (supports go-getter URLs)
- `WithParams(params Params)` - Set parameters for substitution
- `WithSelectors(selectors Selectors)` - Set selectors for filtering rules
- `WithAgent(agent Agent)` - Set target agent (excludes that agent's own rules)
- `WithResume(resume bool)` - Enable resume mode (skips rules)
- `WithUserPrompt(userPrompt string)` - Set user prompt to append to task
- `WithManifestURL(manifestURL string)` - Set manifest URL for additional search paths
- `WithLogger(logger *slog.Logger)` - Set logger

#### `(*Context) Run(ctx context.Context, taskName string) (*Result, error)`

Executes the context assembly for the given task name and returns the assembled result structure with rule and task markdown files (including frontmatter and content).

#### `ParseMarkdownFile[T any](path string, frontmatter *T) (Markdown[T], error)`

Parses a markdown file into frontmatter and content. Generic function that works with any frontmatter type.

#### `ParseTask(text string) (Task, error)`

Parses task text content into blocks of text and slash commands.

#### `ParseParams(s string) (Params, error)`

Parses a string containing key=value pairs with quoted values.

**Examples:**
```go
// Parse quoted key-value pairs
params, _ := ParseParams(`key1="value1" key2="value2"`)
// Result: map[string]string{"key1": "value1", "key2": "value2"}

// Parse with spaces in values
params, _ := ParseParams(`key1="value with spaces" key2="value2"`)
// Result: map[string]string{"key1": "value with spaces", "key2": "value2"}

// Parse with escaped quotes
params, _ := ParseParams(`key1="value with \"escaped\" quotes"`)
// Result: map[string]string{"key1": "value with \"escaped\" quotes"}
```

#### `ParseAgent(s string) (Agent, error)`

Parses a string into an Agent type. Returns error if agent is not supported.

## See Also

- [Main CLI Tool](https://github.com/kitproj/coding-context-cli)
- [Documentation](https://kitproj.github.io/coding-context-cli/)
