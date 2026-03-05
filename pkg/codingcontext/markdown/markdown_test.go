package markdown

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
)

func TestParseError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  ParseError
		want string
	}{
		{
			name: "line and column set",
			err:  ParseError{File: "test.md", Line: 3, Column: 5, Message: "syntax error"},
			want: "test.md:3:5: syntax error",
		},
		{
			name: "line set without column",
			err:  ParseError{File: "test.md", Line: 3, Column: 0, Message: "syntax error"},
			want: "test.md:3: syntax error",
		},
		{
			name: "neither line nor column",
			err:  ParseError{File: "test.md", Line: 0, Column: 0, Message: "syntax error"},
			want: "test.md: syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("ParseError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFromContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty content",
			content: "",
		},
		{
			name:    "simple content",
			content: "Hello world",
		},
		{
			name:    "markdown content",
			content: "# Title\n\nThis is a paragraph with **bold** text.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fm := BaseFrontMatter{Name: "test-rule", Description: "a test rule"}
			md := FromContent(fm, tt.content)

			if md.Content != tt.content {
				t.Errorf("Content = %q, want %q", md.Content, tt.content)
			}

			if md.FrontMatter.Name != fm.Name {
				t.Errorf("FrontMatter.Name = %q, want %q", md.FrontMatter.Name, fm.Name)
			}

			if md.FrontMatter.Description != fm.Description {
				t.Errorf("FrontMatter.Description = %q, want %q", md.FrontMatter.Description, fm.Description)
			}

			if md.Structure == nil {
				t.Error("Structure should not be nil")
			}

			if len(tt.content) > 0 && md.Tokens == 0 {
				t.Error("Tokens should be non-zero for non-empty content")
			}

			if len(tt.content) == 0 && md.Tokens != 0 {
				t.Errorf("Tokens should be zero for empty content, got %d", md.Tokens)
			}
		})
	}
}

func TestParseMarkdownFile_YAMLErrorHasLineInfo(t *testing.T) {
	t.Parallel()

	// Unclosed bracket in YAML produces a parse error that goldmark-meta surfaces
	// with line information. Verify ParseError carries the line number.
	content := "---\nkey: [unclosed\n---\ncontent\n"

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")

	if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var fm BaseFrontMatter

	_, err := ParseMarkdownFile(tmpFile, &fm)
	if err == nil {
		t.Fatal("expected error for invalid YAML frontmatter, got nil")
	}

	parseErr := &ParseError{}

	ok := errors.As(err, &parseErr)
	if !ok {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}

	if parseErr.Line == 0 {
		t.Errorf("ParseError.Line should be > 0 when YAML error includes line info, got 0; error: %v", parseErr)
	}

	if parseErr.File != tmpFile {
		t.Errorf("ParseError.File = %q, want %q", parseErr.File, tmpFile)
	}
}

func TestParseMarkdownFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		content         string
		wantContent     string
		wantFrontmatter map[string]string
		wantErr         bool
	}{
		{
			name: "markdown with frontmatter",
			content: `---
title: Test Title
author: Test Author
---
This is the content
of the markdown file.
`,
			wantContent: "This is the content\nof the markdown file.\n",
			wantFrontmatter: map[string]string{
				"title":  "Test Title",
				"author": "Test Author",
			},
			wantErr: false,
		},
		{
			name: "markdown without frontmatter",
			content: `This is a simple markdown file
without any frontmatter.
`,
			wantContent:     "This is a simple markdown file\nwithout any frontmatter.\n",
			wantFrontmatter: map[string]string{},
			wantErr:         false,
		},
		{
			name: "markdown with title as first line",
			content: `# My Title

This is the content.
`,
			wantContent:     "# My Title\n\nThis is the content.\n",
			wantFrontmatter: map[string]string{},
			wantErr:         false,
		},
		{
			name:            "empty file",
			content:         "",
			wantContent:     "",
			wantFrontmatter: map[string]string{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary file
			tmpDir := t.TempDir()

			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Parse the file
			var frontmatter BaseFrontMatter

			md, err := ParseMarkdownFile(tmpFile, &frontmatter)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMarkdownFile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			// Check content
			if md.Content != tt.wantContent {
				t.Errorf("ParseMarkdownFile() content = %q, want %q", md.Content, tt.wantContent)
			}

			// Check frontmatter
			if len(frontmatter.Content) != len(tt.wantFrontmatter) {
				t.Errorf("ParseMarkdownFile() frontmatter length = %d, want %d", len(frontmatter.Content), len(tt.wantFrontmatter))
			}

			for k, v := range tt.wantFrontmatter {
				if fmVal, ok := frontmatter.Content[k].(string); !ok || fmVal != v {
					t.Errorf("ParseMarkdownFile() frontmatter[%q] = %v, want %q", k, frontmatter.Content[k], v)
				}
			}
		})
	}
}

func TestParseMarkdownFile_FileNotFound(t *testing.T) {
	t.Parallel()

	var frontmatter BaseFrontMatter

	_, err := ParseMarkdownFile("/nonexistent/file.md", &frontmatter)
	if err == nil {
		t.Error("ParseMarkdownFile() expected error for non-existent file, got nil")
	}
	// Verify error message includes file path
	if err != nil && !strings.Contains(err.Error(), "/nonexistent/file.md") {
		t.Errorf("ParseMarkdownFile() error should contain file path, got: %v", err)
	}
}

func TestParseMarkdownFile_ErrorsIncludeFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string // substring that should appear in error
	}{
		{
			name: "invalid YAML in frontmatter",
			content: `---
invalid: yaml: : syntax
---
Content here`,
			want: "failed to parse YAML frontmatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary file
			tmpDir := t.TempDir()

			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Parse the file
			var frontmatter BaseFrontMatter

			_, err := ParseMarkdownFile(tmpFile, &frontmatter)

			// Check that we got an error
			if err == nil {
				t.Fatalf("ParseMarkdownFile() expected error for %s, got nil", tt.name)
			}

			// Check that error contains expected substring
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseMarkdownFile() error should contain %q, got: %v", tt.want, err)
			}

			// Check that error contains file path
			if !strings.Contains(err.Error(), tmpFile) {
				t.Errorf("ParseMarkdownFile() error should contain file path %q, got: %v", tmpFile, err)
			}
		})
	}
}

type testTaskFrontmatter struct {
	TaskName string   `yaml:"task_name"`
	Resume   bool     `yaml:"resume"`
	Priority string   `yaml:"priority"`
	Tags     []string `yaml:"tags"`
}

func assertCustomFrontmatter(t *testing.T, fm testTaskFrontmatter, md Markdown[testTaskFrontmatter], err error,
	wantErr bool, wantContent, wantTaskName, wantPriority string, wantResume bool, wantTags []string,
) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Errorf("ParseMarkdownFile() error = %v, wantErr %v", err, wantErr)

		return
	}

	if md.Content != wantContent {
		t.Errorf("ParseMarkdownFile() content = %q, want %q", md.Content, wantContent)
	}

	if fm.TaskName != wantTaskName {
		t.Errorf("frontmatter.TaskName = %q, want %q", fm.TaskName, wantTaskName)
	}

	if fm.Resume != wantResume {
		t.Errorf("frontmatter.Resume = %v, want %v", fm.Resume, wantResume)
	}

	if fm.Priority != wantPriority {
		t.Errorf("frontmatter.Priority = %q, want %q", fm.Priority, wantPriority)
	}

	if len(fm.Tags) != len(wantTags) {
		t.Errorf("frontmatter.Tags length = %d, want %d", len(fm.Tags), len(wantTags))
	}

	for i, tag := range wantTags {
		if i < len(fm.Tags) && fm.Tags[i] != tag {
			t.Errorf("frontmatter.Tags[%d] = %q, want %q", i, fm.Tags[i], tag)
		}
	}
}

func TestParseMarkdownFile_CustomStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		wantContent  string
		wantTaskName string
		wantResume   bool
		wantPriority string
		wantTags     []string
		wantErr      bool
	}{
		{
			name: "parse task with all fields",
			content: `---
task_name: fix-bug
resume: false
priority: high
tags:
  - backend
  - urgent
---
# Fix Bug

Please fix the bug in the backend service.
`,
			wantContent:  "# Fix Bug\n\nPlease fix the bug in the backend service.\n",
			wantTaskName: "fix-bug",
			wantResume:   false,
			wantPriority: "high",
			wantTags:     []string{"backend", "urgent"},
			wantErr:      false,
		},
		{
			name: "parse task with partial fields",
			content: `---
task_name: deploy
resume: true
---
# Deploy Application

Deploy the application to staging.
`,
			wantContent:  "# Deploy Application\n\nDeploy the application to staging.\n",
			wantTaskName: "deploy",
			wantResume:   true,
			wantPriority: "", // zero value for missing field
			wantTags:     nil,
			wantErr:      false,
		},
		{
			name: "parse without frontmatter",
			content: `# Simple Task

This task has no frontmatter.
`,
			wantContent:  "# Simple Task\n\nThis task has no frontmatter.\n",
			wantTaskName: "", // zero value
			wantResume:   false,
			wantPriority: "",
			wantTags:     nil,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")

			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			var frontmatter testTaskFrontmatter

			md, err := ParseMarkdownFile(tmpFile, &frontmatter)

			assertCustomFrontmatter(t, frontmatter, md, err, tt.wantErr,
				tt.wantContent, tt.wantTaskName, tt.wantPriority, tt.wantResume, tt.wantTags)
		})
	}
}

func TestParseMarkdownFile_MultipleNewlinesAfterFrontmatter(t *testing.T) {
	t.Parallel()
	// This test verifies that multiple newlines after the frontmatter
	// closing delimiter are handled correctly.
	// The parser should:
	// 1. Preserve multiple newlines between frontmatter and content
	// 2. Strip a single newline (treating it as just a separator)
	// 3. Allow the task parser to successfully parse content that starts with newlines
	tests := []struct {
		name        string
		content     string
		wantContent string
	}{
		{
			name: "multiple newlines after frontmatter",
			content: `---
{}
---

Start of context
`,
			wantContent: "\nStart of context\n", // Content copied as-is after frontmatter
		},
		{
			name: "single newline after frontmatter (baseline)",
			content: `---
{}
---
Start of context
`,
			wantContent: "Start of context\n", // Content copied as-is after frontmatter (newline after --- is preserved)
		},
		{
			name: "three newlines after frontmatter",
			content: `---
{}
---


Start of context
`,
			wantContent: "\n\nStart of context\n", // Content copied as-is after frontmatter
		},
		{
			name: "mixed whitespace after frontmatter",
			content: `---
{}
---
  
	 

Start of context
`,
			// Content copied as-is, preserving whitespace (newline after --- is preserved)
			wantContent: "  \n\t \n\nStart of context\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary file
			tmpDir := t.TempDir()

			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Parse the file
			var frontmatter BaseFrontMatter

			md, err := ParseMarkdownFile(tmpFile, &frontmatter)
			if err != nil {
				t.Fatalf("ParseMarkdownFile() error = %v", err)
			}

			// Check content
			if md.Content != tt.wantContent {
				t.Errorf("ParseMarkdownFile() content = %q, want %q", md.Content, tt.wantContent)
			}

			// Verify that the content can be parsed as a task
			// This is the actual use case - content is parsed as a task after frontmatter extraction
			task, err := taskparser.ParseTask(md.Content)
			if err != nil {
				t.Fatalf("ParseTask() failed: %v, content = %q", err, md.Content)
			}

			if len(task) == 0 && strings.TrimSpace(md.Content) != "" {
				t.Errorf("ParseTask() returned empty task for non-empty content: %q", md.Content)
			}
			// Verify that the parsed task content matches the original exactly
			// The parser preserves all content including leading newlines
			taskContent := task.String()
			if taskContent != md.Content {
				t.Errorf("ParseTask() then String() = %q, want %q", taskContent, md.Content)
			}
		})
	}
}
