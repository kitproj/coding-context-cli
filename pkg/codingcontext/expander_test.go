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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expander := NewExpander(tt.params, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			result := expander.expandParameters(tt.content)
			if result != tt.expected {
				t.Errorf("expandParameters() = %q, want %q", result, tt.expected)
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
			expected: "Output: hello",
		},
		{
			name:     "command with multiple words",
			content:  "!`echo hello world`",
			expected: "hello world",
		},
		{
			name:     "multiple commands in content",
			content:  "!`echo foo` and !`echo bar`",
			expected: "foo and bar",
		},
		{
			name:     "command that fails - returns unchanged",
			content:  "!`false` failed",
			expected: "!`false` failed",
		},
		{
			name:     "command with pipes",
			content:  "!`echo test | tr a-z A-Z`",
			expected: "TEST",
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
			name:     "command output trimmed trailing newline",
			content:  "!`echo -n hello` world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expander := NewExpander(Params{}, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			result := expander.expandCommands(tt.content)
			if tt.contains != "" {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("expandCommands() = %q, should contain %q", result, tt.contains)
				}
			} else {
				if result != tt.expected {
					t.Errorf("expandCommands() = %q, want %q", result, tt.expected)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expander := NewExpander(Params{}, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			result := expander.expandPaths(tt.content)
			if result != tt.expected {
				t.Errorf("expandPaths() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExpand(t *testing.T) {
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
			expected: "Hello World from file-${param}",
		},
		{
			name:     "parameter expansion in file content",
			params:   Params{"param": "replaced"},
			content:  "@" + testFile,
			expected: "file-replaced",
		},
		{
			name:     "command generates parameter reference",
			params:   Params{"dynamic": "value"},
			content:  "!`echo '${dynamic}'`",
			expected: "value",
		},
		{
			name:     "all expansion types together",
			params:   Params{"x": "X", "y": "Y"},
			content:  "${x} !`echo middle` ${y}",
			expected: "X middle Y",
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
			expander := NewExpander(tt.params, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			result := expander.Expand(tt.content)
			if result != tt.expected {
				t.Errorf("Expand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewExpander(t *testing.T) {
	tests := []struct {
		name   string
		params Params
		logger *slog.Logger
	}{
		{
			name:   "with params and logger",
			params: Params{"key": "value"},
			logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		},
		{
			name:   "with nil logger - should create default",
			params: Params{},
			logger: nil,
		},
		{
			name:   "with empty params",
			params: Params{},
			logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expander := NewExpander(tt.params, tt.logger)
			if expander == nil {
				t.Error("NewExpander() returned nil")
			}
			if expander.params == nil {
				t.Error("expander.params is nil")
			}
			if expander.logger == nil {
				t.Error("expander.logger is nil")
			}
		})
	}
}
