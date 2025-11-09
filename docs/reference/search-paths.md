---
layout: default
title: Search Paths
parent: Reference
nav_order: 3
---

# Search Paths Reference

Complete reference for where the CLI searches for task files and rule files.

## Task File Search Paths

Task files are searched in the following directories, in order of precedence:

1. `./.agents/tasks/`
2. `~/.agents/tasks/`

### Discovery Rules

- All `.md` files in these directories are examined
- The filename doesn't matter; only the `task_name` frontmatter field
- First match wins (unless selectors create ambiguity)
- Searches stop when a matching task is found

### Example

```
Project structure:
./.agents/tasks/fix-bug.md          (task_name: fix-bug)
~/.agents/tasks/code-review.md      (task_name: code-review)

Commands:
coding-context-cli fix-bug          → Uses ./.agents/tasks/fix-bug.md
coding-context-cli code-review      → Uses ~/.agents/tasks/code-review.md
```

## Rule File Search Paths

Rule files are discovered from multiple locations supporting various AI agent formats.

### Project-Specific Rules (Highest Precedence)

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
| **Cursor** | `.cursor/rules/`, `.cursorrules` |
| **Augment** | `.augment/rules/`, `.augment/guidelines.md` |
| **Windsurf** | `.windsurf/rules/`, `.windsurfrules` |
| **OpenCode.ai** | `.opencode/agent/`, `.opencode/command/`, `.opencode/rules/` |
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
coding-context-cli -C /path/to/project fix-bug

# Equivalent to:
cd /path/to/project && coding-context-cli fix-bug
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
coding-context-cli -s language=Go fix-bug

# Include only planning rules
coding-context-cli -s stage=planning plan-feature
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
coding-context-cli -s language=Go fix-bug
  → Includes: general-standards.md, go-backend.md

coding-context-cli -s language=Python train-model
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
coding-context-cli -s environment=production deploy
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
coding-context-cli -s team=backend fix-bug
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
