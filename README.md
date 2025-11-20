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

The tool automatically discovers and includes rules from these locations in your project, parent directories, and user home directory (`~`).

## Agentic Workflows

This tool plays a crucial role in the **agentic workflow ecosystem** by providing rich, contextual information to AI agents. It complements systems like **GitHub Next's Agentic Workflows** by:

- **Context Preparation**: Assembles rules, guidelines, and task-specific prompts before agent execution
- **Workflow Integration**: Can be invoked in GitHub Actions to provide context to autonomous agents
- **Dynamic Context**: Supports runtime parameters and bootstrap scripts for real-time information
- **Multi-Stage Support**: Different context assemblies for planning, implementation, and validation stages

For a comprehensive guide on using this tool with agentic workflows, see [AGENTIC_WORKFLOWS.md](./AGENTIC_WORKFLOWS.md).

## Installation

You can install the CLI by downloading the latest release from the [releases page](https://github.com/kitproj/coding-context-cli/releases) or by building from source.

### Linux

**AMD64:**
```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_linux_amd64
sudo chmod +x /usr/local/bin/coding-context
```

**ARM64:**
```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_linux_arm64
sudo chmod +x /usr/local/bin/coding-context
```

### MacOS

**Intel (AMD64):**
```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_darwin_amd64
sudo chmod +x /usr/local/bin/coding-context
```

**Apple Silicon (ARM64):**
```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_darwin_arm64
sudo chmod +x /usr/local/bin/coding-context
```

## Usage

```
Usage:
  coding-context-cli [options] <task-name>

Options:
  -C string
    	Change to directory before doing anything. (default ".")
  -d value
    	Remote directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, etc.).
  -p value
    	Parameter to substitute in the prompt. Can be specified multiple times as key=value.
  -r	Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.
  -s value
    	Include rules with matching frontmatter. Can be specified multiple times as key=value.
    	Note: Only matches top-level YAML fields in frontmatter.
  -t	Print task frontmatter at the beginning of output.
```

### Examples

**Basic usage with local files:**
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

**Using remote directories:**
```bash
coding-context-cli \
  -d git::https://github.com/company/shared-rules.git \
  -d s3::https://s3.amazonaws.com/my-bucket/coding-standards \
  fix-bug | llm -m gemini-pro
```

This command will:
1. Download remote directories using go-getter
2. Search for rules and tasks in the downloaded directories
3. Combine them with local rules and tasks
4. Apply the same processing as with local files

The `-d` flag supports various protocols via go-getter:
- `http://` and `https://` - HTTP/HTTPS URLs
- `git::` - Git repositories  
- `s3::` - S3 buckets
- `file://` - Local file paths
- And more (see go-getter documentation)

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
2.  **Rule Bootstrap Scripts**: For each rule file found (e.g., `my-rule.md`), it looks for an executable script named `my-rule-bootstrap`. If found, it runs the script before processing the rule file. These scripts are meant for bootstrapping the environment (e.g., installing tools) and their output is sent to `stderr`, not into the main context.
3.  **Filtering**: If `-s` (include) flag is used, it parses the YAML frontmatter of each rule file to decide whether to include it. Note that selectors can only match top-level YAML fields (e.g., `language: go`), not nested fields.
4.  **Task Prompt**: It searches for a task file with `task_name: <task-name>` in its frontmatter. The filename doesn't matter. If selectors are provided with `-s`, they are used to filter between multiple task files with the same `task_name`.
5.  **Task Bootstrap Script**: For the task file found (e.g., `fix-bug.md`), it looks for an executable script named `fix-bug-bootstrap`. If found, it runs the script before processing the task file. This allows task-specific environment setup or data preparation.
6.  **Parameter Expansion**: It substitutes variables in the task prompt using the `-p` flags.
7.  **Output**: It prints the content of all included rule files, followed by the expanded task prompt, to standard output.
8.  **Token Count**: A running total of estimated tokens is printed to standard error.

### File Search Paths

The tool looks for task and rule files in the following locations, in order of precedence:

**Tasks:**
- `./.agents/tasks/*.md` (any `.md` file with matching `task_name` in frontmatter)
- `~/.agents/tasks/*.md`

**Rules:**
The tool searches for a variety of files and directories, including:
- `CLAUDE.local.md`
- `.agents/rules`, `.cursor/rules`, `.augment/rules`, `.windsurf/rules`, `.opencode/agent`, `.opencode/command`
- `.github/copilot-instructions.md`, `.gemini/styleguide.md`
- `AGENTS.md`, `CLAUDE.md`, `GEMINI.md` (and in parent directories)
- User-specific rules in `~/.agents/rules`, `~/.claude/CLAUDE.md`, `~/.opencode/rules`, etc.

### Remote File System Support

The tool supports loading rules and tasks from remote locations via HTTP/HTTPS URLs. This enables:

- **Shared team guidelines**: Host coding standards on a central server
- **Organization-wide rules**: Distribute common rules across multiple projects
- **Version-controlled context**: Serve rules from Git repositories
- **Dynamic rules**: Update shared rules without modifying individual repositories

**Usage:**

```bash
# Clone a Git repository containing rules
coding-context-cli -d git::https://github.com/company/shared-rules.git fix-bug

# Use multiple remote sources
coding-context-cli \
  -d git::https://github.com/company/shared-rules.git \
  -d https://cdn.company.com/coding-standards \
  deploy

# Mix local and remote directories
coding-context-cli \
  -d git::https://github.com/company/shared-rules.git \
  -s language=Go \
  implement-feature
```

**Supported protocols (via go-getter):**
- `http://` and `https://` - HTTP/HTTPS URLs (downloads tar.gz, zip, or directories)
- `git::` - Git repositories (e.g., `git::https://github.com/user/repo.git`)
- `s3::` - S3 buckets (e.g., `s3::https://s3.amazonaws.com/bucket/path`)
- `file://` - Local file paths
- And more - see [go-getter documentation](https://github.com/hashicorp/go-getter)

**Important notes:**
- Remote directories are downloaded to a temporary location
- Bootstrap scripts work in downloaded directories
- Downloaded directories are cleaned up after execution
- Supports all standard directory structures (`.agents/rules`, `.agents/tasks`, etc.)

**Example: Using a Git repository:**

```bash
# Use a specific branch or tag
coding-context-cli \
  -d 'git::https://github.com/company/shared-rules.git?ref=v1.0' \
  fix-bug

# Use a subdirectory within the repo
coding-context-cli \
  -d 'git::https://github.com/company/mono-repo.git//coding-standards' \
  implement-feature
```

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

#### Task Frontmatter Selectors

Task files can include a `selectors` field in their frontmatter to automatically filter rules without requiring explicit `-s` flags on the command line. This is useful for tasks that always need specific rules.

**Example (`.agents/tasks/implement-go-feature.md`):**
```markdown
---
task_name: implement-feature
selectors:
  language: Go
  stage: implementation
---
# Implement Feature

Implement the feature following Go best practices and implementation guidelines.
```

When you run this task, it automatically applies the selectors:
```bash
# This command automatically includes only rules with language=Go and stage=implementation
coding-context-cli implement-feature
```

This is equivalent to:
```bash
coding-context-cli -s language=Go -s stage=implementation implement-feature
```

**Selectors support OR logic for the same key using arrays:**
```markdown
---
task_name: test-code
selectors:
  language: [Go, Python]
  stage: testing
---
```

This will include rules that match `(language=Go OR language=Python) AND stage=testing`.

**Combining task selectors with command-line selectors:**

Selectors from both the task frontmatter and command line are combined (additive):
```bash
# Task has: selectors.language = Go
# Command adds: -s priority=high
# Result: includes rules matching language=Go AND priority=high
coding-context-cli -s priority=high implement-feature
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

A bootstrap script is an executable file that has the same name as a rule or task file but with a `-bootstrap` suffix. These scripts are used to prepare the environment, for example by installing necessary tools. The output of these scripts is sent to `stderr` and is not part of the AI context.

**Examples:**
- Rule file: `.agents/rules/jira.md`
- Rule bootstrap script: `.agents/rules/jira-bootstrap`
- Task file: `.agents/tasks/fix-bug.md`
- Task bootstrap script: `.agents/tasks/fix-bug-bootstrap`

Bootstrap scripts are executed in the following order:
1. Rule bootstrap scripts run before their corresponding rule files are processed
2. Task bootstrap scripts run after all rules are processed but before the task content is emitted

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

**`.agents/tasks/fix-bug-bootstrap`:**
```bash
#!/bin/bash
# This script fetches the latest issue details before the task runs.
echo "Fetching issue information..." >&2
# Fetch and prepare issue data
```

### Emitting Task Frontmatter

The `-t` flag allows you to include the task's YAML frontmatter at the beginning of the output. This is useful when the AI agent or downstream tool needs access to metadata about the task being executed.

**Example usage:**
```bash
coding-context-cli -t -p issue_number=123 fix-bug
```

**Output format:**
```yaml
---
task_name: fix-bug
resume: false
---
# Fix Bug Task

Fix the bug in issue #123...
```

This can be useful for:
- **Agent decision making**: The AI can see metadata like priority, environment, or stage
- **Workflow automation**: Downstream tools can parse the frontmatter to make decisions
- **Debugging**: You can verify which task variant was selected and what selectors were applied

**Example with selectors in frontmatter:**
```bash
coding-context-cli -t implement-feature
```

If the task has `selectors` in its frontmatter, they will be visible in the output:
```yaml
---
task_name: implement-feature
selectors:
  language: Go
  stage: implementation
---
# Implementation Task
...
```
