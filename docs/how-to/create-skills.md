---
layout: default
title: Create Skills
parent: How-To Guides
nav_order: 4
---

# How to Create Skills

Skills provide specialized capabilities with progressive disclosure. This guide shows you how to create and organize skills for your AI agents.

## What are Skills?

Skills are modular, specialized capabilities that:
- Provide domain-specific knowledge and workflows
- Use progressive disclosure to minimize token usage
- Can be loaded on-demand by AI agents
- Are organized in subdirectories for easy management

## Basic Skill Structure

Each skill is a subdirectory in `.agents/skills/` containing a `SKILL.md` file:

```
.agents/skills/
├── data-analysis/
│   └── SKILL.md
├── pdf-processing/
│   └── SKILL.md
└── api-testing/
    └── SKILL.md
```

## Creating Your First Skill

### Step 1: Create the Skill Directory

```bash
mkdir -p .agents/skills/my-skill
```

### Step 2: Create the SKILL.md File

Create `.agents/skills/my-skill/SKILL.md` with frontmatter and content:

```markdown
---
name: my-skill
description: Brief description of what this skill does and when to use it. Keep it concise - this is shown in the initial context.
---

# My Skill

## When to use this skill
Use this skill when the user needs to:
- First use case
- Second use case
- Third use case

## How to accomplish tasks
1. Step-by-step instructions
2. Code examples
3. Best practices

## Examples
Provide practical examples here.
```

### Step 3: Test the Skill

Run a task to see your skill in the output:

```bash
coding-context my-task
```

Look for the skills section in the output:
```xml
<available_skills>
  <skill>
    <name>my-skill</name>
    <description>Brief description...</description>
    <location>/path/to/.agents/skills/my-skill/SKILL.md</location>
  </skill>
</available_skills>
```

## Frontmatter Fields

### Required Fields

#### `name`
The skill identifier. Must be 1-64 characters.

```yaml
---
name: data-analysis
---
```

#### `description`
Explains what the skill does and when to use it. Must be 1-1024 characters. This is shown in the initial context.

```yaml
---
description: Analyze datasets, generate charts, and create summary reports. Use when working with CSV, Excel, or other tabular data.
---
```

### Optional Fields

#### `license`
The license applied to the skill.

```yaml
---
license: MIT
---
```

#### `compatibility`
Environment requirements. Max 500 characters.

```yaml
---
compatibility: Requires Python 3.8+ with pandas and matplotlib installed
---
```

#### `metadata`
Arbitrary key-value pairs for additional information.

```yaml
---
metadata:
  author: data-team
  version: "2.1"
  category: analytics
  tags: data,visualization,reporting
---
```

## Complete Example

Here's a complete skill for PDF processing:

**.agents/skills/pdf-processing/SKILL.md:**
```markdown
---
name: pdf-processing
description: Extract text and tables from PDF files, fill PDF forms, and merge multiple PDFs. Use when working with PDF documents or when the user mentions PDFs, forms, or document extraction.
license: Apache-2.0
metadata:
  author: document-team
  version: "1.0"
  category: document-processing
---

# PDF Processing

## When to use this skill
Use this skill when the user needs to:
- Extract text or tables from PDF documents
- Fill out PDF forms programmatically
- Merge multiple PDF files into one
- Split PDF files into separate documents

## How to extract text
1. Use pdfplumber for text extraction:
   ```python
   import pdfplumber
   with pdfplumber.open('document.pdf') as pdf:
       text = pdf.pages[0].extract_text()
       print(text)
   ```

2. For tables, use:
   ```python
   with pdfplumber.open('document.pdf') as pdf:
       tables = pdf.pages[0].extract_tables()
   ```

## How to fill forms
1. Use PyPDF2 to fill form fields:
   ```python
   from PyPDF2 import PdfReader, PdfWriter
   
   reader = PdfReader('form.pdf')
   writer = PdfWriter()
   
   # Update form fields
   writer.add_page(reader.pages[0])
   writer.update_page_form_field_values(
       writer.pages[0],
       {"field_name": "value"}
   )
   
   with open('filled_form.pdf', 'wb') as output:
       writer.write(output)
   ```

## How to merge documents
```python
from PyPDF2 import PdfMerger

merger = PdfMerger()
merger.append('document1.pdf')
merger.append('document2.pdf')
merger.write('merged.pdf')
merger.close()
```

## Best Practices
- Always close PDF files after processing
- Handle exceptions for corrupt or password-protected PDFs
- Use pdfplumber for text extraction (better accuracy)
- Use PyPDF2 for form manipulation and merging
```

## Organizing Skills

### Group Related Skills

```
.agents/skills/
├── web-scraping/
│   └── SKILL.md
├── api-testing/
│   └── SKILL.md
└── data/
    ├── data-analysis/
    │   └── SKILL.md
    └── data-visualization/
        └── SKILL.md
```

### Include Supporting Files

Skills can include additional files:

```
.agents/skills/pdf-processing/
├── SKILL.md
├── examples/
│   ├── example1.py
│   └── example2.py
└── references/
    └── REFERENCE.md
```

Reference them in your SKILL.md:
```markdown
For more details, see [the reference guide](references/REFERENCE.md).
```

## Using Selectors with Skills

Skills can be filtered using selectors, just like rules:

```yaml
---
name: go-testing
description: Write and run Go tests with best practices
languages:
  - go
stage: testing
---
```

Filter skills by selector:
```bash
# Only include Go testing skills
coding-context -s languages=go -s stage=testing implement-feature
```

## Progressive Disclosure Pattern

Skills use progressive disclosure to save tokens:

1. **Initial Context**: Only metadata (name, description, location) is included
2. **On-Demand Loading**: AI agents can read the full SKILL.md file when needed

### How It Works

When you run a task, the output includes:
```xml
<available_skills>
  <skill>
    <name>pdf-processing</name>
    <description>Extract text and tables from PDF files...</description>
    <location>/absolute/path/to/.agents/skills/pdf-processing/SKILL.md</location>
  </skill>
</available_skills>
```

The AI agent can:
1. See available skills in the context
2. Decide which skill is relevant
3. Read the full SKILL.md file from the location when needed

## Best Practices

### Writing Good Descriptions

**Good description (clear, actionable):**
```yaml
description: Analyze datasets, generate charts, and create reports. Use when working with CSV, Excel, or tabular data for analysis or visualization.
```

**Poor description (vague, unhelpful):**
```yaml
description: Data stuff. Use for data.
```

### Organizing Content

Structure your skill content clearly:

```markdown
# Skill Name

## When to use this skill
- Clear use cases

## How to accomplish tasks
- Step-by-step instructions
- Code examples
- Configuration details

## Examples
- Practical examples
- Common scenarios

## Best Practices
- Tips and recommendations
- Common pitfalls to avoid

## Troubleshooting
- Common issues and solutions
```

### Naming Conventions

- Use lowercase with hyphens: `data-analysis`, `pdf-processing`
- Be descriptive: `api-testing` not `api`
- Match the directory name to the skill name

## Testing Skills

### Verify Skill Discovery

```bash
# Run any task and check for skills in output
coding-context my-task 2>&1 | grep -A 10 "available_skills"
```

### Test with Selectors

```bash
# Verify selector filtering works
coding-context -s languages=python my-task
```

### Check Skill Metadata

Ensure your skill has:
- Valid YAML frontmatter
- Required fields (name, description)
- Description length: 1-1024 characters
- Name length: 1-64 characters

## Common Issues

### Skill Not Appearing

**Check:**
1. Directory structure: `.agents/skills/skill-name/SKILL.md`
2. Frontmatter is valid YAML
3. Required fields (name, description) are present
4. File is named exactly `SKILL.md` (case-sensitive)

### Validation Errors

**Error: "skill missing required 'name' field"**
- Add `name:` field to frontmatter

**Error: "skill 'name' field must be 1-64 characters"**
- Shorten the skill name

**Error: "skill missing required 'description' field"**
- Add `description:` field to frontmatter

**Error: "skill 'description' field must be 1-1024 characters"**
- Adjust description length

## See Also

- [File Formats Reference](../reference/file-formats) - Detailed skill file specification
- [Search Paths Reference](../reference/search-paths) - Where skills are discovered
- [How to Use Selectors](./use-selectors) - Filtering skills with selectors
