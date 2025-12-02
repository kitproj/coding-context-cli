---
layout: default
title: Use Remote Directories
parent: How-to Guides
nav_order: 6
---

# How to Use Remote Directories

Load rules and tasks from remote locations like Git repositories, HTTP servers, or S3 buckets instead of duplicating files across projects.

## Problem

You want to:
- Share coding standards across multiple projects
- Centralize organizational guidelines
- Version control your rules separately from your code
- Distribute team-specific rules without duplication

## Solution

Use the `-d` flag to load rules and tasks from remote directories. The CLI downloads them to a temporary location, processes them, then cleans up automatically.

## Basic Usage

### Load from Git Repository

```bash
# Clone a Git repository containing rules
coding-context -d git::https://github.com/company/shared-rules.git /fix-bug
```

This downloads the repository, searches for rules and tasks in standard locations (`.agents/rules/`, `.agents/tasks/`, etc.), and includes them in the context.

### Use Specific Version

```bash
# Use a specific tag
coding-context -d 'git::https://github.com/company/rules.git?ref=v1.0.0' /fix-bug

# Use a specific branch
coding-context -d 'git::https://github.com/company/rules.git?ref=main' /fix-bug

# Use a specific commit
coding-context -d 'git::https://github.com/company/rules.git?ref=abc123def' /fix-bug
```

### Use Subdirectory

If your rules are in a subdirectory of a repository, use double slashes (`//`):

```bash
# Get rules from the 'coding-standards' subdirectory
coding-context -d 'git::https://github.com/company/mono-repo.git//coding-standards' /fix-bug
```

## Supported Protocols

The `-d` flag uses [go-getter](https://github.com/hashicorp/go-getter), which supports multiple protocols:

### Git Repositories

```bash
# HTTPS
coding-context -d git::https://github.com/company/rules.git /fix-bug

# SSH (requires SSH keys configured)
coding-context -d git::git@github.com:company/rules.git /fix-bug

# With authentication token
coding-context -d 'git::https://token@github.com/company/private-rules.git' /fix-bug
```

### HTTP/HTTPS

```bash
# Download and extract tar.gz
coding-context -d https://example.com/rules.tar.gz /fix-bug

# Download and extract zip
coding-context -d https://example.com/rules.zip /fix-bug
```

### S3 Buckets

```bash
# S3 bucket
coding-context -d s3::https://s3.amazonaws.com/my-bucket/rules /fix-bug

# With specific region
coding-context -d s3::https://s3-us-west-2.amazonaws.com/my-bucket/rules /fix-bug
```

### Local Files

Useful for testing:

```bash
coding-context -d file:///path/to/local/rules /fix-bug
```

## Advanced Usage

### Multiple Remote Sources

Combine rules from multiple locations:

```bash
coding-context \
  -d git::https://github.com/company/org-standards.git \
  -d git::https://github.com/team/team-rules.git \
  -d https://cdn.company.com/archived-rules.tar.gz \
  /fix-bug
```

Rules from all sources are merged together and processed as if they were in a single directory.

### Mix Remote and Local

Combine remote directories with local project rules:

```bash
# Loads from:
# 1. Remote Git repository (via -d)
# 2. Working directory (automatically added)
# 3. Home directory (automatically added)
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  -s languages=go \
  /fix-bug
```

You can also explicitly add local directories:

```bash
# Explicitly add a local directory
coding-context \
  -d git::https://github.com/company/shared-rules.git \
  -d file:///path/to/local/rules \
  -s languages=go \
  /fix-bug
```

**Note:** The working directory (`-C` or current directory) and home directory are automatically added to search paths, so you don't need to specify them explicitly.

### With Selectors and Parameters

Use remote directories with all normal CLI features:

```bash
coding-context \
  -d git::https://github.com/company/standards.git \
  -s languages=go \
  -s environment=production \
  -p component=auth \
  -p severity=critical \
  /fix-security-issue | llm -m claude-3-5-sonnet-20241022
```

## Repository Structure

Your remote repository should follow the same structure as local directories:

```
shared-rules/
├── .agents/
│   ├── rules/
│   │   ├── coding-standards.md
│   │   ├── security-guidelines.md
│   │   └── testing-requirements.md
│   └── tasks/
│       ├── code-review.md
│       └── fix-security-issue.md
└── README.md
```

The CLI searches for standard locations:
- `.agents/rules/` - Rule files
- `.agents/tasks/` - Task files
- `AGENTS.md`, `CLAUDE.md`, etc. - Single-file rules
- `.github/copilot-instructions.md` - GitHub Copilot rules
- And all other [standard search paths](../reference/search-paths)

## Example Repository

Here's a complete example of setting up a shared rules repository:

### 1. Create Repository

```bash
# Create repository
mkdir shared-coding-rules
cd shared-coding-rules
git init

# Create directory structure
mkdir -p .agents/rules
mkdir -p .agents/tasks
```

### 2. Add Rules

**`.agents/rules/go-standards.md`:**
```markdown
---
languages:
  - go
---

# Go Coding Standards

- Use `gofmt` for formatting
- Write tests for all public functions
- Handle all errors explicitly
- Use meaningful variable names
```

**`.agents/rules/security.md`:**
```markdown
# Security Guidelines

- Never commit secrets or API keys
- Validate all user input
- Use prepared statements for SQL
- Keep dependencies up to date
```

### 3. Add Tasks

**`.agents/tasks/code-review.md`:**
```markdown
---
task_name: code-review
---

# Code Review Task

Review the code changes for:
- Adherence to coding standards
- Security vulnerabilities
- Test coverage
- Documentation
```

### 4. Publish

```bash
# Commit and push
git add .
git commit -m "Initial coding standards"
git tag v1.0.0
git remote add origin https://github.com/company/shared-coding-rules.git
git push -u origin main
git push --tags
```

### 5. Use in Projects

```bash
# In any project
coding-context \
  -d 'git::https://github.com/company/shared-coding-rules.git?ref=v1.0.0' \
  -s languages=go \
  /code-review
```

## Bootstrap Scripts

Bootstrap scripts work in remote directories:

**Remote repository:**
```
.agents/
├── rules/
│   ├── jira-context.md
│   └── jira-context-bootstrap  # Executable script
```

The `jira-context-bootstrap` script runs before processing the rule, just like with local files.

## Important Notes

### Performance

- Remote directories are downloaded on each invocation
- Use specific tags/commits for reproducible builds
- Git shallow clones are used automatically (faster downloads)

### Caching

Remote directories are **not cached** between runs. Each invocation downloads fresh:

```bash
# Downloads repository
coding-context -d git::https://github.com/company/rules.git /fix-bug

# Downloads again (not cached)
coding-context -d git::https://github.com/company/rules.git /code-review
```

For better performance with frequently-used remote directories, consider:
- Cloning locally and using `file://` protocol
- Using a CI/CD cache
- Hosting on a fast CDN

### Cleanup

Downloaded directories are automatically cleaned up after execution. No manual cleanup needed.

### Authentication

**Git over HTTPS with token:**
```bash
export GITHUB_TOKEN="ghp_your_token_here"
coding-context -d "git::https://${GITHUB_TOKEN}@github.com/company/private-rules.git" /fix-bug
```

**Git over SSH:**
```bash
# Uses your SSH keys from ~/.ssh/
coding-context -d git::git@github.com:company/private-rules.git /fix-bug
```

**S3 with AWS credentials:**
```bash
# Uses AWS credentials from environment or ~/.aws/credentials
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
coding-context -d s3::https://s3.amazonaws.com/my-bucket/rules /fix-bug
```

## Troubleshooting

### "Failed to download remote directory"

**Check:**
1. URL is accessible (test with `git clone` or `curl`)
2. Authentication is configured (tokens, SSH keys)
3. Correct protocol prefix (`git::`, `s3::`, etc.)
4. Network connectivity

**Example:**
```bash
# Test Git URL
git clone https://github.com/company/rules.git /tmp/test-clone

# Test HTTP URL
curl -I https://example.com/rules.tar.gz
```

### "No rules found"

The remote directory might not have the expected structure.

**Check:**
```bash
# Download manually and inspect
git clone https://github.com/company/rules.git /tmp/inspect
ls -la /tmp/inspect/.agents/rules/
```

Ensure the repository has files in standard locations like `.agents/rules/`.

### Slow downloads

**Solutions:**
1. Use specific tags instead of branches (more cacheable)
2. Use subdirectories (`//path`) to download less data
3. Host archives on a CDN for faster downloads
4. Consider local caching for frequently-used remote directories

## Best Practices

1. **Version your rules** - Use git tags for stable releases
2. **Document structure** - Add README to explain organization
3. **Test changes** - Test rule changes before publishing
4. **Use subdirectories** - Organize rules by category
5. **Include examples** - Provide example usage in repository
6. **Pin versions** - Use `?ref=v1.0.0` in production

## See Also

- [CLI Reference](../reference/cli) - Complete `-d` flag documentation
- [Search Paths Reference](../reference/search-paths) - Where files are discovered
- [File Formats](../reference/file-formats) - Rule and task file specifications
- [go-getter Documentation](https://github.com/hashicorp/go-getter) - Supported protocols and features

