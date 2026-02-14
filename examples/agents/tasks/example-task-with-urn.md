---
id: task-urn-example
name: Example Task with URN
description: This task demonstrates the use of URN (Uniform Resource Name) for unique identification
urn: urn:example:task-implement-feature-001
agent: copilot
languages:
  - go
---

# Example Task with URN

This task demonstrates the URN frontmatter field for tasks.

## URN Field

The `urn` field provides a globally unique identifier for this task:
- Format: `urn:<namespace-id>:<namespace-specific-string>`
- Example: `urn:example:task-implement-feature-001`
- Optional field - tasks without URNs work normally

## Use Cases

URNs can help with:
1. **Cross-system tracking**: Reference the same task across different tools
2. **Audit trails**: Track which tasks were used in which coding sessions
3. **Documentation**: Link tasks to external documentation or issue trackers
4. **Analytics**: Aggregate metrics about task usage across projects

## URN Validation

URNs are validated according to RFC 2141:
- Valid: `urn:example:my-task-123`
- Valid: `urn:myorg:tasks:feature-x`
- Invalid: `not-a-urn` (will cause parsing error)
- Invalid: `urn:` (missing namespace-specific string)

When URNs are present, they are logged to stderr when tasks are loaded.
