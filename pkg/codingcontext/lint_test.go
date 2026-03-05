package codingcontext

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lintTestDir creates a temp dir registered for cleanup and returns its path.
func lintTestDir(t *testing.T) string {
	t.Helper()

	return t.TempDir()
}

// newLintContext creates a Context configured to use dir as the sole search path.
func newLintContext(dir string) *Context {
	return New(WithSearchPaths(dir))
}

// hasLoadedFile reports whether result contains a loaded file with the given path suffix and kind.
func hasLoadedFile(result *LintResult, pathSuffix string, kind LoadedFileKind) bool {
	for _, f := range result.LoadedFiles {
		if strings.HasSuffix(f.Path, pathSuffix) && f.Kind == kind {
			return true
		}
	}

	return false
}

// hasLintError reports whether result contains a LintError with the given kind and message substring.
func hasLintError(result *LintResult, kind LintErrorKind, msgSubstr string) bool {
	for _, e := range result.Errors {
		if e.Kind == kind && strings.Contains(e.Message, msgSubstr) {
			return true
		}
	}

	return false
}

func TestLint_BasicTaskAndRule(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	createTask(t, dir, "deploy", "", "Deploy the application.")
	createRule(t, dir, ".agents/rules/base.md", "", "Always write tests.")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "deploy")
	if err != nil {
		t.Fatalf("Lint() returned error: %v", err)
	}

	if !hasLoadedFile(result, "deploy.md", LoadedFileKindTask) {
		t.Errorf("expected task file in LoadedFiles, got %+v", result.LoadedFiles)
	}

	if !hasLoadedFile(result, "base.md", LoadedFileKindRule) {
		t.Errorf("expected rule file in LoadedFiles, got %+v", result.LoadedFiles)
	}

	if result.Prompt == "" {
		t.Error("expected non-empty Prompt")
	}
}

func TestLint_TaskNotFound_FatalError(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	cc := newLintContext(dir)

	_, err := cc.Lint(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing task, got nil")
	}

	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestLint_BootstrapFileSkipped(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	createTask(t, dir, "mytask", "", "Do something.")
	createRule(t, dir, ".agents/rules/myrule.md", "", "Some rule.")
	createBootstrapScript(t, dir, ".agents/rules/myrule.md", "#!/bin/sh\nexit 1\n")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "mytask")
	if err != nil {
		t.Fatalf("Lint() returned error (bootstrap should be skipped): %v", err)
	}

	if !hasLoadedFile(result, "myrule-bootstrap", LoadedFileKindBootstrap) {
		t.Errorf("expected bootstrap file in LoadedFiles, got %+v", result.LoadedFiles)
	}
}

func TestLint_CommandExpansionSkipped(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	// Task content with a !`cmd` expansion that would fail if executed
	createTask(t, dir, "cmdtask", "", "Result: !`exit 1`")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "cmdtask")
	if err != nil {
		t.Fatalf("Lint() returned error: %v", err)
	}
	// The literal !`exit 1` should be preserved in the prompt
	if !strings.Contains(result.Prompt, "!`exit 1`") {
		t.Errorf("expected literal command in prompt, got: %s", result.Prompt)
	}
	// No lint error should be generated for skipped commands
	for _, e := range result.Errors {
		if e.Kind == LintErrorKindMissingCommand {
			t.Errorf("unexpected missing-command error: %+v", e)
		}
	}
}

func TestLint_PathRefTracked(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	// Create a file to reference
	refFile := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(refFile, []byte("some data"), 0o600); err != nil {
		t.Fatalf("failed to create ref file: %v", err)
	}

	createTask(t, dir, "pathtask", "", "Content: @"+refFile)

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "pathtask")
	if err != nil {
		t.Fatalf("Lint() returned error: %v", err)
	}

	if !hasLoadedFile(result, "data.txt", LoadedFileKindPathRef) {
		t.Errorf("expected path-ref in LoadedFiles, got %+v", result.LoadedFiles)
	}
	// File content should be included in prompt
	if !strings.Contains(result.Prompt, "some data") {
		t.Errorf("expected file content in prompt, got: %s", result.Prompt)
	}
}

func TestLint_MissingCommand_NonFatal(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	createTask(t, dir, "task1", "", "/missingcmd arg1\nSome text after.")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "task1")
	if err != nil {
		t.Fatalf("Lint() returned error: %v", err)
	}

	if !hasLintError(result, LintErrorKindMissingCommand, "missingcmd") {
		t.Errorf("expected missing-command error, got errors: %+v", result.Errors)
	}
	// Assembly should have continued — prompt should still have the text after
	if !strings.Contains(result.Prompt, "Some text after.") {
		t.Errorf("expected remaining task content in prompt, got: %s", result.Prompt)
	}
}

func TestLint_SkillValidation_NonFatal(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	createTask(t, dir, "task1", "", "Do stuff.")
	// Skill with missing name — invalid
	createSkill(t, dir, ".agents/skills/badskill", "---\ndescription: Some description\n---\nSkill content.")
	// Valid skill
	createSkill(t, dir, ".agents/skills/goodskill", "---\nname: good-skill\ndescription: A good skill\n---\nGood content.")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "task1")
	if err != nil {
		t.Fatalf("Lint() returned error: %v", err)
	}

	if !hasLintError(result, LintErrorKindSkillValidation, "skill missing required 'name'") {
		t.Errorf("expected skill-validation error, got errors: %+v", result.Errors)
	}
	// Good skill should still appear in result
	found := false

	for _, s := range result.Skills.Skills {
		if s.Name == "good-skill" {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected good-skill in Skills, got: %+v", result.Skills.Skills)
	}
}

func TestLint_SelectorNoMatch(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	createTask(t, dir, "task1", "", "Do stuff.")
	createRule(t, dir, ".agents/rules/rule1.md", "name: rule1\nlanguage: go\n", "Go rule.")

	cc := New(
		WithSearchPaths(dir),
		// Selector value that no file has in its frontmatter
		WithSelectors(map[string]map[string]bool{
			"environment": {"production": true},
		}),
	)

	result, err := cc.Lint(context.Background(), "task1")
	if err != nil {
		t.Fatalf("Lint() returned error: %v", err)
	}

	if !hasLintError(result, LintErrorKindSelectorNoMatch, "environment=production") {
		t.Errorf("expected selector-no-match error, got errors: %+v", result.Errors)
	}
}

func TestLint_WithLintOption(t *testing.T) {
	t.Parallel()

	cc := New(WithLint(true))
	if !cc.lintMode {
		t.Error("expected lintMode to be true after WithLint(true)")
	}
}

func TestLint_Command_Tracked(t *testing.T) {
	t.Parallel()
	dir := lintTestDir(t)

	createTask(t, dir, "task1", "", "/mycmd\nText after.")
	createCommand(t, dir, "mycmd", "", "Command content.")

	cc := newLintContext(dir)

	result, err := cc.Lint(context.Background(), "task1")
	if err != nil {
		t.Fatalf("Lint() returned error: %v", err)
	}

	if !hasLoadedFile(result, "mycmd.md", LoadedFileKindCommand) {
		t.Errorf("expected command file in LoadedFiles, got %+v", result.LoadedFiles)
	}
}
