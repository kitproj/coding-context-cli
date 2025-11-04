package lib

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestVisit_SingleFile(t *testing.T) {
	// Create a temporary directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := `---
title: Test Title
author: Test Author
---
This is the content.
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Track visitor calls
	var callCount int
	var capturedFrontMatter FrontMatter
	var capturedContent string

	visitor := func(fm FrontMatter, c string) error {
		callCount++
		capturedFrontMatter = fm
		capturedContent = c
		return nil
	}

	// Visit the file
	err := Visit(testFile, visitor)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	// Verify visitor was called once
	if callCount != 1 {
		t.Errorf("visitor called %d times, want 1", callCount)
	}

	// Verify frontmatter
	if title, ok := capturedFrontMatter["title"].(string); !ok || title != "Test Title" {
		t.Errorf("frontmatter title = %v, want 'Test Title'", capturedFrontMatter["title"])
	}
	if author, ok := capturedFrontMatter["author"].(string); !ok || author != "Test Author" {
		t.Errorf("frontmatter author = %v, want 'Test Author'", capturedFrontMatter["author"])
	}

	// Verify content
	expectedContent := "This is the content.\n"
	if capturedContent != expectedContent {
		t.Errorf("content = %q, want %q", capturedContent, expectedContent)
	}
}

func TestVisit_MultipleFiles(t *testing.T) {
	// Create temporary directory with multiple files
	tmpDir := t.TempDir()
	
	files := []struct {
		name    string
		content string
	}{
		{
			name: "file1.md",
			content: `---
id: 1
---
Content 1
`,
		},
		{
			name: "file2.md",
			content: `---
id: 2
---
Content 2
`,
		},
		{
			name: "file3.md",
			content: `---
id: 3
---
Content 3
`,
		},
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", f.name, err)
		}
	}

	// Track all visited files
	visitedIDs := []int{}
	visitor := func(fm FrontMatter, c string) error {
		if id, ok := fm["id"].(int); ok {
			visitedIDs = append(visitedIDs, id)
		}
		return nil
	}

	// Visit all markdown files
	pattern := filepath.Join(tmpDir, "*.md")
	err := Visit(pattern, visitor)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	// Verify all files were visited
	if len(visitedIDs) != 3 {
		t.Errorf("visited %d files, want 3", len(visitedIDs))
	}
}

func TestVisit_NoFrontMatter(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "Just plain markdown content.\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	var capturedFrontMatter FrontMatter
	var capturedContent string

	visitor := func(fm FrontMatter, c string) error {
		capturedFrontMatter = fm
		capturedContent = c
		return nil
	}

	err := Visit(testFile, visitor)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	// Verify empty frontmatter
	if len(capturedFrontMatter) != 0 {
		t.Errorf("frontmatter length = %d, want 0", len(capturedFrontMatter))
	}

	// Verify content
	if capturedContent != content {
		t.Errorf("content = %q, want %q", capturedContent, content)
	}
}

func TestVisit_ErrorStopsProcessing(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create multiple files
	for i := 1; i <= 5; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".md")
		content := "Content\n"
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Track how many files were visited
	var visitCount int
	testErr := errors.New("test error")

	visitor := func(fm FrontMatter, c string) error {
		visitCount++
		if visitCount == 2 {
			return testErr
		}
		return nil
	}

	// Visit all markdown files
	pattern := filepath.Join(tmpDir, "*.md")
	err := Visit(pattern, visitor)
	
	// Verify it stopped on error
	if err == nil {
		t.Fatal("Visit() expected error, got nil")
	}
	if !errors.Is(err, testErr) {
		t.Errorf("Visit() error = %v, want %v", err, testErr)
	}

	// Verify it stopped after the error (visited 2 files)
	if visitCount != 2 {
		t.Errorf("visited %d files, want 2", visitCount)
	}
}

func TestVisit_NonExistentPattern(t *testing.T) {
	visitor := func(fm FrontMatter, c string) error {
		t.Error("visitor should not be called for non-existent pattern")
		return nil
	}

	// Visit with a pattern that matches no files
	err := Visit("/nonexistent/*.md", visitor)
	
	// Should succeed with no files to visit
	if err != nil {
		t.Errorf("Visit() error = %v, want nil", err)
	}
}

func TestVisit_InvalidPattern(t *testing.T) {
	visitor := func(fm FrontMatter, c string) error {
		t.Error("visitor should not be called for invalid pattern")
		return nil
	}

	// Visit with an invalid glob pattern
	err := Visit("[invalid", visitor)
	
	// Should return an error
	if err == nil {
		t.Error("Visit() expected error for invalid pattern, got nil")
	}
}

func TestVisit_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create a file
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("Content\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	var visitCount int
	visitor := func(fm FrontMatter, c string) error {
		visitCount++
		return nil
	}

	// Visit with a pattern that could match both file and directory
	pattern := filepath.Join(tmpDir, "*")
	err := Visit(pattern, visitor)
	if err != nil {
		t.Fatalf("Visit() error = %v", err)
	}

	// Should only visit the file, not the directory
	if visitCount != 1 {
		t.Errorf("visited %d items, want 1", visitCount)
	}
}

func TestVisit_ParseError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	
	// Create a file with invalid YAML frontmatter
	content := `---
invalid: [unclosed array
---
Content
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	visitor := func(fm FrontMatter, c string) error {
		t.Error("visitor should not be called when parsing fails")
		return nil
	}

	err := Visit(testFile, visitor)
	
	// Should return a parse error
	if err == nil {
		t.Error("Visit() expected parse error, got nil")
	}
}
