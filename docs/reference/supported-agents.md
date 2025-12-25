---
layout: default
title: Supported AI Coding Agents
parent: Reference
nav_order: 4
---

# Supported AI Coding Agents

This document provides a comprehensive list of AI coding agents and assistants that the Coding Context CLI can integrate with. The tool is designed to be agent-agnostic, but provides specific support for popular agents' configuration file locations and conventions.

## Currently Supported Agents

The following agents have explicit support with dedicated configuration file paths:

### 1. GitHub Copilot
- **Website**: [https://github.com/features/copilot](https://github.com/features/copilot)
- **Provider**: GitHub (Microsoft)
- **Config Locations**: 
  - `.github/copilot-instructions.md`
  - `.github/agents`
- **User Rules Path**: `~/.github/agents/AGENTS.md`
- **Description**: AI pair programmer integrated into VS Code, Visual Studio, JetBrains IDEs, and Neovim
- **Agent Flag**: `-a copilot`

### 2. Anthropic Claude
- **Website**: [https://claude.ai/](https://claude.ai/)
- **Provider**: Anthropic
- **Config Locations**: 
  - `CLAUDE.md`
  - `CLAUDE.local.md`
  - `.claude/CLAUDE.md`
- **User Rules Path**: `~/.claude/CLAUDE.md`
- **Description**: Large language model with extended context windows, accessible via API and web interface
- **Agent Flag**: `-a claude`

### 3. Cursor
- **Website**: [https://cursor.sh/](https://cursor.sh/)
- **Provider**: Cursor
- **Config Locations**: 
  - `.cursor/rules`
  - `.cursorrules`
  - `.cursor/commands`
- **User Rules Path**: `~/.cursor/rules/AGENTS.md`
- **Description**: AI-first code editor built on VS Code with integrated AI chat and editing
- **Agent Flag**: `-a cursor`

### 4. Google Gemini
- **Website**: [https://gemini.google.com/](https://gemini.google.com/)
- **Provider**: Google
- **Config Locations**: 
  - `GEMINI.md`
  - `.gemini/styleguide.md`
- **User Rules Path**: `~/.gemini/GEMINI.md`
- **Description**: Google's multimodal AI model with coding capabilities
- **Agent Flag**: `-a gemini`

### 5. Augment
- **Website**: [https://augmentcode.com/](https://augmentcode.com/)
- **Provider**: Augment
- **Config Locations**: 
  - `.augment/rules`
  - `.augment/guidelines.md`
- **User Rules Path**: `~/.augment/AGENTS.md`
- **Description**: AI coding assistant with contextual understanding
- **Agent Flag**: `-a augment`

### 6. Windsurf
- **Website**: [https://codeium.com/windsurf](https://codeium.com/windsurf)
- **Provider**: Codeium
- **Config Locations**: 
  - `.windsurf/rules`
  - `.windsurfrules`
- **User Rules Path**: `~/.windsurf/AGENTS.md`
- **Description**: AI-powered IDE from Codeium
- **Agent Flag**: `-a windsurf`

### 7. OpenCode.ai
- **Website**: [https://opencode.ai/](https://opencode.ai/)
- **Provider**: OpenCode.ai
- **Config Locations**: 
  - `.opencode/agent`
  - `.opencode/command`
  - `.opencode/rules`
- **User Rules Path**: `~/.opencode/AGENTS.md`
- **Description**: Open-source AI coding platform
- **Agent Flag**: `-a opencode`

### 8. Codex
- **Website**: [https://codex.ai/](https://codex.ai/)
- **Provider**: Codex
- **Config Locations**: 
  - `AGENTS.md`
  - `.codex/AGENTS.md`
- **User Rules Path**: `~/.codex/AGENTS.md`
- **Description**: AI coding assistant platform
- **Agent Flag**: `-a codex`

## Additional Agents to Consider

These agents are widely used but may not yet have explicit configuration path support:

### AI-Powered IDEs and Editors

#### 9. Codeium
- **Website**: [https://codeium.com/](https://codeium.com/)
- **Provider**: Codeium
- **Description**: Free AI code completion tool with support for 70+ languages
- **Supported IDEs**: VS Code, JetBrains, Vim/Neovim, Emacs, Eclipse, and more
- **Note**: Parent company of Windsurf

#### 10. Tabnine
- **Website**: [https://www.tabnine.com/](https://www.tabnine.com/)
- **Provider**: Tabnine
- **Description**: AI code completion that can run locally or in the cloud
- **Supported IDEs**: VS Code, JetBrains, Vim, Sublime, Atom, and more
- **Privacy**: Offers local model option for sensitive code

#### 11. Amazon CodeWhisperer (now Amazon Q Developer)
- **Website**: [https://aws.amazon.com/q/developer/](https://aws.amazon.com/q/developer/)
- **Provider**: Amazon Web Services
- **Description**: AI coding companion with AWS service integration
- **Supported IDEs**: VS Code, JetBrains, AWS Cloud9, AWS Lambda console

#### 12. Replit Ghostwriter
- **Website**: [https://replit.com/](https://replit.com/)
- **Provider**: Replit
- **Description**: AI pair programmer integrated into Replit's online IDE
- **Features**: Code completion, generation, and debugging

#### 13. Sourcegraph Cody
- **Website**: [https://sourcegraph.com/cody](https://sourcegraph.com/cody)
- **Provider**: Sourcegraph
- **Description**: AI coding assistant with codebase context awareness
- **Supported IDEs**: VS Code, JetBrains, Neovim
- **Features**: Code search, understanding, and generation

### Standalone AI Coding Tools

#### 14. OpenAI GPT-4 / ChatGPT
- **Website**: [https://openai.com/](https://openai.com/)
- **Provider**: OpenAI
- **Description**: General-purpose LLM with strong coding capabilities
- **Access**: Web interface, API, and third-party integrations

#### 15. GPT-Engineer
- **Website**: [https://github.com/gpt-engineer-org/gpt-engineer](https://github.com/gpt-engineer-org/gpt-engineer)
- **Provider**: Open Source
- **Description**: Autonomous agent that generates entire codebases from prompts

#### 16. Aider
- **Website**: [https://aider.chat/](https://aider.chat/)
- **Provider**: Open Source
- **Description**: AI pair programming in your terminal with Git integration
- **Features**: Direct file editing, Git commits, works with various LLMs

#### 17. Continue
- **Website**: [https://continue.dev/](https://continue.dev/)
- **Provider**: Open Source
- **Description**: Open-source autopilot for VS Code and JetBrains
- **Features**: Chat, edit, and generate with any LLM

#### 18. Copilot++ / Copilot Workspace
- **Website**: [https://githubnext.com/](https://githubnext.com/)
- **Provider**: GitHub Next (Microsoft)
- **Description**: Experimental features and future of GitHub Copilot

### Cloud-Based Development Environments

#### 19. GitHub Codespaces
- **Website**: [https://github.com/features/codespaces](https://github.com/features/codespaces)
- **Provider**: GitHub (Microsoft)
- **Description**: Cloud-based development environment with Copilot integration
- **Features**: Pre-configured containers, VS Code in browser

#### 20. GitLab Duo
- **Website**: [https://about.gitlab.com/gitlab-duo/](https://about.gitlab.com/gitlab-duo/)
- **Provider**: GitLab
- **Description**: AI-powered features across GitLab platform
- **Features**: Code suggestions, chat, vulnerability explanations

### Agent Frameworks and Platforms

#### 21. LangChain
- **Website**: [https://www.langchain.com/](https://www.langchain.com/)
- **Provider**: LangChain
- **Description**: Framework for developing LLM-powered applications
- **Use Case**: Build custom coding agents

#### 22. AutoGPT
- **Website**: [https://github.com/Significant-Gravitas/AutoGPT](https://github.com/Significant-Gravitas/AutoGPT)
- **Provider**: Open Source
- **Description**: Autonomous AI agent framework
- **Use Case**: Task automation including code generation

#### 23. BabyAGI
- **Website**: [https://github.com/yoheinakajima/babyagi](https://github.com/yoheinakajima/babyagi)
- **Provider**: Open Source
- **Description**: AI-powered task management system
- **Use Case**: Autonomous coding task decomposition and execution

### Specialized Coding Assistants

#### 24. Phind
- **Website**: [https://www.phind.com/](https://www.phind.com/)
- **Provider**: Phind
- **Description**: AI search engine for developers with coding focus

#### 25. Bard (now Gemini)
- **Website**: [https://bard.google.com/](https://bard.google.com/)
- **Provider**: Google
- **Description**: Conversational AI with code generation (now merged into Gemini)

#### 26. Perplexity AI
- **Website**: [https://www.perplexity.ai/](https://www.perplexity.ai/)
- **Provider**: Perplexity
- **Description**: AI search with coding capabilities

#### 27. Blackbox AI
- **Website**: [https://www.blackbox.ai/](https://www.blackbox.ai/)
- **Provider**: Blackbox AI
- **Description**: AI coding assistant with real-time knowledge

#### 28. CodeGPT
- **Website**: [https://codegpt.co/](https://codegpt.co/)
- **Provider**: CodeGPT
- **Description**: AI assistant for developers in IDE

#### 29. Pieces for Developers
- **Website**: [https://pieces.app/](https://pieces.app/)
- **Provider**: Pieces
- **Description**: AI-powered code snippet manager and workflow tool

#### 30. Mintlify
- **Website**: [https://mintlify.com/](https://mintlify.com/)
- **Provider**: Mintlify
- **Description**: AI documentation generator from code

### Enterprise and Specialized Solutions

#### 31. Codegen (by Salesforce)
- **Provider**: Salesforce Research
- **Description**: Open-source code generation models

#### 32. StarCoder / BigCode
- **Website**: [https://huggingface.co/bigcode](https://huggingface.co/bigcode)
- **Provider**: Hugging Face / BigCode
- **Description**: Open-source code generation models

#### 33. WizardCoder
- **Provider**: WizardLM team
- **Description**: Code-focused LLM fine-tuned from CodeLlama

#### 34. DeepSeek Coder
- **Website**: [https://github.com/deepseek-ai/DeepSeek-Coder](https://github.com/deepseek-ai/DeepSeek-Coder)
- **Provider**: DeepSeek
- **Description**: Open-source code intelligence model

#### 35. Mistral Codestral
- **Website**: [https://mistral.ai/](https://mistral.ai/)
- **Provider**: Mistral AI
- **Description**: Code-specialized model from Mistral

## Agent Support Tiers

### Tier 1: Full Support
Agents with dedicated configuration paths and `-a` flag support:
- GitHub Copilot
- Claude
- Cursor
- Gemini
- Augment
- Windsurf
- OpenCode.ai
- Codex

### Tier 2: Compatible
Agents that can use the tool via standard input/output but lack specific configuration support:
- Codeium
- Tabnine
- Amazon Q Developer
- Sourcegraph Cody
- Continue
- Aider
- All LLM APIs (OpenAI, Anthropic, Google, etc.)

### Tier 3: Framework Integration
Agent frameworks that can incorporate this tool:
- LangChain
- AutoGPT
- BabyAGI
- Custom agent implementations

## Adding Support for New Agents

To add full support for a new agent:

1. **Define configuration paths** in the codebase (`pkg/codingcontext/paths.go`)
2. **Add agent constant** to the agent types
3. **Define user rules path** for the `-w` flag
4. **Update documentation** in README.md and this file
5. **Add tests** for the new agent's configuration discovery

Example configuration:
```go
// In agent.go
const (
    AgentNewAgent Agent = "newagent"
)

// In paths.go
var newAgentPaths = []string{
    ".newagent/rules",
    ".newagentrules",
}
```

## Recommendations

### For Individual Developers
- **GitHub Copilot** or **Cursor**: Best IDE integration
- **Claude** or **ChatGPT**: Best for complex problem-solving
- **Aider**: Best for terminal-based workflows

### For Teams
- **GitHub Copilot**: Best for Microsoft ecosystem
- **Cursor**: Best for AI-first coding experience
- **Sourcegraph Cody**: Best for large codebases
- **Amazon Q Developer**: Best for AWS-heavy projects

### For Privacy-Conscious Users
- **Tabnine** (local mode): Run models locally
- **Continue**: Open-source, use any LLM
- **Aider**: Open-source, Git-integrated

### For Research and Experimentation
- **OpenAI GPT-4**: Most capable general model
- **Claude**: Largest context window
- **Open-source models**: StarCoder, DeepSeek Coder, WizardCoder

## See Also

- [CLI Reference](./cli) - Command-line options including `-a` flag
- [File Formats](./file-formats) - Configuration file formats
- [Search Paths](./search-paths) - Where the tool looks for agent configs
- [Use with AI Agents](../how-to/use-with-ai-agents) - Practical usage examples
