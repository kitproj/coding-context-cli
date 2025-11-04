package main

import (
	ctx "context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kitproj/coding-context-cli/context"
)

var (
	workDir  string
	params   = make(context.ParamMap)
	includes = make(context.SelectorMap)
)

func main() {
	c, cancel := signal.NotifyContext(ctx.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")
	flag.Var(&params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  coding-context [options] <task-name>")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(c, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(c ctx.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid usage")
	}

	taskName := args[0]

	// Create assembler with configuration
	assembler := context.NewAssembler(context.Config{
		WorkDir:   workDir,
		TaskName:  taskName,
		Params:    params,
		Selectors: includes,
	})

	return assembler.Assemble(c)
}
