package codingcontext

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Context holds the configuration and state for assembling coding context
type Context struct {
	workDir             string
	resume              bool
	params              Params
	includes            Selectors
	remotePaths         []string
	emitTaskFrontmatter bool

	downloadedDirs   []string
	matchingTaskFile string
	taskFrontmatter  FrontMatter // Parsed task frontmatter
	taskContent      string      // Parsed task content (before parameter expansion)
	rules            []Markdown  // Collected rule files
	totalTokens      int
	logger           *slog.Logger
	cmdRunner        func(cmd *exec.Cmd) error
}

// Option is a functional option for configuring a Context
type Option func(*Context)

// WithWorkDir sets the working directory
func WithWorkDir(dir string) Option {
	return func(c *Context) {
		c.workDir = dir
	}
}

// WithResume enables resume mode
func WithResume(resume bool) Option {
	return func(c *Context) {
		c.resume = resume
	}
}

// WithParams sets the parameters
func WithParams(params Params) Option {
	return func(c *Context) {
		c.params = params
	}
}

// WithSelectors sets the selectors
func WithSelectors(selectors Selectors) Option {
	return func(c *Context) {
		c.includes = selectors
	}
}

// WithRemotePaths sets the remote paths
func WithRemotePaths(paths []string) Option {
	return func(c *Context) {
		c.remotePaths = paths
	}
}

// WithEmitTaskFrontmatter enables task frontmatter emission
func WithEmitTaskFrontmatter(emit bool) Option {
	return func(c *Context) {
		c.emitTaskFrontmatter = emit
	}
}

// WithLogger sets the logger
func WithLogger(logger *slog.Logger) Option {
	return func(c *Context) {
		c.logger = logger
	}
}

// New creates a new Context with the given options
func New(opts ...Option) *Context {
	c := &Context{
		workDir:  ".",
		params:   make(Params),
		includes: make(Selectors),
		rules:    make([]Markdown, 0),
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		cmdRunner: func(cmd *exec.Cmd) error {
			return cmd.Run()
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// expandParams expands parameter placeholders in the given content
func (cc *Context) expandParams(content string) string {
	return os.Expand(content, func(key string) string {
		if val, ok := cc.params[key]; ok {
			return val
		}
		// this might not exist, in that case, return the original text
		return fmt.Sprintf("${%s}", key)
	})
}

// Run executes the context assembly for the given task name and returns the assembled result
func (cc *Context) Run(ctx context.Context, taskName string) (*Result, error) {
	if err := cc.downloadRemoteDirectories(ctx); err != nil {
		return nil, fmt.Errorf("failed to download remote directories: %w", err)
	}
	defer cc.cleanupDownloadedDirectories()

	// Add task name to includes so rules can be filtered by task
	cc.includes.SetValue("task_name", taskName)
	cc.includes.SetValue("resume", fmt.Sprint(cc.resume))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	if err := cc.findTaskFile(homeDir, taskName); err != nil {
		return nil, fmt.Errorf("failed to find task file: %w", err)
	}

	// Parse task file early to extract selector labels for filtering rules and tools
	if err := cc.parseTaskFile(); err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}

	// Expand parameters in task content to allow slash commands in parameters
	expandedContent := cc.expandParams(cc.taskContent)

	// Check if the task contains a slash command (after parameter expansion)
	slashTaskName, slashParams, found, err := parseSlashCommand(expandedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse slash command in task: %w", err)
	}
	if found {
		cc.logger.Info("Found slash command in task", "task", slashTaskName, "params", slashParams)

		// Replace parameters completely with slash command parameters
		// The slash command fully replaces both task name and parameters
		cc.params = slashParams

		// Always find and parse the slash command task file, even if it's the same task name
		// This ensures fresh parsing with the new parameters
		if slashTaskName == taskName {
			cc.logger.Info("Reloading slash command task", "task", slashTaskName)
		} else {
			cc.logger.Info("Switching to slash command task", "from", taskName, "to", slashTaskName)
		}

		// Reset task-related state
		cc.matchingTaskFile = ""
		cc.taskFrontmatter = nil
		cc.taskContent = ""

		// Update task_name in includes
		cc.includes.SetValue("task_name", slashTaskName)

		// Find the new task file
		if err := cc.findTaskFile(homeDir, slashTaskName); err != nil {
			return nil, fmt.Errorf("failed to find slash command task file: %w", err)
		}

		// Parse the new task file
		if err := cc.parseTaskFile(); err != nil {
			return nil, fmt.Errorf("failed to parse slash command task file: %w", err)
		}
	}

	if err := cc.findExecuteRuleFiles(ctx, homeDir); err != nil {
		return nil, fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	// Expand parameters in task content (note: this may be a different task than initially loaded
	// if a slash command was found above, which loaded a new task with new parameters)
	expandedTask := cc.expandParams(cc.taskContent)

	// Estimate tokens for task file
	taskTokens := estimateTokens(expandedTask)
	cc.totalTokens += taskTokens
	cc.logger.Info("Including task file", "path", cc.matchingTaskFile, "tokens", taskTokens)
	cc.logger.Info("Total estimated tokens", "tokens", cc.totalTokens)

	// Build and return the result
	result := &Result{
		Rules: cc.rules,
		Task: Markdown{
			Path:        cc.matchingTaskFile,
			FrontMatter: cc.taskFrontmatter,
			Content:     expandedTask,
			Tokens:      taskTokens,
		},
	}

	return result, nil
}

func (cc *Context) downloadRemoteDirectories(ctx context.Context) error {
	for _, remotePath := range cc.remotePaths {
		cc.logger.Info("Downloading remote directory", "path", remotePath)
		localPath, err := downloadRemoteDirectory(ctx, remotePath)
		if err != nil {
			return fmt.Errorf("failed to download remote directory %s: %w", remotePath, err)
		}
		cc.downloadedDirs = append(cc.downloadedDirs, localPath)
		cc.logger.Info("Downloaded to", "path", localPath)
	}

	return nil
}

func (cc *Context) cleanupDownloadedDirectories() {
	for _, dir := range cc.downloadedDirs {
		if dir == "" {
			continue
		}

		if err := os.RemoveAll(dir); err != nil {
			cc.logger.Error("Error cleaning up downloaded directory", "path", dir, "error", err)
		}
	}
}

func (cc *Context) findTaskFile(homeDir string, taskName string) error {
	// find the task prompt by searching for a file with matching task_name in frontmatter
	taskSearchDirs := AllTaskSearchPaths(cc.workDir, homeDir)

	// Add downloaded remote directories to task search paths
	for _, dir := range cc.downloadedDirs {
		taskSearchDirs = append(taskSearchDirs, DownloadedTaskSearchPaths(dir)...)
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

func (cc *Context) taskFileWalker(taskName string) func(path string, info os.FileInfo, err error) error {
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
		var frontmatter FrontMatter

		if _, err = ParseMarkdownFile(path, &frontmatter); err != nil {
			return fmt.Errorf("failed to parse task file %s: %w", path, err)
		}

		// Get task_name from frontmatter, or use filename without .md extension
		fileTaskName, hasTaskName := frontmatter["task_name"]
		var taskNameStr string
		if hasTaskName {
			taskNameStr = fmt.Sprint(fileTaskName)
		} else {
			// Use filename without .md extension as task name
			taskNameStr = strings.TrimSuffix(filepath.Base(path), ".md")
		}

		// Check if this file's task name matches the requested task name
		if taskNameStr != taskName {
			return nil
		}

		// Check if file matches include selectors (task_name is already in includes)
		if !cc.includes.MatchesIncludes(frontmatter) {
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

func (cc *Context) findExecuteRuleFiles(ctx context.Context, homeDir string) error {
	// Skip rule file discovery in resume mode.
	if cc.resume {
		return nil
	}

	// Build the list of rule locations (local and remote)
	rulePaths := AllRulePaths(cc.workDir, homeDir)

	// Append remote directories to rule paths
	for _, dir := range cc.downloadedDirs {
		rulePaths = append(rulePaths, DownloadedRulePaths(dir)...)
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

func (cc *Context) ruleFileWalker(ctx context.Context) func(path string, info os.FileInfo, err error) error {
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
		var frontmatter FrontMatter
		content, err := ParseMarkdownFile(path, &frontmatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file: %w", err)
		}

		// Check if file matches include selectors BEFORE running bootstrap script.
		// Note: Files with duplicate basenames will both be included.
		if !cc.includes.MatchesIncludes(frontmatter) {
			cc.logger.Info("Excluding rule file (does not match include selectors)", "path", path)
			return nil
		}

		if err := cc.runBootstrapScript(ctx, path, ext); err != nil {
			return fmt.Errorf("failed to run bootstrap script (path: %s): %w", path, err)
		}

		// Expand parameters in rule content
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
		cc.logger.Info("Including rule file", "path", path, "tokens", tokens)

		// Collect the rule content with frontmatter
		cc.rules = append(cc.rules, Markdown{
			Path:        path,
			FrontMatter: frontmatter,
			Content:     expanded,
			Tokens:      tokens,
		})

		return nil
	}
}

func (cc *Context) runBootstrapScript(ctx context.Context, path, ext string) error {
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

	cc.logger.Info("Running bootstrap script", "path", bootstrapFilePath)

	cmd := exec.CommandContext(ctx, bootstrapFilePath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cc.cmdRunner(cmd); err != nil {
		return err
	}

	return nil
}

// parseTaskFile parses the task file and extracts selector labels from frontmatter.
// The selectors are added to cc.includes for filtering rules and tools.
// The parsed frontmatter and content are stored in cc.taskFrontmatter and cc.taskContent.
func (cc *Context) parseTaskFile() error {
	cc.taskFrontmatter = make(FrontMatter)

	content, err := ParseMarkdownFile(cc.matchingTaskFile, &cc.taskFrontmatter)
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
