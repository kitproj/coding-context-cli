# Coding Context CLI

A CLI tool for importing and managing coding agent rules. It helps you organize rule files (reusable context) from various AI coding agents and combine them into a unified format.

## Why Use This?

When working with AI coding agents (like GitHub Copilot, Claude, Cursor, Gemini, etc.), each tool has its own way of storing rules and configuration. This tool solves the problem by:

1. **Importing agent-specific rules** - Extract rules from agent-specific formats and locations
2. **Unified output** - Generate a single `rules.md` file with all rules combined
3. **Hierarchical rule search** - Automatically search up the directory tree for ancestor rules
4. **Automating environment setup** - Package bootstrap scripts that prepare the environment before an agent starts work

## Supported Agents

The tool currently supports importing rules from:

- **Claude** - CLAUDE.md, CLAUDE.local.md
- **Gemini** - GEMINI.md, .gemini/styleguide.md
- **Codex** - AGENTS.md
- **Cursor** - .cursor/rules/, AGENTS.md
- **GitHub Copilot** - .github/copilot-instructions.md, .github/agents/, AGENTS.md
- **Augment** - .augment/rules/, .augment/guidelines.md, CLAUDE.md, AGENTS.md
- **Windsurf** - .windsurf/rules/
- **Goose** - AGENTS.md
- **Continue.dev** - .continuerules

Each agent has its own set of rule paths and hierarchy levels (Project, Ancestor, User, System).

## Installation

Download the binary for your platform from the release page:

```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-agent-context-cli/releases/download/v0.0.1/coding-context_v0.0.1_linux_arm64
sudo chmod +x /usr/local/bin/coding-context
```

## Usage

### Import Rules

Import rules from a specific agent:

```bash
coding-context import <agent>
```

**Examples:**

```bash
# Import Codex rules (searches for AGENTS.md in current and ancestor directories)
coding-context import Codex

# Import Claude rules (searches for CLAUDE.md, CLAUDE.local.md)
coding-context import Claude

# Import Cursor rules (searches for .cursor/rules/ directory)
coding-context import Cursor

# Import to a specific output directory
coding-context -o ./output import Gemini
```

### Run Bootstrap Scripts

After importing rules, run the bootstrap scripts to set up the environment:

```bash
coding-context bootstrap
```

This executes the `bootstrap` script which runs all individual bootstrap scripts found in `bootstrap.d/`.

### Other Commands

- **export** - Export rules for a specific agent (TODO - not yet implemented)
- **prompt** - Find and print prompts (TODO - not yet implemented)

## How It Works

### Rule Hierarchy

Rules are organized into four levels (from highest to lowest precedence):

1. **Project Level (0)** - Specific paths in the current working directory
   - Example: `./CLAUDE.local.md`, `./.cursor/rules/`
   
2. **Ancestor Level (1)** - Files searched up the directory tree
   - Example: `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`
   
3. **User Level (2)** - Specific paths in the user's home directory
   - Example: `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`
   
4. **System Level (3)** - System-wide rules
   - Example: `/usr/local/prompts-rules`

### Ancestor Path Search

When a rule is marked as "Ancestor Level", the tool automatically searches up the directory tree from the current working directory to the root, collecting all matching files along the way.

For example, if you run `coding-context import Codex` from `/home/user/project/sub/dir`, it will search for `AGENTS.md` in:
- `/home/user/project/sub/dir/AGENTS.md`
- `/home/user/project/sub/AGENTS.md`
- `/home/user/project/AGENTS.md`
- `/home/user/AGENTS.md`
- `/home/AGENTS.md`
- `/AGENTS.md`

All found files are included in `rules.md`, ordered from closest to farthest (highest to lowest precedence).

## Output Files

The `import` command generates:

- **`rules.md`** - Combined output with all rule files merged together
- **`bootstrap`** - Executable script that runs all bootstrap scripts
- **`bootstrap.d/`** - Individual bootstrap scripts from rule files (with SHA256 hash in filename)

## Bootstrap Scripts

Rule files can have associated bootstrap scripts that set up dependencies or environment. For example:

**File structure:**
```
AGENTS.md
AGENTS-bootstrap
```

The bootstrap script (`AGENTS-bootstrap`) will be copied to `bootstrap.d/` with a hash suffix and made executable. The main `bootstrap` script will run all scripts in `bootstrap.d/`.

**Example bootstrap script:**
```bash
#!/bin/bash
set -euo pipefail
npm install
go mod download
```

## Examples

### Basic Import

Import rules from your current project for Codex:

```bash
# Create AGENTS.md in your project
echo "# Project Rules" > AGENTS.md
echo "Follow these coding standards..." >> AGENTS.md

# Import the rules
coding-context import Codex

# View the output
cat rules.md
```

### Multi-Level Rules

Create rules at different levels for better organization:

```bash
# Project-specific rule (CLAUDE.local.md - personal, not checked into git)
echo "# My Personal Claude Rules" > CLAUDE.local.md

# Project-wide rule (CLAUDE.md - checked into git)
echo "# Team Claude Rules" > CLAUDE.md

# Import both
coding-context import Claude

# Both files are included in rules.md
```

### Cursor with Directory-Based Rules

```bash
# Create Cursor rules directory
mkdir -p .cursor/rules

# Add multiple rule files
echo "# TypeScript Rules" > .cursor/rules/typescript.md
echo "# React Rules" > .cursor/rules/react.mdc

# Import all Cursor rules
coding-context import Cursor
```

### With Bootstrap Scripts

```bash
# Create a rule file
echo "# Setup Instructions" > AGENTS.md

# Create a bootstrap script
cat > AGENTS-bootstrap << 'EOF'
#!/bin/bash
set -euo pipefail
echo "Installing dependencies..."
npm install
go mod download
echo "Setup complete!"
EOF
chmod +x AGENTS-bootstrap

# Import (creates bootstrap.d/ with the script)
coding-context import Codex

# Run the bootstrap
coding-context bootstrap
```

### Hierarchical Rules (Ancestor Search)

```bash
# Create a root-level rule
cd /home/user/myproject
echo "# Root Project Rules" > AGENTS.md

# Create a subdirectory-specific rule
mkdir -p backend/api
cd backend/api
echo "# API-Specific Rules" > AGENTS.md

# Import from the subdirectory (finds both files)
coding-context import Codex

# rules.md will contain:
# 1. backend/api/AGENTS.md (closest, highest precedence)
# 2. /home/user/myproject/AGENTS.md (ancestor, lower precedence)
```

## Global Options

- **`-C <directory>`** - Change to directory before running the command (default: `.`)
- **`-o <directory>`** - Output directory for generated files (default: `.`)

## Agent-Specific Rule Paths

Each agent searches for rules in specific locations:

### Claude
- **Project:** `CLAUDE.local.md` (personal overrides)
- **Ancestor:** `CLAUDE.md` (project-wide)
- **User:** `~/.claude/CLAUDE.md` (global defaults)

### Gemini
- **Project:** `.gemini/styleguide.md`
- **Ancestor:** `GEMINI.md`
- **User:** `~/.gemini/GEMINI.md`

### Codex
- **Ancestor:** `AGENTS.md` (searches up directory tree)
- **User:** `~/.codex/AGENTS.md`

### Cursor
- **Project:** `.cursor/rules/` (directory of .md and .mdc files)
- **Ancestor:** `AGENTS.md`

### GitHub Copilot
- **Project:** `.github/agents/`, `.github/copilot-instructions.md`
- **Ancestor:** `AGENTS.md`

### Augment
- **Project:** `.augment/rules/`, `.augment/guidelines.md`
- **Ancestor:** `CLAUDE.md`, `AGENTS.md`

### Windsurf
- **Project:** `.windsurf/rules/`

### Goose
- **Ancestor:** `AGENTS.md`

### Continue.dev
- **Project:** `.continuerules`

## File Formats

Rule files are standard Markdown (`.md`) or Cursor MDC format (`.mdc`). They can include YAML frontmatter for metadata:

```markdown
---
env: production
language: go
---
# Production Rules

Follow these production-specific guidelines...
```

## Future Commands (TODO)

- **`export <agent>`** - Export rules to agent-specific format
- **`prompt`** - Find and print prompts to stdout

## Development

Build from source:

```bash
git clone https://github.com/kitproj/coding-context-cli
cd coding-context-cli
go build -o coding-context .
```

Run tests:

```bash
go test -v ./...
```

## License

See LICENSE file for details.
