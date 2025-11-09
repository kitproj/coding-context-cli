---
layout: default
title: Getting Started Tutorial
parent: Tutorials
nav_order: 1
---

# Getting Started Tutorial

This tutorial will guide you through your first steps with the Coding Context CLI. By the end, you'll have installed the tool and assembled your first context for an AI agent.

## What You'll Learn

- How to install the CLI
- How to create a basic task file
- How to create a rule file
- How to assemble context and use it with an AI agent

## Prerequisites

- A Unix-like system (Linux or macOS)
- Basic familiarity with the command line
- An AI agent tool (like `llm`, `claude`, or similar)

## Step 1: Install the CLI

First, download and install the Coding Context CLI:

```bash
# For Linux
sudo curl -fsL -o /usr/local/bin/coding-context-cli \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.1.0/coding-context-cli_linux_amd64
sudo chmod +x /usr/local/bin/coding-context-cli

# For macOS
sudo curl -fsL -o /usr/local/bin/coding-context-cli \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.1.0/coding-context-cli_darwin_amd64
sudo chmod +x /usr/local/bin/coding-context-cli
```

Verify the installation:

```bash
coding-context-cli --help
```

You should see the usage instructions.

## Step 2: Set Up Your Project

Create a directory structure for your tasks and rules:

```bash
# Create the directory structure
mkdir -p .agents/tasks
mkdir -p .agents/rules

# Navigate to the agents directory
cd .agents
```

## Step 3: Create Your First Task File

Create a simple bug fix task. Task files define what you want the AI agent to do.

```bash
cat > tasks/fix-bug.md << 'EOF'
---
task_name: fix-bug
---
# Task: Fix Bug

Please analyze the following bug and provide a fix.

## Bug Details
- Issue: ${issue_key}
- Description: ${description}

## Requirements
- Fix the bug
- Add tests to prevent regression
- Update documentation if needed
EOF
```

**What this does:**
- The frontmatter (`---` section) defines `task_name: fix-bug` which is how you'll reference this task
- The `${issue_key}` and `${description}` are placeholders that will be replaced with actual values
- The content provides instructions for the AI agent

## Step 4: Create Your First Rule File

Rules are reusable context snippets that provide guidelines to AI agents.

```bash
cat > rules/coding-standards.md << 'EOF'
---
language: Go
---

# Go Coding Standards

When writing Go code:
- Use `gofmt` for formatting
- Handle all errors explicitly (never ignore errors)
- Write table-driven tests
- Use meaningful variable names
- Add comments for exported functions
EOF
```

**What this does:**
- The frontmatter includes `language: Go`, which allows filtering this rule for Go projects
- The content provides coding standards that the AI agent should follow

## Step 5: Assemble Context

Now let's assemble context for a bug fix:

```bash
# Go back to your project root
cd ..

# Assemble context
coding-context-cli \
  -p issue_key=BUG-123 \
  -p description="Application crashes on startup" \
  -s language=Go \
  fix-bug
```

**What this command does:**
- `-p issue_key=BUG-123` replaces `${issue_key}` in the task
- `-p description="..."` replaces `${description}` in the task
- `-s language=Go` includes only rules with `language: Go` in frontmatter
- `fix-bug` is the task name to use

You should see output containing:
1. The Go coding standards rule
2. The bug fix task with your parameters substituted

## Step 6: Use with an AI Agent

Pipe the output to an AI agent:

```bash
coding-context-cli \
  -p issue_key=BUG-123 \
  -p description="Application crashes on startup" \
  -s language=Go \
  fix-bug | llm -m claude-3-5-sonnet-20241022
```

The AI agent will receive the assembled context and provide a response based on your coding standards and task requirements.

## What You've Learned

You've successfully:
- ✅ Installed the Coding Context CLI
- ✅ Created a task file with parameters
- ✅ Created a rule file with frontmatter selectors
- ✅ Assembled context with parameters and selectors
- ✅ Used the assembled context with an AI agent

## Next Steps

Now that you understand the basics, explore:
- [How-to Guides](../how-to/) - Solve specific problems
- [Reference Documentation](../reference/) - Detailed technical information
- [Explanations](../explanation/) - Understand concepts in depth
