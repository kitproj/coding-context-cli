---
layout: default
title: How the CLI Works
---

# How the CLI Works

This document explains the internal architecture and execution flow of the Coding Context CLI.

## High-Level Architecture

The CLI follows a simple pipeline architecture:

```
Input (CLI args) → Discovery → Filtering → Assembly → Output
```

Each stage transforms data before passing it to the next stage.

## Execution Flow

### 1. Parse Command-Line Arguments

```
coding-context-cli -C /project -s language=Go -p issue=BUG-123 fix-bug
```

The CLI parses:
- Working directory: `/project`
- Selectors: `{language: Go}`
- Parameters: `{issue: BUG-123}`
- Task name: `fix-bug`

### 2. Change Working Directory

If `-C` is specified, change to that directory before processing:

```go
if workDir != "." {
    os.Chdir(workDir)
}
```

This affects:
- Where `./.agents/` is located
- Parent directory traversal
- Bootstrap script execution context

### 3. Discover Rule Files

The CLI searches for rule files in predetermined locations:

```
Search paths (in order):
1. Project-specific: ./.agents/rules/, ./.cursor/rules/, etc.
2. Parent directories: ../AGENTS.md, ../../AGENTS.md, etc.
3. User-specific: ~/.agents/rules/, ~/.claude/CLAUDE.md, etc.
```

For each location:
- List all `.md` and `.mdc` files
- Parse YAML frontmatter (if present)
- Store file path and frontmatter

### 4. Execute Bootstrap Scripts

For each discovered rule file `rule.md`:
- Check if `rule-bootstrap` exists in same directory
- If exists and executable, run it
- Bootstrap output goes to stderr (not included in context)
- Bootstrap failures are logged but don't stop execution

**Example:**
```
Found: .agents/rules/jira-context.md
Check for: .agents/rules/jira-context-bootstrap
If found: Execute and log output to stderr
```

### 5. Filter Rules by Selectors

If selectors are specified (e.g., `-s language=Go`):

```
For each rule:
    If rule has frontmatter:
        For each selector:
            If frontmatter[selector.key] != selector.value:
                Exclude rule
    Include rule (if not excluded)
```

**Examples:**
```yaml
# Rule 1
---
language: Go
stage: testing
---
```
- `-s language=Go` → ✅ Include
- `-s language=Python` → ❌ Exclude
- `-s language=Go -s stage=testing` → ✅ Include
- `-s language=Go -s stage=planning` → ❌ Exclude

Rules without frontmatter are always included (unless resume mode).

### 6. Discover Task Files

Search task file locations for files with matching `task_name`:

```
Search paths:
1. ./.agents/tasks/*.md
2. ~/.agents/tasks/*.md
```

For each `.md` file:
- Parse frontmatter
- Check if `task_name` matches requested task
- Check if selectors match (if specified)
- Store matching tasks

### 7. Select Task

From matching tasks:

**If no selectors:**
- Exactly 1 match required
- 0 matches → Error: "no task found"
- 2+ matches → Error: "multiple tasks found"

**If selectors specified:**
- Filter tasks by selectors
- Exactly 1 match required after filtering
- 0 matches → Error: "no task found"
- 2+ matches → Error: "multiple tasks found"

### 8. Substitute Parameters

In the selected task content, replace `${param}` with values:

```markdown
# Input
Fix bug: ${issue_key}
Description: ${description}

# With -p issue_key=BUG-123 -p description="Crashes"
Fix bug: BUG-123
Description: Crashes
```

Parameter substitution uses environment variable expansion syntax, allowing:
- Simple substitution: `${var}`
- Default values: `${var:-default}`
- Error on unset: `${var:?error message}`

### 9. Assemble Output

Combine all pieces in order:

```
1. All included rule files (content only, no frontmatter)
2. Task content (with parameters substituted)
```

**Example output:**
```markdown
# Go Coding Standards

- Use gofmt
- Handle errors
...

# Python Coding Standards

- Follow PEP 8
- Use type hints
...

# Task: Fix Bug BUG-123

Analyze and fix the following bug...
```

### 10. Count Tokens

Estimate token count using a simple algorithm:

```
tokens ≈ (characters / 4) + (words / 0.75)
```

This provides a rough estimate for monitoring context size.

### 11. Output

Write to stdout and stderr:

**stdout (the context):**
```
[Rule 1 content]

[Rule 2 content]

[Task content with substituted parameters]
```

**stderr (metadata):**
```
Estimated tokens: 1,234
Bootstrap: jira-context-bootstrap executed
Found 3 rule files
Selected task: fix-bug
```

## Special Modes

### Resume Mode (`-r`)

When resume mode is enabled:

1. **Skip all rules:**
   ```
   discoveredRules = []  # Empty, no rules included
   ```

2. **Add implicit selector:**
   ```
   selectors["resume"] = "true"
   ```

This combination:
- Saves tokens by excluding rules
- Selects resume-specific task prompts
- Useful for continuing work in new sessions

### Working Directory (`-C`)

Changes directory before any processing:

```
1. Parse CLI args
2. If -C specified: os.Chdir(directory)
3. Continue with discovery from new directory
```

This affects all relative paths used in discovery.

## Data Structures

### Rule

```go
type Rule struct {
    Path        string
    Content     string
    Frontmatter map[string]string
}
```

### Task

```go
type Task struct {
    Path        string
    Content     string
    Frontmatter map[string]string
}
```

### Selector

```go
type Selector struct {
    Key   string
    Value string
}
```

## Performance Considerations

### File Discovery

- Directories are scanned once at startup
- Parent directory traversal stops at root
- No recursive scanning of entire filesystem

### Frontmatter Parsing

- YAML parsed only for files that are found
- Parsing failures are logged but don't crash
- Only top-level fields are extracted

### Bootstrap Scripts

- Execute sequentially (not in parallel)
- Stderr is buffered and output at end
- Failed scripts log errors but continue

### Token Counting

- Simple character-based estimation
- Not sent to external API
- Minimal performance impact

## Error Handling

### Graceful Failures

The CLI continues execution on:
- Bootstrap script failures
- Frontmatter parsing errors
- Missing optional files

### Fatal Errors

The CLI exits with error on:
- No task found with given name
- Multiple tasks match (ambiguous)
- Required directories not readable
- Invalid YAML in task frontmatter

### Error Messages

Errors are written to stderr with context:

```
Error: no task found with task_name: fix-bug
Searched:
  - ./.agents/tasks/
  - ~/.agents/tasks/

Tip: Check that your task file has 'task_name: fix-bug' in frontmatter
```

## Design Principles

### Simplicity

- Single binary, no dependencies
- Simple file format (Markdown + YAML)
- Clear, predictable behavior

### Composability

- Output to stdout for piping
- Metadata to stderr for logging
- Standard Unix tool conventions

### Flexibility

- Multiple rule sources
- Flexible file organization
- Selector-based filtering

### Convention over Configuration

- Predetermined search paths
- Standard file extensions
- Naming conventions for bootstraps

## Limitations

### No Nested Frontmatter

Selectors only match top-level YAML fields:

```yaml
# ✅ Can match
---
language: Go
---

# ❌ Cannot match nested
---
metadata:
  language: Go
---
```

### No OR Logic in Selectors

Multiple selectors use AND logic:

```bash
# Requires BOTH language=Go AND stage=testing
coding-context-cli -s language=Go -s stage=testing fix-bug

# No way to specify: language=Go OR language=Python
```

### No Rule Ordering

Rules are included in filesystem order, which may vary by platform. If order matters, use a single rule file.

### Static Bootstrap

Bootstrap scripts run once before rule processing, not dynamically during agent execution.

## See Also

- [CLI Reference](../reference/cli) - Command-line interface
- [File Formats Reference](../reference/file-formats) - File specifications
- [Agentic Workflows](./agentic-workflows) - Conceptual overview
