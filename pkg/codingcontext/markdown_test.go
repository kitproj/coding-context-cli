package codingcontext

import (
	"os"
	"path/filepath"
	"testing"
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
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
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
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
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
