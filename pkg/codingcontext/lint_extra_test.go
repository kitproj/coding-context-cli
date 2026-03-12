package codingcontext

import (
	"context"
	"strings"
	"testing"
)

// TestLint_FrontmatterListValue exercises the []any branch in recordFrontmatterValues.
// When a rule has a list-valued frontmatter field (e.g. "languages: [go, python]"),
// each item is tracked individually in allFrontmatterValues. This is required for
// selector coverage validation to report list-keyed selectors correctly.
func TestLint_FrontmatterListValue(t *testing.T) {
	t.Parallel()

	dir := lintTestDir(t)

	createTask(t, dir, "task", "", "Task content.")
	// Rule with a list-valued frontmatter field
	createRule(t, dir, ".agents/rules/multi-lang.md", "languages:\n  - go\n  - python\n", "Multi-language rule.")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "task")
	if err != nil {
		t.Fatalf("Lint() error: %v", err)
	}

	// The rule should be included (no selector filter active)
	if !hasLoadedFile(result, "multi-lang.md", LoadedFileKindRule) {
		t.Errorf("expected multi-lang rule in LoadedFiles, got %+v", result.LoadedFiles)
	}

	// allFrontmatterValues should contain "go" and "python" under "languages"
	// We verify indirectly: add a selector that matches and one that doesn't match,
	// and check that the selector-no-match logic sees list items properly.
	seen := cc.lintCollector.allFrontmatterValues["languages"]
	if !seen["go"] || !seen["python"] {
		t.Errorf("expected 'go' and 'python' in allFrontmatterValues[\"languages\"], got %v", seen)
	}
}

// TestLint_FrontmatterBootstrapSkipped exercises the frontmatterBootstrap != ""
// branch in recordLintBootstrap. When a rule has an inline bootstrap script in
// its frontmatter, lint mode must log and skip — not execute — it, and must NOT
// record it in LoadedFiles (it has no companion file to stat).
func TestLint_FrontmatterBootstrapSkipped(t *testing.T) {
	t.Parallel()

	dir := lintTestDir(t)

	createTask(t, dir, "task", "", "Task content.")
	// Rule with frontmatter-based bootstrap (inline script)
	createRule(t, dir, ".agents/rules/fmrule.md", "bootstrap: |\n  #!/bin/sh\n  exit 1\n", "Rule with inline bootstrap.")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "task")
	if err != nil {
		t.Fatalf("Lint() should not fail when frontmatter bootstrap is skipped: %v", err)
	}

	// The rule itself should be loaded
	if !hasLoadedFile(result, "fmrule.md", LoadedFileKindRule) {
		t.Errorf("expected fmrule.md in LoadedFiles, got %+v", result.LoadedFiles)
	}

	// No bootstrap file entry should appear (frontmatter bootstrap has no companion file)
	for _, f := range result.LoadedFiles {
		if f.Kind == LoadedFileKindBootstrap && strings.HasSuffix(f.Path, "fmrule-bootstrap") {
			t.Errorf("unexpected bootstrap file entry for frontmatter bootstrap: %+v", f)
		}
	}
}

// TestLint_ParseErrorInRuleFile exercises the parse-error branch in makeMarkdownWalkFunc.
// When a rule file has invalid YAML frontmatter, lint mode records a parse error
// (non-fatal) and continues assembly instead of aborting.
func TestLint_ParseErrorInRuleFile(t *testing.T) {
	t.Parallel()

	dir := lintTestDir(t)

	createTask(t, dir, "task", "", "Task content.")
	// Rule with syntactically invalid YAML (duplicate mapping key causes error)
	createRule(t, dir, ".agents/rules/badrule.md", "invalid: yaml: : syntax\n", "Bad rule content.")
	// A valid rule that should still be loaded despite the bad one
	createRule(t, dir, ".agents/rules/goodrule.md", "", "Good rule content.")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "task")
	if err != nil {
		t.Fatalf("Lint() should not fatal on parse error in rule: %v", err)
	}

	// A parse error should be recorded
	if !hasLintError(result, LintErrorKindParse, "") {
		t.Errorf("expected parse error for bad rule YAML, got errors: %+v", result.Errors)
	}

	// The good rule should still be loaded
	if !hasLoadedFile(result, "goodrule.md", LoadedFileKindRule) {
		t.Errorf("expected goodrule.md in LoadedFiles despite bad rule, got %+v", result.LoadedFiles)
	}
}
