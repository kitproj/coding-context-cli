---
layout: default
title: Examples
---

# Examples

This page provides real-world examples demonstrating how to use the Coding Context CLI in various scenarios.

## Basic Examples

### Simple Bug Fix

**Task file (`.agents/tasks/fix-bug.md`):**
```markdown
---
task_name: fix-bug
---
# Bug Fix Task

Analyze and fix the reported bug.
Follow coding standards and add regression tests.
```

**Usage:**
```bash
coding-context-cli fix-bug | ai-agent
```

### Bug Fix with Parameters

**Task file (`.agents/tasks/fix-bug-detailed.md`):**
```markdown
---
task_name: fix-bug-detailed
---
# Bug Fix: ${issue_key}

## Issue Details
- Key: ${issue_key}
- Title: ${issue_title}
- Severity: ${severity}

## Requirements
- Fix the bug
- Add regression tests
- Update documentation if needed
```

**Usage:**
```bash
coding-context-cli \
  -p issue_key=PROJ-1234 \
  -p issue_title="Application crashes on startup" \
  -p severity=critical \
  fix-bug-detailed | ai-agent
```

## Language-Specific Examples

### Go Project

**Rule file (`.agents/rules/go-standards.md`):**
```markdown
---
language: Go
---

# Go Coding Standards

- Use `gofmt` for formatting
- Handle all errors explicitly
- Write table-driven tests
- Use meaningful package names
- Add godoc comments for exported items
```

**Usage:**
```bash
coding-context-cli -s language=Go fix-bug | ai-agent
```

### Python Project

**Rule file (`.agents/rules/python-standards.md`):**
```markdown
---
language: Python
---

# Python Coding Standards

- Follow PEP 8
- Use type hints
- Write docstrings for all public functions
- Use pytest for testing
- Format with black
```

**Usage:**
```bash
coding-context-cli -s language=Python implement-feature | ai-agent
```

## Multi-Stage Workflow Examples

### Feature Development

Create different task files for each stage:

**Planning stage (`.agents/tasks/plan-feature.md`):**
```markdown
---
task_name: feature
stage: planning
---
# Feature Planning: ${feature_name}

Create a detailed implementation plan for: ${feature_name}

## Planning Requirements
- Break down into subtasks
- Identify dependencies
- Estimate complexity
- Consider edge cases
```

**Implementation stage (`.agents/tasks/implement-feature.md`):**
```markdown
---
task_name: feature
stage: implementation
---
# Implement Feature: ${feature_name}

Implement the feature based on the plan.

## Implementation Requirements
- Follow coding standards
- Write unit tests
- Add integration tests
- Update documentation
```

**Testing stage (`.agents/tasks/test-feature.md`):**
```markdown
---
task_name: feature
stage: testing
---
# Test Feature: ${feature_name}

Verify the feature implementation.

## Testing Requirements
- Run all tests
- Check edge cases
- Verify performance
- Review security implications
```

**Usage:**
```bash
# Planning
coding-context-cli \
  -s stage=planning \
  -p feature_name="User Authentication" \
  feature | ai-agent

# Implementation
coding-context-cli \
  -s stage=implementation \
  -p feature_name="User Authentication" \
  feature | ai-agent

# Testing
coding-context-cli \
  -s stage=testing \
  -p feature_name="User Authentication" \
  feature | ai-agent
```

## Code Review Example

**Task file (`.agents/tasks/code-review.md`):**
```markdown
---
task_name: code-review
---
# Code Review: PR #${pr_number}

## Review Focus
- Code quality and readability
- Test coverage
- Security implications
- Performance considerations
- Documentation completeness

## Pull Request
- Number: ${pr_number}
- Title: ${pr_title}
- Author: ${pr_author}

Please provide a thorough review with specific, actionable feedback.
```

**Rule file (`.agents/rules/review-checklist.md`):**
```markdown
---
stage: review
---

# Code Review Checklist

## Code Quality
- [ ] Code follows project conventions
- [ ] Functions are single-purpose
- [ ] Variable names are descriptive
- [ ] No code duplication

## Testing
- [ ] Tests cover new functionality
- [ ] Edge cases are tested
- [ ] Tests are maintainable

## Security
- [ ] Input validation is present
- [ ] No hardcoded secrets
- [ ] Dependencies are up to date

## Documentation
- [ ] Public APIs are documented
- [ ] README is updated if needed
- [ ] Breaking changes are noted
```

**Usage:**
```bash
coding-context-cli \
  -s stage=review \
  -p pr_number=123 \
  -p pr_title="Add user authentication" \
  -p pr_author="johndoe" \
  code-review | ai-agent
```

## Bootstrap Script Example

**Rule file (`.agents/rules/jira-context.md`):**
```markdown
---
source: jira
---

# JIRA Issue Context

This rule provides context from JIRA for the current issue.
The bootstrap script fetches the issue details.
```

**Bootstrap script (`.agents/rules/jira-context-bootstrap`):**
```bash
#!/bin/bash

# Fetch JIRA issue details
ISSUE_KEY="${JIRA_ISSUE_KEY}"

if [ -z "$ISSUE_KEY" ]; then
    echo "No JIRA issue key provided" >&2
    exit 0
fi

echo "Fetching JIRA issue: $ISSUE_KEY" >&2

curl -s -H "Authorization: Bearer ${JIRA_API_TOKEN}" \
    "https://your-domain.atlassian.net/rest/api/3/issue/${ISSUE_KEY}" \
    | jq -r '.fields | {summary, description, status: .status.name}' \
    > /tmp/jira-issue.json

echo "JIRA issue fetched successfully" >&2
```

**Usage:**
```bash
export JIRA_ISSUE_KEY="PROJ-1234"
export JIRA_API_TOKEN="your-token"

coding-context-cli -s source=jira fix-bug | ai-agent
```

## Resume Mode Example

**Initial task (`.agents/tasks/refactor-initial.md`):**
```markdown
---
task_name: refactor
resume: false
---
# Refactoring Task

Analyze the code and create a refactoring plan.

## Guidelines
- Identify code smells
- Propose improvements
- Maintain backward compatibility
- Write refactoring tests
```

**Resume task (`.agents/tasks/refactor-continue.md`):**
```markdown
---
task_name: refactor
resume: true
---
# Continue Refactoring

Continue with the refactoring work.

## Current Status
Review your previous work and:
- Complete remaining refactoring tasks
- Ensure all tests pass
- Update documentation
- Verify backward compatibility
```

**Usage:**
```bash
# Initial session (includes all rules)
coding-context-cli -s resume=false refactor > initial-context.txt
cat initial-context.txt | ai-agent > refactoring-plan.txt

# Continue in new session (skips rules, saves tokens)
coding-context-cli -r refactor | ai-agent
```

## GitHub Actions Integration

### Automated PR Review

```yaml
name: AI Code Review
on: pull_request

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install CLI
        run: |
          curl -fsL -o /usr/local/bin/coding-context-cli \
            https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context-cli_linux_amd64
          chmod +x /usr/local/bin/coding-context-cli
      
      - name: Assemble Context
        run: |
          coding-context-cli \
            -s stage=review \
            -p pr_number=${{ github.event.pull_request.number }} \
            -p pr_title="${{ github.event.pull_request.title }}" \
            code-review > context.txt
      
      - name: Review with AI
        run: cat context.txt | your-ai-agent > review.md
```

### Automated Bug Fix

```yaml
name: AI Bug Fix
on:
  issues:
    types: [labeled]

jobs:
  fix:
    if: contains(github.event.issue.labels.*.name, 'bug')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install CLI
        run: |
          curl -fsL -o /usr/local/bin/coding-context-cli \
            https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context-cli_linux_amd64
          chmod +x /usr/local/bin/coding-context-cli
      
      - name: Assemble Context
        run: |
          coding-context-cli \
            -p issue_number=${{ github.event.issue.number }} \
            -p issue_title="${{ github.event.issue.title }}" \
            fix-bug > context.txt
      
      - name: Fix with AI
        run: cat context.txt | your-ai-agent --apply
```

## Advanced Examples

### Multi-Language Project

**Directory structure:**
```
.agents/
├── rules/
│   ├── go-standards.md        (language: Go)
│   ├── python-standards.md    (language: Python)
│   ├── javascript-standards.md (language: JavaScript)
│   └── general-standards.md   (no language selector)
└── tasks/
    └── implement-feature.md
```

**Usage:**
```bash
# Backend (Go)
coding-context-cli -s language=Go implement-feature

# ML Pipeline (Python)
coding-context-cli -s language=Python implement-feature

# Frontend (JavaScript)
coding-context-cli -s language=JavaScript implement-feature
```

### Environment-Specific Deployment

**Task files:**
```
.agents/tasks/
├── deploy-dev.md         (environment: development)
├── deploy-staging.md     (environment: staging)
└── deploy-production.md  (environment: production)
```

**Usage:**
```bash
# Deploy to different environments
coding-context-cli -s environment=development deploy
coding-context-cli -s environment=staging deploy
coding-context-cli -s environment=production deploy
```

## Template Repository

For a complete example setup, see the [examples directory](https://github.com/kitproj/coding-context-cli/tree/main/examples) in the repository:

```
examples/
├── workflows/          # GitHub Actions examples
│   ├── agentic-code-review.yml
│   ├── agentic-bugfix.yml
│   └── agentic-feature-development.yml
└── agents/            # Agent configuration examples
    ├── rules/
    │   ├── planning.md
    │   ├── implementation-go.md
    │   └── testing.md
    └── tasks/
        ├── plan-feature.md
        ├── code-review.md
        └── fix-bug.md
```

## Tips and Best Practices

1. **Start Simple**: Begin with basic tasks and rules, then add complexity as needed
2. **Use Selectors**: Target specific contexts to reduce token usage
3. **Parameterize**: Make tasks reusable with parameters
4. **Version Control**: Keep `.agents/` in git for team collaboration
5. **Test Locally**: Verify context assembly before using in production
6. **Monitor Tokens**: Keep context size manageable for AI models
7. **Document Well**: Include clear instructions in task prompts
8. **Iterate**: Refine based on AI agent performance

## Community Examples

Want to share your examples? Submit them to the [examples directory](https://github.com/kitproj/coding-context-cli/tree/main/examples) via pull request!

## Next Steps

- Review the [Usage Guide](./usage) for detailed CLI documentation
- Learn about [Agentic Workflows](./agentic-workflows) integration
- Check out the [GitHub repository](https://github.com/kitproj/coding-context-cli) for more examples
