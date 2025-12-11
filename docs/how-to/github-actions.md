---
layout: default
title: Integrate with GitHub Actions
parent: How-to Guides
nav_order: 5
---

# How to Integrate with GitHub Actions

This guide shows you how to use the Coding Context CLI in GitHub Actions workflows.

## Installation in Workflows

Add a step to install the CLI:

```yaml
- name: Install Coding Context CLI
  run: |
    curl -fsL -o /usr/local/bin/coding-context \
      https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context_linux_amd64
    chmod +x /usr/local/bin/coding-context
```

## Automated Code Review

Review pull requests automatically:

```yaml
name: AI Code Review
on: pull_request

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install CLI
        run: |
          curl -fsL -o /usr/local/bin/coding-context \
            https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context_linux_amd64
          chmod +x /usr/local/bin/coding-context
      
      - name: Assemble Context
        run: |
          coding-context \
            -s stage=review \
            -p pr_number=${{ github.event.pull_request.number }} \
            -p pr_title="${{ github.event.pull_request.title }}" \
            /code-review > context.txt
      
      - name: Review with AI
        run: |
          cat context.txt | your-ai-agent > review.md
      
      - name: Post Review
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const review = fs.readFileSync('review.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: review
            });
```

## Automated Bug Fixing

Attempt automatic fixes when bugs are reported:

```yaml
name: AI Bug Fix
on:
  issues:
    types: [labeled]

jobs:
  fix:
    if: contains(github.event.issue.labels.*.name, 'bug')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install CLI
        run: |
          curl -fsL -o /usr/local/bin/coding-context \
            https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context_linux_amd64
          chmod +x /usr/local/bin/coding-context
      
      - name: Get Context
        run: |
          coding-context \
            -p issue_number=${{ github.event.issue.number }} \
            -p issue_title="${{ github.event.issue.title }}" \
            -p issue_body="${{ github.event.issue.body }}" \
            /fix-bug > context.txt
      
      - name: Apply AI Fix
        run: |
          cat context.txt | your-ai-agent --apply-changes
      
      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v5
        with:
          title: "Fix: ${{ github.event.issue.title }}"
          body: "Automated fix for issue #${{ github.event.issue.number }}"
          branch: "fix/issue-${{ github.event.issue.number }}"
```

## Multi-Stage Feature Development

Implement features through multiple stages:

```yaml
name: AI Feature Development
on: workflow_dispatch

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install CLI
        run: |
          curl -fsL -o /usr/local/bin/coding-context \
            https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context_linux_amd64
          chmod +x /usr/local/bin/coding-context
      
      - name: Planning Context
        run: |
          coding-context -s stage=planning plan-feature > plan-context.txt
      
      - name: Create Plan
        run: cat plan-context.txt | your-ai-agent > plan.md
      
      - name: Upload Plan
        uses: actions/upload-artifact@v3
        with:
          name: plan
          path: plan.md
  
  implement:
    needs: plan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install CLI
        run: |
          curl -fsL -o /usr/local/bin/coding-context \
            https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context_linux_amd64
          chmod +x /usr/local/bin/coding-context
      
      - name: Download Plan
        uses: actions/download-artifact@v3
        with:
          name: plan
      
      - name: Implementation Context
        run: |
          coding-context -s stage=implementation implement-feature > impl-context.txt
      
      - name: Implement
        run: |
          cat plan.md impl-context.txt | your-ai-agent --apply
```

## Using Environment Secrets

Pass secrets to bootstrap scripts:

```yaml
- name: Assemble Context with Secrets
  env:
    JIRA_API_TOKEN: ${{ secrets.JIRA_API_TOKEN }}
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    coding-context -s source=jira fix-bug > context.txt
```

## Caching CLI Binary

Cache the CLI to speed up workflows:

```yaml
- name: Cache CLI
  id: cache-cli
  uses: actions/cache@v3
  with:
    path: /usr/local/bin/coding-context
    key: coding-context-v0.0.23-linux-amd64  # Include architecture to avoid cache collisions

- name: Install CLI
  if: steps.cache-cli.outputs.cache-hit != 'true'
  run: |
    curl -fsL -o /usr/local/bin/coding-context \
      https://github.com/kitproj/coding-context-cli/releases/latest/download/coding-context_linux_amd64
    chmod +x /usr/local/bin/coding-context
```

## Working Directory

Use the `-C` flag to run from a different directory:

```yaml
- name: Assemble Context
  run: |
    coding-context -C ./backend -s languages=go /fix-bug > context.txt
```

## Best Practices

1. **Pin CLI version**: Use specific release versions for reproducibility
2. **Store rules in repo**: Keep `.agents/` in version control
3. **Use secrets for API keys**: Never hardcode credentials
4. **Cache the CLI binary**: Speed up workflow execution
5. **Review AI output**: Always review changes before merging
6. **Set up approval gates**: Require human approval for automated PRs

## Troubleshooting

**CLI not found:**
- Ensure installation step completes successfully
- Check file permissions (`chmod +x`)

**No rules found:**
- Verify `.agents/` directory is committed to repo
- Check `actions/checkout` step runs first

**Environment variables not available:**
- Set them in workflow `env:` section
- Pass secrets via `${{ secrets.NAME }}`

## See Also

- [Agentic Workflows Explanation](../explanation/agentic-workflows) - Understand the concepts
- [CLI Reference](../reference/cli) - All command options
- [Creating Tasks](./create-tasks) - Define workflow tasks
