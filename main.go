package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	yaml "github.com/goccy/go-yaml"
)

type codingContext struct {
	workDir             string
	resume              bool
	params              paramMap
	includes            selectorMap
	remotePaths         []string
	emitTaskFrontmatter bool

	downloadedDirs   []string
	matchingTaskFile string
	totalTokens      int
	output           io.Writer
	logOut           io.Writer
	cmdRunner        func(cmd *exec.Cmd) error
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cc := &codingContext{
		params:   make(paramMap),
		includes: make(selectorMap),
		output:   os.Stdout,
		logOut:   flag.CommandLine.Output(),
		cmdRunner: func(cmd *exec.Cmd) error {
			return cmd.Run()
		},
	}

	flag.StringVar(&cc.workDir, "C", ".", "Change to directory before doing anything.")
	flag.BoolVar(&cc.resume, "r", false, "Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.")
	flag.BoolVar(&cc.emitTaskFrontmatter, "t", false, "Print task frontmatter at the beginning of output.")
	flag.Var(&cc.params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&cc.includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Func("d", "Remote directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, etc.).", func(s string) error {
		cc.remotePaths = append(cc.remotePaths, s)
		return nil
	})

	flag.Usage = func() {
		fmt.Fprintf(cc.logOut, "Usage:")
		fmt.Fprintln(cc.logOut)
		fmt.Fprintln(cc.logOut, "  coding-context [options] <task-name>")
		fmt.Fprintln(cc.logOut)
		fmt.Fprintln(cc.logOut, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := cc.run(ctx, flag.Args()); err != nil {
		fmt.Fprintf(cc.logOut, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func (cc *codingContext) run(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid usage")
	}

	if err := os.Chdir(cc.workDir); err != nil {
		return fmt.Errorf("failed to chdir to %s: %w", cc.workDir, err)
	}

	if err := cc.downloadRemoteDirectories(ctx); err != nil {
		return fmt.Errorf("failed to download remote directories: %w", err)
	}
	defer cc.cleanupDownloadedDirectories()

	// Add task name to includes so rules can be filtered by task
	taskName := args[0]
	cc.includes["task_name"] = taskName
	cc.includes["resume"] = fmt.Sprint(cc.resume)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	if err := cc.findTaskFile(homeDir, taskName); err != nil {
		return fmt.Errorf("failed to find task file: %w", err)
	}

	if err := cc.findExecuteRuleFiles(ctx, homeDir); err != nil {
		return fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	if err := cc.writeTaskFileContent(); err != nil {
		return fmt.Errorf("failed to write task file content: %w", err)
	}

	return nil
}

func (cc *codingContext) findTaskFile(homeDir string, taskName string) error {
	// find the task prompt by searching for a file with matching task_name in frontmatter
	taskSearchDirs := allTaskSearchPaths(homeDir)

	// Add downloaded remote directories to task search paths
	for _, dir := range cc.downloadedDirs {
		taskSearchDirs = append(taskSearchDirs, downloadedTaskSearchPaths(dir)...)
	}

	for _, dir := range taskSearchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat task dir %s: %w", dir, err)
		}

		if err := filepath.Walk(dir, cc.taskFileWalker(taskName)); err != nil {
			return err
		}
	}

	if cc.matchingTaskFile == "" {
		return fmt.Errorf("no task file found with task_name=%s matching selectors in frontmatter (searched in %v)", taskName, taskSearchDirs)
	}

	return nil
}

func (cc *codingContext) taskFileWalker(taskName string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip errors
			return err
		} else if info.IsDir() {
			// Skip directories
			return nil
		} else if filepath.Ext(path) != ".md" {
			// Only process .md files as task files
			return nil
		}

		// Parse frontmatter to check task_name
		var frontmatter frontMatter

		if _, err = parseMarkdownFile(path, &frontmatter); err != nil {
			return fmt.Errorf("failed to parse task file %s: %w", path, err)
		}

		// Check if task_name is present in frontmatter
		if _, hasTaskName := frontmatter["task_name"]; !hasTaskName {
			return fmt.Errorf("task file %s is missing required 'task_name' field in frontmatter", path)
		}

		// Check if file matches include selectors (task_name is already in includes)
		if !cc.includes.matchesIncludes(frontmatter) {
			return nil
		}

		// If we already found a matching task, error on duplicate
		if cc.matchingTaskFile != "" {
			return fmt.Errorf("multiple task files found with task_name=%s: %s and %s", taskName, cc.matchingTaskFile, path)
		}

		cc.matchingTaskFile = path

		return nil
	}
}

func (cc *codingContext) findExecuteRuleFiles(ctx context.Context, homeDir string) error {
	// Skip rule file discovery in resume mode.
	if cc.resume {
		return nil
	}

	// Build the list of rule locations (local and remote)
	rulePaths := allRulePaths(homeDir)

	// Append remote directories to rule paths
	for _, dir := range cc.downloadedDirs {
		rulePaths = append(rulePaths, downloadedRulePaths(dir)...)
	}

	for _, rule := range rulePaths {
		// Skip if the path doesn't exist
		if _, err := os.Stat(rule); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat rule path %s: %w", rule, err)
		}

		if err := filepath.Walk(rule, cc.ruleFileWalker(ctx)); err != nil {
			return fmt.Errorf("failed to walk rule dir: %w", err)
		}
	}

	return nil
}

func (cc *codingContext) ruleFileWalker(ctx context.Context) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
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

		if err := cc.runBootstrapScript(ctx, path, ext); err != nil {
			return fmt.Errorf("failed to run bootstrap script: %w", err)
		}

		// Check if file matches include selectors.
		// Note: Files with duplicate basenames will both be included.
		if !cc.includes.matchesIncludes(frontmatter) {
			fmt.Fprintf(cc.logOut, "⪢ Excluding rule file (does not match include selectors): %s\n", path)
			return nil
		}

		// Estimate tokens for this file
		tokens := estimateTokens(content)
		cc.totalTokens += tokens
		fmt.Fprintf(cc.logOut, "⪢ Including rule file: %s (~%d tokens)\n", path, tokens)
		fmt.Fprintln(cc.output, content)

		return nil
	}
}

func (cc *codingContext) runBootstrapScript(ctx context.Context, path, ext string) error {
	// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
	// For example, setup.md -> setup-bootstrap, setup.mdc -> setup-bootstrap
	baseNameWithoutExt := strings.TrimSuffix(path, ext)
	bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

	if _, err := os.Stat(bootstrapFilePath); os.IsNotExist(err) {
		// Doesn't exist, just skip.
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat bootstrap file %s: %w", bootstrapFilePath, err)
	}

	// Bootstrap file exists, make it executable and run it before printing content
	if err := os.Chmod(bootstrapFilePath, 0o755); err != nil {
		return fmt.Errorf("failed to chmod bootstrap file %s: %w", bootstrapFilePath, err)
	}

	fmt.Fprintf(cc.logOut, "⪢ Running bootstrap script: %s\n", bootstrapFilePath)

	cmd := exec.CommandContext(ctx, bootstrapFilePath)
	cmd.Stdout = cc.logOut
	cmd.Stderr = cc.logOut

	if err := cc.cmdRunner(cmd); err != nil {
		return fmt.Errorf("failed to run bootstrap script: %w", err)
	}

	return nil
}

func (cc *codingContext) writeTaskFileContent() error {
	taskMatter := make(map[string]any)

	content, err := parseMarkdownFile(cc.matchingTaskFile, &taskMatter)
	if err != nil {
		return fmt.Errorf("failed to parse prompt file %s: %w", cc.matchingTaskFile, err)
	}

	expanded := os.Expand(content, func(key string) string {
		if val, ok := cc.params[key]; ok {
			return val
		}
		// this might not exist, in that case, return the original text
		return fmt.Sprintf("${%s}", key)
	})

	// Estimate tokens for this file
	tokens := estimateTokens(expanded)
	cc.totalTokens += tokens
	fmt.Fprintf(cc.logOut, "⪢ Including task file: %s (~%d tokens)\n", cc.matchingTaskFile, tokens)

	if cc.emitTaskFrontmatter {
		fmt.Fprintln(cc.output, "---")
		if err := yaml.NewEncoder(cc.output).Encode(taskMatter); err != nil {
			return fmt.Errorf("failed to encode task matter: %w", err)
		}
		fmt.Fprintln(cc.output, "---")
	}

	fmt.Fprintln(cc.output, expanded)

	// Print total token count
	fmt.Fprintf(cc.logOut, "⪢ Total estimated tokens: %d\n", cc.totalTokens)

	return nil
}
