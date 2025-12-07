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
	params      Params
	includes    Selectors
	manifestURL string
	searchPaths []string
	task        Markdown[TaskFrontMatter]   // Parsed task
	rules       []Markdown[RuleFrontMatter] // Collected rule files
	totalTokens int
	logger      *slog.Logger
	cmdRunner   func(cmd *exec.Cmd) error
	resume      bool
	agent       Agent
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

// getMarkdown searches for a markdown file with the given name in the provided search paths
// and returns the parsed markdown with parameters substituted
func (cc *Context) getMarkdown(searchPaths []string, name string, params map[string]string) (Markdown[TaskFrontMatter], error) {
	var matchingFile string

	// Search for the markdown file
	for _, dir := range searchPaths {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return Markdown[TaskFrontMatter]{}, fmt.Errorf("failed to stat dir %s: %w", dir, err)
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".md" {
				return nil
			}

			// Match by filename (without .md extension)
			baseName := strings.TrimSuffix(filepath.Base(path), ".md")
			if baseName != name {
				return nil
			}

			// Parse frontmatter to check selectors
			var frontmatter TaskFrontMatter
			if _, err = ParseMarkdownFile(path, &frontmatter); err != nil {
				return fmt.Errorf("failed to parse markdown file %s: %w", path, err)
			}

			// Check if file matches include selectors
			if !cc.includes.MatchesIncludes(frontmatter.BaseFrontMatter) {
				return nil
			}

			// If we already found a matching file, error on duplicate
			if matchingFile != "" {
				return fmt.Errorf("multiple markdown files found with filename=%s.md: %s and %s", name, matchingFile, path)
			}

			matchingFile = path
			return nil
		})

		if err != nil {
			return Markdown[TaskFrontMatter]{}, err
		}
	}

	if matchingFile == "" {
		return Markdown[TaskFrontMatter]{}, fmt.Errorf("no markdown file found with filename=%s.md matching selectors (searched in %v)", name, searchPaths)
	}

	// Parse the markdown file
	var frontMatter TaskFrontMatter
	markdown, err := ParseMarkdownFile(matchingFile, &frontMatter)
	if err != nil {
		return Markdown[TaskFrontMatter]{}, fmt.Errorf("failed to parse markdown file %s: %w", matchingFile, err)
	}

	// Substitute parameters in the content
	expandedContent := os.Expand(markdown.Content, func(key string) string {
		if val, ok := params[key]; ok {
			return val
		}
		// If not in params map, check cc.params
		if val, ok := cc.params[key]; ok {
			return val
		}
		// Return original placeholder if not found
		return fmt.Sprintf("${%s}", key)
	})

	markdown.Content = expandedContent
	return markdown, nil
}

// getTask searches for a task markdown file and returns it with parameters substituted
func (cc *Context) getTask(taskName string, params map[string]string) (Markdown[TaskFrontMatter], error) {
	var taskSearchDirs []string
	// Add downloaded remote directories to task search paths
	for _, path := range cc.searchPaths {
		dst := downloadDir(path)
		subPaths := taskSearchPaths(dst)
		taskSearchDirs = append(taskSearchDirs, subPaths...)
	}

	return cc.getMarkdown(taskSearchDirs, taskName, params)
}

// getCommand searches for a command markdown file and returns it with parameters substituted
func (cc *Context) getCommand(commandName string, params map[string]string) (Markdown[TaskFrontMatter], error) {
	var commandSearchDirs []string
	// Add downloaded remote directories to command search paths
	for _, path := range cc.searchPaths {
		dst := downloadDir(path)
		subPaths := commandSearchPaths(dst)
		commandSearchDirs = append(commandSearchDirs, subPaths...)
	}

	return cc.getMarkdown(commandSearchDirs, commandName, params)
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
// The taskPrompt should be a task name (e.g., "fix-bug") that will be looked up in task paths.
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

	// Get the task using getTask (this searches task paths and substitutes parameters)
	taskMarkdown, err := cc.getTask(taskPrompt, cc.params)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Set the task frontmatter
	cc.task.FrontMatter = taskMarkdown.FrontMatter

	// Add task name to includes so rules can be filtered
	cc.includes.SetValue("task_name", taskPrompt)

	// Extract selector labels from task frontmatter
	for key, value := range taskMarkdown.FrontMatter.Selectors {
		switch v := value.(type) {
		case []any:
			for _, item := range v {
				cc.includes.SetValue(key, fmt.Sprint(item))
			}
		default:
			cc.includes.SetValue(key, fmt.Sprint(v))
		}
	}

	// Task frontmatter agent field overrides -a flag if -a was not set
	if cc.task.FrontMatter.Agent != "" && !cc.agent.IsSet() {
		if agent, err := ParseAgent(cc.task.FrontMatter.Agent); err == nil {
			cc.agent = agent
		} else {
			cc.logger.Warn("Invalid agent name in task frontmatter, ignoring", "agent", cc.task.FrontMatter.Agent, "error", err)
		}
	}

	// Parse the task markdown content into blocks
	taskBlocks, err := ParseTask(taskMarkdown.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task content into blocks: %w", err)
	}

	// Iterate over blocks and replace commands with their content
	var finalContent strings.Builder
	for i, block := range taskBlocks {
		if block.Text != nil {
			// Text block - append directly
			finalContent.WriteString(block.Text.Content())
		} else if block.SlashCommand != nil {
			// Command block - get the command and append its content
			cc.logger.Info("Processing command", "name", block.SlashCommand.Name, "block", i)

			// Extract parameters from the slash command
			cmdParams := block.SlashCommand.Params()

			// Get the command markdown
			commandMarkdown, err := cc.getCommand(block.SlashCommand.Name, cmdParams)
			if err != nil {
				return nil, fmt.Errorf("failed to get command %q: %w", block.SlashCommand.Name, err)
			}

			// Append the command content
			finalContent.WriteString(commandMarkdown.Content)
		}
	}

	// Set the final task content
	cc.task.Content = finalContent.String()

	if err := cc.findExecuteRuleFiles(ctx, homeDir); err != nil {
		return nil, fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	// Estimate tokens for task
	taskTokens := estimateTokens(cc.task.Content)
	cc.totalTokens += taskTokens
	cc.logger.Info("Including task", "name", taskPrompt, "tokens", taskTokens)
	cc.logger.Info("Total estimated tokens", "tokens", cc.totalTokens)

	// Build and return the result
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
