---
task_name: file-reference-example
description: Example task demonstrating file reference feature
---

# File Reference Example

This task demonstrates how to include file contents in your prompts using the `@` syntax.

## Basic Usage

You can reference files using `@filepath` syntax. The file content will be automatically included in the prompt.

For example, to review a Go file:

Review the implementation in @pkg/codingcontext/file_reference.go for code quality and potential improvements.

## Multiple File References

You can reference multiple files in the same task:

Compare the implementations in @pkg/codingcontext/file_reference.go and @pkg/codingcontext/file_reference_test.go to ensure comprehensive test coverage.

## How It Works

- File references are expanded after parameter substitution
- File paths are resolved relative to the working directory (where you run the command)
- The file content is wrapped in a code block with the filename as a header
- Email addresses (like user@example.com) are not expanded
- Only file paths with extensions or path separators are recognized

## File Reference Patterns

Valid patterns:
- `@filename.ext` - Simple filename with extension
- `@./relative/path/file.txt` - Relative path starting with ./
- `@../parent/file.txt` - Parent directory reference
- `@src/components/Button.tsx` - Path with directory separator

Invalid patterns (not expanded):
- `@username` - Twitter handle (no extension or separator)
- `user@example.com` - Email address
