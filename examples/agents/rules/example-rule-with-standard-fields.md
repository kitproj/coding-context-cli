---
id: rule-example-001
name: Example Rule with Standard Fields
description: This rule demonstrates the use of id, name, and description standard frontmatter fields
languages:
  - go
stage: implementation
---

# Example Rule

This rule demonstrates the new standard frontmatter fields:
- `id`: A unique identifier for the rule
- `name`: A human-readable name  
- `description`: A description of what the rule provides

These fields are metadata only and do not affect rule filtering.
Rules are still filtered using the `languages` and `stage` fields.
