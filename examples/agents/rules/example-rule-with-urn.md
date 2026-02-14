---
id: rule-urn-example
name: Example Rule with URN
description: This rule demonstrates the use of URN (Uniform Resource Name) for unique identification
urn: urn:example:rule-go-implementation-001
languages:
  - go
stage: implementation
---

# Example Rule with URN

This rule demonstrates the URN frontmatter field:
- `urn`: An optional Uniform Resource Name that uniquely identifies this rule across systems
- URNs must follow RFC 2141 format: `urn:<namespace-id>:<namespace-specific-string>`
- Examples: `urn:example:rule-123`, `urn:myorg:coding-rules:go-001`

URNs are metadata only and do not affect rule filtering.
They can be used for:
- Tracking rule usage across different projects
- Linking to external documentation or systems
- Maintaining consistent references across tools

When URNs are present, they are logged to stderr when rules are loaded.
