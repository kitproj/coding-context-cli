package main

import (
	"crypto/sha256"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed bootstrap
var bootstrap string

var (
	dirs      stringSlice
	memories  stringSlice
	tasks     stringSlice
	outputDir = "."
	params    = make(paramMap)
	includes  = make(selectorMap)
	excludes  = make(selectorMap)
)

func main() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	dirs = []string{
		".prompts",
		filepath.Join(userConfigDir, "prompts"),
		"/var/local/prompts",
	}

	flag.Var(&dirs, "d", "Directory that may contain a memories/ or tasks/ subdirectory. Can be specified multiple times.")
	flag.Var(&memories, "m", "Directory containing memories, or a single memory file. Can be specified multiple times.")
	flag.Var(&tasks, "t", "Directory containing tasks, or a single task file. Can be specified multiple times.")
	flag.StringVar(&outputDir, "o", ".", "Directory to write the context files to.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include memories with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Var(&excludes, "S", "Exclude memories with matching frontmatter. Can be specified multiple times as key=value.")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  coding-context <task-name> ")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("invalid usage")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	bootstrapDir := filepath.Join(outputDir, "bootstrap.d")
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		return fmt.Errorf("failed to create bootstrap dir: %w", err)
	}

	output, err := os.Create(filepath.Join(outputDir, "prompt.md"))
	if err != nil {
		return fmt.Errorf("failed to create prompt file: %w", err)
	}
	defer output.Close()

	// Process memory files from three sources:
	// 1. From dirs - look for memories/ subdirectory
	// 2. From memories - direct paths to directories or files
	
	// Process memories from dirs
	for _, dir := range dirs {
		memoryDir := filepath.Join(dir, "memories")
		if err := processMemoryDir(memoryDir, output, bootstrapDir); err != nil {
			// Skip if directory doesn't exist
			if !os.IsNotExist(err) {
				return err
			}
		}
	}
	
	// Process memories from direct paths
	for _, memPath := range memories {
		info, err := os.Stat(memPath)
		if err != nil {
			return fmt.Errorf("failed to stat memory path %s: %w", memPath, err)
		}
		
		if info.IsDir() {
			// It's a directory, walk it for .md files
			if err := processMemoryDir(memPath, output, bootstrapDir); err != nil {
				return err
			}
		} else {
			// It's a file, process it directly
			if filepath.Ext(memPath) == ".md" {
				if err := processMemoryFile(memPath, output, bootstrapDir); err != nil {
					return err
				}
			}
		}
	}

	if err := os.WriteFile(filepath.Join(outputDir, "bootstrap"), []byte(bootstrap), 0755); err != nil {
		return fmt.Errorf("failed to write bootstrap file: %w", err)
	}

	// Process task files from three sources:
	// 1. From dirs - look for tasks/ subdirectory
	// 2. From tasks - direct paths to directories or files
	
	taskName := args[0]
	
	// First try to find task in dirs
	for _, dir := range dirs {
		promptFile := filepath.Join(dir, "tasks", taskName+".md")
		if _, err := os.Stat(promptFile); err == nil {
			return processTaskFile(promptFile, output)
		}
	}
	
	// Then try to find task in direct paths
	for _, taskPath := range tasks {
		info, err := os.Stat(taskPath)
		if err != nil {
			return fmt.Errorf("failed to stat task path %s: %w", taskPath, err)
		}
		
		if info.IsDir() {
			// It's a directory, look for taskName.md
			promptFile := filepath.Join(taskPath, taskName+".md")
			if _, err := os.Stat(promptFile); err == nil {
				return processTaskFile(promptFile, output)
			}
		} else {
			// It's a file, check if the name matches
			baseName := strings.TrimSuffix(filepath.Base(taskPath), ".md")
			if baseName == taskName {
				return processTaskFile(taskPath, output)
			}
		}
	}

	return fmt.Errorf("prompt file not found for task: %s", taskName)
}

// processMemoryDir walks a directory and processes all .md files as memory files
func processMemoryDir(memoryDir string, output *os.File, bootstrapDir string) error {
	// Skip if the directory doesn't exist
	if _, err := os.Stat(memoryDir); os.IsNotExist(err) {
		return err
	}
	
	err := filepath.Walk(memoryDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process .md files as memory files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		return processMemoryFile(path, output, bootstrapDir)
	})
	return err
}

// processMemoryFile processes a single memory file
func processMemoryFile(path string, output *os.File, bootstrapDir string) error {
	// Parse frontmatter to check selectors
	var frontmatter map[string]string
	content, err := parseMarkdownFile(path, &frontmatter)
	if err != nil {
		return fmt.Errorf("failed to parse markdown file: %w", err)
	}

	// Check if file matches include and exclude selectors
	if !includes.matchesIncludes(frontmatter) {
		fmt.Fprintf(os.Stdout, "Excluding memory file (does not match include selectors): %s\n", path)
		return nil
	}
	if !excludes.matchesExcludes(frontmatter) {
		fmt.Fprintf(os.Stdout, "Excluding memory file (matches exclude selectors): %s\n", path)
		return nil
	}

	fmt.Fprintf(os.Stdout, "Including memory file: %s\n", path)

	// Check for a bootstrap file named <markdown-file-without-md-suffix>-bootstrap
	// For example, setup.md -> setup-bootstrap
	baseNameWithoutExt := strings.TrimSuffix(path, ".md")
	bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

	if bootstrapContent, err := os.ReadFile(bootstrapFilePath); err == nil {
		hash := sha256.Sum256(bootstrapContent)
		bootstrapPath := filepath.Join(bootstrapDir, fmt.Sprintf("%x", hash))
		if err := os.WriteFile(bootstrapPath, bootstrapContent, 0700); err != nil {
			return fmt.Errorf("failed to write bootstrap file: %w", err)
		}
	}

	if _, err := output.WriteString(content + "\n\n"); err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	return nil
}

// processTaskFile processes a single task file
func processTaskFile(promptFile string, output *os.File) error {
	fmt.Fprintf(os.Stdout, "Using prompt file: %s\n", promptFile)

	content, err := parseMarkdownFile(promptFile, &struct{}{})
	if err != nil {
		return fmt.Errorf("failed to parse prompt file: %w", err)
	}

	expanded := os.Expand(content, func(key string) string {
		if val, ok := params[key]; ok {
			return val
		}
		return ""
	})

	if _, err := output.WriteString(expanded); err != nil {
		return fmt.Errorf("failed to write expanded prompt: %w", err)
	}

	return nil
}
