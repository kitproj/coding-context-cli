---
task_name: file-reference-example
description: Example task demonstrating file reference feature
---

# File Reference Example

This task demonstrates how to include file contents in your prompts using the `@path` syntax.

## Basic Usage

You can reference files using `@filepath` syntax. The file content will be automatically included in the prompt.

For example, to review a Go file:

Review the implementation in @pkg/codingcontext/file_reference.go for code quality and potential improvements.

## Multiple File References

You can reference multiple files in the same task:

Compare the implementations in @pkg/codingcontext/file_reference.go and @pkg/codingcontext/file_reference_test.go to ensure comprehensive test coverage.

## Paths with Spaces

If your file path contains spaces, escape them with a backslash:

Review the component in @src/components/My\ Component.tsx for issues.

## How It Works

- File references start with `@` and end at the first unescaped space or end of line
- File paths are resolved relative to the working directory (where you run the command)
- The file content is wrapped in a code block with the filename as a header
- If a file cannot be read, the `@path` reference remains in the output

## File Reference Patterns

Valid patterns:
- `@filename.ext` - Simple filename
- `@./relative/path/file.txt` - Relative path starting with ./
- `@../parent/file.txt` - Parent directory reference
- `@src/components/Button.tsx` - Path with directory separator
- `@src/My\ File.tsx` - Path with escaped spaces

## Email Addresses

Email addresses like user@example.com are not treated as file references.
