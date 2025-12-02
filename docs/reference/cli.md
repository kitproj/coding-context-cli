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
coding-context [options] <task-prompt>
```

## Description

The Coding Context CLI assembles context from rule files and task prompts, performs parameter substitution, and outputs the combined context to stdout. This output is designed to be piped to AI agents.

## Arguments

### `<task-prompt>`

**Required.** The task prompt to execute. This can be either:

1. **Free-text prompt**: Used directly as the task content
2. **Slash command**: A prompt containing `/task-name` which triggers task file lookup

**Examples:**
```bash
# Free-text prompt (used directly as task content)
coding-context "Please help me fix the login bug"

# Slash command (looks up fix-bug.md task file)
coding-context /fix-bug

# Slash command with arguments
coding-context "/fix-bug 123"
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

Load rules and tasks from a directory (remote or local). The directory is processed via go-getter, which downloads remote directories to a temporary location before processing and cleans up afterward.

**Note:** The working directory (`-C` or current directory) and home directory (`~`) are automatically added to the search paths, so you don't need to specify them explicitly.

Supports various protocols via [go-getter](https://github.com/hashicorp/go-getter):
- `git::` - Git repositories (HTTPS, SSH)
- `http://`, `https://` - HTTP/HTTPS URLs (tar.gz, zip, directories)
- `s3::` - S3 buckets
- `file://` - Local file paths (or absolute paths without prefix)

**Examples:**
```bash
# Load from Git repository
coding-context -d git::https://github.com/company/shared-rules.git /fix-bug

# Use specific branch or tag
coding-context -d 'git::https://github.com/company/shared-rules.git?ref=v1.0' /fix-bug

# Use subdirectory within repository (note the double slash)
coding-context -d 'git::https://github.com/company/mono-repo.git//standards' /fix-bug

# Load from HTTP archive
coding-context -d https://example.com/coding-rules.tar.gz /fix-bug

# Multiple remote sources
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  -d https://cdn.example.com/team-rules.zip \
  /fix-bug

# Mix local and remote
coding-context \
  -d git::https://github.com/company/org-standards.git \
  -d file:///path/to/local/rules \
  -s language=Go \
  /fix-bug

# Local directories are automatically included
# (workDir and homeDir are added automatically)
coding-context /fix-bug
```

**See also:** [How to Use Remote Directories](../how-to/use-remote-directories)

### `-p <key>=<value>`

**Type:** Key-value pair  
**Repeatable:** Yes

Define a parameter for substitution in task prompts. Variables in task files using `${key}` syntax will be replaced with the specified value.

**Examples:**
```bash
# Single parameter
coding-context -p issue_key=BUG-123 /fix-bug

# Multiple parameters
coding-context \
  -p issue_key=BUG-123 \
  -p description="Application crashes" \
  -p severity=critical \
  /fix-bug
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
coding-context -s resume=false /fix-bug | ai-agent

# Resume session
coding-context -r /fix-bug | ai-agent
```

### `-s <key>=<value>`

**Type:** Key-value pair  
**Repeatable:** Yes

Filter rules and tasks by frontmatter fields. Only rules and tasks where ALL specified selectors match will be included.

**Important:** Only top-level frontmatter fields can be matched. Nested fields are not supported.

**Examples:**
```bash
# Single selector
coding-context -s language=Go /fix-bug

# Multiple selectors (AND logic)
coding-context -s language=Go -s priority=high /fix-bug

# Select specific task variant
coding-context -s environment=production /deploy
```

### `-t`

**Type:** Boolean flag  
**Default:** False

Print the task's YAML frontmatter at the beginning of the output. This includes all frontmatter fields such as `task_name`, `selectors`, `resume`, and any custom fields.

Use this when downstream tools or AI agents need access to task metadata for decision-making or workflow automation.

**Example:**
```bash
# Emit task frontmatter with the assembled context
coding-context -t /fix-bug
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
coding-context -t /implement-feature
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
coding-context /fix-bug 2>errors.log | ai-agent
```

## Environment Variables

The CLI itself doesn't use environment variables, but bootstrap scripts can access any environment variables set in the shell.

**Example:**
```bash
export JIRA_API_KEY="your-key"
export GITHUB_TOKEN="your-token"

coding-context /fix-bug  # Bootstrap scripts can use these variables
```

## Examples

### Basic Usage

```bash
# Free-text prompt (used directly as task content)
coding-context "Please help me review this code for security issues"

# Slash command to execute a task file
coding-context /code-review

# Slash command with arguments
coding-context "/fix-bug 123"

# With parameters
coding-context -p pr_number=123 /code-review

# With selectors
coding-context -s language=Python /fix-bug

# Multiple parameters and selectors
coding-context \
  -s language=Go \
  -s stage=implementation \
  -p feature_name="Authentication" \
  /implement-feature
```

### Working Directory

```bash
# Run from different directory
coding-context -C /path/to/project /fix-bug

# Run from subdirectory
cd backend
coding-context /fix-bug  # Uses backend/.agents/ if it exists
```

### Remote Directories

```bash
# Load from Git repository
coding-context -d git::https://github.com/company/shared-rules.git /fix-bug

# Use specific version
coding-context -d 'git::https://github.com/company/rules.git?ref=v1.0.0' /fix-bug

# Combine multiple sources
coding-context \
  -d git::https://github.com/company/org-standards.git \
  -d git::https://github.com/team/project-rules.git \
  -s language=Go \
  /implement-feature

# Load from HTTP archive
coding-context -d https://cdn.company.com/rules.tar.gz /code-review
```

### Resume Mode

```bash
# Initial invocation
coding-context -s resume=false /implement-feature > context.txt
cat context.txt | ai-agent > plan.txt

# Continue work (skips rules)
coding-context -r /implement-feature | ai-agent
```

### Piping to AI Agents

```bash
# Claude
coding-context /fix-bug | claude

# LLM tool
coding-context /fix-bug | llm -m claude-3-5-sonnet-20241022

# OpenAI
coding-context /code-review | openai api completions.create -m gpt-4

# Save to file first
coding-context /fix-bug > context.txt
cat context.txt | your-ai-agent

# Free-text prompt
coding-context "Please help me debug the auth module" | claude
```

### Token Monitoring

```bash
# See token count in stderr
coding-context /fix-bug 2>&1 | grep -i token

# Separate stdout and stderr
coding-context /fix-bug 2>tokens.log | ai-agent
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
coding-context -s language=Go -s stage=testing /fix-bug
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
coding-context -s metadata.language=Go /fix-bug
```

## Slash Commands

When you provide a task-prompt containing a slash command (e.g., `/task-name arg1 "arg 2"`), the CLI will automatically:

1. Extract the task name and arguments from the slash command
2. Load the referenced task file
3. Pass the slash command arguments as parameters (`$1`, `$2`, `$ARGUMENTS`, etc.)
4. Merge with any existing parameters (slash command parameters take precedence)

This enables dynamic task execution with inline arguments.

### Slash Command Format

```
/task-name arg1 "arg with spaces" arg3
```

### Positional Parameters

Positional arguments are automatically numbered starting from 1:
- `/fix-bug 123` → `$1` = `123`
- `/task arg1 arg2 arg3` → `$1` = `arg1`, `$2` = `arg2`, `$3` = `arg3`

Quoted arguments preserve spaces:
- `/code-review "PR #42"` → `$1` = `PR #42`

### Named Parameters

Named parameters use the format `key="value"` with **mandatory double quotes**:
- `/fix-bug issue="PROJ-123"` → `$1` = `issue="PROJ-123"`, `$issue` = `PROJ-123`
- `/deploy env="production" version="1.2.3"` → `$1` = `env="production"`, `$2` = `version="1.2.3"`, `$env` = `production`, `$version` = `1.2.3`

Named parameters are counted as positional arguments (retaining their original form) while also being available by their key name:
- `/task arg1 key="value" arg2` → `$1` = `arg1`, `$2` = `key="value"`, `$3` = `arg2`, `$key` = `value`

Named parameter values can contain spaces and special characters:
- `/run message="Hello, World!"` → `$1` = `message="Hello, World!"`, `$message` = `Hello, World!`
- `/config query="x=y+z"` → `$1` = `query="x=y+z"`, `$query` = `x=y+z`

**Note:** Unquoted values (e.g., `key=value`) or single-quoted values (e.g., `key='value'`) are treated as regular positional arguments, not named parameters.

Reserved keys (`ARGUMENTS` and numeric keys like `1`, `2`, etc.) cannot be used as named parameter keys and will be ignored.

### Example with Positional Parameters

Create a task file (`implement-feature.md`):
```yaml
---
task_name: implement-feature
---
# Feature: ${1}

Description: ${2}
```

When you run:
```bash
coding-context '/implement-feature login "Add OAuth support"'
```

It will:
1. Parse the slash command `/implement-feature login "Add OAuth support"`
2. Load the `implement-feature` task
3. Substitute `${1}` with `login` and `${2}` with `Add OAuth support`

The output will be:
```
# Feature: login

Description: Add OAuth support
```

This is equivalent to manually running:
```bash
coding-context -p 1=login -p 2="Add OAuth support" /implement-feature
```

### Example with Named Parameters

Create a wrapper task (`fix-issue-wrapper.md`):
```yaml
---
task_name: fix-issue-wrapper
---
/fix-bug issue="PROJ-456" priority="high"
```

The target task (`fix-bug.md`):
```yaml
---
task_name: fix-bug
---
# Fix Bug: ${issue}

Priority: ${priority}
```

When you run:
```bash
coding-context fix-issue-wrapper
```

The output will be:
```
# Fix Bug: PROJ-456

Priority: high
```

This is equivalent to manually running:
```bash
coding-context -p issue=PROJ-456 -p priority=high fix-bug
```

## See Also

- [File Formats Reference](./file-formats) - Task and rule file specifications
- [Search Paths Reference](./search-paths) - Where files are found
- [How to Use with AI Agents](../how-to/use-with-ai-agents) - Practical examples
