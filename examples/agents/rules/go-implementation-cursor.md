---
task_name: implement-feature
language: Go
agent: cursor
---

# Go Implementation Standards for Feature Development

This rule demonstrates standard frontmatter fields for rules:

- **task_name**: `implement-feature` - This rule only applies to the `implement-feature` task
- **language**: `Go` - This rule only applies when the language is Go
- **agent**: `cursor` - This rule is optimized for the Cursor AI agent

## When This Rule Is Included

This rule will be included when:
1. The task being run is `implement-feature` (or has `task_name: implement-feature` selector)
2. AND the task has `language: Go` (or `-s language=Go` is specified)
3. AND the task has `agent: cursor` (or `-a cursor` is specified)

## Go-Specific Implementation Guidelines

When implementing features in Go:
- Follow Go idioms and conventions
- Use table-driven tests
- Handle errors explicitly
- Keep functions small and focused
- Use interfaces for abstraction
