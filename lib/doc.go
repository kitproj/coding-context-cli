// Package lib provides a visitor pattern API for processing markdown files with YAML frontmatter.
//
// The package allows you to parse markdown files that may contain YAML frontmatter
// and process them using a visitor function. This is useful for extracting metadata
// and content from markdown files in a consistent way.
//
// Basic usage:
//
//	visitor := func(frontMatter lib.FrontMatter, content string) error {
//	    // Access frontmatter fields
//	    if title, ok := frontMatter["title"].(string); ok {
//	        fmt.Printf("Title: %s\n", title)
//	    }
//	    
//	    // Process content
//	    fmt.Printf("Content: %s\n", content)
//	    
//	    return nil
//	}
//	
//	// Visit all markdown files matching the pattern
//	if err := lib.Visit("*.md", visitor); err != nil {
//	    log.Fatal(err)
//	}
//
// The visitor function is called for each markdown file that matches the pattern.
// Processing stops on the first error returned by the visitor.
package lib
