package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandFileReferences(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		setupFiles      func(t *testing.T, tmpDir string)
		wantContains    []string
		wantNotContains []string
		wantErr         bool
		errContains     string
	}{
		{
			name:    "no file references",
			content: "This is just plain text without any references.",
			setupFiles: func(t *testing.T, tmpDir string) {
				// No files needed
			},
			wantContains: []string{"This is just plain text without any references."},
			wantErr:      false,
		},
		{
			name:    "single file reference",
			content: "Review the code in @test.txt and provide feedback.",
			setupFiles: func(t *testing.T, tmpDir string) {
				err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("Hello World"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			},
			wantContains: []string{
				"Review the code in",
				"```txt\n# File: test.txt\nHello World```",
				"and provide feedback.",
			},
			wantNotContains: []string{"@test.txt"},
			wantErr:         false,
		},
		{
			name:    "multiple file references",
			content: "Compare @file1.go and @file2.go for differences.",
			setupFiles: func(t *testing.T, tmpDir string) {
				err := os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte("package main\n\nfunc main() {}"), 0644)
				if err != nil {
					t.Fatalf("failed to create file1.go: %v", err)
				}
				err = os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("package test\n\nfunc test() {}"), 0644)
				if err != nil {
					t.Fatalf("failed to create file2.go: %v", err)
				}
			},
			wantContains: []string{
				"Compare",
				"```go\n# File: file1.go\npackage main\n\nfunc main() {}```",
				"and",
				"```go\n# File: file2.go\npackage test\n\nfunc test() {}```",
				"for differences.",
			},
			wantNotContains: []string{"@file1.go", "@file2.go"},
			wantErr:         false,
		},
		{
			name:    "file reference with subdirectory",
			content: "Check @src/components/Button.tsx for issues.",
			setupFiles: func(t *testing.T, tmpDir string) {
				dir := filepath.Join(tmpDir, "src", "components")
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				err = os.WriteFile(filepath.Join(dir, "Button.tsx"), []byte("export const Button = () => <button>Click</button>;"), 0644)
				if err != nil {
					t.Fatalf("failed to create Button.tsx: %v", err)
				}
			},
			wantContains: []string{
				"Check",
				"```tsx\n# File: src/components/Button.tsx\nexport const Button = () => <button>Click</button>;```",
				"for issues.",
			},
			wantNotContains: []string{"@src/components/Button.tsx"},
			wantErr:         false,
		},
		{
			name:    "file reference not found",
			content: "Review @nonexistent.txt please.",
			setupFiles: func(t *testing.T, tmpDir string) {
				// Don't create the file
			},
			wantErr:     true,
			errContains: "failed to read referenced file nonexistent.txt",
		},
		{
			name:    "file reference at start of line",
			content: "@config.yaml\ncontains the configuration.",
			setupFiles: func(t *testing.T, tmpDir string) {
				err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("key: value"), 0644)
				if err != nil {
					t.Fatalf("failed to create config.yaml: %v", err)
				}
			},
			wantContains: []string{
				"```yaml\n# File: config.yaml\nkey: value```",
				"contains the configuration.",
			},
			wantNotContains: []string{"@config.yaml"},
			wantErr:         false,
		},
		{
			name:    "file reference at end of line",
			content: "See the code in @main.go",
			setupFiles: func(t *testing.T, tmpDir string) {
				err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
				if err != nil {
					t.Fatalf("failed to create main.go: %v", err)
				}
			},
			wantContains: []string{
				"See the code in",
				"```go\n# File: main.go\npackage main```",
			},
			wantNotContains: []string{"@main.go"},
			wantErr:         false,
		},
		{
			name:    "file reference with dash and underscore in name",
			content: "Review @my-test_file.js for bugs.",
			setupFiles: func(t *testing.T, tmpDir string) {
				err := os.WriteFile(filepath.Join(tmpDir, "my-test_file.js"), []byte("console.log('test');"), 0644)
				if err != nil {
					t.Fatalf("failed to create my-test_file.js: %v", err)
				}
			},
			wantContains: []string{
				"Review",
				"```js\n# File: my-test_file.js\nconsole.log('test');```",
				"for bugs.",
			},
			wantNotContains: []string{"@my-test_file.js"},
			wantErr:         false,
		},
		{
			name:    "file reference followed by punctuation",
			content: "Check @README.md. Then review @LICENSE.",
			setupFiles: func(t *testing.T, tmpDir string) {
				err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# README"), 0644)
				if err != nil {
					t.Fatalf("failed to create README.md: %v", err)
				}
				err = os.WriteFile(filepath.Join(tmpDir, "LICENSE"), []byte("MIT License"), 0644)
				if err != nil {
					t.Fatalf("failed to create LICENSE: %v", err)
				}
			},
			wantContains: []string{
				"Check",
				"```md\n# File: README.md\n# README```",
				". Then review",
				"```text\n# File: LICENSE\nMIT License```",
			},
			wantNotContains: []string{"@README.md", "@LICENSE"},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for test files
			tmpDir := t.TempDir()

			// Setup test files
			if tt.setupFiles != nil {
				tt.setupFiles(t, tmpDir)
			}

			// Expand file references
			result, err := expandFileReferences(tt.content, tmpDir)

			// Check error expectation
			if tt.wantErr {
				if err == nil {
					t.Errorf("expandFileReferences() expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expandFileReferences() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("expandFileReferences() unexpected error = %v", err)
				return
			}

			// Check that expected strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("expandFileReferences() result missing expected string:\nwant: %q\ngot: %q", want, result)
				}
			}

			// Check that unexpected strings are not present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("expandFileReferences() result contains unexpected string:\ndon't want: %q\ngot: %q", notWant, result)
				}
			}
		})
	}
}

func TestReadReferencedFile(t *testing.T) {
	tests := []struct {
		name        string
		filepath    string
		setupFile   func(t *testing.T, tmpDir string) string
		wantContent string
		wantErr     bool
		errContains string
	}{
		{
			name:     "read existing file",
			filepath: "test.txt",
			setupFile: func(t *testing.T, tmpDir string) string {
				path := filepath.Join(tmpDir, "test.txt")
				err := os.WriteFile(path, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			wantContent: "test content",
			wantErr:     false,
		},
		{
			name:     "read file from subdirectory",
			filepath: "subdir/file.txt",
			setupFile: func(t *testing.T, tmpDir string) string {
				dir := filepath.Join(tmpDir, "subdir")
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				path := filepath.Join(dir, "file.txt")
				err = os.WriteFile(path, []byte("nested content"), 0644)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return tmpDir
			},
			wantContent: "nested content",
			wantErr:     false,
		},
		{
			name:     "file not found",
			filepath: "nonexistent.txt",
			setupFile: func(t *testing.T, tmpDir string) string {
				return tmpDir
			},
			wantErr:     true,
			errContains: "failed to read file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workDir := tt.setupFile(t, tmpDir)

			content, err := readReferencedFile(tt.filepath, workDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("readReferencedFile() expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("readReferencedFile() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("readReferencedFile() unexpected error = %v", err)
				return
			}

			if content != tt.wantContent {
				t.Errorf("readReferencedFile() = %q, want %q", content, tt.wantContent)
			}
		})
	}
}

func TestFormatFileContent(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		content  string
		want     string
	}{
		{
			name:     "go file",
			filepath: "main.go",
			content:  "package main",
			want:     "```go\n# File: main.go\npackage main```",
		},
		{
			name:     "typescript file",
			filepath: "component.tsx",
			content:  "export const Component = () => null;",
			want:     "```tsx\n# File: component.tsx\nexport const Component = () => null;```",
		},
		{
			name:     "file without extension",
			filepath: "LICENSE",
			content:  "MIT License",
			want:     "```text\n# File: LICENSE\nMIT License```",
		},
		{
			name:     "python file",
			filepath: "script.py",
			content:  "print('hello')",
			want:     "```py\n# File: script.py\nprint('hello')```",
		},
		{
			name:     "file with path",
			filepath: "src/utils/helper.js",
			content:  "export function helper() {}",
			want:     "```js\n# File: src/utils/helper.js\nexport function helper() {}```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFileContent(tt.filepath, tt.content)
			if got != tt.want {
				t.Errorf("formatFileContent() = %q, want %q", got, tt.want)
			}
		})
	}
}
