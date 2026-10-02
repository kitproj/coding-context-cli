// Package codingcontext provides context assembly for AI coding agents.
package codingcontext

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/skills"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/tokencount"
)

var (
	// ErrTaskNotFound is returned when the requested task file cannot be found.
	ErrTaskNotFound = errors.New("task not found")
	// ErrCommandNotFound is returned when a referenced command file cannot be found.
	ErrCommandNotFound = errors.New("command not found")
	// ErrSkillMissingName is returned when a skill's frontmatter lacks the required name field.
	ErrSkillMissingName = errors.New("skill missing required 'name' field")
	// ErrSkillNameLength is returned when a skill's name exceeds the maximum length.
	ErrSkillNameLength = errors.New("skill 'name' field must be 1-64 characters")
	// ErrSkillMissingDesc is returned when a skill's frontmatter lacks the required description field.
	ErrSkillMissingDesc = errors.New("skill missing required 'description' field")
	// ErrSkillDescriptionLength is returned when a skill's description exceeds the maximum length.
	ErrSkillDescriptionLength = errors.New("skill 'description' field must be 1-1024 characters")

	// ErrMultipleAgents is returned when more than one agent option (WithAgent, WithLenientAgent) is used.
	// These options are mutually exclusive; only one agent may be set.
	ErrMultipleAgents = errors.New("only one agent option (WithAgent or WithLenientAgent) may be used")
	// ErrCallerNotAllowed is returned when a task does not permit the configured invocation surface or caller.
	ErrCallerNotAllowed = errors.New("task invocation not allowed")

	// ErrInvalidTaskNameNamespace is returned when the task name has an empty namespace.
	ErrInvalidTaskNameNamespace = errors.New("namespace must not be empty")
	// ErrInvalidTaskNameBase is returned when the task name has an empty base name.
	ErrInvalidTaskNameBase = errors.New("task base name must not be empty")
	// ErrInvalidTaskNameDepth is returned when the task name has more than one level of namespacing.
	ErrInvalidTaskNameDepth = errors.New("only one level of namespacing is supported (expected \"namespace/task\")")
)

const (
	maxNamespacedParts = 2
	splitLimit         = 3
)

// Context holds the configuration and state for assembling coding context.
// SearchPath represents a search path with an optional lenient flag.
// When Lenient is true, errors encountered while processing files from this path
// are logged as warnings and skipped rather than treated as fatal errors.
type SearchPath struct {
	Path    string
	Lenient bool
}

type Context struct {
	params           taskparser.Params
	includes         selectors.Selectors
	manifestURL      string
	searchPaths      []SearchPath
	downloadedPaths  []SearchPath
	task             markdown.Markdown[markdown.TaskFrontMatter]   // Parsed task
	rules            []markdown.Markdown[markdown.RuleFrontMatter] // Collected rule files
	ruleSources      []string                                      // First-discovered paths, aligned with rules
	skills           skills.AvailableSkills                        // Discovered skills (metadata only)
	totalTokens      int
	logger           *slog.Logger
	cmdRunner        func(cmd *exec.Cmd) error
	resume           bool
	doBootstrap      bool // Controls whether to discover rules, skills, and run bootstrap scripts
	includeByDefault bool // Controls whether unmatched rules/skills are included by default
	agent            Agent
	lenientAgent     bool   // When true, agent-specific paths are treated as lenient
	agentSetCount    int    // Incremented by WithAgent and WithLenientAgent; >1 means conflict
	namespace        string // Active namespace derived from task name (e.g. "myteam" from "myteam/fix-bug")
	userPrompt       string // User-provided prompt to append to task
	surface          string
	caller           CallerIdentity
	lintMode         bool
	lintCollector    *lintCollector
}

// parseNamespacedTaskName splits a task name into its optional namespace and base name.
// "myteam/fix-bug" → ("myteam", "fix-bug", nil)
// "fix-bug"        → ("", "fix-bug", nil)
// "a/b/c"          → error (only one level of namespacing is supported)
// "/task" or "ns/" → error (empty namespace or base name).
func parseNamespacedTaskName(taskName string) (string, string, error) {
	parts := strings.SplitN(taskName, "/", splitLimit)
	switch len(parts) {
	case 1:
		return "", parts[0], nil
	case maxNamespacedParts:
		ns, base := parts[0], parts[1]
		if ns == "" {
			return "", "", fmt.Errorf("invalid task name %q: %w", taskName, ErrInvalidTaskNameNamespace)
		}

		if base == "" {
			return "", "", fmt.Errorf("invalid task name %q: %w", taskName, ErrInvalidTaskNameBase)
		}

		return ns, base, nil
	default:
		return "", "", fmt.Errorf("invalid task name %q: %w", taskName, ErrInvalidTaskNameDepth)
	}
}

// New creates a new Context with the given options.
func New(opts ...Option) *Context {
	c := &Context{
		params:           make(taskparser.Params),
		includes:         make(selectors.Selectors),
		rules:            make([]markdown.Markdown[markdown.RuleFrontMatter], 0),
		skills:           skills.AvailableSkills{Skills: make([]skills.Skill, 0)},
		logger:           slog.New(slog.NewTextHandler(os.Stderr, nil)),
		doBootstrap:      true, // Default to true for backward compatibility
		includeByDefault: true, // Default to true for backward compatibility
		cmdRunner: func(cmd *exec.Cmd) error {
			return cmd.Run()
		},
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// nameFromPath returns the filename without extension. Used to default Name in frontmatter when omitted.
func nameFromPath(path string) string {
	baseName := filepath.Base(path)
	ext := filepath.Ext(baseName)

	return strings.TrimSuffix(baseName, ext)
}

type markdownVisitor func(path string, fm *markdown.BaseFrontMatter) error

// A ruleFileError names the rule file involved in a discovery or bootstrap failure.
type ruleFileError struct {
	path string
	err  error
}

func (e *ruleFileError) Error() string {
	return fmt.Sprintf("%s: %s", e.path, e.err)
}

func (e *ruleFileError) Unwrap() error {
	return e.err
}

// pendingRule is a parsed, expanded rule that has not been bootstrapped yet.
type pendingRule struct {
	path          string
	bootstrapPath string // Alias companion if present, otherwise canonical companion
	md            markdown.Markdown[markdown.RuleFrontMatter]
	reason        string // selector match explanation
	tokens        int
}

// Run executes the context assembly for the given taskName and returns the assembled result.
// The taskName is looked up in task search paths and its content is parsed into blocks.
// If the taskName cannot be found as a task file, an error is returned.
func (cc *Context) Run(ctx context.Context, taskName string) (*Result, error) {
	if cc.agentSetCount > 1 {
		return nil, ErrMultipleAgents
	}

	// Parse manifest file first to get additional search paths
	manifestPaths, err := cc.parseManifestFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest file: %w", err)
	}

	for _, p := range manifestPaths {
		cc.searchPaths = append(cc.searchPaths, SearchPath{Path: p})
	}

	// Download all remote directories (including those from manifest)
	if err := cc.downloadRemoteDirectories(ctx); err != nil {
		return nil, fmt.Errorf("failed to download remote directories: %w", err)
	}
	defer cc.cleanupDownloadedDirectories()

	// If resume mode is enabled, add resume=true as a selector
	if cc.resume {
		cc.includes.SetValue("resume", "true")
	}

	// Get the task by name
	if err := cc.findTask(taskName); err != nil {
		return nil, fmt.Errorf("find task %q: %w", taskName, err)
	}

	// Log parameters and selectors after task is found
	// This ensures we capture any additions from task/command frontmatter
	cc.logger.Info("Parameters", "params", cc.params.String())
	cc.logger.Info("Selectors", "selectors", cc.includes.String())

	if err := cc.findExecuteRuleFiles(ctx); err != nil {
		return nil, fmt.Errorf("failed to find and execute rule files: %w", err)
	}

	// Discover skills (load metadata only for progressive disclosure)
	if err := cc.discoverSkills(ctx); err != nil {
		return nil, fmt.Errorf("failed to discover skills: %w", err)
	}

	// Estimate tokens for task
	cc.logger.Info("Total estimated tokens", "tokens", cc.totalTokens)

	prompt, err := cc.buildPrompt()
	if err != nil {
		return nil, err
	}

	// Build and return the result
	result := &Result{
		Name:      taskName,
		Namespace: cc.namespace,
		Rules:     cc.rules,
		Task:      cc.task,
		Skills:    cc.skills,
		Tokens:    cc.totalTokens,
		Agent:     cc.agent,
		Prompt:    prompt,
	}

	return result, nil
}

func (cc *Context) visitMarkdownFiles(searchDirFn func(path string) []string, visitor markdownVisitor) error {
	type searchDir struct {
		path    string
		lenient bool
	}

	searchDirs := make([]searchDir, 0, len(cc.downloadedPaths))

	for _, sp := range cc.downloadedPaths {
		for _, dir := range searchDirFn(sp.Path) {
			searchDirs = append(searchDirs, searchDir{path: dir, lenient: sp.Lenient})
		}
	}

	for _, dir := range searchDirs {
		if err := cc.visitMarkdownInDir(dir.path, visitor); err != nil {
			if dir.lenient && !errors.Is(err, ErrCallerNotAllowed) {
				cc.logger.Warn("skipping directory", "path", dir.path, "error", err)

				continue
			}

			return fmt.Errorf("process markdown search path %q: %w", dir.path, err)
		}
	}

	return nil
}

func (cc *Context) visitMarkdownInDir(dir string, visitor markdownVisitor) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}

	if err := filepath.Walk(dir, cc.makeMarkdownWalkFunc(visitor)); err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return nil
}

func (cc *Context) makeMarkdownWalkFunc(visitor markdownVisitor) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}

		ext := filepath.Ext(path) // .md or .mdc
		if info.IsDir() || (ext != ".md" && ext != ".mdc") {
			return nil
		}

		var fm markdown.BaseFrontMatter
		if _, parseErr := markdown.ParseMarkdownFileWithLogger(path, &fm, cc.logger); parseErr != nil {
			if cc.lintCollector != nil {
				var pe *markdown.ParseError
				if errors.As(parseErr, &pe) {
					cc.lintCollector.recordParseError(pe)
				} else {
					cc.lintCollector.recordError(path, LintErrorKindParse, parseErr.Error())
				}
			}

			return nil
		}

		if cc.lintCollector != nil {
			cc.lintCollector.recordFrontmatterValues(fm)
		}

		matches, reason := cc.includes.MatchesIncludes(fm, cc.includeByDefault)
		if !matches {
			if reason != "" {
				cc.logger.Info("Skipping file", "path", path, "reason", reason)
			}

			return nil
		}

		return visitor(path, &fm)
	}
}

// findTask searches for a task markdown file and returns it with parameters substituted.
func (cc *Context) findTask(taskName string) error {
	namespace, baseName, err := parseNamespacedTaskName(taskName)
	if err != nil {
		return err
	}

	cc.namespace = namespace

	// Add task_name for both the full namespaced form and the base name so that
	// existing task_names selectors in rule frontmatter continue to work regardless
	// of whether the task is invoked with or without a namespace prefix.
	cc.includes.SetValue("task_name", taskName)

	if baseName != taskName {
		cc.includes.SetValue("task_name", baseName)
	}

	// Expose the namespace as a selector so rules can restrict themselves to a
	// specific namespace via frontmatter (e.g. `namespace: myteam`).
	// For non-namespaced tasks we add an empty-string sentinel so that rules
	// which declare `namespace: somevalue` in their frontmatter are excluded by
	// the selector matching logic (their value won't be in {"": true}).
	cc.includes.SetValue("namespace", namespace)

	taskFound := false

	namespacedPaths := func(dir string) []string {
		return namespacedTaskSearchPaths(dir, namespace)
	}

	err = cc.visitMarkdownFiles(namespacedPaths, func(path string, _ *markdown.BaseFrontMatter) error {
		// Stop after the first matching file so that a namespace task takes
		// precedence over a global task with the same base name.
		if taskFound {
			return nil
		}

		base := filepath.Base(path)
		ext := filepath.Ext(base)

		if strings.TrimSuffix(base, ext) != baseName {
			return nil
		}

		taskFound = true

		return cc.loadTask(path, taskName)
	})
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	if !taskFound {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, taskName)
	}

	return nil
}

// loadTask parses and processes a task file, populating cc.task.
func (cc *Context) loadTask(path, taskName string) error {
	var frontMatter markdown.TaskFrontMatter

	md, err := markdown.ParseMarkdownFileWithLogger(path, &frontMatter, cc.logger)
	if err != nil {
		return fmt.Errorf("failed to parse task file %s: %w", path, err)
	}

	if cc.lintCollector != nil {
		cc.lintCollector.recordFile(path, LoadedFileKindTask)
	}

	if frontMatter.Name == "" {
		frontMatter.Name = nameFromPath(path)
	}

	if !cc.lintMode {
		if err := cc.authorizeTaskInvocation(frontMatter); err != nil {
			return fmt.Errorf("authorize task %q: %w", taskName, err)
		}
	}

	// Extract selector labels from task frontmatter and add them to cc.includes.
	// This combines CLI selectors (from -s flag) with task selectors using OR logic:
	// rules match if their frontmatter value matches ANY selector value for a given key.
	cc.mergeSelectors(frontMatter.Selectors)

	// Apply the task's default inclusion policy for unmatched rules/skills.
	if frontMatter.IncludeUnmatched != nil {
		cc.includeByDefault = *frontMatter.IncludeUnmatched
	}

	// Task frontmatter agent field overrides -a flag
	if frontMatter.Agent != "" {
		agent, err := ParseAgent(frontMatter.Agent)
		if err != nil {
			return fmt.Errorf("failed to parse agent from task frontmatter: %w", err)
		}

		cc.agent = agent
	}

	// Use the task already parsed by the goldmark extension in ParseMarkdownFile.
	// If a user prompt was appended the content changed, so re-parse the combined string
	// to pick up any slash commands in the appended prompt.
	task := md.Task
	if cc.userPrompt != "" {
		taskContent := cc.appendUserPrompt(md.Content)

		var parseErr error

		task, parseErr = taskparser.ParseTaskWithLogger(taskContent, cc.logger)
		if parseErr != nil {
			return fmt.Errorf("failed to parse task content in file %s: %w", path, parseErr)
		}
	}

	finalContent, err := cc.buildFinalContent(task, path, frontMatter.ExpandParams)
	if err != nil {
		return err
	}

	cc.task = markdown.FromContent(frontMatter, finalContent)
	cc.totalTokens += cc.task.Tokens

	cc.logger.Info("Including task", "name", taskName,
		"reason", fmt.Sprintf("task name matches '%s'", taskName), "tokens", cc.task.Tokens)

	return nil
}

// appendUserPrompt appends the user prompt to the task content if present.
func (cc *Context) appendUserPrompt(content string) string {
	if cc.userPrompt == "" {
		return content
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	cc.logger.Info("Appended user_prompt to task", "user_prompt_length", len(cc.userPrompt))

	return content + "---\n" + cc.userPrompt
}

// buildFinalContent processes each block of a pre-parsed task into a final string.
func (cc *Context) buildFinalContent(task taskparser.Task, path string, expandParams *bool) (string, error) {
	var finalContent strings.Builder

	for _, block := range task {
		blockContent, err := cc.processTaskBlock(block, path, expandParams)
		if err != nil {
			return "", err
		}

		finalContent.WriteString(blockContent)
	}

	return finalContent.String(), nil
}

// processTaskBlock processes a single task block (text or slash command) and returns its content.
func (cc *Context) processTaskBlock(block taskparser.Block, path string, expandParams *bool) (string, error) {
	if block.Text != nil {
		textContent := block.Text.Content()

		if shouldExpandParams(expandParams) {
			var err error

			textContent, err = cc.expandParams(textContent, nil)
			if err != nil {
				return "", fmt.Errorf("failed to expand parameters in task file %s: %w", path, err)
			}
		}

		return textContent, nil
	}

	if block.SlashCommand != nil {
		commandContent, err := cc.findCommand(block.SlashCommand.Name, block.SlashCommand.Params())
		if err != nil {
			if errors.Is(err, ErrCommandNotFound) {
				if cc.lintMode {
					cc.lintCollector.recordError(path, LintErrorKindMissingCommand,
						"command not found: "+block.SlashCommand.Name)
				} else {
					cc.logger.Warn("Command not found, passing through as-is",
						"command", block.SlashCommand.Name)
				}

				return block.SlashCommand.String(), nil
			}

			return "", fmt.Errorf("failed to find command %s: %w", block.SlashCommand.Name, err)
		}

		return commandContent, nil
	}

	return "", nil
}

// findCommand searches for a command markdown file and returns its content.
// Commands now support optional frontmatter with the expand field and selectors.
// Parameters are substituted by default (when expand is nil or true).
// Substitution is skipped only when expand is explicitly set to false.
// If the command has selectors in its frontmatter, they are merged into cc.includes
// to allow commands to specify which rules they need.
func (cc *Context) findCommand(commandName string, params taskparser.Params) (string, error) {
	var content *string

	namespacedCmdPaths := func(dir string) []string {
		return namespacedCommandSearchPaths(dir, cc.namespace)
	}

	err := cc.visitMarkdownFiles(namespacedCmdPaths, func(path string, _ *markdown.BaseFrontMatter) error {
		// Stop after the first matching command so that a namespace command takes
		// precedence over a global command with the same name.
		if content != nil {
			return nil
		}

		baseName := filepath.Base(path)

		ext := filepath.Ext(baseName)
		if strings.TrimSuffix(baseName, ext) != commandName {
			return nil
		}

		var frontMatter markdown.CommandFrontMatter

		md, err := markdown.ParseMarkdownFileWithLogger(path, &frontMatter, cc.logger)
		if err != nil {
			return fmt.Errorf("failed to parse command file %s: %w", path, err)
		}

		if cc.lintCollector != nil {
			cc.lintCollector.recordFile(path, LoadedFileKindCommand)
		}

		if frontMatter.Name == "" {
			frontMatter.Name = nameFromPath(path)
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

		cc.logger.Info("Including command", "name", commandName,
			"reason", fmt.Sprintf("referenced by slash command '/%s'", commandName), "path", path)

		return nil
	})
	if err != nil {
		return "", err
	}

	if content == nil {
		return "", fmt.Errorf("%w: %s", ErrCommandNotFound, commandName)
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
	if cc.lintMode {
		return cc.expandParamsLint(content, params)
	}

	// Merge params with cc.params
	mergedParams := make(taskparser.Params)
	maps.Copy(mergedParams, cc.params)
	maps.Copy(mergedParams, params)

	// Use the expand function to handle all expansion types
	expanded, err := mergedParams.Expand(content)
	if err != nil {
		return "", fmt.Errorf("failed to expand parameters: %w", err)
	}

	return expanded, nil
}

// expandParamsLint is a lint-mode variant of expandParams that skips shell command
// execution and tracks @path file references in the lint collector.
func (cc *Context) expandParamsLint(content string, params taskparser.Params) (string, error) {
	mergedParams := make(taskparser.Params)
	maps.Copy(mergedParams, cc.params)
	maps.Copy(mergedParams, params)

	var pathRefs []string

	expanded, err := mergedParams.ExpandWith(content, taskparser.ExpandOptions{
		SkipCommands: true,
		PathRefs:     &pathRefs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to expand parameters: %w", err)
	}

	if cc.lintCollector != nil {
		for _, ref := range pathRefs {
			cc.lintCollector.recordFile(ref, LoadedFileKindPathRef)
		}
	}

	return expanded, nil
}

// shouldExpandParams returns true if parameter expansion should occur based on the expandParams field.
// If expandParams is nil (not specified), it defaults to true.
func shouldExpandParams(expandParams *bool) bool {
	if expandParams == nil {
		return true
	}

	return *expandParams
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
	if rest, ok := strings.CutPrefix(path, "file://"); ok {
		return rest
	}

	return path
}

func downloadDir(path string) string {
	// hash the path and prepend it with a temporary directory
	hash := sha256.Sum256([]byte(path))
	tempDir := os.TempDir()

	return filepath.Join(tempDir, hex.EncodeToString(hash[:]))
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

	defer func() { _ = os.RemoveAll(manifestFile) }()

	cc.logger.Info("Downloaded manifest file", "path", manifestFile)

	cleanPath := filepath.Clean(manifestFile)

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}

	defer func() { _ = file.Close() }()

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
	for _, sp := range cc.searchPaths {
		// If the path is local, use it directly without downloading
		if isLocalPath(sp.Path) {
			localPath := normalizeLocalPath(sp.Path)
			cc.logger.Info("Using local directory", "path", localPath)
			cc.downloadedPaths = append(cc.downloadedPaths, SearchPath{Path: localPath, Lenient: sp.Lenient})

			continue
		}

		// Download remote directories
		cc.logger.Info("Downloading remote directory", "path", sp.Path)

		dst := downloadDir(sp.Path)
		if _, err := getter.Get(ctx, dst, sp.Path); err != nil {
			if sp.Lenient {
				cc.logger.Warn("skipping remote directory", "path", sp.Path, "error", err)

				continue
			}

			return fmt.Errorf("failed to download remote directory %s: %w", sp.Path, err)
		}

		cc.logger.Info("Downloaded to", "path", dst)
		cc.downloadedPaths = append(cc.downloadedPaths, SearchPath{Path: dst, Lenient: sp.Lenient})
	}

	return nil
}

func (cc *Context) cleanupDownloadedDirectories() {
	for _, sp := range cc.searchPaths {
		// Skip cleanup for local paths - they should not be deleted
		if isLocalPath(sp.Path) {
			continue
		}

		// Only clean up downloaded remote directories
		dst := downloadDir(sp.Path)
		if err := os.RemoveAll(dst); err != nil {
			cc.logger.Error("Error cleaning up downloaded directory", "path", dst, "error", err)
		}
	}
}

// findExecuteRuleFiles discovers rule files in each rule directory, then
// bootstraps and publishes them. Discovery of one directory finishes before
// any of that directory's bootstrap scripts run. A lenient bootstrap failure
// skips only that rule; a parse or expand error still stops discovery of the
// rest of that directory.
func (cc *Context) findExecuteRuleFiles(ctx context.Context) error {
	if !cc.doBootstrap {
		return nil
	}

	namespacedRulePaths := func(dir string) []string {
		return namespacedRuleSearchPaths(dir, cc.namespace)
	}

	// Claim an identity before expansion or bootstrap: a lenient failure must
	// not retry side effects through another alias later in the same run.
	seen := make(map[string]struct{})
	for _, sp := range cc.downloadedPaths {
		for _, dir := range namespacedRulePaths(sp.Path) {
			var pending []pendingRule

			discoveryErr := cc.visitMarkdownInDir(dir, func(path string, baseFm *markdown.BaseFrontMatter) error {
				identity, err := filepath.EvalSymlinks(path)
				if err != nil {
					return &ruleFileError{path: path, err: fmt.Errorf("resolve rule identity: %w", err)}
				}
				identity, err = filepath.Abs(identity)
				if err != nil {
					return &ruleFileError{path: path, err: fmt.Errorf("resolve absolute rule path: %w", err)}
				}
				if _, exists := seen[identity]; exists {
					cc.logger.Debug("Skipping duplicate rule file", "path", path)
					return nil
				}
				seen[identity] = struct{}{}
				rule, err := cc.discoverPendingRule(path, baseFm)
				if err != nil {
					return err
				}
				// Preserve a companion next to the first-discovered alias. If it
				// has none, use the canonical file's companion so deduplication
				// does not silently drop the target's setup script.
				if rule.md.FrontMatter.Bootstrap == "" {
					companion := strings.TrimSuffix(path, filepath.Ext(path)) + "-bootstrap"
					if _, err := os.Stat(companion); os.IsNotExist(err) {
						rule.bootstrapPath = identity
					} else if err != nil {
						return &ruleFileError{path: path, err: fmt.Errorf("stat rule companion: %w", err)}
					}
				}

				pending = append(pending, rule)

				return nil
			})
			if discoveryErr != nil {
				if !sp.Lenient {
					return discoveryErr
				}

				cc.logger.Warn("stopping rule discovery after error", "path", dir, "error", discoveryErr)
			}

			if err := cc.bootstrapAndPublishRules(ctx, pending, sp.Lenient); err != nil {
				return err
			}
		}
	}

	return nil
}

// discoverPendingRule parses and expands a rule file without running its bootstrap script.
func (cc *Context) discoverPendingRule(
	path string,
	baseFm *markdown.BaseFrontMatter,
) (pendingRule, error) {
	var frontmatter markdown.RuleFrontMatter

	md, err := markdown.ParseMarkdownFileWithLogger(path, &frontmatter, cc.logger)
	if err != nil {
		return pendingRule{}, &ruleFileError{path: path, err: fmt.Errorf("parse markdown file: %w", err)}
	}

	if cc.lintCollector != nil {
		cc.lintCollector.recordFile(path, LoadedFileKindRule)
	}

	if frontmatter.Name == "" {
		frontmatter.Name = nameFromPath(path)
	}

	processedContent := md.Content
	if shouldExpandParams(frontmatter.ExpandParams) {
		processedContent, err = cc.expandParams(md.Content, nil)
		if err != nil {
			return pendingRule{}, &ruleFileError{path: path, err: fmt.Errorf("expand parameters: %w", err)}
		}
	}

	tokens := tokencount.EstimateTokens(processedContent)
	_, reason := cc.includes.MatchesIncludes(*baseFm, cc.includeByDefault)

	return pendingRule{
		path:          path,
		bootstrapPath: path,
		md:            markdown.FromContent(frontmatter, processedContent),
		reason:        reason,
		tokens:        tokens,
	}, nil
}

// bootstrapAndPublishRules runs each pending rule's bootstrap script and, on
// success, publishes the rule. When lenient is true, a bootstrap failure logs
// a warning and continues with the remaining rules.
func (cc *Context) bootstrapAndPublishRules(
	ctx context.Context,
	pending []pendingRule,
	lenient bool,
) error {
	for _, rule := range pending {
		if err := cc.runBootstrapScript(ctx, rule.bootstrapPath, rule.md.FrontMatter.Bootstrap); err != nil {
			if lenient {
				cc.logger.Warn(
					"skipping rule file after bootstrap failure",
					"path", rule.path,
					"error", err,
				)

				continue
			}

			return &ruleFileError{path: rule.path, err: fmt.Errorf("run bootstrap script: %w", err)}
		}

		cc.rules = append(cc.rules, rule.md)
		cc.ruleSources = append(cc.ruleSources, rule.path)
		cc.totalTokens += rule.tokens
		cc.logger.Info(
			"Including rule file",
			"path", rule.path,
			"reason", rule.reason,
			"tokens", rule.tokens,
		)
	}

	return nil
}

func (cc *Context) runBootstrapScript(ctx context.Context, path string, frontmatterBootstrap string) error {
	// executablePerm is the permission for executable scripts (0755) - required for direct execution and shebang support.
	const executablePerm = 0o755

	// In lint mode, skip execution but stat-check companion bootstrap files.
	if cc.lintMode {
		cc.recordLintBootstrap(path, frontmatterBootstrap)

		return nil
	}

	// Prefer frontmatter bootstrap if present
	if frontmatterBootstrap != "" {
		cc.logger.Info("Running bootstrap from frontmatter", "path", path)

		tmpFile, err := os.CreateTemp("", "bootstrap-*.sh")
		if err != nil {
			return fmt.Errorf("failed to create temp file for bootstrap script from %s: %w", path, err)
		}

		tmpFilePath := tmpFile.Name()

		defer func() { _ = os.Remove(tmpFilePath) }()

		if _, err := tmpFile.WriteString(frontmatterBootstrap); err != nil {
			_ = tmpFile.Close()

			return fmt.Errorf("failed to write bootstrap script from %s: %w", path, err)
		}

		_ = tmpFile.Close()

		// Scripts must be executable to run; supports shebangs (e.g. #!/usr/bin/python3)
		// #nosec G302 G703 -- bootstrap scripts require 0755; tmpFilePath from CreateTemp is system-generated
		if err := os.Chmod(tmpFilePath, executablePerm); err != nil {
			return fmt.Errorf("failed to chmod bootstrap script from %s: %w", path, err)
		}

		// #nosec G204 G702 -- intentionally executing user-defined bootstrap scripts; path from CreateTemp
		cmd := exec.CommandContext(ctx, tmpFilePath)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr

		if err := cc.cmdRunner(cmd); err != nil {
			return fmt.Errorf("frontmatter bootstrap script failed for %s: %w", path, err)
		}

		return nil
	}

	// Fall back to file-based bootstrap
	baseNameWithoutExt := strings.TrimSuffix(path, filepath.Ext(path))
	bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

	if _, err := os.Stat(bootstrapFilePath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat bootstrap file %s: %w", bootstrapFilePath, err)
	}

	cc.logger.Info("Running bootstrap script", "path", bootstrapFilePath)

	// #nosec G302 -- bootstrap scripts require executablePerm for direct execution and shebang support
	if err := os.Chmod(bootstrapFilePath, executablePerm); err != nil {
		return fmt.Errorf("failed to chmod bootstrap file %s: %w", bootstrapFilePath, err)
	}

	// #nosec G204 G702 -- intentionally executing user-defined bootstrap scripts
	cmd := exec.CommandContext(ctx, bootstrapFilePath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cc.cmdRunner(cmd); err != nil {
		return fmt.Errorf("file-based bootstrap script failed for %s: %w", path, err)
	}

	return nil
}

// discoverSkills searches for skill directories and loads only their metadata (name and description)
// for progressive disclosure. Skills are folders containing a SKILL.md file.
func (cc *Context) discoverSkills(ctx context.Context) error {
	// Skip skill discovery if bootstrap is disabled
	if !cc.doBootstrap {
		return nil
	}

	type skillDir struct {
		path    string
		lenient bool
	}

	// Determine the agent's skill path prefix for lenient agent handling.
	var agentSkillSuffix string
	if cc.lenientAgent && cc.agent.IsSet() {
		agentPaths := getAgentsPaths()
		if cfg, ok := agentPaths[cc.agent]; ok {
			agentSkillSuffix = cfg.skillsPath
		}
	}

	var skillPaths []skillDir

	for _, sp := range cc.downloadedPaths {
		for _, dir := range namespacedSkillSearchPaths(sp.Path, cc.namespace) {
			lenient := sp.Lenient
			if !lenient && agentSkillSuffix != "" && strings.HasSuffix(dir, agentSkillSuffix) {
				lenient = true
			}

			skillPaths = append(skillPaths, skillDir{path: dir, lenient: lenient})
		}
	}

	cache := skillCacheFromContext(ctx)

	for _, dir := range skillPaths {
		if err := cc.discoverSkillsInDir(dir.path, dir.lenient, cache); err != nil {
			return err
		}
	}

	return nil
}

// discoverSkillsInDir discovers skills within a single directory. cache may
// be nil (no caching, parse every call) — see skillCacheFromContext.
func (cc *Context) discoverSkillsInDir(dir string, lenient bool, cache *SkillCache) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		if lenient {
			cc.logger.Warn("skipping skill directory", "path", dir, "error", err)

			return nil
		}

		return fmt.Errorf("failed to stat skill directory %s: %w", dir, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if lenient {
			cc.logger.Warn("skipping skill directory", "path", dir, "error", err)

			return nil
		}

		return fmt.Errorf("failed to read skill directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillFile := filepath.Join(dir, entry.Name(), "SKILL.md")

		if err := cc.loadSkillEntry(skillFile, lenient, cache); err != nil {
			return err
		}
	}

	return nil
}

// loadSkillEntry loads and validates a single skill from its SKILL.md file.
// cache may be nil — see skillCacheFromContext.
func (cc *Context) loadSkillEntry(skillFile string, lenient bool, cache *SkillCache) error {
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		if lenient {
			cc.logger.Warn("skipping skill file", "path", skillFile, "error", err)

			return nil
		}

		return fmt.Errorf("failed to stat skill file %s: %w", skillFile, err)
	}

	var (
		frontmatter markdown.SkillFrontMatter
		parseErr    error
	)

	if cached, ok := cache.get(skillFile); ok {
		frontmatter, parseErr = cached.frontmatter, cached.err
	} else {
		_, parseErr = markdown.ParseMarkdownFileWithLogger(skillFile, &frontmatter, cc.logger)

		// Cache only deterministic outcomes. A parse error means the file was
		// read and its content rejected, so every later call would reject it
		// identically; an I/O error (permissions, fd exhaustion, a flaky
		// network filesystem, a file that vanished mid-walk) says nothing
		// about the next attempt, and caching one would let a single transient
		// blip poison the rest of the run.
		if parseErr == nil {
			cache.set(skillFile, skillParseResult{frontmatter: frontmatter})
		} else if isDeterministicParseError(parseErr) {
			// Discard frontmatter on error: parsing populates it field by
			// field, so a failed parse can leave it partly filled. Only err is
			// consulted while it is non-nil, but caching a half-built struct
			// invites a later reader to trust it.
			cache.set(skillFile, skillParseResult{err: parseErr})
		}
	}

	if parseErr != nil {
		if lenient {
			cc.logger.Warn("skipping skill file: failed to parse YAML frontmatter", "path", skillFile, "error", parseErr)

			return nil
		}

		return fmt.Errorf("failed to parse skill file %s: %w", skillFile, parseErr)
	}

	if cc.lintCollector != nil {
		cc.lintCollector.recordFile(skillFile, LoadedFileKindSkill)
		cc.lintCollector.recordFrontmatterValues(frontmatter.BaseFrontMatter)
	}

	matches, reason := cc.includes.MatchesIncludes(frontmatter.BaseFrontMatter, cc.includeByDefault)
	if !matches {
		if reason != "" {
			cc.logger.Info("Skipping skill", "name", frontmatter.Name, "path", skillFile, "reason", reason)
		}

		return nil
	}

	return cc.validateAndAddSkill(frontmatter, skillFile, reason, lenient)
}

// validateAndAddSkill validates skill metadata and adds it to the skill collection.
func (cc *Context) validateAndAddSkill(frontmatter markdown.SkillFrontMatter, skillFile, reason string, lenient bool) error {
	if frontmatter.Name == "" {
		if lenient {
			// Infer name from the skill's parent directory
			frontmatter.Name = filepath.Base(filepath.Dir(skillFile))
			cc.logger.Warn("using inferred skill name", "name", frontmatter.Name, "path", skillFile)
		} else if cc.lintMode {
			cc.lintCollector.recordError(skillFile, LintErrorKindSkillValidation,
				fmt.Sprintf("%v: %s", ErrSkillMissingName, skillFile))

			return nil
		} else {
			return fmt.Errorf("%w: %s", ErrSkillMissingName, skillFile)
		}
	}

	const maxSkillNameLen = 64
	if len(frontmatter.Name) > maxSkillNameLen {
		if lenient {
			cc.logger.Warn("skill name exceeds maximum length", "path", skillFile, "length", len(frontmatter.Name))
		} else if cc.lintMode {
			cc.lintCollector.recordError(skillFile, LintErrorKindSkillValidation,
				fmt.Sprintf("%v: %s (got %d)", ErrSkillNameLength, skillFile, len(frontmatter.Name)))

			return nil
		} else {
			return fmt.Errorf("%w: %s (got %d)", ErrSkillNameLength, skillFile, len(frontmatter.Name))
		}
	}

	if frontmatter.Description == "" {
		if lenient {
			cc.logger.Warn("skipping skill: missing 'description' field", "path", skillFile)

			return nil
		} else if cc.lintMode {
			cc.lintCollector.recordError(skillFile, LintErrorKindSkillValidation,
				fmt.Sprintf("%v: %s", ErrSkillMissingDesc, skillFile))

			return nil
		}

		return fmt.Errorf("%w: %s", ErrSkillMissingDesc, skillFile)
	}

	const maxSkillDescLen = 1024
	if len(frontmatter.Description) > maxSkillDescLen {
		if lenient {
			cc.logger.Warn("skill description exceeds maximum length", "path", skillFile, "length", len(frontmatter.Description))
		} else if cc.lintMode {
			cc.lintCollector.recordError(skillFile, LintErrorKindSkillValidation,
				fmt.Sprintf("%v: %s (got %d)", ErrSkillDescriptionLength, skillFile, len(frontmatter.Description)))

			return nil
		} else {
			return fmt.Errorf("%w: %s (got %d)", ErrSkillDescriptionLength, skillFile, len(frontmatter.Description))
		}
	}

	absPath, err := filepath.Abs(skillFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for skill %s: %w", skillFile, err)
	}

	cc.skills.Skills = append(cc.skills.Skills, skills.Skill{
		Name:        frontmatter.Name,
		Description: frontmatter.Description,
		Location:    absPath,
	})

	cc.logger.Info("Discovered skill", "name", frontmatter.Name, "reason", reason, "path", absPath)

	return nil
}
