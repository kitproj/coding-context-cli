package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

//go:embed bootstrap
var bootstrap string

var (
	workDir string
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.StringVar(&workDir, "C", ".", "Change to directory before doing anything.")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:\n")
		fmt.Fprintf(w, "  coding-context <command> [options] [arguments]\n\n")
		fmt.Fprintln(w, "Commands:")
		fmt.Fprintln(w, "  import         Import rules from all known agents to default agent")
		fmt.Fprintln(w, "  export <agent> Export rules from default agent to specified agent")
		fmt.Fprintln(w, "  bootstrap      Run bootstrap scripts")
		fmt.Fprintf(w, "  prompt <name>  Find and print a task prompt\n\n")
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
	agentRules, err := initAgentRules()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize agent rules: %v\n", err)
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "import":
		if err := runImport(ctx, agentRules, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "export":
		if err := runExport(ctx, agentRules, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "bootstrap":
		if err := runBootstrap(ctx, agentRules, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "prompt":
		if err := runPrompt(ctx, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}
