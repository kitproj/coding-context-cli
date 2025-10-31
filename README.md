# Coding Context CLI

A CLI tool for managing context files for coding agents. It helps you organize prompts, memories (reusable context), and bootstrap scripts that can be assembled into a single context file for AI coding agents.

## Why Use This?

When working with AI coding agents (like GitHub Copilot, ChatGPT, Claude, etc.), providing the right context is crucial for getting quality results. However, managing this context becomes challenging when:

- **Context is scattered**: Project conventions, coding standards, and setup instructions are spread across multiple documents
- **Repetition is tedious**: You find yourself copy-pasting the same information into every AI chat session
- **Context size is limited**: AI models have token limits, so you need to efficiently select what's relevant
- **Onboarding is manual**: New team members or agents need step-by-step setup instructions

**This tool solves these problems by:**

1. **Centralizing reusable context** - Store project conventions, coding standards, and setup instructions once in "memory" files
2. **Creating task-specific prompts** - Define templated prompts for common tasks (e.g., "add feature", "fix bug", "refactor")
3. **Automating environment setup** - Package bootstrap scripts that prepare the environment before an agent starts work
4. **Filtering context dynamically** - Use selectors to include only relevant context (e.g., production vs. development, Python vs. Go)
5. **Composing everything together** - Generate a single `prompt.md` file combining all relevant context and the task prompt

## When to Use

This tool is ideal for:

- **Working with AI coding agents** - Prepare comprehensive context before starting a coding session
- **Team standardization** - Share common prompts and conventions across your team
- **Complex projects** - Manage large amounts of project-specific context efficiently
- **Onboarding automation** - New developers or agents can run bootstrap scripts to set up their environment
- **Multi-environment projects** - Filter context based on environment (dev/staging/prod) or technology stack

## How It Works

The basic workflow is:

1. **Organize your context** - Create memory files (shared context) and prompt files (task-specific instructions)
2. **Run the CLI** - Execute `coding-context <task-name>` with optional parameters
3. **Get assembled output** - The tool generates:
   - `prompt.md` - Combined context + task prompt with template variables filled in
   - `bootstrap` - Executable script to set up the environment
   - `bootstrap.d/` - Individual bootstrap scripts from your memory files
4. **Use with AI agents** - Share `prompt.md` with your AI coding agent, or run `./bootstrap` to prepare the environment first

**Visual flow:**
```
+---------------------+       +--------------------------+
| Memory Files (*.md) |       | Prompt Template          |
|                     |       | (task-name.md)           |
+----------+----------+       +------------+-------------+
           |                               |
           | Filter by selectors           | Apply template params
           v                               v
+---------------------+       +--------------------------+
| Filtered Memories   +-------+ Rendered Prompt          |
+---------------------+       +------------+-------------+
                                           |
                                           v
                              +----------------------------+
                              | prompt.md (combined output)|
                              +----------------------------+
```

## Installation

Download the binary for your platform from the release page:

```bash
sudo curl -fsL -o /usr/local/bin/coding-context https://github.com/kitproj/coding-agent-context-cli/releases/download/v0.0.1/coding-context_v0.0.1_linux_arm64
sudo chmod +x /usr/local/bin/coding-context
```

### Using Go Install

```bash
go install github.com/kitproj/coding-agent-context-cli@latest
```

## Usage

```
coding-context [options] <task-name>

Options:
  -b                Automatically run the bootstrap script after generating it
  -d <directory>    Add a directory to include in the context (can be used multiple times)
                    Default: .prompts, ~/.config/prompts, /var/local/prompts
  -o <directory>    Output directory for generated files (default: .)
  -p <key=value>    Template parameter for prompt substitution (can be used multiple times)
  -s <key=value>    Include memories with matching frontmatter (can be used multiple times)
  -S <key=value>    Exclude memories with matching frontmatter (can be used multiple times)
```

**Example:**
```bash
coding-context -p feature="Authentication" -p language=Go add-feature
```

**Example with selectors:**
```bash
# Include only production memories
coding-context -s env=production deploy

# Exclude test memories
coding-context -S env=test deploy

# Combine include and exclude selectors
coding-context -s env=production -S language=python deploy
```

## Quick Start

This 4-step guide shows how to set up and generate your first context:

**Step 1: Create a context directory structure**
```bash
mkdir -p .prompts/{tasks,memories}
```

**Step 2: Create a memory file** (`.prompts/memories/project-info.md`)

Memory files are included in every generated context. They contain reusable information like project conventions, architecture notes, or coding standards.

```markdown
# Project Context

- Framework: Go CLI
- Purpose: Manage AI agent context
```

**Step 3: Create a prompt file** (`.prompts/tasks/my-task.md`)

Prompt files define specific tasks. They can use template variables (like `${taskName}` or `$taskName`) that you provide via command-line parameters.

```markdown
# Task: ${taskName}

Please help me with this task. The project uses ${language}.
```

**Step 4: Generate your context file**

```bash
coding-context -p taskName="Fix Bug" -p language=Go my-task
```

**Result:** This generates `./prompt.md` combining your memories and the task prompt with template variables filled in. You can now share this complete context with your AI coding agent!

**What you'll see in `prompt.md`:**
```markdown
# Project Context

- Framework: Go CLI
- Purpose: Manage AI agent context



# Task: Fix Bug

Please help me with this task. The project uses Go.
```


## Directory Structure

The tool searches these directories for context files (in priority order):
1. `.prompts/` (project-local)
2. `~/.config/prompts/` (user-specific)
3. `/var/local/prompts/` (system-wide)

Each directory should contain:
```
.prompts/
├── tasks/          # Task-specific prompt templates
│   └── <task-name>.md
└── memories/         # Reusable context files (included in all outputs)
    └── *.md
```


## File Formats

### Prompt Files

Markdown files with YAML frontmatter and Go template support.

**Example** (`.prompts/tasks/add-feature.md`):
```markdown
# Task: ${feature}

Implement ${feature} in ${language}.
```

Run with:
```bash
coding-context -p feature="User Login" -p language=Go add-feature
```

### Memory Files

Markdown files included in every generated context. Bootstrap scripts can be provided in separate files.

**Example** (`.prompts/memories/setup.md`):
```markdown
---
env: development
language: go
---
# Development Setup

This project requires Node.js dependencies.
```

**Bootstrap file** (`.prompts/memories/setup-bootstrap`):
```bash
#!/bin/bash
npm install
```

For each memory file `<name>.md`, you can optionally create a corresponding `<name>-bootstrap` file that will be executed during setup.


## Filtering Memories with Selectors

Use the `-s` and `-S` flags to filter which memory files are included based on their frontmatter metadata.

### Selector Syntax

- **`-s key=value`** - Include memories where the frontmatter key matches the value
- **`-S key=value`** - Exclude memories where the frontmatter key matches the value
- If a key doesn't exist in a memory's frontmatter, the memory is allowed (not filtered out)
- Multiple selectors of the same type use AND logic (all must match)

### Examples

**Include only production memories:**
```bash
coding-context -s env=production deploy
```

**Exclude test environment:**
```bash
coding-context -S env=test deploy
```

**Combine include and exclude:**
```bash
# Include production but exclude python
coding-context -s env=production -S language=python deploy
```

**Multiple includes:**
```bash
# Only production Go backend memories
coding-context -s env=production -s language=go -s tier=backend deploy
```

### How It Works

When you run with selectors, the tool logs which files are included or excluded:

```
INFO Including memory file path=.prompts/memories/production.md
INFO Excluding memory file (does not match include selectors) path=.prompts/memories/development.md
INFO Including memory file path=.prompts/memories/nofrontmatter.md
```

**Important:** Files without the specified frontmatter keys are still included. This allows you to have generic memories that apply to all scenarios.

If no selectors are specified, all memory files are included.


## Output Files

- **`prompt.md`** - Combined output with all memories and the task prompt
- **`bootstrap`** - Executable script that runs all bootstrap scripts from memories
- **`bootstrap.d/`** - Individual bootstrap scripts (SHA256 named)

Run the bootstrap script to set up your environment:
```bash
./bootstrap
```

Or use the `-b` flag to automatically run the bootstrap script after generating it:
```bash
coding-context -b my-task
```


## Examples

### Basic Usage

```bash
# Create structure
mkdir -p .prompts/{tasks,memories}

# Add a memory
cat > .prompts/memories/conventions.md << 'EOF'
# Coding Conventions

- Use tabs for indentation
- Write tests for all functions
EOF

# Create a task prompt
cat > .prompts/tasks/refactor.md << 'EOF'
# Refactoring Task

Please refactor the codebase to improve code quality.
EOF

# Generate context
coding-context refactor
```

### With Template Parameters

```bash
cat > .prompts/tasks/add-feature.md << 'EOF'
# Add Feature: ${featureName}

Implement ${featureName} in ${language}.
EOF

coding-context -p featureName="Authentication" -p language=Go add-feature
```

### With Bootstrap Scripts

```bash
cat > .prompts/memories/setup.md << 'EOF'
# Project Setup

This Go project uses modules.
EOF

cat > .prompts/memories/setup-bootstrap << 'EOF'
#!/bin/bash
go mod download
EOF
chmod +x .prompts/memories/setup-bootstrap

coding-context -o ./output my-task
cd output && ./bootstrap
```

Alternatively, use the `-b` flag to automatically run the bootstrap script:
```bash
coding-context -o ./output -b my-task
```

### Integrating External CLI Tools

The bootstrap script mechanism is especially useful for integrating external CLI tools like `kitproj/jira-cli` and `kitproj/slack-cli`. These tools can be installed automatically when an agent starts working on a task.

#### Example: Using kitproj/jira-cli

The `kitproj/jira-cli` tool allows agents to interact with Jira issues programmatically. Here's how to set it up:

**Step 1: Create a memory file with Jira context** (`.prompts/memories/jira.md`)

```markdown
---
tools: jira
---
# Jira Integration

This project uses Jira for issue tracking. The `jira` CLI tool is available for interacting with issues.

## Available Commands

- `jira get-issue <issue-id>` - Get details of a Jira issue
- `jira get-comments <issue-id>` - Get all comments on an issue
- `jira add-comment <issue-id> <comment-text>` - Add a comment to an issue
- `jira update-issue-status <issue-id> <status>` - Update the status of an issue
- `jira create-issue <project-key> <summary> <description>` - Create a new issue

## Configuration

The Jira CLI is configured with:
- Server URL: https://your-company.atlassian.net
- Authentication: Token-based (set via JIRA_API_TOKEN environment variable)
```

**Step 2: Create a bootstrap script** (`.prompts/memories/jira-bootstrap`)

```bash
#!/bin/bash
set -euo pipefail

# Install jira-cli if not already installed
if ! command -v jira &> /dev/null; then
    echo "Installing jira-cli..."
    
    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    # Map architecture names
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac
    
    # Download and install the latest version
    VERSION="v0.1.0"  # Update to the latest version
    BINARY_URL="https://github.com/kitproj/jira-cli/releases/download/${VERSION}/jira-cli_${VERSION}_${OS}_${ARCH}"
    
    sudo curl -fsSL -o /usr/local/bin/jira "$BINARY_URL"
    sudo chmod +x /usr/local/bin/jira
    
    echo "jira-cli installed successfully"
else
    echo "jira-cli is already installed"
fi

# Verify installation
jira --version
```

**Step 3: Make the bootstrap script executable**

```bash
chmod +x .prompts/memories/jira-bootstrap
```

**Step 4: Use with a task that needs Jira**

```bash
# The bootstrap will automatically run when you generate context
coding-context -p storyId="PROJ-123" implement-jira-story

# This creates ./bootstrap which installs jira-cli when executed
./bootstrap
```

Now when an agent starts work, the bootstrap script will ensure `jira-cli` is installed and ready to use!

#### Example: Using kitproj/slack-cli

The `kitproj/slack-cli` tool allows agents to send notifications and interact with Slack channels. Here's the setup:

**Step 1: Create a memory file with Slack context** (`.prompts/memories/slack.md`)

```markdown
---
tools: slack
---
# Slack Integration

This project uses Slack for team communication. The `slack` CLI tool is available for sending messages and notifications.

## Available Commands

- `slack send-message <channel> <message>` - Send a message to a channel
- `slack send-thread-reply <channel> <thread-ts> <message>` - Reply to a thread
- `slack upload-file <channel> <file-path>` - Upload a file to a channel
- `slack set-status <status-text> <emoji>` - Set your Slack status
- `slack get-channel-history <channel> <limit>` - Get recent messages from a channel

## Configuration

The Slack CLI requires:
- Workspace: your-workspace.slack.com
- Authentication: Bot token (set via SLACK_BOT_TOKEN environment variable)
- Channels: Use channel IDs or names (e.g., #engineering, #alerts)

## Common Use Cases

- Send build notifications: `slack send-message "#builds" "Build completed successfully"`
- Report deployment status: `slack send-message "#deployments" "Production deployment started"`
- Alert on failures: `slack send-message "#alerts" "Test suite failed on main branch"`
```

**Step 2: Create a bootstrap script** (`.prompts/memories/slack-bootstrap`)

```bash
#!/bin/bash
set -euo pipefail

# Install slack-cli if not already installed
if ! command -v slack &> /dev/null; then
    echo "Installing slack-cli..."
    
    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    # Map architecture names
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac
    
    # Download and install the latest version
    VERSION="v0.1.0"  # Update to the latest version
    BINARY_URL="https://github.com/kitproj/slack-cli/releases/download/${VERSION}/slack-cli_${VERSION}_${OS}_${ARCH}"
    
    sudo curl -fsSL -o /usr/local/bin/slack "$BINARY_URL"
    sudo chmod +x /usr/local/bin/slack
    
    echo "slack-cli installed successfully"
else
    echo "slack-cli is already installed"
fi

# Verify installation
slack --version
```

**Step 3: Make the bootstrap script executable**

```bash
chmod +x .prompts/memories/slack-bootstrap
```

**Step 4: Create a task that uses Slack** (`.prompts/tasks/slack-deploy-alert.md`)

```markdown
# Slack Deployment Alert: ${environment}

## Task

Send a deployment notification to the team via Slack.

## Steps

1. **Prepare the notification message**
   - Include environment: ${environment}
   - Include deployment status
   - Include relevant details (version, commit, etc.)

2. **Send to appropriate channels**
   ```bash
   slack send-message "#deployments" "🚀 Deployment to ${environment} started"
   ```

3. **Update on completion**
   ```bash
   slack send-message "#deployments" "✅ Deployment to ${environment} completed successfully"
   ```

4. **Alert on failures** (if needed)
   ```bash
   slack send-message "#alerts" "❌ Deployment to ${environment} failed. Check logs for details."
   ```

## Success Criteria
- Team is notified of deployment status
- Appropriate channels receive updates
- Messages are clear and actionable
```

**Step 5: Use the task**

```bash
coding-context -p environment="production" slack-deploy-alert
./bootstrap  # Installs slack-cli if needed
```

#### Writing Bootstrap Scripts - Best Practices

When writing bootstrap scripts for external CLI tools:

1. **Check if already installed** - Avoid reinstalling if the tool exists
   ```bash
   if ! command -v toolname &> /dev/null; then
       # Install logic here
   fi
   ```

2. **Handle multiple platforms** - Detect OS and architecture
   ```bash
   OS=$(uname -s | tr '[:upper:]' '[:lower:]')
   ARCH=$(uname -m)
   case "$ARCH" in
       x86_64) ARCH="amd64" ;;
       aarch64|arm64) ARCH="arm64" ;;
   esac
   ```

3. **Use specific versions** - Pin to a specific version for reproducibility
   ```bash
   VERSION="v0.1.0"
   ```

4. **Set error handling** - Use `set -euo pipefail` to catch errors early
   ```bash
   #!/bin/bash
   set -euo pipefail
   ```

5. **Verify installation** - Check that the tool works after installation
   ```bash
   toolname --version
   ```

6. **Provide clear output** - Echo messages to show progress
   ```bash
   echo "Installing toolname..."
   echo "Installation complete"
   ```

### Real-World Task Examples

Here are some practical task templates for common development workflows:

#### Implement Jira Story

**Note:** This example assumes you've set up the Jira CLI integration as shown in the [Using kitproj/jira-cli](#example-using-kitprojjira-cli) section above. The bootstrap script will automatically install the `jira` command.

```bash
cat > .prompts/tasks/implement-jira-story.md << 'EOF'
---
---
# Implement Jira Story: ${storyId}

## Story Details

First, get the full story details from Jira:
```bash
jira get-issue ${storyId}
```

## Requirements

Please implement the feature described in the Jira story. Follow these steps:

1. **Review the Story**
   - Read the story details, acceptance criteria, and comments
   - Get all comments: `jira get-comments ${storyId}`
   - Clarify any uncertainties by adding comments: `jira add-comment ${storyId} "Your question"`

2. **Start Development**
   - Create a feature branch with the story ID in the name (e.g., `feature/${storyId}-implement-auth`)
   - Move the story to "In Progress": `jira update-issue-status ${storyId} "In Progress"`

3. **Implementation**
   - Design the solution following project conventions
   - Implement the feature with proper error handling
   - Add comprehensive unit tests (aim for >80% coverage)
   - Update documentation if needed
   - Ensure all tests pass and code is lint-free

4. **Update Jira Throughout**
   - Add progress updates: `jira add-comment ${storyId} "Completed implementation, working on tests"`
   - Keep stakeholders informed of any blockers or changes

5. **Complete the Story**
   - Ensure all acceptance criteria are met
   - Create a pull request
   - Move to review: `jira update-issue-status ${storyId} "In Review"`
   - Once merged, close: `jira update-issue-status ${storyId} "Done"`

## Success Criteria
- All acceptance criteria are met
- Code follows project coding standards
- Tests are passing
- Documentation is updated
- Jira story is properly tracked through workflow
EOF

# Usage
coding-context -p storyId="PROJ-123" implement-jira-story
```

#### Triage Jira Bug

**Note:** This example requires the Jira CLI integration. See [Using kitproj/jira-cli](#example-using-kitprojjira-cli) for setup instructions.

```bash
cat > .prompts/tasks/triage-jira-bug.md << 'EOF'
---
---
# Triage Jira Bug: ${bugId}

## Get Bug Details

First, retrieve the full bug report from Jira:
```bash
jira get-issue ${bugId}
jira get-comments ${bugId}
```

## Triage Steps

1. **Acknowledge and Take Ownership**
   - Add initial comment: `jira add-comment ${bugId} "Triaging this bug now"`
   - Move to investigation: `jira update-issue-status ${bugId} "In Progress"`

2. **Reproduce the Issue**
   - Follow the steps to reproduce in the bug report
   - Verify the issue exists in the reported environment
   - Document actual vs. expected behavior
   - Update Jira: `jira add-comment ${bugId} "Reproduced on [environment]. Actual: [X], Expected: [Y]"`

3. **Investigate Root Cause**
   - Review relevant code and logs
   - Identify the component/module causing the issue
   - Determine if this is a regression (check git history)
   - Document findings: `jira add-comment ${bugId} "Root cause: [description]"`

4. **Assess Impact**
   - How many users are affected?
   - Is there a workaround available?
   - What is the risk if left unfixed?
   - Add assessment: `jira add-comment ${bugId} "Impact: [severity]. Workaround: [yes/no]. Affected users: [estimate]"`

5. **Provide Triage Report**
   - Root cause analysis
   - Recommended priority level
   - Estimated effort to fix
   - Suggested assignee/team
   - Final summary: `jira add-comment ${bugId} "Triage complete. Priority: [level]. Effort: [estimate]. Recommended assignee: [name]"`

## Output
Provide a detailed triage report with your findings and recommendations, and post it as a comment to the Jira issue.
EOF

# Usage
coding-context -p bugId="PROJ-456" triage-jira-bug
```

#### Respond to Jira Comment

**Note:** This example requires the Jira CLI integration. See [Using kitproj/jira-cli](#example-using-kitprojjira-cli) for setup instructions.

```bash
cat > .prompts/tasks/respond-to-jira-comment.md << 'EOF'
---
---
# Respond to Jira Comment: ${issueId}

## Get Issue and Comments

First, retrieve the issue details and all comments:
```bash
jira get-issue ${issueId}
jira get-comments ${issueId}
```

Review the latest comment and the full context of the issue.

## Instructions

Please analyze the comment and provide a professional response:

1. **Acknowledge** the comment and any concerns raised
2. **Address** each question or point made
3. **Provide** technical details or clarifications as needed
4. **Suggest** next steps or actions if appropriate
5. **Maintain** a collaborative and helpful tone

## Response Guidelines
- Be clear and concise
- Provide code examples if relevant
- Link to documentation when helpful
- Offer to discuss further if needed

## Post Your Response

Once you've formulated your response, add it to the Jira issue:
```bash
jira add-comment ${issueId} "Your detailed response here"
```

If the comment requires action on your part, update the issue status accordingly:
```bash
jira update-issue-status ${issueId} "In Progress"
```
EOF

# Usage
coding-context -p issueId="PROJ-789" respond-to-jira-comment
```

#### Send Slack Notification on Build Completion

**Note:** This example requires the Slack CLI integration. See [Using kitproj/slack-cli](#example-using-kitprojslack-cli) for setup instructions.

```bash
cat > .prompts/tasks/notify-build-status.md << 'EOF'
---
---
# Notify Build Status: ${buildStatus}

## Task

Send a build status notification to the team via Slack.

## Build Information
- Status: ${buildStatus}
- Branch: ${branch}
- Commit: ${commit}
- Build Time: ${buildTime}

## Steps

1. **Prepare the notification message**
   - Determine the appropriate emoji based on status
   - Include all relevant build details
   - Add links to build logs or artifacts

2. **Send notification to #builds channel**

   For successful builds:
   ```bash
   slack send-message "#builds" "✅ Build succeeded on ${branch}
Commit: ${commit}
Time: ${buildTime}
Status: ${buildStatus}"
   ```

   For failed builds:
   ```bash
   slack send-message "#builds" "❌ Build failed on ${branch}
Commit: ${commit}
Time: ${buildTime}
Status: ${buildStatus}
Please check the build logs for details."
   ```

3. **Alert in #alerts channel for failures** (if build failed)
   ```bash
   slack send-message "#alerts" "🚨 Build failure detected on ${branch}. Immediate attention needed."
   ```

4. **Update thread if this is a rebuild**
   If responding to a previous build notification:
   ```bash
   slack send-thread-reply "#builds" "<thread-timestamp>" "Rebuild completed: ${buildStatus}"
   ```

## Success Criteria
- Appropriate channels are notified
- Message includes all relevant details
- Team can quickly assess build status
- Failed builds trigger alerts
EOF

# Usage
coding-context -p buildStatus="SUCCESS" -p branch="main" -p commit="abc123" -p buildTime="2m 30s" notify-build-status
```

#### Post Deployment Notification to Slack

**Note:** This example requires the Slack CLI integration. See [Using kitproj/slack-cli](#example-using-kitprojslack-cli) for setup instructions.

```bash
cat > .prompts/tasks/notify-deployment.md << 'EOF'
---
---
# Notify Deployment: ${environment}

## Task

Communicate deployment status to stakeholders via Slack.

## Deployment Details
- Environment: ${environment}
- Version: ${version}
- Deployer: ${deployer}

## Instructions

1. **Announce deployment start**
   ```bash
   slack send-message "#deployments" "🚀 Deployment to ${environment} started
Version: ${version}
Deployer: ${deployer}
Started at: $(date)"
   ```

2. **Monitor deployment progress**
   - Track deployment steps
   - Note any issues or delays

3. **Send completion notification**

   For successful deployments:
   ```bash
   slack send-message "#deployments" "✅ Deployment to ${environment} completed successfully
Version: ${version}
Completed at: $(date)
All services are healthy and running."
   ```

   For failed deployments:
   ```bash
   slack send-message "#deployments" "❌ Deployment to ${environment} failed
Version: ${version}
Failed at: $(date)
Rolling back to previous version..."
   ```

4. **Alert stakeholders for production deployments**
   ```bash
   slack send-message "#general" "📢 Production deployment completed: version ${version} is now live!"
   ```

5. **Update status thread**
   - Reply to the initial announcement with final status
   - Include any post-deployment tasks or notes

## Success Criteria
- Deployment timeline is clearly communicated
- All stakeholders are informed
- Status updates are timely and accurate
- Issues are escalated appropriately
EOF

# Usage
coding-context -p environment="production" -p version="v2.1.0" -p deployer="deploy-bot" notify-deployment
```

#### Review Pull Request

```bash
cat > .prompts/tasks/review-pull-request.md << 'EOF'
---
---
# Review Pull Request: ${prNumber}

## PR Details
- PR #${prNumber}
- Author: ${author}
- Title: ${title}

## Review Checklist

### Code Quality
- [ ] Code follows project style guidelines
- [ ] No obvious bugs or logic errors
- [ ] Error handling is appropriate
- [ ] No security vulnerabilities introduced
- [ ] Performance considerations addressed

### Testing
- [ ] Tests are included for new functionality
- [ ] Tests cover edge cases
- [ ] All tests pass
- [ ] Test quality is high (clear, maintainable)

### Documentation
- [ ] Public APIs are documented
- [ ] Complex logic has explanatory comments
- [ ] README updated if needed
- [ ] Breaking changes are noted

### Architecture
- [ ] Changes align with project architecture
- [ ] No unnecessary dependencies added
- [ ] Code is modular and reusable
- [ ] Separation of concerns maintained

## Instructions
Please review the pull request thoroughly and provide:
1. Constructive feedback on any issues found
2. Suggestions for improvements
3. Approval or request for changes
4. Specific line-by-line comments where helpful

Be thorough but encouraging. Focus on learning and improvement.
EOF

# Usage
coding-context -p prNumber="42" -p author="Jane" -p title="Add feature X" review-pull-request
```

#### Respond to Pull Request Comment

```bash
cat > .prompts/tasks/respond-to-pull-request-comment.md << 'EOF'
---
---
# Respond to Pull Request Comment

## PR Details
- PR #${prNumber}
- Reviewer: ${reviewer}
- File: ${file}

## Comment
${comment}

## Instructions

Please address the pull request review comment:

1. **Analyze** the feedback carefully
2. **Determine** if the comment is valid
3. **Respond** professionally:
   - If you agree: Acknowledge and describe your fix
   - If you disagree: Respectfully explain your reasoning
   - If unclear: Ask clarifying questions

4. **Make changes** if needed:
   - Fix the issue raised
   - Add tests if applicable
   - Update documentation
   - Ensure code still works

5. **Reply** with:
   - What you changed (with commit reference)
   - Why you made that choice
   - Any additional context needed

## Tone
Be collaborative, open to feedback, and focused on code quality.
EOF

# Usage
coding-context -p prNumber="42" -p reviewer="Bob" -p file="main.go" -p comment="Consider using a switch here" respond-to-pull-request-comment
```

#### Fix Failing Check

```bash
cat > .prompts/tasks/fix-failing-check.md << 'EOF'
---
---
# Fix Failing Check: ${checkName}

## Check Details
- Check Name: ${checkName}
- Branch: ${branch}
- Status: FAILED

## Debugging Steps

1. **Identify the Failure**
   - Review the check logs
   - Identify the specific error message
   - Determine which component is failing

2. **Reproduce Locally**
   - Pull the latest code from ${branch}
   - Run the same check locally
   - Verify you can reproduce the failure

3. **Root Cause Analysis**
   - Is this a new failure or regression?
   - What recent changes might have caused it?
   - Is it environment-specific?

4. **Fix the Issue**
   - Implement the fix
   - Verify the check passes locally
   - Ensure no other checks are broken
   - Add tests to prevent regression

5. **Validate**
   - Run all relevant checks locally
   - Push changes and verify CI passes
   - Update any related documentation

## Common Check Types
- **Tests**: Fix failing unit/integration tests
- **Linter**: Address code style issues
- **Build**: Resolve compilation errors
- **Security**: Fix vulnerability scans
- **Coverage**: Improve test coverage

Please fix the failing check and ensure all CI checks pass.
EOF

# Usage
coding-context -p checkName="Unit Tests" -p branch="main" fix-failing-check
```

## Advanced Usage

### Template Variables

Prompts use shell-style variable expansion via `os.Expand`:

```markdown
${variableName}    # Braced variable substitution
$variableName      # Simple variable substitution (works with alphanumeric names)
```

Variables that are not provided via `-p` flag are replaced with empty strings.

### Directory Priority

When the same task exists in multiple directories, the first match wins:
1. `.prompts/` (highest priority)
2. `~/.config/prompts/`
3. `/var/local/prompts/` (lowest priority)

## Troubleshooting

**"prompt file not found for task"**
- Ensure `<task-name>.md` exists in a `tasks/` subdirectory

**"failed to walk memory dir"**
```bash
mkdir -p .prompts/memories
```

**Template parameter shows `<no value>`**
```bash
coding-context -p myvar="value" my-task
```

**Bootstrap script not executing**
```bash
chmod +x bootstrap
```

