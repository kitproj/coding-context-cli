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
	workDir              string
	resume               bool
	printTaskFrontmatter bool
	params               = make(paramMap)
	includes             = make(selectorMap)
	remotePaths          []string
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.BoolVar(&resume, "r", false, "Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.")
	flag.BoolVar(&printTaskFrontmatter, "t", false, "Print task frontmatter at the beginning of output.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Func("d", "Remote directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, etc.).", func(s string) error {
		remotePaths = append(remotePaths, s)
		return nil
	})

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  coding-context [options] <task-name>")
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
	if len(args) != 1 {
		return fmt.Errorf("invalid usage")
	}

	if err := os.Chdir(workDir); err != nil {
		return fmt.Errorf("failed to chdir to %s: %w", workDir, err)
	}

	// Download remote directories if specified
	var downloadedDirs []string
	defer func() {
		// Clean up downloaded directories on exit
		for _, dir := range downloadedDirs {
			cleanupRemoteDirectory(dir)
		}
	}()

	for _, remotePath := range remotePaths {
		fmt.Fprintf(os.Stderr, "⪢ Downloading remote directory: %s\n", remotePath)
		localPath, err := downloadRemoteDirectory(ctx, remotePath)
		if err != nil {
			return fmt.Errorf("failed to download remote directory %s: %w", remotePath, err)
		}
		downloadedDirs = append(downloadedDirs, localPath)
		fmt.Fprintf(os.Stderr, "⪢ Downloaded to: %s\n", localPath)
	}

	// Add task name to includes so rules can be filtered by task
	taskName := args[0]
	includes["task_name"] = taskName
	includes["resume"] = fmt.Sprint(resume)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// find the task prompt by searching for a file with matching task_name in frontmatter
	taskSearchDirs := []string{
		filepath.Join(".agents", "tasks"),
		filepath.Join(".cursor", "commands"),
		filepath.Join(homeDir, ".agents", "tasks"),
	}

	// Add downloaded remote directories to task search paths
	for _, dir := range downloadedDirs {
		taskSearchDirs = append(taskSearchDirs,
			filepath.Join(dir, ".agents", "tasks"),
			filepath.Join(dir, ".cursor", "commands"),
		)
	}

	var matchingTaskFile string
	for _, dir := range taskSearchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat task dir %s: %w", dir, err)
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Only process .md files as task files
			if filepath.Ext(path) != ".md" {
				return nil
			}

			// Parse frontmatter to check task_name
			var frontmatter frontMatter
			_, err = parseMarkdownFile(path, &frontmatter)
			if err != nil {
				return fmt.Errorf("failed to parse task file %s: %w", path, err)
			}

			// Check if task_name is present in frontmatter
			if _, hasTaskName := frontmatter["task_name"]; !hasTaskName {
				return fmt.Errorf("task file %s is missing required 'task_name' field in frontmatter", path)
			}

			// Check if file matches include selectors (task_name is already in includes)
			if !includes.matchesIncludes(frontmatter) {
				return nil
			}

			// If we already found a matching task, error on duplicate
			if matchingTaskFile != "" {
				return fmt.Errorf("multiple task files found with task_name=%s: %s and %s", taskName, matchingTaskFile, path)
			}

			matchingTaskFile = path

			return nil
		})
		if err != nil {
			return err
		}
	}

	if matchingTaskFile == "" {
		return fmt.Errorf("no task file found with task_name=%s matching selectors in frontmatter (searched in %v)", taskName, taskSearchDirs)
	}

	taskPromptPath := matchingTaskFile

	// Track total tokens
	var totalTokens int

	// Parse task file to get frontmatter and content
	var taskFrontmatter frontMatter
	taskContent, taskRawFrontmatter, err := parseMarkdownFileWithRawFrontmatter(taskPromptPath, &taskFrontmatter)
	if err != nil {
		return fmt.Errorf("failed to parse prompt file %s: %w", taskPromptPath, err)
	}

	// Print task frontmatter at the beginning if requested
	if printTaskFrontmatter {
		fmt.Println("---")
		fmt.Print(taskRawFrontmatter)
		fmt.Println("---")
		fmt.Println()
	}

	// Skip rules processing in resume mode
	if !resume {
		// Build the list of rule locations (local and remote)
		rulePaths := []string{
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
		}

		// Append remote directories to rule paths
		for _, dir := range downloadedDirs {
			rulePaths = append(rulePaths,
				filepath.Join(dir, ".agents", "rules"),
				filepath.Join(dir, ".cursor", "rules"),
				filepath.Join(dir, ".augment", "rules"),
				filepath.Join(dir, ".windsurf", "rules"),
				filepath.Join(dir, ".opencode", "agent"),
				filepath.Join(dir, ".opencode", "command"),
				filepath.Join(dir, ".github", "copilot-instructions.md"),
				filepath.Join(dir, ".gemini", "styleguide.md"),
				filepath.Join(dir, ".github", "agents"),
				filepath.Join(dir, ".augment", "guidelines.md"),
				filepath.Join(dir, "AGENTS.md"),
				filepath.Join(dir, "CLAUDE.md"),
				filepath.Join(dir, "GEMINI.md"),
				filepath.Join(dir, ".cursorrules"),
				filepath.Join(dir, ".windsurfrules"),
			)
		}

		for _, rule := range rulePaths {

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
				var frontmatter frontMatter
				content, err := parseMarkdownFile(path, &frontmatter)
				if err != nil {
					return fmt.Errorf("failed to parse markdown file: %w", err)
				}

				// Check if file matches include selectors.
				// Note: Files with duplicate basenames will both be included.
				if !includes.matchesIncludes(frontmatter) {
					fmt.Fprintf(os.Stderr, "⪢ Excluding rule file (does not match include selectors): %s\n", path)
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

					fmt.Fprintf(os.Stderr, "⪢ Running bootstrap script: %s\n", bootstrapFilePath)

					cmd := exec.CommandContext(ctx, bootstrapFilePath)
					cmd.Stdout = os.Stderr
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						return fmt.Errorf("failed to run bootstrap script: %w", err)
					}
				} else if !os.IsNotExist(err) {
					return fmt.Errorf("failed to stat bootstrap file %s: %w", bootstrapFilePath, err)
				}

				// Estimate tokens for this file
				tokens := estimateTokens(content)
				totalTokens += tokens
				fmt.Fprintf(os.Stderr, "⪢ Including rule file: %s (~%d tokens)\n", path, tokens)
				fmt.Println(content)

				return nil

			})
			if err != nil {
				return fmt.Errorf("failed to walk rule dir: %w", err)
			}
		}
	} // end of if !resume

	expanded := os.Expand(taskContent, func(key string) string {
		if val, ok := params[key]; ok {
			return val
		}
		// this might not exist, in that case, return the original text
		return fmt.Sprintf("${%s}", key)
	})

	// Estimate tokens for this file
	tokens := estimateTokens(expanded)
	totalTokens += tokens
	fmt.Fprintf(os.Stderr, "⪢ Including task file: %s (~%d tokens)\n", taskPromptPath, tokens)

	fmt.Println(expanded)

	// Print total token count
	fmt.Fprintf(os.Stderr, "⪢ Total estimated tokens: %d\n", totalTokens)

	return nil
}
