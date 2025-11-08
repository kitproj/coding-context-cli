---
layout: default
title: Usage Guide
---

# Usage Guide

This guide provides detailed information on using the Coding Context CLI.

## Command-Line Options

```
Usage:
  coding-context-cli [options] <task-name>

Options:
  -C string
    	Change to directory before doing anything. (default ".")
  -p value
    	Parameter to substitute in the prompt. Can be specified multiple times as key=value.
  -r	Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.
  -s value
    	Include rules with matching frontmatter. Can be specified multiple times as key=value.
    	Note: Only matches top-level YAML fields in frontmatter.
```

## Task Files

Task files are Markdown files with a required `task_name` field in the frontmatter. The filename itself doesn't matter - only the `task_name` value is used for selection.

### Basic Task Example

```markdown
---
task_name: fix-bug
---
# Task: Fix Bug

Analyze the issue and implement a fix.
Follow the coding standards and write tests.
```

### Task with Parameters

Task files can contain variables for substitution using the `${variable_name}` syntax:

```markdown
---
task_name: fix-bug
---
# Task: Fix Bug in ${jira_issue_key}

## Issue Details
- Issue: ${jira_issue_key}
- Title: ${issue_title}
- Description: ${issue_description}

Please analyze and fix this bug.
```

Use with:
```bash
coding-context-cli \
  -p jira_issue_key=PROJ-1234 \
  -p issue_title="App crashes" \
  -p issue_description="Application crashes on startup" \
  fix-bug
```

### Multiple Tasks with Selectors

You can create multiple task files with the same `task_name` but different selectors:

**`.agents/tasks/deploy-staging.md`:**
```markdown
---
task_name: deploy
environment: staging
---
# Deploy to Staging

Deploy with extra validation and monitoring.
```

**`.agents/tasks/deploy-prod.md`:**
```markdown
---
task_name: deploy
environment: production
---
# Deploy to Production

Deploy with all safety checks and rollback plan.
```

Use with:
```bash
# Deploy to staging
coding-context-cli -s environment=staging deploy

# Deploy to production
coding-context-cli -s environment=production deploy
```

## Rule Files

Rule files are Markdown (`.md`) or `.mdc` files that contain reusable context snippets. They can have optional YAML frontmatter for filtering.

### Basic Rule Example

```markdown
# Backend Coding Standards

- All new code must be accompanied by unit tests
- Use the standard logging library
- Follow REST API conventions
```

### Rule with Frontmatter

```markdown
---
language: Go
priority: high
---

# Go Backend Standards

- Use `gofmt` for formatting
- Handle all errors explicitly
- Write table-driven tests
- Use meaningful package names
```

### Filtering Rules with Selectors

Use the `-s` flag to filter rules based on frontmatter:

```bash
# Include only Go rules
coding-context-cli -s language=Go fix-bug

# Include only high-priority rules
coding-context-cli -s priority=high code-review

# Multiple selectors (AND logic)
coding-context-cli -s language=Go -s priority=high implement-feature
```

**Important:** Frontmatter selectors only match top-level YAML fields. Nested fields are not supported.

### Language-Specific Rules

Create separate rule files for each language:

**`.agents/rules/python-standards.md`:**
```markdown
---
language: Python
---

# Python Coding Standards
- Follow PEP 8
- Use type hints
- Write docstrings
```

**`.agents/rules/go-standards.md`:**
```markdown
---
language: Go
---

# Go Coding Standards
- Use gofmt
- Handle errors
- Write tests
```

Then select the appropriate rules:
```bash
# Python project
coding-context-cli -s language=Python fix-bug

# Go project
coding-context-cli -s language=Go fix-bug
```

### Common Linguist Languages

When using language selectors, use the exact language names as defined by [GitHub Linguist](https://github.com/github-linguist/linguist/blob/master/lib/linguist/languages.yml):

- **C**, **C#**, **C++**, **CSS**
- **Dart**, **Elixir**, **Go**
- **Haskell**, **HTML**
- **Java**, **JavaScript**
- **Kotlin**, **Lua**
- **Markdown**, **Objective-C**
- **PHP**, **Python**
- **Ruby**, **Rust**
- **Scala**, **Shell**, **Swift**
- **TypeScript**, **YAML**

Note: Use exact capitalization (e.g., `Go` not `go`, `JavaScript` not `javascript`).

## Resume Mode

Resume mode is designed for continuing work on a task where context has already been established. When using the `-r` flag:

1. **Rules are skipped**: All rule files are excluded from output
2. **Resume-specific tasks are selected**: Automatically adds `-s resume=true` selector

This saves tokens and reduces context size when an AI agent is continuing work from a previous session.

### Resume Mode Example

**Initial task (`.agents/tasks/fix-bug-initial.md`):**
```markdown
---
task_name: fix-bug
resume: false
---
# Fix Bug

Analyze the issue and implement a fix.
Follow the coding standards and write tests.
```

**Resume task (`.agents/tasks/fix-bug-resume.md`):**
```markdown
---
task_name: fix-bug
resume: true
---
# Fix Bug - Continue

Continue working on the bug fix.
Review your previous work and complete remaining tasks.
```

**Usage:**
```bash
# Initial invocation (includes all rules)
coding-context-cli -s resume=false fix-bug | ai-agent

# Resume the task (skips rules)
coding-context-cli -r fix-bug | ai-agent
```

## Bootstrap Scripts

Bootstrap scripts are executable files that run before their corresponding rule files are processed. They're used to prepare the environment or fetch dynamic context.

### Bootstrap Script Example

**Rule file:** `.agents/rules/jira.md`

**Bootstrap script:** `.agents/rules/jira-bootstrap`

```bash
#!/bin/bash
# This script runs before jira.md is processed

# Install jira-cli if not present
if ! command -v jira-cli &> /dev/null; then
    echo "Installing jira-cli..." >&2
    # Installation commands here
fi

# Fetch latest issue data
jira-cli get-issue ${JIRA_ISSUE} > /tmp/jira-context.txt
```

**Important Notes:**
- Bootstrap scripts must be executable: `chmod +x script-name`
- Output goes to `stderr`, not the main context
- Script name must match rule file name with `-bootstrap` suffix

## File Search Paths

The tool searches for task and rule files in these locations, in order of precedence:

### Task File Locations

1. `./.agents/tasks/*.md` (any `.md` file with matching `task_name` in frontmatter)
2. `~/.agents/tasks/*.md`
3. `/etc/agents/tasks/*.md`

### Rule File Locations

The tool searches for various configuration formats:

**Project-specific:**
- `CLAUDE.local.md`
- `.agents/rules/*`
- `.cursor/rules/*`
- `.augment/rules/*`
- `.windsurf/rules/*`
- `.opencode/agent/*`, `.opencode/command/*`, `.opencode/rules/*`
- `.github/copilot-instructions.md`, `.github/agents/*`
- `.gemini/styleguide.md`
- `AGENTS.md`, `CLAUDE.md`, `GEMINI.md` (and in parent directories)

**User-specific:**
- `~/.agents/rules/*`
- `~/.claude/CLAUDE.md`
- `~/.opencode/rules/*`

**System-wide:**
- `/etc/agents/rules/*`
- `/etc/opencode/rules/*`

## How It Works

The tool assembles context in the following order:

1. **Rule Files**: Searches predefined locations for rule files (`.md` or `.mdc`)
2. **Bootstrap Scripts**: For each rule file (e.g., `my-rule.md`), looks for `my-rule-bootstrap` and runs it if found
3. **Filtering**: If `-s` flags are used, parses YAML frontmatter to decide whether to include each rule
4. **Task Prompt**: Searches for a task file with `task_name: <task-name>` in frontmatter
5. **Parameter Expansion**: Substitutes variables in the task prompt using `-p` flags
6. **Output**: Prints all included rule files, followed by the expanded task prompt, to stdout
7. **Token Count**: Prints running total of estimated tokens to stderr

## Examples

### Basic Usage

```bash
# Simple task without parameters
coding-context-cli code-review

# Task with parameters
coding-context-cli -p pr_number=123 code-review

# Task with selectors
coding-context-cli -s language=Python fix-bug

# Multiple parameters and selectors
coding-context-cli \
  -s language=Go \
  -s priority=high \
  -p issue_key=PROJ-1234 \
  -p severity=critical \
  fix-bug
```

### Working Directory

```bash
# Run from a different directory
coding-context-cli -C /path/to/project fix-bug

# Run from project subdirectory
cd /path/to/project/backend
coding-context-cli fix-bug
```

### Piping to AI Models

```bash
# Pipe to Claude
coding-context-cli fix-bug | claude

# Pipe to OpenAI
coding-context-cli fix-bug | openai

# Pipe to local LLM
coding-context-cli fix-bug | llm -m gemini-pro

# Pipe to custom script
coding-context-cli fix-bug | ./my-ai-agent.sh
```

### Resume Workflow

```bash
# Initial work session
coding-context-cli -s resume=false implement-feature > context.txt
cat context.txt | ai-agent > implementation-plan.txt

# Continue in new session
coding-context-cli -r implement-feature | ai-agent
```

## Environment Variables

Bootstrap scripts can access environment variables:

```bash
# Set environment variables for bootstrap scripts
export JIRA_API_KEY="your-api-key"
export GITHUB_TOKEN="your-token"

# Run with environment variables available to bootstrap scripts
coding-context-cli fix-bug
```

## Best Practices

1. **Version control your rules**: Keep `.agents/` directory in git
2. **Use selectors effectively**: Target specific contexts for different scenarios
3. **Parameterize tasks**: Make tasks reusable across different instances
4. **Organize by concern**: Structure rules by workflow stage (planning, implementation, testing)
5. **Test context assembly**: Verify context before using in production workflows
6. **Monitor token usage**: Keep context size manageable for AI models
7. **Document your tasks**: Include clear instructions in task prompts
8. **Use bootstrap for dynamic data**: Fetch real-time information when needed

## Troubleshooting

### Task not found

- Verify `task_name` field exists in frontmatter
- Check task file is in a search path
- Try running from correct directory with `-C`

### Rules not included

- Check frontmatter selectors match exactly
- Remember selectors only match top-level fields
- If no selectors specified, all rules are included

### Bootstrap script errors

- Ensure script is executable: `chmod +x script`
- Check script output in stderr
- Verify script name matches pattern: `rule-name-bootstrap`

### Token count too high

- Use selectors to reduce included rules
- Use resume mode (`-r`) to skip rules
- Split large rules into smaller, targeted ones
- Consider language-specific rules

## Next Steps

- Explore [Examples](./examples) for real-world use cases
- Learn about [Agentic Workflows](./agentic-workflows) integration
- Check the [GitHub repository](https://github.com/kitproj/coding-context-cli) for updates
