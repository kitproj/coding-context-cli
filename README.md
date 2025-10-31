# Coding Agent Context CLI

A CLI tool for managing context files for coding agents. It helps you organize prompts and memories (reusable context) that can be assembled into a single context file for AI coding agents.

It's aimed at coding agents with a simple interface for managing task-specific context and reusable knowledge.

## Installation

Download the binary for your platform from the release page:

```bash
sudo curl -fsL -o /usr/local/bin/coding-agent-context https://github.com/kitproj/coding-agent-context-cli/releases/download/v0.0.1/coding-agent-context_v0.0.1_linux_arm64
sudo chmod +x /usr/local/bin/coding-agent-context
```

### Using Go Install

```bash
go install github.com/kitproj/coding-agent-context-cli@latest
```

## Usage

```
coding-agent-context [options] <task-name>

Options:
  -d <directory>    Add a directory to include in the context (can be used multiple times)
                    Default: .coding-agent-context, ~/.config/coding-agent-context, /var/local/coding-agent-context
  -o <directory>    Output directory for generated files (default: .)
  -p <key=value>    Template parameter for prompt substitution (can be used multiple times)
```

**Example:**
```bash
coding-agent-context -p feature="Authentication" -p language=Go add-feature
```

## Quick Start

1. Create a context directory structure:
```bash
mkdir -p .coding-agent-context/{prompts,memories}
```

2. Create a memory file (`.coding-agent-context/memories/project-info.md`):
```markdown
---
---
# Project Context

- Framework: Go CLI
- Purpose: Manage AI agent context
```

3. Create a prompt file (`.coding-agent-context/prompts/my-task.md`):
```markdown
---
---
# Task: {{ .taskName }}

Please help me with this task. The project uses {{ .language }}.
```

4. Run the tool:
```bash
coding-agent-context -p taskName="Fix Bug" -p language=Go my-task
```

This generates `./prompt.md` combining your memories and the task prompt.


## Directory Structure

The tool searches these directories for context files (in priority order):
1. `.coding-agent-context/` (project-local)
2. `~/.config/coding-agent-context/` (user-specific)
3. `/var/local/coding-agent-context/` (system-wide)

Each directory should contain:
```
.coding-agent-context/
├── prompts/          # Task-specific prompt templates
│   └── <task-name>.md
└── memories/         # Reusable context files (included in all outputs)
    └── *.md
```


## File Formats

### Prompt Files

Markdown files with YAML frontmatter and Go template support.

**Example** (`.coding-agent-context/prompts/add-feature.md`):
```markdown
---
---
# Task: {{ .feature }}

Implement {{ .feature }} in {{ .language }}.
```

Run with:
```bash
coding-agent-context -p feature="User Login" -p language=Go add-feature
```

### Memory Files

Markdown files included in every generated context.

**Example** (`.coding-agent-context/memories/setup.md`):
```markdown
---
---
# Development Setup

This project requires Node.js dependencies.
```


## Output Files

- **`prompt.md`** - Combined output with all memories and the task prompt


## Examples

### Basic Usage

```bash
# Create structure
mkdir -p .coding-agent-context/{prompts,memories}

# Add a memory
cat > .coding-agent-context/memories/conventions.md << 'EOF'
---
---
# Coding Conventions

- Use tabs for indentation
- Write tests for all functions
EOF

# Create a task prompt
cat > .coding-agent-context/prompts/refactor.md << 'EOF'
---
---
# Refactoring Task

Please refactor the codebase to improve code quality.
EOF

# Generate context
coding-agent-context refactor
```

### With Template Parameters

```bash
cat > .coding-agent-context/prompts/add-feature.md << 'EOF'
---
---
# Add Feature: {{ .featureName }}

Implement {{ .featureName }} in {{ .language }}.
EOF

coding-agent-context -p featureName="Authentication" -p language=Go add-feature
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
1. `.coding-agent-context/` (highest priority)
2. `~/.config/coding-agent-context/`
3. `/var/local/coding-agent-context/` (lowest priority)

## Troubleshooting

**"prompt file not found for task"**
- Ensure `<task-name>.md` exists in a `prompts/` subdirectory

**"failed to walk memory dir"**
```bash
mkdir -p .coding-agent-context/memories
```

**Template parameter shows `<no value>`**
```bash
coding-agent-context -p myvar="value" my-task
```

