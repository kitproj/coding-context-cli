# Coding Context Standard Specification

**Version:** 1.0  
**Status:** Draft  
**Last Updated:** 2025-12-25

## Abstract

This document specifies the **Coding Context Standard**, a convention-based file format and directory structure for providing rich contextual information to AI coding agents. The standard defines how coding rules, task prompts, reusable commands, and specialized skills are organized, discovered, filtered, and assembled into cohesive context for AI-assisted software development.

## Table of Contents

1. [Introduction](#1-introduction)
2. [Core Concepts](#2-core-concepts)
3. [File Formats](#3-file-formats)
4. [Directory Structure](#4-directory-structure)
5. [Frontmatter Specification](#5-frontmatter-specification)
6. [Content Expansion](#6-content-expansion)
7. [Selector System](#7-selector-system)
8. [Bootstrap Scripts](#8-bootstrap-scripts)
9. [Agent Targeting](#9-agent-targeting)
10. [Discovery and Search Paths](#10-discovery-and-search-paths)
11. [Context Assembly](#11-context-assembly)
12. [Compatibility](#12-compatibility)
13. [Examples](#13-examples)
14. [Rationale and Design Principles](#14-rationale-and-design-principles)

---

## 1. Introduction

### 1.1 Purpose

AI coding agents require comprehensive context to make informed decisions about code generation, review, and modification. This context includes:

- Project-specific coding standards and conventions
- Technology stack guidelines and best practices
- Task-specific instructions and constraints
- Reusable command templates
- Specialized domain knowledge (skills)

The Coding Context Standard provides a structured, convention-based approach to organizing and delivering this information.

### 1.2 Scope

This specification defines:

- **File formats** for rules, tasks, commands, and skills
- **Directory structures** for organizing context files
- **Metadata schemas** using YAML frontmatter
- **Discovery mechanisms** for locating context files
- **Filtering systems** for selecting relevant context
- **Assembly rules** for combining context into a cohesive output
- **Expansion mechanisms** for dynamic content generation

### 1.3 Goals

- **Simplicity**: Easy to create and understand
- **Composability**: Mix and match context from multiple sources
- **Flexibility**: Support diverse workflows and agents
- **Convention over Configuration**: Minimize boilerplate
- **Interoperability**: Compatible with multiple AI agents

---

## 2. Core Concepts

### 2.1 Rules

**Rules** are reusable context snippets that define coding standards, conventions, and guidelines. Rules are:

- Stored as Markdown files (`.md` or `.mdc`)
- Optionally filtered using frontmatter metadata
- Included in context assembly before task content
- Environment-agnostic (applicable across tasks)

**Example use cases:**
- Language-specific coding standards (Go, Python, JavaScript)
- Testing guidelines and patterns
- Security requirements
- Documentation conventions

### 2.2 Tasks

**Tasks** are specific prompts that define what an AI agent should accomplish. Tasks are:

- Stored as Markdown files (`.md`)
- Matched by filename (not frontmatter)
- Can include parameter placeholders for substitution
- Appear after rules in assembled context
- May specify selectors to auto-filter rules

**Example use cases:**
- Fix a bug
- Implement a feature
- Review code
- Write tests
- Refactor code

### 2.3 Commands

**Commands** are reusable content blocks referenced from tasks using slash command syntax (`/command-name`). Commands enable:

- Modular, composable task definitions
- Shared templates across multiple tasks
- Parameter passing to command content

**Example use cases:**
- Pre-deployment checklists
- Common instruction sets
- Boilerplate text

### 2.4 Skills

**Skills** provide specialized capabilities with progressive disclosure. Skills are:

- Discovered but only metadata is initially included
- Full content loaded on-demand by AI agents
- Organized in subdirectories with `SKILL.md` files
- Described with structured metadata

**Example use cases:**
- Data analysis capabilities
- PDF processing utilities
- API testing frameworks
- Specialized domain knowledge

### 2.5 Bootstrap Scripts

**Bootstrap scripts** are executable files that prepare the environment before context assembly. They:

- Run before their corresponding rule/task is processed
- Output to stderr (not included in context)
- Enable dynamic environment setup
- Follow naming convention: `{base-name}-bootstrap`

**Example use cases:**
- Installing required tools
- Fetching external data
- Environment validation
- Credential setup

---

## 3. File Formats

### 3.1 General Structure

All context files follow this structure:

```markdown
---
<optional-yaml-frontmatter>
---

# File Content

Content in Markdown format with optional expansions.
```

### 3.2 Rule Files

**Location**: Various (see [Directory Structure](#4-directory-structure))  
**Extension**: `.md` or `.mdc`  
**Frontmatter**: Optional metadata for filtering

**Minimal Example:**
```markdown
# Coding Standards

- Use consistent formatting
- Write meaningful comments
```

**With Frontmatter:**
```markdown
---
languages:
  - go
  - python
stage: implementation
---

# Implementation Standards

- Write unit tests for all functions
- Handle errors explicitly
```

**Rules:**
- Frontmatter is optional
- Content is pure Markdown
- Top-level YAML fields only (no nested objects for selectors)
- Multiple values use YAML arrays

### 3.3 Task Files

**Location**: `.agents/tasks/`  
**Extension**: `.md`  
**Frontmatter**: Optional metadata (automatically included in output)  
**Matching**: By filename without `.md` extension

**Example:**
```markdown
---
task_name: fix-bug
resume: false
languages:
  - go
selectors:
  stage: implementation
---

# Fix Bug: ${issue_number}

## Instructions

1. Analyze the issue described in ${issue_body}
2. Identify the root cause
3. Implement a minimal fix
4. Add regression tests

## Guidelines

- Make minimal changes
- Ensure backward compatibility
- Update documentation if needed
```

**Rules:**
- `task_name` field is optional metadata
- Tasks matched by filename (e.g., `fix-bug.md` → task name `fix-bug`)
- Frontmatter is used for filtering and metadata but not included in output
- Can specify `selectors` to auto-filter rules
- Supports parameter expansion: `${param_name}`

### 3.4 Command Files

**Location**: `.agents/commands/`, `.cursor/commands/`, `.opencode/command/`  
**Extension**: `.md`  
**Invocation**: Via slash syntax in task content

**Example (`pre-deploy.md`):**
```markdown
---
expand: true
---

# Pre-deployment Checklist

- Run tests: !`npm test`
- Check status: !`git status`
- Verify build: !`make build`
```

**Usage in Task:**
```markdown
# Deploy Application

/pre-deploy

Deploy to ${environment}.

/post-deploy
```

**Rules:**
- Commands are referenced, not matched by name
- Support parameter expansion
- Support inline parameters: `/command param="value"`
- Recursive expansion is prevented (single pass)

### 3.5 Skill Files

**Location**: `.agents/skills/{skill-name}/SKILL.md`  
**Extension**: `.md`  
**Filename**: Must be `SKILL.md`

**Example:**
```markdown
---
name: data-analysis
description: Analyze datasets, generate charts, and create summary reports. Use when working with CSV, Excel, or tabular data.
license: MIT
compatibility: Requires Python 3.8+ with pandas and matplotlib
metadata:
  author: team-name
  version: "1.0"
  tags:
    - data
    - visualization
---

# Data Analysis Skill

## When to Use

Use this skill when you need to:
- Analyze CSV or Excel files
- Generate charts and visualizations
- Calculate statistics and summaries

## How to Analyze Data

1. Load data with pandas:
   ```python
   import pandas as pd
   df = pd.read_csv('data.csv')
   ```

2. Generate summary statistics:
   ```python
   summary = df.describe()
   ```

3. Create visualizations:
   ```python
   import matplotlib.pyplot as plt
   df.plot(kind='bar')
   plt.savefig('chart.png')
   ```
```

**Output Format (XML):**
```xml
<available_skills>
  <skill>
    <name>data-analysis</name>
    <description>Analyze datasets, generate charts...</description>
    <location>/path/to/.agents/skills/data-analysis/SKILL.md</location>
  </skill>
</available_skills>
```

**Rules:**
- `name` (required): 1-64 characters, skill identifier
- `description` (required): 1-1024 characters, what the skill does
- `license` (optional): License information
- `compatibility` (optional): Max 500 characters, environment requirements
- `metadata` (optional): Arbitrary key-value pairs
- Only metadata output in initial context (progressive disclosure)
- AI agents load full content from provided location when needed

---

## 4. Directory Structure

### 4.1 Project-Level Structure

```
project/
├── .agents/
│   ├── rules/          # General rules (*.md, *.mdc)
│   ├── tasks/          # Task definitions (*.md)
│   ├── commands/       # Reusable commands (*.md)
│   └── skills/         # Specialized skills
│       ├── skill-1/
│       │   └── SKILL.md
│       └── skill-2/
│           └── SKILL.md
├── .cursor/
│   ├── rules/          # Cursor-specific rules
│   └── commands/       # Cursor commands
├── .github/
│   ├── copilot-instructions.md
│   └── agents/         # GitHub Copilot rules
├── .opencode/
│   ├── agent/          # OpenCode rules
│   └── command/        # OpenCode commands
├── .augment/
│   ├── rules/
│   └── guidelines.md
├── .windsurf/
│   └── rules/
├── .gemini/
│   └── styleguide.md
├── AGENTS.md           # Generic rules file
├── CLAUDE.md           # Claude-specific rules
├── CLAUDE.local.md     # Local Claude overrides
├── GEMINI.md           # Gemini-specific rules
├── .cursorrules        # Cursor rules file
└── .windsurfrules      # Windsurf rules file
```

### 4.2 User-Level Structure

```
~/ (home directory)
├── .agents/
│   ├── rules/          # User-wide rules
│   └── tasks/          # User-wide tasks
├── .claude/
│   └── CLAUDE.md       # User's Claude rules
├── .codex/
│   └── AGENTS.md       # User's Codex rules
├── .gemini/
│   └── GEMINI.md       # User's Gemini rules
└── .opencode/
    └── rules/          # User's OpenCode rules
```

### 4.3 Remote Directories

Remote directories follow the same structure as local directories and can be sourced via:

- HTTP/HTTPS URLs
- Git repositories (`git::https://...`)
- S3 buckets (`s3::https://...`)
- Local file paths (`file://...`)

---

## 5. Frontmatter Specification

### 5.1 YAML Frontmatter Format

Frontmatter uses standard YAML syntax enclosed in triple dashes:

```yaml
---
field1: value1
field2:
  - value2a
  - value2b
field3: "string value"
---
```

**Rules:**
- Must start and end with `---` on their own lines
- Must be valid YAML
- Must be at the beginning of the file
- Parsing errors should be logged but not fatal
- Only top-level fields are accessible for selectors

### 5.2 Standard Task Fields

#### 5.2.1 `task_name` (optional)
- **Type:** String
- **Purpose:** Metadata identifier for the task
- **Note:** Tasks are matched by filename, not this field

```yaml
---
task_name: fix-bug
---
```

#### 5.2.2 `resume` (optional)
- **Type:** Boolean
- **Purpose:** Indicates if task is for resuming work
- **Default:** `false`
- **Usage:** Selected with `-r` flag or `-s resume=true`

```yaml
---
resume: true
---
```

#### 5.2.3 `languages` (optional)
- **Type:** Array of strings (recommended) or string
- **Purpose:** Metadata about programming languages
- **Note:** Metadata only, does not auto-filter rules
- **Values:** Lowercase language names

```yaml
---
languages:
  - go
  - python
---
```

#### 5.2.4 `agent` (optional)
- **Type:** String
- **Purpose:** Target agent, acts as default selector
- **Values:** `cursor`, `copilot`, `claude`, `gemini`, `opencode`, `augment`, `windsurf`, `codex`

```yaml
---
agent: cursor
---
```

#### 5.2.5 `model` (optional)
- **Type:** String
- **Purpose:** AI model identifier (metadata only)

```yaml
---
model: anthropic.claude-sonnet-4-20250514-v1-0
---
```

#### 5.2.6 `single_shot` (optional)
- **Type:** Boolean
- **Purpose:** Indicates single vs. multi-execution
- **Note:** Metadata only

```yaml
---
single_shot: true
---
```

#### 5.2.7 `timeout` (optional)
- **Type:** String (Go time.Duration format)
- **Purpose:** Task execution timeout
- **Note:** Metadata only

```yaml
---
timeout: 10m
---
```

#### 5.2.8 `selectors` (optional)
- **Type:** Map of key-value pairs
- **Purpose:** Auto-filter rules for this task
- **Supports:** Scalar values and arrays (OR logic)

```yaml
---
selectors:
  languages: go
  stage: implementation
---
```

**With OR logic:**
```yaml
---
selectors:
  languages: [go, python, rust]
  stage: testing
---
```

#### 5.2.9 `expand` (optional)
- **Type:** Boolean
- **Purpose:** Control parameter expansion
- **Default:** `true`

```yaml
---
expand: false
---
```

### 5.3 Standard Rule Fields

#### 5.3.1 `languages` (optional)
- **Type:** Array or string
- **Purpose:** Filter rules by programming language
- **Values:** Lowercase language names

```yaml
---
languages:
  - go
---
```

#### 5.3.2 `stage` (optional)
- **Type:** String
- **Purpose:** Filter by development stage
- **Common values:** `planning`, `implementation`, `testing`, `review`

```yaml
---
stage: implementation
---
```

#### 5.3.3 `agent` (optional)
- **Type:** String
- **Purpose:** Target specific agent

```yaml
---
agent: cursor
---
```

#### 5.3.4 `mcp_server` (optional)
- **Type:** Object
- **Purpose:** Model Context Protocol server configuration

```yaml
---
mcp_server:
  command: python
  args: ["-m", "server"]
  env:
    PYTHON_PATH: /usr/bin/python3
---
```

### 5.4 Standard Skill Fields

#### 5.4.1 `name` (required)
- **Type:** String
- **Length:** 1-64 characters
- **Purpose:** Skill identifier

#### 5.4.2 `description` (required)
- **Type:** String
- **Length:** 1-1024 characters
- **Purpose:** What the skill does and when to use it

#### 5.4.3 `license` (optional)
- **Type:** String
- **Purpose:** License information

#### 5.4.4 `compatibility` (optional)
- **Type:** String
- **Max length:** 500 characters
- **Purpose:** Environment requirements

#### 5.4.5 `metadata` (optional)
- **Type:** Object
- **Purpose:** Arbitrary key-value pairs

### 5.5 Custom Fields

Any additional YAML fields can be used for custom selectors:

```yaml
---
environment: production
region: us-east-1
priority: high
---
```

**Rules:**
- Custom fields enable flexible filtering
- Only top-level fields work with selectors
- Nested objects are stored but not matchable

---

## 6. Content Expansion

Content expansion processes dynamic elements in file content. All expansions occur in a **single pass** to prevent injection attacks.

### 6.1 Parameter Expansion

**Syntax:** `${parameter_name}`

**Purpose:** Substitute values from command-line parameters

**Example:**
```markdown
Issue: ${issue_number}
Title: ${issue_title}
```

**With:** `-p issue_number=123 -p issue_title="Bug Fix"`

**Output:**
```markdown
Issue: 123
Title: Bug Fix
```

**Rules:**
- If parameter not found, placeholder remains unchanged
- Warning logged for missing parameters
- Disabled with `expand: false` in frontmatter

### 6.2 Command Expansion

**Syntax:** `` !`command` ``

**Purpose:** Execute shell commands and include output

**Example:**
```markdown
Current date: !`date +%Y-%m-%d`
Git branch: !`git rev-parse --abbrev-ref HEAD`
```

**Output:**
```markdown
Current date: 2025-12-25
Git branch: main
```

**Rules:**
- Executed with `sh -c`
- If command fails, syntax remains unchanged
- Warning logged for failures
- Output included verbatim (including trailing newlines)

**Security:** Only use with trusted files

### 6.3 Path Expansion

**Syntax:** `@path`

**Purpose:** Include file contents

**Example:**
```markdown
Configuration:
@config.yaml

Documentation:
@docs/api.md
```

**With spaces:**
```markdown
Content: @my\ file\ with\ spaces.txt
```

**Rules:**
- Path delimited by whitespace
- Use `\ ` to escape spaces
- If file not found, syntax remains unchanged
- Warning logged for missing files
- Content included verbatim

### 6.4 Slash Commands

**Syntax:** `/command-name` or `/command-name arg="value"`

**Purpose:** Include command file content

**Example:**
```markdown
/pre-deploy

Deploy to production.

/post-deploy env="production"
```

**Rules:**
- References command files by name (without `.md`)
- Searched in command directories
- Can pass inline parameters
- Command content is expanded and inserted
- Recursive expansion prevented

### 6.5 Expansion Order

1. **Slash commands** are expanded first
2. **Parameter expansion** processes `${...}`
3. **Command expansion** processes `` !`...` ``
4. **Path expansion** processes `@...`
5. Expanded content is **never re-processed**

---

## 7. Selector System

### 7.1 Purpose

Selectors filter rules and tasks based on frontmatter metadata, enabling context-specific rule inclusion.

### 7.2 Selector Syntax

**Command-line:**
```bash
-s key=value
-s languages=go
-s stage=implementation
```

**Task frontmatter:**
```yaml
---
selectors:
  languages: go
  stage: implementation
---
```

### 7.3 Matching Rules

**Simple match:**
```yaml
# Rule frontmatter
---
languages:
  - go
---

# Selector: -s languages=go
# Result: ✅ Match
```

**Array match (any value):**
```yaml
# Rule frontmatter
---
languages:
  - go
  - python
---

# Selector: -s languages=go
# Result: ✅ Match (go is in array)

# Selector: -s languages=rust
# Result: ❌ No match
```

**Multiple selectors (AND logic):**
```bash
# Requires: languages=go AND stage=implementation
-s languages=go -s stage=implementation
```

**OR logic (task frontmatter):**
```yaml
---
selectors:
  languages: [go, python, rust]
---
# Matches rules with languages=go OR python OR rust
```

### 7.4 Special Selectors

#### 7.4.1 `rule_name`
Filter by rule filename (without extension):

```yaml
---
selectors:
  rule_name: [security-standards, go-best-practices]
---
```

#### 7.4.2 `resume`
Filter for resume-specific tasks:

```bash
# Equivalent
-r
-s resume=true
```

### 7.5 Selector Precedence

1. Command-line `-s` flags
2. Task frontmatter `selectors` field
3. Combined with AND logic

**Example:**
```bash
# Task has: selectors.language = go
# Command: -s stage=testing
# Result: language=go AND stage=testing
```

### 7.6 Rules Without Frontmatter

- Rules without frontmatter are **always included** (unless resume mode)
- Cannot be filtered by selectors
- Considered universal rules

---

## 8. Bootstrap Scripts

### 8.1 Purpose

Bootstrap scripts prepare the environment before context assembly, enabling:
- Tool installation
- Data fetching
- Environment validation
- Dynamic setup

### 8.2 Naming Convention

```
{base-filename}-bootstrap
```

**Examples:**
- Rule: `jira-context.md` → Bootstrap: `jira-context-bootstrap`
- Task: `fix-bug.md` → Bootstrap: `fix-bug-bootstrap`

### 8.3 Requirements

- Must be executable (`chmod +x`)
- Located in same directory as associated file
- Extension-less (no `.sh`, `.py`, etc.)

### 8.4 Execution

**Rule bootstraps:**
- Execute before rule content is processed
- Run sequentially (not in parallel)

**Task bootstraps:**
- Execute after all rules processed
- Run before task content is emitted

### 8.5 Output

- **stdout**: Ignored (not included in context)
- **stderr**: Logged for debugging
- **Exit code**: Non-zero logs warning but continues

**Example:**
```bash
#!/bin/bash
# jira-context-bootstrap

if ! command -v jira-cli &> /dev/null
then
    echo "Installing jira-cli..." >&2
    # Installation commands
fi

echo "Fetching issue data..." >&2
# Fetch commands
```

---

## 9. Agent Targeting

### 9.1 Purpose

Support multiple AI coding agents with agent-specific rules and configuration.

### 9.2 Supported Agents

- `cursor` - Cursor IDE
- `copilot` - GitHub Copilot
- `claude` - Anthropic Claude
- `gemini` - Google Gemini
- `opencode` - OpenCode.ai
- `augment` - Augment
- `windsurf` - Windsurf
- `codex` - Codex

### 9.3 Agent Selection

**Command-line flag:**
```bash
-a agent-name
```

**Task frontmatter (overrides flag):**
```yaml
---
agent: cursor
---
```

### 9.4 Agent-Specific Rules

Rules can target specific agents:

```yaml
---
agent: cursor
---

# Cursor-specific coding standards
```

**Filtering behavior:**
- Rules with matching `agent` field are included
- Rules with no `agent` field are included (universal)
- Rules with different `agent` value are excluded

### 9.5 Write Rules Mode

With `-w` flag, rules are written to agent's user configuration:

```bash
coding-context -a copilot -w task-name
```

Writes rules to: `~/.github/agents/AGENTS.md`

**Agent-specific paths:**
- `cursor`: `~/.cursor/rules/`
- `copilot`: `~/.github/agents/`
- `claude`: `~/.claude/`
- `gemini`: `~/.gemini/`
- etc.

---

## 10. Discovery and Search Paths

### 10.1 Search Path Order

1. Directories specified via `-d` flags (in order)
2. Working directory (auto-added): `.`, parent dirs for some files
3. User home directory (auto-added): `~`

### 10.2 Task Discovery

**Search locations (in order):**
```
./.agents/tasks/*.md
~/.agents/tasks/*.md
```

**Matching:**
- By filename without `.md` extension
- `task_name` in frontmatter is optional metadata
- First match wins (unless selectors disambiguate)

### 10.3 Command Discovery

**Search locations:**
```
./.agents/commands/*.md
./.cursor/commands/*.md
./.opencode/command/*.md
```

### 10.4 Skill Discovery

**Search locations:**
```
./.agents/skills/*/SKILL.md
```

**Structure:**
```
.agents/skills/
├── skill-1/
│   └── SKILL.md
└── skill-2/
    └── SKILL.md
```

### 10.5 Rule Discovery

**Standard locations (searched in each directory):**

**Agent-specific directories:**
```
.agents/rules/
.cursor/rules/
.augment/rules/
.windsurf/rules/
.opencode/agent/
.github/agents/
```

**Agent-specific files:**
```
AGENTS.md
CLAUDE.md
CLAUDE.local.md
GEMINI.md
.cursorrules
.windsurfrules
.github/copilot-instructions.md
.gemini/styleguide.md
.augment/guidelines.md
```

**User-level:**
```
~/.agents/rules/
~/.claude/CLAUDE.md
~/.codex/AGENTS.md
~/.gemini/GEMINI.md
~/.opencode/rules/
```

### 10.6 Remote Directories

Remote directories support (via go-getter):

```bash
-d git::https://github.com/org/repo.git
-d https://example.com/rules.tar.gz
-d s3::https://s3.amazonaws.com/bucket/path
-d file:///absolute/path
```

**Features:**
- Downloaded to temporary location
- Processed like local directories
- Cleaned up after execution
- Support same directory structure

---

## 11. Context Assembly

### 11.1 Assembly Order

1. **Rule content** (all included rules, content only)
2. **Skill metadata** (XML format, if skills found)
3. **Task content** (with expansions applied)
4. **User prompt** (if provided, after `---` delimiter)

**Note**: Task frontmatter is used for filtering and metadata purposes but is not included in the output.

### 11.2 Output Format

**To stdout (the context):**
```markdown
# Rule 1 Content

Rule 1 text...

# Rule 2 Content

Rule 2 text...

# Skills

You have access to the following skills. Skills are specialized capabilities that provide domain expertise, workflows, and procedural knowledge. When a task matches a skill's description, you can load the full skill content by reading the SKILL.md file at the location provided.

<available_skills>
  <skill>
    <name>data-analysis</name>
    <description>...</description>
    <location>...</location>
  </skill>
</available_skills>

# Task Content

Fix bug #123...

---
User prompt text if provided.
```

**To stderr (metadata):**
```
INFO: Found 5 rule files
INFO: Selected task: fix-bug
INFO: Executing bootstrap: jira-context-bootstrap
INFO: Estimated tokens: 2,345
```

### 11.3 Resume Mode

With `-r` flag:
1. All rules are **skipped**
2. Implicit selector: `-s resume=true` added
3. Useful for continuing work with established context

### 11.4 Token Estimation

Rough estimate logged to stderr:

```
tokens ≈ (characters / 4) + (words / 0.75)
```

---

## 12. Compatibility

### 12.1 Supported AI Agents

- **Anthropic Claude**: `CLAUDE.md`, `.claude/`
- **GitHub Copilot**: `.github/copilot-instructions.md`, `.github/agents/`
- **Cursor**: `.cursor/rules`, `.cursorrules`
- **OpenCode.ai**: `.opencode/agent`, `.opencode/command`, `.opencode/rules`
- **Augment**: `.augment/rules`, `.augment/guidelines.md`
- **Windsurf**: `.windsurf/rules`, `.windsurfrules`
- **Google Gemini**: `GEMINI.md`, `.gemini/styleguide.md`
- **Codex**: `AGENTS.md`, `.codex/AGENTS.md`

### 12.2 Backward Compatibility

The standard is designed to be backward compatible with existing agent configuration files:

- Generic files like `AGENTS.md` work across agents
- Agent-specific files use their native conventions
- Standard supports both shared and agent-specific rules

### 12.3 Forward Compatibility

- Unknown frontmatter fields are ignored
- New standard fields can be added without breaking existing files
- Implementations should gracefully handle missing optional fields

---

## 13. Examples

### 13.1 Minimal Task

**File:** `.agents/tasks/hello.md`
```markdown
# Hello World Task

This is a minimal task with no frontmatter.
```

**Usage:**
```bash
coding-context hello
```

### 13.2 Task with Parameters

**File:** `.agents/tasks/fix-bug.md`
```markdown
---
task_name: fix-bug
languages:
  - go
---

# Fix Bug #${issue_number}

**Title:** ${issue_title}
**Description:** ${issue_body}

Fix this bug following Go best practices.
```

**Usage:**
```bash
coding-context \
  -p issue_number=123 \
  -p issue_title="Crash on startup" \
  -p issue_body="Application crashes when..." \
  fix-bug
```

### 13.3 Task with Selectors

**File:** `.agents/tasks/implement-feature.md`
```markdown
---
selectors:
  languages: go
  stage: implementation
---

# Implement Feature

Follow Go implementation standards and write tests.
```

**Effect:** Automatically filters to rules with `languages=go` AND `stage=implementation`

### 13.4 Rule with Filtering

**File:** `.agents/rules/go-standards.md`
```markdown
---
languages:
  - go
stage: implementation
---

# Go Implementation Standards

- Use `gofmt` for formatting
- Handle errors explicitly
- Write table-driven tests
```

**Selected by:**
```bash
coding-context -s languages=go -s stage=implementation task-name
```

### 13.5 Command Usage

**File:** `.agents/commands/check-tests.md`
```markdown
# Test Status

Tests: !`go test ./... -v`
Coverage: !`go test -cover ./...`
```

**Used in task:**
```markdown
# Verify Implementation

/check-tests

Ensure all tests pass before proceeding.
```

### 13.6 Skill Definition

**File:** `.agents/skills/data-analysis/SKILL.md`
```markdown
---
name: data-analysis
description: Analyze CSV/Excel data, generate charts, calculate statistics
---

# Data Analysis Skill

Use pandas and matplotlib for data analysis...
```

**Output in context:**
```xml
<available_skills>
  <skill>
    <name>data-analysis</name>
    <description>Analyze CSV/Excel data, generate charts...</description>
    <location>/path/to/.agents/skills/data-analysis/SKILL.md</location>
  </skill>
</available_skills>
```

### 13.7 Bootstrap Script

**File:** `.agents/rules/jira-context-bootstrap`
```bash
#!/bin/bash
# Install jira-cli if needed

if ! command -v jira-cli &> /dev/null
then
    echo "Installing jira-cli..." >&2
    pip install jira-cli >&2
fi

echo "Bootstrap complete" >&2
```

**Make executable:**
```bash
chmod +x .agents/rules/jira-context-bootstrap
```

---

## 14. Rationale and Design Principles

### 14.1 Design Principles

#### 14.1.1 Convention over Configuration
- Predetermined search paths reduce boilerplate
- Standard file extensions and names
- Implicit behaviors (e.g., auto-adding working directory)

#### 14.1.2 Simplicity
- Markdown for human readability
- YAML for structured metadata
- Single binary, no runtime dependencies

#### 14.1.3 Composability
- Mix rules from multiple sources
- Layer project, team, and personal rules
- Output to stdout for piping

#### 14.1.4 Flexibility
- Custom frontmatter fields
- Multiple filtering dimensions
- Support diverse workflows

#### 14.1.5 Security
- Single-pass expansion prevents injection
- Bootstrap output isolated from context
- Explicit execution permissions required

### 14.2 Key Design Decisions

#### 14.2.1 Filename Matching for Tasks
Tasks are matched by filename, not `task_name` field, because:
- Simpler mental model (filename = identifier)
- Avoids conflicts (filename must be unique in directory)
- Easier to discover (just list directory)

#### 14.2.2 Top-Level Frontmatter Only
Selectors only match top-level YAML fields because:
- Simpler implementation
- Predictable behavior
- Encourages flat, readable metadata
- Sufficient for most use cases

#### 14.2.3 Single-Pass Expansion
All expansions occur in one pass to:
- Prevent injection attacks
- Ensure predictable behavior
- Simplify implementation
- Make debugging easier

#### 14.2.4 Bootstrap Output to Stderr
Bootstrap scripts output to stderr (not context) because:
- Separates setup from content
- Allows logging without polluting context
- Enables verification without affecting AI input

#### 14.2.5 Task Frontmatter Usage
Task frontmatter serves metadata purposes:
- Enables task selection and filtering via selectors
- Specifies agent preferences
- Controls workflow behavior (e.g., resume mode)
- Defines rule filtering via `selectors` field

The frontmatter is not included in the output to keep the generated context focused on actionable content.

### 14.3 Limitations

#### 14.3.1 No Nested Selector Matching
Selectors only match top-level YAML fields. This is intentional for simplicity and predictability.

#### 14.3.2 No Native OR Logic in Command-Line
Command-line selectors use AND logic only. OR logic is available via array values in task frontmatter selectors.

#### 14.3.3 No Rule Ordering Guarantees
Rules are included in filesystem order. If order matters, combine into a single file or use a manifest.

#### 14.3.4 Static Bootstrap
Bootstrap scripts run once at assembly time, not dynamically during agent execution.

### 14.4 Future Considerations

Potential future additions while maintaining backward compatibility:

- Rule ordering hints (e.g., `order: 1`)
- Include/exclude patterns for rules
- Nested selector support
- Rule dependencies
- Versioned manifests
- Encrypted parameters

---

## Appendix A: YAML Frontmatter Schema (JSON Schema)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Coding Context Frontmatter",
  "type": "object",
  "properties": {
    "task_name": {
      "type": "string",
      "description": "Optional task identifier (metadata only)"
    },
    "resume": {
      "type": "boolean",
      "description": "Indicates resume-specific task"
    },
    "languages": {
      "oneOf": [
        {"type": "string"},
        {"type": "array", "items": {"type": "string"}}
      ],
      "description": "Programming languages (metadata or selector)"
    },
    "language": {
      "type": "string",
      "description": "Alias for languages (singular form)"
    },
    "stage": {
      "type": "string",
      "description": "Development stage",
      "enum": ["implementation", "planning", "review", "testing"]
    },
    "agent": {
      "type": "string",
      "description": "Target AI agent",
      "enum": ["augment", "claude", "codex", "copilot", "cursor", "gemini", "opencode", "windsurf"]
    },
    "model": {
      "type": "string",
      "description": "AI model identifier"
    },
    "single_shot": {
      "type": "boolean",
      "description": "Single vs. multi-execution"
    },
    "timeout": {
      "type": "string",
      "description": "Timeout duration (Go time.Duration format)"
    },
    "expand": {
      "type": "boolean",
      "description": "Enable/disable parameter expansion",
      "default": true
    },
    "selectors": {
      "type": "object",
      "description": "Auto-filter rules",
      "additionalProperties": {
        "oneOf": [
          {"type": "string"},
          {"type": "array", "items": {"type": "string"}}
        ]
      }
    },
    "mcp_server": {
      "type": "object",
      "description": "Model Context Protocol server config",
      "properties": {
        "command": {"type": "string"},
        "args": {"type": "array", "items": {"type": "string"}},
        "env": {"type": "object", "additionalProperties": {"type": "string"}}
      }
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 64,
      "description": "Skill name (required for skills)"
    },
    "description": {
      "type": "string",
      "minLength": 1,
      "maxLength": 1024,
      "description": "Skill description (required for skills)"
    },
    "license": {
      "type": "string",
      "description": "License information"
    },
    "compatibility": {
      "type": "string",
      "maxLength": 500,
      "description": "Environment requirements"
    },
    "metadata": {
      "type": "object",
      "description": "Arbitrary key-value pairs",
      "additionalProperties": true
    }
  },
  "additionalProperties": true
}
```

---

## Appendix B: File Extension Registry

| Extension | Type | Purpose |
|-----------|------|---------|
| `.md` | Markdown | Rules, tasks, commands, skills |
| `.mdc` | Markdown Context | Alternative rule extension |
| `-bootstrap` | Executable | Bootstrap scripts (no extension) |

---

## Appendix C: Standard Frontmatter Fields Reference

### Task Fields

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `task_name` | string | No | Metadata identifier |
| `resume` | boolean | No | Resume mode indicator |
| `languages` | array/string | No | Programming languages (metadata) |
| `agent` | string | No | Target agent (selector) |
| `model` | string | No | AI model identifier |
| `single_shot` | boolean | No | Execution mode |
| `timeout` | string | No | Timeout duration |
| `selectors` | object | No | Auto-filter rules |
| `expand` | boolean | No | Parameter expansion control |

### Rule Fields

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `languages` | array/string | No | Language filter |
| `stage` | string | No | Development stage filter |
| `agent` | string | No | Agent filter |
| `mcp_server` | object | No | MCP server config |

### Skill Fields

| Field | Type | Required | Purpose |
|-------|------|----------|---------|
| `name` | string | Yes | Skill identifier (1-64 chars) |
| `description` | string | Yes | What skill does (1-1024 chars) |
| `license` | string | No | License information |
| `compatibility` | string | No | Environment requirements (max 500 chars) |
| `metadata` | object | No | Arbitrary key-value pairs |

---

## Appendix D: Common Language Values

Use lowercase language identifiers in frontmatter:

- `c` - C
- `cpp` - C++
- `csharp` - C#
- `css` - CSS
- `dart` - Dart
- `elixir` - Elixir
- `go` - Go
- `haskell` - Haskell
- `html` - HTML
- `java` - Java
- `javascript` - JavaScript
- `kotlin` - Kotlin
- `lua` - Lua
- `markdown` - Markdown
- `objectivec` - Objective-C
- `php` - PHP
- `python` - Python
- `ruby` - Ruby
- `rust` - Rust
- `scala` - Scala
- `shell` - Shell
- `swift` - Swift
- `typescript` - TypeScript
- `yaml` - YAML

---

## Appendix E: Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-12-25 | Initial specification |

---

## Appendix F: References

- [Coding Context CLI Repository](https://github.com/kitproj/coding-context-cli)
- [Documentation Site](https://kitproj.github.io/coding-context-cli/)
- [YAML Specification](https://yaml.org/spec/)
- [Markdown Specification](https://commonmark.org/)
- [Model Context Protocol](https://modelcontextprotocol.io/)

---

## License

This specification is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
