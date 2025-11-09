# Remote Directory Example

This example demonstrates how to use remote directories with coding-context-cli.

## Use Case

You have coding standards and tasks in a central Git repository or other remote location that you want to share across multiple projects. Instead of duplicating these files in each repository, you can load them from a remote directory.

## Example Setup

### 1. Create a Remote Repository

Create a Git repository with your shared rules and tasks:

```
shared-rules/
├── .agents/
│   ├── rules/
│   │   ├── coding-standards.md
│   │   ├── security-guidelines.md
│   │   └── testing-best-practices.md
│   └── tasks/
│       ├── code-review.md
│       └── fix-security-issue.md
└── README.md
```

### 2. Create Rule Files

**Example: `.agents/rules/coding-standards.md`**

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

### 3. Use the Remote Directory

```bash
# Clone from Git repository
coding-context-cli \
  -d git::https://github.com/company/shared-rules.git \
  fix-bug

# Use a specific branch or tag
coding-context-cli \
  -d 'git::https://github.com/company/shared-rules.git?ref=v1.0' \
  implement-feature

# Use a subdirectory within the repo
coding-context-cli \
  -d 'git::https://github.com/company/mono-repo.git//coding-standards' \
  refactor-code

# Mix local and remote directories
coding-context-cli \
  -d git::https://github.com/company/shared-rules.git \
  -s language=Go \
  implement-feature
```

## Supported Protocols

The `-r` flag uses HashiCorp's go-getter library, which supports many protocols:

### Git Repositories

```bash
# HTTPS
coding-context-cli -d git::https://github.com/company/rules.git fix-bug

# SSH
coding-context-cli -d git::git@github.com:company/rules.git fix-bug

# With authentication token
coding-context-cli -d 'git::https://token@github.com/company/rules.git' fix-bug

# Specific branch
coding-context-cli -d 'git::https://github.com/company/rules.git?ref=main' fix-bug

# Specific tag
coding-context-cli -d 'git::https://github.com/company/rules.git?ref=v1.0.0' fix-bug

# Specific commit
coding-context-cli -d 'git::https://github.com/company/rules.git?ref=abc123' fix-bug

# Subdirectory (note the double slash)
coding-context-cli -d 'git::https://github.com/company/mono.git//standards' fix-bug
```

### HTTP/HTTPS

```bash
# Download and extract tar.gz
coding-context-cli -d https://example.com/rules.tar.gz fix-bug

# Download and extract zip
coding-context-cli -d https://example.com/rules.zip fix-bug

# File server directory
coding-context-cli -d https://example.com/rules/ fix-bug
```

### S3 Buckets

```bash
# S3 bucket
coding-context-cli -d s3::https://s3.amazonaws.com/bucket/rules fix-bug

# With region
coding-context-cli -d s3::https://s3-us-west-2.amazonaws.com/bucket/rules fix-bug
```

### Local Files

```bash
# Local directory (useful for testing)
coding-context-cli -d file:///path/to/local/rules fix-bug
```

## Real-World Example: GitHub Repository

```bash
# Load organization-wide coding standards from GitHub
coding-context-cli \
  -d git::https://github.com/company/shared-rules.git \
  -p component=auth \
  fix-security-issue | llm -m claude-3-sonnet
```

## Benefits

1. **Centralized Management**: Update rules in one place, all projects get the latest version
2. **Version Control**: Use git tags/branches to manage different versions of rules
3. **No Duplication**: Don't need to copy rules into every repository
4. **Easy Distribution**: Share rules across teams and organizations
5. **Mix and Match**: Combine remote directories with local project-specific rules
6. **Full Feature Support**: Bootstrap scripts work in downloaded directories

## Important Notes

- Remote directories are downloaded to a temporary location
- Downloaded directories are cleaned up after execution
- Bootstrap scripts are supported in remote directories
- All standard directory structures are supported (`.agents/rules`, `.agents/tasks`, etc.)
- Downloads happen on each invocation (use git caching for better performance)

## Troubleshooting

### Remote directory not accessible

Check:
1. URL is accessible (test with git clone or curl)
2. Authentication is configured (SSH keys, tokens, etc.)
3. Correct protocol prefix (git::, s3::, etc.)
4. Network connectivity

### Performance concerns

- Remote directories are downloaded on every run
- Use git shallow clones for large repos (automatically done by go-getter)
- Consider local caching strategies for frequently used remote directories
- Use specific tags/commits for reproducible builds

## Advanced: Multiple Remote Sources

You can combine multiple remote directories:

```bash
# Load from multiple sources
coding-context-cli \
  -d git::https://github.com/company/standards.git \
  -d git::https://github.com/team/project-rules.git \
  -d https://cdn.company.com/shared-rules.tar.gz \
  implement-feature
```

This allows for:
- Company-wide standards from one repo
- Team-specific rules from another repo  
- Project-specific rules from a CDN
- Local rules in the current directory

All sources are merged together and processed as if they were in a single directory.
