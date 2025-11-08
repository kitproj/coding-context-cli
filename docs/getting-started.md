---
layout: default
title: Getting Started
---

# Getting Started

This guide will help you install and start using the Coding Context CLI.

## Installation

You can install the CLI by downloading the latest release from the [releases page](https://github.com/kitproj/coding-context-cli/releases) or by building from source.

### Linux

```bash
# Download the latest release for Linux AMD64
sudo curl -fsL -o /usr/local/bin/coding-context-cli \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.1.0/coding-context-cli_linux_amd64
sudo chmod +x /usr/local/bin/coding-context-cli

# Verify installation
coding-context-cli --help
```

### macOS

```bash
# Download the latest release for macOS
sudo curl -fsL -o /usr/local/bin/coding-context-cli \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.1.0/coding-context-cli_darwin_amd64
sudo chmod +x /usr/local/bin/coding-context-cli

# Verify installation
coding-context-cli --help
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/kitproj/coding-context-cli.git
cd coding-context-cli

# Build the binary
go build -o coding-context-cli

# Install to /usr/local/bin
sudo mv coding-context-cli /usr/local/bin/

# Verify installation
coding-context-cli --help
```

## First Steps

### 1. Create a Task Directory

Create a directory structure for your agent tasks and rules:

```bash
mkdir -p .agents/tasks
mkdir -p .agents/rules
```

### 2. Create Your First Task

Create a simple bug fix task:

```bash
cat > .agents/tasks/fix-bug.md << 'EOF'
---
task_name: fix-bug
---
# Task: Fix Bug in ${jira_issue_key}

Please analyze the following issue and provide a fix.

## Issue Details
- Issue: ${jira_issue_key}
- Description: ${issue_description}

## Requirements
- Fix the bug
- Add tests to prevent regression
- Update documentation if needed
EOF
```

### 3. Create a Rule File

Create a coding standards rule:

```bash
cat > .agents/rules/coding-standards.md << 'EOF'
---
language: Go
---

# Go Coding Standards

- All new code must be accompanied by unit tests
- Follow the standard Go formatting (use `gofmt`)
- Use meaningful variable names
- Add comments for exported functions
- Handle errors properly, don't ignore them
EOF
```

### 4. Use the CLI

Now you can assemble context and pipe it to an AI agent:

```bash
# Assemble context with parameters
coding-context-cli \
  -p jira_issue_key=PROJ-1234 \
  -p issue_description="Application crashes on startup" \
  -s language=Go \
  fix-bug | llm -m claude-3-5-sonnet-20241022
```

This will:
1. Find the task file with `task_name: fix-bug`
2. Include all rules matching `-s language=Go`
3. Substitute the parameters in the task prompt
4. Output the assembled context to stdout (piped to `llm`)

## Next Steps

- Read the [Usage Guide](./usage) for detailed CLI options and advanced features
- Explore [Examples](./examples) for real-world use cases
- Learn about [Agentic Workflows](./agentic-workflows) integration with GitHub Actions

## Common Patterns

### Code Review

```bash
coding-context-cli -p pr_number=42 code-review | ai-agent
```

### Feature Implementation

```bash
coding-context-cli \
  -s stage=planning \
  -p feature_name="User Authentication" \
  implement-feature | ai-agent
```

### Bug Triage

```bash
coding-context-cli \
  -p severity=critical \
  -s language=Python \
  triage-bug | ai-agent
```

## Troubleshooting

### No task found error

If you get an error like "no task found with task_name: xxx", make sure:
- Your task file has the correct `task_name` in the frontmatter
- The task file is in one of the search paths (`.agents/tasks/`, `~/.agents/tasks/`, `/etc/agents/tasks/`)

### Rules not included

If rules aren't being included:
- Check that the frontmatter selectors match (e.g., `-s language=Go` requires `language: Go` in the rule frontmatter)
- Remember that frontmatter selectors only match top-level YAML fields
- If no selectors are specified, all rules are included by default

### Bootstrap scripts not running

Bootstrap scripts must:
- Be executable (`chmod +x script-name`)
- Have the same name as the rule file with `-bootstrap` suffix
- Be in the same directory as the rule file

## Getting Help

- Check the [Usage Guide](./usage) for detailed documentation
- Visit the [GitHub repository](https://github.com/kitproj/coding-context-cli) to report issues
- Review the [Examples](./examples) for common patterns and templates
