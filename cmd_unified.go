package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// runUnified runs the unified command that combines import, export, bootstrap, and prompt
func runUnified(ctx context.Context, args []string) error {
	// Define flags for unified command
	var params paramMap
	var includes selectorMap
	var excludes selectorMap
	unifiedFlags := flag.NewFlagSet("unified", flag.ExitOnError)
	unifiedFlags.Var(&params, "p", "Template parameter (key=value)")
	unifiedFlags.Var(&includes, "s", "Include rules with matching frontmatter (key=value)")
	unifiedFlags.Var(&excludes, "S", "Exclude rules with matching frontmatter (key=value)")

	if err := unifiedFlags.Parse(args); err != nil {
		return err
	}

	unifiedArgs := unifiedFlags.Args()
	if len(unifiedArgs) < 2 {
		return fmt.Errorf("usage: coding-context <agent_name> <task_name> [-p key=value] [-s key=value] [-S key=value]")
	}

	agentName := Agent(unifiedArgs[0])
	taskName := unifiedArgs[1]

	// Step 1: Import from all tools to the ".agents" structure
	fmt.Fprintf(os.Stderr, "Step 1: Importing from all agents...\n")
	importRules, err := initImportRules()
	if err != nil {
		return fmt.Errorf("failed to initialize import rules: %w", err)
	}
	if err := runImport(ctx, importRules, []string{}); err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Step 2: Export to requested agent structure
	fmt.Fprintf(os.Stderr, "\nStep 2: Exporting to %s...\n", agentName)
	exportRules, err := initExportRules()
	if err != nil {
		return fmt.Errorf("failed to initialize export rules: %w", err)
	}

	// Build export args with selectors
	exportArgs := []string{string(agentName)}
	for k, v := range includes {
		exportArgs = append(exportArgs, "-s", fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range excludes {
		exportArgs = append(exportArgs, "-S", fmt.Sprintf("%s=%s", k, v))
	}

	if err := runExport(ctx, exportRules, exportArgs); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Step 3: Bootstrap
	fmt.Fprintf(os.Stderr, "\nStep 3: Running bootstrap...\n")
	if err := runBootstrap(ctx, exportRules, []string{}); err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}

	// Step 4: Print the task prompt to stdout
	fmt.Fprintf(os.Stderr, "\nStep 4: Generating prompt for task '%s'...\n", taskName)

	// Build prompt args with params
	promptArgs := []string{taskName}
	for k, v := range params {
		promptArgs = append(promptArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}

	if err := runPrompt(ctx, promptArgs); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	return nil
}
