# Coding Context CLI - Slide Deck

This directory contains a presentation slide deck for the Coding Context CLI tool.

## Files

- **[SLIDES.md](./SLIDES.md)** - Main slide deck in Marp format
- **[SLIDES.pdf](./SLIDES.pdf)** - Pre-generated PDF (ready to use)

## Viewing the Slides

### Quick Start: Download the PDF

The easiest way to view or present the slides is to download the pre-generated PDF:
- **[SLIDES.pdf](./SLIDES.pdf)** - Ready to open in any PDF viewer

### Generate Your Own

The slides are written in [Marp](https://marp.app/) markdown format. You have several options to generate or view them:

### Option 1: Marp CLI (Recommended)

Install Marp CLI globally:

```bash
npm install -g @marp-team/marp-cli
```

Then view the slides in various formats:

```bash
# View as HTML in browser
marp SLIDES.md -o slides.html && open slides.html

# Export to PDF
marp SLIDES.md -o slides.pdf

# Export to PowerPoint
marp SLIDES.md -o slides.pptx

# Watch mode (auto-reload on changes)
marp -w SLIDES.md
```

### Option 2: Marp for VS Code

1. Install the [Marp for VS Code](https://marketplace.visualstudio.com/items?itemName=marp-team.marp-vscode) extension
2. Open `SLIDES.md` in VS Code
3. Click the preview icon or press `Ctrl+K V` (Windows/Linux) or `Cmd+K V` (macOS)
4. Use the export function to save as HTML, PDF, or PowerPoint

### Option 3: Online Viewer

Use the Marp Web interface:
1. Go to https://web.marp.app/
2. Copy the contents of `SLIDES.md`
3. Paste into the editor
4. View and export

### Option 4: Docker

If you have Docker installed:

```bash
# Export to HTML
docker run --rm -v $PWD:/home/marp/app/ -e LANG=en_US.UTF-8 marpteam/marp-cli SLIDES.md -o slides.html

# Export to PDF
docker run --rm -v $PWD:/home/marp/app/ -e LANG=en_US.UTF-8 marpteam/marp-cli SLIDES.md -o slides.pdf

# Export to PowerPoint
docker run --rm -v $PWD:/home/marp/app/ -e LANG=en_US.UTF-8 marpteam/marp-cli SLIDES.md -o slides.pptx
```

## Presenting

### Keyboard Shortcuts (HTML output)

When viewing the HTML version in a browser:

- **Arrow keys** / **Space** / **Page Up/Down** - Navigate slides
- **F** - Fullscreen mode
- **Esc** - Exit fullscreen
- **?** - Show keyboard shortcuts

### Presenter Notes

To add presenter notes (not visible in slides):

```markdown
---

## Slide Title

Slide content here

<!--
Presenter notes go here
They won't appear in the slides
-->

---
```

## Customizing the Slides

### Themes

Change the theme by modifying the frontmatter at the top of `SLIDES.md`:

```yaml
---
marp: true
theme: default  # Options: default, gaia, uncover
---
```

### Custom Styles

The slides include custom CSS in the frontmatter. Modify the `style:` section to change colors, fonts, or layout:

```yaml
style: |
  section {
    font-size: 28px;
  }
  h1 {
    color: #2c3e50;
  }
```

### Page Classes

Apply different layouts to specific slides:

```markdown
---

<!-- _class: lead -->
# Centered Title Slide

---

<!-- _class: invert -->
# Dark Background Slide

---
```

## Export Formats

### HTML

Best for:
- Web hosting
- Interactive viewing
- Easy sharing via URL

```bash
marp SLIDES.md -o slides.html
```

### PDF

Best for:
- Printing
- Offline distribution
- Universal compatibility

```bash
marp SLIDES.md -o slides.pdf
```

### PowerPoint (PPTX)

Best for:
- Further editing in PowerPoint/Keynote
- Corporate environments
- Adding animations

```bash
marp SLIDES.md -o slides.pptx
```

## Hosting the Slides

### GitHub Pages

Add to your repository's GitHub Pages:

```bash
marp SLIDES.md -o docs/slides.html
git add docs/slides.html
git commit -m "Add slide deck"
git push
```

Access at: `https://<username>.github.io/<repo>/slides.html`

### Self-Hosted

Upload the HTML file to any web server:

```bash
marp SLIDES.md -o index.html
# Upload index.html to your server
```

## Slide Deck Contents

The slide deck covers:

1. **Introduction** - Problem statement and solution
2. **Key Features** - Dynamic assembly, rules, tasks, remote directories
3. **Installation** - Quick setup for Linux and macOS
4. **Usage** - Basic commands and examples
5. **Advanced Features** - Selectors, bootstrap scripts, expansion
6. **Integration** - AI agents, GitHub Actions, agentic workflows
7. **Best Practices** - Tips for effective usage
8. **Examples** - Real-world use cases
9. **Resources** - Documentation and community links

## Tips for Presenting

1. **Practice navigation** - Familiarize yourself with the slide flow
2. **Prepare demos** - Have terminal windows ready for live demonstrations
3. **Know your audience** - Adjust depth based on technical level
4. **Use examples** - Refer to the example sections when questions arise
5. **Share resources** - Point to documentation links at the end

## Updating the Slides

When making changes:

1. Edit `SLIDES.md`
2. Test locally: `marp SLIDES.md -o test.html && open test.html`
3. Commit changes to version control
4. Re-export in desired formats

## License

The slide deck is part of the Coding Context CLI project and is licensed under the MIT License.

## Contributing

To suggest improvements to the slides:

1. Fork the repository
2. Edit `SLIDES.md`
3. Test your changes with Marp
4. Submit a pull request

## Resources

- [Marp Documentation](https://marpit.marp.app/)
- [Marp CLI](https://github.com/marp-team/marp-cli)
- [Marp for VS Code](https://marketplace.visualstudio.com/items?itemName=marp-team.marp-vscode)
- [Coding Context CLI Documentation](https://kitproj.github.io/coding-context-cli/)
