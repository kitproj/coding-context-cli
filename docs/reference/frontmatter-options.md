---
layout: default
title: Frontmatter Options
parent: Reference
nav_order: 6
---

# Frontmatter Options Reference

This document provides a comprehensive list of frontmatter options that can be used in task files, rule files, command files, and skill files.

Frontmatter is YAML metadata placed at the beginning of markdown files between `---` delimiters. It controls file behavior, filtering, and metadata.

## Table of Contents

1. [Currently Implemented Options](#currently-implemented-options)
2. [Potential Future Options](#potential-future-options)
3. [Category Reference](#category-reference)

---

## Currently Implemented Options

These options are currently supported by the coding-context CLI.

### Task File Options

Task frontmatter controls task behavior and rule filtering.

#### `task_name` (optional)
- **Type:** String
- **Purpose:** Metadata identifier for the task
- **Note:** Tasks are matched by filename, not this field
- **Example:**
  ```yaml
  task_name: fix-bug
  ```

#### `agent` (optional)
- **Type:** String
- **Purpose:** Specifies the target AI agent and filters rules
- **Values:** `cursor`, `opencode`, `copilot`, `claude`, `gemini`, `augment`, `windsurf`, `codex`
- **Example:**
  ```yaml
  agent: copilot
  ```

#### `languages` (optional)
- **Type:** Array or String
- **Purpose:** Programming language(s) for metadata (does not filter)
- **Recommended:** Use lowercase array format
- **Example:**
  ```yaml
  languages:
    - go
    - python
  ```

#### `model` (optional)
- **Type:** String
- **Purpose:** AI model identifier for task execution
- **Example:**
  ```yaml
  model: anthropic.claude-sonnet-4-20250514-v1-0
  ```

#### `single_shot` (optional)
- **Type:** Boolean
- **Purpose:** Whether task runs once or multiple times
- **Default:** `false`
- **Example:**
  ```yaml
  single_shot: true
  ```

#### `timeout` (optional)
- **Type:** String (time.Duration format)
- **Purpose:** Task execution timeout
- **Format:** `30s`, `5m`, `1h`, `1h30m`
- **Example:**
  ```yaml
  timeout: 10m
  ```

#### `resume` (optional)
- **Type:** Boolean
- **Purpose:** Indicates if task is for resume mode
- **Use Case:** Continue work without re-sending rules
- **Example:**
  ```yaml
  resume: true
  ```

#### `selectors` (optional)
- **Type:** Map of key-value pairs
- **Purpose:** Automatically filter rules when task runs
- **Supports:** OR logic with arrays for same key
- **Example:**
  ```yaml
  selectors:
    stage: implementation
    languages: [go, python]
    priority: high
  ```

#### `expand` (optional)
- **Type:** Boolean
- **Purpose:** Controls parameter expansion (${variable})
- **Default:** `true`
- **Use Case:** Preserve templates for AI agents
- **Example:**
  ```yaml
  expand: false
  ```

### Rule File Options

Rule frontmatter controls rule filtering and metadata.

#### `rule_name` (optional)
- **Type:** String
- **Purpose:** Optional identifier for the rule file
- **Example:**
  ```yaml
  rule_name: go-testing-standards
  ```

#### `task_names` (optional)
- **Type:** Array
- **Purpose:** Specific tasks this rule applies to
- **Logic:** OR logic (matches any listed task)
- **Example:**
  ```yaml
  task_names:
    - implement-feature
    - refactor-code
  ```

#### `languages` (optional)
- **Type:** Array
- **Purpose:** Programming languages this rule applies to
- **Logic:** OR logic (matches any listed language)
- **Recommended:** Use lowercase values
- **Example:**
  ```yaml
  languages:
    - go
    - python
  ```

#### `agent` (optional)
- **Type:** String
- **Purpose:** AI agent this rule is intended for
- **Values:** `cursor`, `opencode`, `copilot`, `claude`, `gemini`, `augment`, `windsurf`, `codex`
- **Example:**
  ```yaml
  agent: cursor
  ```

#### `mcp_server` (optional)
- **Type:** Object
- **Purpose:** MCP (Model Context Protocol) server configuration
- **Use Case:** Define server processes for AI agents
- **Example:**
  ```yaml
  mcp_server:
    command: python
    args: ["-m", "server"]
    env:
      PYTHON_PATH: /usr/bin/python3
  ```

#### `expand` (optional)
- **Type:** Boolean
- **Purpose:** Controls parameter expansion in rule content
- **Default:** `true`
- **Example:**
  ```yaml
  expand: false
  ```

### Command File Options

Command frontmatter controls command behavior.

#### `expand` (optional)
- **Type:** Boolean
- **Purpose:** Controls parameter expansion in command content
- **Default:** `true`
- **Use Case:** Preserve template syntax
- **Example:**
  ```yaml
  expand: false
  ```

#### `selectors` (optional)
- **Type:** Map of key-value pairs
- **Purpose:** Filter commands or add to task selectors
- **Example:**
  ```yaml
  selectors:
    environment: production
  ```

### Skill File Options

Skill frontmatter defines skill metadata for progressive disclosure.

#### `name` (required)
- **Type:** String
- **Purpose:** Skill identifier
- **Constraints:** 1-64 characters, lowercase alphanumeric and hyphens
- **Example:**
  ```yaml
  name: data-analysis
  ```

#### `description` (required)
- **Type:** String
- **Purpose:** What the skill does and when to use it
- **Constraints:** 1-1024 characters
- **Example:**
  ```yaml
  description: Analyze datasets, generate charts, and create summary reports. Use when the user needs to work with CSV, Excel, or other tabular data formats.
  ```

#### `license` (optional)
- **Type:** String
- **Purpose:** License applied to the skill
- **Example:**
  ```yaml
  license: MIT
  ```

#### `compatibility` (optional)
- **Type:** String
- **Purpose:** Environment requirements
- **Constraints:** Max 500 characters
- **Example:**
  ```yaml
  compatibility: Requires Python 3.8+ with pandas and matplotlib
  ```

#### `metadata` (optional)
- **Type:** Map of string key-value pairs
- **Purpose:** Arbitrary metadata
- **Example:**
  ```yaml
  metadata:
    author: team-name
    version: "1.0"
    category: data-processing
  ```

#### `allowed-tools` (optional, experimental)
- **Type:** String (space-delimited list)
- **Purpose:** Pre-approved tools for the skill
- **Example:**
  ```yaml
  allowed-tools: pandas numpy matplotlib seaborn
  ```

---

## Potential Future Options

These options are not currently implemented but could be useful additions based on common patterns in static site generators, documentation tools, and AI/ML systems.

### General Metadata Options

These could apply to any file type (tasks, rules, commands, skills).

#### `title`
- **Type:** String
- **Purpose:** Human-readable title for the file
- **Use Case:** Display in UI, logs, or documentation
- **Example:**
  ```yaml
  title: Fix Critical Bug in Authentication
  ```

#### `description`
- **Type:** String
- **Purpose:** Brief description of the file's purpose
- **Use Case:** Tooltips, documentation, search
- **Example:**
  ```yaml
  description: Guidelines for implementing Go features with proper error handling
  ```

#### `author`
- **Type:** String or Array
- **Purpose:** Author(s) of the content
- **Use Case:** Attribution, contact information
- **Example:**
  ```yaml
  author: Jane Doe
  # or
  authors:
    - Jane Doe
    - John Smith
  ```

#### `created`
- **Type:** String (ISO 8601 date)
- **Purpose:** Creation date
- **Use Case:** Version tracking, freshness indicators
- **Example:**
  ```yaml
  created: 2024-01-15
  ```

#### `updated`
- **Type:** String (ISO 8601 date)
- **Purpose:** Last update date
- **Use Case:** Version tracking, freshness indicators
- **Example:**
  ```yaml
  updated: 2024-03-20
  ```

#### `version`
- **Type:** String (semantic version)
- **Purpose:** Version number for the file
- **Use Case:** Compatibility checking, change tracking
- **Example:**
  ```yaml
  version: 1.2.0
  ```

#### `tags`
- **Type:** Array
- **Purpose:** Categorization and search keywords
- **Use Case:** Filtering, grouping, search
- **Example:**
  ```yaml
  tags:
    - security
    - authentication
    - backend
  ```

#### `categories`
- **Type:** Array
- **Purpose:** High-level categorization
- **Use Case:** Organization, navigation
- **Example:**
  ```yaml
  categories:
    - backend
    - api
  ```

#### `deprecated`
- **Type:** Boolean
- **Purpose:** Mark content as deprecated
- **Use Case:** Migration warnings, version management
- **Example:**
  ```yaml
  deprecated: true
  ```

#### `deprecation_message`
- **Type:** String
- **Purpose:** Explanation for deprecation
- **Use Case:** Migration guidance
- **Example:**
  ```yaml
  deprecation_message: Use implement-feature-v2 instead. This version will be removed in v2.0.
  ```

### Task-Specific Options

#### `priority`
- **Type:** String or Integer
- **Purpose:** Task priority level
- **Values:** `critical`, `high`, `medium`, `low` or 1-5
- **Use Case:** Task ordering, resource allocation
- **Example:**
  ```yaml
  priority: high
  ```

#### `complexity`
- **Type:** String or Integer
- **Purpose:** Estimated task complexity
- **Values:** `simple`, `moderate`, `complex`, `very-complex` or 1-10
- **Use Case:** Resource planning, model selection
- **Example:**
  ```yaml
  complexity: moderate
  ```

#### `estimated_time`
- **Type:** String (time.Duration format)
- **Purpose:** Estimated time to complete
- **Use Case:** Planning, resource allocation
- **Example:**
  ```yaml
  estimated_time: 2h
  ```

#### `prerequisites`
- **Type:** Array
- **Purpose:** Tasks or conditions required before this task
- **Use Case:** Workflow dependencies, validation
- **Example:**
  ```yaml
  prerequisites:
    - setup-environment
    - review-requirements
  ```

#### `outputs`
- **Type:** Array
- **Purpose:** Expected outputs or artifacts
- **Use Case:** Validation, workflow planning
- **Example:**
  ```yaml
  outputs:
    - source code
    - unit tests
    - documentation
  ```

#### `context_files`
- **Type:** Array
- **Purpose:** Files that should be included in context
- **Use Case:** Auto-include relevant files
- **Example:**
  ```yaml
  context_files:
    - src/auth/*.go
    - docs/authentication.md
  ```

#### `max_tokens`
- **Type:** Integer
- **Purpose:** Maximum token budget for task
- **Use Case:** Cost control, model selection
- **Example:**
  ```yaml
  max_tokens: 100000
  ```

#### `temperature`
- **Type:** Float (0.0 to 2.0)
- **Purpose:** Model temperature parameter
- **Use Case:** Control creativity vs consistency
- **Example:**
  ```yaml
  temperature: 0.7
  ```

#### `requires_approval`
- **Type:** Boolean
- **Purpose:** Whether task requires human approval
- **Use Case:** Workflow control, safety
- **Example:**
  ```yaml
  requires_approval: true
  ```

#### `approval_criteria`
- **Type:** Array or String
- **Purpose:** Criteria for approval
- **Use Case:** Checklist, validation
- **Example:**
  ```yaml
  approval_criteria:
    - All tests pass
    - Code review completed
    - Security scan passed
  ```

#### `interactive`
- **Type:** Boolean
- **Purpose:** Whether task requires user interaction
- **Use Case:** Workflow planning
- **Example:**
  ```yaml
  interactive: true
  ```

#### `retry_on_failure`
- **Type:** Boolean or Integer
- **Purpose:** Whether to retry on failure (or max retries)
- **Use Case:** Resilience, automation
- **Example:**
  ```yaml
  retry_on_failure: 3
  ```

#### `fallback_task`
- **Type:** String
- **Purpose:** Alternative task if this one fails
- **Use Case:** Error handling, workflows
- **Example:**
  ```yaml
  fallback_task: manual-fix-bug
  ```

### Rule-Specific Options

#### `stage`
- **Type:** String or Array
- **Purpose:** Development stage(s) rule applies to
- **Values:** `planning`, `implementation`, `testing`, `review`, `deployment`
- **Use Case:** Phase-specific rules
- **Example:**
  ```yaml
  stage: implementation
  # or
  stages:
    - implementation
    - testing
  ```

#### `scope`
- **Type:** String or Array
- **Purpose:** Code scope rule applies to
- **Values:** `frontend`, `backend`, `database`, `api`, `ui`, `infrastructure`
- **Use Case:** Component-specific rules
- **Example:**
  ```yaml
  scope: backend
  ```

#### `severity`
- **Type:** String
- **Purpose:** Importance level of the rule
- **Values:** `error`, `warning`, `info`, `suggestion`
- **Use Case:** Rule enforcement level
- **Example:**
  ```yaml
  severity: error
  ```

#### `enforceable`
- **Type:** Boolean
- **Purpose:** Whether rule can be automatically enforced
- **Use Case:** Linting, validation
- **Example:**
  ```yaml
  enforceable: true
  ```

#### `enforcer`
- **Type:** String
- **Purpose:** Tool that enforces this rule
- **Values:** `eslint`, `pylint`, `gofmt`, `prettier`
- **Use Case:** Tooling integration
- **Example:**
  ```yaml
  enforcer: gofmt
  ```

#### `applies_to`
- **Type:** Array
- **Purpose:** File patterns rule applies to
- **Use Case:** Selective application
- **Example:**
  ```yaml
  applies_to:
    - "*.go"
    - "src/**/*.js"
  ```

#### `excludes`
- **Type:** Array
- **Purpose:** File patterns to exclude
- **Use Case:** Exceptions
- **Example:**
  ```yaml
  excludes:
    - "vendor/**"
    - "*.test.go"
  ```

#### `dependencies`
- **Type:** Array
- **Purpose:** Other rules this depends on
- **Use Case:** Rule ordering, validation
- **Example:**
  ```yaml
  dependencies:
    - base-coding-standards
    - go-error-handling
  ```

#### `conflicts_with`
- **Type:** Array
- **Purpose:** Rules that conflict with this one
- **Use Case:** Validation, warnings
- **Example:**
  ```yaml
  conflicts_with:
    - alternative-error-handling
  ```

#### `framework`
- **Type:** String or Array
- **Purpose:** Framework(s) rule applies to
- **Values:** `react`, `vue`, `angular`, `django`, `flask`, `express`
- **Use Case:** Framework-specific rules
- **Example:**
  ```yaml
  framework: react
  ```

#### `library`
- **Type:** String or Array
- **Purpose:** Library/package rule applies to
- **Use Case:** Library-specific guidelines
- **Example:**
  ```yaml
  library: pandas
  ```

### AI/ML Specific Options

These options are specific to AI coding agents and ML workflows.

#### `model_parameters`
- **Type:** Object
- **Purpose:** Model-specific parameters
- **Use Case:** Fine-tune AI behavior
- **Example:**
  ```yaml
  model_parameters:
    temperature: 0.7
    top_p: 0.9
    max_tokens: 2000
    frequency_penalty: 0.0
    presence_penalty: 0.0
  ```

#### `prompt_template`
- **Type:** String
- **Purpose:** Template for prompt generation
- **Use Case:** Custom prompt formatting
- **Example:**
  ```yaml
  prompt_template: "As a ${role}, ${task}"
  ```

#### `system_prompt`
- **Type:** String
- **Purpose:** System-level instructions for AI
- **Use Case:** Behavior modification
- **Example:**
  ```yaml
  system_prompt: You are an expert Go developer specializing in microservices.
  ```

#### `response_format`
- **Type:** String
- **Purpose:** Expected response format
- **Values:** `markdown`, `json`, `xml`, `yaml`, `code`, `diff`
- **Use Case:** Response validation
- **Example:**
  ```yaml
  response_format: json
  ```

#### `validation_rules`
- **Type:** Array or Object
- **Purpose:** Rules to validate AI output
- **Use Case:** Quality control
- **Example:**
  ```yaml
  validation_rules:
    - must_include_tests: true
    - must_include_docs: true
    - max_lines: 500
  ```

#### `tools`
- **Type:** Array
- **Purpose:** Tools AI can use
- **Use Case:** Tool access control
- **Example:**
  ```yaml
  tools:
    - file_search
    - code_interpreter
    - web_browser
  ```

#### `context_window`
- **Type:** Integer
- **Purpose:** Maximum context window size
- **Use Case:** Token management
- **Example:**
  ```yaml
  context_window: 128000
  ```

#### `few_shot_examples`
- **Type:** Array or String (file path)
- **Purpose:** Examples for few-shot learning
- **Use Case:** Improve AI performance
- **Example:**
  ```yaml
  few_shot_examples: examples/go-refactoring.md
  ```

#### `chain_of_thought`
- **Type:** Boolean
- **Purpose:** Enable chain-of-thought reasoning
- **Use Case:** Complex problem solving
- **Example:**
  ```yaml
  chain_of_thought: true
  ```

#### `reflection`
- **Type:** Boolean
- **Purpose:** Enable self-reflection after generation
- **Use Case:** Quality improvement
- **Example:**
  ```yaml
  reflection: true
  ```

### Workflow Options

#### `workflow`
- **Type:** String
- **Purpose:** Workflow this file belongs to
- **Use Case:** Workflow organization
- **Example:**
  ```yaml
  workflow: feature-implementation
  ```

#### `stage`
- **Type:** String
- **Purpose:** Stage in workflow
- **Values:** `planning`, `implementation`, `testing`, `review`, `deployment`
- **Example:**
  ```yaml
  stage: implementation
  ```

#### `next_step`
- **Type:** String
- **Purpose:** Next task in workflow
- **Use Case:** Workflow automation
- **Example:**
  ```yaml
  next_step: write-tests
  ```

#### `on_success`
- **Type:** String or Object
- **Purpose:** Action to take on success
- **Use Case:** Workflow automation
- **Example:**
  ```yaml
  on_success: deploy-to-staging
  ```

#### `on_failure`
- **Type:** String or Object
- **Purpose:** Action to take on failure
- **Use Case:** Error handling
- **Example:**
  ```yaml
  on_failure: notify-team
  ```

#### `parallel_tasks`
- **Type:** Array
- **Purpose:** Tasks that can run in parallel
- **Use Case:** Workflow optimization
- **Example:**
  ```yaml
  parallel_tasks:
    - run-tests
    - run-linter
    - run-security-scan
  ```

### Environment Options

#### `environment`
- **Type:** String or Array
- **Purpose:** Target environment(s)
- **Values:** `development`, `staging`, `production`, `test`
- **Use Case:** Environment-specific rules/tasks
- **Example:**
  ```yaml
  environment: production
  ```

#### `platform`
- **Type:** String or Array
- **Purpose:** Target platform(s)
- **Values:** `linux`, `windows`, `macos`, `docker`, `kubernetes`
- **Use Case:** Platform-specific content
- **Example:**
  ```yaml
  platform: linux
  ```

#### `arch`
- **Type:** String or Array
- **Purpose:** Target architecture(s)
- **Values:** `amd64`, `arm64`, `x86`
- **Use Case:** Architecture-specific content
- **Example:**
  ```yaml
  arch: amd64
  ```

#### `min_version`
- **Type:** String
- **Purpose:** Minimum tool/language version required
- **Use Case:** Compatibility checking
- **Example:**
  ```yaml
  min_version: "go1.21"
  ```

#### `max_version`
- **Type:** String
- **Purpose:** Maximum tool/language version supported
- **Use Case:** Compatibility checking
- **Example:**
  ```yaml
  max_version: "go1.23"
  ```

### Access Control Options

#### `visibility`
- **Type:** String
- **Purpose:** Visibility level
- **Values:** `public`, `private`, `internal`, `team`
- **Use Case:** Access control
- **Example:**
  ```yaml
  visibility: internal
  ```

#### `teams`
- **Type:** Array
- **Purpose:** Teams with access
- **Use Case:** Team-specific content
- **Example:**
  ```yaml
  teams:
    - backend-team
    - platform-team
  ```

#### `users`
- **Type:** Array
- **Purpose:** Specific users with access
- **Use Case:** User-specific content
- **Example:**
  ```yaml
  users:
    - jane.doe
    - john.smith
  ```

#### `roles`
- **Type:** Array
- **Purpose:** Required roles
- **Values:** `developer`, `reviewer`, `admin`, `architect`
- **Use Case:** Role-based content
- **Example:**
  ```yaml
  roles:
    - developer
    - reviewer
  ```

### Quality & Testing Options

#### `test_coverage_required`
- **Type:** Float (percentage) or Boolean
- **Purpose:** Minimum test coverage required
- **Use Case:** Quality gates
- **Example:**
  ```yaml
  test_coverage_required: 80.0
  ```

#### `code_review_required`
- **Type:** Boolean
- **Purpose:** Whether code review is mandatory
- **Use Case:** Quality control
- **Example:**
  ```yaml
  code_review_required: true
  ```

#### `reviewers`
- **Type:** Array
- **Purpose:** Required reviewers
- **Use Case:** Review assignment
- **Example:**
  ```yaml
  reviewers:
    - senior-dev-1
    - senior-dev-2
  ```

#### `lint_rules`
- **Type:** Object or String (config file path)
- **Purpose:** Linting rules to apply
- **Use Case:** Code quality
- **Example:**
  ```yaml
  lint_rules: .eslintrc.json
  ```

### Documentation Options

#### `doc_url`
- **Type:** String (URL)
- **Purpose:** Link to related documentation
- **Use Case:** Reference, learning
- **Example:**
  ```yaml
  doc_url: https://docs.example.com/go-standards
  ```

#### `example_url`
- **Type:** String (URL) or Array
- **Purpose:** Link(s) to examples
- **Use Case:** Learning, reference
- **Example:**
  ```yaml
  example_url: https://github.com/example/repo/blob/main/examples
  ```

#### `video_url`
- **Type:** String (URL)
- **Purpose:** Link to tutorial video
- **Use Case:** Learning
- **Example:**
  ```yaml
  video_url: https://youtube.com/watch?v=example
  ```

#### `references`
- **Type:** Array
- **Purpose:** Related resources, articles, or files
- **Use Case:** Context, learning
- **Example:**
  ```yaml
  references:
    - https://go.dev/doc/effective_go
    - docs/coding-standards.md
  ```

### Security Options

#### `security_level`
- **Type:** String
- **Purpose:** Security classification
- **Values:** `public`, `internal`, `confidential`, `restricted`
- **Use Case:** Security control
- **Example:**
  ```yaml
  security_level: confidential
  ```

#### `sensitive_data`
- **Type:** Boolean
- **Purpose:** Whether content handles sensitive data
- **Use Case:** Compliance, warnings
- **Example:**
  ```yaml
  sensitive_data: true
  ```

#### `compliance`
- **Type:** Array
- **Purpose:** Compliance requirements
- **Values:** `gdpr`, `hipaa`, `pci-dss`, `sox`
- **Use Case:** Regulatory compliance
- **Example:**
  ```yaml
  compliance:
    - gdpr
    - hipaa
  ```

#### `security_scan_required`
- **Type:** Boolean
- **Purpose:** Whether security scanning is required
- **Use Case:** Security control
- **Example:**
  ```yaml
  security_scan_required: true
  ```

---

## Category Reference

### By File Type

#### All File Types
- `title`, `description`, `author`, `created`, `updated`, `version`
- `tags`, `categories`, `deprecated`, `deprecation_message`
- `doc_url`, `example_url`, `video_url`, `references`

#### Task Files Only
- `task_name`, `agent`, `languages`, `model`, `single_shot`, `timeout`
- `resume`, `selectors`, `expand`
- `priority`, `complexity`, `estimated_time`, `prerequisites`, `outputs`
- `context_files`, `max_tokens`, `temperature`, `requires_approval`
- `interactive`, `retry_on_failure`, `fallback_task`
- `workflow`, `stage`, `next_step`, `on_success`, `on_failure`

#### Rule Files Only
- `rule_name`, `task_names`, `languages`, `agent`, `mcp_server`, `expand`
- `stage`, `scope`, `severity`, `enforceable`, `enforcer`
- `applies_to`, `excludes`, `dependencies`, `conflicts_with`
- `framework`, `library`

#### Command Files Only
- `expand`, `selectors`

#### Skill Files Only
- `name`, `description`, `license`, `compatibility`, `metadata`, `allowed-tools`

### By Use Case

#### Filtering & Selection
- `agent`, `languages`, `selectors`, `task_names`, `stage`, `scope`
- `framework`, `library`, `environment`, `platform`, `arch`
- `teams`, `users`, `roles`

#### Metadata Only (No Filtering)
- `model`, `single_shot`, `timeout`, `title`, `description`, `author`
- `created`, `updated`, `version`, `tags`, `categories`

#### Workflow Control
- `resume`, `workflow`, `next_step`, `on_success`, `on_failure`
- `parallel_tasks`, `prerequisites`, `fallback_task`

#### Quality Control
- `test_coverage_required`, `code_review_required`, `reviewers`
- `lint_rules`, `security_scan_required`, `validation_rules`

#### AI Behavior
- `model_parameters`, `prompt_template`, `system_prompt`, `response_format`
- `tools`, `context_window`, `few_shot_examples`, `chain_of_thought`, `reflection`
- `temperature`, `max_tokens`

#### Security & Compliance
- `security_level`, `sensitive_data`, `compliance`, `security_scan_required`
- `visibility`, `requires_approval`

---

## Examples

### Comprehensive Task Example

```yaml
---
# Basic Metadata
task_name: implement-auth-feature
title: Implement OAuth2 Authentication
description: Add OAuth2 authentication with Google and GitHub providers

# Author & Version
author: Security Team
created: 2024-01-15
updated: 2024-03-20
version: 1.2.0

# Task Categorization
tags:
  - security
  - authentication
  - oauth2
categories:
  - backend
  - security

# AI Agent Configuration
agent: copilot
model: gpt-4-turbo
temperature: 0.7
max_tokens: 100000

# Filtering
languages:
  - go
selectors:
  stage: implementation
  scope: backend
  priority: high

# Task Execution
single_shot: false
timeout: 2h
estimated_time: 4h
interactive: false

# Quality Requirements
requires_approval: true
approval_criteria:
  - All tests pass
  - Security scan passed
  - Code review completed
test_coverage_required: 85.0
code_review_required: true
reviewers:
  - security-lead
  - backend-lead

# Dependencies & Workflow
prerequisites:
  - setup-oauth-credentials
  - review-security-requirements
workflow: feature-implementation
stage: implementation
next_step: write-integration-tests
on_success: deploy-to-staging
on_failure: notify-security-team

# Environment
environment: development
platform: linux

# Security
security_level: confidential
sensitive_data: true
compliance:
  - gdpr
  - oauth2-spec
security_scan_required: true

# Documentation
doc_url: https://docs.example.com/oauth2-implementation
example_url: https://github.com/example/oauth2-example
references:
  - docs/security/oauth2.md
  - docs/api/authentication.md

# Context Files
context_files:
  - src/auth/*.go
  - src/middleware/auth.go
  - docs/security/*.md

# Output Expectations
outputs:
  - OAuth2 implementation
  - Unit tests
  - Integration tests
  - API documentation
  - Security audit log
---

# Task Content...
```

### Comprehensive Rule Example

```yaml
---
# Basic Metadata
rule_name: go-security-standards
title: Go Security Best Practices
description: Security guidelines for Go applications

# Author & Version
author: Security Team
created: 2024-01-10
version: 2.1.0
updated: 2024-03-15

# Categorization
tags:
  - security
  - go
  - best-practices
categories:
  - security
  - backend

# Filtering
languages:
  - go
stage: implementation
scope: backend
agent: copilot
framework: gin

# Application Scope
task_names:
  - implement-feature
  - fix-security-bug
  - security-audit
applies_to:
  - "*.go"
  - "!*_test.go"
excludes:
  - "vendor/**"
  - "third_party/**"

# Rule Enforcement
severity: error
enforceable: true
enforcer: gosec

# Dependencies
dependencies:
  - base-go-standards
  - error-handling-standards
conflicts_with:
  - legacy-security-rules

# Environment
environment:
  - production
  - staging
platform: linux

# Security
security_level: internal
compliance:
  - owasp-top-10
  - cwe-top-25

# Documentation
doc_url: https://docs.example.com/go-security
example_url: https://github.com/example/secure-go-examples
references:
  - https://golang.org/doc/security
  - docs/security/go-security.md

# MCP Server Configuration
mcp_server:
  command: gosec
  args: ["-fmt=json", "."]
  env:
    GOSEC_CONFIG: ".gosec.json"
---

# Rule Content...
```

### Comprehensive Skill Example

```yaml
---
# Required Fields
name: advanced-data-analysis
description: |
  Advanced data analysis and visualization skill. Analyze complex datasets,
  generate statistical insights, create interactive visualizations, and
  produce comprehensive reports. Use for CSV, Excel, JSON, SQL databases,
  and big data formats.

# Skill Metadata
license: MIT
compatibility: |
  Requires Python 3.8+, pandas 2.0+, matplotlib 3.5+, seaborn 0.12+,
  plotly 5.0+, scipy 1.9+, scikit-learn 1.2+

metadata:
  author: Data Science Team
  version: "2.3.0"
  category: data-analysis
  subcategory: advanced-analytics
  created: "2024-01-01"
  updated: "2024-03-15"
  documentation: https://docs.example.com/skills/data-analysis
  examples: https://github.com/example/skill-examples/data-analysis
  
# Tool Access
allowed-tools: pandas numpy matplotlib seaborn plotly scipy sklearn jupyter

# Categorization
tags:
  - data-science
  - statistics
  - visualization
  - machine-learning

# Filtering
languages:
  - python
stage: implementation
scope: data-analysis

# Environment
environment: development
platform:
  - linux
  - macos
min_version: "python3.8"

# Security
security_level: internal
sensitive_data: true
compliance:
  - data-privacy-policy

# Documentation
doc_url: https://docs.example.com/skills/advanced-data-analysis
example_url: https://github.com/example/skills/data-analysis/examples
video_url: https://youtube.com/watch?v=data-analysis-tutorial
references:
  - https://pandas.pydata.org/docs/
  - https://matplotlib.org/stable/tutorials/
  - docs/data-analysis-guide.md
---

# Skill Content...
```

---

## Notes on Implementation

### Current Implementation
- The `BaseFrontMatter` struct captures all frontmatter fields in a `Content` map
- Type-specific structs (Task, Rule, Command, Skill) define typed fields
- Custom `UnmarshalJSON` methods populate both typed fields and the Content map
- This allows forward compatibility with new fields

### Adding New Options
To add new frontmatter options:

1. **For typed access**: Add field to appropriate struct in `pkg/codingcontext/markdown/frontmatter.go`
2. **For generic access**: Fields are automatically captured in `Content` map
3. **For filtering**: Add to selector logic in `pkg/codingcontext/selectors/`
4. **For documentation**: Update this file and relevant reference docs

### Backward Compatibility
- All new fields should be optional
- Use `omitempty` in YAML/JSON tags
- Provide sensible defaults
- Don't break existing files without frontmatter

### Best Practices
- Use lowercase for field names
- Use arrays for OR logic (e.g., `languages: [go, python]`)
- Use objects for complex structures (e.g., `mcp_server`)
- Document purpose and example for each field
- Consider whether field should filter or be metadata-only
