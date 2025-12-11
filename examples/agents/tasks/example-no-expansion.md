---
task_name: example-no-expansion
expand: false
---

# Example Task Without Parameter Expansion

This task demonstrates how to disable parameter expansion using the `expand: false` frontmatter field.

## Usage

When `expand` is set to `false`, parameter placeholders are preserved as-is:

- Issue Number: ${issue_number}
- Issue Title: ${issue_title}
- Repository: ${repo}
- Branch: ${branch}

This is useful when:
1. You want to pass parameters directly to an AI agent that will handle its own substitution
2. You're using template syntax that conflicts with the parameter expansion syntax
3. You want to preserve the template for later processing

## Example

```bash
coding-context -p issue_number=123 -p issue_title="Bug Fix" example-no-expansion
```

Even though parameters are passed on the command line, they won't be expanded in the output.
The placeholders `${issue_number}` and `${issue_title}` will remain unchanged.
