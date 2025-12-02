---
layout: default
title: Use Frontmatter Selectors
parent: How-to Guides
nav_order: 3
---

# How to Use Frontmatter Selectors

This guide shows you how to use frontmatter selectors to filter which rules and tasks are included.

## Basic Selector Usage

Include only rules matching a specific frontmatter field:

```bash
# Include only Go rules
coding-context -s languages=go /fix-bug
```

This includes only rules with `languages: [ go ]` in their frontmatter.

## Multiple Selectors (AND Logic)

Combine multiple selectors - all must match:

```bash
# Include only Go testing rules
coding-context -s languages=go -s stage=testing /implement-feature
```

This includes only rules with BOTH `languages: [ go ]` AND `stage: testing`.

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`).

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
coding-context -s environment=staging /deploy

# Select production task
coding-context -s environment=production /deploy
```

## Common Selector Patterns

### By Language

```bash
# Python project
coding-context -s languages=python /fix-bug

# JavaScript project
coding-context -s languages=javascript /code-review

# Multi-language (run separately)
coding-context -s languages=go /implement-backend
coding-context -s languages=javascript /implement-frontend
```

### By Stage

```bash
# Planning phase
coding-context -s stage=planning /plan-feature

# Implementation phase
coding-context -s stage=implementation /implement-feature

# Testing phase
coding-context -s stage=testing /test-feature
```

### By Priority

```bash
# High priority rules only
coding-context -s priority=high /fix-critical-bug

# Include all priorities (no selector)
coding-context /fix-bug
```

### By Source

```bash
# Include JIRA context
coding-context -s source=jira /fix-bug

# Include GitHub context
coding-context -s source=github /code-review
```

## Resume Mode

The `-r` flag is shorthand for `-s resume=true` plus skipping all rules:

```bash
# These are equivalent:
coding-context -r /fix-bug
coding-context -s resume=true /fix-bug  # but also skips rules
```

Use resume mode when continuing work in a new session to save tokens.

## Task Frontmatter Selectors

Instead of specifying selectors on the command line every time, you can embed them directly in task files using the `selectors` field.

### Basic Task Selectors

**Task file (`.agents/tasks/implement-go-feature.md`):**
```markdown
---
task_name: implement-feature
selectors:
  languages: go
  stage: implementation
---
# Implement Feature in Go
...
```

**Usage:**
```bash
# Automatically applies language=go and stage=implementation
coding-context /implement-feature
```

This is equivalent to:
```bash
coding-context -s languages=go -s stage=implementation /implement-feature
```

### Array Selectors (OR Logic)

Use arrays for OR logic within the same selector key:

**Task file:**
```markdown
---
task_name: refactor-code
selectors:
  language: [go, python, javascript]
  stage: refactoring
---
```

**Usage:**
```bash
# Includes rules matching (go OR python OR javascript) AND refactoring
coding-context /refactor-code
```

### Combining Command-Line and Task Selectors

Selectors from task frontmatter and the command line are combined (additive):

**Task file with embedded selectors:**
```markdown
---
task_name: deploy
selectors:
  stage: deployment
---
```

**Usage:**
```bash
# Combines task selectors with command-line selectors
# Result: stage=deployment AND environment=production
coding-context -s environment=production /deploy
```

### When to Use Task Frontmatter Selectors

**Use task frontmatter selectors when:**
- A task always needs specific rules (e.g., language-specific tasks)
- You want to simplify command-line invocations
- The selectors are intrinsic to the task's purpose

**Use command-line selectors when:**
- Selectors vary between invocations
- You need runtime flexibility
- Multiple users run the same task differently

### Viewing Task Frontmatter

Task frontmatter (including selectors) is automatically included in the output:

```bash
coding-context /implement-feature
```

**Output:**
```yaml
---
task_name: implement-feature
selectors:
  languages: go
  stage: implementation
---
# Task content...
```

## Understanding Selector Matching

**Rules are included if:**
- No selectors are specified (all rules included), OR
- All specified selectors match the rule's frontmatter

**Tasks are selected by:**
- Matching filename (without `.md` extension)
- Matching all selectors (if specified)

**Note:** Tasks are matched by filename, not by `task_name` in frontmatter. The `task_name` field is optional metadata.

**Important limitations:**
- Only top-level frontmatter fields can be matched
- Nested fields (e.g., `metadata.version`) are NOT supported
- Selector values must match exactly (case-sensitive)

## Examples with Rules

**Rule with multiple frontmatter fields:**
```markdown
---
languages:
  - go
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
coding-context -s languages=go /fix-bug
coding-context -s languages=go -s stage=testing /fix-bug
coding-context -s priority=high /fix-bug

# Does NOT match
coding-context -s languages=python /fix-bug
coding-context -s languages=go -s stage=planning /fix-bug
```

## Debugging Selectors

Check which rules are included:

```bash
# Output to file and review
coding-context -s languages=go /fix-bug > output.txt
less output.txt

# Check token count
coding-context -s languages=go /fix-bug 2>&1 | grep -i token
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
