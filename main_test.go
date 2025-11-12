package main

import (
	"bytes"
	"context"
	"fmt"
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
		params      paramMap
		includes    selectorMap
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
			params: paramMap{
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

			var output, logOut bytes.Buffer
			cc := &codingContext{
				workDir:  tmpDir,
				resume:   tt.resume,
				params:   tt.params,
				includes: tt.includes,
				output:   &output,
				logOut:   &logOut,
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil // Mock command runner
				},
			}

			if cc.params == nil {
				cc.params = make(paramMap)
			}
			if cc.includes == nil {
				cc.includes = make(selectorMap)
			}

			err = cc.run(context.Background(), tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("run() expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("run() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("run() unexpected error: %v\nLog output:\n%s", err, logOut.String())
				}
			}
		})
	}
}

func TestFindTaskFile(t *testing.T) {
	tests := []struct {
		name           string
		taskName       string
		includes       selectorMap
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
			includes: selectorMap{
				"env": "prod",
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
			includes: selectorMap{
				"env": "dev",
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
			name:     "task missing task_name field",
			taskName: "my_task",
			setupFiles: func(t *testing.T, tmpDir string) {
				taskDir := filepath.Join(tmpDir, ".agents", "tasks")
				createMarkdownFile(t, filepath.Join(taskDir, "task.md"),
					"env: prod",
					"# Task without name")
			},
			wantErr:     true,
			errContains: "missing required 'task_name' field",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setupFiles(t, tmpDir)

			cc := &codingContext{
				includes: tt.includes,
			}
			if cc.includes == nil {
				cc.includes = make(selectorMap)
			}
			cc.includes["task_name"] = tt.taskName

			// Set downloadedDirs if specified in test case
			if len(tt.downloadedDirs) > 0 {
				cc.downloadedDirs = make([]string, len(tt.downloadedDirs))
				for i, dir := range tt.downloadedDirs {
					cc.downloadedDirs[i] = filepath.Join(tmpDir, dir)
				}
			}

			err := cc.findTaskFile(tmpDir, tt.taskName)

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
		name              string
		resume            bool
		includes          selectorMap
		setupFiles        func(t *testing.T, tmpDir string)
		downloadedDirs    []string // Directories to add to downloadedDirs
		wantTokens        int
		wantMinTokens     bool // Check that tokens > 0
		expectInOutput    string
		expectNotInOutput string
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
			includes: selectorMap{
				"env": "prod",
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
			includes: selectorMap{
				"env": "prod",
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

			var output, logOut bytes.Buffer
			cc := &codingContext{
				resume:   tt.resume,
				includes: tt.includes,
				output:   &output,
				logOut:   &logOut,
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil // Mock command runner
				},
			}
			if cc.includes == nil {
				cc.includes = make(selectorMap)
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

			outputStr := output.String()
			if tt.expectInOutput != "" && !strings.Contains(outputStr, tt.expectInOutput) {
				t.Errorf("findExecuteRuleFiles() output should contain %q, got:\n%s", tt.expectInOutput, outputStr)
			}
			if tt.expectNotInOutput != "" && strings.Contains(outputStr, tt.expectNotInOutput) {
				t.Errorf("findExecuteRuleFiles() output should not contain %q, got:\n%s", tt.expectNotInOutput, outputStr)
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
			cc := &codingContext{
				logOut: &logOut,
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
		name                string
		taskFile            string
		params              paramMap
		emitTaskFrontmatter bool
		setupFiles          func(t *testing.T, tmpDir string) string // returns task file path
		expectInOutput      string
		wantErr             bool
	}{
		{
			name:     "simple task",
			taskFile: "task.md",
			params:   paramMap{},
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
			params: paramMap{
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
			params:   paramMap{},
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
			params: paramMap{
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
			name:                "task with frontmatter emission enabled",
			taskFile:            "task.md",
			params:              paramMap{},
			emitTaskFrontmatter: true,
			setupFiles: func(t *testing.T, tmpDir string) string {
				taskPath := filepath.Join(tmpDir, "task.md")
				createMarkdownFile(t, taskPath,
					"task_name: test_task\nenv: production\nversion: 1.0",
					"# Task with Frontmatter\nThis task has frontmatter.")
				return taskPath
			},
			expectInOutput: "task_name: test_task",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			taskPath := tt.setupFiles(t, tmpDir)

			var output, logOut bytes.Buffer
			cc := &codingContext{
				matchingTaskFile:    taskPath,
				params:              tt.params,
				emitTaskFrontmatter: tt.emitTaskFrontmatter,
				output:              &output,
				logOut:              &logOut,
			}

			err := cc.writeTaskFileContent()

			if tt.wantErr {
				if err == nil {
					t.Errorf("writeTaskFileContent() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("writeTaskFileContent() unexpected error: %v", err)
				}
			}

			outputStr := output.String()
			if tt.expectInOutput != "" {
				if !strings.Contains(outputStr, tt.expectInOutput) {
					t.Errorf("writeTaskFileContent() output should contain %q, got:\n%s", tt.expectInOutput, outputStr)
				}
			}

			// Additional checks for frontmatter emission
			if tt.emitTaskFrontmatter {
				// Verify frontmatter delimiters are present
				if !strings.Contains(outputStr, "---") {
					t.Errorf("writeTaskFileContent() with emitTaskFrontmatter=true should contain '---' delimiters, got:\n%s", outputStr)
				}
				// Verify YAML frontmatter structure
				if !strings.Contains(outputStr, "task_name:") {
					t.Errorf("writeTaskFileContent() with emitTaskFrontmatter=true should contain 'task_name:' field, got:\n%s", outputStr)
				}
				// Verify task content is still present
				if !strings.Contains(outputStr, "# Task with Frontmatter") {
					t.Errorf("writeTaskFileContent() should contain task content, got:\n%s", outputStr)
				}
			}

			if !tt.wantErr && cc.totalTokens <= 0 {
				t.Errorf("writeTaskFileContent() expected tokens > 0, got %d", cc.totalTokens)
			}
		})
	}
}

func TestTaskFileWalker(t *testing.T) {
	tests := []struct {
		name          string
		taskName      string
		includes      selectorMap
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
			name:        "task missing task_name",
			taskName:    "test",
			fileInfo:    fileInfoMock{isDir: false, name: "task.md"},
			filePath:    "task.md",
			fileContent: "---\nother: value\n---\n# Task",
			wantErr:     true,
			errContains: "missing required 'task_name' field",
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

			cc := &codingContext{
				includes:         tt.includes,
				matchingTaskFile: tt.existingMatch,
			}
			if cc.includes == nil {
				cc.includes = make(selectorMap)
			}
			cc.includes["task_name"] = tt.taskName

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
		includes         selectorMap
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
			includes:         selectorMap{"env": "prod"},
			fileInfo:         fileInfoMock{isDir: false, name: "rule.md"},
			filePath:         "rule.md",
			fileContent:      "---\nenv: dev\n---\n# Dev Rule",
			expectInOutput:   false,
			expectExcludeLog: true,
			wantErr:          false,
		},
		{
			name:           "include rule with matching selector",
			includes:       selectorMap{"env": "prod"},
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

			var output, logOut bytes.Buffer
			cc := &codingContext{
				includes: tt.includes,
				output:   &output,
				logOut:   &logOut,
				cmdRunner: func(cmd *exec.Cmd) error {
					return nil // Mock command runner
				},
			}
			if cc.includes == nil {
				cc.includes = make(selectorMap)
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

			outputStr := output.String()
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
func (f *fileInfoMock) Sys() interface{}   { return nil }
