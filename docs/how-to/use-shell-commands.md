---
layout: default
title: Use Shell Commands
parent: How-To Guides
nav_order: 7
---

# Use Shell Commands in Tasks and Rules

This guide shows you how to inject the output of shell commands into your tasks and rules using the `!`command`` syntax.

## Overview

You can execute shell commands and inject their output directly into your prompts. This is useful for:

- Analyzing test results
- Reviewing recent commits
- Inspecting file contents
- Getting system information
- Any dynamic information that can be obtained via shell commands

## Syntax

To execute a shell command and include its output, use the following syntax on its own line:

```markdown
!`command`
```

The command must:
- Be on its own line
- Be wrapped in backticks
- Start with `!`

## Examples

### Basic Command

```markdown
---
task_name: show-date
---

# Current Date

Today is:
!`date +%Y-%m-%d`
```

### Analyzing Test Coverage

```markdown
---
task_name: analyze-coverage
description: Analyze test coverage
---

# Test Coverage Analysis

Here are the current test results:
!`go test -cover ./... 2>&1 | grep -E "(^ok|coverage:)" | head -10`

Based on these results, suggest improvements to increase coverage.
```

### Reviewing Recent Changes

```markdown
---
task_name: review-commits
description: Review recent changes
---

# Code Review of Recent Commits

Recent git commits:
!`git log --oneline -10`

Review these changes and suggest any improvements.
```

### Multiple Commands

You can use multiple shell commands in the same file:

```markdown
---
task_name: project-status
---

# Project Status

## Git Status
!`git status --short`

## Recent Commits
!`git log --oneline -5`

## Test Results
!`go test ./... -cover 2>&1 | tail -5`
```

## Command Execution

Commands are executed:
- In your project's root directory (or the directory specified with `-C`)
- Using `sh -c`, so you can use shell features like pipes, redirects, and command substitution
- With their `stdout` captured and injected into the prompt
- Before parameter substitution in task files

## Error Handling

If a command fails:
- The error is logged to stderr
- A comment is inserted in place of the output: `<!-- Error executing command 'cmd': error message -->`
- Processing continues for other commands

## Tips

1. **Use pipes and filters** to limit output:
   ```markdown
   !`git log --oneline -10`  # Only last 10 commits
   !`go test ./... 2>&1 | grep FAIL`  # Only failures
   ```

2. **Combine with parameters** in task files:
   ```markdown
   ---
   task_name: review-pr
   ---
   
   PR #${pr_number} changes:
   !`git diff main...feature-branch`
   ```

3. **Use in both rules and tasks**:
   - Rules: For context that applies to multiple tasks
   - Tasks: For task-specific dynamic information

4. **Keep commands simple**: Complex logic should go in bootstrap scripts

## Best Practices

- Keep command output concise to avoid token limits
- Use filtering (grep, head, tail) to limit output
- Test commands manually before adding them
- Consider command execution time
- Use meaningful context around command output
- Document what each command does

## See Also

- [Create Tasks](create-tasks.md) - Learn about task files
- [Create Rules](create-rules.md) - Learn about rule files
- [Bootstrap Scripts](../reference/bootstrap-scripts.md) - For more complex setup
