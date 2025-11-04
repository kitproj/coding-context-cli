# Basic Library Usage Example

This example demonstrates how to use the coding-context-cli as a library.

## Running the Example

```bash
go run main.go
```

Note: This example expects to find a task file named `fix-bug.md` in one of the standard search paths:
- `./.agents/tasks/fix-bug.md`
- `~/.agents/tasks/fix-bug.md`
- `/etc/agents/tasks/fix-bug.md`

The example shows how to:
- Create parameters for substitution
- Create selectors for filtering rules
- Configure and create an Assembler
- Execute the context assembly
