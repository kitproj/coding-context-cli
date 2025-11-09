---
layout: default
title: CLI Reference
parent: Reference
nav_order: 1
---

# CLI Reference

Complete reference for the `coding-context-cli` command-line interface.

## Synopsis

```
coding-context-cli [options] <task-name>
```

## Description

The Coding Context CLI assembles context from rule files and task prompts, performs parameter substitution, and outputs the combined context to stdout. This output is designed to be piped to AI agents.

## Arguments

### `<task-name>`

**Required.** The name of the task to execute. This matches the `task_name` field in task file frontmatter, not the filename.

**Example:**
```bash
coding-context-cli fix-bug
```

## Options

### `-C <directory>`

**Type:** String  
**Default:** `.` (current directory)

Change to the specified directory before processing files.

**Example:**
```bash
coding-context-cli -C /path/to/project fix-bug
```

### `-p <key>=<value>`

**Type:** Key-value pair  
**Repeatable:** Yes

Define a parameter for substitution in task prompts. Variables in task files using `${key}` syntax will be replaced with the specified value.

**Examples:**
```bash
# Single parameter
coding-context-cli -p issue_key=BUG-123 fix-bug

# Multiple parameters
coding-context-cli \
  -p issue_key=BUG-123 \
  -p description="Application crashes" \
  -p severity=critical \
  fix-bug
```

### `-r`

**Type:** Boolean flag  
**Default:** False

Enable resume mode. This does two things:
1. Skips outputting all rule files (saves tokens)
2. Automatically adds `-s resume=true` selector

Use this when continuing work in a new session where context has already been established.

**Example:**
```bash
# Initial session
coding-context-cli -s resume=false fix-bug | ai-agent

# Resume session
coding-context-cli -r fix-bug | ai-agent
```

### `-s <key>=<value>`

**Type:** Key-value pair  
**Repeatable:** Yes

Filter rules and tasks by frontmatter fields. Only rules and tasks where ALL specified selectors match will be included.

**Important:** Only top-level frontmatter fields can be matched. Nested fields are not supported.

**Examples:**
```bash
# Single selector
coding-context-cli -s language=Go fix-bug

# Multiple selectors (AND logic)
coding-context-cli -s language=Go -s priority=high fix-bug

# Select specific task variant
coding-context-cli -s environment=production deploy
```

## Exit Codes

- `0` - Success
- Non-zero - Error occurred (check stderr for details)

## Output

### Standard Output (stdout)

The assembled context, consisting of:
1. All matching rule files
2. The selected task prompt (with parameters substituted)

This output is intended to be piped to an AI agent.

### Standard Error (stderr)

- Token count estimates
- Bootstrap script output
- Error messages
- Progress information

**Example:**
```bash
coding-context-cli fix-bug 2>errors.log | ai-agent
```

## Environment Variables

The CLI itself doesn't use environment variables, but bootstrap scripts can access any environment variables set in the shell.

**Example:**
```bash
export JIRA_API_KEY="your-key"
export GITHUB_TOKEN="your-token"

coding-context-cli fix-bug  # Bootstrap scripts can use these variables
```

## Examples

### Basic Usage

```bash
# Simple task execution
coding-context-cli code-review

# With parameters
coding-context-cli -p pr_number=123 code-review

# With selectors
coding-context-cli -s language=Python fix-bug

# Multiple parameters and selectors
coding-context-cli \
  -s language=Go \
  -s stage=implementation \
  -p feature_name="Authentication" \
  implement-feature
```

### Working Directory

```bash
# Run from different directory
coding-context-cli -C /path/to/project fix-bug

# Run from subdirectory
cd backend
coding-context-cli fix-bug  # Uses backend/.agents/ if it exists
```

### Resume Mode

```bash
# Initial invocation
coding-context-cli -s resume=false implement-feature > context.txt
cat context.txt | ai-agent > plan.txt

# Continue work (skips rules)
coding-context-cli -r implement-feature | ai-agent
```

### Piping to AI Agents

```bash
# Claude
coding-context-cli fix-bug | claude

# LLM tool
coding-context-cli fix-bug | llm -m claude-3-5-sonnet-20241022

# OpenAI
coding-context-cli code-review | openai api completions.create -m gpt-4

# Save to file first
coding-context-cli fix-bug > context.txt
cat context.txt | your-ai-agent
```

### Token Monitoring

```bash
# See token count in stderr
coding-context-cli fix-bug 2>&1 | grep -i token

# Separate stdout and stderr
coding-context-cli fix-bug 2>tokens.log | ai-agent
cat tokens.log  # View token information
```

## File Discovery

The CLI searches for files in specific locations. See [Search Paths Reference](./search-paths) for details.

## Frontmatter Matching

Selectors (`-s`) only match top-level YAML frontmatter fields.

**Works:**
```yaml
---
language: Go
stage: testing
---
```
```bash
coding-context-cli -s language=Go -s stage=testing fix-bug
```

**Doesn't Work:**
```yaml
---
metadata:
  language: Go
  stage: testing
---
```
```bash
# This WON'T match nested fields
coding-context-cli -s metadata.language=Go fix-bug
```

## See Also

- [File Formats Reference](./file-formats) - Task and rule file specifications
- [Search Paths Reference](./search-paths) - Where files are found
- [How to Use with AI Agents](../how-to/use-with-ai-agents) - Practical examples
