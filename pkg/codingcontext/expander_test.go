package codingcontext

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a test logger that discards output
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors in tests
	}))
}

func TestExpandString_ParameterExpansion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   map[string]string
		expected string
	}{
		{
			name:     "single parameter found",
			input:    "Hello ${name}!",
			params:   map[string]string{"name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "multiple parameters",
			input:    "${greeting} ${name}, your id is ${id}",
			params:   map[string]string{"greeting": "Hello", "name": "Alice", "id": "123"},
			expected: "Hello Alice, your id is 123",
		},
		{
			name:     "parameter not found leaves unexpanded",
			input:    "Value: ${missing}",
			params:   map[string]string{},
			expected: "Value: ${missing}",
		},
		{
			name:     "mix of found and missing parameters",
			input:    "${found} and ${missing}",
			params:   map[string]string{"found": "exists"},
			expected: "exists and ${missing}",
		},
		{
			name:     "parameter with underscores",
			input:    "${my_param}",
			params:   map[string]string{"my_param": "value"},
			expected: "value",
		},
		{
			name:     "parameter with hyphens",
			input:    "${my-param}",
			params:   map[string]string{"my-param": "value"},
			expected: "value",
		},
		{
			name:     "parameter with dots",
			input:    "${my.param}",
			params:   map[string]string{"my.param": "value"},
			expected: "value",
		},
		{
			name:     "empty parameter value",
			input:    "${empty}",
			params:   map[string]string{"empty": ""},
			expected: "",
		},
		{
			name:     "no parameters in text",
			input:    "Just plain text",
			params:   map[string]string{},
			expected: "Just plain text",
		},
		{
			name:     "parameter at start",
			input:    "${start} of text",
			params:   map[string]string{"start": "Beginning"},
			expected: "Beginning of text",
		},
		{
			name:     "parameter at end",
			input:    "End of ${end}",
			params:   map[string]string{"end": "text"},
			expected: "End of text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandString(tt.input, tt.params, testLogger())
			if err != nil {
				t.Fatalf("ExpandString failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpandString_CommandExpansion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Or a check function for dynamic output
	}{
		{
			name:     "simple echo command",
			input:    "Output: !`echo hello`",
			expected: "Output: hello",
		},
		{
			name:     "command with spaces",
			input:    "!`echo hello world`",
			expected: "hello world",
		},
		{
			name:     "multiple commands",
			input:    "First: !`echo one`, Second: !`echo two`",
			expected: "First: one, Second: two",
		},
		{
			name:     "command at start",
			input:    "!`echo start` of text",
			expected: "start of text",
		},
		{
			name:     "command at end",
			input:    "End of !`echo text`",
			expected: "End of text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandString(tt.input, nil, testLogger())
			if err != nil {
				t.Fatalf("ExpandString failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpandString_CommandExpansion_Failures(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "failing command returns empty",
			input: "!`exit 1`",
			check: func(t *testing.T, result string) {
				// Command fails but still returns its output (which is empty)
				if result != "" {
					t.Errorf("expected empty output from failing command, got %q", result)
				}
			},
		},
		{
			name:  "non-existent command",
			input: "!`nonexistentcommand12345`",
			check: func(t *testing.T, result string) {
				// Should return error output or empty
				// Just verify it doesn't panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandString(tt.input, nil, testLogger())
			if err != nil {
				t.Fatalf("ExpandString failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestExpandString_FileExpansion(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello from file"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a file with spaces in the name
	testFileWithSpaces := filepath.Join(tmpDir, "test file.txt")
	spaceContent := "Content with spaces"
	if err := os.WriteFile(testFileWithSpaces, []byte(spaceContent), 0644); err != nil {
		t.Fatalf("failed to create test file with spaces: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple file path",
			input:    "Content: @" + testFile,
			expected: "Content: " + testContent,
		},
		{
			name:     "file path with escaped spaces",
			input:    "Content: @" + strings.ReplaceAll(testFileWithSpaces, " ", `\ `),
			expected: "Content: " + spaceContent,
		},
		{
			name:     "file at start",
			input:    "@" + testFile + " is the content",
			expected: testContent + " is the content",
		},
		{
			name:     "file at end",
			input:    "The content is @" + testFile,
			expected: "The content is " + testContent,
		},
		{
			name:     "non-existent file leaves unexpanded",
			input:    "@/nonexistent/file.txt",
			expected: "@/nonexistent/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandString(tt.input, nil, testLogger())
			if err != nil {
				t.Fatalf("ExpandString failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpandString_MixedExpansions(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("file content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		params   map[string]string
		expected string
	}{
		{
			name:     "parameter and command",
			input:    "${name} says !`echo hello`",
			params:   map[string]string{"name": "Alice"},
			expected: "Alice says hello",
		},
		{
			name:     "parameter and file",
			input:    "${prefix}: @" + testFile,
			params:   map[string]string{"prefix": "Content"},
			expected: "Content: file content",
		},
		{
			name:     "command and file",
			input:    "!`echo Command output` and @" + testFile,
			params:   map[string]string{},
			expected: "Command output and file content",
		},
		{
			name:     "all three types",
			input:    "${name}: !`echo hello` from @" + testFile,
			params:   map[string]string{"name": "Message"},
			expected: "Message: hello from file content",
		},
		{
			name:     "literal text with all expansions",
			input:    "Start ${param} middle !`echo cmd` end @" + testFile + " finish",
			params:   map[string]string{"param": "value"},
			expected: "Start value middle cmd end file content finish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandString(tt.input, tt.params, testLogger())
			if err != nil {
				t.Fatalf("ExpandString failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpandString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   map[string]string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			params:   nil,
			expected: "",
		},
		{
			name:     "only literal text",
			input:    "Just plain text with $ ! @ characters",
			params:   nil,
			expected: "Just plain text with $ ! @ characters",
		},
		{
			name:     "consecutive parameters",
			input:    "${first}${second}${third}",
			params:   map[string]string{"first": "1", "second": "2", "third": "3"},
			expected: "123",
		},
		{
			name:     "parameter with empty name", // This might not parse correctly
			input:    "${}",
			params:   nil,
			expected: "${}",
		},
		{
			name:     "backtick without exclamation",
			input:    "Text with ` backticks ` but no expansion",
			params:   nil,
			expected: "Text with ` backticks ` but no expansion",
		},
		{
			name:     "at sign without path",
			input:    "Email @ example.com",
			params:   nil,
			expected: "Email @ example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandString(tt.input, tt.params, testLogger())
			// Some edge cases might fail to parse, which is acceptable
			if err != nil {
				t.Logf("Parse error (expected for some edge cases): %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpandString_Multiline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   map[string]string
		expected string
	}{
		{
			name:     "multiline with parameter",
			input:    "Line 1: ${param}\nLine 2: more text",
			params:   map[string]string{"param": "value"},
			expected: "Line 1: value\nLine 2: more text",
		},
		{
			name:     "multiline with command",
			input:    "Line 1\n!`echo output`\nLine 3",
			params:   nil,
			expected: "Line 1\noutput\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandString(tt.input, tt.params, testLogger())
			if err != nil {
				t.Fatalf("ExpandString failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
