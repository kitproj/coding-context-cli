package main

import (
	"context"
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

var (
	workDir  string
	rules    map[string]string
	params   = make(paramMap)
	includes = make(selectorMap)
	excludes = make(selectorMap)
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Var(&excludes, "S", "Exclude rules with matching frontmatter. Can be specified multiple times as key=value.")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  coding-context [options] <agent-name> <task-name>")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(ctx, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("invalid usage")
	}

	if err := os.Chdir(workDir); err != nil {
		return fmt.Errorf("failed to chdir to %s: %w", workDir, err)
	}

	agentName := args[0]
	taskName := args[1]
	includes["task_name"] = taskName

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	for source, target := range map[string]string{

		// --- USER RULES (GLOBAL PERSUASION, HOME DIRECTORY) ---
		// These files define the agent's persona and rules across all projects.
		// Standardized value format: ~/.agents/rules/<purpose>.md
		homeDir + "/.claude/CLAUDE.md": homeDir + ".agents/rules/CLAUDE.md",  // Claude Code global persona [1]
		homeDir + "/.gemini/GEMINI.md": homeDir + "/.agents/rules/GEMINI.md", // Gemini CLI global context [2]
		homeDir + "/.codex/AGENTS.md":  homeDir + ".agents/rules/CODEX.md",   // Codex CLI global guidance [3]

		// --- PROJECT RULES (ANCSTOR & PWD-SPECIFIC GUIDANCE) ---
		// These files/directories are located within the repository or CWD hierarchy.
		// Standardized value format:.agents/rules/<location/purpose>

		// Standardized Static Rule Files (e.g., AGENTS.md, CLAUDE.md, GEMINI.md)
		"./CLAUDE.md":        ".agents/rules/CLAUDE.md",   // Claude Code project root instructions [1]
		"./AGENTS.md":        ".agents/rules/AGENTS.md",   // Codex/Cursor/Copilot root guidance [3, 4]
		"./GEMINI.md":        ".agents/rules/GEMINI.md",   // Gemini CLI ancestor context [2]
		"../AGENTS.md":       ".agents/rules/AGENTS.1.md", // Copilot/Codex ancestor AGENTS.md [5]
		"../../AGENTS.md":    ".agents/rules/AGENTS.2.md", // Codex/Cursor sub-directory specific guidance [3, 4]
		"../../../AGENTS.md": ".agents/rules/AGENTS.3.md", // Codex/Cursor sub-directory specific guidance [3, 4]

		// Tool-Specific/Specialized Rule Files
		"./CLAUDE.local.md":               ".agents/rules/CLAUDE.local.md",         // Claude Code private project override [1]
		".gemini/styleguide.md":           ".agents/rules/gemini_styleguide.md",    // Gemini CLI code review style guide [6]
		".augment/guidelines.md":          ".agents/rules/augment_guidelines.md",   // Augment CLI legacy guidance file [7]
		".github/copilot-instructions.md": ".agents/rules/copilot_instructions.md", // Copilot workspace instructions [8]

		// Structured Rule Directories (These keys represent recursive rule folders)
		// Standardized value format:.agents/rules/structured_dir/
		".cursor/rules":   ".agents/rules/cursor",   // Cursor IDE declarative rules folder [4]
		".augment/rules":  ".agents/rules/augment",  // Augment CLI custom rules folder (supports frontmatter) [7]
		".windsurf/rules": ".agents/rules/windsurf", // Windsurf rules folder (searched recursively) [9]
		".github/agents":  ".agents/rules/github",   // GitHub Copilot custom agent definition folder
	} {
		// Skip if the path doesn't exist
		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
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

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := os.WriteFile(target, []byte(content), 0600); err != nil {
				return fmt.Errorf("failed to write to rules file: %w", err)
			}

			// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
			// For example, setup.md -> setup-bootstrap, setup.mdc -> setup-bootstrap
			baseNameWithoutExt := strings.TrimSuffix(path, ext)
			bootstrapSoucePath := baseNameWithoutExt + "-bootstrap"

			bootstrapContent, err := os.ReadFile(bootstrapSoucePath)
			if os.IsNotExist(err) {
				return nil
			} else if err != nil {
				return err
			}

			bootstrapTargetPath := filepath.Join(filepath.Dir(target), filepath.Base(bootstrapSoucePath))

			if err := os.WriteFile(bootstrapTargetPath, bootstrapContent, 0600); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk rule dir: %w", err)
		}
	}

	// Track total tokens
	var totalTokens int

	// AgentRulesPathLookup maps agent tool names to their specific rule and instruction file paths.
	// Ancestral paths are modeled explicitly using the current directory (./) and one parent directory (../).
	var agentRulesPathLookup = map[string]map[string]string{
		"ClaudeCode": {
			".agents/rules":                            "CLAUDE.md",
			"../.agents/rules":                         "../CLAUDE.md",
			"../../.agents/rules":                      "../../CLAUDE.md",
			filepath.Join(homeDir, ".agents", "rules"): filepath.Join(homeDir, ".claude", "CLAUDE.md"),
			"/etc/agents/rules":                        "CLAUDE.md",
		},
		"Cursor": {
			".agents/rules":                            "./.cursor/rules",
			"../.agents/rules":                         "../AGENTS.md",
			"../../.agents/rules":                      "../../AGENTS.md",
			filepath.Join(homeDir, ".agents", "rules"): "AGENTS.md",
			"/etc/agents/rules":                        "AGENTS.md",
		},
		"Windsurf": {
			".agents/rules":                            "./.windsurf/rules",
			"../.agents/rules":                         "../.windsurf/rules",
			"../../.agents/rules":                      "../../.windsurf/rules",
			filepath.Join(homeDir, ".agents", "rules"): "./.windsurf/rules",
			"/etc/agents/rules":                        ".windsurfrules",
		},
		"Codex": {
			".agents/rules":                            "AGENTS.md",
			"../.agents/rules":                         "../AGENTS.md",
			"../../.agents/rules":                      "../../AGENTS.md",
			filepath.Join(homeDir, ".agents", "rules"): filepath.Join(homeDir, ".codex", "AGENTS.md"),
			"/etc/agents/rules":                        "AGENTS.md",
		},
		"GitHubCopilot": {
			".agents/rules":                            ".github/copilot-instructions.md",
			"../.agents/rules":                         "../AGENTS.md",
			"../../.agents/rules":                      "../../AGENTS.md",
			filepath.Join(homeDir, ".agents", "rules"): ".github/copilot-instructions.md",
			"/etc/agents/rules":                        "AGENTS.md",
		},
		"AugmentCLI": {
			".agents/rules":                            ".augment/rules",
			"../.agents/rules":                         "../AGENTS.md",
			"../../.agents/rules":                      "../../AGENTS.md",
			filepath.Join(homeDir, ".agents", "rules"): "AGENTS.md",
			"/etc/agents/rules":                        "AGENTS.md",
		},
		"Goose": {
			".agents/rules":                            ".goosehints",
			"../.agents/rules":                         ".goosehints",
			"../../.agents/rules":                      ".goosehints",
			filepath.Join(homeDir, ".agents", "rules"): filepath.Join(homeDir, ".config", "goose", ".goosehints"),
			"/etc/agents/rules":                        ".goosehints",
		},
		"Gemini": {
			".agents/rules":                            "GEMINI.md",
			"../.agents/rules":                         "../GEMINI.md",
			"../../.agents/rules":                      "../../GEMINI.md",
			filepath.Join(homeDir, ".agents", "rules"): filepath.Join(homeDir, ".gemini", "GEMINI.md"),
			"/etc/agents/rules":                        "GEMINI.md",
		},
	}

	agentRulePaths := agentRulesPathLookup[agentName]

	// delete every target path
	for _, target := range agentRulePaths {
		err := os.RemoveAll(target)
		if os.IsExist(err) {
			continue
		} else if err != nil {
			return err
		}
	}

	for _, rulePath := range []string{
		".agents/rules",
		filepath.Join(userConfigDir, "agents", "rules"),
		"/etc/agents/tasks",
	} {

		err = filepath.Walk(rulePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only process .md files as rule files
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

			// Check if file matches include and exclude selectors.
			// Note: Files with duplicate basenames will both be included.
			if !includes.matchesIncludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (does not match include selectors): %s\n", path)
				return nil
			}

			if !excludes.matchesExcludes(frontmatter) {
				fmt.Fprintf(os.Stderr, "Excluding rule file (matches exclude selectors): %s\n", path)
				return nil
			}

			target, ok := agentRulePaths[rulePath]
			if !ok {
				panic("failed to look up path, this is a bug")
			}

			targetIsDir := filepath.Ext(target) == ""

			if targetIsDir {
				target = filepath.Join(target, path)
			}

			f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			if _, err := f.WriteString(content); err != nil {
				return fmt.Errorf("failed to write to rules file: %w", err)
			}
			if err := f.Close(); err != nil {
				return err
			}

			totalTokens += estimateTokens(content)

			// Check for a bootstrap file named <markdown-file-without-md/mdc-suffix>-bootstrap
			// For example, setup.md -> setup-bootstrap, setup.mdc -> setup-bootstrap
			baseNameWithoutExt := strings.TrimSuffix(path, ext)
			bootstrapSoucePath := baseNameWithoutExt + "-bootstrap"

			_, err = os.Stat(bootstrapSoucePath)
			if os.IsNotExist(err) {
				return nil
			} else if err != nil {
				return err
			}

			cmd := exec.CommandContext(ctx, bootstrapSoucePath)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}

			return nil
		})
	}

	for _, path := range []string{
		".agents/tasks",
		filepath.Join(userConfigDir, "agents", "tasks"),
		"/etc/agents/tasks",
	} {
		stat, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to stat task path %s: %w", path, err)
		}
		if stat.IsDir() {
			path = filepath.Join(path, taskName+".md")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return fmt.Errorf("failed to stat task file %s: %w", path, err)
			}
		}

		content, err := parseMarkdownFile(path, &struct{}{})
		if err != nil {
			return fmt.Errorf("failed to parse prompt file: %w", err)
		}

		expanded := os.Expand(content, func(key string) string {
			if val, ok := params[key]; ok {
				return val
			}
			// this might not exist, in that case, return the original text
			return fmt.Sprintf("${%s}", key)
		})

		// Estimate tokens for this file
		tokens := estimateTokens(expanded)
		totalTokens += tokens
		fmt.Fprintf(os.Stderr, "Using task file: %s (~%d tokens)\n", path, tokens)

		fmt.Fprintln(os.Stdout, expanded)

		// Print total token count
		fmt.Fprintf(os.Stderr, "Total estimated tokens: %d\n", totalTokens)

		return nil
	}

	return fmt.Errorf("prompt file not found for task: %s", taskName)
}
