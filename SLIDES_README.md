# Coding Context CLI - Presentation Slides

This directory contains presentation slides explaining the Coding Context CLI tool.

## Slide Deck

**File**: `SLIDES.md`

A comprehensive slide deck covering:
- What problem the tool solves
- How it works
- Key features
- Installation and usage
- Rule and task files
- Advanced features
- Integration with AI agents and agentic workflows
- Real-world examples

## Viewing the Slides

### Option 1: Marp CLI (Recommended)

[Marp](https://marp.app/) is a Markdown presentation ecosystem.

**Install Marp CLI:**
```bash
npm install -g @marp-team/marp-cli
```

**View slides in browser:**
```bash
marp SLIDES.md --preview
```

**Export to HTML:**
```bash
marp SLIDES.md -o slides.html
```

**Export to PDF:**
```bash
marp SLIDES.md -o slides.pdf
```

**Export to PowerPoint:**
```bash
marp SLIDES.md -o slides.pptx
```

### Option 2: Marp for VS Code

Install the [Marp for VS Code](https://marketplace.visualstudio.com/items?itemName=marp-team.marp-vscode) extension:

1. Install the extension from VS Code marketplace
2. Open `SLIDES.md` in VS Code
3. Click the "Open Preview" button (or press `Ctrl+Shift+V` / `Cmd+Shift+V`)
4. Use the "Export" command to save as HTML/PDF/PPTX

### Option 3: Online Viewers

You can also view the slides using online Markdown presentation tools:

- **Marp Web**: https://web.marp.app/ (paste the content)
- **HackMD**: https://hackmd.io/ (supports Marp syntax)
- **Slides.com**: https://slides.com/ (import Markdown)

### Option 4: Plain Markdown

If you just want to read the content, you can view `SLIDES.md` as regular Markdown in:
- GitHub (automatically rendered)
- Any Markdown viewer
- Your favorite text editor

## Slide Features

The slide deck includes:

- **52 slides** covering all aspects of the tool
- Clear visual diagrams showing architecture and workflows
- Code examples with syntax highlighting
- Real-world use cases
- Installation instructions for multiple platforms
- Integration examples with GitHub Actions
- Best practices and tips

## Customization

The slides use Marp's default theme. To customize:

1. **Change theme**: Edit the frontmatter theme (options: `default`, `gaia`, `uncover`)
2. **Modify colors**: Add custom CSS in the frontmatter
3. **Add images**: Use standard Markdown image syntax
4. **Change layout**: Use Marp directives like `<!-- _class: invert -->`

Example customization:
```markdown
---
marp: true
theme: gaia
paginate: true
style: |
  section {
    background-color: #1e1e1e;
    color: white;
  }
---
```

## Presenting Tips

1. **Full screen**: Use your browser's full-screen mode (F11) or Marp's presentation mode
2. **Navigation**: Use arrow keys or click to advance slides
3. **Speaker notes**: Add notes with `<!-- Note: ... -->` (visible in presenter mode)
4. **Print handouts**: Export to PDF and print 2-4 slides per page
5. **Share online**: Export to HTML and host on GitHub Pages or your website

## Content Overview

### Slides 1-10: Introduction
- What problem does it solve?
- How it works
- Key features
- Supported AI agents

### Slides 11-20: Getting Started
- Installation
- Basic usage
- Examples
- Rule files

### Slides 21-30: Core Concepts
- Task files
- Frontmatter filtering
- Task selectors
- Bootstrap scripts

### Slides 31-40: Advanced Features
- Remote file support
- Resume mode
- Agent-specific mode
- File search paths

### Slides 41-52: Integration and Best Practices
- Agentic workflows
- GitHub Actions examples
- Use cases
- Best practices
- Real-world examples

## Updating the Slides

To update the slides:

1. Edit `SLIDES.md` with your changes
2. Use `---` to separate slides
3. Use standard Markdown syntax
4. Test with Marp preview
5. Commit and push changes

## License

These slides are part of the Coding Context CLI project and are licensed under the Apache 2.0 License.
