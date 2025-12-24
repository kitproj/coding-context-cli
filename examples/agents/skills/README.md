# Agent Skills Examples

This directory contains example skills that demonstrate the Agent Skills format.

## What are Agent Skills?

Agent Skills are a lightweight, open format for extending AI agent capabilities with specialized knowledge and workflows. Each skill is a folder containing a `SKILL.md` file with metadata and instructions.

## Skills in this directory

### pdf-processing

Extract text and tables from PDF files, fill PDF forms, and merge multiple PDFs.

```bash
coding-context -C examples/test-workspace your-task
```

### data-analysis

Analyze datasets, generate charts, and create summary reports for CSV, Excel, and other tabular data formats.

## Skill Structure

Each skill folder contains:

```
skill-name/
├── SKILL.md          # Required: instructions + metadata
├── scripts/          # Optional: executable code
├── references/       # Optional: documentation
└── assets/           # Optional: templates, resources
```

## Required Frontmatter

The `SKILL.md` file must contain YAML frontmatter with at least:

- `name`: A unique identifier (1-64 characters, lowercase alphanumeric and hyphens only)
- `description`: What the skill does and when to use it (1-1024 characters)

Example:

```yaml
---
name: pdf-processing
description: Extract text and tables from PDF files, fill forms, merge documents.
---
```

## Optional Frontmatter

- `license`: License name or reference
- `compatibility`: Environment requirements
- `metadata`: Additional key-value pairs
- `allowed-tools`: Pre-approved tools (experimental)

## Progressive Disclosure

Skills use progressive disclosure for efficient context management:

1. **Discovery**: At startup, only the `name` and `description` are loaded
2. **Activation**: When relevant, the agent loads the full `SKILL.md` content
3. **Execution**: Scripts and resources are loaded as needed

## Selectors

Skills can be filtered using selectors in their frontmatter, just like rules:

```yaml
---
name: dev-skill
description: A skill for development environments
env: development
---
```

Then use with:

```bash
coding-context -s env=development your-task
```

Skills without the selector key are included by default (OR logic).

## Testing

To see skills in action:

```bash
# Navigate to the test workspace
cd examples/test-workspace

# Run a simple task and see discovered skills
coding-context simple
```

The output will include an `<available_skills>` section with skill metadata.
