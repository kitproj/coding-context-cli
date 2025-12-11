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

**Required.** The name of a task file to look up (without `.md` extension). The task file is searched in task search paths (`.agents/tasks/`, etc.).

Task files can contain slash commands (e.g., `/command-name arg`) which reference command files for modular content reuse.

**Examples:**
```bash
# Task name (looks up fix-bug.md task file)
coding-context fix-bug

# With parameters
coding-context -p issue_key=BUG-123 fix-bug

# With selectors
coding-context -s languages=go fix-bug
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
  -d file:///path/to/local/rules \
  -s languages=go \
  fix-bug

# Local directories are automatically included
# (workDir and homeDir are added automatically)
coding-context fix-bug
```

**See also:** [How to Use Remote Directories](../how-to/use-remote-directories)

### `-m <url>`

**Type:** String (URL)  
**Default:** (empty)

Load a manifest file containing search paths (one per line). The manifest file is downloaded via go-getter and each line is treated as a search path to be added to the `-d` flag list. Every line is included as-is without trimming.

**Examples:**
```bash
# Load search paths from a manifest file
coding-context -m https://example.com/manifest.txt fix-bug

# Combine manifest with additional directories
coding-context \
  -m https://example.com/manifest.txt \
  -d git::https://github.com/company/extra-rules.git \
  fix-bug
```

**Manifest file format (`manifest.txt`):**
```
git::https://github.com/company/shared-rules.git
https://cdn.example.com/coding-standards.tar.gz
file:///path/to/local/rules
```

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

### `-a <agent>`

**Type:** String  
**Default:** (empty)

Specify the target agent being used. When set, this excludes that agent's own rule paths (since the agent reads those itself) while including rules from other agents and generic rules.

**Supported agents:** `cursor`, `opencode`, `copilot`, `claude`, `gemini`, `augment`, `windsurf`, `codex`

**How it works:**
- When `-a cursor` is specified, paths like `.cursor/rules` and `.cursorrules` are excluded
- Rules from other agents (e.g., `.opencode/agent`, `.github/copilot-instructions.md`) are included
- Generic rules from `.agents/rules` are always included
- The agent name is automatically added as a selector for rule filtering

**Example:**
```bash
# Using Cursor - excludes .cursor/ paths, includes others
coding-context -a cursor fix-bug

# Using GitHub Copilot - excludes .github/copilot-instructions.md, includes others
coding-context -a copilot implement-feature
```

**Note:** Task files can override this with an `agent` field in their frontmatter.

**See also:** [Targeting a Specific Agent](../../README.md#targeting-a-specific-agent) in README

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

**Important:** 
- Only top-level frontmatter fields can be matched. Nested fields are not supported.
- For language filtering, use `-s languages=go` (plural `languages`)
- This is different from the `-p` flag, which is for parameter substitution, not filtering

**Examples:**
```bash
# Single selector
coding-context -s languages=go fix-bug

# Multiple selectors (AND logic)
coding-context -s languages=go -s priority=high fix-bug

# Select specific task variant
coding-context -s environment=production deploy
```

**Note:** When filtering by language, use `-s languages=go` (plural). The selector key is `languages` (plural), matching the frontmatter field name.

### `-a <agent>`

**Type:** String (agent name)  
**Default:** (empty)

Specify the default agent to use. This acts as a fallback if the task doesn't specify an agent in its frontmatter.

**Supported agents:**
- `cursor` - [Cursor](https://cursor.sh/)
- `opencode` - [OpenCode.ai](https://opencode.ai/)
- `copilot` - [GitHub Copilot](https://github.com/features/copilot)
- `claude` - [Anthropic Claude](https://claude.ai/)
- `gemini` - [Google Gemini](https://gemini.google.com/)
- `augment` - [Augment](https://augmentcode.com/)
- `windsurf` - [Windsurf](https://codeium.com/windsurf)
- `codex` - [Codex](https://codex.ai/)

**Agent Precedence:**
- If the task specifies an `agent` field in its frontmatter, that agent **overrides** the `-a` flag
- The `-a` flag serves as a **default** agent when the task doesn't specify one
- This allows tasks to specify their preferred agent while supporting a command-line default

**Examples:**
```bash
# Use copilot as the default agent
coding-context -a copilot fix-bug

# Task with agent field will override -a flag
# If fix-bug.md has "agent: claude", it will use claude instead of copilot
coding-context -a copilot fix-bug
```

### `-w`

**Type:** Boolean flag  
**Default:** False

Write rules mode. When enabled:
1. Rules are written to the agent's user-specific file (e.g., `~/.github/agents/AGENTS.md` for copilot)
2. Only the task prompt (with frontmatter) is output to stdout
3. Rules are not included in stdout

This is useful for separating rules from task prompts, allowing AI agents to read rules from their standard configuration files while keeping the task prompt clean.

**Requirements:**
- Requires an agent to be specified (via task's `agent` field or `-a` flag)

**Agent-specific file paths:**
- `cursor`: `~/.cursor/rules/AGENTS.md`
- `opencode`: `~/.opencode/rules/AGENTS.md`
- `copilot`: `~/.github/agents/AGENTS.md`
- `claude`: `~/.claude/CLAUDE.md`
- `gemini`: `~/.gemini/GEMINI.md`
- `augment`: `~/.augment/rules/AGENTS.md`
- `windsurf`: `~/.windsurf/rules/AGENTS.md`
- `codex`: `~/.codex/AGENTS.md`

**Examples:**
```bash
# Write rules to copilot's config, output only task to stdout
coding-context -a copilot -w fix-bug

# Task specifies agent field (agent: claude), rules written to ~/.claude/CLAUDE.md
coding-context -w fix-bug

# Combine with other options
coding-context -a copilot -w -s languages=go -p issue=123 fix-bug
```

**Use case:**
This mode is particularly useful when working with AI coding agents that read rules from specific configuration files. Instead of including all rules in the prompt (consuming tokens), you can write them to the agent's config file once and only send the task prompt.

## Exit Codes

- `0` - Success
- Non-zero - Error occurred (check stderr for details)

## Output

### Standard Output (stdout)

The assembled context, consisting of:
1. Task frontmatter (YAML format) - always included when task has frontmatter
2. All matching rule files
3. The selected task prompt (with parameters substituted)

Task frontmatter is automatically included at the beginning of the output when present. This includes all frontmatter fields such as `task_name`, `selectors`, `resume`, `language`, `agent`, and any custom fields.

**Example output:**
```yaml
---
resume: false
---
# Rule content here...

# Fix Bug Task
...
```

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
# Free-text prompt (used directly as task content)
coding-context "Please help me review this code for security issues"

# Task name lookup
coding-context code-review

# With parameters
coding-context -p pr_number=123 code-review

# With selectors
coding-context -s languages=python fix-bug

# Multiple parameters and selectors
coding-context \
  -s languages=go \
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
  -s languages=go \
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

# Free-text prompt
coding-context "Please help me debug the auth module" | claude
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
languages:
  - go
stage: testing
---
```
```bash
coding-context -s languages=go -s stage=testing fix-bug
```

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`). Use `languages:` (plural) with array format in frontmatter.

**Doesn't Work:**
```yaml
---
metadata:
  languages: go
  stage: testing
---
```
```bash
# This WON'T match nested fields
coding-context -s metadata.language=go fix-bug
```

## Slash Commands in Task Files

Task files can contain slash commands (e.g., `/command-name arg`) to reference reusable command files. This enables modular, composable task definitions.

### How Slash Commands Work

When a task file contains a slash command like `/pre-deploy` or `/greet name="Alice"`, the CLI:

1. Looks up the command file (e.g., `pre-deploy.md`) in command search paths
2. Processes any arguments passed to the slash command
3. Substitutes the slash command with the command file's content
4. Passes arguments as parameters to the command file

### Slash Command Format in Task Files

```markdown
---
---
# My Task

/command-name

/another-command arg1 "arg with spaces"

/command-with-params key="value" count="42"
```

### Positional Parameters

Positional arguments are automatically numbered starting from 1:
- `/greet Alice` → `${1}` = `Alice` (in command file)
- `/deploy staging 1.2.3` → `${1}` = `staging`, `${2}` = `1.2.3`

Quoted arguments preserve spaces:
- `/notify "Build failed"` → `${1}` = `Build failed`

The special parameter `${ARGUMENTS}` contains all arguments as a space-separated string.

### Named Parameters

Named parameters use the format `key="value"` with **mandatory double quotes**:
- `/deploy env="production"` → `${env}` = `production`, `${1}` = `env="production"`
- `/notify message="Hello, World!"` → `${message}` = `Hello, World!`, `${1}` = `message="Hello, World!"`

Named parameters are also available as positional parameters (retaining their original form):
- `/task arg1 key="value" arg2` → `${1}` = `arg1`, `${2}` = `key="value"`, `${3}` = `arg2`, `${key}` = `value`

**Note:** Unquoted values (e.g., `key=value`) or single-quoted values (e.g., `key='value'`) are treated as regular positional arguments, not named parameters.

### Example with Positional Parameters

Create a command file (`.agents/commands/greet.md`):
```markdown
---
# greet command
---
Hello, ${1}! Welcome to the project.
```

Use it in a task file (`.agents/tasks/welcome.md`):
```markdown
---
---
# Welcome Task

/greet Alice

/greet Bob
```

When you run:
```bash
coding-context welcome
```

The output will include:
```
Hello, Alice! Welcome to the project.

Hello, Bob! Welcome to the project.
```

### Example with Named Parameters

Create a command file (`.agents/commands/deploy-step.md`):
```markdown
---
# deploy-step command
---
## Deploy to ${env}

Version: ${version}
Environment: ${env}
```

Use it in a task file (`.agents/tasks/deploy.md`):
```markdown
---
---
# Deployment Task

/deploy-step env="staging" version="1.2.3"

/deploy-step env="production" version="1.2.3"
```

When you run:
```bash
coding-context deploy
```

The output will include:
```
## Deploy to staging

Version: 1.2.3
Environment: staging

## Deploy to production

Version: 1.2.3
Environment: production
```

## See Also

- [File Formats Reference](./file-formats) - Task and rule file specifications
- [Search Paths Reference](./search-paths) - Where files are found
- [How to Use with AI Agents](../how-to/use-with-ai-agents) - Practical examples
