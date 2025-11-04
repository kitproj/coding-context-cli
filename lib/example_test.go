package lib_test

import (
	"fmt"
	"log"

	"github.com/kitproj/coding-context-cli/lib"
)

// Example demonstrates how to use the visitor pattern to process markdown files.
func Example() {
	// Define a visitor function that processes each markdown file
	visitor := func(path string, frontMatter lib.FrontMatter, content string) error {
		// Access frontmatter fields
		if title, ok := frontMatter["title"].(string); ok {
			fmt.Printf("Title: %s\n", title)
		}
		
		// Process content
		fmt.Printf("Content length: %d bytes\n", len(content))
		
		return nil
	}
	
	// Visit all markdown files in a directory
	if err := lib.Visit("testdata/*.md", visitor); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// Example_stoppingOnError demonstrates how the visitor stops on the first error.
func Example_stoppingOnError() {
	visitor := func(path string, frontMatter lib.FrontMatter, content string) error {
		// Check for required fields
		if _, ok := frontMatter["required_field"]; !ok {
			return fmt.Errorf("missing required field in frontmatter")
		}
		
		// Process the file...
		return nil
	}
	
	// Visit will stop on the first error
	if err := lib.Visit("*.md", visitor); err != nil {
		fmt.Printf("Stopped processing: %v\n", err)
		return
	}
}
