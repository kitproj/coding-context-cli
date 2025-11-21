package codingcontext_test

import (
	"context"
	"fmt"
	"log"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

// TaskMetadata represents the custom frontmatter structure for a task
type TaskMetadata struct {
	TaskName    string         `yaml:"task_name"`
	Resume      bool           `yaml:"resume"`
	Priority    string         `yaml:"priority"`
	Environment string         `yaml:"environment"`
	Selectors   map[string]any `yaml:"selectors"`
}

// ExampleResult_ParseTaskFrontmatter demonstrates how to parse task frontmatter
// into a custom struct when using the coding-context library.
func ExampleResult_ParseTaskFrontmatter() {
	// Create a context and run it to get a result
	// In a real application, you would configure this properly
	cc := codingcontext.New(
		codingcontext.WithWorkDir("."),
	)

	// Assuming there's a task file with frontmatter like:
	// ---
	// task_name: deploy
	// priority: high
	// environment: production
	// ---
	result, err := cc.Run(context.Background(), "deploy")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the task frontmatter into your custom struct
	var taskMeta TaskMetadata
	if err := result.ParseTaskFrontmatter(&taskMeta); err != nil {
		log.Fatal(err)
	}

	// Now you can use the strongly-typed task metadata
	fmt.Printf("Task: %s\n", taskMeta.TaskName)
	fmt.Printf("Priority: %s\n", taskMeta.Priority)
	fmt.Printf("Environment: %s\n", taskMeta.Environment)

	// You can also access the generic frontmatter map directly
	if priority, ok := result.Task.FrontMatter["priority"]; ok {
		fmt.Printf("Priority from map: %v\n", priority)
	}
}

// ExampleParseMarkdownFile demonstrates how to parse a markdown file
// with frontmatter into a custom struct.
func ExampleParseMarkdownFile() {
	// Define your custom struct with yaml tags
	type TaskFrontmatter struct {
		TaskName string   `yaml:"task_name"`
		Resume   bool     `yaml:"resume"`
		Priority string   `yaml:"priority"`
		Tags     []string `yaml:"tags"`
	}

	// Parse the markdown file
	var frontmatter TaskFrontmatter
	content, err := codingcontext.ParseMarkdownFile("path/to/task.md", &frontmatter)
	if err != nil {
		log.Fatal(err)
	}

	// Access the parsed frontmatter
	fmt.Printf("Task: %s\n", frontmatter.TaskName)
	fmt.Printf("Resume: %v\n", frontmatter.Resume)
	fmt.Printf("Priority: %s\n", frontmatter.Priority)
	fmt.Printf("Tags: %v\n", frontmatter.Tags)

	// Access the content
	fmt.Printf("Content length: %d\n", len(content))
}
