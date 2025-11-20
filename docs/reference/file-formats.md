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
task_name: <required-task-identifier>
<optional-frontmatter-fields>
---

# Task content in Markdown

Content can include ${parameter_placeholders}.
```

### Frontmatter Fields

#### `task_name` (required)

**Type:** String  
**Purpose:** Identifies the task for selection via command line

**Example:**
```yaml
---
task_name: fix-bug
---
```

**Usage:**
```bash
coding-context fix-bug
```

#### Custom Fields (optional)

Any additional YAML fields can be used for selector-based filtering.

**Example:**
```yaml
---
task_name: deploy
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
task_name: implement-feature
selectors:
  language: Go
  stage: implementation
---
```

**Usage:**
```bash
# Automatically includes rules with language=Go AND stage=implementation
coding-context implement-feature
```

This is equivalent to:
```bash
coding-context -s language=Go -s stage=implementation implement-feature
```

**OR Logic with Arrays:**

You can specify multiple values for the same key using YAML arrays for OR logic:

```yaml
---
task_name: test-code
selectors:
  language: [Go, Python, JavaScript]
  stage: testing
---
```

This matches rules where `(language=Go OR language=Python OR language=JavaScript) AND stage=testing`.

**Combining with Command-Line Selectors:**

Selectors from the task frontmatter and command-line `-s` flags are combined (additive):

```bash
# Task frontmatter has: selectors.language = Go
# Command line adds: -s priority=high
# Result: Rules must match language=Go AND priority=high
coding-context -s priority=high implement-feature
```

**Special Selector: `rule_name`**

You can filter to specific rule files by their base filename (without extension):

```yaml
---
task_name: my-task
selectors:
  rule_name: [security-standards, go-best-practices]
---
```

This would only include the rules from `security-standards.md` and `go-best-practices.md`.

### Parameter Substitution

Use `${parameter_name}` syntax for dynamic values.

**Example:**
```markdown
---
task_name: fix-bug
---
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
  fix-bug
```

### File Location

Task files must be in one of these directories:
- `./.agents/tasks/`
- `./.cursor/commands/`
- `./.opencode/command/`
- `~/.agents/tasks/`

The filename itself doesn't matter if `task_name` is specified in frontmatter. If `task_name` is not specified, the filename (without `.md` extension) is used as the task name.

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

**Common fields:**
```yaml
---
language: Go
stage: implementation
priority: high
team: backend
source: jira
---
```

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
language: Go
stage: testing
---

# ❌ Not supported (nested fields)
---
metadata:
  language: Go
  stage: testing
---
```

**Selectors match top-level only:**
```bash
# Works with top-level fields
coding-context -s language=Go fix-bug

# Doesn't work with nested fields
coding-context -s metadata.language=Go fix-bug  # Won't match
```

### Data Types

Frontmatter values are treated as strings for matching:

```yaml
---
priority: 1
enabled: true
language: Go
---
```

```bash
# All values are matched as strings
coding-context -s priority=1 task       # Matches priority: 1
coding-context -s enabled=true task     # Matches enabled: true
coding-context -s language=Go task      # Matches language: Go
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
- ✅ Task files have `task_name` in frontmatter
- ✅ YAML frontmatter is well-formed
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
