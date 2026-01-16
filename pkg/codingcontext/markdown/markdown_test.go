package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
)

func TestParseMarkdownFile(t *testing.T) {
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
			// Create a temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o644); err != nil {
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
			want: "failed to unmarshal frontmatter in file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o644); err != nil {
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

func TestParseMarkdownFile_CustomStruct(t *testing.T) {
	// Define a custom struct for task frontmatter
	type TaskFrontmatter struct {
		TaskName string   `yaml:"task_name"`
		Resume   bool     `yaml:"resume"`
		Priority string   `yaml:"priority"`
		Tags     []string `yaml:"tags"`
	}

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
			// Create a temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Parse the file into custom struct
			var frontmatter TaskFrontmatter
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

			// Check frontmatter fields
			if frontmatter.TaskName != tt.wantTaskName {
				t.Errorf("frontmatter.TaskName = %q, want %q", frontmatter.TaskName, tt.wantTaskName)
			}
			if frontmatter.Resume != tt.wantResume {
				t.Errorf("frontmatter.Resume = %v, want %v", frontmatter.Resume, tt.wantResume)
			}
			if frontmatter.Priority != tt.wantPriority {
				t.Errorf("frontmatter.Priority = %q, want %q", frontmatter.Priority, tt.wantPriority)
			}
			if len(frontmatter.Tags) != len(tt.wantTags) {
				t.Errorf("frontmatter.Tags length = %d, want %d", len(frontmatter.Tags), len(tt.wantTags))
			}
			for i, tag := range tt.wantTags {
				if i < len(frontmatter.Tags) && frontmatter.Tags[i] != tag {
					t.Errorf("frontmatter.Tags[%d] = %q, want %q", i, frontmatter.Tags[i], tag)
				}
			}
		})
	}
}

func TestParseMarkdownFile_MultipleNewlinesAfterFrontmatter(t *testing.T) {
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
			wantContent: "  \n\t \n\nStart of context\n", // Content copied as-is, preserving whitespace (newline after --- is preserved)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o644); err != nil {
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
