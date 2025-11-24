package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	yaml "github.com/goccy/go-yaml"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	var workDir string
	var resume bool
	var agent codingcontext.Agent
	params := make(codingcontext.Params)
	includes := make(codingcontext.Selectors)
	var searchPaths []string

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.BoolVar(&resume, "r", false, "Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.")
	flag.Var(&agent, "a", "Target agent to use (excludes rules from other agents). Supported agents: cursor, opencode, copilot, claude, gemini, augment, windsurf, codex.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Func("d", "Directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::,, file:// etc.).", func(s string) error {
		searchPaths = append(searchPaths, s)
		return nil
	})

	flag.Usage = func() {
		logger.Info("Usage:")
		logger.Info("  coding-context [options] <task-name>")
		logger.Info("")
		logger.Info("Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		logger.Error("Error", "error", fmt.Errorf("invalid usage"))
		flag.Usage()
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Error", "error", fmt.Errorf("failed to get user home directory: %w", err))
		os.Exit(1)
	}

	searchPaths = append(searchPaths, homeDir)
	searchPaths = append(searchPaths, workDir)

	cc := codingcontext.New(
		codingcontext.WithWorkDir(workDir),
		codingcontext.WithParams(params),
		codingcontext.WithSelectors(includes),
		codingcontext.WithSearchPaths(searchPaths...),
		codingcontext.WithLogger(logger),
		codingcontext.WithResume(resume),
		codingcontext.WithAgent(agent),
	)

	result, err := cc.Run(ctx, args[0])
	if err != nil {
		logger.Error("Error", "error", err)
		flag.Usage()
		os.Exit(1)
	}

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
