# Coding Context CLI - GitHub Copilot Instructions

## Project Overview

This is a Go-based command-line tool that assembles dynamic context for AI coding agents. The CLI collects context from rule files, task prompts, and parameters, then outputs a combined context for AI models.

### Key Concepts

- **Rules**: Reusable context snippets (`.md` or `.mdc` files) with optional YAML frontmatter
- **Tasks**: Task-specific prompts identified by `task_name` in frontmatter
- **Selectors**: Frontmatter-based filtering (e.g., `-s language=Go`)
- **Parameters**: Variable substitution in task prompts (e.g., `-p issue_number=123`)
- **Bootstrap Scripts**: Executable scripts that run before rules are processed

## Project Structure

```
.
├── main.go              # CLI entry point and main logic
├── markdown.go          # Frontmatter parsing
├── param_map.go         # Parameter map implementation
├── selector_map.go      # Selector filtering logic
├── token_counter.go     # Token estimation
├── frontmatter.go       # Frontmatter extraction
├── examples/            # Example rules and tasks
│   └── agents/
│       ├── rules/       # Example rule files
│       └── tasks/       # Example task files
└── docs/                # Documentation (Diataxis framework)
```

## Go Coding Standards

### Style and Formatting

- Follow standard Go formatting (`gofmt`)
- Use `goimports` for import organization
- Run `make lint` before committing
- Keep functions focused and small
- Use meaningful variable and function names

### Error Handling

- Always check and handle errors explicitly
- Wrap errors with context using `fmt.Errorf`
- Return errors rather than panicking
- Use custom error types when appropriate

### Testing

- Write unit tests for all new functions
- Use **table-driven tests** for multiple scenarios (this is the project standard)
- Aim for **>80% test coverage**
- Use meaningful test names that describe the scenario (e.g., `TestSelectorMap_MatchesIncludes/single_include_-_match`)
- Test edge cases and error conditions
- Run `go test -v ./...` to verify all tests pass

### Dependencies

- Minimize external dependencies
- Use standard library when possible
- Current dependencies: `go.yaml.in/yaml/v2` for YAML parsing
- Run `go mod tidy` to clean up dependencies

## Build and Development

### Commands

```bash
# Lint the code
make lint

# Build the project
go build -v ./...

# Run all tests
go test -v ./...

# Run integration tests
go test -v -tags=integration ./...
```

### Development Workflow

1. Make changes to code
2. Run `make lint` to ensure code quality
3. Run `go test -v ./...` to verify tests pass
4. Build with `go build -v ./...`

## File Formats

### Rule Files

Rule files use Markdown with optional YAML frontmatter:

```markdown
---
language: Go
stage: implementation
---

# Your Rule Content Here
```

**Important**: Frontmatter selectors only match **top-level fields**. Nested fields are not supported.

### Task Files

Task files require a `task_name` field in frontmatter:

```markdown
---
task_name: fix-bug
resume: false
---

# Task Content
Variable substitution: ${issue_number}
```

### Bootstrap Scripts

- Executable files named `<rule-name>-bootstrap`
- Output goes to `stderr`, not included in AI context
- Used for environment setup (e.g., installing tools)

## Common Patterns

### Adding New Functionality

When adding features:
1. Follow existing patterns in the codebase
2. Add table-driven tests
3. Update documentation in `docs/` if user-facing
4. Handle errors appropriately
5. Maintain backwards compatibility

### Frontmatter Parsing

The project uses a custom frontmatter parser (`markdown.go`):
- Extracts YAML between `---` delimiters
- Parses into `map[string]interface{}`
- Only top-level fields are accessible for selectors

### File Discovery

The tool searches multiple locations in order:
- Current directory (`./.agents/`, `.github/`, etc.)
- Parent directories (for `AGENTS.md`, `CLAUDE.md`, etc.)
- User home directory (`~/.agents/`, etc.)
- System-wide (`/etc/agents/`, etc.)

## Language Specifics

### Go Version

- Target: Go 1.24.4 (specified in `go.mod`)
- Use modern Go idioms and patterns

### Concurrency

- Use `context.Context` for cancellation (already used in main)
- Signal handling via `signal.NotifyContext` (already implemented)

## Documentation

The project uses the **Diataxis framework** for documentation:
- **How-to guides**: `docs/how-to/`
- **Explanation**: `docs/explanation/`
- **Reference**: `docs/reference/`

When adding features, update relevant documentation sections.

## CI/CD

### GitHub Actions

The project uses GitHub Actions for:
- Go build and test (`.github/workflows/go.yml`)
- GitHub Pages deployment (`.github/workflows/pages.yml`)
- Releases (`.github/workflows/release.yml`)

All PRs must pass the Go workflow (lint, build, test).

## Important Notes

- **Resume Mode**: The `-r` flag skips rules and adds `-s resume=true` selector
- **Token Counting**: Estimates are rough (4 chars ≈ 1 token)
- **Selector Matching**: Only top-level YAML fields are supported
- **Backwards Compatibility**: Maintain compatibility with existing rule/task files
- **Minimal Changes**: Prefer small, focused changes over large refactors
