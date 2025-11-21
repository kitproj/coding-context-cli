---
layout: default
title: Create Task Files
parent: How-to Guides
nav_order: 1
---

# How to Create Task Files

This guide shows you how to create task files for different scenarios.

## Basic Task File

Create a simple task without parameters:

```markdown
---
task_name: code-review
---
# Code Review Task

Please review the code changes with focus on:
- Code quality
- Test coverage
- Security implications
```

Save as `.agents/tasks/code-review.md`.

Use with:
```bash
coding-context code-review
```

## Task with Parameters

Create a task that accepts dynamic values:

```markdown
---
task_name: implement-feature
---
# Feature Implementation: ${feature_name}

Implement the following feature: ${feature_name}

## Requirements
- ${requirements}

## Success Criteria
- ${success_criteria}
```

Use with:
```bash
coding-context \
  -p feature_name="User Authentication" \
  -p requirements="OAuth2 support, secure password storage" \
  -p success_criteria="All tests pass, security audit clean" \
  implement-feature
```

## Multiple Tasks with Selectors

Create multiple variations of the same task using selectors:

**For staging environment (`.agents/tasks/deploy-staging.md`):**
```markdown
---
task_name: deploy
environment: staging
---
# Deploy to Staging

Deploy with extra validation and monitoring.
```

**For production environment (`.agents/tasks/deploy-production.md`):**
```markdown
---
task_name: deploy
environment: production
---
# Deploy to Production

Deploy with all safety checks and rollback plan.
```

Use with:
```bash
# Deploy to staging
coding-context -s environment=staging deploy

# Deploy to production
coding-context -s environment=production deploy
```

## Resume Mode Tasks

Create separate tasks for initial and resume sessions:

**Initial task (`.agents/tasks/refactor-initial.md`):**
```markdown
---
task_name: refactor
resume: false
---
# Refactoring Task

Analyze the code and create a refactoring plan.
```

**Resume task (`.agents/tasks/refactor-resume.md`):**
```markdown
---
task_name: refactor
resume: true
---
# Continue Refactoring

Continue with the refactoring work from your previous session.
```

Use with:
```bash
# Initial session
coding-context -s resume=false refactor

# Resume session (uses -r flag to skip rules and select resume task)
coding-context -r refactor
```

## Tasks with Embedded Selectors

Instead of requiring `-s` flags on every invocation, you can embed selectors directly in the task frontmatter. This is useful for tasks that always need specific rules.

**Example (`.agents/tasks/implement-go-feature.md`):**
```markdown
---
task_name: implement-feature
selectors:
  language: Go
  stage: implementation
---
# Implement Feature in Go

Implement the feature following Go best practices and implementation guidelines.

Feature name: ${feature_name}
Requirements: ${requirements}
```

**Usage:**
```bash
# Automatically applies language=Go and stage=implementation selectors
coding-context -p feature_name="User Auth" implement-feature
```

**Example with OR logic using arrays:**
```markdown
---
task_name: write-tests
selectors:
  language: [Go, Python]
  stage: testing
---
# Write Tests

Write comprehensive tests for the code.
```

This matches rules where `(language=Go OR language=Python) AND stage=testing`.

**Combining embedded and command-line selectors:**
```bash
# Task has: selectors.language = Go
# Command adds: -s priority=high
# Result: Includes rules matching language=Go AND priority=high
coding-context -s priority=high implement-feature
```

## Task Frontmatter

Task frontmatter is automatically included in the output. This is useful when downstream tools need access to task metadata.

**Example:**
```bash
coding-context implement-feature
```

**Output:**
```yaml
---
task_name: implement-feature
selectors:
  language: Go
  stage: implementation
---
# Implement Feature in Go
...
```

## Best Practices

1. **Use descriptive task names**: Make them clear and specific
2. **Include clear instructions**: Be explicit about what the AI should do
3. **Use parameters for dynamic content**: Don't hardcode values that change
4. **Organize by purpose**: Keep related tasks together
5. **Document your tasks**: Add comments explaining complex requirements

## See Also

- [File Formats Reference](../reference/file-formats) - Technical specification
- [Using Selectors](./use-selectors) - Filter tasks and rules
- [Creating Rules](./create-rules) - Create reusable context
