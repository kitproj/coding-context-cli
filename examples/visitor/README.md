# Visitor Pattern Example

This example demonstrates how to use the visitor pattern to customize rule processing.

## What This Example Shows

The example creates a custom `LoggingVisitor` that:
- Logs detailed information about each rule as it's processed
- Tracks the total number of rules processed
- Maintains the default behavior of writing content to stdout

## Running the Example

```bash
go run main.go
```

## Prerequisites

This example expects to find a task file named `fix-bug.md` in one of the standard search paths.

You can create a simple task file for testing:

```bash
mkdir -p .agents/tasks
cat > .agents/tasks/fix-bug.md << 'TASK_EOF'
# Fix Bug Task

Please analyze and fix the bug.
TASK_EOF
```

## Use Cases for Custom Visitors

Custom visitors enable many advanced scenarios:

1. **Logging and Monitoring**: Track which rules are being used
2. **Filtering**: Skip certain rules based on custom logic
3. **Transformation**: Modify rule content before output
4. **Analytics**: Collect statistics about rule usage
5. **Custom Output**: Write to multiple destinations or formats
6. **Caching**: Store processed rules for reuse
7. **Validation**: Verify rule content meets requirements

## Example Output

When running with the logging visitor, you'll see output like:

```
=== Rule #1 ===
Path: /path/to/.agents/rules/setup.md
Tokens: 42
Frontmatter:
  language: go

# Setup Guide
...

=== Summary ===
Total rules processed: 3
```
