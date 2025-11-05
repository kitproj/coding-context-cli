package context

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config holds the configuration for context assembly
type Config struct {
	// WorkDir is the working directory to use
	WorkDir string
	// TaskName is the name of the task to execute
	TaskName string
	// Params are parameters for substitution in task prompts
	Params ParamMap
	// Selectors are frontmatter selectors for filtering rules
	Selectors SelectorMap
	// Stdout is where assembled context is written (defaults to os.Stdout)
	Stdout io.Writer
	// Stderr is where progress messages are written (defaults to os.Stderr)
	Stderr io.Writer
	// Visitor is called for each selected rule (defaults to DefaultRuleVisitor)
	Visitor RuleVisitor
}

// Assembler assembles context from rule and task files
type Assembler struct {
	config Config
}

// NewAssembler creates a new context assembler with the given configuration
func NewAssembler(config Config) *Assembler {
	if config.Stdout == nil {
		config.Stdout = os.Stdout
	}
	if config.Stderr == nil {
		config.Stderr = os.Stderr
	}
	if config.Params == nil {
		config.Params = make(ParamMap)
	}
	if config.Selectors == nil {
		config.Selectors = make(SelectorMap)
	}
	if config.Visitor == nil {
		config.Visitor = &DefaultRuleVisitor{
			stdout: config.Stdout,
			stderr: config.Stderr,
		}
	}
	return &Assembler{config: config}
}

// Assemble assembles the context and writes it to the configured output
func (a *Assembler) Assemble(ctx context.Context) error {
	// Change to work directory if specified
	if a.config.WorkDir != "" {
		if err := os.Chdir(a.config.WorkDir); err != nil {
			return fmt.Errorf("failed to chdir to %s: %w", a.config.WorkDir, err)
		}
	}

	// Add task name to selectors so rules can be filtered by task
	a.config.Selectors["task_name"] = a.config.TaskName

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// find the task prompt
	var taskPromptPath string
	taskPromptPaths := []string{
		filepath.Join(".agents", "tasks", a.config.TaskName+".md"),
		filepath.Join(homeDir, ".agents", "tasks", a.config.TaskName+".md"),
		filepath.Join("/etc", "agents", "tasks", a.config.TaskName+".md"),
	}
	for _, path := range taskPromptPaths {
		if _, err := os.Stat(path); err == nil {
			taskPromptPath = path
			break
		}
	}

	if taskPromptPath == "" {
		return fmt.Errorf("prompt file not found for task: %s in %v", a.config.TaskName, taskPromptPaths)
	}

	// Track total tokens
	var totalTokens int

	for _, rule := range []string{
		"CLAUDE.local.md",

		".agents/rules",
		".cursor/rules",
		".augment/rules",
		".windsurf/rules",
		".opencode/agent",
		".opencode/command",

		".github/copilot-instructions.md",
		".gemini/styleguide.md",
		".github/agents",
		".augment/guidelines.md",

		"AGENTS.md",
		"CLAUDE.md",
		"GEMINI.md",

		".cursorrules",
		".windsurfrules",

		// ancestors
		"../AGENTS.md",
		"../CLAUDE.md",
		"../GEMINI.md",

		"../../AGENTS.md",
		"../../CLAUDE.md",
		"../../GEMINI.md",

		// user
		filepath.Join(homeDir, ".agents", "rules"),
		filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		filepath.Join(homeDir, ".codex", "AGENTS.md"),
		filepath.Join(homeDir, ".gemini", "GEMINI.md"),
		filepath.Join(homeDir, ".opencode", "rules"),

		// system
		"/etc/agents/rules",
		"/etc/opencode/rules",
	} {

		// Skip if the path doesn't exist
		if _, err := os.Stat(rule); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat rule path %s: %w", rule, err)
		}

		err := filepath.Walk(rule, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Only process .md and .mdc files as rule files
			ext := filepath.Ext(path)
			if ext != ".md" && ext != ".mdc" {
				return nil
			}

			// Parse frontmatter to check selectors
			var frontmatter map[string]string
			content, err := ParseMarkdownFile(path, &frontmatter)
			if err != nil {
				return fmt.Errorf("failed to parse markdown file: %w", err)
			}

			// Check if file matches include selectors.
			// Note: Files with duplicate basenames will both be included.
			if !a.config.Selectors.MatchesIncludes(frontmatter) {
				fmt.Fprintf(a.config.Stderr, "⪢ Excluding rule file (does not match include selectors): %s\n", path)
				return nil
			}

			// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
			// For example, setup.md -> setup-bootstrap, setup.mdc -> setup-bootstrap
			baseNameWithoutExt := strings.TrimSuffix(path, ext)
			bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

			if _, err := os.Stat(bootstrapFilePath); err == nil {
				// Bootstrap file exists, make it executable and run it before printing content
				if err := os.Chmod(bootstrapFilePath, 0755); err != nil {
					return fmt.Errorf("failed to chmod bootstrap file %s: %w", bootstrapFilePath, err)
				}

				fmt.Fprintf(a.config.Stderr, "⪢ Running bootstrap script: %s\n", bootstrapFilePath)

				cmd := exec.CommandContext(ctx, bootstrapFilePath)
				cmd.Stdout = a.config.Stderr
				cmd.Stderr = a.config.Stderr

				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to run bootstrap script: %w", err)
				}
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("failed to stat bootstrap file %s: %w", bootstrapFilePath, err)
			}

			// Create Rule object and visit it
			tokens := EstimateTokens(content)
			totalTokens += tokens
			
			ruleObj := &Rule{
				Path:        path,
				Content:     content,
				Frontmatter: frontmatter,
				Tokens:      tokens,
			}
			
			// Visit the rule using the configured visitor
			if err := a.config.Visitor.VisitRule(ctx, ruleObj); err != nil {
				return fmt.Errorf("visitor error for rule %s: %w", path, err)
			}

			return nil

		})
		if err != nil {
			return fmt.Errorf("failed to walk rule dir: %w", err)
		}
	}

	content, err := ParseMarkdownFile(taskPromptPath, &struct{}{})
	if err != nil {
		return fmt.Errorf("failed to parse prompt file %s: %w", taskPromptPath, err)
	}

	expanded := os.Expand(content, func(key string) string {
		if val, ok := a.config.Params[key]; ok {
			return val
		}
		// this might not exist, in that case, return the original text
		return fmt.Sprintf("${%s}", key)
	})

	// Estimate tokens for this file
	tokens := EstimateTokens(expanded)
	totalTokens += tokens
	fmt.Fprintf(a.config.Stderr, "⪢ Including task file: %s (~%d tokens)\n", taskPromptPath, tokens)

	fmt.Fprintln(a.config.Stdout, expanded)

	// Print total token count
	fmt.Fprintf(a.config.Stderr, "⪢ Total estimated tokens: %d\n", totalTokens)

	return nil
}
