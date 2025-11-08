# Coding Agent Context CLI

A command-line interface for dynamically assembling context for AI coding agents.

This tool collects context from predefined rule files and a task-specific prompt, substitutes parameters, and prints a single, combined context to standard output. This is useful for feeding a large amount of relevant information into an AI model like Claude, Gemini, or OpenAI's GPT series.

**📖 [View Full Documentation](https://kitproj.github.io/coding-context-cli/)**

## Features

- **Dynamic Context Assembly**: Merges context from various source files.
- **Task-Specific Prompts**: Use different prompts for different tasks (e.g., `feature`, `bugfix`).
- **Rule-Based Context**: Define reusable context snippets (rules) that can be included or excluded.
- **Frontmatter Filtering**: Select rules based on metadata using frontmatter selectors (matches top-level YAML fields only).
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

## Agentic Workflows

This tool plays a crucial role in the **agentic workflow ecosystem** by providing rich, contextual information to AI agents. It complements systems like **GitHub Next's Agentic Workflows** by:

- **Context Preparation**: Assembles rules, guidelines, and task-specific prompts before agent execution
- **Workflow Integration**: Can be invoked in GitHub Actions to provide context to autonomous agents
- **Dynamic Context**: Supports runtime parameters and bootstrap scripts for real-time information
- **Multi-Stage Support**: Different context assemblies for planning, implementation, and validation stages

For a comprehensive guide on using this tool with agentic workflows, see [AGENTIC_WORKFLOWS.md](./AGENTIC_WORKFLOWS.md).

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
  -r	Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.
  -s value
    	Include rules with matching frontmatter. Can be specified multiple times as key<op>value.
    	Operators: = (equals), := (includes), != (not equals), !: (not includes)
    	Note: Only matches top-level YAML fields in frontmatter.
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
3.  **Filtering**: If `-s` (include) flag is used, it parses the YAML frontmatter of each rule file to decide whether to include it. Note that selectors can only match top-level YAML fields (e.g., `language: go`), not nested fields.
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

### Resume Mode

Resume mode is designed for continuing work on a task where you've already established context. When using the `-r` flag:

1. **Rules are skipped**: All rule files are excluded from output, saving tokens and reducing context size
2. **Resume-specific task prompts are selected**: Automatically adds `-s resume=true` selector to find task files with `resume: true` in their frontmatter

This is particularly useful in agentic workflows where an AI agent has already been primed with rules and is continuing work from a previous session.

**The `-r` flag is shorthand for:**
- Adding `-s resume=true` selector
- Skipping all rules output

**Example usage:**

```bash
# Initial task invocation (includes all rules, uses task with resume: false)
coding-context-cli -s resume=false fix-bug | ai-agent

# Resume the task (skips rules, uses task with resume: true)
coding-context-cli -r fix-bug | ai-agent
```

**Example task files for resume mode:**

Initial task (`.agents/tasks/fix-bug-initial.md`):
```markdown
---
task_name: fix-bug
resume: false
---
# Fix Bug

Analyze the issue and implement a fix.
Follow the coding standards and write tests.
```

Resume task (`.agents/tasks/fix-bug-resume.md`):
```markdown
---
task_name: fix-bug
resume: true
---
# Fix Bug - Continue

Continue working on the bug fix.
Review your previous work and complete remaining tasks.
```

With this approach, you can have multiple task prompts for the same task name, differentiated by the `resume` frontmatter field. Use `-s resume=false` to select the initial task (with rules), or `-r` to select the resume task (without rules).

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

To include this rule only when working on Go code, you would use `-s language=Go`:

```bash
coding-context-cli -s language=Go fix-bug
```

This will include all rules with `language: Go` in their frontmatter, excluding rules for other languages.

**Example: Multi-Language Rules**

You can also create rules that apply to multiple languages using YAML arrays:

```markdown
---
language:
  - TypeScript
  - JavaScript
---

# Web Development Standards

- Use ESLint for linting
- Write unit tests with Jest
- Follow modern JavaScript/TypeScript best practices
```

To include rules for TypeScript (whether specified as a scalar or in an array), use the `:=` (includes) operator:

```bash
# Include rules that apply to TypeScript
coding-context-cli -s language:=TypeScript implement-feature
```

This will match both:
- Rules with `language: TypeScript` (scalar)
- Rules with `language: [TypeScript, JavaScript]` (array containing TypeScript)

**Selector Operators:**

- `=` : Exact match (e.g., `language=Go`)
- `:=` : Includes - matches if value is in array or equals scalar (e.g., `language:=TypeScript`)
- `!=` : Not equals - excludes exact matches (e.g., `env!=staging`)
- `!:` : Not includes - excludes if value is in array or equals scalar (e.g., `language!:Python`)

For more details, see the [selector documentation](https://kitproj.github.io/coding-context-cli/how-to/use-selectors).

**Example: Language-Specific Rules**

You can create multiple language-specific rule files:

- `.agents/rules/python-standards.md` with `language: Python`
- `.agents/rules/javascript-standards.md` with `language: JavaScript`
- `.agents/rules/go-standards.md` with `language: Go`

Then select only the relevant rules:

```bash
# Work on Python code with Python-specific rules
coding-context-cli -s language=Python fix-bug

# Work on JavaScript code with JavaScript-specific rules
coding-context-cli -s language=JavaScript enhance-feature

# Exclude Python rules
coding-context-cli -s language!=Python implement-feature
```

**Common Linguist Languages**

When using language selectors, use the exact language names as defined by [GitHub Linguist](https://github.com/github-linguist/linguist/blob/master/lib/linguist/languages.yml). Here are common languages with correct capitalization:

- **C**: `C`
- **C#**: `C#`
- **C++**: `C++`
- **CSS**: `CSS`
- **Dart**: `Dart`
- **Elixir**: `Elixir`
- **Go**: `Go`
- **Haskell**: `Haskell`
- **HTML**: `HTML`
- **Java**: `Java`
- **JavaScript**: `JavaScript`
- **Kotlin**: `Kotlin`
- **Lua**: `Lua`
- **Markdown**: `Markdown`
- **Objective-C**: `Objective-C`
- **PHP**: `PHP`
- **Python**: `Python`
- **Ruby**: `Ruby`
- **Rust**: `Rust`
- **Scala**: `Scala`
- **Shell**: `Shell`
- **Swift**: `Swift`
- **TypeScript**: `TypeScript`
- **YAML**: `YAML`

Note the capitalization - for example, use `Go` not `go`, `JavaScript` not `javascript`, and `TypeScript` not `typescript`.

**Note:** Frontmatter selectors can only match top-level YAML fields. For example:
- ✅ Works: `language: Go` matches `-s language=Go`
- ❌ Doesn't work: Nested fields like `metadata.version: 1.0` cannot be matched with `-s metadata.version=1.0`

If you need to filter on nested data, flatten your frontmatter structure to use top-level fields only.

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
