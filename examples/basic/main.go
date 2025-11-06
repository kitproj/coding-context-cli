package main

import (
	"context"
	"fmt"
	"os"

	ctxlib "github.com/kitproj/coding-context-cli/context"
)

func main() {
	// Create parameters for substitution
	params := make(ctxlib.ParamMap)
	params["component"] = "authentication"
	params["issue"] = "password reset bug"

	// Create selectors for filtering rules
	selectors := make(ctxlib.SelectorMap)
	selectors["language"] = "go"

	// Configure the assembler
	config := ctxlib.Config{
		WorkDir:   ".",
		TaskName:  "fix-bug",
		Params:    params,
		Selectors: selectors,
	}

	// Create the assembler
	assembler := ctxlib.NewAssembler(config)

	// Assemble the context
	ctx := context.Background()
	task, err := assembler.Assemble(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Print the task content
	fmt.Println(task.Content)
}
