---
layout: default
title: Specification
parent: Reference
nav_order: 1
---

# Coding Context Standard Specification

The **Coding Context Standard** is a convention-based file format and directory structure for providing rich contextual information to AI coding agents.

## Full Specification

The complete specification is available in the repository:

**[View SPECIFICATION.md](https://github.com/kitproj/coding-context-cli/blob/main/SPECIFICATION.md)**

## Quick Overview

The standard defines:

### File Formats
- **Rules**: Reusable context snippets (`.md`, `.mdc` files)
- **Tasks**: Specific prompts matched by filename
- **Commands**: Reusable content blocks referenced via slash syntax
- **Skills**: Specialized capabilities with progressive disclosure

### Directory Structure
- `.agents/` - Generic AI agent files
- `.cursor/`, `.github/`, `.opencode/` - Agent-specific directories
- Standard file locations in project and home directories

### Frontmatter Metadata
- YAML frontmatter for structured metadata
- Standard fields: `languages`, `stage`, `agent`, `selectors`, etc.
- Custom fields for flexible filtering

### Content Expansion
- **Parameter expansion**: `${param_name}`
- **Command expansion**: `` !`shell command` ``
- **Path expansion**: `@path/to/file`
- **Slash commands**: `/command-name`

### Selector System
- Filter rules based on frontmatter metadata
- Command-line: `-s key=value`
- Task frontmatter: `selectors` field
- Supports AND logic and array-based OR logic

### Bootstrap Scripts
- Executable scripts named `{base-name}-bootstrap`
- Run before rule/task processing
- Enable dynamic environment setup

## Key Design Principles

1. **Convention over Configuration**: Predetermined search paths and naming conventions
2. **Simplicity**: Markdown files with YAML frontmatter
3. **Composability**: Mix rules from multiple sources
4. **Flexibility**: Support diverse workflows and agents
5. **Security**: Single-pass expansion prevents injection attacks

## Compatibility

The standard is compatible with multiple AI coding agents:

- Anthropic Claude
- GitHub Copilot
- Cursor
- OpenCode.ai
- Augment
- Windsurf
- Google Gemini
- Codex

## Version

**Current Version:** 1.0 (2025-12-25)

## See Also

- [File Formats Reference](./file-formats) - Detailed file format documentation
- [Search Paths Reference](./search-paths) - Discovery mechanisms
- [CLI Reference](./cli) - Command-line interface
