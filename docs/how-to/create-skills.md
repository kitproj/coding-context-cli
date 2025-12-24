---
layout: default
title: Create Skills
parent: How-To Guides
nav_order: 5
---

# How to Create Skills

This guide shows you how to create skill files that provide reusable capabilities for AI agents.

## What are Skills?

Skills are reusable capability definitions that AI agents can leverage during task execution. Unlike rules (which provide context and guidelines), skills define specific capabilities, methodologies, or domain expertise that agents can use.

Skills follow the format described at [agentskills.io](https://agentskills.io) and are output in XML-like format for easy reference.

## Basic Skill Structure

Skills are stored in `.agents/skills/(skill-name)/SKILL.md`:

```
.agents/skills/
├── code-review/
│   └── SKILL.md
├── debugging/
│   └── SKILL.md
└── testing/
    └── SKILL.md
```

## Creating Your First Skill

### 1. Create the Directory Structure

```bash
mkdir -p .agents/skills/code-review
```

### 2. Create the Skill File

Create `.agents/skills/code-review/SKILL.md`:

```markdown
---
skill_name: code-review
languages:
  - go
  - python
  - javascript
---

# Code Review Skill

This skill provides comprehensive code review capabilities including:

## Capabilities

- **Style Checking**: Verify code follows established style guides
- **Security Analysis**: Identify potential security vulnerabilities
- **Performance Review**: Suggest performance optimizations
- **Best Practices**: Ensure code follows language-specific best practices

## Usage

When performing code reviews, consider:

1. **Readability**: Is the code easy to understand?
2. **Maintainability**: Can it be easily modified?
3. **Test Coverage**: Are there adequate tests?
4. **Error Handling**: Are errors properly handled?
5. **Documentation**: Is the code well-documented?

## Review Checklist

- [ ] Code follows style guide
- [ ] No security vulnerabilities
- [ ] Adequate test coverage
- [ ] Clear error handling
- [ ] Documentation is complete
```

### 3. Test the Skill

```bash
coding-context -s languages=go code-review-task
```

The skill will be included in a dedicated "Skills" section:

```xml
# Skills

The following skills are available for use in this task...

<skill name="code-review">
# Code Review Skill
...
</skill>
```

## Skill Selection with Frontmatter

### Task-Specific Skills

Create a skill that only applies to specific tasks:

```markdown
---
skill_name: debugging
task_names:
  - fix-bug
  - troubleshoot
---

# Debugging Skill

Systematic approach to finding and fixing bugs...
```

This skill will only be included when running `fix-bug` or `troubleshoot` tasks.

### Language-Specific Skills

Create skills for specific programming languages:

```markdown
---
skill_name: go-concurrency
languages:
  - go
---

# Go Concurrency Skill

Best practices for Go goroutines and channels...
```

Use it:
```bash
coding-context -s languages=go implement-feature
```

### Agent-Specific Skills

Create skills optimized for specific AI agents:

```markdown
---
skill_name: cursor-shortcuts
agent: cursor
---

# Cursor IDE Skill

Keyboard shortcuts and features specific to Cursor...
```

## Common Skill Types

### 1. Domain Expertise Skills

```markdown
---
skill_name: security-review
---

# Security Review Skill

## Common Vulnerabilities

- SQL Injection
- Cross-Site Scripting (XSS)
- Authentication Issues
...
```

### 2. Methodology Skills

```markdown
---
skill_name: tdd-approach
---

# Test-Driven Development Skill

## TDD Cycle

1. Write a failing test
2. Write minimal code to pass
3. Refactor
4. Repeat
```

### 3. Tool Usage Skills

```markdown
---
skill_name: git-workflow
---

# Git Workflow Skill

## Branch Strategy

- `main`: Production-ready code
- `develop`: Integration branch
- `feature/*`: New features
...
```

## Skill vs Rule Guidelines

Use **Skills** for:
- Reusable capabilities agents can apply
- Methodologies and approaches
- Tool-specific instructions
- Domain expertise (security, testing, etc.)

Use **Rules** for:
- Project-specific context
- Coding standards and style guides
- Architecture decisions
- Team conventions

## Best Practices

### 1. Clear Skill Names

```yaml
# ✅ Good - descriptive name
skill_name: code-review

# ❌ Bad - too generic
skill_name: review
```

### 2. Structured Content

Organize skills with clear sections:
- Overview/Purpose
- Capabilities
- Usage instructions
- Examples
- Checklists

### 3. Focused Scope

Each skill should cover one area of expertise:

```markdown
# ✅ Good - focused on one capability
skill_name: security-analysis

# ❌ Bad - trying to do too much
skill_name: all-code-quality-checks
```

### 4. Use Selectors Wisely

Add selectors to limit when skills are included:

```yaml
# Include only for specific tasks
task_names:
  - security-review
  - penetration-test

# Include only for specific languages
languages:
  - go
  - rust
```

## Example: Creating a Testing Skill

Complete example of a comprehensive testing skill:

**File:** `.agents/skills/testing/SKILL.md`

```markdown
---
skill_name: testing
languages:
  - go
  - python
  - javascript
---

# Testing Skill

Comprehensive testing approach for software development.

## Test Types

### Unit Tests
- Test individual functions/methods
- Fast execution
- No external dependencies

### Integration Tests
- Test component interactions
- May use test databases
- Slower than unit tests

### End-to-End Tests
- Test complete workflows
- Use real or staging environments
- Slowest but most comprehensive

## Testing Strategy

1. **Start with unit tests** - Cover all pure functions
2. **Add integration tests** - Test component boundaries
3. **Include edge cases** - Test error conditions
4. **Consider E2E tests** - For critical user workflows

## Code Coverage Goals

- **Minimum**: 70% for new code
- **Target**: 80-90% overall
- **Critical paths**: 100% coverage

## Best Practices

- ✅ Test behavior, not implementation
- ✅ Use descriptive test names
- ✅ Follow AAA pattern (Arrange, Act, Assert)
- ✅ Keep tests independent
- ✅ Mock external dependencies

## Testing Checklist

- [ ] All public functions have unit tests
- [ ] Edge cases are covered
- [ ] Error handling is tested
- [ ] Integration points are tested
- [ ] Tests run quickly
- [ ] Tests are maintainable
```

## Troubleshooting

### Skill Not Being Included

Check:
1. File is named exactly `SKILL.md` (case-sensitive)
2. File is in a subdirectory under `.agents/skills/`
3. Selectors match (use verbose logging to debug)

```bash
# Check what's being included
coding-context your-task 2>&1 | grep "Including skill"
```

### Wrong Skills Being Included

Review your selectors:

```bash
# Check current selectors
coding-context your-task 2>&1 | grep "Selectors"
```

Add more specific selectors to your skill frontmatter.

## See Also

- [File Formats Reference](../reference/file-formats#skill-files) - Skill format specification
- [Search Paths Reference](../reference/search-paths#skill-file-search-paths) - Where skills are found
- [How to Create Rules](./create-rules) - Similar concepts for rules
- [agentskills.io](https://agentskills.io) - Skill format specification
