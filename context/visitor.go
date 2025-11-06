package context

import (
"context"
"fmt"
"io"
"log/slog"
)

// RuleVisitor defines the interface for visiting rules as they are selected
type RuleVisitor interface {
// VisitRule is called for each rule that matches the selection criteria
// It receives the context and the rule document
// Returning an error will stop the assembly process
VisitRule(ctx context.Context, rule *Document) error
}

// DefaultRuleVisitor is the default implementation that writes rules to stdout
type DefaultRuleVisitor struct {
stdout io.Writer
logger *slog.Logger
}

// VisitRule writes the rule content to stdout and logs progress
func (v *DefaultRuleVisitor) VisitRule(ctx context.Context, rule *Document) error {
if v.logger == nil {
v.logger = slog.Default()
}
v.logger.Info("including rule file", "path", rule.Path, "tokens", rule.Tokens)
fmt.Fprintln(v.stdout, rule.Content)
return nil
}
