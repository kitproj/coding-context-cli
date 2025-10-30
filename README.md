# coding-agent-context-cli

A CLI tool for managing context files for coding agents. This tool helps you organize prompts, memories (reusable context), and bootstrap scripts that can be assembled into a single context file for AI coding agents.

## Overview

`coding-agent-context-cli` allows you to:
- Create task-specific prompts with template parameters
- Store reusable context information (memories) that apply across multiple tasks
- Include bootstrap scripts that run before your agent starts work
- Generate a unified `prompt.md` file combining all relevant context

## Installation

### From Source

```bash
git clone https://github.com/kitproj/coding-agent-context-cli.git
cd coding-agent-context-cli
go build -o coding-agent-context .
```

Move the binary to your PATH:
```bash
sudo mv coding-agent-context /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/kitproj/coding-agent-context-cli@latest
```

## Quick Start

1. Create a context directory structure:
```bash
mkdir -p .coding-agent-context/{prompts,memories}
```

2. Create a prompt file (`.coding-agent-context/prompts/my-task.md`):
```markdown
---
---
# Task: {{ .taskName }}

Please help me with this task. The project uses {{ .language }}.
```

3. Create a memory file (`.coding-agent-context/memories/project-info.md`):
```markdown
---
---
# Project Context

- Framework: Go CLI
- Purpose: Manage AI agent context
```

4. Run the tool:
```bash
coding-agent-context -p taskName="Fix Bug" -p language=Go my-task
```

This generates `./prompt.md` combining your memories and prompt.

## Directory Structure

The tool looks for context files in these directories (in order):
1. `.coding-agent-context/` (project-local)
2. `~/.config/coding-agent-context/` (user-specific)
3. `/var/local/coding-agent-context/` (system-wide)

Each directory can contain:
```
.coding-agent-context/
├── prompts/          # Task-specific prompt templates
│   ├── task1.md
│   └── task2.md
└── memories/         # Reusable context files
    ├── project.md
    └── conventions.md
```

## Usage

```
coding-agent-context [options] <task-name>
```

### Options

- `-d <directory>` - Add a directory to include in the context. Can be specified multiple times.
  - Default directories: `.coding-agent-context`, `~/.config/coding-agent-context`, `/var/local/coding-agent-context`
  
- `-o <directory>` - Output directory for generated files (default: `.`)

- `-p <key=value>` - Template parameter for prompt substitution. Can be specified multiple times.

### Arguments

- `<task-name>` - The name of the task/prompt file (without `.md` extension)

## File Formats

### Prompt Files

Prompt files are Markdown documents with optional YAML frontmatter and support Go template syntax.

**Location:** `.coding-agent-context/prompts/<task-name>.md`

**Example:**
```markdown
---
---
# Task: {{ .feature }}

Implement {{ .feature }} in {{ .language }}.

Requirements:
- Write tests
- Follow existing patterns
```

Use it:
```bash
coding-agent-context -p feature="User Login" -p language=Go implement-feature
```

### Memory Files

Memory files are Markdown documents that get included in every generated context. They can include bootstrap scripts.

**Location:** `.coding-agent-context/memories/*.md`

**Example:**
```markdown
---
---
# Coding Standards

- Use tabs for indentation
- Write descriptive commit messages
- All functions must have tests
```

**Example with Bootstrap:**
```markdown
---
bootstrap: |
  #!/bin/bash
  npm install
  npm run build
---
# Development Setup

This project requires Node.js dependencies to be installed.
```

## Output Files

Running the tool generates:

- **`prompt.md`** - Combined output including all memories and the task prompt
- **`bootstrap`** - Shell script that executes all bootstrap scripts from memories
- **`bootstrap.d/`** - Directory containing individual bootstrap scripts (identified by SHA256 hash)

### Bootstrap Execution

The generated `bootstrap` script executes all bootstrap scripts defined in memory files:

```bash
# Run the bootstrap script
./bootstrap
```

The bootstrap script automatically finds and executes all scripts in `bootstrap.d/`.

## Examples

### Example 1: Simple Task

Create a prompt:
```bash
mkdir -p .coding-agent-context/prompts
cat > .coding-agent-context/prompts/refactor.md << 'EOF'
---
---
# Refactoring Task

Please refactor the codebase to improve code quality.
EOF
```

Run:
```bash
coding-agent-context refactor
```

### Example 2: Parameterized Prompt

Create a parameterized prompt:
```bash
cat > .coding-agent-context/prompts/add-feature.md << 'EOF'
---
---
# Add Feature: {{ .featureName }}

Implement {{ .featureName }} following these requirements:
- Language: {{ .language }}
- Framework: {{ .framework }}
EOF
```

Run with parameters:
```bash
coding-agent-context \
  -p featureName="Authentication" \
  -p language=Go \
  -p framework="net/http" \
  add-feature
```

### Example 3: With Memory and Bootstrap

Create a memory with bootstrap:
```bash
mkdir -p .coding-agent-context/memories
cat > .coding-agent-context/memories/setup.md << 'EOF'
---
bootstrap: |
  #!/bin/bash
  set -e
  echo "Installing dependencies..."
  go mod download
  echo "Dependencies installed!"
---
# Project Setup

This Go project uses modules for dependency management.
EOF
```

Create a simple prompt:
```bash
cat > .coding-agent-context/prompts/test.md << 'EOF'
---
---
# Write Tests

Add unit tests for the new features.
EOF
```

Run and execute bootstrap:
```bash
coding-agent-context -o ./output test
cd output
./bootstrap  # Runs go mod download
```

### Example 4: Multiple Context Directories

```bash
# Add project-specific context
coding-agent-context -d .coding-agent-context -d ../shared-context my-task
```

### Example 5: Custom Output Location

```bash
# Generate context in a specific directory
coding-agent-context -o /tmp/agent-context my-task
```

## Use Cases

### For AI Coding Agents

Provide consistent context to your coding agent:
1. Store project conventions in memories
2. Create task-specific prompts
3. Generate unified context before each agent session

### For Team Workflows

Share context across teams:
1. Project context in `.coding-agent-context/` (version controlled)
2. Personal preferences in `~/.config/coding-agent-context/`
3. Organization standards in `/var/local/coding-agent-context/`

### For Automated Workflows

Include bootstrap scripts for environment setup:
- Install dependencies
- Configure tools
- Run health checks

## Advanced Usage

### Template Functions

Prompts use Go's `text/template` syntax. All parameters passed via `-p` are available:

```markdown
{{ .variableName }}        # Simple substitution
{{ if .debug }}Debug{{ end }}  # Conditionals
```

### Directory Priority

When the same task exists in multiple directories, the first match wins:
1. `.coding-agent-context/` (highest priority)
2. `~/.config/coding-agent-context/`
3. `/var/local/coding-agent-context/` (lowest priority)

### Bootstrap Script Hashing

Bootstrap scripts are stored by SHA256 hash to avoid duplication and ensure consistency.

## Troubleshooting

### "prompt file not found for task"

Make sure your prompt file:
- Is named `<task-name>.md`
- Is in a `prompts/` subdirectory
- Is in one of the searched directories

### "failed to walk memory dir"

Ensure the `memories/` directory exists in at least one context directory:
```bash
mkdir -p .coding-agent-context/memories
```

### Template parameter shows `<no value>`

You forgot to pass the parameter via `-p`:
```bash
coding-agent-context -p myvar="value" my-task
```

### Bootstrap script not executing

1. Check the script has execute permissions:
```bash
chmod +x bootstrap
```

2. Ensure the script is valid shell syntax

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## See Also

- [Go text/template documentation](https://pkg.go.dev/text/template)
- [YAML frontmatter](https://yaml.org/)
