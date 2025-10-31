package main

import (
	"crypto/sha256"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed bootstrap
var bootstrap string

var (
	dirs      stringSlice
	outputDir = "."
	params    = make(paramMap)
)

func main() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	dirs = []string{
		".coding-agent-context",
		filepath.Join(userConfigDir, "coding-agent-context"),
		"/var/local/coding-agent-context",
	}

	flag.Var(&dirs, "d", "Directory to include in the context. Can be specified multiple times.")
	flag.StringVar(&outputDir, "o", ".", "Directory to write the context files to.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  coding-agent-context <task-name> ")
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

	for _, dir := range dirs {
		memoryDir := filepath.Join(dir, "memories")
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

			slog.Info("Including memory file", "path", path)

			content, err := parseMarkdownFile(path, &struct{}{})
			if err != nil {
				return fmt.Errorf("failed to parse markdown file: %w", err)
			}

			// Check for a bootstrap file named <markdown-file-without-md-suffix>-bootstrap
			// For example, setup.md -> setup-bootstrap
			ext := filepath.Ext(path)
			baseNameWithoutExt := path[:len(path)-len(ext)]
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
		promptFile := filepath.Join(dir, "prompts", taskName+".md")

		if _, err := os.Stat(promptFile); err == nil {
			slog.Info("Using prompt file", "path", promptFile)

			content, err := parseMarkdownFile(promptFile, &struct{}{})
			if err != nil {
				return fmt.Errorf("failed to parse prompt file: %w", err)
			}

			t, err := template.New("prompt").Parse(content)
			if err != nil {
				return fmt.Errorf("failed to parse prompt template: %w", err)
			}

			if err := t.Execute(output, params); err != nil {
				return fmt.Errorf("failed to execute prompt template: %w", err)
			}

			return nil

		}
	}

	return fmt.Errorf("prompt file not found for task: %s", taskName)
}
