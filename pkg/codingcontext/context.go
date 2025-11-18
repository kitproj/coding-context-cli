package codingcontext

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

// Context represents a coding context builder that assembles rules and tasks
type Context struct {
	workDir             string
	resume              bool
	params              ParamMap
	includes            SelectorMap
	remotePaths         []string
	emitTaskFrontmatter bool

	downloadedDirs   []string
	matchingTaskFile string
	taskFrontmatter  FrontMatter // Parsed task frontmatter
	taskContent      string      // Parsed task content (before parameter expansion)
	totalTokens      int
	output           io.Writer
	logOut           io.Writer
	cmdRunner        func(cmd *exec.Cmd) error
}

// Option is a functional option for configuring Context
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

// WithParams sets the parameter map
func WithParams(params ParamMap) Option {
	return func(c *Context) {
		c.params = params
	}
}

// WithIncludes sets the selector map
func WithIncludes(includes SelectorMap) Option {
	return func(c *Context) {
		c.includes = includes
	}
}

// WithRemotePaths sets the remote paths to download
func WithRemotePaths(paths []string) Option {
	return func(c *Context) {
		c.remotePaths = paths
	}
}

// WithEmitTaskFrontmatter enables emitting task frontmatter
func WithEmitTaskFrontmatter(emit bool) Option {
	return func(c *Context) {
		c.emitTaskFrontmatter = emit
	}
}

// WithOutput sets the output writer for the assembled context
func WithOutput(w io.Writer) Option {
	return func(c *Context) {
		c.output = w
	}
}

// WithLogOutput sets the output writer for log messages
func WithLogOutput(w io.Writer) Option {
	return func(c *Context) {
		c.logOut = w
	}
}

// WithCmdRunner sets a custom command runner (for testing)
func WithCmdRunner(runner func(cmd *exec.Cmd) error) Option {
	return func(c *Context) {
		c.cmdRunner = runner
	}
}

// New creates a new Context with the given options
func New(opts ...Option) *Context {
	c := &Context{
		workDir:  ".",
		params:   make(ParamMap),
		includes: make(SelectorMap),
		output:   os.Stdout,
		logOut:   os.Stderr,
		cmdRunner: func(cmd *exec.Cmd) error {
			return cmd.Run()
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Run executes the context assembly for the given task name
func (c *Context) Run(ctx context.Context, taskName string) error {
	if err := os.Chdir(c.workDir); err != nil {
		return fmt.Errorf("failed to chdir to %s: %w", c.workDir, err)
	}

	if err := c.downloadRemoteDirectories(ctx); err != nil {
		return fmt.Errorf("failed to download remote directories: %w", err)
	}
	defer c.cleanupDownloadedDirectories()

	// Add task name to includes so rules can be filtered by task
	c.includes.SetValue("task_name", taskName)
	c.includes.SetValue("resume", fmt.Sprint(c.resume))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	if err := c.findTaskFile(homeDir, taskName); err != nil {
		return fmt.Errorf("failed to find task file: %w", err)
	}

	// Parse task file early to extract selector labels for filtering rules and tools
	if err := c.parseTaskFile(); err != nil {
		return fmt.Errorf("failed to parse task file: %w", err)
	}

	if err := c.printTaskFrontmatter(); err != nil {
		return fmt.Errorf("failed to emit task frontmatter: %w", err)
	}

	if err := c.findExecuteRuleFiles(ctx, homeDir); err != nil {
		return fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	// Run bootstrap script for the task file if it exists
	taskExt := filepath.Ext(c.matchingTaskFile)
	if err := c.runBootstrapScript(ctx, c.matchingTaskFile, taskExt); err != nil {
		return fmt.Errorf("failed to run task bootstrap script: %w", err)
	}

	if err := c.emitTaskFileContent(); err != nil {
		return fmt.Errorf("failed to emit task file content: %w", err)
	}

	return nil
}

// TotalTokens returns the total estimated tokens in the assembled context
func (c *Context) TotalTokens() int {
	return c.totalTokens
}

func (c *Context) downloadRemoteDirectories(ctx context.Context) error {
	for _, remotePath := range c.remotePaths {
		fmt.Fprintf(c.logOut, "⪢ Downloading remote directory: %s\n", remotePath)
		localPath, err := DownloadRemoteDirectory(ctx, remotePath)
		if err != nil {
			return fmt.Errorf("failed to download remote directory %s: %w", remotePath, err)
		}
		c.downloadedDirs = append(c.downloadedDirs, localPath)
		fmt.Fprintf(c.logOut, "⪢ Downloaded to: %s\n", localPath)
	}

	return nil
}

func (c *Context) cleanupDownloadedDirectories() {
	for _, dir := range c.downloadedDirs {
		if dir == "" {
			continue
		}

		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(c.logOut, "⪢ Error cleaning up downloaded directory %s: %v\n", dir, err)
		}
	}
}

func (c *Context) findTaskFile(homeDir string, taskName string) error {
	// find the task prompt by searching for a file with matching task_name in frontmatter
	taskSearchDirs := AllTaskSearchPaths(homeDir)

	// Add downloaded remote directories to task search paths
	for _, dir := range c.downloadedDirs {
		taskSearchDirs = append(taskSearchDirs, DownloadedTaskSearchPaths(dir)...)
	}

	for _, dir := range taskSearchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat task dir %s: %w", dir, err)
		}

		if err := filepath.Walk(dir, c.taskFileWalker(taskName)); err != nil {
			return err
		}
	}

	if c.matchingTaskFile == "" {
		return fmt.Errorf("no task file found with task_name=%s matching selectors in frontmatter (searched in %v)", taskName, taskSearchDirs)
	}

	return nil
}

func (c *Context) taskFileWalker(taskName string) func(path string, info os.FileInfo, err error) error {
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
		if !c.includes.MatchesIncludes(frontmatter) {
			return nil
		}

		// If we already found a matching task, error on duplicate
		if c.matchingTaskFile != "" {
			return fmt.Errorf("multiple task files found with task_name=%s: %s and %s", taskName, c.matchingTaskFile, path)
		}

		c.matchingTaskFile = path

		return nil
	}
}

func (c *Context) findExecuteRuleFiles(ctx context.Context, homeDir string) error {
	// Skip rule file discovery in resume mode.
	if c.resume {
		return nil
	}

	// Build the list of rule locations (local and remote)
	rulePaths := AllRulePaths(homeDir)

	// Append remote directories to rule paths
	for _, dir := range c.downloadedDirs {
		rulePaths = append(rulePaths, DownloadedRulePaths(dir)...)
	}

	for _, rule := range rulePaths {
		// Skip if the path doesn't exist
		if _, err := os.Stat(rule); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat rule path %s: %w", rule, err)
		}

		if err := filepath.Walk(rule, c.ruleFileWalker(ctx)); err != nil {
			return fmt.Errorf("failed to walk rule dir: %w", err)
		}
	}

	return nil
}

func (c *Context) ruleFileWalker(ctx context.Context) func(path string, info os.FileInfo, err error) error {
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
		if !c.includes.MatchesIncludes(frontmatter) {
			fmt.Fprintf(c.logOut, "⪢ Excluding rule file (does not match include selectors): %s\n", path)
			return nil
		}

		if err := c.runBootstrapScript(ctx, path, ext); err != nil {
			return fmt.Errorf("failed to run bootstrap script (path: %s): %w", path, err)
		}

		// Estimate tokens for this file
		tokens := EstimateTokens(content)
		c.totalTokens += tokens
		fmt.Fprintf(c.logOut, "⪢ Including rule file: %s (~%d tokens)\n", path, tokens)
		fmt.Fprintln(c.output, content)

		return nil
	}
}

func (c *Context) runBootstrapScript(ctx context.Context, path, ext string) error {
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

	fmt.Fprintf(c.logOut, "⪢ Running bootstrap script: %s\n", bootstrapFilePath)

	cmd := exec.CommandContext(ctx, bootstrapFilePath)
	cmd.Stdout = c.logOut
	cmd.Stderr = c.logOut

	if err := c.cmdRunner(cmd); err != nil {
		return err
	}

	return nil
}

// parseTaskFile parses the task file and extracts selector labels from frontmatter.
// The selectors are added to c.includes for filtering rules and tools.
// The parsed frontmatter and content are stored in c.taskFrontmatter and c.taskContent.
func (c *Context) parseTaskFile() error {
	c.taskFrontmatter = make(FrontMatter)

	content, err := ParseMarkdownFile(c.matchingTaskFile, &c.taskFrontmatter)
	if err != nil {
		return fmt.Errorf("failed to parse task file %s: %w", c.matchingTaskFile, err)
	}

	c.taskContent = content

	// Extract selector labels from frontmatter
	// Look for a "selectors" field that contains a map of key-value pairs
	// Values can be strings or arrays (for OR logic)
	if selectorsRaw, ok := c.taskFrontmatter["selectors"]; ok {
		selectorsMap, ok := selectorsRaw.(map[string]any)
		if !ok {
			// Try to handle it as a map[interface{}]interface{} (common YAML unmarshal result)
			if selectorsMapAny, ok := selectorsRaw.(map[any]any); ok {
				selectorsMap = make(map[string]any)
				for k, v := range selectorsMapAny {
					selectorsMap[fmt.Sprint(k)] = v
				}
			} else {
				return fmt.Errorf("task file %s has invalid 'selectors' field: expected map, got %T", c.matchingTaskFile, selectorsRaw)
			}
		}

		// Add selectors to includes
		// Convert all values to map[string]bool for OR logic
		for key, value := range selectorsMap {
			switch v := value.(type) {
			case []any:
				// Convert []any to map[string]bool
				for _, item := range v {
					c.includes.SetValue(key, fmt.Sprint(item))
				}
			case string:
				// Convert string to single value in map
				c.includes.SetValue(key, v)
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
				// Convert scalar numeric or boolean to string
				c.includes.SetValue(key, fmt.Sprint(v))
			default:
				return fmt.Errorf("task file %s has invalid selector value for key %q: expected string or array, got %T", c.matchingTaskFile, key, value)
			}
		}
	}

	return nil
}

// printTaskFrontmatter emits only the task frontmatter to the output and
// only if emitTaskFrontmatter is true.
func (c *Context) printTaskFrontmatter() error {
	if !c.emitTaskFrontmatter {
		return nil
	}

	fmt.Fprintln(c.output, "---")
	if err := yaml.NewEncoder(c.output).Encode(c.taskFrontmatter); err != nil {
		return fmt.Errorf("failed to encode task frontmatter: %w", err)
	}
	fmt.Fprintln(c.output, "---")
	return nil
}

// emitTaskFileContent emits the parsed task content to the output.
// It expands parameters and estimates tokens.
func (c *Context) emitTaskFileContent() error {
	expanded := os.Expand(c.taskContent, func(key string) string {
		if val, ok := c.params[key]; ok {
			return val
		}
		// this might not exist, in that case, return the original text
		return fmt.Sprintf("${%s}", key)
	})

	// Estimate tokens for this file
	tokens := EstimateTokens(expanded)
	c.totalTokens += tokens
	fmt.Fprintf(c.logOut, "⪢ Including task file: %s (~%d tokens)\n", c.matchingTaskFile, tokens)

	fmt.Fprintln(c.output, expanded)

	// Print total token count
	fmt.Fprintf(c.logOut, "⪢ Total estimated tokens: %d\n", c.totalTokens)

	return nil
}
