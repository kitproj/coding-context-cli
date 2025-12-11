---
layout: default
title: Home
nav_order: 1
---

# Coding Context CLI

A command-line interface for dynamically assembling context for AI coding agents.

This tool collects context from predefined rule files and task-specific prompts, substitutes parameters, and outputs a combined context to standard output. Use it to provide rich, relevant information to AI models like Claude, Gemini, or GPT.

## Quick Example

```bash
# Install
sudo curl -fsL -o /usr/local/bin/coding-context \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.0.23/coding-context_v0.0.23_linux_amd64
sudo chmod +x /usr/local/bin/coding-context

# Use with an AI agent
coding-context -p issue_key=BUG-123 -s languages=go /fix-bug | llm -m claude-3-5-sonnet-20241022
```

## Documentation Structure

This documentation follows the [Diataxis](https://diataxis.fr/) framework, organized by your needs:

### 📚 [Tutorials](./tutorials/getting-started) - *Learning-oriented*

Step-by-step guides to get you started:
- [Getting Started Tutorial](./tutorials/getting-started) - Your first steps with the CLI

**Best for:** Learning the basics, first-time users, hands-on practice

### 🛠️ [How-to Guides](./how-to/) - *Problem-oriented*

Practical guides to solve specific problems:
- [Create Task Files](./how-to/create-tasks) - Define what AI agents should do
- [Create Rule Files](./how-to/create-rules) - Provide reusable context
- [Use Frontmatter Selectors](./how-to/use-selectors) - Filter rules and tasks
- [Use Remote Directories](./how-to/use-remote-directories) - Load rules from Git, HTTP, or S3
- [Use with AI Agents](./how-to/use-with-ai-agents) - Integrate with various AI tools
- [Integrate with GitHub Actions](./how-to/github-actions) - Automate with CI/CD

**Best for:** Solving specific problems, achieving specific goals, practical tasks

### 📖 [Reference](./reference/) - *Information-oriented*

Technical specifications and API details:
- [CLI Reference](./reference/cli) - Command-line options and arguments
- [File Formats](./reference/file-formats) - Task and rule file specifications
- [Search Paths](./reference/search-paths) - Where files are discovered

**Best for:** Looking up specific details, understanding options, technical specifications

### 💡 [Explanation](./explanation/) - *Understanding-oriented*

Conceptual guides to deepen your understanding:
- [Agentic Workflows](./explanation/agentic-workflows) - Understanding autonomous AI workflows
- [Architecture](./explanation/architecture) - How the CLI works internally

**Best for:** Understanding concepts, learning why things work the way they do, big picture

## Key Features

- **Dynamic Context Assembly**: Merges context from various source files
- **Remote Directories**: Load rules from Git, HTTP, S3, and other sources
- **Task-Specific Prompts**: Different prompts for different tasks
- **Rule-Based Context**: Reusable context snippets
- **Frontmatter Filtering**: Select rules based on metadata
- **Bootstrap Scripts**: Fetch or generate context dynamically
- **Parameter Substitution**: Inject runtime values
- **Token Estimation**: Monitor context size

## Supported AI Agent Formats

Automatically discovers rules from configuration files for:
- **Anthropic Claude**, **Cursor**, **GitHub Copilot**
- **Google Gemini**, **OpenCode.ai**, **Windsurf**
- **Augment**, **Codex**, and generic `.agents/` directories

See [Search Paths Reference](./reference/search-paths) for complete list.

## Quick Links

- [GitHub Repository](https://github.com/kitproj/coding-context-cli) - Source code and releases
- [Report Issues](https://github.com/kitproj/coding-context-cli/issues) - Bug reports and feature requests
- [License](https://github.com/kitproj/coding-context-cli/blob/main/LICENSE) - MIT License

## Need Help?

- **New to the CLI?** Start with the [Getting Started Tutorial](./tutorials/getting-started)
- **Have a specific goal?** Check the [How-to Guides](./how-to/)
- **Looking for details?** See the [Reference Documentation](./reference/)
- **Want to understand concepts?** Read the [Explanations](./explanation/)
