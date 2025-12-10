package codingcontext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnescapePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no escaping",
			input:    "path/to/file.txt",
			expected: "path/to/file.txt",
		},
		{
			name:     "escaped space",
			input:    "path/to/My\\ File.txt",
			expected: "path/to/My File.txt",
		},
		{
			name:     "multiple escaped spaces",
			input:    "src/My\\ Component\\ Name.tsx",
			expected: "src/My Component Name.tsx",
		},
		{
			name:     "escaped backslash",
			input:    "path\\\\to\\\\file",
			expected: "path\\to\\file",
		},
		{
			name:     "trailing backslash",
			input:    "path/to/file\\",
			expected: "path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unescapePath(tt.input)
			if result != tt.expected {
				t.Errorf("unescapePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandFileReferences(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	componentFile := filepath.Join(srcDir, "Button.tsx")
	if err := os.WriteFile(componentFile, []byte("function Button() {}"), 0o644); err != nil {
		t.Fatalf("Failed to create component file: %v", err)
	}

	spacedFile := filepath.Join(srcDir, "My Component.tsx")
	if err := os.WriteFile(spacedFile, []byte("spaced content"), 0o644); err != nil {
		t.Fatalf("Failed to create spaced file: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected []string // Strings that should be present in output
		notIn    []string // Strings that should NOT be present
	}{
		{
			name:     "no file references",
			input:    "This is plain text without any file references.",
			expected: []string{"This is plain text"},
			notIn:    []string{"File:"},
		},
		{
			name:     "single file reference",
			input:    "Review @test.txt for issues.",
			expected: []string{"test content", "File: test.txt", "Review", "for issues"},
			notIn:    []string{"@test.txt"},
		},
		{
			name:     "file reference with path",
			input:    "Check @src/Button.tsx please.",
			expected: []string{"function Button()", "File: src/Button.tsx", "Check", "please"},
		},
		{
			name:     "multiple file references",
			input:    "Compare @test.txt and @src/Button.tsx files.",
			expected: []string{"test content", "function Button()", "File: test.txt", "File: src/Button.tsx"},
		},
		{
			name:     "file reference with escaped spaces",
			input:    "Review @src/My\\ Component.tsx now.",
			expected: []string{"spaced content", "File: src/My Component.tsx", "Review", "now"},
		},
		{
			name:     "email addresses not expanded",
			input:    "Contact user@example.com for help.",
			expected: []string{"Contact user@example.com for help"},
			notIn:    []string{"File:"},
		},
		{
			name:     "file reference at start",
			input:    "@test.txt is important.",
			expected: []string{"test content", "File: test.txt", "is important"},
		},
		{
			name:     "file reference at end",
			input:    "Check this: @test.txt",
			expected: []string{"test content", "File: test.txt", "Check this:"},
		},
		{
			name:     "file reference followed by punctuation",
			input:    "Review @test.txt, @src/Button.tsx.",
			expected: []string{"test content", "function Button()", "File: test.txt", "File: src/Button.tsx"},
		},
		{
			name:     "nonexistent file keeps reference",
			input:    "Review @nonexistent.txt please.",
			expected: []string{"@nonexistent.txt", "Review", "please"},
			notIn:    []string{"File: nonexistent.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandFileReferences(tt.input, tmpDir)

			for _, check := range tt.expected {
				if !strings.Contains(result, check) {
					t.Errorf("Expected result to contain %q, got: %s", check, result)
				}
			}

			for _, check := range tt.notIn {
				if strings.Contains(result, check) {
					t.Errorf("Expected result NOT to contain %q, got: %s", check, result)
				}
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
		{
			name: "absolute path",
			setup: func() string {
				path := filepath.Join(tmpDir, "absolute.txt")
				if err := os.WriteFile(path, []byte("absolute content"), 0o644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			filePath: filepath.Join(tmpDir, "absolute.txt"),
			baseDir:  "/some/other/dir", // Should be ignored for absolute paths
			wantErr:  false,
			checkFunc: func(t *testing.T, content string) {
				if content != "absolute content" {
					t.Errorf("Expected 'absolute content', got: %s", content)
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
		{
			name:     "empty content",
			filePath: "empty.txt",
			content:  "",
			checks:   []string{"File: empty.txt", "```"},
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
