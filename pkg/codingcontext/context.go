package codingcontext

import (
	"bufio"
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
	manifestURL      string
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

// WithManifestURL sets the manifest URL
func WithManifestURL(manifestURL string) Option {
	return func(c *Context) {
		c.manifestURL = manifestURL
	}
}

// WithSearchPaths adds one or more search paths
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

// extractSelectors extracts selector labels from frontmatter and adds them to cc.includes
func (cc *Context) extractSelectors(frontMatter *TaskFrontMatter) {
	for key, value := range frontMatter.Selectors {
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

// Run executes the context assembly for the given taskPrompt and returns the assembled result.
// The taskPrompt can be either:
// - A free-text prompt (used directly as task content)
// - A prompt containing a slash command (e.g., "/fix-bug 123") which triggers task lookup
func (cc *Context) Run(ctx context.Context, taskPrompt string) (*Result, error) {
	// Parse manifest file first to get additional search paths
	manifestPaths, err := cc.parseManifestFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest file: %w", err)
	}
	cc.searchPaths = append(cc.searchPaths, manifestPaths...)

	// Download all remote directories (including those from manifest)
	if err := cc.downloadRemoteDirectories(ctx); err != nil {
		return nil, fmt.Errorf("failed to download remote directories: %w", err)
	}
	defer cc.cleanupDownloadedDirectories()

	// If resume mode is enabled, add resume=true as a selector
	if cc.resume {
		cc.includes.SetValue("resume", "true")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Expand parameters in taskPrompt to allow slash commands in parameters
	expandedTaskPrompt := cc.expandParams(taskPrompt)

	// Parse the task prompt into blocks using the new parser
	taskBlocks, err := ParseTask(expandedTaskPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task prompt: %w", err)
	}

	// Process all blocks to assemble the final task content and frontmatter
	var taskContent strings.Builder
	cc.task = Markdown[TaskFrontMatter]{
		FrontMatter: TaskFrontMatter{},
	}
	var taskFiles []string                // Track all task files used
	var firstFrontmatter *TaskFrontMatter // Track first task's frontmatter

	for i, block := range taskBlocks {
		if block.Text != nil {
			// Text block - append directly to content
			taskContent.WriteString(block.Text.Content())
		} else if block.SlashCommand != nil {
			// Slash command block - find and expand the task file
			cc.logger.Info("Processing slash command", "task", block.SlashCommand.Name, "block", i)

			// Extract parameters from the slash command using the new Params() method
			slashParams := block.SlashCommand.Params()

			// Merge slash command parameters with existing parameters
			// Slash command parameters take precedence
			for k, v := range slashParams {
				cc.params[k] = v
			}

			// Add task name to includes so rules can be filtered by task
			cc.includes.SetValue("task_name", block.SlashCommand.Name)

			// Reset matchingTaskFile before finding the next one
			cc.matchingTaskFile = ""

			// Find the task file
			if err := cc.findTaskFile(block.SlashCommand.Name); err != nil {
				return nil, fmt.Errorf("failed to find task file for command %q: %w", block.SlashCommand.Name, err)
			}

			taskFiles = append(taskFiles, cc.matchingTaskFile)

			// Parse the task file
			var frontMatter TaskFrontMatter
			taskMarkdown, err := ParseMarkdownFile(cc.matchingTaskFile, &frontMatter)
			if err != nil {
				return nil, fmt.Errorf("failed to parse task file %s: %w", cc.matchingTaskFile, err)
			}

			// Use the first task's frontmatter only (as per new requirement)
			if firstFrontmatter == nil {
				firstFrontmatter = &frontMatter
				cc.task.FrontMatter = frontMatter
			}

			// Expand parameters in the task content
			expandedContent := cc.expandParams(taskMarkdown.Content)
			taskContent.WriteString(expandedContent)

			// Extract selector labels from frontmatter
			cc.extractSelectors(&frontMatter)
		}
	}

	// Set the assembled content
	cc.task.Content = taskContent.String()

	// Set a composite task file name if multiple tasks were used
	if len(taskFiles) == 0 {
		cc.matchingTaskFile = "<inline>"
	} else if len(taskFiles) == 1 {
		cc.matchingTaskFile = taskFiles[0]
	} else {
		cc.matchingTaskFile = fmt.Sprintf("<composite: %d tasks>", len(taskFiles))
	}

	// Task frontmatter agent field overrides -a flag if -a was not set
	if cc.task.FrontMatter.Agent != "" && !cc.agent.IsSet() {
		if agent, err := ParseAgent(cc.task.FrontMatter.Agent); err == nil {
			cc.agent = agent
		} else {
			cc.logger.Warn("Invalid agent name in task frontmatter, ignoring", "agent", cc.task.FrontMatter.Agent, "error", err)
		}
	}

	if err := cc.findExecuteRuleFiles(ctx, homeDir); err != nil {
		return nil, fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	// Expand parameters in task content
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

// parseManifestFile downloads a manifest file from a Go Getter URL and returns
// the list of search paths (one per line). Every line is included as-is without trimming.
func (cc *Context) parseManifestFile(ctx context.Context) ([]string, error) {
	if cc.manifestURL == "" {
		return nil, nil
	}

	manifestFile := downloadDir(cc.manifestURL)

	// Download the manifest file using go-getter's GetFile function
	// GetFile is specifically for downloading single files (not directories)
	if _, err := getter.GetFile(ctx, manifestFile, cc.manifestURL); err != nil {
		return nil, fmt.Errorf("failed to download manifest file %s: %w", cc.manifestURL, err)
	}
	defer os.RemoveAll(manifestFile)

	cc.logger.Info("Downloaded manifest file", "path", manifestFile)

	// Read and parse the manifest file
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	var paths []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		paths = append(paths, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	cc.logger.Info("Parsed manifest file", "url", cc.manifestURL, "paths", len(paths))

	return paths, nil
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

func (cc *Context) findTaskFile(taskName string) error {

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
	cc.extractSelectors(&cc.task.FrontMatter)

	return nil
}
