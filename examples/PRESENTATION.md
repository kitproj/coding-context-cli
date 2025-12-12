# Presentation Example

This example demonstrates how to use the slide deck for presentations about the Coding Context CLI.

## Quick Start

The main slide deck is available in two formats:
- [`../SLIDES.md`](../SLIDES.md) - Marp markdown format (editable)
- [`../SLIDES.pdf`](../SLIDES.pdf) - Pre-generated PDF (ready to use)

**For immediate use:** Download and open [`SLIDES.pdf`](../SLIDES.pdf) in any PDF viewer.

## Viewing Options

### 1. PDF Viewer (Easiest)

Simply open [`../SLIDES.pdf`](../SLIDES.pdf) in any PDF viewer:
- Adobe Acrobat Reader
- Preview (macOS)
- Edge, Chrome, or Firefox
- Your system's default PDF viewer

### 2. VS Code (For Editing)

1. Install the [Marp for VS Code](https://marketplace.visualstudio.com/items?itemName=marp-team.marp-vscode) extension
2. Open `SLIDES.md` in VS Code
3. Press `Ctrl+K V` (Windows/Linux) or `Cmd+K V` (macOS) to open preview
4. Use the export button to generate HTML, PDF, or PowerPoint

### 2. Command Line

```bash
# Install Marp CLI
npm install -g @marp-team/marp-cli

# Generate HTML
marp SLIDES.md -o slides.html

# Generate PDF
marp SLIDES.md --pdf -o slides.pdf

# Generate PowerPoint
marp SLIDES.md --pptx -o slides.pptx

# Watch mode (auto-reload)
marp -w SLIDES.md
```

### 3. Docker

```bash
# HTML
docker run --rm -v $PWD:/home/marp/app/ marpteam/marp-cli SLIDES.md

# PDF
docker run --rm -v $PWD:/home/marp/app/ marpteam/marp-cli --pdf SLIDES.md

# PowerPoint
docker run --rm -v $PWD:/home/marp/app/ marpteam/marp-cli --pptx SLIDES.md
```

## Presentation Scenarios

### Scenario 1: Quick Demo (5 minutes)

Focus on these slides:
- Slide 1-2: Introduction and Problem
- Slide 3-4: Solution and Key Features
- Slide 8: Installation
- Slide 9-10: Basic Usage and Examples
- Slide 50: Resources

**Demo commands:**
```bash
# Simple example
coding-context fix-bug | llm -m claude-3-5-sonnet-20241022

# With parameters
coding-context -p issue=BUG-123 -s languages=go fix-bug
```

### Scenario 2: Technical Deep Dive (15 minutes)

Cover these sections:
- Slides 1-7: Introduction and Features
- Slides 8-16: Installation and Usage
- Slides 17-28: Rule Files, Tasks, and Advanced Features
- Slides 35-42: Best Practices and Examples
- Slide 50: Resources

**Demo commands:**
```bash
# Show rule discovery
coding-context -s languages=go fix-bug | head -50

# Show with selectors
coding-context -s languages=go -s stage=implementation implement-feature

# Show remote rules
coding-context -d git::https://github.com/company/rules.git fix-bug
```

### Scenario 3: Full Workshop (30-45 minutes)

Use the entire slide deck with hands-on exercises:

1. **Introduction** (5 min) - Slides 1-7
2. **Installation** (5 min) - Slides 8-9 + hands-on
3. **Basic Usage** (10 min) - Slides 10-16 + demos
4. **Advanced Features** (10 min) - Slides 17-28 + demos
5. **Integration** (5 min) - Slides 29-34
6. **Best Practices** (5 min) - Slides 35-42
7. **Q&A** (5-10 min)

**Hands-on exercises:**

```bash
# Exercise 1: Create a simple rule
mkdir -p .agents/rules
cat > .agents/rules/testing.md << 'EOF'
---
stage: testing
---
# Testing Standards
- Write unit tests for all new code
- Aim for >80% coverage
EOF

# Exercise 2: Create a task
mkdir -p .agents/tasks
cat > .agents/tasks/write-tests.md << 'EOF'
---
selectors:
  stage: testing
---
# Write Tests for ${module}

Create comprehensive unit tests.
EOF

# Exercise 3: Run the CLI
coding-context -s stage=testing -p module=auth write-tests

# Exercise 4: Try with remote rules
coding-context -d git::https://github.com/kitproj/coding-context-cli.git//examples/agents/rules write-tests
```

## Customizing for Your Audience

### For Developers

Emphasize:
- Technical implementation (slides 17-28)
- CLI usage and options (slides 9-16)
- Integration examples (slides 35-42)
- Code examples and demos

### For Managers/Leadership

Emphasize:
- Problem statement (slides 1-2)
- Business benefits (slides 3-7)
- Team collaboration features (slides 24, 35-42)
- Workflow integration (slides 29-34)

### For DevOps/SRE

Emphasize:
- GitHub Actions integration (slide 30-32)
- Remote directories (slide 24)
- Bootstrap scripts (slide 25)
- Agentic workflows (slides 29-34)

## Presentation Tips

1. **Prepare Your Environment**
   ```bash
   # Have these ready in terminal windows
   cd /path/to/demo/project
   coding-context --help
   ls -la .agents/
   ```

2. **Use Real Examples**
   - Show actual rule files from your project
   - Demonstrate with real task scenarios
   - Use familiar tech stack in examples

3. **Interactive Demos**
   - Let audience suggest selector values
   - Show token counts in real-time
   - Demonstrate different AI agents

4. **Common Questions**
   - Q: "Can we use with our existing CI/CD?"
     A: Yes! Show slide 30 (GitHub Actions integration)
   
   - Q: "How do we share rules across teams?"
     A: Show slide 24 (Remote Directories)
   
   - Q: "What about security?"
     A: Show slide 44 (Security & Privacy)

## After the Presentation

Share these resources with attendees:

- **Documentation**: https://kitproj.github.io/coding-context-cli/
- **GitHub Repo**: https://github.com/kitproj/coding-context-cli
- **Slide Deck**: Available in the repository at `SLIDES.md`
- **Getting Started Guide**: https://kitproj.github.io/coding-context-cli/tutorials/getting-started

## Feedback

After presenting, consider:
- Collecting feedback on which topics resonated
- Noting questions for FAQ additions
- Updating examples based on audience interests
- Contributing improvements back to the slide deck

## Additional Resources

- [Full documentation](https://kitproj.github.io/coding-context-cli/)
- [SLIDES_README.md](../SLIDES_README.md) - Detailed viewing instructions
- [AGENTIC_WORKFLOWS.md](../AGENTIC_WORKFLOWS.md) - Workflow integration guide
- [Examples directory](./agents/) - Sample rules and tasks
