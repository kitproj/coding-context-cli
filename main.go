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
	"text/template"
)

//go:embed bootstrap
var bootstrap string

var (
	dirs      stringSlice
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
		".github/prompts",
		filepath.Join(userConfigDir, "prompts"),
		"/var/local/prompts",
	}

	flag.Var(&dirs, "d", "Directory to include in the context. Can be specified multiple times.")
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

// findPromptFile looks for a prompt file with the given task name
// Supports both .md and .prompt.md extensions for VS Code compatibility
func findPromptFile(dir, taskName string) (string, error) {
	// Try .prompt.md first (VS Code Copilot format)
	promptFile := filepath.Join(dir, "tasks", taskName+".prompt.md")
	if _, err := os.Stat(promptFile); err == nil {
		return promptFile, nil
	}
	
	// Fall back to .md (standard format)
	promptFile = filepath.Join(dir, "tasks", taskName+".md")
	if _, err := os.Stat(promptFile); err == nil {
		return promptFile, nil
	}
	
	return "", os.ErrNotExist
}

// convertVSCodeVariables converts VS Code variable syntax ${var} to Go template syntax {{ .var }}
// Also handles input variables like ${input:varName} -> {{ .varName }}
func convertVSCodeVariables(content string) string {
	result := content
	
	for i := 0; i < len(result); {
		// Check for ${input:varName} or ${input:varName:placeholder}
		if i+8 <= len(result) && result[i:i+8] == "${input:" {
			// Find the closing }
			end := i + 8
			for end < len(result) && result[end] != '}' {
				end++
			}
			if end < len(result) {
				// Extract the content between ${input: and }
				varPart := result[i+8 : end]
				// Split by : to get variable name (ignore placeholder if present)
				colonIdx := strings.Index(varPart, ":")
				var varName string
				if colonIdx >= 0 {
					varName = varPart[:colonIdx]
				} else {
					varName = varPart
				}
				
				// Replace with Go template syntax
				replacement := "{{ ." + varName + " }}"
				result = result[:i] + replacement + result[end+1:]
				i += len(replacement)
				continue
			}
		}
		
		// Check for simple ${varName}
		if i+2 <= len(result) && result[i:i+2] == "${" {
			// Find the closing }
			end := i + 2
			for end < len(result) && result[end] != '}' {
				end++
			}
			if end < len(result) {
				// Extract variable name
				varName := result[i+2 : end]
				
				// Skip special VS Code variables that we don't support
				if strings.HasPrefix(varName, "workspace") || 
				   strings.HasPrefix(varName, "file") || 
				   strings.HasPrefix(varName, "selection") ||
				   strings.HasPrefix(varName, "selected") {
					i = end + 1
					continue
				}
				
				// Replace with Go template syntax
				replacement := "{{ ." + varName + " }}"
				result = result[:i] + replacement + result[end+1:]
				i += len(replacement)
				continue
			}
		}
		i++
	}
	
	return result
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

			// Only process .md and .prompt.md files as memory files
			ext := filepath.Ext(path)
			if ext != ".md" {
				return nil
			}

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

			// Check for a bootstrap file named <markdown-file-without-md-suffix>-bootstrap
			// For example, setup.md -> setup-bootstrap or setup.prompt.md -> setup.prompt-bootstrap
			var baseNameWithoutExt string
			if strings.HasSuffix(path, ".prompt.md") {
				baseNameWithoutExt = strings.TrimSuffix(path, ".prompt.md")
			} else {
				baseNameWithoutExt = strings.TrimSuffix(path, ".md")
			}
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
		promptFile, err := findPromptFile(dir, taskName)
		if err != nil {
			continue
		}

		slog.Info("Using prompt file", "path", promptFile)

		content, err := parseMarkdownFile(promptFile, &struct{}{})
		if err != nil {
			return fmt.Errorf("failed to parse prompt file: %w", err)
		}

		// Convert VS Code variable syntax to Go template syntax
		content = convertVSCodeVariables(content)

		t, err := template.New("prompt").Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse prompt template: %w", err)
		}

		if err := t.Execute(output, params); err != nil {
			return fmt.Errorf("failed to execute prompt template: %w", err)
		}

		return nil
	}

	return fmt.Errorf("prompt file not found for task: %s", taskName)
}
