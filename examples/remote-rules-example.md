# Remote Rules Example

This example demonstrates how to use remote rule files with coding-context-cli.

## Use Case

You have coding standards hosted on a central server or CDN that you want to share across multiple projects. Instead of duplicating these files in each repository, you can load them from a remote URL.

## Example Setup

### 1. Host Your Rules

Host your rule files on any HTTP/HTTPS server. For example:

- GitHub Raw Files: `https://raw.githubusercontent.com/org/repo/main/rules/standards.md`
- GitHub Pages: `https://org.github.io/repo/rules/standards.md`
- CDN: `https://cdn.example.com/rules/standards.md`
- Internal Server: `https://internal.company.com/rules/standards.md`

### 2. Create a Remote Rule File

**Example: `coding-standards.md`**

```markdown
---
language: Go
---
# Go Coding Standards

- Use `gofmt` for formatting
- Write tests for all public functions
- Use meaningful variable names
- Add comments for exported functions
```

### 3. Use the Remote Rule

```bash
# Single remote rule
coding-context-cli \
  -remote-rule https://example.com/rules/coding-standards.md \
  fix-bug

# Multiple remote rules
coding-context-cli \
  -remote-rule https://example.com/rules/coding-standards.md \
  -remote-rule https://example.com/rules/security-guidelines.md \
  -remote-rule https://example.com/rules/performance-best-practices.md \
  implement-feature

# Mix local and remote rules
coding-context-cli \
  -remote-rule https://example.com/shared/org-standards.md \
  -s language=Go \
  refactor-code
```

## Real-World Example: GitHub Raw Files

You can use GitHub to host your shared rules:

```bash
# Load organization-wide coding standards from GitHub
coding-context-cli \
  -remote-rule https://raw.githubusercontent.com/company/shared-rules/main/coding-standards.md \
  -remote-rule https://raw.githubusercontent.com/company/shared-rules/main/security.md \
  -p component=auth \
  fix-security-issue | llm -m claude-3-sonnet
```

## Benefits

1. **Centralized Management**: Update rules in one place, all projects get the latest version
2. **Version Control**: Use git tags/branches to manage different versions of rules
3. **No Duplication**: Don't need to copy rules into every repository
4. **Easy Distribution**: Share rules across teams and organizations
5. **Mix and Match**: Combine remote rules with local project-specific rules

## Important Notes

- Remote files are fetched on each invocation (no caching)
- Bootstrap scripts are NOT supported for remote files
- Missing remote files are silently skipped
- Works with any HTTP/HTTPS endpoint
- Respects standard HTTP status codes (404 = not found, etc.)

## Troubleshooting

### Remote file not loading

Check:
1. URL is accessible (test with `curl`)
2. File is served with correct Content-Type
3. No authentication required (basic auth not currently supported)
4. URL is direct to the file (not a directory listing)

### Performance concerns

- Remote files are fetched on every run
- Consider caching layer if you have many remote rules
- Use a CDN for better performance across locations

## Advanced: Dynamic Rules

You can even use dynamic endpoints that generate rules on-the-fly:

```bash
# Load rules from an API endpoint
coding-context-cli \
  -remote-rule https://api.company.com/rules/current/coding-standards \
  implement-feature
```

This allows for:
- Conditional rules based on project type
- Time-based rules (e.g., different standards during migration periods)
- Team-specific rules
- Environment-specific guidelines
