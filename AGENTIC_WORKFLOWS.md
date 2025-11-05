# Agentic Workflows and Coding Context CLI

## Overview

This document explains the relationship between **GitHub Next's Agentic Workflows** research project and the **Coding Context CLI** tool, and how they complement each other in the broader AI-powered software development ecosystem.

## What are Agentic Workflows?

**GitHub Next's Agentic Workflows** is a research initiative exploring how AI agents can autonomously execute multi-step workflows using GitHub's infrastructure. The key concepts include:

1. **Autonomous Execution**: AI agents can define, plan, and execute complex multi-step workflows without continuous human intervention
2. **GitHub Actions Integration**: Leverages GitHub Actions as the execution environment for agent-driven tasks
3. **Tool Coordination**: Agents orchestrate multiple tools, APIs, and services to accomplish goals
4. **Context-Aware Decision Making**: Agents make decisions based on repository state, code changes, test results, and other contextual information
5. **Iterative Refinement**: Agents can observe outcomes and adjust their approach dynamically

## How Coding Context CLI Fits In

The **Coding Context CLI** serves as a **context preparation and assembly tool** that complements agentic workflows by addressing a critical challenge: **providing rich, relevant context to AI agents**.

### The Context Challenge

AI agents need comprehensive context to make informed decisions:
- Project-specific coding standards and conventions
- Repository structure and architecture
- Technology stack and dependencies
- Team practices and guidelines
- Task-specific requirements and constraints

Manually assembling this context for each task is tedious and error-prone. This is where Coding Context CLI excels.

### The Relationship

```
┌─────────────────────────────────────────────────────────────────┐
│                    Agentic Workflow Ecosystem                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────┐         ┌─────────────────────┐           │
│  │  Context Layer   │────────▶│  Execution Layer     │           │
│  ├──────────────────┤         ├─────────────────────┤           │
│  │                  │         │                      │           │
│  │  Coding Context  │         │  GitHub Actions     │           │
│  │  CLI             │         │  (Agentic Workflows)│           │
│  │                  │         │                      │           │
│  │  • Rules         │         │  • Workflow def     │           │
│  │  • Guidelines    │         │  • Step execution   │           │
│  │  • Tasks         │         │  • Tool calling     │           │
│  │  • Parameters    │         │  • State mgmt       │           │
│  └──────────────────┘         └─────────────────────┘           │
│         │                              │                         │
│         │                              │                         │
│         └──────────────┬───────────────┘                         │
│                        ▼                                         │
│              ┌─────────────────────┐                             │
│              │   AI Agent          │                             │
│              │   (Claude, GPT,     │                             │
│              │    Gemini, etc.)    │                             │
│              └─────────────────────┘                             │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Complementary Roles

| Aspect | Coding Context CLI | Agentic Workflows |
|--------|-------------------|-------------------|
| **Purpose** | Context preparation | Workflow execution |
| **When** | Before agent invocation | During task execution |
| **What** | Assembles rules, guidelines, task prompts | Orchestrates multi-step processes |
| **Where** | Runs locally or in CI/CD | Runs in GitHub Actions |
| **Output** | Formatted context for AI models | Completed tasks, PRs, updates |

## Integration Patterns

### Pattern 1: Pre-Workflow Context Assembly

Use Coding Context CLI to prepare context before initiating an agentic workflow:

```yaml
# .github/workflows/agentic-code-review.yml
name: Agentic Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Assemble Context
        run: |
          coding-context-cli \
            -s task=code-review \
            -p pr_number=${{ github.event.pull_request.number }} \
            code-review > context.txt
      
      - name: Execute AI Review
        uses: github/agent-action@v1
        with:
          context-file: context.txt
          task: review-pull-request
```

### Pattern 2: Dynamic Context for Agent Tasks

Agents can invoke Coding Context CLI during workflow execution to get task-specific context:

```yaml
# .github/workflows/agentic-bugfix.yml
name: Agentic Bug Fix
on:
  issues:
    types: [labeled]

jobs:
  fix-bug:
    if: contains(github.event.issue.labels.*.name, 'bug')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Get Bug Context
        id: context
        run: |
          coding-context-cli \
            -s severity=${{ github.event.issue.labels[0].name }} \
            -p issue_number=${{ github.event.issue.number }} \
            fix-bug > context.txt
          
      - name: Agent Fix Bug
        uses: github/agent-action@v1
        with:
          context-file: context.txt
          task: implement-fix
```

### Pattern 3: Multi-Stage Workflows

Complex workflows can use different context assemblies for different stages:

```yaml
# .github/workflows/agentic-feature.yml
name: Agentic Feature Development
on: workflow_dispatch

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Planning Context
        run: coding-context-cli -s stage=planning plan-feature > plan-context.txt
      - name: Create Plan
        uses: github/agent-action@v1
        with:
          context-file: plan-context.txt
  
  implement:
    needs: plan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Implementation Context
        run: coding-context-cli -s stage=implementation implement-feature > impl-context.txt
      - name: Implement Feature
        uses: github/agent-action@v1
        with:
          context-file: impl-context.txt
  
  test:
    needs: implement
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Testing Context
        run: coding-context-cli -s stage=testing test-feature > test-context.txt
      - name: Run Tests
        uses: github/agent-action@v1
        with:
          context-file: test-context.txt
```

## Best Practices

### 1. Version Control Your Context Rules

Store your `.agents/rules` and `.agents/tasks` in version control alongside your code:

```
.agents/
├── rules/
│   ├── coding-standards.md
│   ├── architecture.md
│   ├── testing.md
│   └── security.md
└── tasks/
    ├── code-review.md
    ├── fix-bug.md
    ├── implement-feature.md
    └── refactor.md
```

### 2. Use Selectors for Context Precision

Leverage frontmatter selectors to provide precise context:

```markdown
---
task_name: fix-bug
severity: critical
language: go
---
# Critical Bug Fix Guidelines

When fixing critical bugs:
- Prioritize data integrity
- Add regression tests
- Document the root cause
```

### 3. Parameterize Task Prompts

Use parameters to inject runtime information:

```markdown
---
task_name: review-pr
---
# Pull Request Review: #${pr_number}

Review the following pull request with focus on:
- Code quality
- Test coverage
- Security implications

PR: ${pr_url}
```

### 4. Bootstrap Dynamic Context

Use bootstrap scripts to fetch real-time information:

```bash
#!/bin/bash
# .agents/rules/jira-bootstrap

# Fetch issue details from Jira
curl -s "https://jira.example.com/api/issue/${JIRA_ISSUE}" \
  | jq -r '.fields.description' \
  > /tmp/jira-context.txt
```

### 5. Organize Rules by Concern

Structure rules to match your workflow stages:

```
.agents/rules/
├── planning/
│   ├── requirements.md
│   └── architecture-decisions.md
├── implementation/
│   ├── coding-standards.md
│   └── api-guidelines.md
└── validation/
    ├── testing-requirements.md
    └── code-review-checklist.md
```

## GitHub Copilot Integration

The Coding Context CLI already supports GitHub Copilot's configuration format:

- `.github/copilot-instructions.md` - Global instructions for Copilot
- `.github/agents/*.md` - Agent-specific configurations

This allows seamless integration with GitHub's official AI tooling:

```
.github/
├── copilot-instructions.md        # Shared instructions
├── agents/
│   ├── code-review-agent.md      # Code review agent config
│   ├── testing-agent.md          # Testing agent config
│   └── documentation-agent.md    # Docs agent config
└── workflows/
    └── agentic-*.yml              # Workflow definitions
```

## Future Possibilities

As agentic workflows evolve, this tool could be extended to:

1. **Workflow Context Injection**: Automatically inject workflow state into context
2. **Agent Memory**: Persist and retrieve context across workflow runs
3. **Context Validation**: Validate that assembled context meets agent requirements
4. **Dynamic Rule Selection**: Use AI to select relevant rules based on task analysis
5. **Feedback Loop**: Learn from workflow outcomes to improve context assembly

## Conclusion

The Coding Context CLI and GitHub's Agentic Workflows are complementary technologies:

- **Coding Context CLI** ensures agents have the right information (context)
- **Agentic Workflows** ensure agents can act on that information (execution)

Together, they enable sophisticated, autonomous software development workflows where AI agents can understand project requirements, follow team conventions, and execute complex tasks with minimal human intervention.

By investing in well-structured context rules and task definitions, teams can build a foundation for increasingly capable agentic workflows that leverage institutional knowledge and best practices encoded in the repository.

## Resources

- [GitHub Next - Agentic Workflows](https://githubnext.com/projects/agentic-workflows/)
- [Coding Context CLI Repository](https://github.com/kitproj/coding-context-cli)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Copilot for Business](https://github.com/features/copilot)
