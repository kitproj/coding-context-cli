package main

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
			var frontmatter map[string]string
			content, err := parseMarkdownFile(tmpFile, &frontmatter)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMarkdownFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check content
			if content != tt.wantContent {
				t.Errorf("parseMarkdownFile() content = %q, want %q", content, tt.wantContent)
			}

			// Check frontmatter
			if len(frontmatter) != len(tt.wantFrontmatter) {
				t.Errorf("parseMarkdownFile() frontmatter length = %d, want %d", len(frontmatter), len(tt.wantFrontmatter))
			}
			for k, v := range tt.wantFrontmatter {
				if frontmatter[k] != v {
					t.Errorf("parseMarkdownFile() frontmatter[%q] = %q, want %q", k, frontmatter[k], v)
				}
			}
		})
	}
}

func TestParseMarkdownFile_FileNotFound(t *testing.T) {
	var frontmatter map[string]string
	_, err := parseMarkdownFile("/nonexistent/file.md", &frontmatter)
	if err == nil {
		t.Error("parseMarkdownFile() expected error for non-existent file, got nil")
	}
}
