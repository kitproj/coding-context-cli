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
	var listTasks bool
	var agent codingcontext.Agent
	params := make(codingcontext.Params)
	includes := make(codingcontext.Selectors)
	var remotePaths []string

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.BoolVar(&resume, "r", false, "Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.")
	flag.BoolVar(&listTasks, "list-tasks", false, "List all available tasks and exit.")
	flag.Var(&agent, "a", "Target agent to use (excludes rules from other agents). Supported agents: cursor, opencode, copilot, claude, gemini, augment, windsurf, codex.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Func("d", "Remote directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, etc.).", func(s string) error {
		remotePaths = append(remotePaths, s)
		return nil
	})

	flag.Usage = func() {
		logger.Info("Usage:")
		logger.Info("  coding-context [options] <task-name>")
		logger.Info("  coding-context --list-tasks")
		logger.Info("")
		logger.Info("Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	cc := codingcontext.New(
		codingcontext.WithWorkDir(workDir),
		codingcontext.WithParams(params),
		codingcontext.WithSelectors(includes),
		codingcontext.WithRemotePaths(remotePaths),
		codingcontext.WithLogger(logger),
		codingcontext.WithResume(resume),
		codingcontext.WithAgent(agent),
	)

	// Handle list-tasks mode
	if listTasks {
		tasks, err := cc.ListTasks(ctx)
		if err != nil {
			logger.Error("Error listing tasks", "error", err)
			os.Exit(1)
		}

		if len(tasks) == 0 {
			logger.Info("No tasks found")
			return
		}

		fmt.Println("Available tasks:")
		fmt.Println()

		for _, task := range tasks {
			fmt.Printf("  %s", task.TaskName)

			// Add variant info if present
			if task.Resume {
				fmt.Print(" (resume)")
			}
			if len(task.Selectors) > 0 {
				fmt.Print(" [")
				first := true
				for k, v := range task.Selectors {
					if !first {
						fmt.Print(", ")
					}
					fmt.Printf("%s=%v", k, v)
					first = false
				}
				fmt.Print("]")
			}

			fmt.Println()

			// Print description if available
			if task.Description != "" {
				fmt.Printf("    %s\n", task.Description)
			}

			// Print path for reference
			logger.Info("Task file", "task", task.TaskName, "path", task.Path)
		}

		return
	}

	args := flag.Args()
	if len(args) != 1 {
		logger.Error("Error", "error", fmt.Errorf("invalid usage"))
		flag.Usage()
		os.Exit(1)
	}

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
