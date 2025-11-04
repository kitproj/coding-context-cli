// Package lib provides a visitor pattern API for processing markdown files with YAML frontmatter.
//
// The package allows you to parse markdown files that may contain YAML frontmatter
// and process them using a visitor function. This is useful for extracting metadata
// and content from markdown files in a consistent way.
//
// Basic usage:
//
//	visitor := func(path string, frontMatter lib.FrontMatter, content string) error {
//	    // Access the file path
//	    fmt.Printf("Processing: %s\n", path)
//	    
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
//
// The package provides three main functions:
//   - Visit: processes files matching a glob pattern
//   - VisitPath: processes a single file or directory recursively
//   - VisitPaths: processes multiple files/directories
package lib
