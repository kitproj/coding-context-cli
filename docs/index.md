---
layout: default
title: Home
---

# Coding Context CLI

A command-line interface for dynamically assembling context for AI coding agents.

This tool collects context from predefined rule files and a task-specific prompt, substitutes parameters, and prints a single, combined context to standard output. This is useful for feeding a large amount of relevant information into an AI model like Claude, Gemini, or OpenAI's GPT series.

## Quick Start

```bash
# Install the CLI
sudo curl -fsL -o /usr/local/bin/coding-context-cli \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.1.0/coding-context-cli_linux_amd64
sudo chmod +x /usr/local/bin/coding-context-cli

# Use it with an AI agent
coding-context-cli -p jira_issue_key=PROJ-1234 fix-bug | llm -m gemini-pro
```

## Features

- **Dynamic Context Assembly**: Merges context from various source files
- **Task-Specific Prompts**: Use different prompts for different tasks (e.g., `feature`, `bugfix`)
- **Rule-Based Context**: Define reusable context snippets (rules) that can be included or excluded
- **Frontmatter Filtering**: Select rules based on metadata using frontmatter selectors
- **Bootstrap Scripts**: Run scripts to fetch or generate context dynamically
- **Parameter Substitution**: Inject values into your task prompts
- **Token Estimation**: Get an estimate of the total token count for the generated context

## Supported Coding Agents

This tool is compatible with configuration files from various AI coding agents and IDEs:

- **[Anthropic Claude](https://claude.ai/)**: `CLAUDE.md`, `CLAUDE.local.md`, `.claude/CLAUDE.md`
- **[Codex](https://codex.ai/)**: `AGENTS.md`, `.codex/AGENTS.md`
- **[Cursor](https://cursor.sh/)**: `.cursor/rules`, `.cursorrules`
- **[Augment](https://augmentcode.com/)**: `.augment/rules`, `.augment/guidelines.md`
- **[Windsurf](https://codeium.com/windsurf)**: `.windsurf/rules`, `.windsurfrules`
- **[OpenCode.ai](https://opencode.ai/)**: `.opencode/agent`, `.opencode/command`, `.opencode/rules`
- **[GitHub Copilot](https://github.com/features/copilot)**: `.github/copilot-instructions.md`, `.github/agents`
- **[Google Gemini](https://gemini.google.com/)**: `GEMINI.md`, `.gemini/styleguide.md`
- **Generic AI Agents**: `AGENTS.md`, `.agents/rules`

## Agentic Workflows

This tool plays a crucial role in the **agentic workflow ecosystem** by providing rich, contextual information to AI agents. It complements systems like **GitHub Next's Agentic Workflows** by:

- **Context Preparation**: Assembles rules, guidelines, and task-specific prompts before agent execution
- **Workflow Integration**: Can be invoked in GitHub Actions to provide context to autonomous agents
- **Dynamic Context**: Supports runtime parameters and bootstrap scripts for real-time information
- **Multi-Stage Support**: Different context assemblies for planning, implementation, and validation stages

For a comprehensive guide on using this tool with agentic workflows, see the [Agentic Workflows](./agentic-workflows) guide.

## Documentation

- [Getting Started](./getting-started) - Installation and first steps
- [Usage Guide](./usage) - Detailed usage instructions and CLI reference
- [Agentic Workflows](./agentic-workflows) - Integration with GitHub Actions and AI workflows
- [Examples](./examples) - Real-world examples and templates

## GitHub Repository

Visit the [GitHub repository](https://github.com/kitproj/coding-context-cli) to:
- Report issues
- Contribute to the project
- View the source code
- Download the latest releases

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/kitproj/coding-context-cli/blob/main/LICENSE) file for details.
