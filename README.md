# Coding Agent Context CLI

A command-line interface and Go library for dynamically assembling context for AI coding agents.

This tool collects context from predefined rule files and a task-specific prompt, substitutes parameters, and prints a single, combined context to standard output. This is useful for feeding a large amount of relevant information into an AI model like Claude, Gemini, or OpenAI's GPT series.

The package also provides a library API for processing markdown files with YAML frontmatter using a visitor pattern.

## Features

- **Dynamic Context Assembly**: Merges context from various source files.
- **Task-Specific Prompts**: Use different prompts for different tasks (e.g., `feature`, `bugfix`).
- **Rule-Based Context**: Define reusable context snippets (rules) that can be included or excluded.
- **Frontmatter Filtering**: Select rules based on metadata using frontmatter selectors.
- **Bootstrap Scripts**: Run scripts to fetch or generate context dynamically.
- **Parameter Substitution**: Inject values into your task prompts.
- **Token Estimation**: Get an estimate of the total token count for the generated context.

## Supported Coding Agents

This tool is compatible with configuration files from various AI coding agents and IDEs:

- **[Anthropic Claude](https://claude.ai/)**: `CLAUDE.md`, `CLAUDE.local.md`, `.claude/CLAUDE.md`
- **[Codex](https://codex.ai/)**: `AGENTS.md`, `.codex/AGENTS.md`
- **[Cursor](https://cursor.sh/)**: `.cursor/rules`, `.cursorrules`
- **[Augment](https://augmentcode.com/)**: `.augment/rules`, `.augment/guidelines.md`
- **[Windsurf](https://codeium.com/windsurf)**: `.windsurf/rules`, `.windsurfrules`
- **[OpenCode.ai](https://opencode.ai/)**: `.opencode/agent`, `.opencode/command`, `.opencode/rules`
- **[GitHub Copilot](https://github.com/features/copilot)**: `.github/copilot-instructions.md`, `.github/agents`
- **[Google Gemini](https://gemini.google.com/)**: `GEMINI.md`, `.gemini/styleguide.md`
- **Generic AI Agents**: `AGENTS.md`, `.agents/rules`

The tool automatically discovers and includes rules from these locations in your project, parent directories, user home directory (`~`), and system-wide directories (`/etc`).

## Installation

### As a CLI Tool

You can install the CLI by downloading the latest release from the [releases page](https://github.com/kitproj/coding-context-cli/releases) or by building from source.

```bash
# Example for Linux
sudo curl -fsL -o /usr/local/bin/coding-context-cli https://github.com/kitproj/coding-context-cli/releases/download/v0.1.0/coding-context-cli_linux_amd64
sudo chmod +x /usr/local/bin/coding-context-cli
```

### As a Go Library

To use this package as a library in your Go project:

```bash
go get github.com/kitproj/coding-context-cli
```

## Usage

### As a CLI Tool

```
Usage:
  coding-context-cli [options] <task-name>

Options:
  -C string
    	Change to directory before doing anything. (default ".")
  -p value
    	Parameter to substitute in the prompt. Can be specified multiple times as key=value.
  -s value
    	Include rules with matching frontmatter. Can be specified multiple times as key=value.
```

#### CLI Example

```bash
coding-context-cli -p jira_issue_key=PROJ-1234 fix-bug | llm -m gemini-pro
```

This command will:
1. Find the `fix-bug.md` task file.
2. Find all rule files in the search paths.
3. Filter the rules based on selectors.
4. Execute any associated bootstrap scripts.
5. Substitute `${jira_issue_key}` with `PROJ-1234` in the task prompt.
6. Print the combined context (rules + task) to `stdout`.
7. Pipe the output to another program (in this case, `llm`).

### As a Go Library

The package provides a visitor pattern API for processing markdown files with YAML frontmatter.

#### Library API

```go
// FrontMatter represents the YAML frontmatter of a markdown file
type FrontMatter map[string]any

// MarkdownVisitor is a function that processes a markdown file's frontmatter and content
type MarkdownVisitor func(frontMatter FrontMatter, content string) error

// Visit parses markdown files matching the given pattern and calls the visitor for each file.
// It stops on the first error.
func Visit(pattern string, visitor MarkdownVisitor) error
```

#### Library Example

```go
package main

import (
    "fmt"
    context "github.com/kitproj/coding-context-cli"
)

func main() {
    // Define a visitor function to process each markdown file
    visitor := func(frontMatter context.FrontMatter, content string) error {
        // Access frontmatter fields
        if title, ok := frontMatter["title"].(string); ok {
            fmt.Printf("Title: %s\n", title)
        }
        
        // Process content
        fmt.Printf("Content length: %d bytes\n", len(content))
        
        // Return error to stop processing
        return nil
    }
    
    // Visit all markdown files in current directory
    if err := context.Visit("*.md", visitor); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
}
```

The visitor pattern allows you to:
- Process markdown files with YAML frontmatter
- Access both frontmatter metadata and content separately
- Stop processing on the first error
- Use glob patterns to select files

### Example Tasks

The `<task-name>` is the name of the task you want the agent to perform. Here are some common examples:

- `triage-bug`
- `review-pull-request`
- `fix-broken-build`
- `migrate-java-version`
- `enhance-docs`
- `remove-feature-flag`
- `speed-up-build`

Each of these would have a corresponding `.md` file in a `tasks` directory (e.g., `triage-bug.md`).

## How It Works

The tool assembles the context in the following order:

1.  **Rule Files**: It searches a list of predefined locations for rule files (`.md` or `.mdc`). These locations include the current directory, ancestor directories, user's home directory, and system-wide directories.
2.  **Bootstrap Scripts**: For each rule file found (e.g., `my-rule.md`), it looks for an executable script named `my-rule-bootstrap`. If found, it runs the script before processing the rule file. These scripts are meant for bootstrapping the environment (e.g., installing tools) and their output is sent to `stderr`, not into the main context.
3.  **Filtering**: If `-s` (include) flag is used, it parses the YAML frontmatter of each rule file to decide whether to include it.
4.  **Task Prompt**: It finds the task prompt file (e.g., `<task-name>.md`) in one of the search paths.
5.  **Parameter Expansion**: It substitutes variables in the task prompt using the `-p` flags.
6.  **Output**: It prints the content of all included rule files, followed by the expanded task prompt, to standard output.
7.  **Token Count**: A running total of estimated tokens is printed to standard error.

### File Search Paths

The tool looks for task and rule files in the following locations, in order of precedence:

**Tasks:**
- `./.agents/tasks/<task-name>.md`
- `~/.agents/tasks/<task-name>.md`
- `/etc/agents/tasks/<task-name>.md`

**Rules:**
The tool searches for a variety of files and directories, including:
- `CLAUDE.local.md`
- `.agents/rules`, `.cursor/rules`, `.augment/rules`, `.windsurf/rules`, `.opencode/agent`, `.opencode/command`
- `.github/copilot-instructions.md`, `.gemini/styleguide.md`
- `AGENTS.md`, `CLAUDE.md`, `GEMINI.md` (and in parent directories)
- User-specific rules in `~/.agents/rules`, `~/.claude/CLAUDE.md`, `~/.opencode/rules`, etc.
- System-wide rules in `/etc/agents/rules`, `/etc/opencode/rules`.

## File Formats

### Task Files

Task files are Markdown files that can contain variables for substitution.

**Example (`.agents/tasks/fix-bug.md`):**
```markdown
# Task: Fix Bug in ${jira_issue_key}

Here is the context for the bug. Please analyze the following files and provide a fix.
```

### Rule Files

Rule files are Markdown (`.md`) or `.mdc` files, optionally with YAML frontmatter for filtering.

**Example (`.agents/rules/backend.md`):**
```markdown
---
language: Go
---

# Backend Coding Standards

- All new code must be accompanied by unit tests.
- Use the standard logging library.
```

To include this rule only when working on the backend, you would use `-s system=backend`.

### Bootstrap Scripts

A bootstrap script is an executable file that has the same name as a rule file but with a `-bootstrap` suffix. These scripts are used to prepare the environment, for example by installing necessary tools. The output of these scripts is sent to `stderr` and is not part of the AI context.

**Example:**
- Rule file: `.agents/rules/jira.md`
- Bootstrap script: `.agents/rules/jira-bootstrap`

If `jira-bootstrap` is an executable script, it will be run before its corresponding rule file is processed.

**`.agents/rules/jira-bootstrap`:**
```bash
#!/bin/bash
# This script installs the jira-cli if it's not already present.
if ! command -v jira-cli &> /dev/null
then
    echo "Installing jira-cli..." >&2
    # Add installation commands here
fi
```
