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
coding-context [options] <task-name> [user-prompt]
```

## Description

The Coding Context CLI assembles context from rule files and task prompts, performs parameter substitution, and outputs the combined context to stdout. This output is designed to be piped to AI agents.

## Arguments

### `<task-name>`

**Required.** The name of a task file to look up (without `.md` extension). The task file is searched in task search paths (`.agents/tasks/`, etc.).

Task files can contain slash commands (e.g., `/command-name arg`) which reference command files for modular content reuse.

### `[user-prompt]` (optional)

**Optional.** Additional text to append to the task content. This text is appended after a delimiter (`---`) and can contain:
- Slash commands (e.g., `/command-name arg`) which will be expanded
- Parameter substitution placeholders (e.g., `${param}`)

The user-prompt is processed the same way as task file content, allowing you to dynamically extend the task at runtime.

**Examples:**
```bash
# Task name only (looks up fix-bug.md task file)
coding-context fix-bug

# Task with user-prompt
coding-context fix-bug "Focus on the authentication module"

# User-prompt with parameters
coding-context -p issue_key=BUG-123 fix-bug "Check the error logs in /var/log"

# User-prompt with slash commands
coding-context fix-bug "/pre-checks and then analyze the code"

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

**Parameter Parsing Features:**

The `-p` flag supports flexible parameter parsing with the following features:

- **Basic key-value pairs**: `key=value`
- **Multiple values per key**: Duplicate keys are collected into a list (e.g., `-p tag=frontend -p tag=backend` results in `tag` having both values)
- **Quoted values**: Use single (`'`) or double (`"`) quotes for values containing spaces or special characters
  - `-p description="Application crashes on startup"`
  - `-p name='John Doe'`
- **Escape sequences**: Supported in both quoted and unquoted values
  - Standard: `\n` (newline), `\t` (tab), `\r` (carriage return), `\\` (backslash)
  - Quotes: `\"` (double quote), `\'` (single quote)
  - Unicode: `\uXXXX` where XXXX are four hexadecimal digits
  - Hex: `\xHH` where HH are two hexadecimal digits
  - Octal: `\OOO` where OOO are up to three octal digits
- **Case-insensitive keys**: Keys are automatically converted to lowercase
- **UTF-8 support**: Full Unicode support in keys and values
- **Flexible separators**: Multiple `-p` flags can be used, or a single flag can contain comma or whitespace-separated pairs
- **Empty values**: Unquoted empty values (`key=`) result in empty parameter, quoted empty values (`key=""`) result in empty string

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

Specify the target agent being used. This is currently used for:
1. **Write Rules Mode**: With `-w` flag, determines where to write rules (e.g., `~/.github/agents/AGENTS.md` for copilot)

> **Note:** Agent-based rule filtering is not currently implemented. All rules are included regardless of the `-a` value.

**Supported agents:** `cursor`, `opencode`, `copilot`, `claude`, `gemini`, `augment`, `windsurf`, `codex`

For a complete list of all compatible agents, see [Supported Agents](./supported-agents).

**How it works:**
- The agent value is stored in the context (can come from `-a` flag or task frontmatter)
- With `-w` flag, the agent determines the user rules path for writing
- All rules are currently included regardless of agent value

**Agent Precedence:**
- If a task specifies an `agent` field in its frontmatter, that takes precedence over the `-a` flag
- The `-a` flag is used when the task doesn't specify an agent
- Either the task's agent field or `-a` flag can be used to set the agent

**Example:**
```bash
# Use with write rules mode
coding-context -a copilot -w implement-feature
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

# Resume mode with write rules: rules are skipped, only task output to stdout
coding-context -a copilot -w -r fix-bug
```

**Note on Resume Mode:**
When using `-w` with `-r` (resume mode), no rules file is written since rules are not collected in resume mode. Only the task prompt is output to stdout.

**Use case:**
This mode is particularly useful when working with AI coding agents that read rules from specific configuration files. Instead of including all rules in the prompt (consuming tokens), you can write them to the agent's config file once and only send the task prompt.

## Exit Codes

- `0` - Success
- Non-zero - Error occurred (check stderr for details)

## Output

### Standard Output (stdout)

The assembled context, consisting of:
1. All matching rule files (content only, without frontmatter)
2. Available skills metadata (XML format) - included when skills are discovered
3. The selected task content (with parameters substituted, without frontmatter)

**Note:** Task frontmatter is used for filtering and metadata but is **NOT included in the output**. Only the task content below the frontmatter delimiters appears in the final output.

Skills metadata (when present) is output as XML after rules and before the task content, listing available skills for progressive disclosure.

**Example output:**
```markdown
# Rule content here...

<available_skills>
  <skill>
    <name>data-analysis</name>
    <description>Analyze datasets and generate reports...</description>
    <location>/path/to/.agents/skills/data-analysis/SKILL.md</location>
  </skill>
</available_skills>

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
# Task name lookup
coding-context code-review

# Task with user-prompt
coding-context code-review "Focus on security vulnerabilities"

# With parameters
coding-context -p pr_number=123 code-review

# User-prompt with parameters
coding-context -p issue=BUG-456 fix-bug "Check the database connection logic"

# With selectors
coding-context -s languages=python fix-bug

# Multiple parameters and selectors
coding-context \
  -s languages=go \
  -s stage=implementation \
  -p feature_name="Authentication" \
  implement-feature

# User-prompt with slash commands
coding-context implement-feature "/pre-checks and validate the requirements"
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

# With user-prompt
coding-context fix-bug "Focus on edge cases" | claude

# LLM tool
coding-context fix-bug | llm -m claude-3-5-sonnet-20241022

# OpenAI
coding-context code-review | openai api completions.create -m gpt-4

# Save to file first
coding-context fix-bug > context.txt
cat context.txt | your-ai-agent

# User-prompt with parameters
coding-context -p issue=123 fix-bug "Check logs in /var/log" | claude
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
