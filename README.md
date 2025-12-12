# Coding Agent Context CLI

A command-line interface for dynamically assembling context for AI coding agents.

This tool collects context from predefined rule files and a task-specific prompt, substitutes parameters, and prints a single, combined context to standard output. This is useful for feeding a large amount of relevant information into an AI model like Claude, Gemini, or OpenAI's GPT series.

**📖 [View Full Documentation](https://kitproj.github.io/coding-context-cli/)**  
**📊 [View Slide Deck](./SLIDES.md)** | [Download PDF](./SLIDES.pdf) | [How to Present](./SLIDES_README.md)

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
- **Generic AI Agents**: `AGENTS.md`, `.agents/rules`, `.agents/commands` (tasks), `.agents/tasks`

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
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.23/coding-context_v0.0.23_linux_amd64
sudo chmod +x /usr/local/bin/coding-context
```

**ARM64:**
```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.23/coding-context_v0.0.23_linux_arm64
sudo chmod +x /usr/local/bin/coding-context
```

### MacOS

**Intel (AMD64):**
```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.23/coding-context_v0.0.23_darwin_amd64
sudo chmod +x /usr/local/bin/coding-context
```

**Apple Silicon (ARM64):**
```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-context-cli/releases/download/v0.0.23/coding-context_v0.0.23_darwin_arm64
sudo chmod +x /usr/local/bin/coding-context
```

## Usage

```
Usage:
  coding-context [options] <task-name>

Options:
  -C string
    	Change to directory before doing anything. (default ".")
  -d value
    	Remote directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, etc.).
  -m string
    	Go Getter URL to a manifest file containing search paths (one per line). Every line is included as-is.
  -p value
    	Parameter to substitute in the prompt. Can be specified multiple times as key=value.
  -r	Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.
  -s value
    	Include rules with matching frontmatter. Can be specified multiple times as key=value.
    	Note: Only matches top-level YAML fields in frontmatter.
  -a string
    	Default agent to use if task doesn't specify one. Excludes that agent's own rule paths (since the agent reads those itself). Supported agents: cursor, opencode, copilot, claude, gemini, augment, windsurf, codex.
  -w	Write rules to agent's config file and output only task to stdout. Requires agent (via task or -a flag).
```

### Examples

**Basic usage with local files:**
```bash
coding-context -p jira_issue_key=PROJ-1234 fix-bug | llm -m gemini-pro
```

This command will:
1. Find a task file named `fix-bug.md` in the task search paths.
2. Find all rule files in the search paths.
3. Filter the rules based on selectors.
4. Execute any associated bootstrap scripts.
5. Substitute `${jira_issue_key}` with `PROJ-1234` in the task prompt.
6. Print the combined context (rules + task) to `stdout`.
7. Pipe the output to another program (in this case, `llm`).

**Using remote directories:**
```bash
coding-context \
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

### Content Expansion Features

Task and rule content supports three types of dynamic expansion:

1. **Parameter Expansion**: Use `${parameter_name}` syntax to substitute parameter values from `-p` flags
   ```markdown
   Issue: ${issue_key}
   Description: ${description}
   ```

2. **Command Expansion**: Use `` !`command` `` syntax to execute shell commands and include their output
   ```markdown
   Current date: !`date +%Y-%m-%d`
   Git branch: !`git rev-parse --abbrev-ref HEAD`
   ```

3. **Path Expansion**: Use `@path` syntax to include file contents
   ```markdown
   Current configuration:
   @config.yaml
   ```

**Security Note:** All expansions are processed in a single pass to prevent injection attacks. Expanded content is never re-processed.

### Example Tasks

The `<task-name>` corresponds to the filename (without the `.md` extension) of task files. Here are some common examples:

- `triage-bug`
- `review-pull-request`
- `fix-broken-build`
- `migrate-java-version`
- `enhance-docs`
- `remove-feature-flag`
- `speed-up-build`

Each of these would have a corresponding `.md` file (e.g., `triage-bug.md`, `fix-broken-build.md`).

## How It Works

The tool assembles the context in the following order:

1.  **Rule Files**: It searches a list of predefined locations for rule files (`.md` or `.mdc`). These locations include the current directory, ancestor directories, user's home directory, and system-wide directories.
2.  **Rule Bootstrap Scripts**: For each rule file found (e.g., `my-rule.md`), it looks for an executable script named `my-rule-bootstrap`. If found, it runs the script before processing the rule file. These scripts are meant for bootstrapping the environment (e.g., installing tools) and their output is sent to `stderr`, not into the main context.
3.  **Filtering**: If `-s` (include) flag is used, it parses the YAML frontmatter of each rule file to decide whether to include it. Note that selectors can only match top-level YAML fields (e.g., `language: go`), not nested fields.
4.  **Task Prompt**: It searches for a task file matching the filename (without `.md` extension). Tasks are matched by filename, not by `task_name` in frontmatter. If selectors are provided with `-s`, they are used to filter between multiple task files with the same filename.
5.  **Task Bootstrap Script**: For the task file found (e.g., `fix-bug.md`), it looks for an executable script named `fix-bug-bootstrap`. If found, it runs the script before processing the task file. This allows task-specific environment setup or data preparation.
6.  **Parameter Expansion**: It substitutes variables in the task prompt using the `-p` flags.
7.  **Output**: It prints the content of all included rule files, followed by the expanded task prompt, to standard output.
8.  **Token Count**: A running total of estimated tokens is printed to standard error.

### File Search Paths

The tool looks for task and rule files in the following locations, in order of precedence:

**Tasks:**
- `./.agents/tasks/*.md` (task name matches filename without `.md` extension)
- `~/.agents/tasks/*.md`

**Commands** (reusable content blocks referenced via slash commands like `/command-name` inside task content):
- `./.agents/commands/*.md`
- `./.cursor/commands/*.md`
- `./.opencode/command/*.md`

**Rules:**
The tool searches for a variety of files and directories, including:
- `CLAUDE.local.md`
- `.agents/rules`, `.cursor/rules`, `.augment/rules`, `.windsurf/rules`, `.opencode/agent`, `.opencode/rules`
- `.github/copilot-instructions.md`, `.github/agents`, `.gemini/styleguide.md`
- `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, `.codex/AGENTS.md`
- User-specific rules in `~/.agents/rules`, `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, `~/.gemini/GEMINI.md`, `~/.opencode/rules`, etc.

### Remote File System Support

The tool supports loading rules and tasks from remote locations via HTTP/HTTPS URLs. This enables:

- **Shared team guidelines**: Host coding standards on a central server
- **Organization-wide rules**: Distribute common rules across multiple projects
- **Version-controlled context**: Serve rules from Git repositories
- **Dynamic rules**: Update shared rules without modifying individual repositories

**Usage:**

```bash
# Clone a Git repository containing rules
coding-context -d git::https://github.com/company/shared-rules.git /fix-bug

# Use multiple remote sources
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  -d https://cdn.company.com/coding-standards \
  /deploy

# Mix local and remote directories
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  -s languages=go \
  /implement-feature
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
coding-context \
  -d 'git::https://github.com/company/shared-rules.git?ref=v1.0' \
  /fix-bug

# Use a subdirectory within the repo
coding-context \
  -d 'git::https://github.com/company/mono-repo.git//coding-standards' \
  /implement-feature
```

## File Formats

### Task Files

Task files are Markdown files located in task search directories (e.g., `.agents/tasks/`). Tasks are matched by filename (without the `.md` extension), not by frontmatter fields. The `task_name` field in frontmatter is optional metadata. Task files can contain variables for substitution and can use selectors in frontmatter to filter between multiple task files with the same filename in different locations.

**Example (`.agents/tasks/fix-bug.md`):**
```markdown
# Task: Fix Bug in ${jira_issue_key}

Here is the context for the bug. Please analyze the following files and provide a fix.
```

**Example with selectors for multiple prompts (`.agents/tasks/deploy-staging.md`):**
```markdown
---
environment: staging
---
# Deploy to Staging

Deploy the application to the staging environment with extra validation.
```

**Example for production (`.agents/tasks/deploy-prod.md`):**
```markdown
---
environment: production
---
# Deploy to Production

Deploy the application to production with all safety checks.
```

You can then select the appropriate task using:
```bash
# Deploy to staging
coding-context -s environment=staging /deploy

# Deploy to production
coding-context -s environment=production /deploy
```

#### Task Frontmatter Selectors

Task files can include a `selectors` field in their frontmatter to automatically filter rules without requiring explicit `-s` flags on the command line. This is useful for tasks that always need specific rules.

**Example (`.agents/tasks/implement-go-feature.md`):**
```markdown
---
selectors:
  languages: go
  stage: implementation
---
# Implement Feature

Implement the feature following Go best practices and implementation guidelines.
```

When you run this task, it automatically applies the selectors:
```bash
# This command automatically includes only rules with languages=go and stage=implementation
coding-context implement-feature
```

This is equivalent to:
```bash
coding-context -s languages=go -s stage=implementation /implement-feature
```

**Selectors support OR logic for the same key using arrays:**
```markdown
---
selectors:
  languages: [go, python]
  stage: testing
---
```

This will include rules that match `(languages=go OR languages=python) AND stage=testing`.

**Combining task selectors with command-line selectors:**

Selectors from both the task frontmatter and command line are combined (additive):
```bash
# Task has: selectors.languages = go
# Command adds: -s priority=high
# Result: includes rules matching languages=go AND priority=high
coding-context -s priority=high implement-feature
```

### Parameter Expansion Control

By default, parameter expansion occurs in all task and rule content. You can disable this behavior using the `expand` frontmatter field.

**Example (task with expansion disabled):**
```yaml
---
expand: false
---

Issue: ${issue_number}
Title: ${issue_title}
```

When `expand: false` is set, parameter placeholders like `${variable}` are preserved as-is in the output, rather than being replaced with values from `-p` flags.

**Use cases:**
- Passing templates to AI agents that handle their own parameter substitution
- Preserving template syntax for later processing
- Avoiding conflicts with other templating systems

The `expand` field works in:
- Task files (`.agents/tasks/*.md`)
- Command files (`.agents/commands/*.md`)
- Rule files (`.agents/rules/*.md`)

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
coding-context -s resume=false /fix-bug | ai-agent

# Resume the task (skips rules, uses task with resume: true)
coding-context -r /fix-bug | ai-agent
```

**Example task files for resume mode:**

Initial task (`.agents/tasks/fix-bug-initial.md`):
```markdown
---
resume: false
---
# Fix Bug

Analyze the issue and implement a fix.
Follow the coding standards and write tests.
```

Resume task (`.agents/tasks/fix-bug-resume.md`):
```markdown
---
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
languages:
  - go
---

# Backend Coding Standards

- All new code must be accompanied by unit tests.
- Use the standard logging library.
```

To include this rule only when working on Go code, you would use `-s languages=go`:

```bash
coding-context -s languages=go /fix-bug
```

This will include all rules with `languages: [ go ]` in their frontmatter, excluding rules for other languages.

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`). The frontmatter field is `languages` (plural) with array format.

**Example: Language-Specific Rules**

You can create multiple language-specific rule files:

- `.agents/rules/python-standards.md` with `languages: [ python ]`
- `.agents/rules/javascript-standards.md` with `languages: [ javascript ]`
- `.agents/rules/go-standards.md` with `languages: [ go ]`

Then select only the relevant rules:

```bash
# Work on Python code with Python-specific rules
coding-context -s languages=python /fix-bug

# Work on JavaScript code with JavaScript-specific rules
coding-context -s languages=javascript /enhance-feature
```

**Language Values**

When using language selectors, language values should be **lowercase** (e.g., `go`, `python`, `javascript`, `java`, `typescript`). The frontmatter field should be `languages` (plural) in array format:

```yaml
---
languages:
  - go
  - python
---
```

**Common languages (lowercase):**
- `c`, `csharp` (C#), `cpp` (C++), `css`
- `dart`, `elixir`, `go`, `haskell`, `html`
- `java`, `javascript`, `kotlin`, `lua`, `markdown`
- `objectivec` (Objective-C), `php`, `python`, `ruby`, `rust`
- `scala`, `shell`, `swift`, `typescript`, `yaml`

**Note:** Language values should be lowercase in frontmatter and selectors.

**Note:** Frontmatter selectors can only match top-level YAML fields. For example:
- ✅ Works: `languages: [ go ]` matches `-s languages=go`
- ❌ Doesn't work: Nested fields like `metadata.version: 1.0` cannot be matched with `-s metadata.version=1.0`

If you need to filter on nested data, flatten your frontmatter structure to use top-level fields only.

### Targeting a Specific Agent

When working with a specific AI coding agent, the agent itself will read its own configuration files. The `-a` flag lets you specify which agent you're using, automatically excluding that agent's specific rule paths while including rules from other agents and generic rules.

**Supported agents:**
- `cursor` - Excludes `.cursor/rules`, `.cursorrules`; includes other agents and generic rules
- `opencode` - Excludes `.opencode/agent`, `.opencode/command`; includes other agents and generic rules
- `copilot` - Excludes `.github/copilot-instructions.md`, `.github/agents`; includes other agents and generic rules
- `claude` - Excludes `.claude/`, `CLAUDE.md`, `CLAUDE.local.md`; includes other agents and generic rules
- `gemini` - Excludes `.gemini/`, `GEMINI.md`; includes other agents and generic rules
- `augment` - Excludes `.augment/`; includes other agents and generic rules
- `windsurf` - Excludes `.windsurf/`, `.windsurfrules`; includes other agents and generic rules
- `codex` - Excludes `.codex/`, `AGENTS.md`; includes other agents and generic rules

**Example: Using Cursor:**

```bash
# When using Cursor, exclude .cursor/ and .cursorrules (Cursor reads those itself)
# But include rules from other agents and generic rules
coding-context -a cursor /fix-bug
```

**How it works:**
- The `-a` flag sets the target agent
- The target agent's own paths are excluded (e.g., `.cursor/` for cursor)
- Rules from other agents are included (e.g., `.opencode/`, `.github/copilot-instructions.md`)
- Generic rules (from `.agents/rules`) are always included
- The agent name is automatically added as a selector, so generic rules can filter themselves with `agent: cursor` in frontmatter

**Example generic rule with agent filtering:**

```markdown
---
agent: cursor
---
# This rule only applies when using Cursor
Use Cursor-specific features...
```

**Agent field in task frontmatter:**

Tasks can specify an `agent` field in their frontmatter, which overrides the `-a` command-line flag:

```markdown
---
agent: cursor
---
# This task automatically sets the agent to cursor
```

This is useful for tasks designed for specific agents, ensuring the correct agent context is used regardless of command-line flags.

**Use cases:**
- **Avoid duplication**: The agent reads its own config, so exclude it from the context
- **Cross-agent rules**: Include rules from other agents that might be relevant
- **Generic rules**: Always include generic rules, with optional agent-specific filtering
- **Task-specific agents**: Tasks can enforce a specific agent context

The exclusion happens before rule processing, so excluded paths are never loaded or counted toward token estimates.

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

### Command Files (Slash Commands)

Command files are reusable content blocks that can be referenced from task files using slash command syntax (e.g., `/command-name`). They enable modular, composable task definitions.

**Example command file (`.agents/commands/pre-deploy.md`):**
```markdown
---
expand: true
---
# Pre-deployment Checklist

Run tests: !`npm test`
Build status: !`git status`
```

**Using commands in a task (`.agents/tasks/deploy.md`):**
```markdown
# Deployment Task

/pre-deploy

Deploy the application to ${environment}.

/post-deploy
```

Commands can also receive inline parameters:
```markdown
/greet name="Alice"
/deploy env="production" version="1.2.3"
```

### Task Frontmatter

Task frontmatter is **always** automatically included at the beginning of the output when a task file has frontmatter. This allows the AI agent or downstream tool to access metadata about the task being executed. There is no flag needed to enable this - it happens automatically.

**Example usage:**
```bash
coding-context -p issue_number=123 fix-bug
```

**Output format:**
```yaml
---
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
coding-context implement-feature
```

If the task has `selectors` in its frontmatter, they will be visible in the output:
```yaml
---
selectors:
  languages: go
  stage: implementation
---
# Implementation Task
...
```

## Presentations

A comprehensive slide deck is available for presenting and learning about the Coding Context CLI:

- **[View Slide Deck](./SLIDES.md)** - Full presentation with 50+ slides
- **[Presentation Guide](./SLIDES_README.md)** - How to view, export, and present
- **[Example Usage](./examples/PRESENTATION.md)** - Presentation scenarios and tips

The slides are written in [Marp](https://marp.app/) format and can be:
- Viewed in VS Code with the Marp extension
- Exported to HTML, PDF, or PowerPoint
- Presented directly from a browser
- Customized for your audience

Perfect for:
- **Team onboarding** - Introduce the tool to new team members
- **Tech talks** - Present at meetups or conferences
- **Workshops** - Run hands-on training sessions
- **Documentation** - Visual reference for features
