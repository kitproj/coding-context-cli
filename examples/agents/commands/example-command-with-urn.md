---
id: command-urn-example
name: Example Command with URN
description: This command demonstrates the use of URN (Uniform Resource Name) for unique identification
urn: urn:example:command-setup-environment-001
---

# Example Command with URN

This command demonstrates the URN frontmatter field for commands.

Commands can have URNs just like tasks and rules:
- Format: `urn:<namespace-id>:<namespace-specific-string>`
- Example: `urn:example:command-setup-environment-001`
- Optional field

When URNs are present, they are logged to stderr when commands are loaded via slash commands.
