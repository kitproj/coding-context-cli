# Shell Output Demo

This example demonstrates the **shell-output** feature, which allows you to execute shell scripts and include their output in the context assembled for AI agents.

## What is Shell Output?

The shell-output feature enables you to:
- Execute shell scripts that gather dynamic runtime information
- Include the stdout of these scripts in the AI context
- Filter scripts using frontmatter selectors (similar to rules)
- Provide real-time system state, git status, test results, etc. to AI agents

## Directory Structure

```
.opencode/shell-output/          # Shell scripts to execute
  git-status                     # Script that outputs git status
  git-status.md                  # Optional metadata with frontmatter
  system-info                    # Script that outputs system info
.agents/tasks/
  demo-task.md                   # Task that uses shell outputs
```

## How It Works

1. Place executable scripts in `.opencode/shell-output/`
2. Optionally create `<script-name>.md` files with frontmatter for filtering
3. When the CLI runs, it executes these scripts and captures their stdout
4. The output is included in the context between rules and the task

## Running the Example

From this directory:

```bash
cd /path/to/coding-context-cli/examples/shell-output-demo
coding-context demo-shell-output
```

Or from the repository root:

```bash
coding-context -C examples/shell-output-demo demo-shell-output
```

## Expected Output

The output will include:
1. Git repository status (from `git-status` script)
2. System information (from `system-info` script)
3. Task content (from `demo-task.md`)

## Using with Selectors

You can filter shell outputs using selectors, just like rules:

**git-status.md:**
```yaml
---
purpose: git-context
---
```

Then use selector to include only git-related scripts:
```bash
coding-context -s purpose=git-context demo-shell-output
```

## Resume Mode

Shell outputs are automatically skipped in resume mode (like rules):

```bash
coding-context -r demo-shell-output  # Skips shell outputs
```

## Use Cases

- **Git context**: Include current branch, recent commits, file changes
- **Test results**: Run and include test suite output
- **System state**: CPU, memory, disk usage for debugging
- **Environment info**: Installed tools, versions, configuration
- **Dynamic data**: Fetch data from APIs, databases, or services
- **Code metrics**: Lines of code, test coverage, complexity metrics

## Notes

- Scripts must be executable (chmod +x)
- stdout is captured and included in context
- stderr goes to the terminal (not included in context)
- Scripts are executed in the order they're discovered
- .md files in shell-output are treated as metadata, not executables
