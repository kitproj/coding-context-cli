package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var workDir string
	var resume bool
	var emitTaskFrontmatter bool
	params := make(codingcontext.ParamMap)
	includes := make(codingcontext.SelectorMap)
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
		fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), "  coding-context [options] <task-name>")
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(flag.CommandLine.Output(), "Error: invalid usage")
		flag.Usage()
		os.Exit(1)
	}

	taskName := args[0]

	// Create context using library
	cc := codingcontext.New(
		codingcontext.WithWorkDir(workDir),
		codingcontext.WithResume(resume),
		codingcontext.WithParams(params),
		codingcontext.WithIncludes(includes),
		codingcontext.WithRemotePaths(remotePaths),
		codingcontext.WithEmitTaskFrontmatter(emitTaskFrontmatter),
		codingcontext.WithOutput(os.Stdout),
		codingcontext.WithLogOutput(flag.CommandLine.Output()),
	)

	if err := cc.Run(ctx, taskName); err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}
