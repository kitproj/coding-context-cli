package main

import (
	"crypto/sha256"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

//go:embed bootstrap
var bootstrap string

var (
	dirs      stringSlice
	files     stringSlice
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

	flag.Var(&dirs, "d", "Directory to include in the context. Can be specified multiple times.")
	flag.Var(&files, "f", "Specific file to include as memory. Can be specified multiple times.")
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

	// Process specific files first (if provided)
	for _, file := range files {
		// Resolve relative paths
		absPath := file
		if !filepath.IsAbs(file) {
			var err error
			absPath, err = filepath.Abs(file)
			if err != nil {
				return fmt.Errorf("failed to resolve path %s: %w", file, err)
			}
		}

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("specific file not found: %s", file)
		} else if err != nil {
			return fmt.Errorf("failed to stat specific file %s: %w", file, err)
		}

		// Process the file
		if err := processMemoryFile(absPath, output, bootstrapDir); err != nil {
			return fmt.Errorf("failed to process specific file %s: %w", file, err)
		}
	}

	// Process memory directories
	for _, dir := range dirs {
		memoryDir := filepath.Join(dir, "memories")
		
		// Skip if the directory doesn't exist
		if _, err := os.Stat(memoryDir); os.IsNotExist(err) {
			continue
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
		if err != nil {
			return fmt.Errorf("failed to walk memory dir: %w", err)
		}
	}

	if err := os.WriteFile(filepath.Join(outputDir, "bootstrap"), []byte(bootstrap), 0755); err != nil {
		return fmt.Errorf("failed to write bootstrap file: %w", err)
	}

	taskName := args[0]
	for _, dir := range dirs {
		promptFile := filepath.Join(dir, "tasks", taskName+".md")

		if _, err := os.Stat(promptFile); err == nil {
			slog.Info("Using prompt file", "path", promptFile)

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
	}

	return fmt.Errorf("prompt file not found for task: %s", taskName)
}

// processMemoryFile processes a single memory file and writes its content to the output
func processMemoryFile(path string, output *os.File, bootstrapDir string) error {
	// Parse frontmatter to check selectors
	var frontmatter map[string]string
	content, err := parseMarkdownFile(path, &frontmatter)
	if err != nil {
		return fmt.Errorf("failed to parse markdown file: %w", err)
	}

	// Check if file matches include and exclude selectors
	if !includes.matchesIncludes(frontmatter) {
		slog.Info("Excluding memory file (does not match include selectors)", "path", path)
		return nil
	}
	if !excludes.matchesExcludes(frontmatter) {
		slog.Info("Excluding memory file (matches exclude selectors)", "path", path)
		return nil
	}

	slog.Info("Including memory file", "path", path)

	// Check for a bootstrap file
	// For .md files: setup.md -> setup-bootstrap
	// For other files: .cursorrules -> .cursorrules-bootstrap
	var bootstrapFilePath string
	if filepath.Ext(path) == ".md" {
		baseNameWithoutExt := strings.TrimSuffix(path, ".md")
		bootstrapFilePath = baseNameWithoutExt + "-bootstrap"
	} else {
		bootstrapFilePath = path + "-bootstrap"
	}

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
