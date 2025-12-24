package codingcontext

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-getter/v2"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/selectors"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/tokencount"
)

// Context holds the configuration and state for assembling coding context
type Context struct {
	params          taskparser.Params
	includes        selectors.Selectors
	manifestURL     string
	searchPaths     []string
	downloadedPaths []string
	task            markdown.Markdown[markdown.TaskFrontMatter]   // Parsed task
	rules           []markdown.Markdown[markdown.RuleFrontMatter] // Collected rule files
	totalTokens     int
	logger          *slog.Logger
	cmdRunner       func(cmd *exec.Cmd) error
	resume          bool
	agent           Agent
	userPrompt      string // User-provided prompt to append to task
}

// New creates a new Context with the given options
func New(opts ...Option) *Context {
	c := &Context{
		params:   make(taskparser.Params),
		includes: make(selectors.Selectors),
		rules:    make([]markdown.Markdown[markdown.RuleFrontMatter], 0),
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
			return fmt.Errorf("failed to stat directory %s: %w", dir, err)
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("failed to walk path %s: %w", path, err)
			}
			ext := filepath.Ext(path) // .md or .mdc
			if info.IsDir() || ext != ".md" && ext != ".mdc" {
				return nil
			}

			// If selectors are provided, check if the file matches
			// Parse frontmatter to check selectors
			var fm markdown.BaseFrontMatter
			if _, err := markdown.ParseMarkdownFile(path, &fm); err != nil {
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
			return fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return nil
}

// findTask searches for a task markdown file and returns it with parameters substituted
func (cc *Context) findTask(taskName string) error {
	// Add task name to includes so rules can be filtered
	cc.includes.SetValue("task_name", taskName)

	taskFound := false
	err := cc.visitMarkdownFiles(taskSearchPaths, func(path string) error {
		baseName := filepath.Base(path)
		ext := filepath.Ext(baseName)
		if strings.TrimSuffix(baseName, ext) != taskName {
			return nil
		}

		taskFound = true
		var frontMatter markdown.TaskFrontMatter
		md, err := markdown.ParseMarkdownFile(path, &frontMatter)
		if err != nil {
			return fmt.Errorf("failed to parse task file %s: %w", path, err)
		}

		// Extract selector labels from task frontmatter and add them to cc.includes.
		// This combines CLI selectors (from -s flag) with task selectors using OR logic:
		// rules match if their frontmatter value matches ANY selector value for a given key.
		// For example: if CLI has env=development and task has env=production,
		// rules with either env=development OR env=production will be included.
		cc.mergeSelectors(frontMatter.Selectors)

		// Task frontmatter agent field overrides -a flag
		if frontMatter.Agent != "" {
			agent, err := ParseAgent(frontMatter.Agent)
			if err != nil {
				return fmt.Errorf("failed to parse agent from task frontmatter: %w", err)
			}
			cc.agent = agent
		}

		// Append user_prompt to task content before parsing
		// This allows user_prompt to be processed uniformly with task content
		taskContent := md.Content
		if cc.userPrompt != "" {
			// Add delimiter to separate task from user_prompt
			if !strings.HasSuffix(taskContent, "\n") {
				taskContent += "\n"
			}
			taskContent += "---\n" + cc.userPrompt
			cc.logger.Info("Appended user_prompt to task", "user_prompt_length", len(cc.userPrompt))
		}

		// Parse the task content (including user_prompt) to separate text blocks from slash commands
		task, err := taskparser.ParseTask(taskContent)
		if err != nil {
			return fmt.Errorf("failed to parse task content in file %s: %w", path, err)
		}

		// Build the final content by processing each block
		// Text blocks are expanded if expand is not false
		// Slash command arguments are NOT expanded here - they are passed as literals
		// to command files where they may be substituted via ${param} templates
		finalContent := strings.Builder{}
		for _, block := range task {
			if block.Text != nil {
				textContent := block.Text.Content()
				// Expand parameters in text blocks only if expand is not explicitly set to false
				if shouldExpandParams(frontMatter.ExpandParams) {
					textContent, err = cc.expandParams(textContent, nil)
					if err != nil {
						return fmt.Errorf("failed to expand parameters in task file %s: %w", path, err)
					}
				}
				finalContent.WriteString(textContent)
			} else if block.SlashCommand != nil {
				commandContent, err := cc.findCommand(block.SlashCommand.Name, block.SlashCommand.Params())
				if err != nil {
					return fmt.Errorf("failed to find command %s: %w", block.SlashCommand.Name, err)
				}
				finalContent.WriteString(commandContent)
			}
		}

		cc.task = markdown.Markdown[markdown.TaskFrontMatter]{
			FrontMatter: frontMatter,
			Content:     finalContent.String(),
			Tokens:      tokencount.EstimateTokens(finalContent.String()),
		}
		cc.totalTokens += cc.task.Tokens

		cc.logger.Info("Including task", "tokens", cc.task.Tokens)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}
	if !taskFound {
		return fmt.Errorf("task not found: %s", taskName)
	}
	return nil
}

// findCommand searches for a command markdown file and returns its content.
// Commands now support optional frontmatter with the expand field and selectors.
// Parameters are substituted by default (when expand is nil or true).
// Substitution is skipped only when expand is explicitly set to false.
// If the command has selectors in its frontmatter, they are merged into cc.includes
// to allow commands to specify which rules they need.
func (cc *Context) findCommand(commandName string, params taskparser.Params) (string, error) {
	var content *string
	err := cc.visitMarkdownFiles(commandSearchPaths, func(path string) error {
		baseName := filepath.Base(path)
		ext := filepath.Ext(baseName)
		if strings.TrimSuffix(baseName, ext) != commandName {
			return nil
		}

		var frontMatter markdown.CommandFrontMatter
		md, err := markdown.ParseMarkdownFile(path, &frontMatter)
		if err != nil {
			return fmt.Errorf("failed to parse command file %s: %w", path, err)
		}

		// Extract selector labels from command frontmatter and add them to cc.includes.
		// This combines CLI selectors, task selectors, and command selectors using OR logic:
		// rules match if their frontmatter value matches ANY selector value for a given key.
		cc.mergeSelectors(frontMatter.Selectors)

		// Expand parameters only if expand is not explicitly set to false
		var processedContent string
		if shouldExpandParams(frontMatter.ExpandParams) {
			processedContent, err = cc.expandParams(md.Content, params)
			if err != nil {
				return fmt.Errorf("failed to expand parameters in command file %s: %w", path, err)
			}
		} else {
			processedContent = md.Content
		}
		content = &processedContent

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

// mergeSelectors adds selectors from a map into cc.includes.
// This is used to combine selectors from task and command frontmatter with CLI selectors.
// The merge uses OR logic: rules match if their frontmatter value matches ANY selector value for a given key.
func (cc *Context) mergeSelectors(selectors map[string]any) {
	for key, value := range selectors {
		switch v := value.(type) {
		case []any:
			for _, item := range v {
				cc.includes.SetValue(key, fmt.Sprint(item))
			}
		default:
			cc.includes.SetValue(key, fmt.Sprint(v))
		}
	}
}

// expandParams performs all types of content expansion:
// - Parameter expansion: ${param_name}
// - Command expansion: !`command`
// - Path expansion: @path
// If params is provided, it is merged with cc.params (with params taking precedence).
func (cc *Context) expandParams(content string, params taskparser.Params) (string, error) {
	// Merge params with cc.params
	mergedParams := make(taskparser.Params)
	maps.Copy(mergedParams, cc.params)
	maps.Copy(mergedParams, params)

	// Use the expand function to handle all expansion types
	return mergedParams.Expand(content)
}

// shouldExpandParams returns true if parameter expansion should occur based on the expandParams field.
// If expandParams is nil (not specified), it defaults to true.
func shouldExpandParams(expandParams *bool) bool {
	if expandParams == nil {
		return true
	}
	return *expandParams
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

	// Build the combined prompt from all rules and task content
	var promptBuilder strings.Builder
	for _, rule := range cc.rules {
		promptBuilder.WriteString(rule.Content)
		promptBuilder.WriteString("\n")
	}
	promptBuilder.WriteString(cc.task.Content)

	// Build and return the result
	result := &Result{
		Rules:  cc.rules,
		Task:   cc.task,
		Tokens: cc.totalTokens,
		Agent:  cc.agent,
		Prompt: promptBuilder.String(),
	}

	return result, nil
}

// isLocalPath checks if a path is a local file system path.
// Returns true for:
// - file:// URLs (e.g., file:///path/to/dir)
// - Absolute paths (e.g., /path/to/dir)
// - Relative paths (e.g., ./path or ../path)
// Returns false for remote protocols like git::, https://, s3::, etc.
func isLocalPath(path string) bool {
	// Check if path starts with file:// protocol
	if strings.HasPrefix(path, "file://") {
		return true
	}

	// Check if it's an absolute or relative local path
	// (no protocol prefix like git::, https://, s3::, etc.)
	if !strings.Contains(path, "://") && !strings.Contains(path, "::") {
		return true
	}

	return false
}

// normalizeLocalPath converts a local path to a usable file system path.
// For file:// URLs, it strips the protocol prefix.
// For other local paths, it returns them as-is.
func normalizeLocalPath(path string) string {
	if strings.HasPrefix(path, "file://") {
		return strings.TrimPrefix(path, "file://")
	}
	return path
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
		// If the path is local, use it directly without downloading
		if isLocalPath(path) {
			localPath := normalizeLocalPath(path)
			cc.logger.Info("Using local directory", "path", localPath)
			cc.downloadedPaths = append(cc.downloadedPaths, localPath)
			continue
		}

		// Download remote directories
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
		// Skip cleanup for local paths - they should not be deleted
		if isLocalPath(path) {
			continue
		}

		// Only clean up downloaded remote directories
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
		var frontmatter markdown.RuleFrontMatter
		md, err := markdown.ParseMarkdownFile(path, &frontmatter)
		if err != nil {
			return fmt.Errorf("failed to parse markdown file %s: %w", path, err)
		}

		// Expand parameters only if expand is not explicitly set to false
		var processedContent string
		if shouldExpandParams(frontmatter.ExpandParams) {
			processedContent, err = cc.expandParams(md.Content, nil)
			if err != nil {
				return fmt.Errorf("failed to expand parameters in file %s: %w", path, err)
			}
		} else {
			processedContent = md.Content
		}
		tokens := tokencount.EstimateTokens(processedContent)

		cc.rules = append(cc.rules, markdown.Markdown[markdown.RuleFrontMatter]{
			FrontMatter: frontmatter,
			Content:     processedContent,
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
