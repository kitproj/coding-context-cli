# Agent Skills

Agent Skills are a lightweight format for extending AI agent capabilities with specialized knowledge and workflows. This document describes how skills work in the coding-context CLI.

## Overview

Skills are folders containing a `SKILL.md` file with metadata and instructions that tell an agent how to perform specific tasks. Skills can bundle scripts, templates, and reference materials alongside the core instructions.

## Directory Structure

Skills must be placed in the `.agents/skills` directory. Each skill is a subdirectory containing at minimum a `SKILL.md` file:

```
.agents/skills/
├── pdf-processing/
│   └── SKILL.md          # Required
├── data-analysis/
│   ├── SKILL.md          # Required
│   ├── scripts/          # Optional
│   ├── references/       # Optional
│   └── assets/           # Optional
```

## SKILL.md Format

The `SKILL.md` file must contain YAML frontmatter followed by Markdown content.

### Required Fields

```yaml
---
name: pdf-processing
description: Extract text and tables from PDF files, fill PDF forms, and merge multiple PDFs. Use when working with PDF documents or when the user mentions PDFs, forms, or document extraction.
---
```

- **name**: A short identifier (1-64 characters)
  - Must match the parent directory name
  - Only lowercase letters, numbers, and hyphens
  - Cannot start or end with a hyphen
  - No consecutive hyphens
  
- **description**: What the skill does and when to use it (1-1024 characters)
  - Should describe both functionality and use cases
  - Include specific keywords to help agents identify relevant tasks

### Optional Fields

```yaml
---
name: pdf-processing
description: Extract text and tables from PDF files...
license: MIT
compatibility: Requires Python 3.8+
metadata:
  author: example-org
  version: "1.0"
allowed-tools: bash pip python
---
```

- **license**: License name or reference to a bundled license file
- **compatibility**: Environment requirements (max 500 characters)
- **metadata**: Additional key-value properties
- **allowed-tools**: Space-delimited list of pre-approved tools (experimental)

### Body Content

The Markdown body after the frontmatter contains the skill instructions. There are no format restrictions.

```markdown
# PDF Processing Skill

## When to use this skill

Use this skill when the user needs to:
- Extract text from PDF files
- Fill PDF forms programmatically
- Merge multiple PDF files

## How to extract text

1. Install the required library:
   ```bash
   pip install pdfplumber
   ```

2. Use the following code:
   ```python
   import pdfplumber
   # ...
   ```
```

## Discovery and Loading

Skills are discovered at startup and their metadata (name and description) is loaded into the context. This allows the CLI to know which skills are available without loading full instructions.

### Skill Discovery Locations

Skills are searched in the following paths:
- `.agents/skills` in the current directory
- `.agents/skills` in parent directories
- `.agents/skills` in the user's home directory

### Selector Filtering

Skills support the same selector filtering as rules. You can add custom frontmatter fields and use the `-s` flag to filter which skills are included:

```yaml
---
name: python-skill
description: Python programming utilities
language: python
---
```

Run with: `coding-context -s language=python task-name`

## Resume Mode

Like rules, skills are skipped when running in resume mode (`-r` flag).

## Validation

The CLI validates skills during discovery:

1. **File name**: Must be exactly `SKILL.md`
2. **Required fields**: Both `name` and `description` must be present
3. **Name format**: Must follow the naming rules (1-64 chars, lowercase, etc.)
4. **Name match**: Skill name must match the parent directory name
5. **Description length**: Must be 1-1024 characters
6. **Compatibility length**: If provided, must be ≤ 500 characters

If validation fails, the CLI will report an error and exit.

## Example

Here's a complete example skill:

**`.agents/skills/pdf-processing/SKILL.md`**:

```yaml
---
name: pdf-processing
description: Extract text and tables from PDF files, fill PDF forms, and merge multiple PDFs. Use when working with PDF documents.
license: MIT
metadata:
  author: example-org
  version: "1.0"
---

# PDF Processing Skill

## When to use this skill

Use this skill when the user needs to work with PDF files.

## How to extract text

Use pdfplumber for text extraction:

\```python
import pdfplumber

with pdfplumber.open("document.pdf") as pdf:
    for page in pdf.pages:
        text = page.extract_text()
        print(text)
\```

## Best practices

- Always close file handles properly
- Handle encrypted PDFs with appropriate passwords
- Validate PDF structure before processing
```

## Integration

Skills are included in the `Result` struct returned by the CLI and can be accessed programmatically or via the API. The skills are loaded with their full content and token estimates.

## See Also

- [File Formats](file-formats.md) - Other file format specifications
- [Search Paths](search-paths.md) - How the CLI discovers files
- [CLI Reference](cli.md) - Command-line options
