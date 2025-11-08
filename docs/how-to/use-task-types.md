---
layout: default
title: Use Task Types
---

# How to Use Task Types

This guide shows you how to use the task type pattern to apply different rules for different types of work (e.g., POC vs. production code, bug fixes vs. new features).

## What are Task Types?

Task types are a pattern for categorizing different kinds of work and applying appropriate rules for each. For example:

- **`poc`** - Proof of concept / prototypes (skip tests, move fast)
- **`bug-fix`** - Bug fixes (require regression tests)
- **`feature-implementation`** - New features (full test coverage, documentation)
- **`refactor`** - Code refactoring (maintain test coverage)
- **`research`** - Research tasks (documentation-focused)

## Basic Usage

### Step 1: Create Rules for Different Task Types

Create rule files that specify which task type they apply to:

**`.agents/rules/testing-poc.md`:**
```markdown
---
task_type: poc
---

# POC Testing Guidelines

For proofs of concept:
- Skip comprehensive test coverage
- Basic manual testing is sufficient
- Focus on demonstrating the concept works
```

**`.agents/rules/testing-bugfix.md`:**
```markdown
---
task_type: bug-fix
---

# Bug Fix Testing Guidelines

For bug fixes:
- MUST include regression tests
- Test should fail without the fix
- Test edge cases related to the bug
```

**`.agents/rules/testing-feature.md`:**
```markdown
---
task_type: feature-implementation
---

# Feature Testing Guidelines

For new features:
- Comprehensive unit test coverage (>80%)
- Integration tests for critical paths
- Document test cases in PR description
```

### Step 2: Create Tasks with Task Type Selectors

Create task files that automatically select rules based on task type:

**`.agents/tasks/create-poc.md`:**
```markdown
---
task_name: create-poc
selector:
  task_type: poc
---

# Create Proof of Concept

Build a quick prototype to validate the concept.
Move fast and skip comprehensive testing for now.
```

**`.agents/tasks/fix-bug.md`:**
```markdown
---
task_name: fix-bug
selector:
  task_type: bug-fix
---

# Fix Bug

Fix the reported bug with appropriate regression tests.
```

### Step 3: Run Tasks

When you run a task, it automatically selects the appropriate rules:

```bash
# POC task - uses POC testing rules
coding-context-cli create-poc

# Bug fix task - uses bug-fix testing rules
coding-context-cli fix-bug
```

## Combining Task Types with Languages

You can combine `task_type` with `language` for even more specific rule selection:

**`.agents/rules/go-bugfix-testing.md`:**
```markdown
---
language: Go
task_type: bug-fix
---

# Go Bug Fix Testing

Use Go table-driven tests for regression testing:

\```go
func TestBugFix_IssueXXX(t *testing.T) {
    tests := []struct {
        name string
        input string
        want string
    }{
        {name: "regression case", input: "bad", want: "good"},
    }
    // ...
}
\```
```

**`.agents/tasks/fix-go-bug.md`:**
```markdown
---
task_name: fix-go-bug
selector:
  language: Go
  task_type: bug-fix
---

# Fix Go Bug

Fix the bug following Go testing standards.
```

Now when you run `coding-context-cli fix-go-bug`, it will only include:
- Rules with both `language: Go` AND `task_type: bug-fix`
- Rules without any frontmatter (always included)

## Common Task Type Examples

### Research Task

**`.agents/tasks/research.md`:**
```markdown
---
task_name: research
selector:
  task_type: research
---

# Research Task

Research: ${research_topic}

Provide a comprehensive analysis with:
- Current state of the art
- Available options and tradeoffs
- Recommendations
```

**`.agents/rules/research-guidelines.md`:**
```markdown
---
task_type: research
---

# Research Guidelines

- Focus on documentation and analysis
- No code implementation required
- Cite sources and references
- Present options objectively
```

### Triage Task

**`.agents/tasks/triage-bug.md`:**
```markdown
---
task_name: triage-bug
selector:
  task_type: triage
---

# Triage Bug

Analyze issue #${issue_number} and determine:
- Severity and priority
- Root cause area
- Estimated effort
```

**`.agents/rules/triage-guidelines.md`:**
```markdown
---
task_type: triage
---

# Triage Guidelines

- Don't fix the bug, just analyze it
- Identify affected components
- Assess impact and risk
- Recommend priority level
```

## Overriding Task Type

You can override the task's task_type from the command line:

```bash
# Task defines task_type=poc, but override to use feature-implementation rules
coding-context-cli -s task_type=feature-implementation create-poc
```

This is useful when you want to "upgrade" a POC to production-ready code.

## Best Practices

1. **Define clear task types** - Use consistent task type names across your rules
2. **Document expectations** - Each task type's rules should clearly state what's expected
3. **Start with common types** - Begin with `poc`, `bug-fix`, and `feature-implementation`
4. **Combine with language** - Use both `task_type` and `language` for language-specific standards
5. **Keep rules focused** - Each rule file should address one concern (testing, documentation, etc.)

## Task Type Suggestions

Common task types to consider:

- **`poc`** - Quick prototypes, skip comprehensive tests
- **`bug-fix`** - Bug fixes, require regression tests
- **`feature-implementation`** - New features, full coverage
- **`refactor`** - Code improvements, maintain coverage
- **`research`** - Investigation and analysis
- **`triage`** - Bug analysis and prioritization
- **`migration`** - Large-scale code migrations
- **`documentation`** - Documentation updates
- **`security-fix`** - Security patches, extra scrutiny
- **`performance`** - Performance optimization
- **`experiment`** - Experimental changes

## Troubleshooting

**No rules are included:**
- Check that your rule files have the correct `task_type` in frontmatter
- Verify the task's `selector` field is properly formatted
- Remember that ALL selectors must match (AND logic)

**Wrong rules are included:**
- Use command-line selectors to override: `-s task_type=desired-type`
- Check for rules without frontmatter (they're always included)

**Task not found:**
- Ensure the `task_name` matches exactly
- Task selection happens before selector application

## See Also

- [File Formats Reference](../reference/file-formats) - Full specification of the `selector` field
- [Use Frontmatter Selectors](./use-selectors) - More on selector syntax
- [Create Tasks](./create-tasks) - How to create task files
- [Create Rules](./create-rules) - How to create rule files
