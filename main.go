package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	workDir  string
	params   = make(paramMap)
	includes = make(selectorMap)
	excludes = make(selectorMap)
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Var(&excludes, "S", "Exclude rules with matching frontmatter. Can be specified multiple times as key=value.")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  coding-context [options] <task-name> [persona-name]")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(ctx, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("invalid usage")
	}

	if err := os.Chdir(workDir); err != nil {
		return fmt.Errorf("failed to chdir to %s: %w", workDir, err)
	}

	// Add task name to includes so rules can be filtered by task
	taskName := args[0]
	includes["task_name"] = taskName

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rules := []string{
		"./CLAUDE.local.md",

		"./.agents/rules",
		"./.cursor/rules",
		"./.augment/rules",
		"./.windsurf/rules",

		"./.github/copilot-instructions.md",
		"./.gemini/styleguide.md",
		"./.github/agents",
		"./.augment/guidelines.md",

		"AGENTS.md",
		"CLAUDE.md",
		"GEMINI.md",

		"./.cursorrules",
		"./.windsurfrules",

		// ancestors
		"../AGENTS.md",
		"../CLAUDE.md",
		"../GEMINI.md",

		"../../AGENTS.md",
		"../../CLAUDE.md",
		"../../GEMINI.md",

		// user
		filepath.Join(homeDir, "agents", "rules"),
		filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		filepath.Join(homeDir, ".codex", "AGENTS.md"),
		filepath.Join(homeDir, ".gemini", "GEMINI.md"),

		// system
		"/etc/agents/rules",
	}

	tasks := []string{
		".agents/tasks",
		filepath.Join(homeDir, "agents", "tasks"),
		"/etc/agents/tasks",
	}

	// Track total tokens
	var totalTokens int

	for _, rule := range rules {

		// Skip if the path doesn't exist
		if _, err := os.Stat(rule); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
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
			content, err := parseMarkdownFile(path, &frontmatter)
			if err != nil {
				return fmt.Errorf("failed to parse markdown file: %w", err)
			}

			// Check if file matches include and exclude selectors.
			// Note: Files with duplicate basenames will both be included.
			if !includes.matchesIncludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (does not match include selectors): %s\n", path)
				return nil
			}
			if !excludes.matchesExcludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (matches exclude selectors): %s\n", path)
				return nil
			}

			// Estimate tokens for this file
			tokens := estimateTokens(content)
			totalTokens += tokens
			fmt.Fprintf(os.Stderr, "Including rule file: %s (~%d tokens)\n", path, tokens)
			fmt.Println(content)

			// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
			// For example, setup.md -> setup-bootstrap, setup.mdc -> setup-bootstrap
			baseNameWithoutExt := strings.TrimSuffix(path, ext)
			bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

			if _, err := os.Stat(bootstrapFilePath); os.IsNotExist(err) {
				return nil
			} else if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Running bootstrap script: %s\n", bootstrapFilePath)

			cmd := exec.CommandContext(ctx, bootstrapFilePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run bootstrap script: %w", err)
			}
			return nil

		})
		if err != nil {
			return fmt.Errorf("failed to walk rule dir: %w", err)
		}
	}

	for _, path := range tasks {
		stat, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat task path %s: %w", path, err)
		}
		if stat.IsDir() {
			path = filepath.Join(path, taskName+".md")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return fmt.Errorf("failed to stat task file %s: %w", path, err)
			}
		}

		content, err := parseMarkdownFile(path, &struct{}{})
		if err != nil {
			return fmt.Errorf("failed to parse prompt file: %w", err)
		}

		expanded := os.Expand(content, func(key string) string {
			if val, ok := params[key]; ok {
				return val
			}
			// this might not exist, in that case, return the original text
			return fmt.Sprintf("${%s}", key)
		})

		// Estimate tokens for this file
		tokens := estimateTokens(expanded)
		totalTokens += tokens
		fmt.Fprintf(os.Stderr, "Using task file: %s (~%d tokens)\n", path, tokens)

		fmt.Println(expanded)

		// Print total token count
		fmt.Fprintf(os.Stderr, "Total estimated tokens: %d\n", totalTokens)

		return nil
	}

	return fmt.Errorf("prompt file not found for task: %s", taskName)
}
