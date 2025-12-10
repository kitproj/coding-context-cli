---
task_name: file-reference-example
description: Example task demonstrating file reference feature
---

# File Reference Example

This task demonstrates how to include file contents in your prompts using the `${file:path}` syntax.

## Basic Usage

You can reference files using `${file:filepath}` syntax, similar to parameter expansion. The file content will be automatically included in the prompt.

For example, to review a Go file:

Review the implementation in ${file:pkg/codingcontext/file_reference.go} for code quality and potential improvements.

## Multiple File References

You can reference multiple files in the same task:

Compare the implementations in ${file:pkg/codingcontext/file_reference.go} and ${file:pkg/codingcontext/file_reference_test.go} to ensure comprehensive test coverage.

## How It Works

- File references use the same expansion mechanism as parameters
- File references are expanded during parameter substitution using the `file:` prefix
- File paths are resolved relative to the working directory (where you run the command)
- The file content is wrapped in a code block with the filename as a header
- If a file cannot be read, the placeholder remains unexpanded and a warning is logged

## File Reference Patterns

Valid patterns:
- `${file:filename.ext}` - Simple filename
- `${file:./relative/path/file.txt}` - Relative path starting with ./
- `${file:../parent/file.txt}` - Parent directory reference
- `${file:src/components/Button.tsx}` - Path with directory separator

## Combining with Parameters

You can combine file references with regular parameters:

Review ${file:${source_file}} and check if it matches the requirements in ${file:docs/requirements.md}.
