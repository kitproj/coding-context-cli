# Coding Context CLI Documentation

This directory contains the source files for the GitHub Pages documentation website.

## Website URL

The documentation is available at: **https://kitproj.github.io/coding-context-cli/**

## Structure

- `index.md` - Home page with overview and quick start
- `tutorials/` - Step-by-step learning guides
- `how-to/` - Problem-oriented practical guides
- `reference/` - Technical specifications
- `explanation/` - Conceptual documentation
- `_config.yml` - Jekyll configuration

## Building Locally

To build and preview the site locally:

```bash
# Install Jekyll and dependencies
gem install bundler jekyll

# Create a Gemfile in the docs directory
cat > Gemfile << 'EOF'
source 'https://rubygems.org'
gem 'github-pages', group: :jekyll_plugins
gem 'just-the-docs'
EOF

# Install dependencies
bundle install

# Serve the site locally
bundle exec jekyll serve --source . --baseurl ""
```

Then open http://localhost:4000 in your browser.

## Deployment

The site is automatically deployed to GitHub Pages when changes are pushed to the `main` branch via the `.github/workflows/pages.yml` workflow.

## Theme

The site uses the Just the Docs theme, which provides excellent navigation and search capabilities. The theme is configured in `_config.yml`.

## Updating Documentation

1. Edit the Markdown files in this directory
2. Test changes locally (optional)
3. Commit and push to the repository
4. Changes will be automatically deployed

## Navigation

The site uses the Just the Docs theme, which automatically generates navigation from frontmatter in each page (using `nav_order`, `parent`, and `has_children` fields).
