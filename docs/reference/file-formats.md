---
layout: default
title: File Formats
parent: Reference
nav_order: 2
---

# File Formats Reference

Technical specification for task files, rule files, and bootstrap scripts.

## Task Files

Task files define what the AI agent should do. They are Markdown files with YAML frontmatter.

### Format

```markdown
---
<optional-frontmatter-fields>
---

# Task content in Markdown

Content can include ${parameter_placeholders}.
```

**Note:** The `task_name` field is optional. Tasks are matched by filename (without `.md` extension), not by `task_name` in frontmatter. The `task_name` field is useful for metadata and appears in the frontmatter output.

### Frontmatter Fields

#### `task_name` (optional)

**Type:** String  
**Purpose:** Metadata field that identifies the task. Tasks are actually matched by filename (without `.md` extension), not by this field. This field is useful for metadata and can be used in task frontmatter output.

**Example:**
```yaml
---
---
```

**Usage:**
```bash
# Task is matched by filename "fix-bug.md", not by task_name field
coding-context fix-bug
```

**Note:** The `task_name` field is optional. If omitted, the task is still matched by its filename.

#### `languages` (optional, standard field)

**Type:** Array (recommended) or String  
**Purpose:** Metadata field that specifies the programming language(s) for the task. This field does NOT filter rules - it is metadata only and appears in the task frontmatter output.

The `languages` field is a **standard frontmatter field** that provides metadata about which programming language(s) the task relates to. Unlike selectors, this field does not automatically filter rules. To filter rules by language, use the `selectors` field or the `-s languages=go` command-line flag.

**Recommended format (array with lowercase values):**
```yaml
---
languages:
  - go
---
```

**Example (multiple languages):**
```yaml
---
languages:
  - go
  - python
  - javascript
---
```

**Note:** Both `language` (singular) and `languages` (plural) are accepted in YAML and map to the same field, but `languages` (plural) is recommended. Language values should be lowercase (e.g., `go`, `python`, `javascript`). The field is stored in frontmatter output but does not affect rule filtering.

**To filter rules by language, use selectors:**

**Important distinction:**
- Frontmatter metadata field: `languages:` (plural) - does NOT filter rules
- Selector key: `languages:` (plural) - used for filtering rules

**In task frontmatter selectors:**
```yaml
---
selectors:
  languages: go
---
```

**On the command line:**
```bash
coding-context -s languages=go implement-feature
```

**Note:** 
- Use `-s languages=go` (selector flag, plural `languages`)
- Do NOT use `-p languages=go` (`-p` is for parameter substitution, not filtering)
- Language values should be lowercase (e.g., `go`, `python`, `javascript`)

#### `single_shot` (optional, standard field)

**Type:** Boolean  
**Purpose:** Indicates whether the task should be run once or many times; stored in frontmatter output but does not filter rules

The `single_shot` field is a **standard frontmatter field** that provides metadata about task execution. It does not act as a selector.

**Example:**
```yaml
---
single_shot: true
---
```

**Common values:**
- `true` - Task runs once
- `false` - Task can run multiple times

#### `timeout` (optional, standard field)

**Type:** String (time.Duration format)  
**Purpose:** Specifies the timeout duration for the task using Go's time.Duration format; stored in frontmatter output but does not filter rules

The `timeout` field is a **standard frontmatter field** that provides metadata about task execution limits. It does not act as a selector.

**Example:**
```yaml
---
timeout: 10m
---
```

**Common time.Duration formats:**
- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `1h30m` - 1 hour 30 minutes

#### `mcp_server` (optional, standard field)

**Type:** Object (MCP server configuration)  
**Purpose:** Specifies a single MCP (Model Context Protocol) server configuration for the task; stored in frontmatter output but does not filter rules

The `mcp_server` field is a **standard frontmatter field** that defines one MCP server configuration. It does not act as a selector. The field is an object with both standard configuration fields and support for arbitrary custom fields.

**Standard configuration fields:**
- `command`: The executable to run (e.g., "npx", "python", "docker")
- `args`: Array of command-line arguments
- `env`: Map of environment variables
- `type`: Connection protocol - "stdio" (default), "http", or "sse"
- `url`: Endpoint URL (required for HTTP/SSE types)
- `headers`: Custom HTTP headers (for HTTP/SSE types)

**Example:**
```yaml
---
mcp_server:
  command: python
  args: ["-m", "server"]
  env:
    PYTHON_PATH: /usr/bin/python3
  custom_config:
    host: localhost
    port: 5432
    ssl: true
---
```

**Additional arbitrary fields:**
You can include any custom fields for your specific server needs (e.g., `custom_config`, `monitoring`, `cache_enabled`, etc.). All fields are preserved in the configuration.

**Note:** Each task or rule can specify one MCP server configuration. The format supports the standard MCP server fields plus arbitrary custom fields for flexibility.

#### `agent` (optional, standard field)

**Type:** String  
**Purpose:** Specifies the target agent and automatically filters rules with matching agent selector

The `agent` field is a **standard frontmatter field** that acts as a default selector. When a task specifies an agent, only rules with that same agent value (or no agent field) will be included in the context.

**Example:**
```yaml
---
agent: cursor
---
```

**Supported agents:** `cursor`, `copilot`, `claude`, `gemini`, `opencode`, `augment`, `windsurf`, `codex`

**Behavior:**
- Rules with `agent: cursor` are included
- Rules without an `agent` field are included (generic rules)
- Rules with different agent values (e.g., `agent: copilot`) are excluded

**Equivalent command-line usage:**
```bash
# These are equivalent:
coding-context implement-feature  # (task has agent: cursor)
coding-context -a cursor implement-feature
```

#### `model` (optional, standard field)

**Type:** String  
**Purpose:** Specifies the AI model to use; stored in frontmatter output but does not filter rules

The `model` field is a **standard frontmatter field** that provides metadata about which AI model should be used for the task. Unlike the `agent` field, the `model` field does not act as a selector and does not filter rules.

**Example:**
```yaml
---
agent: copilot
model: anthropic.claude-sonnet-4-20250514-v1-0
---
```

**Common model values:**
- `anthropic.claude-sonnet-4-20250514-v1-0`
- `gpt-4`
- `gpt-4-turbo`
- `gemini-pro`

**Note:** The model field is purely informational and appears in the task frontmatter output for the AI agent to use as configuration.

#### Custom Fields (optional)

Any additional YAML fields can be used for selector-based filtering.

**Example:**
```yaml
---
environment: production
region: us-east-1
---
```

**Usage:**
```bash
coding-context -s environment=production -s region=us-east-1 deploy
```

#### `selectors` (optional)

**Type:** Map of key-value pairs  
**Purpose:** Automatically filter rules and tasks without requiring `-s` flags on the command line

The `selectors` field allows a task to specify which rules should be included when the task is executed. This is equivalent to passing `-s` flags but is declared in the task file itself.

**Example:**
```yaml
---
selectors:
  languages: go
  stage: implementation
---
```

**Usage:**
```bash
# Automatically includes rules with languages=go AND stage=implementation
coding-context implement-feature
```

This is equivalent to:
```bash
coding-context -s languages=go -s stage=implementation implement-feature
```

**OR Logic with Arrays:**

You can specify multiple values for the same key using YAML arrays for OR logic:

```yaml
---
selectors:
  languages: [go, python, javascript]
  stage: testing
---
```

This matches rules where `(languages=go OR languages=python OR languages=javascript) AND stage=testing`.

**Combining with Command-Line Selectors:**

Selectors from the task frontmatter and command-line `-s` flags are combined (additive):

```bash
# Task frontmatter has: selectors.languages = go
# Command line adds: -s priority=high
# Result: Rules must match languages=go AND priority=high
coding-context -s priority=high implement-feature
```

**Special Selector: `rule_name`**

You can filter to specific rule files by their base filename (without extension):

```yaml
---
selectors:
  rule_name: [security-standards, go-best-practices]
---
```

This would only include the rules from `security-standards.md` and `go-best-practices.md`.

#### `expand` (optional)

**Type:** Boolean  
**Purpose:** Controls whether parameter expansion should occur in the task content. Defaults to `true` if not specified.

When set to `false`, parameter placeholders like `${variable}` are preserved as-is in the output, rather than being replaced with values from `-p` flags.

**Example (with parameter expansion disabled):**
```yaml
---
expand: false
---

Issue: ${issue_number}
Title: ${issue_title}
```

**Usage:**
```bash
# Even with -p flags, parameters won't be expanded
coding-context -p issue_number=123 -p issue_title="Bug" preserve-template
# Output will contain: ${issue_number} and ${issue_title}
```

**Use cases:**
- Passing templates to AI agents that handle their own parameter substitution
- Preserving template syntax that conflicts with the parameter expansion format
- Keeping templates intact for later processing

**Default behavior (expand: true or omitted):**
```yaml
---
# expand defaults to true
---

Issue: ${issue_number}
Title: ${issue_title}
```

```bash
coding-context -p issue_number=123 -p issue_title="Bug" normal-task
# Output will contain: Issue: 123 and Title: Bug
```

### Content Expansion

Task and command content supports three types of dynamic expansion, processed in a single pass to prevent injection attacks.

#### Parameter Expansion

Use `${parameter_name}` syntax to substitute parameter values from `-p` flags.

**Syntax:** `${parameter_name}`

**Example:**
```markdown
# Fix Bug: ${issue_key}

Issue: ${issue_key}
Description: ${description}
Severity: ${severity}
```

**Usage:**
```bash
coding-context \
  -p issue_key=BUG-123 \
  -p description="Crashes on startup" \
  -p severity=critical \
  /fix-bug
```

**Behavior:** If a parameter is not found, the placeholder remains unchanged (e.g., `${missing}` stays as `${missing}`) and a warning is logged.

#### Command Expansion

Use `` !`command` `` syntax to execute shell commands and include their output.

**Syntax:** `` !`command` ``

**Example:**
```markdown
# System Information

Current date: !`date +%Y-%m-%d`
Current user: !`whoami`
Git branch: !`git rev-parse --abbrev-ref HEAD`
```

**Output:**
```
Current date: 2025-12-11
Current user: alex
Git branch: main
```

**Behavior:** 
- Command output is included as-is (including any trailing newlines)
- If the command fails, the original syntax remains unchanged (e.g., `` !`false` `` stays as `` !`false` ``) and a warning is logged
- Commands are executed using `sh -c`

**Security Note:** Only use with trusted task files, as commands are executed with your user permissions.

#### Path Expansion

Use `@path` syntax to include the contents of a file.

**Syntax:** `@path` (delimited by whitespace; use `\ ` to escape spaces in filenames)

**Example:**
```markdown
# Current Configuration

@config.yaml

# API Documentation

@docs/api.md
```

**With spaces in filenames:**
```markdown
Content from file: @my\ file\ with\ spaces.txt
```

**Behavior:**
- File content is included verbatim
- If the file is not found, the original syntax remains unchanged (e.g., `@missing.txt` stays as `@missing.txt`) and a warning is logged
- Path can be absolute or relative to the current directory

#### Security: Single-Pass Expansion

All three expansion types are processed in a **single pass, rune-by-rune** to prevent injection attacks:

- Expanded content is **never re-processed** for further expansions
- Command output containing `${param}` will not be expanded
- File content containing `` !`command` `` will not be executed
- Parameter values containing `@path` will not be read as files

This prevents command injection where expanded content could trigger further, unintended expansions.

### File Location

Task files must be in one of these directories:
- `./.agents/tasks/`
- `./.cursor/commands/`
- `./.opencode/command/`
- `~/.agents/tasks/`

Tasks are matched by filename (without `.md` extension). The `task_name` field in frontmatter is optional and used only for metadata. For example, a file named `fix-bug.md` is matched by the command `/fix-bug`, regardless of whether it has `task_name` in its frontmatter.

## Command Files

Command files are reusable content blocks that can be referenced from task files using slash command syntax (e.g., `/command-name`). They are Markdown files with optional YAML frontmatter.

### Format

```markdown
---
<optional-frontmatter-fields>
---

# Command content in Markdown

This content will be substituted when the command is referenced.
```

### Frontmatter Fields (optional)

#### `expand` (optional)

**Type:** Boolean  
**Purpose:** Controls whether parameter expansion should occur in the command content. Defaults to `true` if not specified.

When set to `false`, parameter placeholders like `${variable}` are preserved as-is in the output.

**Example:**
```yaml
---
expand: false
---

Deploy to ${environment} with version ${version}
```

**Usage in a task:**
```yaml
---
---

/deploy-steps
```

**Command line:**
```bash
coding-context -p environment=prod -p version=1.0 my-task
# Command output will contain: ${environment} and ${version} (not expanded)
```

This is useful when commands contain template syntax that should be preserved.

### Slash Command Syntax

Commands are referenced from tasks using slash command syntax:

```markdown
# Deployment Steps

/pre-deploy

/deploy

/post-deploy
```

Commands can also receive inline parameters:

```markdown
/greet name="Alice"
/deploy env="production" version="1.2.3"
```

### File Locations

Command files must be in one of these directories:
- `./.agents/commands/`
- `./.cursor/commands/`
- `./.opencode/command/`
- `~/.agents/commands/`

Commands are matched by filename (without `.md` extension). For example, a file named `deploy.md` is matched by the slash command `/deploy`.

## Rule Files

Rule files provide reusable context snippets. They are Markdown or `.mdc` files with optional YAML frontmatter.

### Format

```markdown
---
<optional-frontmatter-fields>
---

# Rule content in Markdown

Guidelines, standards, or context for AI agents.
```

### Frontmatter Fields (optional)

All frontmatter fields are optional and used for filtering.

**Standard fields for rules:**

#### `task_names` (rule selector)

Specifies which task(s) this rule applies to. Can be a string or array. The field name is `task_names` (plural).

```yaml
---
task_names:
  - fix-bug
---
# This rule only applies to the 'fix-bug' task
```

**Multiple tasks (OR logic):**
```yaml
---
task_names:
  - fix-bug
  - implement-feature
  - refactor
---
# This rule applies to any of these three tasks
```

**Behavior:**
- When a task is run (e.g., `coding-context fix-bug`), the task name `fix-bug` is automatically added as a selector `task_name=fix-bug` (singular)
- Rules with `task_names: [ fix-bug ]` (plural) should match this selector
- Rules without `task_names` are included for all tasks (generic rules)
- **Note:** The code uses `task_names` (plural) in rule frontmatter, but sets selector as `task_name` (singular)

#### `language` or `languages` (rule selector)

Specifies which programming language(s) this rule applies to. Can be a string or array. Language values should be lowercase. The recommended format is `languages:` (plural) with an array.

**Recommended format (array):**
```yaml
---
languages:
  - go
---
# This rule only applies when languages=go is selected
```

**Multiple languages (OR logic):**
```yaml
---
languages:
  - go
  - python
  - javascript
---
# This rule applies to any of these languages
```

**Behavior:**
- Rules with `languages: [ go ]` are included when `-s languages=go` is specified (or via task `selectors.languages`)
- Rules without `languages` are included (generic rules)
- The task's `languages` field (metadata) does NOT automatically filter rules - use `selectors.languages` or `-s languages=go` instead
- Language values should be lowercase (e.g., `go`, `python`, `javascript`)
- Both `language` (singular) and `languages` (plural) are accepted in frontmatter, but `languages` (plural) with array format is recommended
- **Important:** When using selectors (`-s` flag or `selectors:` in frontmatter), use `languages` (plural) as the key

#### `agent` (rule selector)

Specifies which AI agent this rule is intended for.

```yaml
---
agent: cursor
---
# Rule specific to Cursor AI agent
```

**Behavior:**
- If task/CLI specifies `agent: cursor`, only rules with `agent: cursor` or no agent field are included
- Rules without an agent field are considered generic and always included (unless other selectors exclude them)

#### `mcp_server` (rule metadata)

Specifies an MCP server configuration that needs to be running for this rule. Does not filter rules. The field is an object with standard and arbitrary custom fields.

```yaml
---
mcp_server:
  command: python
  args: ["-m", "server"]
  env:
    PYTHON_PATH: /usr/bin/python3
  custom_config:
    host: localhost
    port: 5432
---
# Metadata indicating required MCP server
```

**Note:** This field is informational and does not affect rule selection.

#### `expand` (optional)

**Type:** Boolean  
**Purpose:** Controls whether parameter expansion should occur in the rule content. Defaults to `true` if not specified.

When set to `false`, parameter placeholders like `${variable}` are preserved as-is in the output.

**Example:**
```yaml
---
languages:
  - go
expand: false
---

Use version ${version} when building the project.
```

**Usage:**
```bash
coding-context -p version=1.2.3 my-task
# Rule output will contain: ${version} (not expanded)
```

This is useful when rules contain template syntax that should be preserved for the AI agent to process.

**Other common fields:**
```yaml
---
languages:
  - go
stage: implementation
priority: high
team: backend
agent: cursor
---
```

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`). Use `languages:` (plural) with array format.

### Supported Extensions

- `.md` - Markdown files
- `.mdc` - Markdown component files

### File Locations

Rules are discovered in many locations. See [Search Paths Reference](./search-paths) for the complete list.

## Bootstrap Scripts

Bootstrap scripts are executable files that run before their associated rule file is processed.

### Naming Convention

For a rule file named `my-rule.md`, the bootstrap script must be named `my-rule-bootstrap` (no extension).

**Example:**
- Rule: `.agents/rules/jira-context.md`
- Bootstrap: `.agents/rules/jira-context-bootstrap`

### Requirements

1. **Executable permission:** `chmod +x script-name`
2. **Same directory:** Must be in same directory as the rule file
3. **Naming:** Must match rule filename plus `-bootstrap` suffix

### Output Handling

- Bootstrap script output goes to **stderr**, not the main context
- The script's stdout is not captured
- Use stderr for logging and status messages

**Example:**
```bash
#!/bin/bash
# my-rule-bootstrap

echo "Fetching data..." >&2  # Goes to stderr
curl -s "https://api.example.com/data" > /tmp/data.json
echo "Data fetched successfully" >&2
```

### Environment Access

Bootstrap scripts can access all environment variables from the parent process.

**Example:**
```bash
#!/bin/bash

# Access environment variables
API_KEY="${JIRA_API_KEY}"
ISSUE="${JIRA_ISSUE_KEY}"

if [ -z "$API_KEY" ]; then
    echo "Error: JIRA_API_KEY not set" >&2
    exit 1
fi

# Fetch and process data
curl -s -H "Authorization: Bearer $API_KEY" \
    "https://api.example.com/issue/$ISSUE" \
    | jq -r '.fields' > /tmp/issue-data.json
```

## YAML Frontmatter Specification

### Valid Frontmatter

```yaml
---
key: value
another_key: another value
numeric_key: 123
boolean_key: true
---
```

### Limitations

**Top-level fields only:**
```yaml
# ✅ Supported
---
languages:
  - go
stage: testing
---

# ❌ Not supported (nested fields)
---
metadata:
  language: go
  stage: testing
---
```

**Selectors match top-level only:**
```bash
# Works with top-level fields
coding-context -s languages=go fix-bug

# Doesn't work with nested fields
coding-context -s metadata.language=go fix-bug  # Won't match
```

### Data Types

Frontmatter values are treated as strings for matching:

```yaml
---
priority: 1
enabled: true
languages:
  - go
---
```

```bash
# All values are matched as strings
coding-context -s priority=1 task       # Matches priority: 1
coding-context -s enabled=true task     # Matches enabled: true
coding-context -s languages=go task      # Matches languages: [ go ]
```

## Special Behaviors

### Multiple Tasks with Same `task_name`

If multiple task files have the same `task_name`, selectors determine which one is used.

**Without selectors:**
- Error: "multiple tasks found with task_name: X"

**With selectors:**
- The task matching all selectors is used
- If no task matches: "no task found"
- If multiple match: Error

### Rules Without Frontmatter

Rules without frontmatter are always included (unless resume mode is active).

```markdown
# General Standards

These standards apply to all projects.
```

This rule is included in every context assembly.

### Resume Mode Special Handling

The `-r` flag:
1. Skips all rule file output
2. Adds implicit `-s resume=true` selector

**Equivalent commands:**
```bash
# These are NOT exactly equivalent:
coding-context -r fix-bug                    # Skips rules
coding-context -s resume=true fix-bug        # Includes rules
```

## Validation

The CLI validates:
- ✅ Task files match by filename (`.md` files with matching base name)
- ✅ YAML frontmatter is well-formed (if present)
- ✅ At most one task matches the selectors

The CLI does NOT validate:
- Content format or structure
- Parameter references exist
- Bootstrap script success/failure

## See Also

- [CLI Reference](./cli) - Command-line options
- [Search Paths Reference](./search-paths) - Where files are found
- [How to Create Tasks](../how-to/create-tasks) - Practical guide
- [How to Create Rules](../how-to/create-rules) - Practical guide
