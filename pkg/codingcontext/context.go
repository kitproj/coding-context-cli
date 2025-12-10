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
	params          Params
	includes        Selectors
	manifestURL     string
	searchPaths     []string
	downloadedPaths []string
	task            Markdown[TaskFrontMatter]   // Parsed task
	rules           []Markdown[RuleFrontMatter] // Collected rule files
	totalTokens     int
	logger          *slog.Logger
	cmdRunner       func(cmd *exec.Cmd) error
	resume          bool
	agent           Agent
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

type markdownVisitor func(path string) error

// findMarkdownFile searches for a markdown file by name in the given directories.
// Returns the path to the file if found, or an error if not found or multiple files match.
func (cc *Context) visitMarkdownFiles(searchDirFn func(path string) []string, visitor markdownVisitor) error {

	var searchDirs []string
	for _, path := range cc.downloadedPaths {
		searchDirs = append(searchDirs, searchDirFn(path)...)
	}

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			ext := filepath.Ext(path) // .md or .mdc
			if info.IsDir() || ext != ".md" && ext != ".mdc" {
				return nil
			}

			// If selectors are provided, check if the file matches
			// Parse frontmatter to check selectors
			var fm BaseFrontMatter
			if _, err := ParseMarkdownFile(path, &fm); err != nil {
				// Skip files that can't be parsed
				return nil
			}

			// Skip files that don't match selectors
			if !cc.includes.MatchesIncludes(fm) {
				return nil
			}
			return visitor(path)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// findTask searches for a task markdown file and returns it with parameters substituted
func (cc *Context) findTask(taskName string) error {

	// Add task name to includes so rules can be filtered
	cc.includes.SetValue("task_name", taskName)

	err := cc.visitMarkdownFiles(taskSearchPaths, func(path string) error {
		baseName := filepath.Base(path)
		ext := filepath.Ext(baseName)
		if strings.TrimSuffix(baseName, ext) != taskName {
			return nil
		}

		var frontMatter TaskFrontMatter
		md, err := ParseMarkdownFile(path, &frontMatter)
		if err != nil {
			return err
		}

		// Extract selector labels from task frontmatter and add them to cc.includes.
		// This combines CLI selectors (from -s flag) with task selectors using OR logic:
		// rules match if their frontmatter value matches ANY selector value for a given key.
		// For example: if CLI has env=development and task has env=production,
		// rules with either env=development OR env=production will be included.
		for key, value := range frontMatter.Selectors {
			switch v := value.(type) {
			case []any:
				for _, item := range v {
					cc.includes.SetValue(key, fmt.Sprint(item))
				}
			default:
				cc.includes.SetValue(key, fmt.Sprint(v))
			}
		}

		// Task frontmatter agent field overrides -a flag
		if frontMatter.Agent != "" {
			agent, err := ParseAgent(frontMatter.Agent)
			if err != nil {
				return err
			}
			cc.agent = agent
		}

		expandedContent := cc.expandParams(md.Content, nil)

		// Trim leading/trailing whitespace to avoid parser issues with empty lines
		expandedContent = strings.TrimSpace(expandedContent)

		task, err := ParseTask(expandedContent)
		if err != nil {
			return err
		}

		finalContent := strings.Builder{}
		for _, block := range task {
			if block.Text != nil {
				finalContent.WriteString(block.Text.Content())
			} else if block.SlashCommand != nil {
				commandContent, err := cc.findCommand(block.SlashCommand.Name, block.SlashCommand.Params())
				if err != nil {
					return err
				}
				finalContent.WriteString(commandContent)
			}
		}

		cc.task = Markdown[TaskFrontMatter]{
			FrontMatter: frontMatter,
			Content:     finalContent.String(),
			Tokens:      estimateTokens(finalContent.String()),
		}
		cc.totalTokens += cc.task.Tokens

		cc.logger.Info("Including task", "tokens", cc.task.Tokens)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}
	if cc.task.Content == "" {
		return fmt.Errorf("task not found: %s", taskName)
	}
	return nil
}

// findCommand searches for a command markdown file and returns it with parameters substituted.
// Commands don't have frontmatter, so only the content is returned.
func (cc *Context) findCommand(commandName string, params map[string]string) (string, error) {
	var content *string
	err := cc.visitMarkdownFiles(commandSearchPaths, func(path string) error {
		var frontMatter CommandFrontMatter
		md, err := ParseMarkdownFile(path, &frontMatter)
		if err != nil {
			return err
		}

		expanded := cc.expandParams(md.Content, params)
		content = &expanded

		return nil
	})
	if err != nil {
		return "", err
	}
	if content == nil {
		return "", fmt.Errorf("command not found: %s", commandName)
	}
	return *content, nil
}

// expandParams substitutes parameter placeholders in the given content.
// Supports both regular parameters (${param_name}) and file references (${file:path/to/file.txt})
func (cc *Context) expandParams(content string, params map[string]string) string {
	return os.Expand(content, func(key string) string {
		// Check if this is a file reference (starts with "file:")
		if strings.HasPrefix(key, "file:") {
			filePath := strings.TrimPrefix(key, "file:")

			// Determine base directory for file resolution
			// The first downloadedPath is always the working directory (passed via -C flag or default ".")
			// This is set in main.go where workDir is added as the first search path
			baseDir := "."
			if len(cc.downloadedPaths) > 0 {
				baseDir = cc.downloadedPaths[0]
			}

			// Read and format the file content
			fileContent, err := readFileReference(filePath, baseDir)
			if err != nil {
				cc.logger.Warn("Failed to expand file reference", "file", filePath, "error", err)
				// Return original placeholder on error
				return fmt.Sprintf("${%s}", key)
			}

			return formatFileContent(filePath, fileContent)
		}

		// Regular parameter expansion
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
}

// Run executes the context assembly for the given taskName and returns the assembled result.
// The taskName is looked up in task search paths and its content is parsed into blocks.
// If the taskName cannot be found as a task file, an error is returned.
func (cc *Context) Run(ctx context.Context, taskName string) (*Result, error) {
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

	// Get the task by name
	if err := cc.findTask(taskName); err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if err := cc.findExecuteRuleFiles(ctx, homeDir); err != nil {
		return nil, fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	// Estimate tokens for task
	cc.logger.Info("Total estimated tokens", "tokens", cc.totalTokens)

	// Build and return the result
	result := &Result{
		Rules:  cc.rules,
		Task:   cc.task,
		Tokens: cc.totalTokens,
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
		cc.downloadedPaths = append(cc.downloadedPaths, dst)
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

	err := cc.visitMarkdownFiles(func(path string) []string { return rulePaths(path, path == homeDir) }, func(path string) error {
		var frontmatter RuleFrontMatter
		md, err := ParseMarkdownFile(path, &frontmatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file: %w", err)
		}

		expandedContent := cc.expandParams(md.Content, nil)
		tokens := estimateTokens(expandedContent)

		cc.rules = append(cc.rules, Markdown[RuleFrontMatter]{
			FrontMatter: frontmatter,
			Content:     expandedContent,
			Tokens:      tokens,
		})

		cc.totalTokens += tokens

		cc.logger.Info("Including rule file", "path", path, "tokens", tokens)

		if err := cc.runBootstrapScript(ctx, path); err != nil {
			return fmt.Errorf("failed to run bootstrap script: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	return nil
}

func (cc *Context) runBootstrapScript(ctx context.Context, path string) error {
	// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
	// For example, setup.md -> setup-bootstrap, setup.mdc -> setup-bootstrap
	baseNameWithoutExt := strings.TrimSuffix(path, filepath.Ext(path))
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

	return cc.cmdRunner(cmd)
}
