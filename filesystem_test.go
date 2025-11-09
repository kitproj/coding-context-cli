package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileSystem_Local(t *testing.T) {
	fsys, path, err := NewFileSystem("/tmp/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fsys.IsRemote() {
		t.Error("expected local file system")
	}

	if path != "/tmp/test.md" {
		t.Errorf("expected path /tmp/test.md, got %s", path)
	}
}

func TestNewFileSystem_HTTP(t *testing.T) {
	fsys, path, err := NewFileSystem("http://example.com/rules/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fsys.IsRemote() {
		t.Error("expected remote file system")
	}

	if path != "rules/test.md" {
		t.Errorf("expected path rules/test.md, got %s", path)
	}
}

func TestNewFileSystem_HTTPS(t *testing.T) {
	fsys, path, err := NewFileSystem("https://example.com/agents/rules/coding.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fsys.IsRemote() {
		t.Error("expected remote file system")
	}

	if path != "agents/rules/coding.md" {
		t.Errorf("expected path agents/rules/coding.md, got %s", path)
	}
}

func TestHTTPFileSystem_Open(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test.md" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("# Test Content\n\nThis is a test file."))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	hfs := NewHTTPFileSystem(server.URL)

	// Test opening an existing file
	f, err := hfs.Open("test.md")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "# Test Content\n\nThis is a test file."
	if string(content) != expected {
		t.Errorf("expected content %q, got %q", expected, string(content))
	}

	// Test opening a non-existent file
	_, err = hfs.Open("nonexistent.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "file does not exist") && !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("expected 'file does not exist' or 'no such file or directory' error, got: %v", err)
	}
}

func TestHTTPFileSystem_Stat(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test.md" {
			w.Header().Set("Content-Length", "42")
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	hfs := NewHTTPFileSystem(server.URL)

	// Test stat on existing file
	info, err := hfs.Stat("test.md")
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Name() != "test.md" {
		t.Errorf("expected name test.md, got %s", info.Name())
	}

	if info.Size() != 42 {
		t.Errorf("expected size 42, got %d", info.Size())
	}

	// Test stat on non-existent file
	_, err = hfs.Stat("nonexistent.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLocalFileSystem(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test File\n\nLocal content."

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	lfs := &LocalFileSystem{}

	// Test Open
	f, err := lfs.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected content %q, got %q", content, string(data))
	}

	// Test Stat
	info, err := lfs.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Name() != "test.md" {
		t.Errorf("expected name test.md, got %s", info.Name())
	}

	// Test Walk
	visitedPaths := []string{}
	err = lfs.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		visitedPaths = append(visitedPaths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk: %v", err)
	}

	if len(visitedPaths) < 2 { // Should visit tmpDir and testFile
		t.Errorf("expected at least 2 paths, got %d", len(visitedPaths))
	}
}

func TestOpenFile_Local(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test\n\nContent here."

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test opening local file
	f, fsys, err := openFile(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	if fsys.IsRemote() {
		t.Error("expected local file system")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected content %q, got %q", content, string(data))
	}
}

func TestOpenFile_HTTP(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Remote Content\n\nFrom HTTP."))
	}))
	defer server.Close()

	// Test opening remote file
	url := server.URL + "/test.md"
	f, fsys, err := openFile(url)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	if !fsys.IsRemote() {
		t.Error("expected remote file system")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "# Remote Content\n\nFrom HTTP."
	if string(data) != expected {
		t.Errorf("expected content %q, got %q", expected, string(data))
	}
}

func TestStatFile_Local(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	info, fsys, err := statFile(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if fsys.IsRemote() {
		t.Error("expected local file system")
	}

	if info.Name() != "test.md" {
		t.Errorf("expected name test.md, got %s", info.Name())
	}
}

func TestStatFile_HTTP(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	url := server.URL + "/test.md"
	info, fsys, err := statFile(url)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if !fsys.IsRemote() {
		t.Error("expected remote file system")
	}

	if info.Name() != "test.md" {
		t.Errorf("expected name test.md, got %s", info.Name())
	}

	if info.Size() != 100 {
		t.Errorf("expected size 100, got %d", info.Size())
	}
}

func TestStatFile_NotFound(t *testing.T) {
	// Create a test HTTP server that always returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	url := server.URL + "/nonexistent.md"
	_, _, err := statFile(url)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Check that it's a not found error
	if !strings.Contains(err.Error(), "file does not exist") && !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("expected 'file does not exist' or 'no such file or directory' error, got: %v", err)
	}
}

func TestWalkPath_Local(t *testing.T) {
	// Create a temporary directory with files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	visitedPaths := []string{}
	err := walkPath(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		visitedPaths = append(visitedPaths, path)
		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk: %v", err)
	}

	if len(visitedPaths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(visitedPaths))
	}
}

func TestWalkPath_HTTP(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "42")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	url := server.URL + "/test.md"
	visitedPaths := []string{}
	err := walkPath(url, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		visitedPaths = append(visitedPaths, path)
		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk: %v", err)
	}

	// For HTTP, we should visit just the single file
	if len(visitedPaths) != 1 {
		t.Errorf("expected 1 path for HTTP file, got %d", len(visitedPaths))
	}
}

func TestResolveRulePath(t *testing.T) {
	homeDir := "/home/user"

	tests := []struct {
		name     string
		rule     string
		expected string
	}{
		{
			name:     "URL http",
			rule:     "http://example.com/rules/test.md",
			expected: "http://example.com/rules/test.md",
		},
		{
			name:     "URL https",
			rule:     "https://example.com/rules/test.md",
			expected: "https://example.com/rules/test.md",
		},
		{
			name:     "Home directory path",
			rule:     "~/.agents/rules",
			expected: filepath.Join(homeDir, ".agents/rules"),
		},
		{
			name:     "Absolute path",
			rule:     "/etc/agents/rules",
			expected: "/etc/agents/rules",
		},
		{
			name:     "Relative path",
			rule:     ".agents/rules",
			expected: ".agents/rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveRulePath(tt.rule, homeDir)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseMarkdownFile_HTTP(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`---
language: Go
---
# Coding Standards

Use gofmt.`))
	}))
	defer server.Close()

	url := server.URL + "/test.md"

	var fm frontMatter
	content, err := parseMarkdownFile(url, &fm)
	if err != nil {
		t.Fatalf("failed to parse markdown file: %v", err)
	}

	if fm["language"] != "Go" {
		t.Errorf("expected language=Go, got %v", fm["language"])
	}

	expectedContent := "# Coding Standards\n\nUse gofmt.\n"
	if content != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, content)
	}
}
