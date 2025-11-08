# Task Type Selector Examples

This directory contains examples demonstrating how to use task type selectors to provide different rules and guidelines for different kinds of work.

## Task Types

Different tasks require different approaches:

- **`poc`** - Proof of concepts and prototypes
- **`production`** - Production-ready code
- **`research`** - Technical research and investigation
- **`review`** - Code review tasks
- **`bugfix`** - Bug fixes (can use production or specific bugfix rules)

## Example Usage

### Proof of Concept

When building a quick POC to validate an approach:

```bash
coding-context-cli -s task_type=poc poc-feature
```

This will:
- ✅ Include POC guidelines (no tests, no comments, focus on speed)
- ✅ Include planning guidelines (general, always included)
- ❌ Exclude production standards
- ❌ Exclude testing requirements
- ❌ Exclude implementation standards

### Production Feature

When implementing production-ready code:

```bash
coding-context-cli -s task_type=production implement-feature
```

This will:
- ✅ Include production standards (comprehensive tests, documentation)
- ✅ Include testing requirements
- ✅ Include implementation standards
- ✅ Include planning guidelines (general, always included)
- ❌ Exclude POC guidelines
- ❌ Exclude research guidelines

### Research

When conducting technical research:

```bash
coding-context-cli -s task_type=research research-topic
```

This will:
- ✅ Include research guidelines (thorough documentation, comparisons)
- ✅ Include planning guidelines (general, always included)
- ❌ Exclude production standards
- ❌ Exclude POC guidelines
- ❌ Exclude testing requirements

### Code Review

When reviewing code:

```bash
coding-context-cli -s task_type=review review-code
```

This will:
- ✅ Include review guidelines (constructive feedback, categorized issues)
- ✅ Include planning guidelines (general, always included)
- ❌ Exclude implementation standards
- ❌ Exclude production standards

## File Structure

```
examples/agents/
├── rules/
│   ├── implementation-go.md          # task_type: production, stage: implementation, language: go
│   ├── testing.md                    # task_type: production, stage: testing
│   ├── production-standards.md       # task_type: production
│   ├── poc-guidelines.md             # task_type: poc
│   ├── research-guidelines.md        # task_type: research
│   ├── review-guidelines.md          # task_type: review
│   └── planning.md                   # No task_type (applies to all)
└── tasks/
    ├── poc-feature.md                # task_name: poc-feature, task_type: poc
    ├── production-feature.md         # task_name: implement-feature, task_type: production
    ├── research-task.md              # task_name: research-topic, task_type: research
    ├── review-code.md                # task_name: review-code, task_type: review
    ├── fix-bug.md                    # task_name: fix-bug, resume: false
    ├── fix-bug-resume.md             # task_name: fix-bug, resume: true
    ├── plan-feature.md               # task_name: plan-feature
    └── code-review.md                # task_name: code-review
```

## How It Works

### Rule Filtering

Rules with a `task_type` field in their frontmatter will only be included when that task type is selected:

```markdown
---
task_type: production
---
# Production Standards

These rules only apply to production tasks.
```

Rules without a `task_type` field are always included (unless filtered by other selectors):

```markdown
---
# No task_type specified
---
# General Guidelines

These apply to all tasks.
```

### Task Selection

Tasks can specify their type in frontmatter:

```markdown
---
task_name: poc-feature
task_type: poc
---
# POC Feature

Quick proof of concept guidelines.
```

Then invoke with the matching selector:

```bash
coding-context-cli -s task_type=poc poc-feature
```

### Multiple Selectors

You can combine multiple selectors:

```bash
# Production Go implementation
coding-context-cli -s task_type=production -s language=go -s stage=implementation implement-feature
```

This allows fine-grained control over which rules are included.

## Benefits

1. **Speed for POCs**: Skip unnecessary overhead like tests and documentation when doing quick explorations
2. **Quality for Production**: Enforce high standards when building production code
3. **Flexibility**: Use the same tool for different workflows
4. **Maintainability**: Clear separation of concerns in rule files
5. **Token Efficiency**: Only send relevant context to AI models

## See Also

- [Main README](../../README.md)
- [Agentic Workflows Guide](../../AGENTIC_WORKFLOWS.md)
