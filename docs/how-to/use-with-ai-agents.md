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
coding-context -s resume=false /fix-bug > context-initial.txt
cat context-initial.txt | ai-agent > analysis.txt

# Step 2: Implementation (skip rules with -r)
coding-context -r /fix-bug > context-resume.txt
cat context-resume.txt analysis.txt | ai-agent > implementation.txt
```

## With GitHub Copilot

If you're using GitHub Copilot, the CLI can prepare context for custom instructions:

```bash
# Generate context
coding-context implement-feature > .github/copilot-context.md

# Copilot will read this file automatically
```

## Environment Variables for Bootstrap Scripts

Pass environment variables to bootstrap scripts:

```bash
# Set environment variables
export JIRA_API_KEY="your-api-key"
export GITHUB_TOKEN="your-token"
export DATABASE_URL="your-db-url"

# Bootstrap scripts can access these
coding-context -s source=jira /fix-bug | ai-agent
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
   coding-context -s languages=go -s priority=high /fix-bug
   ```

2. **Use resume mode to skip rules:**
   ```bash
   coding-context -r /fix-bug
   ```

3. **Split into multiple requests:**
   ```bash
   # First request: Planning
   coding-context -s stage=planning /plan-feature | ai-agent
   
   # Second request: Implementation
   coding-context -s stage=implementation /implement-feature | ai-agent
   ```

## See Also

- [CLI Reference](../reference/cli) - All command-line options
- [GitHub Actions Integration](./github-actions) - Automate with CI/CD
- [Creating Tasks](./create-tasks) - Define what AI should do
