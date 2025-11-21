package codingcontext_test

import (
	"context"
	"fmt"
	"log"

	"github.com/goccy/go-yaml"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

// ExampleMarkdown_FrontMatter demonstrates how to access task frontmatter
// when using the coding-context library.
func ExampleMarkdown_FrontMatter() {
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

	// Access the task frontmatter Content map directly
	taskName, _ := result.Task.FrontMatter.Content["task_name"].(string)
	priority, _ := result.Task.FrontMatter.Content["priority"].(string)
	environment, _ := result.Task.FrontMatter.Content["environment"].(string)

	// Now you can use the frontmatter values
	fmt.Printf("Task: %s\n", taskName)
	fmt.Printf("Priority: %s\n", priority)
	fmt.Printf("Environment: %s\n", environment)

	// You can also access rule frontmatter the same way
	for _, rule := range result.Rules {
		if language, ok := rule.FrontMatter.Content["language"].(string); ok {
			if stage, ok := rule.FrontMatter.Content["stage"].(string); ok {
				fmt.Printf("Rule: language=%s, stage=%s\n", language, stage)
			}
		}
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

	// Parse the markdown file into a FrontMatter
	var frontmatterMap codingcontext.FrontMatter
	content, err := codingcontext.ParseMarkdownFile("path/to/task.md", &frontmatterMap)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the Content into your struct if needed
	var frontmatter TaskFrontmatter
	yamlBytes, _ := yaml.Marshal(frontmatterMap.Content)
	yaml.Unmarshal(yamlBytes, &frontmatter)

	// Access the parsed frontmatter
	fmt.Printf("Task: %s\n", frontmatter.TaskName)
	fmt.Printf("Resume: %v\n", frontmatter.Resume)
	fmt.Printf("Priority: %s\n", frontmatter.Priority)
	fmt.Printf("Tags: %v\n", frontmatter.Tags)

	// Access the content
	fmt.Printf("Content length: %d\n", len(content))
}
