package codingcontext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandFileReferences(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"test.txt":              "Hello, World!",
		"src/component.tsx":     "function Button() { return <button>Click me</button>; }",
		"docs/readme.md":        "# Documentation\n\nThis is a test.",
		"file with spaces.txt":  "Content with spaces in filename",
		"deeply/nested/file.go": "package main\n\nfunc main() {}",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	tests := []struct {
		name      string
		input     string
		baseDir   string
		wantErr   bool
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:    "no file references",
			input:   "This is plain text without any file references.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if result != "This is plain text without any file references." {
					t.Errorf("Expected unchanged text, got: %s", result)
				}
			},
		},
		{
			name:    "single file reference",
			input:   "Review the file @test.txt for issues.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "Hello, World!") {
					t.Errorf("Expected file content to be included, got: %s", result)
				}
				if !strings.Contains(result, "File: test.txt") {
					t.Errorf("Expected file header, got: %s", result)
				}
			},
		},
		{
			name:    "file reference with path",
			input:   "Check @src/component.tsx for performance issues.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "function Button()") {
					t.Errorf("Expected component content, got: %s", result)
				}
				if !strings.Contains(result, "File: src/component.tsx") {
					t.Errorf("Expected file header with path, got: %s", result)
				}
			},
		},
		{
			name:    "multiple file references",
			input:   "Compare @test.txt and @docs/readme.md files.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "Hello, World!") {
					t.Errorf("Expected first file content, got: %s", result)
				}
				if !strings.Contains(result, "# Documentation") {
					t.Errorf("Expected second file content, got: %s", result)
				}
			},
		},
		{
			name:    "email addresses should not be expanded",
			input:   "Contact me at user@example.com for questions.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if result != "Contact me at user@example.com for questions." {
					t.Errorf("Email should not be expanded, got: %s", result)
				}
			},
		},
		{
			name:    "file not found error",
			input:   "Check @nonexistent.txt file.",
			baseDir: tmpDir,
			wantErr: true,
			checkFunc: func(t *testing.T, result string) {
				// Error should be returned, result doesn't matter
			},
		},
		{
			name:    "relative path with ./",
			input:   "Review @./test.txt file.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "Hello, World!") {
					t.Errorf("Expected file content with ./ path, got: %s", result)
				}
			},
		},
		{
			name:    "deeply nested path",
			input:   "Check @deeply/nested/file.go implementation.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "package main") {
					t.Errorf("Expected nested file content, got: %s", result)
				}
			},
		},
		{
			name:    "file reference in markdown context",
			input:   "# Review Component\n\nReview the component in @src/component.tsx.\nCheck for performance issues.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "function Button()") {
					t.Errorf("Expected component content in markdown, got: %s", result)
				}
				if !strings.Contains(result, "# Review Component") {
					t.Errorf("Expected original content to be preserved, got: %s", result)
				}
			},
		},
		{
			name:    "file reference at start of line",
			input:   "@test.txt\nSome other text.",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "Hello, World!") {
					t.Errorf("Expected file content at start, got: %s", result)
				}
			},
		},
		{
			name:    "file reference at end of line",
			input:   "Check this file: @test.txt",
			baseDir: tmpDir,
			wantErr: false,
			checkFunc: func(t *testing.T, result string) {
				if !strings.Contains(result, "Hello, World!") {
					t.Errorf("Expected file content at end, got: %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandFileReferences(tt.input, tt.baseDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("expandFileReferences() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestReadFileReference(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string // Returns the file path to read
		filePath  string
		baseDir   string
		wantErr   bool
		checkFunc func(t *testing.T, content string)
	}{
		{
			name: "read simple file",
			setup: func() string {
				path := filepath.Join(tmpDir, "simple.txt")
				if err := os.WriteFile(path, []byte("simple content"), 0o644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return "simple.txt"
			},
			filePath: "simple.txt",
			baseDir:  tmpDir,
			wantErr:  false,
			checkFunc: func(t *testing.T, content string) {
				if content != "simple content" {
					t.Errorf("Expected 'simple content', got: %s", content)
				}
			},
		},
		{
			name:     "file not found",
			filePath: "nonexistent.txt",
			baseDir:  tmpDir,
			wantErr:  true,
		},
		{
			name: "read file with subdirectory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "subdir")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				path := filepath.Join(dir, "nested.txt")
				if err := os.WriteFile(path, []byte("nested content"), 0o644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return "subdir/nested.txt"
			},
			filePath: "subdir/nested.txt",
			baseDir:  tmpDir,
			wantErr:  false,
			checkFunc: func(t *testing.T, content string) {
				if content != "nested content" {
					t.Errorf("Expected 'nested content', got: %s", content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			content, err := readFileReference(tt.filePath, tt.baseDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("readFileReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, content)
			}
		})
	}
}

func TestFormatFileContent(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		content  string
		checks   []string // Strings that should be present in output
	}{
		{
			name:     "simple content",
			filePath: "test.txt",
			content:  "Hello, World!",
			checks:   []string{"File: test.txt", "Hello, World!", "```"},
		},
		{
			name:     "content with path",
			filePath: "src/components/Button.tsx",
			content:  "function Button() {}",
			checks:   []string{"File: src/components/Button.tsx", "function Button() {}", "```"},
		},
		{
			name:     "multiline content",
			filePath: "code.go",
			content:  "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}",
			checks:   []string{"File: code.go", "package main", "func main()", "```"},
		},
		{
			name:     "content without trailing newline",
			filePath: "test.txt",
			content:  "no newline",
			checks:   []string{"File: test.txt", "no newline", "```"},
		},
		{
			name:     "content with trailing newline",
			filePath: "test.txt",
			content:  "with newline\n",
			checks:   []string{"File: test.txt", "with newline", "```"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileContent(tt.filePath, tt.content)

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf("Expected result to contain %q, got: %s", check, result)
				}
			}

			// Verify the format structure
			if !strings.HasPrefix(result, "\n\n") {
				t.Errorf("Expected result to start with double newline")
			}
			if !strings.HasSuffix(result, "```\n\n") {
				t.Errorf("Expected result to end with closing backticks and double newline, got: %q", result)
			}
		})
	}
}

func TestFileReferencePattern(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMatch  bool
		wantGroups []string // Expected matched filepath (without @)
	}{
		{
			name:       "simple filename",
			input:      "@test.txt",
			wantMatch:  true,
			wantGroups: []string{"test.txt"},
		},
		{
			name:       "relative path with ./",
			input:      "@./src/file.go",
			wantMatch:  true,
			wantGroups: []string{"./src/file.go"},
		},
		{
			name:       "relative path with subdirectory",
			input:      "@src/components/Button.tsx",
			wantMatch:  true,
			wantGroups: []string{"src/components/Button.tsx"},
		},
		{
			name:       "parent directory",
			input:      "@../parent/file.txt",
			wantMatch:  true,
			wantGroups: []string{"../parent/file.txt"},
		},
		{
			name:      "email address",
			input:     "user@example.com",
			wantMatch: false,
		},
		{
			name:      "twitter handle",
			input:     "@username",
			wantMatch: false,
		},
		{
			name:       "multiple references",
			input:      "Compare @file1.txt and @file2.txt",
			wantMatch:  true,
			wantGroups: []string{"file1.txt", "file2.txt"},
		},
		{
			name:       "reference at start",
			input:      "@file.txt is important",
			wantMatch:  true,
			wantGroups: []string{"file.txt"},
		},
		{
			name:       "reference at end",
			input:      "Check @file.txt",
			wantMatch:  true,
			wantGroups: []string{"file.txt"},
		},
		{
			name:       "deeply nested path",
			input:      "@a/b/c/d/e/file.ext",
			wantMatch:  true,
			wantGroups: []string{"a/b/c/d/e/file.ext"},
		},
		{
			name:       "file with hyphens and underscores",
			input:      "@my-file_name.txt",
			wantMatch:  true,
			wantGroups: []string{"my-file_name.txt"},
		},
		{
			name:       "file in sentence",
			input:      "Review the component in @src/Button.tsx for issues.",
			wantMatch:  true,
			wantGroups: []string{"src/Button.tsx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := fileReferencePattern.FindAllStringSubmatch(tt.input, -1)

			if tt.wantMatch {
				if len(matches) == 0 {
					t.Errorf("Expected pattern to match, but got no matches")
					return
				}

				if len(tt.wantGroups) > 0 {
					if len(matches) != len(tt.wantGroups) {
						t.Errorf("Expected %d matches, got %d", len(tt.wantGroups), len(matches))
						return
					}

					for i, match := range matches {
						if len(match) < 2 {
							t.Errorf("Match %d has insufficient capture groups", i)
							continue
						}
						// match[1] is the first capture group (the filepath without @)
						if match[1] != tt.wantGroups[i] {
							t.Errorf("Match %d: expected %q, got %q", i, tt.wantGroups[i], match[1])
						}
					}
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("Expected no match, but got %d matches: %v", len(matches), matches)
				}
			}
		})
	}
}
