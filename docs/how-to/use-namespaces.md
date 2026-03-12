---
layout: default
title: Use Namespaces
parent: How-to Guides
nav_order: 4
---

# How to Use Namespaces

Namespaces let multiple teams share a single `.agents/` directory without conflicts. Each team's tasks, rules, commands, and skills live under their own subdirectory and take precedence over global assets, while still inheriting the shared global layer.

## When to Use Namespaces

Use namespaces when:
- Multiple teams share a repository and each needs distinct rules, bootstraps, or tasks
- You want to prevent one team's selectors or context from leaking into another's
- A team needs to override a global command or skill with their own version

If you're a single team, namespaces are not required — the global `.agents/` layout works as-is.

## Directory Structure

```
.agents/
├── tasks/                        # Global tasks (no namespace)
│   └── fix-bug.md
├── rules/                        # Global rules (always included)
│   └── coding-standards.md
├── commands/                     # Global commands
│   └── deploy.md
├── skills/                       # Global skills
│   └── data-analysis/
│       └── SKILL.md
└── namespaces/                   # Namespace root
    ├── myteam/
    │   ├── tasks/                # Tasks accessed as "myteam/<name>"
    │   │   └── build.md
    │   ├── rules/                # Namespace-specific rules (included first)
    │   │   └── team-rules.md
    │   ├── commands/             # Can override global commands
    │   │   └── deploy.md        # Shadows the global deploy command
    │   └── skills/
    │       └── special-tool/
    │           └── SKILL.md
    └── otherteam/
        ├── tasks/
        ├── rules/
        ├── commands/
        └── skills/
```

Only the generic `.agents/` structure supports namespacing. Agent-specific paths (`.cursor/`, `.claude/`, `.github/`, etc.) are not namespaced.

## Task Name Format

Select a namespace by using `namespace/taskname` as the task argument:

| Task argument | Namespace | Base task name |
|---|---|---|
| `fix-bug` | _(none — global)_ | `fix-bug` |
| `myteam/fix-bug` | `myteam` | `fix-bug` |

Only a single level of namespacing is supported. `myteam/fix-bug` is valid; `myteam/subteam/fix-bug` is an error.

```bash
# Global task (existing behaviour, unchanged)
coding-context fix-bug

# Namespaced task
coding-context myteam/fix-bug

# Namespaced task with parameters
coding-context -p issue=BUG-123 myteam/fix-bug
```

## Resolution Rules

| Asset | No namespace | With namespace |
|---|---|---|
| **Task** | `.agents/tasks/<name>.md` | `.agents/namespaces/<ns>/tasks/<name>.md` first; falls back to global |
| **Rules** | `.agents/rules/` only | Namespace rules **first**, then **all** global rules (both always included) |
| **Commands** | `.agents/commands/` | Namespace commands searched first; first match wins |
| **Skills** | `.agents/skills/` | Both namespace and global skills discovered; namespace skills listed first |

## Quick Start

### 1. Create the namespace directory structure

```bash
mkdir -p .agents/namespaces/myteam/{tasks,rules,commands,skills}
```

### 2. Create a namespaced task

```markdown
# .agents/namespaces/myteam/tasks/build.md

Build the myteam service using our internal pipeline.
```

### 3. Create a namespace-specific rule (optional)

```markdown
# .agents/namespaces/myteam/rules/team-standards.md
---
name: myteam-standards
---

# myteam Coding Standards

Always prefix internal service calls with `svc.`.
```

### 4. Run the namespaced task

```bash
coding-context myteam/build
```

The assembled context will include:
1. Rules from `.agents/namespaces/myteam/rules/` (namespace rules, first)
2. Rules from `.agents/rules/` (global rules, always included)
3. Skills from both namespace and global directories
4. The task from `.agents/namespaces/myteam/tasks/build.md`

## Falling Back to Global Tasks

If a task doesn't exist in the namespace directory, the tool falls back to the global task directory automatically:

```
.agents/tasks/common-task.md              # global task
.agents/namespaces/myteam/tasks/          # no common-task.md here

$ coding-context myteam/common-task      # resolves to .agents/tasks/common-task.md
```

This allows namespaces to selectively override tasks without having to duplicate every task.

## Overriding Global Commands

A namespace command with the same name as a global command takes precedence:

```
.agents/commands/deploy.md                        # global deploy
.agents/namespaces/myteam/commands/deploy.md      # myteam deploy (overrides global)

# When running any myteam/* task, /deploy expands the myteam version
```

## Scoping Rules to a Namespace

Global rules are always included for namespaced tasks. If a global rule should only apply to a specific namespace, add `namespace: <value>` to its frontmatter using the existing selector system:

```markdown
---
# .agents/rules/myteam-only-rule.md
namespace: myteam
---

# myteam Internal Requirements

Only apply this rule for myteam tasks.
```

This rule will be included when running `myteam/*` tasks and excluded for all other namespaces or global tasks, because the `namespace` selector is automatically set based on the task name.

Rules with **no** `namespace` frontmatter field are always included regardless of namespace.

## The `namespace` Selector

When a namespaced task is run, the tool automatically injects `namespace=<value>` into the selector set. This means:

- Rules with `namespace: myteam` in frontmatter are included only for `myteam/*` tasks
- Rules with no `namespace` field are always included (the existing behaviour)
- For non-namespaced tasks (e.g., `fix-bug`), the selector is set to `namespace=""`, so rules with any explicit `namespace:` value are excluded

You can observe this with the lint command:

```bash
# No namespace selector errors
coding-context lint myteam/build
```

## Team Isolation Example

Two teams sharing a repository with no overlap:

```
.agents/
├── rules/
│   └── company-wide.md           # Always included for everyone
└── namespaces/
    ├── backend/
    │   ├── tasks/
    │   │   └── deploy-service.md
    │   └── rules/
    │       └── go-standards.md   # Only included for backend/* tasks
    └── frontend/
        ├── tasks/
        │   └── deploy-app.md
        └── rules/
            └── ts-standards.md   # Only included for frontend/* tasks
```

```bash
# Backend team runs their tasks — gets company-wide.md + go-standards.md
coding-context backend/deploy-service

# Frontend team runs their tasks — gets company-wide.md + ts-standards.md
coding-context frontend/deploy-app
```

If backend wanted to restrict their rule further to _only_ backend tasks (in case the file path isn't already enough), they could add `namespace: backend` to the frontmatter.

## Error Cases

| Input | Error |
|---|---|
| `myteam/subteam/build` | Only one level of namespacing is supported |
| `/build` | Namespace must not be empty |
| `myteam/` | Task base name must not be empty |

## See Also

- [Create Task Files](./create-tasks) - General task file documentation
- [Use Frontmatter Selectors](./use-selectors) - How selectors work, including the `namespace` selector
- [Search Paths Reference](../reference/search-paths) - Complete path resolution reference
