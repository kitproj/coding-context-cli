package main

import (
	"context"
	"fmt"
	"os"

	ctxlib "github.com/kitproj/coding-context-cli/context"
)

// LoggingVisitor is a custom visitor that logs rule information
type LoggingVisitor struct {
	RuleCount int
}

func (v *LoggingVisitor) VisitRule(ctx context.Context, rule *ctxlib.Rule) error {
	v.RuleCount++

	// Log detailed information about each rule
	fmt.Fprintf(os.Stderr, "\n=== Rule #%d ===\n", v.RuleCount)
	fmt.Fprintf(os.Stderr, "Path: %s\n", rule.Path)
	fmt.Fprintf(os.Stderr, "Tokens: %d\n", rule.Tokens)

	// Log frontmatter
	if len(rule.Frontmatter) > 0 {
		fmt.Fprintf(os.Stderr, "Frontmatter:\n")
		for key, value := range rule.Frontmatter {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", key, value)
		}
	}

	// Write the content to stdout (maintaining default behavior)
	fmt.Println(rule.Content)

	return nil
}

func main() {
	// Create a custom logging visitor
	visitor := &LoggingVisitor{}

	// Configure the assembler with the custom visitor
	config := ctxlib.Config{
		WorkDir:  ".",
		TaskName: "fix-bug",
		Visitor:  visitor,
	}

	// Assemble the context
	assembler := ctxlib.NewAssembler(config)
	ctx := context.Background()
	if err := assembler.Assemble(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\n=== Summary ===\n")
	fmt.Fprintf(os.Stderr, "Total rules processed: %d\n", visitor.RuleCount)
}
