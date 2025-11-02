package main

import (
	"context"
	"crypto/sha256"
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

//go:embed bootstrap
var bootstrap string

var (
	workDir      string
	rules        stringSlice
	personas     stringSlice
	tasks        stringSlice
	outputDir    = "."
	params       = make(paramMap)
	includes     = make(selectorMap)
	excludes     = make(selectorMap)
	runBootstrap bool
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rules = []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
		"CLAUDE.md",
		".cursorrules",
		".cursor/rules/",
		".instructions.md",
		".continuerules",
		".prompts/rules",
		filepath.Join(userConfigDir, "prompts", "rules"),
		"/var/local/prompts/rules",
	}

	personas = []string{
		".prompts/personas",
		filepath.Join(userConfigDir, "prompts", "personas"),
		"/var/local/prompts/personas",
	}

	tasks = []string{
		".prompts/tasks",
		filepath.Join(userConfigDir, "prompts", "tasks"),
		"/var/local/prompts/tasks",
	}

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.Var(&rules, "m", "Directory containing rules, or a single rule file. Can be specified multiple times.")
	flag.Var(&personas, "r", "Directory containing personas, or a single persona file. Can be specified multiple times.")
	flag.Var(&tasks, "t", "Directory containing tasks, or a single task file. Can be specified multiple times.")
	flag.StringVar(&outputDir, "o", ".", "Directory to write the context files to.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Var(&excludes, "S", "Exclude rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.BoolVar(&runBootstrap, "b", false, "Automatically run the bootstrap script after generating it.")

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

	// Optional persona argument after task name
	var personaName string
	if len(args) > 1 {
		personaName = args[1]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	bootstrapDir := filepath.Join(outputDir, "bootstrap.d")
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		return fmt.Errorf("failed to create bootstrap dir: %w", err)
	}

	// Track total tokens
	var totalTokens int

	// Create persona.md file
	personaOutput, err := os.Create(filepath.Join(outputDir, "persona.md"))
	if err != nil {
		return fmt.Errorf("failed to create persona file: %w", err)
	}
	defer personaOutput.Close()

	// Process persona first if provided
	if personaName != "" {
		personaFound := false
		for _, path := range personas {
			stat, err := os.Stat(path)
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				return fmt.Errorf("failed to stat persona path %s: %w", path, err)
			}
			if stat.IsDir() {
				path = filepath.Join(path, personaName+".md")
				if _, err := os.Stat(path); os.IsNotExist(err) {
					continue
				} else if err != nil {
					return fmt.Errorf("failed to stat persona file %s: %w", path, err)
				}
			}

			content, err := parseMarkdownFile(path, &struct{}{})
			if err != nil {
				return fmt.Errorf("failed to parse persona file: %w", err)
			}

			// Estimate tokens for this file
			tokens := estimateTokens(content)
			totalTokens += tokens
			fmt.Fprintf(os.Stdout, "Using persona file: %s (~%d tokens)\n", path, tokens)

			// Personas don't need variable expansion or filters
			if _, err := personaOutput.WriteString(content); err != nil {
				return fmt.Errorf("failed to write persona: %w", err)
			}

			personaFound = true
			break
		}

		if !personaFound {
			return fmt.Errorf("persona file not found for persona: %s", personaName)
		}
	}

	// Create rules.md file
	rulesOutput, err := os.Create(filepath.Join(outputDir, "rules.md"))
	if err != nil {
		return fmt.Errorf("failed to create rules file: %w", err)
	}
	defer rulesOutput.Close()

	for _, rule := range rules {

		// Skip if the path doesn't exist
		if _, err := os.Stat(rule); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(rule, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Only process .md files as rule files
			if filepath.Ext(path) != ".md" {
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
				fmt.Fprintf(os.Stdout, "Excluding rule file (does not match include selectors): %s\n", path)
				return nil
			}
			if !excludes.matchesExcludes(frontmatter) {
				fmt.Fprintf(os.Stdout, "Excluding rule file (matches exclude selectors): %s\n", path)
				return nil
			}

			// Estimate tokens for this file
			tokens := estimateTokens(content)
			totalTokens += tokens
			fmt.Fprintf(os.Stdout, "Including rule file: %s (~%d tokens)\n", path, tokens)

			// Check for a bootstrap file named <markdown-file-without-md-suffix>-bootstrap
			// For example, setup.md -> setup-bootstrap
			baseNameWithoutExt := strings.TrimSuffix(path, ".md")
			bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

			if bootstrapContent, err := os.ReadFile(bootstrapFilePath); err == nil {
				hash := sha256.Sum256(bootstrapContent)
				// Use original filename as prefix with first 4 bytes of hash as 8-char hex suffix
				// e.g., jira-bootstrap-9e2e8bc8
				baseBootstrapName := filepath.Base(bootstrapFilePath)
				bootstrapFileName := fmt.Sprintf("%s-%08x", baseBootstrapName, hash[:4])
				bootstrapPath := filepath.Join(bootstrapDir, bootstrapFileName)
				if err := os.WriteFile(bootstrapPath, bootstrapContent, 0700); err != nil {
					return fmt.Errorf("failed to write bootstrap file: %w", err)
				}
			}

			if _, err := rulesOutput.WriteString(content + "\n\n"); err != nil {
				return fmt.Errorf("failed to write to rules file: %w", err)
			}

			return nil

		})
		if err != nil {
			return fmt.Errorf("failed to walk rule dir: %w", err)
		}
	}

	if err := os.WriteFile(filepath.Join(outputDir, "bootstrap"), []byte(bootstrap), 0755); err != nil {
		return fmt.Errorf("failed to write bootstrap file: %w", err)
	}

	// Create task.md file
	taskOutput, err := os.Create(filepath.Join(outputDir, "task.md"))
	if err != nil {
		return fmt.Errorf("failed to create task file: %w", err)
	}
	defer taskOutput.Close()

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
		fmt.Fprintf(os.Stdout, "Using task file: %s (~%d tokens)\n", path, tokens)

		if _, err := taskOutput.WriteString(expanded); err != nil {
			return fmt.Errorf("failed to write expanded task: %w", err)
		}

		// Print total token count
		fmt.Fprintf(os.Stdout, "Total estimated tokens: %d\n", totalTokens)

		// Run bootstrap if requested
		if runBootstrap {
			bootstrapPath := filepath.Join(outputDir, "bootstrap")

			// Convert to absolute path
			absBootstrapPath, err := filepath.Abs(bootstrapPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for bootstrap script: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Running bootstrap script: %s\n", absBootstrapPath)

			cmd := exec.CommandContext(ctx, absBootstrapPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = outputDir

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run bootstrap script: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("prompt file not found for task: %s", taskName)
}
