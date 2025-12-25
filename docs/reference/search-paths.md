---
layout: default
title: Search Paths
parent: Reference
nav_order: 3
---

# Search Paths Reference

Complete reference for where the CLI searches for task files and rule files.

## Search Paths Overview

The CLI searches for rules and tasks in directories specified via the `-d` flag. The working directory (`-C` or current directory) and home directory (`~`) are **automatically added** to the search paths, so they don't need to be specified explicitly.

All directories (local and remote) are processed via go-getter, which downloads remote directories to temporary locations and processes local directories directly.

### Task File Search Paths

Within each directory, task files are searched in the following locations:

1. `.agents/tasks/`

**Note:** Task files are matched by filename (without `.md` extension), not by `task_name` in frontmatter.

### Command File Search Paths (for slash commands)

Command files are referenced via slash commands inside task content. Within each directory, command files are searched in:

1. `.agents/commands/`
2. `.cursor/commands/`
3. `.opencode/command/`

### Skill File Search Paths

Skill files provide specialized capabilities with progressive disclosure. Within each directory, skill files are searched in:

1. `.agents/skills/*/SKILL.md` (each subdirectory in `.agents/skills/` can contain a `SKILL.md` file)
2. `.cursor/skills/*/SKILL.md` (each subdirectory in `.cursor/skills/` can contain a `SKILL.md` file)
3. `.opencode/skills/*/SKILL.md` (each subdirectory in `.opencode/skills/` can contain a `SKILL.md` file)
4. `.github/skills/*/SKILL.md` (each subdirectory in `.github/skills/` can contain a `SKILL.md` file)
5. `.claude/skills/*/SKILL.md` (each subdirectory in `.claude/skills/` can contain a `SKILL.md` file)
6. `.gemini/skills/*/SKILL.md` (each subdirectory in `.gemini/skills/` can contain a `SKILL.md` file)
7. `.augment/skills/*/SKILL.md` (each subdirectory in `.augment/skills/` can contain a `SKILL.md` file)
8. `.windsurf/skills/*/SKILL.md` (each subdirectory in `.windsurf/skills/` can contain a `SKILL.md` file)
9. `.codex/skills/*/SKILL.md` (each subdirectory in `.codex/skills/` can contain a `SKILL.md` file)

**Example:**
```
.agents/skills/
├── data-analysis/
│   └── SKILL.md
├── pdf-processing/
│   └── SKILL.md
└── api-testing/
    └── SKILL.md

.cursor/skills/
├── code-review/
│   └── SKILL.md
└── refactoring/
    └── SKILL.md

.opencode/skills/
├── testing/
│   └── SKILL.md
└── debugging/
    └── SKILL.md

.github/skills/
├── deployment/
│   └── SKILL.md
└── ci-cd/
    └── SKILL.md

.claude/skills/
├── analysis/
│   └── SKILL.md
└── writing/
    └── SKILL.md

.gemini/skills/
├── search/
│   └── SKILL.md
└── multimodal/
    └── SKILL.md

.codex/skills/
├── code-gen/
│   └── SKILL.md
└── refactoring/
    └── SKILL.md
```

### Discovery Rules

- All `.md` files in these directories are examined
- Tasks are matched by filename (without `.md` extension), not by `task_name` in frontmatter
- The `task_name` field in frontmatter is optional and used only for metadata
- If multiple files have the same filename, selectors are used to choose between them
- First match wins (unless selectors create ambiguity)
- Searches stop when a matching task is found
- Directories are searched in the order they appear in `-d` flags, then the automatically-added working directory and home directory

### Example

```
Project structure:
./.agents/tasks/fix-bug.md            (task file, matched by filename "fix-bug")
./.agents/tasks/code-review.md        (task file, matched by filename "code-review")
./.agents/commands/deploy-checks.md   (command file, referenced via slash command in tasks)
~/.agents/tasks/plan-feature.md       (task file in home directory)

Commands:
coding-context fix-bug           → Uses ./.agents/tasks/fix-bug.md (from working directory)
coding-context code-review       → Uses ./.agents/tasks/code-review.md (from working directory)
coding-context plan-feature      → Uses ~/.agents/tasks/plan-feature.md (from home directory)

# Command files are NOT invoked directly, but referenced inside task content via slash commands like:
# /deploy-checks arg1 arg2
```
```

**Note:** The working directory and home directory are automatically added to search paths, so tasks in those locations are found automatically.

## Rule File Search Paths

Rule files are discovered from directories specified via the `-d` flag (plus automatically-added working directory and home directory). Within each directory, the CLI searches for all standard file patterns listed below.

### Directory Processing Order

1. Directories specified via `-d` flags (in order)
2. Working directory (`-C` flag or current directory) - added automatically
3. Home directory (`~`) - added automatically

### Rule File Locations Within Each Directory

**Agent-specific directories:**
```
.agents/rules/
.cursor/rules/
.augment/rules/
.windsurf/rules/
.opencode/agent/
.github/agents/
```

**Specific files:**
```
CLAUDE.local.md
.github/copilot-instructions.md
.gemini/styleguide.md
.augment/guidelines.md
```

**Standard files:**
```
AGENTS.md
CLAUDE.md
GEMINI.md
.cursorrules
.windsurfrules
```

**User-specific locations (only in home directory):**
```
.agents/rules/
.claude/CLAUDE.md
.codex/AGENTS.md
.gemini/GEMINI.md
.opencode/rules/
```

## Supported AI Agent Formats

The CLI automatically discovers rules from configuration files for these AI coding agents:

| Agent | File Locations |
|-------|----------------|
| **Anthropic Claude** | `CLAUDE.md`, `CLAUDE.local.md`, `.claude/CLAUDE.md` |
| **Codex** | `AGENTS.md`, `.codex/AGENTS.md` |
| **Cursor** | `.cursor/rules/`, `.cursorrules`, `.cursor/commands/` (commands, not tasks) |
| **Augment** | `.augment/rules/`, `.augment/guidelines.md` |
| **Windsurf** | `.windsurf/rules/`, `.windsurfrules` |
| **OpenCode.ai** | `.opencode/agent/`, `.opencode/rules/`, `.opencode/command/` (commands, not tasks) |
| **GitHub Copilot** | `.github/copilot-instructions.md`, `.github/agents/` |
| **Google Gemini** | `GEMINI.md`, `.gemini/styleguide.md` |
| **Generic** | `AGENTS.md`, `.agents/rules/`, `.agents/tasks/` (tasks), `.agents/commands/` (commands) |

## Discovery Behavior

### File Types

The CLI processes:
- `.md` files (Markdown)
- `.mdc` files (Markdown component)

Other file types are ignored.

### Directory Processing

The CLI searches within each directory specified in search paths. It does not traverse parent directories automatically. Each directory is searched independently for the standard file patterns listed above.

**Example:**
```
Search paths:
1. /home/user/projects/myapp/backend/ (working directory, auto-added)
2. /home/user/ (home directory, auto-added)

Searches in /home/user/projects/myapp/backend/:
- .agents/rules/
- .agents/tasks/
- CLAUDE.md
- AGENTS.md
- etc.

Searches in /home/user/:
- .agents/rules/
- .claude/CLAUDE.md
- etc.
```

### Precedence Order

When multiple rule files exist across different directories:
1. Directories specified via `-d` flags (in order)
2. Working directory
3. Home directory

Within each directory, all matching files are included (unless filtered by selectors).

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

The `-C` option changes the working directory, which is automatically added to the search paths:

```bash
# Search from /path/to/project
coding-context -C /path/to/project fix-bug

# The working directory is automatically included, equivalent to:
coding-context -d file:///path/to/project fix-bug
```

The working directory is automatically included in search paths, so rules and tasks in that directory are discovered automatically.

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
coding-context -s languages=go fix-bug

# Include only planning rules
coding-context -s stage=planning plan-feature
```

## Examples

### Multi-Language Project

```
.agents/
└── rules/
    ├── general-standards.md        (no frontmatter - always included)
    ├── go-backend.md               (languages: [ go ])
    ├── python-ml.md                (languages: [ python ])
    └── javascript-frontend.md      (languages: [ javascript ])

Commands:
coding-context -s languages=go fix-bug
  → Includes: general-standards.md, go-backend.md

coding-context -s languages=python train-model
  → Includes: general-standards.md, python-ml.md

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`).
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
- Check that directories are in search paths (working directory and home directory are added automatically)
- Verify that `.agents/rules/` directory exists in one of the search path directories
- Verify files have `.md` or `.mdc` extension
- Check file permissions (must be readable)
- For remote directories, verify the download succeeded (check stderr logs)

**Task not found:**
- Verify that `.agents/tasks/` directory exists in one of the search path directories (working directory or home directory)
- Check that the filename (without `.md` extension) matches the task name you're using (e.g., `fix-bug.md` for `fix-bug`)
- Ensure filename has `.md` extension
- Verify the directory containing the task is in search paths (working directory and home directory are added automatically)
- Note: Tasks are matched by filename, not by `task_name` in frontmatter
- Note: Commands (in `.agents/commands/`, `.cursor/commands/`, `.opencode/command/`) are NOT tasks - they're referenced via slash commands inside task content

**Rules not filtered correctly:**
- Verify frontmatter YAML is valid
- Check selector spelling and capitalization
- Remember: only top-level frontmatter fields match

## See Also

- [File Formats Reference](./file-formats) - File specifications
- [CLI Reference](./cli) - Command-line options
- [How to Create Rules](../how-to/create-rules) - Organizing rules
