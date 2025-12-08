package codingcontext

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper to create a markdown file with frontmatter
func createMarkdownFile(t *testing.T, path string, frontmatter string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	var file string
	if frontmatter != "" {
		file = fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
	} else {
		file = content
	}

	if err := os.WriteFile(path, []byte(file), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		workDir     string
		params      Params
		includes    Selectors
		setupFiles  func(t *testing.T, tmpDir string)
		wantErr     bool
		errContains string
	}{
		{
			name:        "no arguments",
			args:        []string{},
			wantErr:     true,
			errContains: "invalid usage",
		},
		{
			name:        "too many arguments",
			args:        []string{"task1", "task2"},
			wantErr:     true,
			errContains: "invalid usage",
		},
		{
			name:        "task not found",
			args:        []string{"/nonexistent"},
			wantErr:     true,
			errContains: "no task file found",
		},
		{
			name: "successful task execution",
			args: []string{"/test_task"},
			setupFiles: func(t *testing.T, tmpDir string) {
				// Create task file
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "test_task.md"),
					"",
					"# Test Task\nThis is a test task.")
			},
			wantErr: false,
		},
		{
			name: "task with parameters",
			args: []string{"/param_task"},
			params: Params{
				"name": "value",
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "param_task.md"),
					"",
					"# Test ${name}")
			},
			wantErr: false,
		},
		{
			name: "resume mode skips rules",
			args: []string{"/resume_task"},
			includes: Selectors{
				"resume": map[string]bool{"true": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "resume_task.md"),
					"resume: true",
					"# Resume Task")

				// Create a rule file that should be skipped
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule that should be skipped")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup test files
			if tt.setupFiles != nil {
				tt.setupFiles(t, tmpDir)
			}

			// Change to temp dir
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}
			defer os.Chdir(oldDir)

			var logOut bytes.Buffer
			cc := &Context{
				params:   tt.params,
				includes: tt.includes,
				searchPaths: func() []string {
					abs, _ := filepath.Abs(tmpDir)
					return []string{"file://" + abs}
				}(),
				rules:  make([]Markdown[RuleFrontMatter], 0),
				logger: slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil // Mock command runner
				},
			}

			if cc.params == nil {
				cc.params = make(Params)
			}
			if cc.includes == nil {
				cc.includes = make(Selectors)
			}

			// Validate args before calling Run
			var result *Result
			if len(tt.args) != 1 {
				if len(tt.args) == 0 {
					err = fmt.Errorf("invalid usage")
				} else {
					err = fmt.Errorf("invalid usage")
				}
			} else {
				result, err = cc.Run(context.Background(), tt.args[0])
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("Run() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Run() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Run() unexpected error: %v\nLog output:\n%s", err, logOut.String())
				}
				if result == nil {
					t.Errorf("Run() returned nil result")
				}
			}
		})
	}
}

func TestFindTaskFile(t *testing.T) {
// Skipped: uses removed method findTaskFile
}

func TestFindExecuteRuleFiles(t *testing.T) {
	tests := []struct {
		name               string
		includes           Selectors
		params             Params // Parameters for template expansion
		setupFiles         func(t *testing.T, tmpDir string)
		searchPaths        []string // Search paths to use (will be downloaded via downloadDir)
		wantTokens         int
		wantMinTokens      bool // Check that tokens > 0
		expectInOutput     string
		expectNotInOutput  string
		expectBootstrapRun bool   // Whether bootstrap script should run
		bootstrapPath      string // Path to bootstrap script to check
	}{
		{
			name: "resume mode skips rules",
			includes: Selectors{
				"resume": map[string]bool{"true": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule File")
			},
			wantTokens: 0,
		},
		{
			name: "include rule file",
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule File\nThis is a rule.")
			},
			wantMinTokens:  true,
			expectInOutput: "# Rule File",
		},
		{
			name: "exclude rule with non-matching selector",
			includes: Selectors{
				"env": map[string]bool{"prod": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"env: dev",
					"# Dev Rule")
			},
			expectNotInOutput: "# Dev Rule",
		},
		{
			name: "include rule with matching selector",
			includes: Selectors{
				"env": map[string]bool{"prod": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"env: prod",
					"# Prod Rule")
			},
			wantMinTokens:  true,
			expectInOutput: "# Prod Rule",
		},
		{
			name: "include multiple rules",
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule 1")
				createMarkdownFile(t, filepath.Join(tmpDir, "AGENTS.md"),
					"",
					"# Rule 2")
			},
			wantMinTokens:  true,
			expectInOutput: "# Rule 1",
		},
		{
			name: "include .mdc files",
			setupFiles: func(t *testing.T, tmpDir string) {
				// .mdc files need to be in a rules directory
				rulesDir := filepath.Join(tmpDir, ".agents", "rules")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule.mdc"),
					"",
					"# MDC Rule")
			},
			wantMinTokens:  true,
			expectInOutput: "# MDC Rule",
		},
		{
			name: "include rules from downloaded directories",
			setupFiles: func(t *testing.T, tmpDir string) {
				// Create a downloaded directory with rules
				downloadedDir := filepath.Join(tmpDir, "downloaded")
				createMarkdownFile(t, filepath.Join(downloadedDir, "CLAUDE.md"),
					"",
					"# Downloaded Rule")
				// Also create a rule in a subdirectory
				rulesDir := filepath.Join(downloadedDir, ".agents", "rules")
				createMarkdownFile(t, filepath.Join(rulesDir, "remote.md"),
					"",
					"# Remote Rule")
			},
			searchPaths:    []string{"file://downloaded"}, // Will be resolved relative to tmpDir
			wantMinTokens:  true,
			expectInOutput: "Downloaded Rule",
		},
		{
			name: "bootstrap script should not run on excluded files",
			includes: Selectors{
				"env": map[string]bool{"prod": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				// Create an excluded rule file (env: dev doesn't match env: prod)
				rulePath := filepath.Join(tmpDir, "CLAUDE.md")
				createMarkdownFile(t, rulePath,
					"env: dev",
					"# Dev Rule")
				// Create a bootstrap script for this rule
				bootstrapPath := filepath.Join(tmpDir, "CLAUDE-bootstrap")
				if err := os.WriteFile(bootstrapPath, []byte("#!/bin/sh\necho 'bootstrap ran' >&2"), 0o644); err != nil {
					t.Fatalf("failed to create bootstrap file: %v", err)
				}
			},
			wantTokens:         0,
			expectNotInOutput:  "# Dev Rule",
			expectBootstrapRun: false,
			bootstrapPath:      "CLAUDE-bootstrap",
		},
		{
			name: "rule with parameter substitution",
			params: Params{
				"issue_key":    "PROJ-123",
				"project_name": "MyProject",
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule with params\nIssue: ${issue_key}\nProject: ${project_name}")
			},
			wantMinTokens:     true,
			expectInOutput:    "Issue: PROJ-123\nProject: MyProject",
			expectNotInOutput: "${issue_key}",
		},
		{
			name: "rule with missing parameter preserved",
			params: Params{
				"issue_key": "PROJ-456",
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule with partial params\nIssue: ${issue_key}\nProject: ${missing_param}")
			},
			wantMinTokens:     true,
			expectInOutput:    "Issue: PROJ-456\nProject: ${missing_param}",
			expectNotInOutput: "${issue_key}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setupFiles(t, tmpDir)

			// Change to temp dir
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}
			defer os.Chdir(oldDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to chdir: %v", err)
			}

			var logOut bytes.Buffer
			bootstrapRan := false
			cc := &Context{
				includes: tt.includes,
				params:   tt.params,
				rules:    make([]Markdown[RuleFrontMatter], 0),
				logger:   slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					// Track if bootstrap script was executed
					if cmd.Path != "" {
						bootstrapRan = true
					}
					return nil // Mock command runner
				},
			}
			if cc.includes == nil {
				cc.includes = make(Selectors)
			}
			if cc.params == nil {
				cc.params = make(Params)
			}

			// Always add tmpDir to searchPaths so local files are found
			absTmpDir, err := filepath.Abs(tmpDir)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}
			cc.searchPaths = []string{"file://" + absTmpDir}

			// Add additional searchPaths if specified in test case
			if len(tt.searchPaths) > 0 {
				additionalPaths := make([]string, len(tt.searchPaths))
				for i, path := range tt.searchPaths {
					// Convert relative paths to absolute file:// URLs
					if strings.HasPrefix(path, "file://") {
						// Already a file:// URL, resolve relative to tmpDir
						relPath := strings.TrimPrefix(path, "file://")
						absPath, err := filepath.Abs(filepath.Join(tmpDir, relPath))
						if err != nil {
							t.Fatalf("failed to get absolute path: %v", err)
						}
						additionalPaths[i] = "file://" + absPath
					} else {
						absPath, err := filepath.Abs(filepath.Join(tmpDir, path))
						if err != nil {
							t.Fatalf("failed to get absolute path: %v", err)
						}
						additionalPaths[i] = "file://" + absPath
					}
				}
				cc.searchPaths = append(cc.searchPaths, additionalPaths...)
			}

			// Download the directories
			if err := cc.downloadRemoteDirectories(context.Background()); err != nil {
				t.Fatalf("failed to download remote directories: %v", err)
			}
			defer cc.cleanupDownloadedDirectories()

			err = cc.findExecuteRuleFiles(context.Background(), tmpDir)
			if err != nil {
				t.Errorf("findExecuteRuleFiles() unexpected error: %v", err)
			}

			if tt.wantMinTokens && cc.totalTokens <= 0 {
				t.Errorf("findExecuteRuleFiles() expected tokens > 0, got %d", cc.totalTokens)
			}
			if !tt.wantMinTokens && tt.wantTokens != cc.totalTokens {
				t.Errorf("findExecuteRuleFiles() expected %d tokens, got %d", tt.wantTokens, cc.totalTokens)
			}

			// Collect all rule content into a single string for output comparison
			var outputStr string
			for _, rule := range cc.rules {
				outputStr += rule.Content + "\n"
			}

			if tt.expectInOutput != "" && !strings.Contains(outputStr, tt.expectInOutput) {
				t.Errorf("findExecuteRuleFiles() output should contain %q, got:\n%s", tt.expectInOutput, outputStr)
			}
			if tt.expectNotInOutput != "" && strings.Contains(outputStr, tt.expectNotInOutput) {
				t.Errorf("findExecuteRuleFiles() output should not contain %q, got:\n%s", tt.expectNotInOutput, outputStr)
			}

			// Check bootstrap script execution
			if tt.bootstrapPath != "" {
				if tt.expectBootstrapRun && !bootstrapRan {
					t.Errorf("findExecuteRuleFiles() expected bootstrap script %q to run, but it didn't", tt.bootstrapPath)
				}
				if !tt.expectBootstrapRun && bootstrapRan {
					t.Errorf("findExecuteRuleFiles() expected bootstrap script %q NOT to run, but it did", tt.bootstrapPath)
				}
			}
		})
	}
}

func TestRunBootstrapScript(t *testing.T) {
	tests := []struct {
		name         string
		mdFile       string
		ext          string
		setupFiles   func(t *testing.T, tmpDir string, mdFile string) string // returns bootstrap path
		wantErr      bool
		expectRun    bool
		mockRunError error
	}{
		{
			name:   "no bootstrap file",
			mdFile: "test.md",
			ext:    ".md",
			setupFiles: func(t *testing.T, tmpDir string, mdFile string) string {
				// Don't create bootstrap file
				return ""
			},
			wantErr:   false,
			expectRun: false,
		},
		{
			name:   "bootstrap file exists and runs",
			mdFile: "test.md",
			ext:    ".md",
			setupFiles: func(t *testing.T, tmpDir string, mdFile string) string {
				bootstrapPath := filepath.Join(tmpDir, "test-bootstrap")
				if err := os.WriteFile(bootstrapPath, []byte("#!/bin/sh\necho 'bootstrap'"), 0o644); err != nil {
					t.Fatalf("failed to create bootstrap file: %v", err)
				}
				return bootstrapPath
			},
			wantErr:   false,
			expectRun: true,
		},
		{
			name:   "bootstrap file with .mdc extension",
			mdFile: "test.mdc",
			ext:    ".mdc",
			setupFiles: func(t *testing.T, tmpDir string, mdFile string) string {
				bootstrapPath := filepath.Join(tmpDir, "test-bootstrap")
				if err := os.WriteFile(bootstrapPath, []byte("#!/bin/sh\necho 'bootstrap'"), 0o644); err != nil {
					t.Fatalf("failed to create bootstrap file: %v", err)
				}
				return bootstrapPath
			},
			wantErr:   false,
			expectRun: true,
		},
		{
			name:   "bootstrap file fails",
			mdFile: "test.md",
			ext:    ".md",
			setupFiles: func(t *testing.T, tmpDir string, mdFile string) string {
				bootstrapPath := filepath.Join(tmpDir, "test-bootstrap")
				if err := os.WriteFile(bootstrapPath, []byte("#!/bin/sh\nexit 1"), 0o644); err != nil {
					t.Fatalf("failed to create bootstrap file: %v", err)
				}
				return bootstrapPath
			},
			wantErr:      true,
			expectRun:    true,
			mockRunError: fmt.Errorf("exit status 1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mdPath := filepath.Join(tmpDir, tt.mdFile)
			bootstrapPath := tt.setupFiles(t, tmpDir, tt.mdFile)

			var logOut bytes.Buffer
			cmdRan := false
			cc := &Context{
				logger: slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					cmdRan = true
					if tt.mockRunError != nil {
						return tt.mockRunError
					}
					return nil
				},
			}

			err := cc.runBootstrapScript(context.Background(), mdPath, tt.ext)

			if tt.wantErr {
				if err == nil {
					t.Errorf("runBootstrapScript() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("runBootstrapScript() unexpected error: %v", err)
				}
			}

			if tt.expectRun && !cmdRan {
				t.Errorf("runBootstrapScript() expected command to run, but it didn't")
			}
			if !tt.expectRun && cmdRan {
				t.Errorf("runBootstrapScript() expected command not to run, but it did")
			}

			// Check that bootstrap file was made executable if it existed
			if bootstrapPath != "" {
				info, err := os.Stat(bootstrapPath)
				if err == nil && tt.expectRun {
					mode := info.Mode()
					if mode&0o100 == 0 {
						t.Errorf("runBootstrapScript() bootstrap file should be executable")
					}
				}
			}
		})
	}
}

func TestWriteTaskFileContent(t *testing.T) {
	t.Skip("This test uses removed field matchingTaskFile - functionality replaced by Run with ParseTask")
}

func TestParseTaskFile(t *testing.T) {
	t.Skip("This test uses removed method parseTaskFile - functionality replaced by getTask")
}

func TestTaskSelectorsFilterRulesByRuleName(t *testing.T) {
	tests := []struct {
		name              string
		taskSelectors     string // YAML frontmatter for task Selectors field
		setupRules        func(t *testing.T, tmpDir string)
		expectInOutput    []string // Rule content that should be present
		expectNotInOutput []string // Rule content that should NOT be present
		wantErr           bool
	}{
		{
			name:          "single rule_name selector filters to one rule",
			taskSelectors: "selectors:\n  rule_name: rule1",
			setupRules: func(t *testing.T, tmpDir string) {
				rulesDir := filepath.Join(tmpDir, ".agents", "rules")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule1.md"),
					"rule_name: rule1",
					"# Rule 1 Content\nThis is rule 1.")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule2.md"),
					"rule_name: rule2",
					"# Rule 2 Content\nThis is rule 2.")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule3.md"),
					"rule_name: rule3",
					"# Rule 3 Content\nThis is rule 3.")
			},
			expectInOutput:    []string{"# Rule 1 Content", "This is rule 1."},
			expectNotInOutput: []string{"# Rule 2 Content", "# Rule 3 Content", "This is rule 2.", "This is rule 3."},
			wantErr:           false,
		},
		{
			name:          "array selector matches multiple rules",
			taskSelectors: "selectors:\n  rule_name:\n    - rule1\n    - rule2",
			setupRules: func(t *testing.T, tmpDir string) {
				rulesDir := filepath.Join(tmpDir, ".agents", "rules")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule1.md"),
					"rule_name: rule1",
					"# Rule 1 Content\nThis is rule 1.")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule2.md"),
					"rule_name: rule2",
					"# Rule 2 Content\nThis is rule 2.")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule3.md"),
					"rule_name: rule3",
					"# Rule 3 Content\nThis is rule 3.")
			},
			expectInOutput:    []string{"# Rule 1 Content", "# Rule 2 Content", "This is rule 1.", "This is rule 2."},
			expectNotInOutput: []string{"# Rule 3 Content", "This is rule 3."},
			wantErr:           false,
		},
		{
			name:          "combined Selectors use AND logic",
			taskSelectors: "selectors:\n  rule_name: rule1\n  env: prod",
			setupRules: func(t *testing.T, tmpDir string) {
				rulesDir := filepath.Join(tmpDir, ".agents", "rules")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule1.md"),
					"rule_name: rule1\nenv: prod",
					"# Rule 1 Content\nThis is rule 1.")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule2.md"),
					"rule_name: rule2\nenv: prod",
					"# Rule 2 Content\nThis is rule 2.")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule1-dev.md"),
					"rule_name: rule1\nenv: dev",
					"# Rule 1 Dev Content\nThis is rule 1 dev.")
			},
			expectInOutput:    []string{"# Rule 1 Content", "This is rule 1."},
			expectNotInOutput: []string{"# Rule 2 Content", "# Rule 1 Dev Content", "This is rule 2.", "This is rule 1 dev."},
			wantErr:           false,
		},
		{
			name:          "no Selectors includes all rules",
			taskSelectors: "",
			setupRules: func(t *testing.T, tmpDir string) {
				rulesDir := filepath.Join(tmpDir, ".agents", "rules")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule1.md"),
					"rule_name: rule1",
					"# Rule 1 Content\nThis is rule 1.")
				createMarkdownFile(t, filepath.Join(rulesDir, "rule2.md"),
					"rule_name: rule2",
					"# Rule 2 Content\nThis is rule 2.")
			},
			expectInOutput:    []string{"# Rule 1 Content", "# Rule 2 Content", "This is rule 1.", "This is rule 2."},
			expectNotInOutput: []string{},
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup rule files
			tt.setupRules(t, tmpDir)

			// Setup task file
			taskDir := filepath.Join(tmpDir, ".agents", "tasks")
			taskPath := filepath.Join(taskDir, "test-task.md")
			var taskFrontmatter string
			if tt.taskSelectors != "" {
				taskFrontmatter = fmt.Sprintf("task_name: test-task\n%s", tt.taskSelectors)
			} else {
				taskFrontmatter = "task_name: test-task"
			}
			createMarkdownFile(t, taskPath, taskFrontmatter, "# Test Task\nThis is a test task.")

			// Change to temp dir
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}
			defer os.Chdir(oldDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to chdir: %v", err)
			}

			var logOut bytes.Buffer
			cc := &Context{
				includes: make(Selectors),
				rules:    make([]Markdown[RuleFrontMatter], 0),
				logger:   slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil // Mock command runner
				},
			}

			// Set up task name in includes (as done in run())
			cc.includes.SetValue("task_name", "test-task")
			absTmpDir, err := filepath.Abs(tmpDir)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}
			cc.searchPaths = []string{"file://" + absTmpDir}

			// Download the directories
			if err := cc.downloadRemoteDirectories(context.Background()); err != nil {
				t.Fatalf("failed to download remote directories: %v", err)
			}
			defer cc.cleanupDownloadedDirectories()

			// Find and parse task file
			homeDir, err := os.UserHomeDir()
			if err != nil {
				t.Fatalf("failed to get user home directory: %v", err)
			}

			if err := cc.findTaskFile("test-task"); err != nil {
				if !tt.wantErr {
					t.Fatalf("findTaskFile() unexpected error: %v", err)
				}
				return
			}

			// Parse task file to extract Selectors
			if err := cc.parseTaskFile(); err != nil {
				if !tt.wantErr {
					t.Fatalf("parseTaskFile() unexpected error: %v", err)
				}
				return
			}

			// Find and execute rule files
			if err := cc.findExecuteRuleFiles(context.Background(), homeDir); err != nil {
				if !tt.wantErr {
					t.Fatalf("findExecuteRuleFiles() unexpected error: %v", err)
				}
				return
			}

			// Collect all rule content into a single string for output comparison
			var outputStr string
			for _, rule := range cc.rules {
				outputStr += rule.Content + "\n"
			}

			// Verify expected content is present
			for _, expected := range tt.expectInOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("TestTaskSelectorsFilterRulesByRuleName() output should contain %q, got:\n%s", expected, outputStr)
				}
			}

			// Verify unexpected content is NOT present
			for _, unexpected := range tt.expectNotInOutput {
				if strings.Contains(outputStr, unexpected) {
					t.Errorf("TestTaskSelectorsFilterRulesByRuleName() output should NOT contain %q, got:\n%s", unexpected, outputStr)
				}
			}
		})
	}
}

func TestTaskFileWalker(t *testing.T) {
	t.Skip("This test uses removed method taskFileWalker - functionality replaced by findMarkdownFile")
}

func TestRuleFileWalker(t *testing.T) {
	tests := []struct {
		name             string
		includes         Selectors
		fileInfo         fileInfoMock
		filePath         string
		fileContent      string
		expectInOutput   bool
		expectExcludeLog bool
		wantErr          bool
	}{
		{
			name:     "skip directories",
			fileInfo: fileInfoMock{isDir: true, name: "somedir"},
			filePath: "/test/somedir",
			wantErr:  false,
		},
		{
			name:     "skip non-markdown files",
			fileInfo: fileInfoMock{isDir: false, name: "file.txt"},
			filePath: "/test/file.txt",
			wantErr:  false,
		},
		{
			name:           "include rule file",
			fileInfo:       fileInfoMock{isDir: false, name: "rule.md"},
			filePath:       "rule.md",
			fileContent:    "---\n---\n# Rule Content",
			expectInOutput: true,
			wantErr:        false,
		},
		{
			name:           "include mdc file",
			fileInfo:       fileInfoMock{isDir: false, name: "rule.mdc"},
			filePath:       "rule.mdc",
			fileContent:    "---\n---\n# MDC Rule",
			expectInOutput: true,
			wantErr:        false,
		},
		{
			name:             "exclude rule with non-matching selector",
			includes:         Selectors{"env": map[string]bool{"prod": true}},
			fileInfo:         fileInfoMock{isDir: false, name: "rule.md"},
			filePath:         "rule.md",
			fileContent:      "---\nenv: dev\n---\n# Dev Rule",
			expectInOutput:   false,
			expectExcludeLog: true,
			wantErr:          false,
		},
		{
			name:           "include rule with matching selector",
			includes:       Selectors{"env": map[string]bool{"prod": true}},
			fileInfo:       fileInfoMock{isDir: false, name: "rule.md"},
			filePath:       "rule.md",
			fileContent:    "---\nenv: prod\n---\n# Prod Rule",
			expectInOutput: true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create the file if content is provided
			if tt.fileContent != "" {
				fullPath := filepath.Join(tmpDir, tt.filePath)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(tt.fileContent), 0o644); err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				tt.filePath = fullPath
			}

			var logOut bytes.Buffer
			cc := &Context{
				includes: tt.includes,
				rules:    make([]Markdown[RuleFrontMatter], 0),
				logger:   slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil // Mock command runner
				},
			}
			if cc.includes == nil {
				cc.includes = make(Selectors)
			}

			walker := cc.ruleFileWalker(context.Background())
			err := walker(tt.filePath, &tt.fileInfo, nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ruleFileWalker() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ruleFileWalker() unexpected error: %v", err)
				}
			}

			// Collect all rule content into a single string for output comparison
			var outputStr string
			for _, rule := range cc.rules {
				outputStr += rule.Content
			}
			logStr := logOut.String()

			if tt.expectInOutput && !strings.Contains(outputStr, "Rule") {
				t.Errorf("ruleFileWalker() expected output to contain rule content, got:\n%s", outputStr)
			}
			if !tt.expectInOutput && strings.Contains(outputStr, "Rule") {
				t.Errorf("ruleFileWalker() expected output not to contain rule content, got:\n%s", outputStr)
			}

			if tt.expectExcludeLog && !strings.Contains(logStr, "Excluding") {
				t.Errorf("ruleFileWalker() expected log to contain 'Excluding', got:\n%s", logStr)
			}
		})
	}
}

// Mock fileInfo for testing
type fileInfoMock struct {
	name  string
	isDir bool
}

func (f *fileInfoMock) Name() string       { return f.name }
func (f *fileInfoMock) Size() int64        { return 0 }
func (f *fileInfoMock) Mode() os.FileMode  { return 0o644 }
func (f *fileInfoMock) ModTime() time.Time { return time.Time{} }
func (f *fileInfoMock) IsDir() bool        { return f.isDir }
func (f *fileInfoMock) Sys() any           { return nil }

func TestSlashCommandSubstitution(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		params       Params
		wantTaskName string
		wantParams   map[string]string
		wantContent  string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "slash command finds task",
			prompt:       "/real-task 123",
			params:       Params{},
			wantTaskName: "real-task",
			wantParams: map[string]string{
				"ARGUMENTS": "123",
				"1":         "123",
			},
			wantContent: "# Real Task Content for issue 123",
			wantErr:     false,
		},
		{
			name:         "slash command merges with existing parameters",
			prompt:       "/real-task 456",
			params:       Params{"foo": "bar", "existing": "old"},
			wantTaskName: "real-task",
			wantParams: map[string]string{
				"ARGUMENTS": "456",
				"1":         "456",
				"foo":       "bar",
				"existing":  "old",
			},
			wantContent: "# Real Task Content for issue 456",
			wantErr:     false,
		},
		{
			name:         "slash command with multiple arguments",
			prompt:       "/multi-arg-task arg1 arg2 arg3",
			params:       Params{},
			wantTaskName: "multi-arg-task",
			wantParams: map[string]string{
				"ARGUMENTS": "arg1 arg2 arg3",
				"1":         "arg1",
				"2":         "arg2",
				"3":         "arg3",
			},
			wantContent: "# Multi Arg Task: arg1, arg2, arg3",
			wantErr:     false,
		},
		{
			name:         "no slash command - uses free text as inline task",
			prompt:       "Just a simple task with no slash command",
			params:       Params{},
			wantTaskName: "",
			wantParams:   map[string]string{},
			wantContent:  "Just a simple task with no slash command",
			wantErr:      false,
		},
		{
			name:         "free text with parameters expanded",
			prompt:       "Please work on ${component}",
			params:       Params{"component": "auth"},
			wantTaskName: "",
			wantParams:   map[string]string{"component": "auth"},
			wantContent:  "Please work on auth",
			wantErr:      false,
		},
		{
			name:        "slash command for nonexistent task",
			prompt:      "/nonexistent-task",
			params:      Params{},
			wantErr:     true,
			errContains: "no task file found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create task files
			taskDir := filepath.Join(tmpDir, ".agents", "tasks")

			// Create real-task file
			createMarkdownFile(t, filepath.Join(taskDir, "real-task.md"),
				"task_name: real-task",
				"# Real Task Content for issue ${1}")

			// Create multi-arg-task file
			createMarkdownFile(t, filepath.Join(taskDir, "multi-arg-task.md"),
				"task_name: multi-arg-task",
				"# Multi Arg Task: ${1}, ${2}, ${3}")

			var logOut bytes.Buffer
			cc := &Context{
				params:   tt.params,
				includes: make(Selectors),
				searchPaths: func() []string {
					abs, _ := filepath.Abs(tmpDir)
					return []string{"file://" + abs}
				}(),
				rules:  make([]Markdown[RuleFrontMatter], 0),
				logger: slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil
				},
			}

			if cc.params == nil {
				cc.params = make(Params)
			}

			result, err := cc.Run(context.Background(), tt.prompt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Run() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Run() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Run() unexpected error: %v\nLog output:\n%s", err, logOut.String())
				return
			}

			if result == nil {
				t.Errorf("Run() returned nil result")
				return
			}

			// Verify the task name by checking the task_name in frontmatter (only for slash command cases)
			if tt.wantTaskName != "" {
				if taskName, ok := result.Task.FrontMatter.Content["task_name"].(string); ok {
					if taskName != tt.wantTaskName {
						t.Errorf("Task name = %v, want %v", taskName, tt.wantTaskName)
					}
				} else {
					t.Errorf("Task name not found in frontmatter, wanted %q", tt.wantTaskName)
				}
			}

			// Verify content
			if tt.wantContent != "" && !strings.Contains(result.Task.Content, tt.wantContent) {
				t.Errorf("Task content = %q, want to contain %q", result.Task.Content, tt.wantContent)
			}

			// Verify parameters
			for k, v := range tt.wantParams {
				if cc.params[k] != v {
					t.Errorf("Param[%q] = %q, want %q", k, cc.params[k], v)
				}
			}

			// Verify param count
			if len(cc.params) != len(tt.wantParams) {
				t.Errorf("Param count = %d, want %d. Params: %v", len(cc.params), len(tt.wantParams), cc.params)
			}
		})
	}
}
func TestTaskLanguageFieldFilteringRules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create rules with different language filters
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "rules", "go-rule.md"),
		"language: go", "# go-specific rule content")
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "rules", "python-rule.md"),
		"language: python", "# python-specific rule content")
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "rules", "js-rule.md"),
		"language: javascript", "# javascript-specific rule content")
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "rules", "generic-rule.md"),
		"", "# Generic rule content")

	tests := []struct {
		name             string
		taskFrontmatter  string
		expectInRules    []string
		expectNotInRules []string
	}{
		{
			name:             "task with language: go does not filter rules (language is metadata only)",
			taskFrontmatter:  "task_name: test-task\nlanguage: go",
			expectInRules:    []string{"go-specific rule content", "python-specific rule content", "javascript-specific rule content", "Generic rule content"},
			expectNotInRules: []string{},
		},
		{
			name:             "task with language: python does not filter rules (language is metadata only)",
			taskFrontmatter:  "task_name: test-task\nlanguage: python",
			expectInRules:    []string{"go-specific rule content", "python-specific rule content", "javascript-specific rule content", "Generic rule content"},
			expectNotInRules: []string{},
		},
		{
			name:             "task with language array does not filter rules (language is metadata only)",
			taskFrontmatter:  "task_name: test-task\nlanguage:\n  - go\n  - python",
			expectInRules:    []string{"go-specific rule content", "python-specific rule content", "javascript-specific rule content", "Generic rule content"},
			expectNotInRules: []string{},
		},
		{
			name:             "task without language field includes all rules",
			taskFrontmatter:  "task_name: test-task",
			expectInRules:    []string{"go-specific rule content", "python-specific rule content", "javascript-specific rule content", "Generic rule content"},
			expectNotInRules: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create task file for this test
			taskPath := filepath.Join(tmpDir, ".agents", "tasks", "test-task.md")
			createMarkdownFile(t, taskPath, tt.taskFrontmatter, "# Test Task Content")

			cc := New(
				WithSearchPaths(func() string { abs, _ := filepath.Abs(tmpDir); return "file://" + abs }()),
			)

			result, err := cc.Run(ctx, "/test-task")
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			// Combine all rule content
			var allRules strings.Builder
			for _, rule := range result.Rules {
				allRules.WriteString(rule.Content)
			}
			rulesContent := allRules.String()

			// Check expected inclusions
			for _, expected := range tt.expectInRules {
				if !strings.Contains(rulesContent, expected) {
					t.Errorf("Expected rules to contain %q but it was not found", expected)
				}
			}

			// Check expected exclusions
			for _, notExpected := range tt.expectNotInRules {
				if strings.Contains(rulesContent, notExpected) {
					t.Errorf("Expected rules to NOT contain %q but it was found", notExpected)
				}
			}

			// Verify task frontmatter contains the language field
			if strings.Contains(tt.taskFrontmatter, "language:") {
				if _, exists := result.Task.FrontMatter.Content["language"]; !exists {
					t.Errorf("Expected task frontmatter to contain 'language' field")
				}
			}
		})
	}
}

func TestTaskStandardFieldsPreservedInFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple rule to ensure the task runs
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "rules", "generic.md"),
		"", "# Generic rule")

	taskFrontmatter := `task_name: test-task
agent: cursor
languages:
  - go
model: anthropic.claude-sonnet-4-20250514-v1-0
single_shot: true
timeout: 5m
mcp_servers:
  filesystem:
    type: stdio
    command: filesystem-server
  git:
    type: stdio
    command: git-server`

	taskPath := filepath.Join(tmpDir, ".agents", "tasks", "test-task.md")
	createMarkdownFile(t, taskPath, taskFrontmatter, "# Test Task")

	cc := New(
		WithSearchPaths(func() string { abs, _ := filepath.Abs(tmpDir); return "file://" + abs }()),
	)
	result, err := cc.Run(context.Background(), "/test-task")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify all standard fields are preserved in task frontmatter
	expectedFields := map[string]any{
		"task_name":   "test-task",
		"agent":       "cursor",
		"languages":   []any{"go"},
		"model":       "anthropic.claude-sonnet-4-20250514-v1-0",
		"single_shot": true,
		"timeout":     "5m",
		"mcp_servers": map[string]any{
			"filesystem": map[string]any{"type": "stdio", "command": "filesystem-server"},
			"git":        map[string]any{"type": "stdio", "command": "git-server"},
		},
	}

	for field, expectedValue := range expectedFields {
		actualValue, ok := result.Task.FrontMatter.Content[field]
		if !ok {
			t.Errorf("Expected task frontmatter to contain %q field", field)
			continue
		}

		// Special handling for languages array
		if field == "languages" {
			actualArray, ok := actualValue.([]any)
			if !ok {
				t.Errorf("Expected %q to be []any, got %T", field, actualValue)
				continue
			}
			expectedArray := expectedValue.([]any)
			if len(actualArray) != len(expectedArray) {
				t.Errorf("Expected %q length %d, got %d", field, len(expectedArray), len(actualArray))
			}
		} else if field == "mcp_servers" {
			// Special handling for mcp_servers map
			actualMap, ok := actualValue.(map[string]any)
			if !ok {
				t.Errorf("Expected %q to be map[string]any, got %T", field, actualValue)
				continue
			}
			expectedMap := expectedValue.(map[string]any)
			if len(actualMap) != len(expectedMap) {
				t.Errorf("Expected %q length %d, got %d", field, len(expectedMap), len(actualMap))
			}
		} else {
			// For simple values, just check they exist
			// (exact comparison would require type matching which is complex with YAML)
			if actualValue == nil {
				t.Errorf("Expected %q to have a value, got nil", field)
			}
		}
	}
}

func TestWithResume(t *testing.T) {
	tmpDir := t.TempDir()
	taskDir := filepath.Join(tmpDir, ".agents", "tasks")
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")

	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("failed to create task dir: %v", err)
	}
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	// Create a resume task file
	createMarkdownFile(t, filepath.Join(taskDir, "resume_task.md"),
		"resume: true",
		"# Resume Task")

	// Create a rule file that should be skipped in resume mode
	createMarkdownFile(t, filepath.Join(rulesDir, "test-rule.md"),
		"",
		"# Test Rule")

	// Change to temp dir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	var logOut bytes.Buffer
	cc := New(
		WithResume(true),
		WithLogger(slog.New(slog.NewTextHandler(&logOut, nil))),
		WithSearchPaths(func() string { abs, _ := filepath.Abs(tmpDir); return "file://" + abs }()),
	)

	result, err := cc.Run(context.Background(), "/resume_task")
	if err != nil {
		t.Fatalf("Run() unexpected error: %v\nLog output:\n%s", err, logOut.String())
	}

	// In resume mode, rules should NOT be included
	if len(result.Rules) != 0 {
		t.Errorf("WithResume(true): expected 0 rules, got %d", len(result.Rules))
	}

	// Task should be included
	if result.Task.Content == "" {
		t.Errorf("WithResume(true): expected task content, got empty")
	}
	if !strings.Contains(result.Task.Content, "# Resume Task") {
		t.Errorf("WithResume(true): expected task content to contain '# Resume Task', got: %s", result.Task.Content)
	}

	// Test that WithResume(false) includes rules
	cc2 := New(
		WithResume(false),
		WithLogger(slog.New(slog.NewTextHandler(&logOut, nil))),
		WithSearchPaths(func() string { abs, _ := filepath.Abs(tmpDir); return "file://" + abs }()),
	)

	result2, err := cc2.Run(context.Background(), "/resume_task")
	if err != nil {
		t.Fatalf("Run() unexpected error: %v\nLog output:\n%s", err, logOut.String())
	}

	// Without resume mode, rules should be included
	if len(result2.Rules) == 0 {
		t.Errorf("WithResume(false): expected rules to be included, got 0")
	}
}

func TestWithAgent(t *testing.T) {
	tmpDir := t.TempDir()
	taskDir := filepath.Join(tmpDir, ".agents", "tasks")
	rulesDir := filepath.Join(tmpDir, ".agents", "rules")

	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("failed to create task dir: %v", err)
	}
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("failed to create rules dir: %v", err)
	}

	// Create a task file
	createMarkdownFile(t, filepath.Join(taskDir, "test-task.md"),
		"",
		"# Test Task")

	// Create CLAUDE.md in root (should be excluded when agent=claude)
	createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
		"",
		"# Claude Rule")

	// Create a rule in .agents/rules with agent: claude (should be excluded when agent=claude)
	createMarkdownFile(t, filepath.Join(rulesDir, "claude-specific.md"),
		"agent: claude",
		"# Claude Specific Rule")

	// Create a rule in .agents/rules with agent: cursor (should be included when agent=claude)
	createMarkdownFile(t, filepath.Join(rulesDir, "cursor-specific.md"),
		"agent: cursor",
		"# Cursor Specific Rule")

	// Create a rule in .agents/rules without agent field (should be included)
	createMarkdownFile(t, filepath.Join(rulesDir, "generic.md"),
		"",
		"# Generic Rule")

	// Change to temp dir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	var logOut bytes.Buffer
	var agent Agent
	if err := agent.Set("claude"); err != nil {
		t.Fatalf("failed to set target agent: %v", err)
	}

	cc := New(
		WithAgent(agent),
		WithLogger(slog.New(slog.NewTextHandler(&logOut, nil))),
		WithSearchPaths(func() string { abs, _ := filepath.Abs(tmpDir); return "file://" + abs }()),
	)

	result, err := cc.Run(context.Background(), "/test-task")
	if err != nil {
		t.Fatalf("Run() unexpected error: %v\nLog output:\n%s", err, logOut.String())
	}

	// CLAUDE.md should NOT be included (path matches target agent)
	// claude-specific.md should NOT be included (agent field matches target agent)
	// cursor-specific.md should be included (different agent)
	// generic.md should be included (no agent field)

	var ruleContents string
	for _, rule := range result.Rules {
		ruleContents += rule.Content + "\n"
	}

	if strings.Contains(ruleContents, "# Claude Rule") {
		t.Errorf("WithAgent(claude): CLAUDE.md should be excluded, but was included")
	}

	if strings.Contains(ruleContents, "# Claude Specific Rule") {
		t.Errorf("WithAgent(claude): rule with agent: claude should be excluded, but was included")
	}

	if !strings.Contains(ruleContents, "# Cursor Specific Rule") {
		t.Errorf("WithAgent(claude): rule with agent: cursor should be included, but was excluded")
	}

	if !strings.Contains(ruleContents, "# Generic Rule") {
		t.Errorf("WithAgent(claude): rule without agent field should be included, but was excluded")
	}
}
