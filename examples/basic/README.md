# Basic Library Usage Example

This example demonstrates how to use the coding-context-cli as a library.

## Running the Example

```bash
go run main.go
```

## Prerequisites

This example expects to find a task file named `fix-bug.md` in one of the standard search paths:
- `./.agents/tasks/fix-bug.md`
- `~/.agents/tasks/fix-bug.md`
- `/etc/agents/tasks/fix-bug.md`

You can create a simple task file for testing:

```bash
mkdir -p .agents/tasks
cat > .agents/tasks/fix-bug.md << 'EOF'
# Fix Bug Task

Please fix the bug in ${component} related to ${issue}.

Analyze the code and provide a fix.
EOF
```

## What This Example Shows

The example demonstrates how to:
- Create parameters for substitution
- Create selectors for filtering rules
- Configure and create an Assembler
- Execute the context assembly
