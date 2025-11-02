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
		fmt.Fprintf(w, "  coding-context <agent_name> <task_name> [-p key=value] [-s key=value] [-S key=value]\n")
		fmt.Fprintf(w, "  coding-context <command> [options] [arguments]\n\n")
		fmt.Fprintln(w, "Unified Command (imports, exports, bootstraps, and generates prompt):")
		fmt.Fprintln(w, "  <agent_name> <task_name>")
		fmt.Fprintln(w, "                      Run full workflow for specified agent and task")
		fmt.Fprintln(w, "                      Flags: -p key=value, -s key=value, -S key=value")
		fmt.Fprintln(w, "Individual Commands:")
		fmt.Fprintln(w, "  import              Import rules from all known agents to default agent")
		fmt.Fprintln(w, "  export <agent> [-s key=value] [-S key=value]")
		fmt.Fprintln(w, "                      Export rules from default agent to specified agent")
		fmt.Fprintln(w, "  bootstrap           Run bootstrap scripts")
		fmt.Fprintf(w, "  prompt <name> [-p key=value]\n")
		fmt.Fprintf(w, "                      Find and print a task prompt\n\n")
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

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "import":
		// Initialize import rules
		importRules, err := initImportRules()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to initialize import rules: %v\n", err)
			os.Exit(1)
		}
		if err := runImport(ctx, importRules, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "export":
		// Initialize export rules
		exportRules, err := initExportRules()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to initialize export rules: %v\n", err)
			os.Exit(1)
		}
		if err := runExport(ctx, exportRules, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "bootstrap":
		// Initialize export rules for bootstrap (reads from Default agent)
		exportRules, err := initExportRules()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to initialize export rules: %v\n", err)
			os.Exit(1)
		}
		if err := runBootstrap(ctx, exportRules, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "prompt":
		if err := runPrompt(ctx, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		// Try to interpret as unified command: agent_name task_name ...
		// If first arg looks like an agent name, treat as unified command
		if err := runUnified(ctx, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
