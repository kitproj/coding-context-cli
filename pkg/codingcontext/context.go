package codingcontext

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-getter/v2"
)

// Context holds the configuration and state for assembling coding context
type Context struct {
	params           Params
	includes         Selectors
	searchPaths      []string
	matchingTaskFile string
	task             Markdown[TaskFrontMatter]   // Parsed task
	rules            []Markdown[RuleFrontMatter] // Collected rule files
	totalTokens      int
	logger           *slog.Logger
	cmdRunner        func(cmd *exec.Cmd) error
	resume           bool
	agent            Agent
}

// Option is a functional option for configuring a Context
type Option func(*Context)

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

// WithSearchPath sets a search path
func WithSearchPaths(paths ...string) Option {
	return func(c *Context) {
		c.searchPaths = append(c.searchPaths, paths...)
	}
}

// WithLogger sets the logger
func WithLogger(logger *slog.Logger) Option {
	return func(c *Context) {
		c.logger = logger
	}
}

// WithResume enables resume mode, which skips rule discovery and bootstrap scripts
func WithResume(resume bool) Option {
	return func(c *Context) {
		c.resume = resume
	}
}

// WithAgent sets the target agent, which excludes that agent's own rules
func WithAgent(agent Agent) Option {
	return func(c *Context) {
		c.agent = agent
	}
}

// New creates a new Context with the given options
func New(opts ...Option) *Context {
	c := &Context{
		params:   make(Params),
		includes: make(Selectors),
		rules:    make([]Markdown[RuleFrontMatter], 0),
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

	// If resume mode is enabled, add resume=true as a selector
	if cc.resume {
		cc.includes.SetValue("resume", "true")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	err = cc.findTaskFile(homeDir, taskName)
	if err != nil {
		return nil, fmt.Errorf("failed to find task file: %w", err)
	}

	// Parse task file early to extract selector labels for filtering rules and tools
	if err := cc.parseTaskFile(); err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}

	// Task frontmatter agent field overrides -a flag if -a was not set
	if cc.task.FrontMatter.Agent != "" && !cc.agent.IsSet() {
		if agent, err := ParseAgent(cc.task.FrontMatter.Agent); err == nil {
			cc.agent = agent
		} else {
			cc.logger.Warn("Invalid agent name in task frontmatter, ignoring", "agent", cc.task.FrontMatter.Agent, "error", err)
		}
	}

	// Expand parameters in task content to allow slash commands in parameters
	expandedContent := cc.expandParams(cc.task.Content)

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
		cc.task = Markdown[TaskFrontMatter]{}

		// Update task_name in includes for rule filtering
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
	expandedTask := cc.expandParams(cc.task.Content)

	// Estimate tokens for task file
	taskTokens := estimateTokens(expandedTask)
	cc.totalTokens += taskTokens
	cc.logger.Info("Including task file", "path", cc.matchingTaskFile, "tokens", taskTokens)
	cc.logger.Info("Total estimated tokens", "tokens", cc.totalTokens)

	// Build and return the result
	cc.task.Content = expandedTask
	cc.task.Tokens = taskTokens
	result := &Result{
		Rules: cc.rules,
		Task:  cc.task,
	}

	return result, nil
}

func downloadDir(path string) string {
	// hash the path and prepend it with a temporary directory
	hash := sha256.Sum256([]byte(path))
	tempDir := os.TempDir()
	return filepath.Join(tempDir, fmt.Sprintf("%x", hash))
}

func (cc *Context) downloadRemoteDirectories(ctx context.Context) error {
	for _, path := range cc.searchPaths {
		cc.logger.Info("Downloading remote directory", "path", path)
		dst := downloadDir(path)
		if _, err := getter.Get(ctx, dst, path); err != nil {
			return fmt.Errorf("failed to download remote directory %s: %w", path, err)
		}
		cc.logger.Info("Downloaded to", "path", dst)
	}

	return nil
}

func (cc *Context) cleanupDownloadedDirectories() {
	for _, path := range cc.searchPaths {
		dst := downloadDir(path)
		if err := os.RemoveAll(dst); err != nil {
			cc.logger.Error("Error cleaning up downloaded directory", "path", dst, "error", err)
		}
	}
}

func (cc *Context) findTaskFile(homeDir string, taskName string) error {

	var taskSearchDirs []string
	// Add downloaded remote directories to task search paths
	for _, path := range cc.searchPaths {
		dst := downloadDir(path)
		subPaths := taskSearchPaths(dst)
		taskSearchDirs = append(taskSearchDirs, subPaths...)
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
		return fmt.Errorf("no task file found with filename=%s.md matching selectors in frontmatter (searched in %v)", taskName, taskSearchDirs)
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

		// Match by filename (without .md extension)
		baseName := strings.TrimSuffix(filepath.Base(path), ".md")
		if baseName != taskName {
			return nil
		}

		// Parse frontmatter to check selectors
		var frontmatter TaskFrontMatter
		if _, err = ParseMarkdownFile(path, &frontmatter); err != nil {
			return fmt.Errorf("failed to parse task file %s: %w", path, err)
		}

		// Check if file matches include selectors
		if !cc.includes.MatchesIncludes(frontmatter.BaseFrontMatter) {
			return nil
		}

		// If we already found a matching task, error on duplicate
		if cc.matchingTaskFile != "" {
			return fmt.Errorf("multiple task files found with filename=%s.md: %s and %s", taskName, cc.matchingTaskFile, path)
		}

		cc.matchingTaskFile = path

		return nil
	}
}

func (cc *Context) findExecuteRuleFiles(ctx context.Context, homeDir string) error {
	// Skip rule file discovery if resume mode is enabled
	// Check cc.resume directly first, then fall back to selector check for backward compatibility
	if cc.resume || (cc.includes != nil && cc.includes.GetValue("resume", "true")) {
		return nil
	}

	var ruleSearchDirs []string
	for _, path := range cc.searchPaths {
		dst := downloadDir(path)
		subPaths := rulePaths(dst, path == homeDir)
		ruleSearchDirs = append(ruleSearchDirs, subPaths...)
	}

	// Build the list of rule locations (local and remote)
	for _, rule := range ruleSearchDirs {
		// Skip if this path should be excluded based on target agent
		if cc.agent.ShouldExcludePath(rule) {
			cc.logger.Info("Excluding rule path (target agent filtering)", "path", rule)
			continue
		}

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

		// Skip if this file path should be excluded based on target agent
		if cc.agent.ShouldExcludePath(path) {
			cc.logger.Info("Excluding rule file (target agent filtering)", "path", path)
			return nil
		}

		// Parse frontmatter to check selectors
		var frontmatter RuleFrontMatter
		content, err := ParseMarkdownFile(path, &frontmatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file: %w", err)
		}

		// Exclude rules whose frontmatter agent field matches the target agent
		if cc.agent != "" && frontmatter.Agent != "" {
			if string(cc.agent) == frontmatter.Agent {
				cc.logger.Info("Excluding rule file (agent field matches target agent)", "path", path, "agent", frontmatter.Agent)
				return nil
			}
		}

		// Check if file matches include selectors BEFORE running bootstrap script.
		// Note: Files with duplicate basenames will both be included.
		if !cc.includes.MatchesIncludes(frontmatter.BaseFrontMatter) {
			cc.logger.Info("Excluding rule file (does not match include selectors)", "path", path)
			return nil
		}

		if err := cc.runBootstrapScript(ctx, path, ext); err != nil {
			return fmt.Errorf("failed to run bootstrap script (path: %s): %w", path, err)
		}

		// Expand parameters in rule content
		expanded := os.Expand(content.Content, func(key string) string {
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
		cc.rules = append(cc.rules, Markdown[RuleFrontMatter]{
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

	if cc.cmdRunner != nil {
		if err := cc.cmdRunner(cmd); err != nil {
			return err
		}
	} else {
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// parseTaskFile parses the task file and extracts selector labels from frontmatter.
// The selectors are added to cc.includes for filtering rules and tools.
// The parsed task is stored in cc.task.
func (cc *Context) parseTaskFile() error {
	var frontMatter TaskFrontMatter

	task, err := ParseMarkdownFile(cc.matchingTaskFile, &frontMatter)
	if err != nil {
		return fmt.Errorf("failed to parse task file %s: %w", cc.matchingTaskFile, err)
	}

	cc.task = task

	// Extract selector labels from frontmatter
	// Look for a "selectors" field that contains a map of key-value pairs
	// Values can be strings or arrays (for OR logic)
	for key, value := range cc.task.FrontMatter.Selectors {
		switch v := value.(type) {
		case []any:
			// Convert []any to multiple selector values for OR logic
			for _, item := range v {
				cc.includes.SetValue(key, fmt.Sprint(item))
			}
		default:
			cc.includes.SetValue(key, fmt.Sprint(v))
		}
	}

	return nil
}
