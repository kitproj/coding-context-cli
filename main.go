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
	var emitTaskFrontmatter bool
	params := make(codingcontext.Params)
	includes := make(codingcontext.Selectors)
	var remotePaths []string

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.BoolVar(&resume, "r", false, "Resume mode: skip outputting rules and select task with 'resume: true' in frontmatter.")
	flag.BoolVar(&emitTaskFrontmatter, "t", false, "Print task frontmatter at the beginning of output.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Func("d", "Remote directory containing rules and tasks. Can be specified multiple times. Supports various protocols via go-getter (http://, https://, git::, s3::, etc.).", func(s string) error {
		remotePaths = append(remotePaths, s)
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

	cc := codingcontext.New(
		codingcontext.WithWorkDir(workDir),
		codingcontext.WithResume(resume),
		codingcontext.WithParams(params),
		codingcontext.WithSelectors(includes),
		codingcontext.WithRemotePaths(remotePaths),
		codingcontext.WithEmitTaskFrontmatter(emitTaskFrontmatter),
		codingcontext.WithLogger(logger),
	)

	result, err := cc.Run(ctx, args[0])
	if err != nil {
		logger.Error("Error", "error", err)
		flag.Usage()
		os.Exit(1)
	}

	// Output task frontmatter if requested
	if emitTaskFrontmatter && result.Task.FrontMatter != nil {
		fmt.Println("---")
		if err := yaml.NewEncoder(os.Stdout).Encode(result.Task.FrontMatter); err != nil {
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
