package main

import (
	"fmt"
)

// Example demonstrates how to use the library API to process markdown files
func Example_usingVisitor() {
	// Define a visitor function that processes each markdown file
	visitor := func(frontMatter FrontMatter, content string) error {
		// Access frontmatter fields
		if title, ok := frontMatter["title"].(string); ok {
			fmt.Printf("Title: %s\n", title)
		}
		
		// Process content
		fmt.Printf("Content length: %d bytes\n", len(content))
		
		return nil
	}
	
	// Visit all markdown files in a directory
	if err := Visit("*.md", visitor); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}

// Example demonstrates stopping on first error
func Example_stoppingOnError() {
	visitor := func(frontMatter FrontMatter, content string) error {
		// Check for required fields
		if _, ok := frontMatter["required_field"]; !ok {
			return fmt.Errorf("missing required field in frontmatter")
		}
		
		// Process the file...
		return nil
	}
	
	// Visit will stop on the first error
	if err := Visit("*.md", visitor); err != nil {
		fmt.Printf("Stopped processing: %v\n", err)
		return
	}
}
