---
task_name: example-with-standard-fields
agent: cursor
language: go
model: anthropic.claude-sonnet-4-20250514-v1-0
single_shot: false
timeout: 10m
selectors:
  stage: implementation
---

# Example Task with Standard Frontmatter Fields

This task demonstrates all standard frontmatter fields supported by the coding-context CLI.

## Standard Fields (Default Selectors)

These fields automatically filter rules when present in task frontmatter:

- **agent**: `cursor` - Only includes rules with `agent: cursor` (or no agent field)
- **language**: `go` - Only includes rules with `language: go` (or no language field)

## Standard Fields (Metadata Only)

These fields are stored in frontmatter and passed through to output, but do NOT filter rules:

- **model**: `anthropic.claude-sonnet-4-20250514-v1-0` - AI model to use for this task
- **single_shot**: `false` - Task can be run multiple times
- **timeout**: `10m` - Task timeout as time.Duration (10 minutes)

## Custom Selectors

Additional filtering criteria beyond the standard fields:

- **selectors.stage**: `implementation` - Only includes rules with `stage: implementation`

## How Filtering Works

When this task runs, rules are included if they match ALL of the following:
1. `agent: cursor` OR no agent field
2. `language: go` OR no language field  
3. `stage: implementation` OR no stage field
4. `task_name: example-with-standard-fields` OR no task_name field

Rules without any selectors are always included (generic rules).

## Usage

```bash
coding-context-cli example-with-standard-fields
```

The output will include:
1. Task frontmatter with all standard fields
2. Only rules matching the selectors
3. The task content
