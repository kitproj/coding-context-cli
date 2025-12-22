package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	yaml "github.com/goccy/go-yaml"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/selectors"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	var workDir string
	var resume bool
	var writeRules bool
	var agent codingcontext.Agent
	params := make(taskparser.Params)
	includes := make(selectors.Selectors)
	var searchPaths []string
	var manifestURL string

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.BoolVar(&resume, "r", false, "Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.")
	flag.BoolVar(&writeRules, "w", false, "Write rules to the agent's user rules path and only print the prompt to stdout. Requires agent (via task 'agent' field or -a flag).")
	flag.Var(&agent, "a", "Target agent to use. Required when using -w to write rules to the agent's user rules path. Supported agents: cursor, opencode, copilot, claude, gemini, augment, windsurf, codex.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Func("d", "Directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, file:// etc.).", func(s string) error {
		searchPaths = append(searchPaths, s)
		return nil
	})
	flag.StringVar(&manifestURL, "m", "", "Go Getter URL to a manifest file containing search paths (one per line). Every line is included as-is.")

	flag.Usage = func() {
		logger.Info("Usage:")
		logger.Info("  coding-context [options] <task-name> [user-prompt]")
		logger.Info("")
		logger.Info("The task-name is the name of a task file to look up in task search paths (.agents/tasks).")
		logger.Info("The user-prompt is optional text to append to the task. It can contain slash commands")
		logger.Info("(e.g., '/command-name') which will be expanded, and parameter substitution (${param}).")
		logger.Info("")
		logger.Info("Task content can contain slash commands (e.g., '/command-name arg') which reference")
		logger.Info("command files in command search paths (.cursor/commands, .agents/commands, etc.).")
		logger.Info("")
		logger.Info("Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		logger.Error("Error", "error", fmt.Errorf("invalid usage: expected one task name argument and optional user-prompt"))
		flag.Usage()
		os.Exit(1)
	}

	taskName := args[0]
	var userPrompt string
	if len(args) == 2 {
		userPrompt = args[1]
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Error", "error", fmt.Errorf("failed to get user home directory: %w", err))
		os.Exit(1)
	}

	searchPaths = append(searchPaths, "file://"+workDir)
	searchPaths = append(searchPaths, "file://"+homeDir)

	cc := codingcontext.New(
		codingcontext.WithParams(params),
		codingcontext.WithSelectors(includes),
		codingcontext.WithSearchPaths(searchPaths...),
		codingcontext.WithLogger(logger),
		codingcontext.WithResume(resume),
		codingcontext.WithAgent(agent),
		codingcontext.WithManifestURL(manifestURL),
		codingcontext.WithUserPrompt(userPrompt),
	)

	result, err := cc.Run(ctx, taskName)
	if err != nil {
		logger.Error("Error", "error", err)
		flag.Usage()
		os.Exit(1)
	}

	// If writeRules flag is set, write rules to UserRulePath and only output task
	if writeRules {
		// Get the user rule path from the agent (could be from task or -a flag)
		if !result.Agent.IsSet() {
			logger.Error("Error", "error", fmt.Errorf("-w flag requires an agent to be specified (via task 'agent' field or -a flag)"))
			os.Exit(1)
		}

		// Skip writing rules file in resume mode since no rules are collected
		if !resume {
			relativePath := result.Agent.UserRulePath()
			if relativePath == "" {
				logger.Error("Error", "error", fmt.Errorf("no user rule path available for agent"))
				os.Exit(1)
			}

			// Construct full path by joining with home directory
			rulesFile := filepath.Join(homeDir, relativePath)
			rulesDir := filepath.Dir(rulesFile)

			// Create directory if it doesn't exist
			if err := os.MkdirAll(rulesDir, 0o755); err != nil {
				logger.Error("Error", "error", fmt.Errorf("failed to create rules directory %s: %w", rulesDir, err))
				os.Exit(1)
			}

			// Build rules content, trimming each rule and joining with consistent spacing
			var rulesContent strings.Builder
			for i, rule := range result.Rules {
				if i > 0 {
					rulesContent.WriteString("\n\n")
				}
				rulesContent.WriteString(strings.TrimSpace(rule.Content))
			}
			rulesContent.WriteString("\n")

			if err := os.WriteFile(rulesFile, []byte(rulesContent.String()), 0o644); err != nil {
				logger.Error("Error", "error", fmt.Errorf("failed to write rules to %s: %w", rulesFile, err))
				os.Exit(1)
			}

			logger.Info("Rules written", "path", rulesFile)
		}

		// Output only task frontmatter and content
		if taskContent := result.Task.FrontMatter.Content; taskContent != nil {
			fmt.Println("---")
			if err := yaml.NewEncoder(os.Stdout).Encode(taskContent); err != nil {
				logger.Error("Failed to encode task frontmatter", "error", err)
				os.Exit(1)
			}
			fmt.Println("---")
		}
		fmt.Println(result.Task.Content)
	} else {
		// Normal mode: output everything
		// Output task frontmatter (always enabled)
		if taskContent := result.Task.FrontMatter.Content; taskContent != nil {
			fmt.Println("---")
			if err := yaml.NewEncoder(os.Stdout).Encode(taskContent); err != nil {
				logger.Error("Failed to encode task frontmatter", "error", err)
				os.Exit(1)
			}
			fmt.Println("---")
		}

		// Output all rules
		for _, rule := range result.Rules {
			fmt.Println(rule.Content)
		}

		// Output task
		fmt.Println(result.Task.Content)
	}
}
