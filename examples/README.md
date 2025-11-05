# Examples: Agentic Workflows Integration

This directory contains examples demonstrating how to use the Coding Context CLI with agentic workflows, particularly with GitHub Actions.

## Directory Structure

```
examples/
├── workflows/          # GitHub Actions workflow examples
│   ├── agentic-code-review.yml
│   ├── agentic-bugfix.yml
│   └── agentic-feature-development.yml
└── agents/            # Example agent configuration
    ├── rules/         # Context rules for different stages
    │   ├── planning.md
    │   ├── implementation-go.md
    │   └── testing.md
    └── tasks/         # Task definitions
        ├── plan-feature.md
        ├── code-review.md
        └── fix-bug.md
```

## Workflow Examples

### 1. Agentic Code Review (`agentic-code-review.yml`)

Automatically reviews pull requests using AI with context assembled from your repository's coding standards and guidelines.

**Features:**
- Triggered on PR creation/update
- Assembles review context with PR-specific parameters
- Posts review comments
- Tracks review metrics

**Usage:**
Copy to `.github/workflows/agentic-code-review.yml` in your repository.

### 2. Agentic Bug Fix (`agentic-bugfix.yml`)

Autonomously attempts to fix bugs when issues are labeled as bugs.

**Features:**
- Triggered when 'bug' label is added to issues
- Determines bug severity from labels
- Creates fix branch and implements solution
- Runs tests to verify the fix
- Creates PR with the fix

**Usage:**
Copy to `.github/workflows/agentic-bugfix.yml` in your repository.

### 3. Agentic Feature Development (`agentic-feature-development.yml`)

Multi-stage workflow that plans, implements, tests, and documents features autonomously.

**Features:**
- Manual trigger with feature details
- Planning phase creates feature plan
- Implementation phase writes code
- Testing phase generates tests
- Review phase creates documentation
- Creates comprehensive PR

**Usage:**
Copy to `.github/workflows/agentic-feature-development.yml` in your repository.

## Agent Configuration Examples

### Rules

Rules provide context for different stages and aspects of development:

- **`planning.md`**: Guidelines for feature planning and architecture
- **`implementation-go.md`**: Go-specific coding standards (can be adapted for other languages)
- **`testing.md`**: Testing requirements and best practices

Rules use frontmatter selectors to be included only when relevant:

```markdown
---
stage: planning
priority: high
---

# Your rule content here
```

### Tasks

Tasks define specific workflows for agents to execute:

- **`plan-feature.md`**: Creates comprehensive feature plans
- **`code-review.md`**: Performs code reviews on PRs
- **`fix-bug.md`**: Analyzes and fixes bugs

Tasks use parameter substitution for dynamic content:

```markdown
---
task_name: fix-bug
---

# Bug Fix Task

Issue: #${issue_number}
Title: ${issue_title}
```

## Getting Started

1. **Copy examples to your repository:**
   ```bash
   # Copy agent configuration
   mkdir -p .agents/rules .agents/tasks
   cp examples/agents/rules/* .agents/rules/
   cp examples/agents/tasks/* .agents/tasks/
   
   # Copy workflow examples
   mkdir -p .github/workflows
   cp examples/workflows/agentic-*.yml .github/workflows/
   ```

2. **Customize for your project:**
   - Update rules to match your coding standards
   - Modify tasks for your specific workflows
   - Adjust workflow triggers and parameters
   - Replace `your-org/ai-agent-action@v1` with your actual AI agent action

3. **Test locally:**
   ```bash
   # Test context assembly
   coding-context-cli -s stage=planning plan-feature
   
   # Test with parameters
   coding-context-cli \
     -s task=code-review \
     -p pr_number=123 \
     code-review
   ```

4. **Deploy workflows:**
   - Commit the workflows and agent configuration
   - Push to your repository
   - Workflows will activate based on their triggers

## Integration with AI Agents

These examples assume you have an AI agent action available. The agent action should:

1. Accept a context file as input
2. Execute the specified task using the AI model
3. Output results in a structured format
4. Support parameters like model, temperature, max-tokens

Example agent action usage:

```yaml
- name: Execute AI Task
  uses: your-org/ai-agent-action@v1
  with:
    context-file: context.txt
    task: code-review
    model: claude-3-5-sonnet-20241022
    temperature: 0.2
```

## Adapting for Different Languages

The `implementation-go.md` example can be adapted for other languages:

- Create `implementation-python.md`, `implementation-javascript.md`, etc.
- Add `language: python` to frontmatter
- Use selectors in workflows: `-s language=python`

## Best Practices

1. **Version control your rules**: Rules are part of your project's standards
2. **Use selectors effectively**: Target specific contexts for different scenarios
3. **Parameterize tasks**: Make tasks reusable across different instances
4. **Test context assembly**: Verify context before using in workflows
5. **Monitor token usage**: Keep context size manageable for AI models
6. **Iterate on prompts**: Refine task prompts based on agent performance

## Contributing

To add new examples:
1. Create the workflow or agent configuration
2. Add documentation explaining the use case
3. Test thoroughly
4. Submit a pull request

## Related Documentation

- [AGENTIC_WORKFLOWS.md](../AGENTIC_WORKFLOWS.md) - Detailed guide on agentic workflows
- [README.md](../README.md) - Coding Context CLI documentation

## License

These examples are provided as-is for educational and reference purposes.
