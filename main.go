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
	taskFrontmatter  frontMatter // Parsed task frontmatter
	taskContent      string      // Parsed task content (before parameter expansion)
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
	cc.includes.SetValue("task_name", taskName)
	cc.includes.SetValue("resume", fmt.Sprint(cc.resume))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	if err := cc.findTaskFile(homeDir, taskName); err != nil {
		return fmt.Errorf("failed to find task file: %w", err)
	}

	// Parse task file early to extract selector labels for filtering rules and tools
	if err := cc.parseTaskFile(); err != nil {
		return fmt.Errorf("failed to parse task file: %w", err)
	}

	if err := cc.printTaskFrontmatter(); err != nil {
		return fmt.Errorf("failed to emit task frontmatter: %w", err)
	}

	if err := cc.findExecuteRuleFiles(ctx, homeDir); err != nil {
		return fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	// Run bootstrap script for the task file if it exists
	taskExt := filepath.Ext(cc.matchingTaskFile)
	if err := cc.runBootstrapScript(ctx, cc.matchingTaskFile, taskExt); err != nil {
		return fmt.Errorf("failed to run task bootstrap script: %w", err)
	}

	if err := cc.emitTaskFileContent(); err != nil {
		return fmt.Errorf("failed to emit task file content: %w", err)
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
		// If not present, skip this file (it's not a task file)
		if _, hasTaskName := frontmatter["task_name"]; !hasTaskName {
			return nil
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

		// Check if file matches include selectors BEFORE running bootstrap script.
		// Note: Files with duplicate basenames will both be included.
		if !cc.includes.matchesIncludes(frontmatter) {
			fmt.Fprintf(cc.logOut, "⪢ Excluding rule file (does not match include selectors): %s\n", path)
			return nil
		}

		if err := cc.runBootstrapScript(ctx, path, ext); err != nil {
			return fmt.Errorf("failed to run bootstrap script (path: %s): %w", path, err)
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
		return err
	}

	return nil
}

// parseTaskFile parses the task file and extracts selector labels from frontmatter.
// The selectors are added to cc.includes for filtering rules and tools.
// The parsed frontmatter and content are stored in cc.taskFrontmatter and cc.taskContent.
func (cc *codingContext) parseTaskFile() error {
	cc.taskFrontmatter = make(frontMatter)

	content, err := parseMarkdownFile(cc.matchingTaskFile, &cc.taskFrontmatter)
	if err != nil {
		return fmt.Errorf("failed to parse task file %s: %w", cc.matchingTaskFile, err)
	}

	cc.taskContent = content

	// Extract selector labels from frontmatter
	// Look for a "selectors" field that contains a map of key-value pairs
	// Values can be strings or arrays (for OR logic)
	if selectorsRaw, ok := cc.taskFrontmatter["selectors"]; ok {
		selectorsMap, ok := selectorsRaw.(map[string]any)
		if !ok {
			// Try to handle it as a map[interface{}]interface{} (common YAML unmarshal result)
			if selectorsMapAny, ok := selectorsRaw.(map[any]any); ok {
				selectorsMap = make(map[string]any)
				for k, v := range selectorsMapAny {
					selectorsMap[fmt.Sprint(k)] = v
				}
			} else {
				return fmt.Errorf("task file %s has invalid 'selectors' field: expected map, got %T", cc.matchingTaskFile, selectorsRaw)
			}
		}

		// Add selectors to includes
		// Convert all values to map[string]bool for OR logic
		for key, value := range selectorsMap {
			switch v := value.(type) {
			case []any:
				// Convert []any to map[string]bool
				for _, item := range v {
					cc.includes.SetValue(key, fmt.Sprint(item))
				}
			case string:
				// Convert string to single value in map
				cc.includes.SetValue(key, v)
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
				// Convert scalar numeric or boolean to string
				cc.includes.SetValue(key, fmt.Sprint(v))
			default:
				return fmt.Errorf("task file %s has invalid selector value for key %q: expected string or array, got %T", cc.matchingTaskFile, key, value)
			}
		}
	}

	return nil
}

// printTaskFrontmatter emits only the task frontmatter to the output and
// only if emitTaskFrontmatter is true.
func (cc *codingContext) printTaskFrontmatter() error {
	if !cc.emitTaskFrontmatter {
		return nil
	}

	fmt.Fprintln(cc.output, "---")
	if err := yaml.NewEncoder(cc.output).Encode(cc.taskFrontmatter); err != nil {
		return fmt.Errorf("failed to encode task frontmatter: %w", err)
	}
	fmt.Fprintln(cc.output, "---")
	return nil
}

// emitTaskFileContent emits the parsed task content to the output.
// It expands parameters and estimates tokens.
func (cc *codingContext) emitTaskFileContent() error {
	expanded := os.Expand(cc.taskContent, func(key string) string {
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

	fmt.Fprintln(cc.output, expanded)

	// Print total token count
	fmt.Fprintf(cc.logOut, "⪢ Total estimated tokens: %d\n", cc.totalTokens)

	return nil
}
