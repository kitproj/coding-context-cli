---
layout: default
title: Use with AI Agents
parent: How-to Guides
nav_order: 4
---

# How to Use with AI Agents

This guide shows you how to use the Coding Context CLI with various AI agents and tools.

## Basic Usage

Pipe the assembled context to any AI agent:

```bash
coding-context fix-bug | your-ai-agent
```

## With Claude CLI

```bash
coding-context \
  -p issue_key=BUG-123 \
  -s languages=go \
  /fix-bug | claude
```

## With OpenAI API

```bash
coding-context code-review | openai api completions.create \
  -m gpt-4 \
  --stream
```

## With LLM Tool

The [llm](https://llm.datasette.io/) tool supports many models:

```bash
# Using Claude
coding-context fix-bug | llm -m claude-3-5-sonnet-20241022

# Using Gemini
coding-context code-review | llm -m gemini-pro

# Using local models
coding-context implement-feature | llm -m llama2
```

## Saving Context to File

Save the context for later use or inspection:

```bash
# Save to file
coding-context fix-bug > context.txt

# Review the context
cat context.txt

# Use with AI agent
cat context.txt | claude
```

## Multi-Step Workflows

Use context in iterative workflows:

```bash
# Step 1: Initial analysis
coding-context -s resume=false fix-bug > context-initial.txt
cat context-initial.txt | ai-agent > analysis.txt

# Step 2: Implementation (skip rules with --skip-bootstrap)
coding-context -r --skip-bootstrap fix-bug > context-resume.txt
cat context-resume.txt analysis.txt | ai-agent > implementation.txt
```

## With GitHub Copilot

If you're using GitHub Copilot, the CLI can prepare context for custom instructions:

```bash
# Generate context
coding-context implement-feature > .github/copilot-context.md

# Copilot will read this file automatically
```

## Write-Rules Mode

Write-rules mode (`-w` flag) separates rules from tasks, allowing AI agents to read rules from their standard configuration files while keeping task prompts clean.

### Benefits

- **Token Savings**: Avoid including all rules in every prompt
- **Agent Integration**: Write rules to agent-specific config files
- **Clean Prompts**: Output only the task to stdout

### Basic Usage

```bash
# Write rules to agent's config file, output task to stdout
coding-context -a copilot -w fix-bug | llm -m claude-3-5-sonnet
```

This will:
1. Write all rules to `~/.github/agents/AGENTS.md`
2. Output only the task prompt to stdout
3. The AI agent reads rules from its config file

### Agent-Specific Paths

Each agent has a designated configuration file:

```bash
# GitHub Copilot
coding-context -a copilot -w fix-bug  # → ~/.github/agents/AGENTS.md

# Claude
coding-context -a claude -w fix-bug   # → ~/.claude/CLAUDE.md

# Cursor
coding-context -a cursor -w fix-bug   # → ~/.cursor/rules/AGENTS.md

# Gemini
coding-context -a gemini -w fix-bug   # → ~/.gemini/GEMINI.md
```

### Task-Specified Agent

Tasks can specify their preferred agent in frontmatter:

**Task file (`deploy.md`):**
```yaml
---
agent: claude
---
# Deploy to Production
...
```

**Usage:**
```bash
# Task's agent field is used (writes to ~/.claude/CLAUDE.md)
coding-context -w deploy

# Task agent overrides -a flag
coding-context -a copilot -w deploy  # Still uses claude
```

### Workflow Example

```bash
# 1. Initial setup: Write rules once
coding-context -a copilot -w setup-project

# 2. Run multiple tasks without re-including rules
coding-context -a copilot -w fix-bug | llm
coding-context -a copilot -w code-review | llm
coding-context -a copilot -w refactor | llm

# 3. Update rules when needed
coding-context -a copilot -w -s languages=go update-rules
```

## Environment Variables for Bootstrap Scripts

Pass environment variables to bootstrap scripts:

```bash
# Set environment variables
export JIRA_API_KEY="your-api-key"
export GITHUB_TOKEN="your-token"
export DATABASE_URL="your-db-url"

# Bootstrap scripts can access these
coding-context -s source=jira fix-bug | ai-agent
```

## Token Count Monitoring

The CLI prints token estimates to stderr:

```bash
# See token count while piping to AI
coding-context fix-bug 2>&1 | tee >(grep -i token >&2) | ai-agent

# Or redirect stderr to file
coding-context fix-bug 2> tokens.log | ai-agent
```

## Batch Processing

Process multiple tasks:

```bash
# Process multiple bug fixes
for issue in BUG-101 BUG-102 BUG-103; do
  coding-context \
    -p issue_key=$issue \
    /fix-bug | ai-agent > "fix-$issue.txt"
done
```

## Custom AI Agent Scripts

Create a wrapper script for your preferred setup:

```bash
#!/bin/bash
# ai-fix-bug.sh

ISSUE_KEY=$1
DESCRIPTION=$2

coding-context \
  -s languages=go \
  -s priority=high \
  -p issue_key="$ISSUE_KEY" \
  -p description="$DESCRIPTION" \
  /fix-bug | llm -m claude-3-5-sonnet-20241022
```

Use with:
```bash
chmod +x ai-fix-bug.sh
./ai-fix-bug.sh BUG-123 "Application crashes on startup"
```

## Handling Large Contexts

If your context exceeds token limits:

1. **Use selectors to reduce included rules:**
   ```bash
   coding-context -s priority=high fix-bug
   ```

2. **Use bootstrap disabled to skip rules:**
   ```bash
   coding-context -r fix-bug
   ```

3. **Split into multiple requests:**
   ```bash
   # First request: Planning
   coding-context -s stage=planning plan-feature | ai-agent
   
   # Second request: Implementation
   coding-context -s stage=implementation implement-feature | ai-agent
   ```

## See Also

- [CLI Reference](../reference/cli) - All command-line options
- [GitHub Actions Integration](./github-actions) - Automate with CI/CD
- [Creating Tasks](./create-tasks) - Define what AI should do
