# Coding Context CLI

A CLI tool for managing context files for coding agents. It helps you organize prompts, memories (reusable context), and bootstrap scripts that can be assembled into a single context file for AI coding agents.

## Why Use This?

When working with AI coding agents (like GitHub Copilot, ChatGPT, Claude, etc.), providing the right context is crucial for getting quality results. However, managing this context becomes challenging when:

- **Context is scattered**: Project conventions, coding standards, and setup instructions are spread across multiple documents
- **Repetition is tedious**: You find yourself copy-pasting the same information into every AI chat session
- **Context size is limited**: AI models have token limits, so you need to efficiently select what's relevant
- **Onboarding is manual**: New team members or agents need step-by-step setup instructions

**This tool solves these problems by:**

1. **Centralizing reusable context** - Store project conventions, coding standards, and setup instructions once in "memory" files
2. **Creating task-specific prompts** - Define templated prompts for common tasks (e.g., "add feature", "fix bug", "refactor")
3. **Automating environment setup** - Package bootstrap scripts that prepare the environment before an agent starts work
4. **Filtering context dynamically** - Use selectors to include only relevant context (e.g., production vs. development, Python vs. Go)
5. **Composing everything together** - Generate a single `prompt.md` file combining all relevant context and the task prompt

## When to Use

This tool is ideal for:

- **Working with AI coding agents** - Prepare comprehensive context before starting a coding session
- **Team standardization** - Share common prompts and conventions across your team
- **Complex projects** - Manage large amounts of project-specific context efficiently
- **Onboarding automation** - New developers or agents can run bootstrap scripts to set up their environment
- **Multi-environment projects** - Filter context based on environment (dev/staging/prod) or technology stack

## How It Works

The basic workflow is:

1. **Organize your context** - Create memory files (shared context) and prompt files (task-specific instructions)
2. **Run the CLI** - Execute `coding-context <task-name>` with optional parameters
3. **Get assembled output** - The tool generates:
   - `prompt.md` - Combined context + task prompt with template variables filled in
   - `bootstrap` - Executable script to set up the environment
   - `bootstrap.d/` - Individual bootstrap scripts from your memory files
4. **Use with AI agents** - Share `prompt.md` with your AI coding agent, or run `./bootstrap` to prepare the environment first

**Visual flow:**
```
Memory Files (*.md)              Prompt Template (task-name.md)
        |                                    |
        | (filter by selectors)              | (apply template params)
        v                                    v
   Filtered Memories  -------+-------  Rendered Prompt
                              |
                              v
                        prompt.md (combined output)
```

## Installation

Download the binary for your platform from the release page:

```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-agent-context-cli/releases/download/v0.0.1/coding-context_v0.0.1_linux_arm64
sudo chmod +x /usr/local/bin/coding-context
```

### Using Go Install

```bash
go install github.com/kitproj/coding-agent-context-cli@latest
```

## Usage

```
coding-context [options] <task-name>

Options:
  -d <directory>    Add a directory to include in the context (can be used multiple times)
                    Default: .coding-context, ~/.config/coding-context, /var/local/coding-context
  -o <directory>    Output directory for generated files (default: .)
  -p <key=value>    Template parameter for prompt substitution (can be used multiple times)
  -s <key=value>    Include memories with matching frontmatter (can be used multiple times)
  -S <key=value>    Exclude memories with matching frontmatter (can be used multiple times)
```

**Example:**
```bash
coding-context -p feature="Authentication" -p language=Go add-feature
```

**Example with selectors:**
```bash
# Include only production memories
coding-context -s env=production deploy

# Exclude test memories
coding-context -S env=test deploy

# Combine include and exclude selectors
coding-context -s env=production -S language=python deploy
```

## Quick Start

This 4-step guide shows how to set up and generate your first context:

**Step 1: Create a context directory structure**
```bash
mkdir -p .coding-context/{prompts,memories}
```

**Step 2: Create a memory file** (`.coding-context/memories/project-info.md`)

Memory files are included in every generated context. They contain reusable information like project conventions, architecture notes, or coding standards.

```markdown
---
---
# Project Context

- Framework: Go CLI
- Purpose: Manage AI agent context
```

**Step 3: Create a prompt file** (`.coding-context/prompts/my-task.md`)

Prompt files define specific tasks. They can use template variables (like `{{ .taskName }}`) that you provide via command-line parameters.

```markdown
---
---
# Task: {{ .taskName }}

Please help me with this task. The project uses {{ .language }}.
```

**Step 4: Generate your context file**

```bash
coding-context -p taskName="Fix Bug" -p language=Go my-task
```

**Result:** This generates `./prompt.md` combining your memories and the task prompt with template variables filled in. You can now share this complete context with your AI coding agent!

**What you'll see in `prompt.md`:**
```markdown
# Project Context

- Framework: Go CLI
- Purpose: Manage AI agent context



# Task: Fix Bug

Please help me with this task. The project uses Go.
```


## Directory Structure

The tool searches these directories for context files (in priority order):
1. `.coding-context/` (project-local)
2. `~/.config/coding-context/` (user-specific)
3. `/var/local/coding-context/` (system-wide)

Each directory should contain:
```
.coding-context/
├── prompts/          # Task-specific prompt templates
│   └── <task-name>.md
└── memories/         # Reusable context files (included in all outputs)
    └── *.md
```


## File Formats

### Prompt Files

Markdown files with YAML frontmatter and Go template support.

**Example** (`.coding-context/prompts/add-feature.md`):
```markdown
---
---
# Task: {{ .feature }}

Implement {{ .feature }} in {{ .language }}.
```

Run with:
```bash
coding-context -p feature="User Login" -p language=Go add-feature
```

### Memory Files

Markdown files included in every generated context. Bootstrap scripts can be provided in separate files.

**Example** (`.coding-context/memories/setup.md`):
```markdown
---
env: development
language: go
---
# Development Setup

This project requires Node.js dependencies.
```

**Bootstrap file** (`.coding-context/memories/setup-bootstrap`):
```bash
#!/bin/bash
npm install
```

For each memory file `<name>.md`, you can optionally create a corresponding `<name>-bootstrap` file that will be executed during setup.


## Filtering Memories with Selectors

Use the `-s` and `-S` flags to filter which memory files are included based on their frontmatter metadata.

### Selector Syntax

- **`-s key=value`** - Include memories where the frontmatter key matches the value
- **`-S key=value`** - Exclude memories where the frontmatter key matches the value
- If a key doesn't exist in a memory's frontmatter, the memory is allowed (not filtered out)
- Multiple selectors of the same type use AND logic (all must match)

### Examples

**Include only production memories:**
```bash
coding-context -s env=production deploy
```

**Exclude test environment:**
```bash
coding-context -S env=test deploy
```

**Combine include and exclude:**
```bash
# Include production but exclude python
coding-context -s env=production -S language=python deploy
```

**Multiple includes:**
```bash
# Only production Go backend memories
coding-context -s env=production -s language=go -s tier=backend deploy
```

### How It Works

When you run with selectors, the tool logs which files are included or excluded:

```
INFO Including memory file path=.coding-context/memories/production.md
INFO Excluding memory file (does not match include selectors) path=.coding-context/memories/development.md
INFO Including memory file path=.coding-context/memories/nofrontmatter.md
```

**Important:** Files without the specified frontmatter keys are still included. This allows you to have generic memories that apply to all scenarios.

If no selectors are specified, all memory files are included.


## Output Files

- **`prompt.md`** - Combined output with all memories and the task prompt
- **`bootstrap`** - Executable script that runs all bootstrap scripts from memories
- **`bootstrap.d/`** - Individual bootstrap scripts (SHA256 named)

Run the bootstrap script to set up your environment:
```bash
./bootstrap
```


## Examples

### Basic Usage

```bash
# Create structure
mkdir -p .coding-context/{prompts,memories}

# Add a memory
cat > .coding-context/memories/conventions.md << 'EOF'
---
---
# Coding Conventions

- Use tabs for indentation
- Write tests for all functions
EOF

# Create a task prompt
cat > .coding-context/prompts/refactor.md << 'EOF'
---
---
# Refactoring Task

Please refactor the codebase to improve code quality.
EOF

# Generate context
coding-context refactor
```

### With Template Parameters

```bash
cat > .coding-context/prompts/add-feature.md << 'EOF'
---
---
# Add Feature: {{ .featureName }}

Implement {{ .featureName }} in {{ .language }}.
EOF

coding-context -p featureName="Authentication" -p language=Go add-feature
```

### With Bootstrap Scripts

```bash
cat > .coding-context/memories/setup.md << 'EOF'
---
---
# Project Setup

This Go project uses modules.
EOF

cat > .coding-context/memories/setup-bootstrap << 'EOF'
#!/bin/bash
go mod download
EOF
chmod +x .coding-context/memories/setup-bootstrap

coding-context -o ./output my-task
cd output && ./bootstrap
```

## Advanced Usage

### Template Functions

Prompts use Go's `text/template` syntax:

```markdown
{{ .variableName }}                                    # Simple substitution
{{ if .debug }}Debug mode enabled{{ else }}Production mode{{ end }}  # Conditionals
```

### Directory Priority

When the same task exists in multiple directories, the first match wins:
1. `.coding-context/` (highest priority)
2. `~/.config/coding-context/`
3. `/var/local/coding-context/` (lowest priority)

## Troubleshooting

**"prompt file not found for task"**
- Ensure `<task-name>.md` exists in a `prompts/` subdirectory

**"failed to walk memory dir"**
```bash
mkdir -p .coding-context/memories
```

**Template parameter shows `<no value>`**
```bash
coding-context -p myvar="value" my-task
```

**Bootstrap script not executing**
```bash
chmod +x bootstrap
```

