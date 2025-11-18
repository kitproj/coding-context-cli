package codingcontext_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

// Example demonstrates how to use the codingcontext library programmatically
func Example() {
	// Create a temporary directory for the example
	tmpDir, err := os.MkdirTemp("", "example-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create a sample task file
	taskDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		fmt.Printf("Failed to create task dir: %v\n", err)
		return
	}

	taskContent := `---
task_name: example-task
---
# Example Task

This is an example task with parameter: ${param1}
`
	if err := os.WriteFile(filepath.Join(taskDir, "example.md"), []byte(taskContent), 0o644); err != nil {
		fmt.Printf("Failed to write task file: %v\n", err)
		return
	}

	// Create a sample rule file
	ruleContent := `# Example Rule

This is an example rule.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "CLAUDE.md"), []byte(ruleContent), 0o644); err != nil {
		fmt.Printf("Failed to write rule file: %v\n", err)
		return
	}

	// Set up parameters
	params := make(codingcontext.ParamMap)
	params.Set("param1=value1")

	// Set up selectors
	includes := make(codingcontext.SelectorMap)

	// Create output buffers
	var output, logOut bytes.Buffer

	// Create a new context
	ctx := codingcontext.New(
		codingcontext.WithWorkDir(tmpDir),
		codingcontext.WithParams(params),
		codingcontext.WithIncludes(includes),
		codingcontext.WithOutput(&output),
		codingcontext.WithLogOutput(&logOut),
	)

	// Run the context assembly
	if err := ctx.Run(context.Background(), "example-task"); err != nil {
		fmt.Printf("Failed to run context: %v\n", err)
		return
	}

	// Access the total tokens
	_ = ctx.TotalTokens()

	fmt.Println("Context assembled successfully")
	// Output: Context assembled successfully
}

// ExampleWithCmdRunner demonstrates using a custom command runner
func ExampleWithCmdRunner() {
	tmpDir, err := os.MkdirTemp("", "example-cmd-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create task file
	taskDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		fmt.Printf("Failed to create task dir: %v\n", err)
		return
	}

	taskContent := `---
task_name: test-task
---
# Test Task
`
	if err := os.WriteFile(filepath.Join(taskDir, "test.md"), []byte(taskContent), 0o644); err != nil {
		fmt.Printf("Failed to write task file: %v\n", err)
		return
	}

	// Custom command runner that prevents actual execution
	cmdRunner := func(cmd *exec.Cmd) error {
		// Mock execution - do nothing
		return nil
	}

	var output, logOut bytes.Buffer

	ctx := codingcontext.New(
		codingcontext.WithWorkDir(tmpDir),
		codingcontext.WithOutput(&output),
		codingcontext.WithLogOutput(&logOut),
		codingcontext.WithCmdRunner(cmdRunner),
	)

	if err := ctx.Run(context.Background(), "test-task"); err != nil {
		fmt.Printf("Failed to run context: %v\n", err)
		return
	}

	fmt.Println("Custom command runner used successfully")
	// Output: Custom command runner used successfully
}

// TestExampleRun ensures the example code actually works
func TestExampleRun(t *testing.T) {
	Example()
	ExampleWithCmdRunner()
}
