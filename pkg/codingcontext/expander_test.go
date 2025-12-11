package codingcontext

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandParameters(t *testing.T) {
	tests := []struct {
		name     string
		params   Params
		content  string
		expected string
	}{
		{
			name:     "single parameter expansion",
			params:   Params{"name": "Alice"},
			content:  "Hello, ${name}!",
			expected: "Hello, Alice!",
		},
		{
			name:     "multiple parameter expansions",
			params:   Params{"first": "John", "last": "Doe"},
			content:  "${first} ${last}",
			expected: "John Doe",
		},
		{
			name:     "parameter not found - returns unchanged with warning",
			params:   Params{},
			content:  "Value: ${missing}",
			expected: "Value: ${missing}",
		},
		{
			name:     "mixed found and not found parameters",
			params:   Params{"found": "yes"},
			content:  "${found} and ${notfound}",
			expected: "yes and ${notfound}",
		},
		{
			name:     "no parameters to expand",
			params:   Params{"key": "value"},
			content:  "Plain text without parameters",
			expected: "Plain text without parameters",
		},
		{
			name:     "parameter with special characters",
			params:   Params{"path": "/tmp/file.txt"},
			content:  "File: ${path}",
			expected: "File: /tmp/file.txt",
		},
		{
			name:     "unclosed parameter - treated as literal",
			params:   Params{"name": "value"},
			content:  "Text ${name and more",
			expected: "Text ${name and more",
		},
		{
			name:     "empty parameter name - expands to empty",
			params:   Params{"": "value"},
			content:  "Text ${} more",
			expected: "Text value more",
		},
		{
			name:     "parameter at end of string",
			params:   Params{"end": "final"},
			content:  "Start ${end}",
			expected: "Start final",
		},
		{
			name:     "nested braces - outer takes precedence",
			params:   Params{"outer": "value"},
			content:  "${outer{inner}}",
			expected: "${outer{inner}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expand(tt.content, tt.params, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			if result != tt.expected {
				t.Errorf("expand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExpandCommands(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		contains string // Use contains for commands with variable output
	}{
		{
			name:     "simple echo command",
			content:  "Output: !`echo hello`",
			expected: "Output: hello\n",
		},
		{
			name:     "command with multiple words",
			content:  "!`echo hello world`",
			expected: "hello world\n",
		},
		{
			name:     "multiple commands in content",
			content:  "!`echo foo` and !`echo bar`",
			expected: "foo\n and bar\n",
		},
		{
			name:     "command that fails - returns output (empty for false)",
			content:  "!`false` failed",
			expected: " failed",
		},
		{
			name:     "command with pipes",
			content:  "!`echo test | tr a-z A-Z`",
			expected: "TEST\n",
		},
		{
			name:     "no commands to expand",
			content:  "Plain text without commands",
			expected: "Plain text without commands",
		},
		{
			name:     "command with newline in output",
			content:  "!`printf 'line1\\nline2'`",
			expected: "line1\nline2",
		},
		{
			name:     "command output not trimmed",
			content:  "!`echo -n hello` world",
			expected: "hello world",
		},
		{
			name:     "unclosed backtick - treated as literal",
			content:  "Text !`echo test more",
			expected: "Text !`echo test more",
		},
		{
			name:     "empty command",
			content:  "Text !`` more",
			expected: "Text  more",
		},
		{
			name:     "command at end of string",
			content:  "Start !`echo end`",
			expected: "Start end\n",
		},
		{
			name:     "command with error output",
			content:  "Error: !`cat /nonexistent/file 2>&1`",
			contains: "No such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expand(tt.content, Params{}, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			if tt.contains != "" {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("expand() = %q, should contain %q", result, tt.contains)
				}
			} else {
				if result != tt.expected {
					t.Errorf("expand() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestExpandPaths(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	testFile2 := filepath.Join(tmpDir, "test2.txt")
	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	testFileWithSpace := filepath.Join(tmpDir, "test file.txt")
	if err := os.WriteFile(testFileWithSpace, []byte("spaced content"), 0644); err != nil {
		t.Fatalf("failed to create test file with space: %v", err)
	}

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "single file expansion",
			content:  "File content: @" + testFile1,
			expected: "File content: content1",
		},
		{
			name:     "multiple file expansions",
			content:  "First: @" + testFile1 + " Second: @" + testFile2,
			expected: "First: content1 Second: content2",
		},
		{
			name:     "file not found - returns unchanged",
			content:  "Missing: @/nonexistent/file.txt",
			expected: "Missing: @/nonexistent/file.txt",
		},
		{
			name:     "file path with escaped space",
			content:  "Content: @" + strings.ReplaceAll(testFileWithSpace, " ", "\\ "),
			expected: "Content: spaced content",
		},
		{
			name:     "no paths to expand",
			content:  "Plain text without @ paths",
			expected: "Plain text without @ paths",
		},
		{
			name:     "@ not at start or after whitespace is not expanded",
			content:  "email@example.com",
			expected: "email@example.com",
		},
		{
			name:     "@ after newline",
			content:  "line1\n@" + testFile1,
			expected: "line1\ncontent1",
		},
		{
			name:     "path at end without trailing whitespace",
			content:  "End: @" + testFile1,
			expected: "End: content1",
		},
		{
			name:     "lone @ with no path",
			content:  "Text @ more",
			expected: "Text @ more",
		},
		{
			name:     "multiple consecutive @ symbols",
			content:  "Text @@ more",
			expected: "Text @@ more",
		},
		{
			name:     "path with backslash not escaping space - whole path not found",
			content:  "Path: @" + testFile1 + "\\notspace",
			expected: "Path: @" + testFile1 + "\\notspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expand(tt.content, Params{}, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			if result != tt.expected {
				t.Errorf("expand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExpandCombined(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(testFile, []byte("file-${param}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		params   Params
		content  string
		expected string
	}{
		{
			name:     "combined expansions - command, path, parameter",
			params:   Params{"name": "World"},
			content:  "!`echo Hello` ${name} from @" + testFile,
			expected: "Hello\n World from file-${param}",
		},
		{
			name:     "file content NOT re-expanded (security fix)",
			params:   Params{"param": "replaced"},
			content:  "@" + testFile,
			expected: "file-${param}", // Changed: file content is not re-expanded
		},
		{
			name:     "command output NOT re-expanded (security fix)",
			params:   Params{"dynamic": "value"},
			content:  "!`echo '${dynamic}'`",
			expected: "${dynamic}\n", // Changed: command output is not re-expanded
		},
		{
			name:     "all expansion types together",
			params:   Params{"x": "X", "y": "Y"},
			content:  "${x} !`echo middle` ${y}",
			expected: "X middle\n Y",
		},
		{
			name:     "no expansions needed",
			params:   Params{},
			content:  "plain text",
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expand(tt.content, tt.params, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			if result != tt.expected {
				t.Errorf("expand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExpandBasic(t *testing.T) {
	// Test basic expansion functionality
	content := "Hello ${name}!"
	params := Params{"name": "World"}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	result := expand(content, params, logger)
	expected := "Hello World!"
	if result != expected {
		t.Errorf("expand() = %q, want %q", result, expected)
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "simple filename",
			path:    "file.txt",
			wantErr: false,
		},
		{
			name:    "absolute path",
			path:    "/tmp/file.txt",
			wantErr: false,
		},
		{
			name:    "relative path with subdirectory",
			path:    "subdir/file.txt",
			wantErr: false,
		},
		{
			name:    "path with null byte - rejected",
			path:    "file\x00.txt",
			wantErr: true,
		},
		{
			name:    "path with directory traversal - allowed (legitimate use case)",
			path:    "../../../etc/passwd",
			wantErr: false,
		},
		{
			name:    "path with .. - allowed",
			path:    "dir/../file.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandSecurityNoReExpansion(t *testing.T) {
	tests := []struct {
		name     string
		params   Params
		content  string
		expected string
		desc     string
	}{
		{
			name:     "parameter value with command syntax not expanded",
			params:   Params{"evil": "!`echo INJECTED`"},
			content:  "Value: ${evil}",
			expected: "Value: !`echo INJECTED`",
			desc:     "Parameter containing command syntax should not be executed",
		},
		{
			name:     "parameter value with path syntax not expanded",
			params:   Params{"path": "@/etc/passwd"},
			content:  "Path: ${path}",
			expected: "Path: @/etc/passwd",
			desc:     "Parameter containing path syntax should not be read",
		},
		{
			name:     "command output with parameter syntax not expanded",
			params:   Params{"secret": "SECRET"},
			content:  "!`echo '${secret}'`",
			expected: "${secret}\n",
			desc:     "Command output containing parameter syntax should not be expanded",
		},
		{
			name:     "command output with path syntax not expanded",
			params:   Params{},
			content:  "!`echo '@/etc/passwd'`",
			expected: "@/etc/passwd\n",
			desc:     "Command output containing path syntax should not be read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expand(tt.content, tt.params, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			if result != tt.expected {
				t.Errorf("Security test failed: %s\nexpand() = %q, want %q", tt.desc, result, tt.expected)
			}
		})
	}
}
