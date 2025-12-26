---
layout: default
title: Create Rule Files
parent: How-to Guides
nav_order: 2
---

# How to Create Rule Files

This guide shows you how to create rule files that provide reusable context to AI agents.

## Basic Rule File

Create a simple rule without frontmatter:

```markdown
# General Coding Standards

- Write clear, readable code
- Add comments for complex logic
- Follow the project's style guide
- Write tests for new functionality
```

Save as `.agents/rules/general-standards.md`.

This rule will be included in all context assemblies.

## Language-Specific Rules

Create rules that apply only to specific programming languages:

**Go standards (`.agents/rules/go-standards.md`):**
```markdown
---
languages:
  - go
---

# Go Coding Standards

- Use `gofmt` for formatting
- Handle all errors explicitly
- Write table-driven tests
- Use meaningful package names
```

**Python standards (`.agents/rules/python-standards.md`):**
```markdown
---
languages:
  - python
---

# Python Coding Standards

- Follow PEP 8
- Use type hints
- Write docstrings
- Use pytest for testing
```

Use with:
```bash
# Include only Go rules
coding-context -s languages=go fix-bug

# Include only Python rules
coding-context -s languages=python fix-bug
```

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`).

## Rules with Multiple Selectors

Create rules with multiple frontmatter fields for fine-grained filtering:

```markdown
---
languages:
  - go
stage: testing
priority: high
---

# Go Testing Best Practices

When writing tests:
- Use table-driven tests
- Test edge cases
- Mock external dependencies
- Aim for >80% coverage
```

Use with:
```bash
# Include rules for Go testing
coding-context -s stage=testing implement-feature
```

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`).

## Stage-Specific Rules

Create rules for different workflow stages:

**Planning rules (`.agents/rules/planning-guidelines.md`):**
```markdown
---
stage: planning
---

# Planning Guidelines

- Break down features into small tasks
- Identify dependencies
- Consider edge cases
- Document assumptions
```

**Implementation rules (`.agents/rules/implementation-guidelines.md`):**
```markdown
---
stage: implementation
---

# Implementation Guidelines

- Follow coding standards
- Write tests alongside code
- Keep commits small and focused
- Update documentation
```

Use with:
```bash
# Planning phase
coding-context -s stage=planning plan-feature

# Implementation phase
coding-context -s stage=implementation implement-feature
```

## Rules with Bootstrap Scripts

Create rules that fetch dynamic context:

**Rule file (`.agents/rules/jira-context.md`):**
```markdown
---
source: jira
---

# JIRA Context

Issue details are fetched by the bootstrap script.
```

**Bootstrap script (`.agents/rules/jira-context-bootstrap`):**
```bash
#!/bin/bash
# Make this executable: chmod +x jira-context-bootstrap

if [ -z "$JIRA_ISSUE_KEY" ]; then
    exit 0
fi

echo "Fetching JIRA issue: $JIRA_ISSUE_KEY" >&2

# Fetch and process JIRA data
curl -s -H "Authorization: Bearer $JIRA_API_TOKEN" \
    "https://your-domain.atlassian.net/rest/api/3/issue/${JIRA_ISSUE_KEY}" \
    | jq -r '.fields | {summary, description}' \
    > /tmp/jira-context.json
```

Use with:
```bash
export JIRA_ISSUE_KEY="PROJ-123"
export JIRA_API_TOKEN="your-token"

coding-context -s source=jira fix-bug
```

## Best Practices

1. **Keep rules focused**: Each rule should address one concern
2. **Use frontmatter selectors**: Make rules conditionally includable
3. **Use lowercase language values**: Language values should be lowercase (e.g., `go`, `python`, `javascript`)
4. **Organize by category**: Group related rules together
5. **Update rules as standards evolve**: Keep them current

## Common Language Values

Use lowercase values in your `languages:` frontmatter:

- c, c++, c#, css
- dart, elixir, go
- haskell, html
- java, javascript
- kotlin, lua
- markdown, objective-c
- php, python
- ruby, rust
- scala, shell, swift
- typescript, yaml

**Note:** Language values should be lowercase (e.g., `go`, `python`, `javascript`).

## See Also

- [File Formats Reference](../reference/file-formats) - Technical specification
- [Using Selectors](./use-selectors) - Filter rules effectively
- [Search Paths Reference](../reference/search-paths) - Where rules are found
