package context

import (
	"context"
	"fmt"
	"io"
)

// Rule represents a rule file that has been selected for inclusion
type Rule struct {
	// Path is the absolute path to the rule file
	Path string
	// Content is the parsed content of the rule file (without frontmatter)
	Content string
	// Frontmatter contains the YAML frontmatter metadata
	Frontmatter map[string]string
	// Tokens is the estimated token count for this rule
	Tokens int
}

// RuleVisitor defines the interface for visiting rules as they are selected
type RuleVisitor interface {
	// VisitRule is called for each rule that matches the selection criteria
	// It receives the context and the rule information
	// Returning an error will stop the assembly process
	VisitRule(ctx context.Context, rule *Rule) error
}

// DefaultRuleVisitor is the default implementation that writes rules to stdout
type DefaultRuleVisitor struct {
	stdout io.Writer
	stderr io.Writer
}

// VisitRule writes the rule content to stdout and logs progress to stderr
func (v *DefaultRuleVisitor) VisitRule(ctx context.Context, rule *Rule) error {
	fmt.Fprintf(v.stderr, "⪢ Including rule file: %s (~%d tokens)\n", rule.Path, rule.Tokens)
	fmt.Fprintln(v.stdout, rule.Content)
	return nil
}
