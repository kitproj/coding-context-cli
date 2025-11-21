---
layout: default
title: CLI Reference
parent: Reference
nav_order: 1
---

# CLI Reference

Complete reference for the `coding-context` command-line interface.

## Synopsis

```
coding-context [options] <task-name>
```

## Description

The Coding Context CLI assembles context from rule files and task prompts, performs parameter substitution, and outputs the combined context to stdout. This output is designed to be piped to AI agents.

## Arguments

### `<task-name>`

**Required.** The name of the task to execute. This matches the `task_name` field in task file frontmatter, not the filename.

**Example:**
```bash
coding-context fix-bug
```

## Options

### `-C <directory>`

**Type:** String  
**Default:** `.` (current directory)

Change to the specified directory before processing files.

**Example:**
```bash
coding-context -C /path/to/project fix-bug
```

### `-d <url>`

**Type:** String (URL or path)  
**Repeatable:** Yes

Load rules and tasks from a remote directory. The directory is downloaded to a temporary location before processing and cleaned up afterward.

Supports various protocols via [go-getter](https://github.com/hashicorp/go-getter):
- `git::` - Git repositories (HTTPS, SSH)
- `http://`, `https://` - HTTP/HTTPS URLs (tar.gz, zip, directories)
- `s3::` - S3 buckets
- `file://` - Local file paths

**Examples:**
```bash
# Load from Git repository
coding-context -d git::https://github.com/company/shared-rules.git fix-bug

# Use specific branch or tag
coding-context -d 'git::https://github.com/company/shared-rules.git?ref=v1.0' fix-bug

# Use subdirectory within repository (note the double slash)
coding-context -d 'git::https://github.com/company/mono-repo.git//standards' fix-bug

# Load from HTTP archive
coding-context -d https://example.com/coding-rules.tar.gz fix-bug

# Multiple remote sources
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  -d https://cdn.example.com/team-rules.zip \
  fix-bug

# Mix local and remote
coding-context \
  -d git::https://github.com/company/org-standards.git \
  -s language=Go \
  fix-bug
```

**See also:** [How to Use Remote Directories](../how-to/use-remote-directories)

### `-p <key>=<value>`

**Type:** Key-value pair  
**Repeatable:** Yes

Define a parameter for substitution in task prompts. Variables in task files using `${key}` syntax will be replaced with the specified value.

**Examples:**
```bash
# Single parameter
coding-context -p issue_key=BUG-123 fix-bug

# Multiple parameters
coding-context \
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
coding-context -s resume=false fix-bug | ai-agent

# Resume session
coding-context -r fix-bug | ai-agent
```

### `-s <key>=<value>`

**Type:** Key-value pair  
**Repeatable:** Yes

Filter rules and tasks by frontmatter fields. Only rules and tasks where ALL specified selectors match will be included.

**Important:** Only top-level frontmatter fields can be matched. Nested fields are not supported.

**Examples:**
```bash
# Single selector
coding-context -s language=Go fix-bug

# Multiple selectors (AND logic)
coding-context -s language=Go -s priority=high fix-bug

# Select specific task variant
coding-context -s environment=production deploy
```

### `--slash-command`

**Type:** Boolean flag  
**Default:** False

Enable slash command parsing in task content. When enabled, if the task contains a slash command (e.g., `/task-name arg1 "arg 2"`), the CLI will:
1. Extract the task name and arguments from the slash command
2. Load the referenced task instead of the original task
3. Pass the slash command arguments as parameters (`$1`, `$2`, `$ARGUMENTS`, etc.)

This enables wrapper tasks that can dynamically delegate to other tasks with arguments.

**Slash Command Format:**
```
/task-name arg1 "arg with spaces" arg3
```

**Examples:**
```bash
# Wrapper task that contains: /fix-bug 123 "critical issue"
coding-context --slash-command wrapper-task

# Equivalent to manually running:
coding-context -p 1=123 -p 2="critical issue" fix-bug
```

**Use Case Example:**

Create a wrapper task (`wrapper.md`):
```yaml
---
task_name: wrapper
---
Please execute: /implement-feature login "Add OAuth support"
```

The target task (`implement-feature.md`):
```yaml
---
task_name: implement-feature
---
# Feature: ${1}

Description: ${2}
...
```

When run with `coding-context --slash-command wrapper`, it will:
1. Parse the slash command `/implement-feature login "Add OAuth support"`
2. Load `implement-feature` task
3. Substitute `$1` with `login` and `$2` with `Add OAuth support`

### `-t`

**Type:** Boolean flag  
**Default:** False

Print the task's YAML frontmatter at the beginning of the output. This includes all frontmatter fields such as `task_name`, `selectors`, `resume`, and any custom fields.

Use this when downstream tools or AI agents need access to task metadata for decision-making or workflow automation.

**Example:**
```bash
# Emit task frontmatter with the assembled context
coding-context -t fix-bug
```

**Output:**
```yaml
---
task_name: fix-bug
resume: false
---
# Fix Bug Task
...
```

**Example with selectors:**
```bash
coding-context -t implement-feature
```

If the task includes `selectors` in frontmatter, they appear in the output:
```yaml
---
task_name: implement-feature
selectors:
  language: Go
  stage: implementation
---
# Implementation
...
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
coding-context fix-bug 2>errors.log | ai-agent
```

## Environment Variables

The CLI itself doesn't use environment variables, but bootstrap scripts can access any environment variables set in the shell.

**Example:**
```bash
export JIRA_API_KEY="your-key"
export GITHUB_TOKEN="your-token"

coding-context fix-bug  # Bootstrap scripts can use these variables
```

## Examples

### Basic Usage

```bash
# Simple task execution
coding-context code-review

# With parameters
coding-context -p pr_number=123 code-review

# With selectors
coding-context -s language=Python fix-bug

# Multiple parameters and selectors
coding-context \
  -s language=Go \
  -s stage=implementation \
  -p feature_name="Authentication" \
  implement-feature
```

### Working Directory

```bash
# Run from different directory
coding-context -C /path/to/project fix-bug

# Run from subdirectory
cd backend
coding-context fix-bug  # Uses backend/.agents/ if it exists
```

### Remote Directories

```bash
# Load from Git repository
coding-context -d git::https://github.com/company/shared-rules.git fix-bug

# Use specific version
coding-context -d 'git::https://github.com/company/rules.git?ref=v1.0.0' fix-bug

# Combine multiple sources
coding-context \
  -d git::https://github.com/company/org-standards.git \
  -d git::https://github.com/team/project-rules.git \
  -s language=Go \
  implement-feature

# Load from HTTP archive
coding-context -d https://cdn.company.com/rules.tar.gz code-review
```

### Resume Mode

```bash
# Initial invocation
coding-context -s resume=false implement-feature > context.txt
cat context.txt | ai-agent > plan.txt

# Continue work (skips rules)
coding-context -r implement-feature | ai-agent
```

### Piping to AI Agents

```bash
# Claude
coding-context fix-bug | claude

# LLM tool
coding-context fix-bug | llm -m claude-3-5-sonnet-20241022

# OpenAI
coding-context code-review | openai api completions.create -m gpt-4

# Save to file first
coding-context fix-bug > context.txt
cat context.txt | your-ai-agent
```

### Token Monitoring

```bash
# See token count in stderr
coding-context fix-bug 2>&1 | grep -i token

# Separate stdout and stderr
coding-context fix-bug 2>tokens.log | ai-agent
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
coding-context -s language=Go -s stage=testing fix-bug
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
coding-context -s metadata.language=Go fix-bug
```

## See Also

- [File Formats Reference](./file-formats) - Task and rule file specifications
- [Search Paths Reference](./search-paths) - Where files are found
- [How to Use with AI Agents](../how-to/use-with-ai-agents) - Practical examples
