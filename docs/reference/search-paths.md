---
layout: default
title: Search Paths
parent: Reference
nav_order: 3
---

# Search Paths Reference

Complete reference for where the CLI searches for task files and rule files.

## Remote Directories

When using the `-d` flag, the CLI downloads remote directories to a temporary location and includes them in the search paths.

**Example:**
```bash
coding-context -d git::https://github.com/company/shared-rules.git fix-bug
```

The downloaded directory is searched for rules and tasks in all standard locations (`.agents/rules/`, `.agents/tasks/`, `AGENTS.md`, etc.) before being automatically cleaned up.

Multiple remote directories can be specified and are processed in the order given:
```bash
coding-context \
  -d git::https://github.com/company/org-standards.git \
  -d git::https://github.com/team/team-rules.git \
  fix-bug
```

See [How to Use Remote Directories](../how-to/use-remote-directories) for complete documentation.

## Local Search Paths

### Task File Search Paths

Task files are searched in the following directories, in order of precedence:

1. `./.agents/tasks/`
2. `./.cursor/commands/`
3. `./.opencode/command/`
4. `~/.agents/tasks/`

### Discovery Rules

- All `.md` files in these directories are examined
- If `task_name` is present in frontmatter, it's used for task identification
- If `task_name` is absent, the filename (without `.md` extension) is used as the task name
- First match wins (unless selectors create ambiguity)
- Searches stop when a matching task is found
- Remote directories (via `-d` flag) are searched before local directories

### Example

```
Project structure:
./.agents/tasks/fix-bug.md            (task_name: fix-bug)
./.opencode/command/review-code.md    (task_name: review-code)
./.opencode/command/deploy.md         (no task_name, uses filename)
~/.agents/tasks/code-review.md        (task_name: code-review)

Commands:
coding-context fix-bug          → Uses ./.agents/tasks/fix-bug.md
coding-context review-code      → Uses ./.opencode/command/review-code.md
coding-context deploy           → Uses ./.opencode/command/deploy.md
coding-context code-review      → Uses ~/.agents/tasks/code-review.md
coding-context deploy           → Uses ~/.config/opencode/command/deploy.md
```

## Rule File Search Paths

Rule files are discovered from multiple locations supporting various AI agent formats.

### Remote Directories (Highest Precedence)

When using `-d` flag, remote directories are searched first:

```bash
coding-context -d git::https://github.com/company/rules.git fix-bug
```

The remote directory is searched for all standard file patterns listed below.

### Project-Specific Rules

**Agent-specific directories:**
```
./.agents/rules/
./.cursor/rules/
./.augment/rules/
./.windsurf/rules/
./.opencode/agent/
./.opencode/command/
./.opencode/rules/
./.github/agents/
./.codex/
```

**Specific files:**
```
./CLAUDE.local.md
./.github/copilot-instructions.md
./.gemini/styleguide.md
```

**Standard files (searched in current and parent directories):**
```
./AGENTS.md
./CLAUDE.md
./GEMINI.md
../ (continues up to root)
```

### User-Specific Rules (Medium Precedence)

```
~/.agents/rules/
~/.claude/CLAUDE.md
~/.cursor/rules/
~/.augment/rules/
~/.windsurf/rules/
~/.opencode/rules/
~/.github/agents/
~/.codex/AGENTS.md
~/.gemini/styleguide.md
```

## Supported AI Agent Formats

The CLI automatically discovers rules from configuration files for these AI coding agents:

| Agent | File Locations |
|-------|----------------|
| **Anthropic Claude** | `CLAUDE.md`, `CLAUDE.local.md`, `.claude/CLAUDE.md` |
| **Codex** | `AGENTS.md`, `.codex/AGENTS.md` |
| **Cursor** | `.cursor/rules/`, `.cursorrules`, `.cursor/commands/` (tasks) |
| **Augment** | `.augment/rules/`, `.augment/guidelines.md` |
| **Windsurf** | `.windsurf/rules/`, `.windsurfrules` |
| **OpenCode.ai** | `.opencode/agent/`, `.opencode/command/` (tasks), `.opencode/rules/` |
| **GitHub Copilot** | `.github/copilot-instructions.md`, `.github/agents/` |
| **Google Gemini** | `GEMINI.md`, `.gemini/styleguide.md` |
| **Generic** | `AGENTS.md`, `.agents/rules/` |

## Discovery Behavior

### File Types

The CLI processes:
- `.md` files (Markdown)
- `.mdc` files (Markdown component)

Other file types are ignored.

### Directory Traversal

For standard files (like `AGENTS.md`, `CLAUDE.md`):
1. Start in current directory (or `-C` directory)
2. Check for file
3. Move to parent directory
4. Repeat until root or file found

**Example:**
```
/home/user/projects/myapp/backend/

Searches:
/home/user/projects/myapp/backend/AGENTS.md
/home/user/projects/myapp/AGENTS.md
/home/user/projects/AGENTS.md
/home/user/AGENTS.md
/home/AGENTS.md
/AGENTS.md
```

### Precedence Order

When multiple rule files exist:
1. Project-specific (`./.agents/rules/`)
2. Parent directories (moving up)
3. User-specific (`~/.agents/rules/`)

All matching files are included (unless filtered by selectors).

## Bootstrap Script Discovery

For each rule file `rule-name.md`, the CLI looks for `rule-name-bootstrap` in the same directory.

**Example:**
```
./.agents/rules/jira-context.md
./.agents/rules/jira-context-bootstrap  ← Must be here
```

**Not searched:**
```
./.agents/rules/jira-context.md
./.agents/tasks/jira-context-bootstrap  ← Wrong directory
./jira-context-bootstrap                ← Wrong directory
```

## Working Directory

The `-C` option changes the working directory before searching:

```bash
# Search from /path/to/project
coding-context -C /path/to/project fix-bug

# Equivalent to:
cd /path/to/project && coding-context fix-bug
```

This affects:
- Where `./.agents/` is located
- Parent directory traversal starting point
- Bootstrap script execution directory

## Custom Organization

You can organize rules in subdirectories:

```
.agents/
├── rules/
│   ├── planning/
│   │   ├── requirements.md
│   │   └── architecture.md
│   ├── implementation/
│   │   ├── go-standards.md
│   │   └── python-standards.md
│   └── testing/
│       └── test-requirements.md
└── tasks/
    ├── plan.md
    ├── implement.md
    └── test.md
```

All `.md` files in `.agents/rules/` and its subdirectories are discovered.

## Filtering

Regardless of where rules are found, they can be filtered using selectors:

```bash
# Include only Go rules (from any location)
coding-context -s language=Go fix-bug

# Include only planning rules
coding-context -s stage=planning plan-feature
```

## Examples

### Multi-Language Project

```
.agents/
└── rules/
    ├── general-standards.md        (no frontmatter - always included)
    ├── go-backend.md               (language: Go)
    ├── python-ml.md                (language: Python)
    └── javascript-frontend.md      (language: JavaScript)

Commands:
coding-context -s language=Go fix-bug
  → Includes: general-standards.md, go-backend.md

coding-context -s language=Python train-model
  → Includes: general-standards.md, python-ml.md
```

### Environment Tiers

```
.agents/rules/
├── security-base.md       (no frontmatter)
├── dev-config.md          (environment: development)
├── staging-config.md      (environment: staging)
└── prod-config.md         (environment: production)

Commands:
coding-context -s environment=production deploy
  → Includes: security-base.md, prod-config.md
```

### Team-Specific Rules

```
.agents/rules/
├── company-wide.md        (no frontmatter)
├── backend-team.md        (team: backend)
└── frontend-team.md       (team: frontend)

~/.agents/rules/
└── personal-preferences.md

Commands:
coding-context -s team=backend fix-bug
  → Includes: company-wide.md, backend-team.md, personal-preferences.md
```

## Troubleshooting

**No rules found:**
- Check that `.agents/rules/` directory exists
- Verify files have `.md` or `.mdc` extension
- Check file permissions (must be readable)

**Task not found:**
- Verify `.agents/tasks/` directory exists
- Check `task_name` field in frontmatter
- Ensure filename has `.md` extension

**Rules not filtered correctly:**
- Verify frontmatter YAML is valid
- Check selector spelling and capitalization
- Remember: only top-level frontmatter fields match

## See Also

- [File Formats Reference](./file-formats) - File specifications
- [CLI Reference](./cli) - Command-line options
- [How to Create Rules](../how-to/create-rules) - Organizing rules
