---
marp: true
theme: default
paginate: true
backgroundColor: #fff
backgroundImage: url('https://marp.app/assets/hero-background.svg')
---

<!-- _class: invert -->

# Coding Context CLI

### Dynamic Context Assembly for AI Coding Agents

**A command-line tool that assembles rich context for AI models**

---

## What Problem Does It Solve?

**Challenge**: AI coding agents need comprehensive context to make informed decisions

- Project-specific coding standards
- Repository architecture and conventions
- Technology stack guidelines
- Task-specific requirements
- Team practices

**Problem**: Manually assembling context is tedious and error-prone

**Solution**: Coding Context CLI automates context assembly

---

## How It Works

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│              │    │              │    │              │
│  Rule Files  │───▶│  Context CLI │───▶│  AI Agent    │
│              │    │              │    │              │
└──────────────┘    └──────────────┘    └──────────────┘
       │                   │                    │
   Guidelines         Assembly           Smart Decisions
   Standards          Filtering          Code Generation
   Best Practices     Parameters         Bug Fixes
```

The tool collects, filters, and combines context from multiple sources into a single output.

---

## Key Features

✅ **Dynamic Context Assembly** - Merges context from various source files
✅ **Task-Specific Prompts** - Different prompts for different tasks
✅ **Rule-Based Context** - Reusable context snippets (rules)
✅ **Frontmatter Filtering** - Select rules based on metadata
✅ **Bootstrap Scripts** - Fetch or generate context dynamically
✅ **Parameter Substitution** - Inject values into task prompts
✅ **Token Estimation** - Know your context size
✅ **Remote File Support** - Load rules from Git repos, HTTP, S3

---

## Supported AI Agents

Works with configuration files from various AI coding agents:

| Agent | Configuration Files |
|-------|-------------------|
| **Claude** | `CLAUDE.md`, `.claude/CLAUDE.md` |
| **GitHub Copilot** | `.github/copilot-instructions.md`, `.github/agents` |
| **Cursor** | `.cursor/rules`, `.cursorrules` |
| **Windsurf** | `.windsurf/rules`, `.windsurfrules` |
| **OpenCode.ai** | `.opencode/agent`, `.opencode/rules` |
| **Google Gemini** | `GEMINI.md`, `.gemini/styleguide.md` |
| **Generic** | `AGENTS.md`, `.agents/rules`, `.agents/tasks` |

---

## Installation

### Linux (AMD64)
```bash
sudo curl -fsL -o /usr/local/bin/coding-context \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_linux_amd64
sudo chmod +x /usr/local/bin/coding-context
```

### MacOS (Apple Silicon)
```bash
sudo curl -fsL -o /usr/local/bin/coding-context \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_darwin_arm64
sudo chmod +x /usr/local/bin/coding-context
```

Also available for ARM64 Linux, Intel Mac, and build from source!

---

## Basic Usage

```bash
# Simple usage with a task
coding-context /fix-bug | llm -m claude

# With parameters
coding-context -p issue_number=123 /fix-bug | llm

# With selectors (filter rules)
coding-context -s languages=go /implement-feature | llm

# Multiple selectors
coding-context -s languages=go -s stage=implementation /task | llm
```

**Output**: Combined context from rules + task prompt → fed to AI model

---

## Example: Fix a Bug

```bash
coding-context -p jira_issue_key=PROJ-1234 /fix-bug | llm -m gemini-pro
```

**What happens:**
1. ✅ Finds task file: `fix-bug.md` in `.agents/tasks/`
2. ✅ Collects all rule files from search paths
3. ✅ Filters rules based on selectors (if any)
4. ✅ Runs bootstrap scripts (if any)
5. ✅ Substitutes `${jira_issue_key}` with `PROJ-1234`
6. ✅ Outputs combined context to stdout
7. ✅ Pipes to AI model for processing

---

## Rule Files

**Rules** are reusable context snippets (Markdown files with optional YAML frontmatter)

**Example**: `.agents/rules/backend.md`
```markdown
---
languages:
  - go
stage: implementation
---

# Backend Coding Standards

- All new code must be accompanied by unit tests
- Use the standard logging library
- Follow error handling best practices
```

**Usage**: `coding-context -s languages=go -s stage=implementation /task`

---

## Task Files

**Tasks** are specific prompts for different coding activities

**Example**: `.agents/tasks/fix-bug.md`
```markdown
---
task_name: fix-bug
---

# Task: Fix Bug in ${jira_issue_key}

Analyze the following files and provide a fix:
- Root cause analysis
- Proposed solution
- Test cases to prevent regression
```

**Usage**: `coding-context -p jira_issue_key=PROJ-123 /fix-bug`

---

## Frontmatter Filtering

**Frontmatter** = YAML metadata at the top of files

**Use selectors to filter rules:**
```bash
# Only include rules with languages=go
coding-context -s languages=go /task

# Multiple selectors (AND logic)
coding-context -s languages=go -s stage=testing /task

# OR logic (in frontmatter)
---
languages:
  - go
  - python
---
```

**Matches**: Rules where `languages` includes `go` OR `python`

---

## Task Selectors

Tasks can automatically apply selectors via frontmatter:

**Example**: `.agents/tasks/implement-go-feature.md`
```markdown
---
task_name: implement-feature
selectors:
  languages: go
  stage: implementation
---

# Implement Feature

Implement following Go best practices...
```

**Usage**: `coding-context /implement-feature`
**Result**: Automatically includes only rules matching `languages=go AND stage=implementation`

---

## Bootstrap Scripts

**Bootstrap scripts** prepare the environment before processing

**Example**: `.agents/rules/jira-bootstrap`
```bash
#!/bin/bash
# Install jira-cli if not present
if ! command -v jira-cli &> /dev/null
then
    echo "Installing jira-cli..." >&2
    npm install -g jira-cli
fi
```

**Naming**: `<rule-name>-bootstrap` or `<task-name>-bootstrap`
**Output**: Goes to `stderr`, not included in AI context
**Use cases**: Installing tools, fetching data, environment setup

---

## Remote File Support

Load rules and tasks from remote locations:

```bash
# Load from Git repository
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  /fix-bug

# Load from multiple sources
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  -d https://cdn.company.com/coding-standards \
  /deploy

# Mix local and remote
coding-context \
  -d git::https://github.com/company/rules.git \
  -s languages=go \
  /implement-feature
```

**Supports**: Git, HTTP/HTTPS, S3, and more (via go-getter)

---

## Resume Mode

**Resume mode** (`-r`) is for continuing work with existing context:

```bash
# Initial invocation (with all rules)
coding-context -s resume=false /fix-bug | ai-agent

# Resume (skip rules, use resume-specific task)
coding-context -r /fix-bug | ai-agent
```

**What it does:**
- ✅ Skips all rule files (saves tokens)
- ✅ Automatically adds `-s resume=true` selector
- ✅ Selects task with `resume: true` in frontmatter

**Use case**: Continue multi-turn conversations without re-sending rules

---

## Agent-Specific Mode

**Problem**: AI agents read their own config files
**Solution**: Use `-a` flag to exclude agent's own paths

```bash
# Using Cursor: exclude .cursor/ paths (Cursor reads them)
# But include rules from other agents
coding-context -a cursor /fix-bug
```

**Supported agents**: `cursor`, `opencode`, `copilot`, `claude`, `gemini`, `augment`, `windsurf`, `codex`

**Benefits:**
- Avoids duplication
- Includes cross-agent rules
- Reduces token usage

---

## File Search Paths

The tool searches multiple locations automatically:

**Rules:**
- Current directory: `.agents/rules/`, `.cursor/rules/`, `.github/agents/`
- Parent directories: `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`
- User home: `~/.agents/rules/`, `~/.claude/CLAUDE.md`
- System-wide: `/etc/agents/`

**Tasks:**
- `.agents/tasks/*.md`
- `~/.agents/tasks/*.md`

**Commands** (referenced in tasks):
- `.agents/commands/*.md`
- `.cursor/commands/*.md`

---

## Agentic Workflows Integration

**Coding Context CLI** complements **GitHub Agentic Workflows**

```
┌─────────────────────────────────────────┐
│      Agentic Workflow Ecosystem          │
├─────────────────────────────────────────┤
│                                          │
│  Context Layer          Execution Layer │
│  ┌────────────────┐   ┌──────────────┐  │
│  │ Coding Context │──▶│ GitHub       │  │
│  │ CLI            │   │ Actions      │  │
│  │                │   │              │  │
│  │ Rules          │   │ Workflows    │  │
│  │ Guidelines     │   │ AI Agents    │  │
│  │ Tasks          │   │ Steps        │  │
│  └────────────────┘   └──────────────┘  │
│           │                    │         │
│           └──────┬─────────────┘         │
│                  ▼                        │
│          ┌──────────────┐                │
│          │  AI Agent    │                │
│          └──────────────┘                │
└─────────────────────────────────────────┘
```

---

## GitHub Actions Example

```yaml
name: AI Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Coding Context CLI
        run: |
          curl -fsL -o /usr/local/bin/coding-context \
            https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_linux_amd64
          chmod +x /usr/local/bin/coding-context
      
      - name: Assemble Context
        run: |
          coding-context \
            -s task=code-review \
            -p pr_number=${{ github.event.pull_request.number }} \
            /code-review > context.txt
      
      - name: Run AI Review
        run: cat context.txt | llm -m claude-3-opus > review.md
```

---

## Use Cases

**1. Code Reviews**
```bash
coding-context -p pr_number=123 /review-pull-request | llm
```

**2. Bug Fixes**
```bash
coding-context -p issue_id=PROJ-456 /fix-bug | llm
```

**3. Feature Implementation**
```bash
coding-context -s languages=go /implement-feature | llm
```

**4. Documentation**
```bash
coding-context /enhance-docs | llm
```

**5. CI/CD Integration**
```bash
coding-context -s stage=testing /run-tests | llm
```

---

## Advanced Features

**Token Estimation**
```bash
coding-context /task 2>&1 | grep "Total tokens"
# Output: Total tokens: ~2,450
```

**Manifest Files** (centralized search paths)
```bash
coding-context -m https://company.com/paths.txt /task
```

**Change Directory**
```bash
coding-context -C /path/to/project /task
```

**Multiple Parameters**
```bash
coding-context \
  -p env=staging \
  -p version=1.2.3 \
  -p author=alice \
  /deploy
```

---

## Example Project Structure

```
my-project/
├── .agents/
│   ├── rules/
│   │   ├── go-standards.md          # languages: [go]
│   │   ├── python-standards.md      # languages: [python]
│   │   └── testing-guidelines.md    # stage: testing
│   ├── tasks/
│   │   ├── fix-bug.md
│   │   ├── implement-feature.md
│   │   └── code-review.md
│   └── commands/
│       └── run-tests.md
├── .github/
│   └── copilot-instructions.md
├── CLAUDE.md                         # Claude-specific rules
└── AGENTS.md                         # Generic rules
```

---

## Best Practices

✅ **Organize rules by concern**: Language, stage, technology
✅ **Use descriptive task names**: `fix-bug`, `implement-feature`
✅ **Leverage selectors**: Filter rules for specific contexts
✅ **Keep rules focused**: One topic per rule file
✅ **Use lowercase language names**: `go`, `python`, `javascript`
✅ **Add bootstrap scripts**: For dynamic data fetching
✅ **Version control your rules**: Track changes over time
✅ **Share rules across teams**: Use remote directories
✅ **Estimate token usage**: Monitor with stderr output

---

## Language-Specific Rules

**Structure your rules by language:**

```markdown
# .agents/rules/go-standards.md
---
languages: [go]
---
- Use gofmt for formatting
- Handle all errors explicitly
- Write table-driven tests

# .agents/rules/python-standards.md
---
languages: [python]
---
- Follow PEP 8 style guide
- Use type hints
- Write docstrings
```

**Usage**: `coding-context -s languages=go /task`

---

## Multi-Stage Workflows

**Define rules for different stages:**

```markdown
# .agents/rules/planning.md
---
stage: planning
---
# Planning Guidelines
- Write detailed design docs
- Consider edge cases

# .agents/rules/implementation.md
---
stage: implementation
---
# Implementation Guidelines
- Write tests first
- Keep functions small
```

**Usage**: `coding-context -s stage=implementation /task`

---

## Real-World Example

**Scenario**: Fix a production bug in a Go microservice

```bash
coding-context \
  -s languages=go \
  -s stage=bugfix \
  -s environment=production \
  -p issue_id=PROD-789 \
  -p severity=critical \
  /fix-production-bug | llm -m claude-3-opus
```

**Result**: AI receives context with:
- Go coding standards
- Bugfix guidelines
- Production-specific rules
- Issue details
- Severity level

---

## Token Management

**Monitor token usage:**
```bash
coding-context /task 2>&1 | tee >(grep "tokens" >&2) | llm
```

**Output to stderr:**
```
Loading rule: .agents/rules/go-standards.md (450 tokens)
Loading rule: .agents/rules/testing.md (320 tokens)
Loading task: /fix-bug (180 tokens)
Total tokens: ~950
```

**Tip**: Use selectors and resume mode to reduce token count

---

## Comparison with Manual Context

**Without Coding Context CLI:**
```bash
# Manual copy-paste from multiple files
cat .agents/rules/go.md >> context.txt
cat .agents/rules/testing.md >> context.txt
cat .agents/tasks/fix-bug.md >> context.txt
sed -i 's/${issue_id}/PROJ-123/g' context.txt
cat context.txt | llm
```

**With Coding Context CLI:**
```bash
coding-context -s languages=go -p issue_id=PROJ-123 /fix-bug | llm
```

**Benefits**: Automated, consistent, repeatable, version-controlled

---

## Community and Support

📚 **Documentation**: https://kitproj.github.io/coding-context-cli/
🐙 **GitHub**: https://github.com/kitproj/coding-context-cli
📦 **Releases**: https://github.com/kitproj/coding-context-cli/releases
💬 **Issues**: https://github.com/kitproj/coding-context-cli/issues

**Contributing**: PRs welcome!
**License**: Apache 2.0

---

## Key Takeaways

1. **Automates context assembly** for AI coding agents
2. **Works with multiple AI agents** (Claude, Copilot, Cursor, etc.)
3. **Flexible filtering** with selectors and frontmatter
4. **Dynamic content** via bootstrap scripts
5. **Remote file support** for shared team rules
6. **Integrates with CI/CD** and agentic workflows
7. **Token-aware** for managing context size
8. **Simple CLI** that pipes to any AI model

---

<!-- _class: invert -->

# Get Started Today!

### Install
```bash
sudo curl -fsL -o /usr/local/bin/coding-context \
  https://github.com/kitproj/coding-context-cli/releases/download/v0.0.16/coding-context_v0.0.16_linux_amd64
sudo chmod +x /usr/local/bin/coding-context
```

### Try It
```bash
coding-context -p issue=123 /fix-bug | llm -m claude
```

### Learn More
**https://kitproj.github.io/coding-context-cli/**

---

<!-- _class: invert -->

# Questions?

**Documentation**: https://kitproj.github.io/coding-context-cli/
**GitHub**: https://github.com/kitproj/coding-context-cli

Thank you! 🎉
