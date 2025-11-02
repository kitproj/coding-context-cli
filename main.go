package main

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

//go:embed bootstrap
var bootstrap string

var (
	workDir   string
	outputDir = "."
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.StringVar(&outputDir, "o", ".", "Directory to write the context files to.")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:\n")
		fmt.Fprintf(w, "  coding-context <command> [options] [arguments]\n\n")
		fmt.Fprintln(w, "Commands:")
		fmt.Fprintln(w, "  import <agent>  Import rules for the specified agent")
		fmt.Fprintln(w, "  export <agent>  Export rules for the specified agent (TODO)")
		fmt.Fprintln(w, "  bootstrap       Run bootstrap scripts")
		fmt.Fprintf(w, "  prompt          Find and print prompts (TODO)\n\n")
		fmt.Fprintln(w, "Global Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Change to work directory
	if err := os.Chdir(workDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to chdir to %s: %v\n", workDir, err)
		os.Exit(1)
	}

	// Initialize agent rules
	if err := initAgentRules(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize agent rules: %v\n", err)
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "import":
		if err := runImport(ctx, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "export":
		fmt.Fprintln(os.Stderr, "Error: export command not yet implemented")
		os.Exit(1)
	case "bootstrap":
		if err := runBootstrapCommand(ctx, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "prompt":
		fmt.Fprintln(os.Stderr, "Error: prompt command not yet implemented")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func runImport(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: coding-context import <agent>")
	}

	agentName := Agent(args[0])

	// Check if agent is valid
	rulePaths, ok := agentRules[agentName]
	if !ok {
		return fmt.Errorf("unknown agent: %s", agentName)
	}

	// Expand ancestor paths
	rulePaths = expandAncestorPaths(rulePaths)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	bootstrapDir := filepath.Join(outputDir, "bootstrap.d")
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		return fmt.Errorf("failed to create bootstrap dir: %w", err)
	}

	// Track total tokens
	var totalTokens int

	// Create rules.md file
	rulesOutput, err := os.Create(filepath.Join(outputDir, "rules.md"))
	if err != nil {
		return fmt.Errorf("failed to create rules file: %w", err)
	}
	defer rulesOutput.Close()

	// Process each rule path
	for _, rp := range rulePaths {
		// Skip if the path doesn't exist
		if _, err := os.Stat(rp.Path); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(rp.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Only process .md and .mdc files as rule files
			ext := filepath.Ext(path)
			if ext != ".md" && ext != ".mdc" {
				return nil
			}

			// Parse frontmatter
			var frontmatter map[string]string
			content, err := parseMarkdownFile(path, &frontmatter)
			if err != nil {
				return fmt.Errorf("failed to parse markdown file: %w", err)
			}

			// Estimate tokens for this file
			tokens := estimateTokens(content)
			totalTokens += tokens
			fmt.Fprintf(os.Stdout, "Including rule file: %s (level %d, ~%d tokens)\n", path, rp.Level, tokens)

			// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
			baseNameWithoutExt := strings.TrimSuffix(path, ext)
			bootstrapFilePath := baseNameWithoutExt + "-bootstrap"

			if bootstrapContent, err := os.ReadFile(bootstrapFilePath); err == nil {
				hash := sha256.Sum256(bootstrapContent)
				baseBootstrapName := filepath.Base(bootstrapFilePath)
				bootstrapFileName := fmt.Sprintf("%s-%08x", baseBootstrapName, hash[:4])
				bootstrapPath := filepath.Join(bootstrapDir, bootstrapFileName)
				if err := os.WriteFile(bootstrapPath, bootstrapContent, 0700); err != nil {
					return fmt.Errorf("failed to write bootstrap file: %w", err)
				}
			}

			if _, err := rulesOutput.WriteString(content + "\n\n"); err != nil {
				return fmt.Errorf("failed to write to rules file: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk rule path: %w", err)
		}
	}

	if err := os.WriteFile(filepath.Join(outputDir, "bootstrap"), []byte(bootstrap), 0755); err != nil {
		return fmt.Errorf("failed to write bootstrap file: %w", err)
	}

	// Print total token count
	fmt.Fprintf(os.Stdout, "Total estimated tokens: %d\n", totalTokens)

	return nil
}

func runBootstrapCommand(ctx context.Context, args []string) error {
	bootstrapPath := filepath.Join(outputDir, "bootstrap")

	// Convert to absolute path
	absBootstrapPath, err := filepath.Abs(bootstrapPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for bootstrap script: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Running bootstrap script: %s\n", absBootstrapPath)

	cmd := exec.CommandContext(ctx, absBootstrapPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = outputDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run bootstrap script: %w", err)
	}

	return nil
}
