---
layout: default
title: Use Frontmatter Selectors
---

# How to Use Frontmatter Selectors

This guide shows you how to use frontmatter selectors to filter which rules and tasks are included.

## Selector Operators

Selectors support multiple operators for different matching behaviors:

| Operator | Name | Description | Example |
|----------|------|-------------|---------|
| `=` | Equals | Exact match (default) | `language=Go` |
| `:=` | Includes | Value in array or equals scalar | `language:=TypeScript` |
| `!=` | Not Equals | Value does not equal | `env!=staging` |
| `!:` | Not Includes | Value not in array | `language!:Python` |

### Equals (`=`)

Matches when the frontmatter value exactly equals the selector value.

```bash
# Include only Go rules
coding-context-cli -s language=Go fix-bug
```

**Frontmatter:**
```yaml
---
language: Go
---
```

### Includes (`:=`)

Matches when:
- The frontmatter value is a scalar and equals the selector value, OR
- The frontmatter value is an array and contains the selector value

This is useful for rules that apply to multiple languages.

```bash
# Include rules for TypeScript (works with both scalar and array)
coding-context-cli -s language:=TypeScript implement-feature
```

**Frontmatter examples that match:**
```yaml
---
# Scalar match
language: TypeScript
---
```

```yaml
---
# Array match
language:
  - TypeScript
  - JavaScript
---
```

### Not Equals (`!=`)

Matches when the frontmatter value does not equal the selector value. Useful for excluding specific values.

```bash
# Exclude staging-specific rules
coding-context-cli -s env!=staging deploy
```

**Frontmatter:**
```yaml
---
env: production  # Matches (not staging)
---
```

### Not Includes (`!:`)

Matches when:
- The frontmatter value is a scalar and does not equal the selector value, OR
- The frontmatter value is an array and does not contain the selector value

```bash
# Exclude Python rules
coding-context-cli -s language!:Python implement-feature
```

**Frontmatter examples that match:**
```yaml
---
language: Go  # Matches (not Python)
---
```

```yaml
---
language:
  - Go
  - TypeScript  # Matches (Python not in array)
---
```

**Frontmatter examples that don't match:**
```yaml
---
language: Python  # Doesn't match (is Python)
---
```

```yaml
---
language:
  - Python
  - Go  # Doesn't match (Python in array)
---
```

## Basic Selector Usage

Include only rules matching a specific frontmatter field:

```bash
# Include only Go rules
coding-context-cli -s language=Go fix-bug
```

This includes only rules with `language: Go` in their frontmatter.

## Multiple Selectors (AND Logic)

Combine multiple selectors - all must match:

```bash
# Include only Go testing rules
coding-context-cli -s language=Go -s stage=testing implement-feature
```

This includes only rules with BOTH `language: Go` AND `stage: testing`.

You can also combine different operators:

```bash
# Include TypeScript rules but exclude testing stage
coding-context-cli -s language:=TypeScript -s stage!=testing implement-feature
```

## Selecting Tasks

Use selectors to choose between multiple task files with the same `task_name`:

**Staging task (`.agents/tasks/deploy-staging.md`):**
```markdown
---
task_name: deploy
environment: staging
---
Deploy to staging environment.
```

**Production task (`.agents/tasks/deploy-production.md`):**
```markdown
---
task_name: deploy
environment: production
---
Deploy to production environment.
```

**Usage:**
```bash
# Select staging task
coding-context-cli -s environment=staging deploy

# Select production task
coding-context-cli -s environment=production deploy
```

## Common Selector Patterns

### By Language

**Single language (exact match):**
```bash
# Python project
coding-context-cli -s language=Python fix-bug

# JavaScript project
coding-context-cli -s language=JavaScript code-review
```

**Multiple languages (using includes operator):**
```bash
# Include rules that support TypeScript (either as scalar or in array)
coding-context-cli -s language:=TypeScript implement-feature

# Include rules that support Go
coding-context-cli -s language:=Go implement-backend
```

**Example rule for multiple languages:**
```markdown
---
language:
  - TypeScript
  - JavaScript
---
# Web Development Standards

Standards that apply to both TypeScript and JavaScript.
```

**Excluding languages:**
```bash
# Exclude Python rules
coding-context-cli -s language!=Python fix-bug

# Exclude rules with Python in them (scalar or array)
coding-context-cli -s language!:Python implement-feature
```

### By Stage

```bash
# Planning phase
coding-context-cli -s stage=planning plan-feature

# Implementation phase
coding-context-cli -s stage=implementation implement-feature

# Testing phase
coding-context-cli -s stage=testing test-feature

# Everything except testing
coding-context-cli -s stage!=testing implement-feature
```

### By Priority

```bash
# High priority rules only
coding-context-cli -s priority=high fix-critical-bug

# Exclude low priority rules
coding-context-cli -s priority!=low fix-bug

# Include all priorities (no selector)
coding-context-cli fix-bug
```

### By Source

```bash
# Include JIRA context
coding-context-cli -s source=jira fix-bug

# Include GitHub context
coding-context-cli -s source=github code-review
```

## Resume Mode

The `-r` flag is shorthand for `-s resume=true` plus skipping all rules:

```bash
# These are equivalent:
coding-context-cli -r fix-bug
coding-context-cli -s resume=true fix-bug  # but also skips rules
```

Use resume mode when continuing work in a new session to save tokens.

## Understanding Selector Matching

**Rules are included if:**
- No selectors are specified (all rules included), OR
- All specified selectors match the rule's frontmatter

**Tasks are selected by:**
- Matching `task_name` (required)
- Matching all selectors (if specified)

**Important limitations:**
- Only top-level frontmatter fields can be matched
- Nested fields (e.g., `metadata.version`) are NOT supported
- Selector values must match exactly (case-sensitive)

## Examples with Rules

**Rule with multiple frontmatter fields:**
```markdown
---
language: Go
stage: testing
priority: high
team: backend
---

# Go Backend Testing Standards
...
```

**Matching selectors:**
```bash
# Matches
coding-context-cli -s language=Go fix-bug
coding-context-cli -s language=Go -s stage=testing fix-bug
coding-context-cli -s priority=high fix-bug

# Does NOT match
coding-context-cli -s language=Python fix-bug
coding-context-cli -s language=Go -s stage=planning fix-bug
```

## Debugging Selectors

Check which rules are included:

```bash
# Output to file and review
coding-context-cli -s language=Go fix-bug > output.txt
less output.txt

# Check token count
coding-context-cli -s language=Go fix-bug 2>&1 | grep -i token
```

## Best Practices

1. **Use consistent naming**: Standardize frontmatter field names across rules
2. **Be specific when needed**: Use multiple selectors for fine-grained control
3. **Document your selectors**: Note which selectors rules support
4. **Test your selectors**: Verify the expected rules are included
5. **Use language names correctly**: Follow GitHub Linguist capitalization

## Troubleshooting

**No rules included:**
- Check frontmatter spelling and capitalization
- Verify selectors match rule frontmatter exactly
- Remember: All selectors must match (AND logic)

**Wrong rules included:**
- Check for rules without frontmatter (always included if no selectors)
- Verify unique frontmatter values across rules

**Task not found:**
- Ensure `task_name` matches exactly
- Check that selectors don't over-filter (try without selectors)

## See Also

- [Creating Rules](./create-rules) - Add frontmatter to rules
- [Creating Tasks](./create-tasks) - Use selectors in tasks
- [File Formats Reference](../reference/file-formats) - Frontmatter specification
