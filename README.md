# Coding Agent Context CLI

A command-line interface for dynamically assembling context for AI coding agents.

This tool collects context from predefined rule files and a task-specific prompt, substitutes parameters, and prints a single, combined context to standard output. This is useful for feeding a large amount of relevant information into an AI model like Claude, Gemini, or OpenAI's GPT series.

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

You can install the CLI by downloading the latest release from the [releases page](https://github.com/kitproj/coding-context-cli/releases) or by building from source.

```bash
# Example for Linux
sudo curl -fsL -o /usr/local/bin/coding-context-cli https://github.com/kitproj/coding-context-cli/releases/download/v0.1.0/coding-context-cli_linux_amd64
sudo chmod +x /usr/local/bin/coding-context-cli
```

## Usage

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

### Example

```bash
coding-context-cli -p jira_issue_key=PROJ-1234 fix-bug | llm -m gemini-pro
```

This command will:
1. Find a task file with `task_name: fix-bug` in its frontmatter.
2. Find all rule files in the search paths.
3. Filter the rules based on selectors.
4. Execute any associated bootstrap scripts.
5. Substitute `${jira_issue_key}` with `PROJ-1234` in the task prompt.
6. Print the combined context (rules + task) to `stdout`.
7. Pipe the output to another program (in this case, `llm`).

### Example Tasks

The `<task-name>` is the value of the `task_name` field in the frontmatter of task files. Here are some common examples:

- `triage-bug`
- `review-pull-request`
- `fix-broken-build`
- `migrate-java-version`
- `enhance-docs`
- `remove-feature-flag`
- `speed-up-build`

Each of these would have a corresponding `.md` file with `task_name` in the frontmatter (e.g., a file with `task_name: triage-bug`).

## How It Works

The tool assembles the context in the following order:

1.  **Rule Files**: It searches a list of predefined locations for rule files (`.md` or `.mdc`). These locations include the current directory, ancestor directories, user's home directory, and system-wide directories.
2.  **Bootstrap Scripts**: For each rule file found (e.g., `my-rule.md`), it looks for an executable script named `my-rule-bootstrap`. If found, it runs the script before processing the rule file. These scripts are meant for bootstrapping the environment (e.g., installing tools) and their output is sent to `stderr`, not into the main context.
3.  **Filtering**: If `-s` (include) flag is used, it parses the YAML frontmatter of each rule file to decide whether to include it.
4.  **Task Prompt**: It searches for a task file with `task_name: <task-name>` in its frontmatter. The filename doesn't matter. If selectors are provided with `-s`, they are used to filter between multiple task files with the same `task_name`.
5.  **Parameter Expansion**: It substitutes variables in the task prompt using the `-p` flags.
6.  **Output**: It prints the content of all included rule files, followed by the expanded task prompt, to standard output.
7.  **Token Count**: A running total of estimated tokens is printed to standard error.

### File Search Paths

The tool looks for task and rule files in the following locations, in order of precedence:

**Tasks:**
- `./.agents/tasks/*.md` (any `.md` file with matching `task_name` in frontmatter)
- `~/.agents/tasks/*.md`
- `/etc/agents/tasks/*.md`

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

Task files are Markdown files with a required `task_name` field in the frontmatter. The filename itself doesn't matter - only the `task_name` value is used for selection. Task files can contain variables for substitution and can use selectors in frontmatter to provide different prompts for the same task.

**Example (`.agents/tasks/fix-bug.md`):**
```markdown
---
task_name: fix-bug
---
# Task: Fix Bug in ${jira_issue_key}

Here is the context for the bug. Please analyze the following files and provide a fix.
```

**Example with selectors for multiple prompts (`.agents/tasks/deploy-staging.md`):**
```markdown
---
task_name: deploy
environment: staging
---
# Deploy to Staging

Deploy the application to the staging environment with extra validation.
```

**Example for production (`.agents/tasks/deploy-prod.md`):**
```markdown
---
task_name: deploy
environment: production
---
# Deploy to Production

Deploy the application to production with all safety checks.
```

You can then select the appropriate task using:
```bash
# Deploy to staging
coding-context-cli -s environment=staging deploy

# Deploy to production
coding-context-cli -s environment=production deploy
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
