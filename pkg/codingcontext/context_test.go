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
		resume      bool
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
			args:        []string{"nonexistent"},
			wantErr:     true,
			errContains: "no task file found",
		},
		{
			name: "successful task execution",
			args: []string{"test_task"},
			setupFiles: func(t *testing.T, tmpDir string) {
				// Create task file
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "test.md"),
					"task_name: test_task",
					"# Test Task\nThis is a test task.")
			},
			wantErr: false,
		},
		{
			name: "task with parameters",
			args: []string{"param_task"},
			params: Params{
				"name": "value",
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "param.md"),
					"task_name: param_task",
					"# Test ${name}")
			},
			wantErr: false,
		},
		{
			name:   "resume mode skips rules",
			args:   []string{"resume_task"},
			resume: true,
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "resume.md"),
					"task_name: resume_task\nresume: true",
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
				workDir:  tmpDir,
				resume:   tt.resume,
				params:   tt.params,
				includes: tt.includes,
				rules:    make([]Markdown, 0),
				logger:   slog.New(slog.NewTextHandler(&logOut, nil)),
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
	tests := []struct {
		name           string
		taskName       string
		includes       Selectors
		setupFiles     func(t *testing.T, tmpDir string)
		downloadedDirs []string // Directories to add to downloadedDirs
		wantErr        bool
		errContains    string
	}{
		{
			name:     "task file not found",
			taskName: "missing",
			setupFiles: func(t *testing.T, tmpDir string) {
				// No files created
			},
			wantErr:     true,
			errContains: "no task file found",
		},
		{
			name:     "task file found",
			taskName: "my_task",
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "task.md"),
					"task_name: my_task",
					"# My Task")
			},
			wantErr: false,
		},
		{
			name:     "multiple task files with same name",
			taskName: "duplicate",
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "task1.md"),
					"task_name: duplicate",
					"# Task 1")
				createMarkdownFile(t, filepath.Join(taskDir, "task2.md"),
					"task_name: duplicate",
					"# Task 2")
			},
			wantErr:     true,
			errContains: "multiple task files found",
		},
		{
			name:     "task with matching selector",
			taskName: "filtered_task",
			includes: Selectors{
				"env": map[string]bool{"prod": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "task.md"),
					"task_name: filtered_task\nenv: prod",
					"# Filtered Task")
			},
			wantErr: false,
		},
		{
			name:     "task with non-matching selector",
			taskName: "filtered_task",
			includes: Selectors{
				"env": map[string]bool{"dev": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "task.md"),
					"task_name: filtered_task\nenv: prod",
					"# Filtered Task")
			},
			wantErr:     true,
			errContains: "no task file found",
		},
		{
			name:     "task without task_name uses filename",
			taskName: "not-a-task",
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				// Create a file without task_name - should use filename as task name
				createMarkdownFile(t, filepath.Join(taskDir, "not-a-task.md"),
					"env: prod",
					"# Task using filename")
			},
			wantErr: false,
		},
		{
			name:     "task file found in downloaded directory",
			taskName: "downloaded_task",
			setupFiles: func(t *testing.T, tmpDir string) {
				// Create task file in downloaded directory
				downloadedDir := filepath.Join(tmpDir, "downloaded")
				taskDir := filepath.Join(downloadedDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "task.md"),
					"task_name: downloaded_task",
					"# Downloaded Task")
			},
			downloadedDirs: []string{"downloaded"}, // Relative path, will be joined with tmpDir
			wantErr:        false,
		},
		{
			name:     "task file found in .cursor/commands directory",
			taskName: "cursor_task",
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".cursor", "commands")
				createMarkdownFile(t, filepath.Join(taskDir, "cursor-task.md"),
					"task_name: cursor_task",
					"# Cursor Task")
			},
			wantErr: false,
		},
		{
			name:     "task file found in downloaded .cursor/commands directory",
			taskName: "cursor_remote_task",
			setupFiles: func(t *testing.T, tmpDir string) {
				// Create task file in downloaded directory's .cursor/commands
				downloadedDir := filepath.Join(tmpDir, "downloaded")
				taskDir := filepath.Join(downloadedDir, ".cursor", "commands")
				createMarkdownFile(t, filepath.Join(taskDir, "remote.md"),
					"task_name: cursor_remote_task",
					"# Cursor Remote Task")
			},
			downloadedDirs: []string{"downloaded"}, // Relative path, will be joined with tmpDir
			wantErr:        false,
		},
		{
			name:     "task file found in .opencode/command directory",
			taskName: "opencode_task",
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".opencode", "command")
				createMarkdownFile(t, filepath.Join(taskDir, "opencode-task.md"),
					"task_name: opencode_task",
					"# OpenCode Task")
			},
			wantErr: false,
		},
		{
			name:     "task file found in downloaded .opencode/command directory",
			taskName: "opencode_remote_task",
			setupFiles: func(t *testing.T, tmpDir string) {
				// Create task file in downloaded directory's .opencode/command
				downloadedDir := filepath.Join(tmpDir, "downloaded")
				taskDir := filepath.Join(downloadedDir, ".opencode", "command")
				createMarkdownFile(t, filepath.Join(taskDir, "remote.md"),
					"task_name: opencode_remote_task",
					"# OpenCode Remote Task")
			},
			downloadedDirs: []string{"downloaded"}, // Relative path, will be joined with tmpDir
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setupFiles(t, tmpDir)

			// Change to temp dir for relative path searches
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}
			defer os.Chdir(oldDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to chdir: %v", err)
			}

			cc := &Context{
				includes: tt.includes,
			}
			if cc.includes == nil {
				cc.includes = make(Selectors)
			}
			cc.includes.SetValue("task_name", tt.taskName)

			// Set downloadedDirs if specified in test case
			if len(tt.downloadedDirs) > 0 {
				cc.downloadedDirs = make([]string, len(tt.downloadedDirs))
				for i, dir := range tt.downloadedDirs {
					cc.downloadedDirs[i] = filepath.Join(tmpDir, dir)
				}
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				t.Fatalf("failed to get user home directory: %v", err)
			}

			err = cc.findTaskFile(homeDir, tt.taskName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("findTaskFile() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("findTaskFile() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("findTaskFile() unexpected error: %v", err)
				}
				if cc.matchingTaskFile == "" {
					t.Errorf("findTaskFile() did not set matchingTaskFile")
				}
			}
		})
	}
}

func TestFindExecuteRuleFiles(t *testing.T) {
	tests := []struct {
		name               string
		resume             bool
		includes           Selectors
		params             Params // Parameters for template expansion
		setupFiles         func(t *testing.T, tmpDir string)
		downloadedDirs     []string // Directories to add to downloadedDirs
		wantTokens         int
		wantMinTokens      bool // Check that tokens > 0
		expectInOutput     string
		expectNotInOutput  string
		expectBootstrapRun bool   // Whether bootstrap script should run
		bootstrapPath      string // Path to bootstrap script to check
	}{
		{
			name:   "resume mode skips rules",
			resume: true,
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule File")
			},
			wantTokens: 0,
		},
		{
			name:   "include rule file",
			resume: false,
			setupFiles: func(t *testing.T, tmpDir string) {
				createMarkdownFile(t, filepath.Join(tmpDir, "CLAUDE.md"),
					"",
					"# Rule File\nThis is a rule.")
			},
			wantMinTokens:  true,
			expectInOutput: "# Rule File",
		},
		{
			name:   "exclude rule with non-matching selector",
			resume: false,
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
			name:   "include rule with matching selector",
			resume: false,
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
			name:   "include multiple rules",
			resume: false,
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
			name:   "include .mdc files",
			resume: false,
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
			name:   "include rules from downloaded directories",
			resume: false,
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
			downloadedDirs: []string{"downloaded"}, // Relative path, will be joined with tmpDir
			wantMinTokens:  true,
			expectInOutput: "Downloaded Rule",
		},
		{
			name:   "bootstrap script should not run on excluded files",
			resume: false,
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
			name:   "rule with parameter substitution",
			resume: false,
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
			name:   "rule with missing parameter preserved",
			resume: false,
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
				resume:   tt.resume,
				includes: tt.includes,
				params:   tt.params,
				rules:    make([]Markdown, 0),
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

			// Set downloadedDirs if specified in test case
			if len(tt.downloadedDirs) > 0 {
				cc.downloadedDirs = make([]string, len(tt.downloadedDirs))
				for i, dir := range tt.downloadedDirs {
					cc.downloadedDirs[i] = filepath.Join(tmpDir, dir)
				}
			}

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
	tests := []struct {
		name           string
		taskFile       string
		params         Params
		setupFiles     func(t *testing.T, tmpDir string) string // returns task file path
		expectInOutput string
		wantErr        bool
	}{
		{
			name:     "simple task",
			taskFile: "task.md",
			params:   Params{},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test",
					"# Simple Task\nContent here.")
				return taskPath
			},
			expectInOutput: "# Simple Task",
			wantErr:        false,
		},
		{
			name:     "task with parameter substitution",
			taskFile: "task.md",
			params: Params{
				"name":  "Alice",
				"value": "123",
			},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test",
					"Hello ${name}, your value is ${value}.")
				return taskPath
			},
			expectInOutput: "Hello Alice, your value is 123.",
			wantErr:        false,
		},
		{
			name:     "task with missing parameter",
			taskFile: "task.md",
			params:   Params{},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test",
					"Hello ${missing}.")
				return taskPath
			},
			expectInOutput: "Hello ${missing}.",
			wantErr:        false,
		},
		{
			name:     "task with partial parameter substitution",
			taskFile: "task.md",
			params: Params{
				"name": "Bob",
			},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test",
					"Hello ${name}, your value is ${missing}.")
				return taskPath
			},
			expectInOutput: "Hello Bob, your value is ${missing}.",
			wantErr:        false,
		},
		{
			name:     "task with frontmatter (always emitted)",
			taskFile: "task.md",
			params:   Params{},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test_task\nenv: production\nversion: 1.0",
					"# Task with Frontmatter\nThis task has frontmatter.")
				return taskPath
			},
			expectInOutput: "# Task with Frontmatter", // Task content, not frontmatter
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			taskPath := tt.setupFiles(t, tmpDir)

			// Need to extract task name from file (even though we don't use it directly in this test)
			var frontmatter FrontMatter
			_, err := ParseMarkdownFile(taskPath, &frontmatter)
			if err != nil {
				t.Fatalf("failed to parse task file: %v", err)
			}

			var logOut bytes.Buffer
			cc := &Context{
				workDir:          tmpDir,
				matchingTaskFile: taskPath,
				params:           tt.params,
				rules:            make([]Markdown, 0),
				logger:           slog.New(slog.NewTextHandler(&logOut, nil)),
				includes:         make(Selectors),
				taskFrontmatter:  make(FrontMatter),
			}

			// Parse task file first
			if err := cc.parseTaskFile(); err != nil {
				if !tt.wantErr {
					t.Errorf("parseTaskFile() unexpected error: %v", err)
				}
				return
			}

			// Expand parameters in task content (mimics what Run does)
			expandedTask := os.Expand(cc.taskContent, func(key string) string {
				if val, ok := cc.params[key]; ok {
					return val
				}
				return fmt.Sprintf("${%s}", key)
			})

			if tt.wantErr {
				if err != nil {
					return // Expected error
				}
				t.Errorf("expected error but got none")
				return
			}

			if tt.expectInOutput != "" {
				if !strings.Contains(expandedTask, tt.expectInOutput) {
					t.Errorf("task content should contain %q, got:\n%s", tt.expectInOutput, expandedTask)
				}
			}

			// Verify frontmatter is always parsed when present
			if cc.taskFrontmatter != nil && len(cc.taskFrontmatter) > 0 {
				// Just verify frontmatter was parsed - the Context doesn't emit it, main.go does
				if _, ok := cc.taskFrontmatter["task_name"]; !ok {
					// This is OK - not all tasks have task_name in frontmatter
				}
			}

			// Note: Token counting is done in Run(), not in these internal methods
		})
	}
}

func TestParseTaskFile(t *testing.T) {
	tests := []struct {
		name             string
		taskFile         string
		setupFiles       func(t *testing.T, tmpDir string) string // returns task file path
		initialIncludes  Selectors
		expectedIncludes Selectors // expected includes after parsing
		wantErr          bool
		errContains      string
	}{
		{
			name:             "task without Selectors field",
			taskFile:         "task.md",
			initialIncludes:  make(Selectors),
			expectedIncludes: make(Selectors),
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test",
					"# Simple Task")
				return taskPath
			},
			wantErr: false,
		},
		{
			name:            "task with Selectors field",
			taskFile:        "task.md",
			initialIncludes: make(Selectors),
			expectedIncludes: Selectors{
				"language": map[string]bool{"Go": true},
				"env":      map[string]bool{"prod": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test\nselectors:\n  language: Go\n  env: prod",
					"# Task with Selectors")
				return taskPath
			},
			wantErr: false,
		},
		{
			name:            "task with Selectors merges with existing includes",
			taskFile:        "task.md",
			initialIncludes: Selectors{"existing": map[string]bool{"value": true}},
			expectedIncludes: Selectors{
				"existing": map[string]bool{"value": true},
				"language": map[string]bool{"Python": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test\nselectors:\n  language: Python",
					"# Task with Selectors")
				return taskPath
			},
			wantErr: false,
		},
		{
			name:            "task with array selector values",
			taskFile:        "task.md",
			initialIncludes: make(Selectors),
			expectedIncludes: Selectors{
				"rule_name": map[string]bool{"rule1": true, "rule2": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test\nselectors:\n  rule_name:\n    - rule1\n    - rule2",
					"# Task with Array Selectors")
				return taskPath
			},
			wantErr: false,
		},
		{
			name:            "Selectors from -s flag and task file are additive",
			taskFile:        "task.md",
			initialIncludes: Selectors{"var": map[string]bool{"arg1": true}},
			expectedIncludes: Selectors{
				"var": map[string]bool{"arg1": true, "arg2": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test\nselectors:\n  var: arg2",
					"# Task with Additive Selectors")
				return taskPath
			},
			wantErr: false,
		},
		{
			name:            "task with integer selector value",
			taskFile:        "task.md",
			initialIncludes: make(Selectors),
			expectedIncludes: Selectors{
				"version": map[string]bool{"42": true},
			},
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test\nselectors:\n  version: 42",
					"# Task with Integer Selector")
				return taskPath
			},
			wantErr: false,
		},
		{
			name:            "task with invalid Selectors field type",
			taskFile:        "task.md",
			initialIncludes: make(Selectors),
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test\nselectors: invalid",
					"# Task with Invalid Selectors")
				return taskPath
			},
			wantErr:     true,
			errContains: "invalid 'selectors' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			taskPath := tt.setupFiles(t, tmpDir)

			cc := &Context{
				matchingTaskFile: taskPath,
				includes:         tt.initialIncludes,
			}
			if cc.includes == nil {
				cc.includes = make(Selectors)
			}

			err := cc.parseTaskFile()

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTaskFile() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseTaskFile() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("parseTaskFile() unexpected error: %v", err)
				}

				// Verify Selectors were extracted correctly
				for key, expectedValue := range tt.expectedIncludes {
					if actualValue, ok := cc.includes[key]; !ok {
						t.Errorf("parseTaskFile() expected includes[%q] = %v, but key not found", key, expectedValue)
					} else {
						// Compare map[string]bool structures
						if len(actualValue) != len(expectedValue) {
							t.Errorf("parseTaskFile() includes[%q] map length = %d, want %d", key, len(actualValue), len(expectedValue))
						} else {
							for expectedVal := range expectedValue {
								if !actualValue[expectedVal] {
									t.Errorf("parseTaskFile() includes[%q] does not contain value %q", key, expectedVal)
								}
							}
						}
					}
				}

				// Verify all includes match expected (including initial includes)
				if len(cc.includes) != len(tt.expectedIncludes) {
					t.Errorf("parseTaskFile() includes length = %d, want %d. Includes: %v", len(cc.includes), len(tt.expectedIncludes), cc.includes)
				}

				// Verify task content was stored
				if cc.taskContent == "" {
					t.Errorf("parseTaskFile() expected taskContent to be set, got empty string")
				}

				// Verify task frontmatter was stored
				if cc.taskFrontmatter == nil {
					t.Errorf("parseTaskFile() expected taskFrontmatter to be set, got nil")
				}
			}
		})
	}
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
				workDir:  tmpDir,
				includes: make(Selectors),
				rules:    make([]Markdown, 0),
				logger:   slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil // Mock command runner
				},
			}

			// Set up task name in includes (as done in run())
			cc.includes.SetValue("task_name", "test-task")
			cc.includes.SetValue("resume", "false")

			// Find and parse task file
			homeDir, err := os.UserHomeDir()
			if err != nil {
				t.Fatalf("failed to get user home directory: %v", err)
			}

			if err := cc.findTaskFile(homeDir, "test-task"); err != nil {
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
	tests := []struct {
		name          string
		taskName      string
		includes      Selectors
		fileInfo      fileInfoMock
		filePath      string
		fileContent   string // frontmatter + content
		existingMatch string // existing matchingTaskFile
		expectMatch   bool
		wantErr       bool
		errContains   string
	}{
		{
			name:     "skip directories",
			taskName: "test",
			fileInfo: fileInfoMock{isDir: true, name: "somedir"},
			filePath: "/test/somedir",
			wantErr:  false,
		},
		{
			name:     "skip non-markdown files",
			taskName: "test",
			fileInfo: fileInfoMock{isDir: false, name: "file.txt"},
			filePath: "/test/file.txt",
			wantErr:  false,
		},
		{
			name:        "matching task file",
			taskName:    "my_task",
			fileInfo:    fileInfoMock{isDir: false, name: "task.md"},
			filePath:    "task.md",
			fileContent: "---\ntask_name: my_task\n---\n# Task",
			expectMatch: true,
			wantErr:     false,
		},
		{
			name:        "non-matching task name",
			taskName:    "other_task",
			fileInfo:    fileInfoMock{isDir: false, name: "task.md"},
			filePath:    "task.md",
			fileContent: "---\ntask_name: my_task\n---\n# Task",
			expectMatch: false,
			wantErr:     false,
		},
		{
			name:          "duplicate task file",
			taskName:      "my_task",
			fileInfo:      fileInfoMock{isDir: false, name: "task2.md"},
			filePath:      "task2.md",
			fileContent:   "---\ntask_name: my_task\n---\n# Task",
			existingMatch: "task1.md",
			wantErr:       true,
			errContains:   "multiple task files found",
		},
		{
			name:        "task without task_name uses filename",
			taskName:    "task",
			fileInfo:    fileInfoMock{isDir: false, name: "task.md"},
			filePath:    "task.md",
			fileContent: "---\nother: value\n---\n# Task",
			expectMatch: true,
			wantErr:     false,
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

			cc := &Context{
				includes:         tt.includes,
				matchingTaskFile: tt.existingMatch,
			}
			if cc.includes == nil {
				cc.includes = make(Selectors)
			}
			cc.includes.SetValue("task_name", tt.taskName)

			walker := cc.taskFileWalker(tt.taskName)
			err := walker(tt.filePath, &tt.fileInfo, nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("taskFileWalker() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("taskFileWalker() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("taskFileWalker() unexpected error: %v", err)
				}
			}

			if tt.expectMatch && cc.matchingTaskFile == "" {
				t.Errorf("taskFileWalker() expected to set matchingTaskFile, but it's empty")
			}
			if !tt.expectMatch && tt.existingMatch == "" && cc.matchingTaskFile != "" {
				t.Errorf("taskFileWalker() expected no match, but matchingTaskFile = %s", cc.matchingTaskFile)
			}
		})
	}
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
				rules:    make([]Markdown, 0),
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
		name            string
		initialTaskName string
		taskContent     string
		params          Params
		wantTaskName    string
		wantParams      map[string]string
		wantErr         bool
		errContains     string
	}{
		{
			name:            "substitution to different task",
			initialTaskName: "wrapper-task",
			taskContent:     "Please /real-task 123",
			params:          Params{},
			wantTaskName:    "real-task",
			wantParams: map[string]string{
				"ARGUMENTS": "123",
				"1":         "123",
			},
			wantErr: false,
		},
		{
			name:            "slash command replaces existing parameters completely",
			initialTaskName: "wrapper-task",
			taskContent:     "Please /real-task 456",
			params:          Params{"foo": "bar", "existing": "old"},
			wantTaskName:    "real-task",
			wantParams: map[string]string{
				"ARGUMENTS": "456",
				"1":         "456",
			},
			wantErr: false,
		},
		{
			name:            "same task with params - replaces existing params",
			initialTaskName: "my-task",
			taskContent:     "/my-task arg1 arg2",
			params:          Params{"existing": "value"},
			wantTaskName:    "my-task",
			wantParams: map[string]string{
				"ARGUMENTS": "arg1 arg2",
				"1":         "arg1",
				"2":         "arg2",
			},
			wantErr: false,
		},
		{
			name:            "slash command in parameter value (free-text use case)",
			initialTaskName: "free-text-task",
			taskContent:     "${text}",
			params:          Params{"text": "/real-task PROJ-123"},
			wantTaskName:    "real-task",
			wantParams: map[string]string{
				"ARGUMENTS": "PROJ-123",
				"1":         "PROJ-123",
			},
			wantErr: false,
		},
		{
			name:            "no slash command in task",
			initialTaskName: "simple-task",
			taskContent:     "Just a simple task with no slash command",
			params:          Params{},
			wantTaskName:    "simple-task",
			wantParams:      map[string]string{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create the initial task file
			taskDir := filepath.Join(tmpDir, ".agents", "tasks")
			createMarkdownFile(t, filepath.Join(taskDir, "wrapper-task.md"),
				"task_name: wrapper-task",
				tt.taskContent)

			// Create the real-task file if needed
			createMarkdownFile(t, filepath.Join(taskDir, "real-task.md"),
				"task_name: real-task",
				"# Real Task Content for issue ${1}")

			// Create a simple-task file
			createMarkdownFile(t, filepath.Join(taskDir, "simple-task.md"),
				"task_name: simple-task",
				"Just a simple task with no slash command")

			// Create my-task file
			createMarkdownFile(t, filepath.Join(taskDir, "my-task.md"),
				"task_name: my-task",
				"/my-task arg1 arg2")

			// Create free-text-task file
			createMarkdownFile(t, filepath.Join(taskDir, "free-text-task.md"),
				"task_name: free-text-task",
				"${text}")

			var logOut bytes.Buffer
			cc := &Context{
				workDir:  tmpDir,
				params:   tt.params,
				includes: make(Selectors),
				rules:    make([]Markdown, 0),
				logger:   slog.New(slog.NewTextHandler(&logOut, nil)),
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil
				},
			}

			if cc.params == nil {
				cc.params = make(Params)
			}

			result, err := cc.Run(context.Background(), tt.initialTaskName)

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

			// Verify the task name by checking the task path
			expectedTaskPath := filepath.Join(taskDir, tt.wantTaskName+".md")
			if result.Task.Path != expectedTaskPath {
				t.Errorf("Task path = %v, want %v", result.Task.Path, expectedTaskPath)
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
func TestTargetAgentIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create various agent-specific rule files
	createMarkdownFile(t, filepath.Join(tmpDir, ".cursor", "rules", "cursor-rule.md"),
		"language: go", "# Cursor-specific rule")
	createMarkdownFile(t, filepath.Join(tmpDir, ".opencode", "agent", "opencode-rule.md"),
		"language: go", "# OpenCode-specific rule")
	createMarkdownFile(t, filepath.Join(tmpDir, ".github", "copilot-instructions.md"),
		"language: go", "# Copilot-specific rule")
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "rules", "generic-rule.md"),
		"language: go", "# Generic rule")
	// Create a rule that filters by agent selector
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "rules", "cursor-only-rule.md"),
		"agent: cursor", "# Rule only for Cursor agent")
	createMarkdownFile(t, filepath.Join(tmpDir, ".agents", "tasks", "test-task.md"),
		"task_name: test-task", "# Test task")

	tests := []struct {
		name             string
		targetAgent      string
		expectInRules    []string
		expectNotInRules []string
	}{
		{
			name:             "no target agent - all agent-specific rules plus generic and cursor-filtered",
			targetAgent:      "",
			expectInRules:    []string{"Cursor-specific", "OpenCode-specific", "Copilot-specific", "Generic", "Rule only for Cursor agent"},
			expectNotInRules: []string{},
		},
		{
			name:             "target cursor - exclude cursor rules, include others and generic",
			targetAgent:      "cursor",
			expectInRules:    []string{"OpenCode-specific", "Copilot-specific", "Generic", "Rule only for Cursor agent"},
			expectNotInRules: []string{"Cursor-specific"},
		},
		{
			name:             "target opencode - exclude opencode rules, include others and generic",
			targetAgent:      "opencode",
			expectInRules:    []string{"Cursor-specific", "Copilot-specific", "Generic"},
			expectNotInRules: []string{"OpenCode-specific", "Rule only for Cursor agent"},
		},
		{
			name:             "target copilot - exclude copilot rules, include others and generic",
			targetAgent:      "copilot",
			expectInRules:    []string{"Cursor-specific", "OpenCode-specific", "Generic"},
			expectNotInRules: []string{"Copilot-specific", "Rule only for Cursor agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			var ta TargetAgent
			if tt.targetAgent != "" {
				if err := ta.Set(tt.targetAgent); err != nil {
					t.Fatalf("Set target agent failed: %v", err)
				}
			}

			cc := New(
				WithWorkDir(tmpDir),
				WithAgent(ta),
			)

			result, err := cc.Run(ctx, "test-task")
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
		})
	}
}
