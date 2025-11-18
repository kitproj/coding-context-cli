# How to Use Shell Outputs

Shell outputs allow you to execute scripts and include their output in the context assembled for AI agents. This enables you to provide dynamic, runtime information alongside static rules and tasks.

## Overview

The shell-output feature works by:
1. Finding executable scripts in `.opencode/shell-output/` directories
2. Executing them and capturing their stdout
3. Including the output in the AI context
4. Supporting optional frontmatter-based filtering

## Basic Usage

### Step 1: Create the Directory

Create the shell-output directory in your project:

```bash
mkdir -p .opencode/shell-output
```

### Step 2: Add Executable Scripts

Create executable scripts that output information you want to include:

**Example: `.opencode/shell-output/git-status`**
```bash
#!/bin/bash
echo "## Current Git Status"
echo ""
git status --short
echo ""
echo "Recent commits:"
git log --oneline -5
```

Make it executable:
```bash
chmod +x .opencode/shell-output/git-status
```

### Step 3: Run Your Task

When you run your task, the shell output will be automatically included:

```bash
coding-context my-task
```

The output will contain:
1. All matching rules (if any)
2. **Shell outputs** (newly executed)
3. Your task content

## Advanced Features

### Filtering with Selectors

Add optional metadata files to enable selector-based filtering.

**Example: `.opencode/shell-output/test-results.md`**
```yaml
---
language: Go
type: testing
---
Metadata for the test-results script.
```

**Example: `.opencode/shell-output/test-results`**
```bash
#!/bin/bash
echo "## Go Test Results"
go test ./... -v | tail -20
```

Use selectors to include only specific outputs:
```bash
coding-context -s language=Go my-task
```

### Multiple Scripts

You can have multiple scripts in the shell-output directory. They execute in the order they're discovered:

```
.opencode/shell-output/
  01-system-info
  02-git-status
  03-test-results
  04-code-coverage
```

Tip: Prefix with numbers to control execution order.

### Skip in Resume Mode

Shell outputs are automatically skipped in resume mode (like rules):

```bash
# First run: includes shell outputs
coding-context my-task

# Resume: skips shell outputs
coding-context -r my-task
```

## Common Use Cases

### Git Context

**`.opencode/shell-output/git-context`**
```bash
#!/bin/bash
echo "## Git Repository Context"
echo ""
echo "Branch: $(git rev-parse --abbrev-ref HEAD)"
echo "Last commit: $(git log -1 --oneline)"
echo ""
echo "Modified files:"
git status --short
```

### Test Results

**`.opencode/shell-output/run-tests`**
```bash
#!/bin/bash
echo "## Test Suite Results"
echo ""
npm test 2>&1 | tail -30
```

### System Information

**`.opencode/shell-output/system-info`**
```bash
#!/bin/bash
echo "## Build Environment"
echo ""
echo "OS: $(uname -s)"
echo "Node: $(node --version 2>/dev/null || echo 'not installed')"
echo "Go: $(go version 2>/dev/null | awk '{print $3}' || echo 'not installed')"
```

### Code Metrics

**`.opencode/shell-output/code-stats`**
```bash
#!/bin/bash
echo "## Code Statistics"
echo ""
echo "Lines of code:"
find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
echo ""
echo "Test coverage:"
go test -cover ./... 2>/dev/null | grep coverage
```

### API Response

**`.opencode/shell-output/api-data`**
```bash
#!/bin/bash
echo "## Latest API Data"
echo ""
curl -s https://api.example.com/status | jq .
```

## Remote Directories

Shell outputs work with remote directories too:

```bash
coding-context \
  -d git::https://github.com/company/shared-scripts.git \
  my-task
```

The remote directory can include `.opencode/shell-output/` scripts.

## Best Practices

### 1. Keep Outputs Focused

Only include relevant information. Too much output increases token usage and cost.

```bash
# Good: Concise and relevant
git log --oneline -5

# Avoid: Too verbose
git log --stat -100
```

### 2. Handle Errors Gracefully

Check if commands exist before running them:

```bash
#!/bin/bash
if command -v go &> /dev/null; then
    echo "Go version: $(go version)"
else
    echo "Go not installed"
fi
```

### 3. Use Timeouts

Prevent scripts from hanging:

```bash
#!/bin/bash
timeout 5s npm test || echo "Tests timed out"
```

### 4. Control Output Size

Limit output length to avoid excessive tokens:

```bash
#!/bin/bash
# Get last 20 lines of logs
tail -20 /var/log/app.log
```

### 5. Secure Sensitive Data

Never include secrets or credentials in shell outputs:

```bash
#!/bin/bash
# Good: Safe information
echo "Database host: ${DB_HOST}"

# Bad: Don't do this!
# echo "Database password: ${DB_PASSWORD}"
```

## Troubleshooting

### Script Not Executing

**Problem**: Shell output not appearing in context.

**Solutions**:
1. Check if script is executable: `chmod +x .opencode/shell-output/my-script`
2. Verify script path is correct
3. Check selector filters aren't excluding it

### Script Fails

**Problem**: Script execution fails.

**Solutions**:
1. Test script independently: `./.opencode/shell-output/my-script`
2. Check stderr output in terminal
3. Add error handling in script
4. Verify required commands are installed

### Too Many Tokens

**Problem**: Shell outputs are using too many tokens.

**Solutions**:
1. Limit output length: `head -20`, `tail -10`
2. Use `grep` or `awk` to filter output
3. Summarize data instead of including raw output
4. Use selectors to include only necessary scripts

## Example Project Structure

Complete example with shell outputs, rules, and tasks:

```
my-project/
  .opencode/
    shell-output/
      git-status           # Git context
      git-status.md        # Metadata with selectors
      test-results         # Test output
      system-info          # Environment info
  .agents/
    rules/
      go-standards.md      # Static rules
    tasks/
      fix-bug.md           # Task definition
```

Run with:
```bash
cd my-project
coding-context -s language=Go fix-bug
```

## Learn More

- [Create Tasks](create-tasks.md) - Define task prompts
- [Use Selectors](use-selectors.md) - Filter with frontmatter
- [GitHub Actions Integration](github-actions.md) - Use in CI/CD
- [Shell Output Example](../../examples/shell-output-demo/) - Working demo

## Next Steps

1. Create your first shell output script
2. Test it with a simple task
3. Add metadata for selector support
4. Integrate into your AI workflow
