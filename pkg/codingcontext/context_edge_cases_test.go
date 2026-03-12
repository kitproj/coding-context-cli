package codingcontext

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMdcExtensionRuleDiscovery verifies that .mdc files are treated as rule files.
// The walk function explicitly allows .mdc extensions alongside .md; if that check is
// ever dropped, this test will catch the regression.
func TestMdcExtensionRuleDiscovery(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "Task content")

	// Place a rule as a .mdc file
	ruleDir := filepath.Join(dir, ".agents", "rules")
	if err := os.MkdirAll(ruleDir, 0o750); err != nil {
		t.Fatalf("failed to create rule dir: %v", err)
	}

	rulePath := filepath.Join(ruleDir, "style.mdc")
	if err := os.WriteFile(rulePath, []byte("MDC rule content"), 0o600); err != nil {
		t.Fatalf("failed to write .mdc file: %v", err)
	}

	c := New(WithSearchPaths(dir))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Rules) != 1 {
		t.Fatalf("expected 1 rule from .mdc file, got %d", len(result.Rules))
	}

	if !strings.Contains(result.Rules[0].Content, "MDC rule content") {
		t.Errorf("expected .mdc rule content, got %q", result.Rules[0].Content)
	}
}

// TestSkillValidation_NameExactlyAtLimit verifies that a skill name of exactly 64
// characters is accepted (boundary should be inclusive).
func TestSkillValidation_NameExactlyAtLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "Task")

	name64 := strings.Repeat("x", 64)
	skillContent := "---\nname: " + name64 + "\ndescription: Valid description.\n---\n"
	createSkill(t, dir, ".agents/skills/myskill", skillContent)

	c := New(WithSearchPaths(dir))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error for 64-char skill name: %v", err)
	}

	if len(result.Skills.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
	}

	if result.Skills.Skills[0].Name != name64 {
		t.Errorf("expected skill name %q, got %q", name64, result.Skills.Skills[0].Name)
	}
}

// TestSkillValidation_NameOverLimit verifies that a skill name of 65 characters is
// rejected with ErrSkillNameLength.
func TestSkillValidation_NameOverLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "Task")

	name65 := strings.Repeat("x", 65)
	skillContent := "---\nname: " + name65 + "\ndescription: Valid description.\n---\n"
	createSkill(t, dir, ".agents/skills/myskill", skillContent)

	c := New(WithSearchPaths(dir))

	_, err := c.Run(context.Background(), "task")
	if err == nil {
		t.Fatal("expected Run() to fail for 65-char skill name")
	}

	if !errors.Is(err, ErrSkillNameLength) {
		t.Errorf("expected ErrSkillNameLength, got: %v", err)
	}
}

// TestSkillValidation_DescExactlyAtLimit verifies that a skill description of exactly
// 1024 characters is accepted.
func TestSkillValidation_DescExactlyAtLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "Task")

	desc1024 := strings.Repeat("d", 1024)
	skillContent := "---\nname: valid-skill\ndescription: " + desc1024 + "\n---\n"
	createSkill(t, dir, ".agents/skills/myskill", skillContent)

	c := New(WithSearchPaths(dir))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error for 1024-char description: %v", err)
	}

	if len(result.Skills.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
	}
}

// TestSkillValidation_DescOverLimit verifies that a skill description of 1025
// characters is rejected with ErrSkillDescriptionLength.
func TestSkillValidation_DescOverLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "Task")

	desc1025 := strings.Repeat("d", 1025)
	skillContent := "---\nname: valid-skill\ndescription: " + desc1025 + "\n---\n"
	createSkill(t, dir, ".agents/skills/myskill", skillContent)

	c := New(WithSearchPaths(dir))

	_, err := c.Run(context.Background(), "task")
	if err == nil {
		t.Fatal("expected Run() to fail for 1025-char description")
	}

	if !errors.Is(err, ErrSkillDescriptionLength) {
		t.Errorf("expected ErrSkillDescriptionLength, got: %v", err)
	}
}

// TestBootstrapFailurePropagates verifies that when the bootstrap script runner
// returns an error, Run() propagates it as a failure.  The rule is appended to
// cc.rules before bootstrap runs, so the error must surface via Run() rather than
// being silently ignored.
func TestBootstrapFailurePropagates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "Task")
	createRule(t, dir, ".agents/rules/rule.md", "", "Rule content")
	createBootstrapScript(t, dir, ".agents/rules/rule.md", "#!/bin/sh\nexit 1")

	c := New(WithSearchPaths(dir))
	// Replace the default runner with one that always fails
	c.cmdRunner = func(_ *exec.Cmd) error {
		return errors.New("simulated bootstrap failure") //nolint:err113
	}

	_, err := c.Run(context.Background(), "task")
	if err == nil {
		t.Fatal("expected Run() to fail when bootstrap script errors")
	}

	if !strings.Contains(err.Error(), "bootstrap") {
		t.Errorf("expected bootstrap-related error message, got: %v", err)
	}
}

// TestSkillSelectorFiltering verifies that skills whose frontmatter contains a
// selector key that does NOT match any active selector value are excluded.
// This exercises the MatchesIncludes path inside loadSkillEntry.
func TestSkillSelectorFiltering(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Task adds env: development to the includes via its frontmatter selectors
	createTask(t, dir, "task", "selectors:\n  env: development", "Task content")

	// This skill declares env: production — it should be excluded
	createSkill(t, dir, ".agents/skills/prod-skill",
		"---\nname: prod-skill\ndescription: A production skill.\nenv: production\n---\n")

	// This skill has no env selector — it should be included
	createSkill(t, dir, ".agents/skills/generic-skill",
		"---\nname: generic-skill\ndescription: A generic skill.\n---\n")

	c := New(WithSearchPaths(dir))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	skillNames := make([]string, 0, len(result.Skills.Skills))
	for _, s := range result.Skills.Skills {
		skillNames = append(skillNames, s.Name)
	}

	for _, name := range skillNames {
		if name == "prod-skill" {
			t.Error("prod-skill (env: production) should be excluded when env: development is active")
		}
	}

	found := false

	for _, name := range skillNames {
		if name == "generic-skill" {
			found = true
		}
	}

	if !found {
		t.Errorf("generic-skill (no env selector) should be included; got skills: %v", skillNames)
	}
}

// TestMergeSelectorsIntegerYAMLValue verifies that task frontmatter selectors whose
// YAML values parse as integers (not strings) still match rule frontmatter correctly.
// The mergeSelectors function uses fmt.Sprint, and MatchesIncludes does the same;
// if either were changed to a type assertion, this test would panic or fail.
func TestMergeSelectorsIntegerYAMLValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// YAML parses bare integers without quotes — the selector value is an int
	createTask(t, dir, "task", "selectors:\n  priority: 1", "Task content")
	createRule(t, dir, ".agents/rules/p1.md", "priority: 1", "Priority 1 rule")
	createRule(t, dir, ".agents/rules/p2.md", "priority: 2", "Priority 2 rule")
	createRule(t, dir, ".agents/rules/any.md", "", "No priority rule")

	c := New(WithSearchPaths(dir))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Priority-1 rule and the no-priority rule should be included; priority-2 excluded
	var p1Found, p2Found bool

	for _, rule := range result.Rules {
		if strings.Contains(rule.Content, "Priority 1") {
			p1Found = true
		}

		if strings.Contains(rule.Content, "Priority 2") {
			p2Found = true
		}
	}

	if !p1Found {
		t.Error("expected Priority 1 rule to be included (priority: 1 matches)")
	}

	if p2Found {
		t.Error("expected Priority 2 rule to be excluded (priority: 2 does not match priority: 1)")
	}
}

// TestMergeSelectorsYAMLBoolValue verifies selector matching when YAML parses a
// selector value as a boolean (true/false rather than a quoted string).
func TestMergeSelectorsYAMLBoolValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// YAML parses bare `true` as a bool, not a string
	createTask(t, dir, "task", "selectors:\n  experimental: true", "Task content")
	createRule(t, dir, ".agents/rules/exp.md", "experimental: true", "Experimental rule")
	createRule(t, dir, ".agents/rules/stable.md", "experimental: false", "Stable rule")
	createRule(t, dir, ".agents/rules/any.md", "", "No experimental flag")

	c := New(WithSearchPaths(dir))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var expFound, stableFound bool

	for _, rule := range result.Rules {
		if strings.Contains(rule.Content, "Experimental rule") {
			expFound = true
		}

		if strings.Contains(rule.Content, "Stable rule") {
			stableFound = true
		}
	}

	if !expFound {
		t.Error("expected Experimental rule to be included (experimental: true matches)")
	}

	if stableFound {
		t.Error("expected Stable rule to be excluded (experimental: false does not match true)")
	}
}

// TestEmptyTaskBody verifies that a task file containing only frontmatter and an
// empty body is accepted without error and produces an empty content string.
func TestEmptyTaskBody(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "empty-body", "priority: high", "")

	c := New(WithSearchPaths(dir), WithBootstrap(false))

	result, err := c.Run(context.Background(), "empty-body")
	if err != nil {
		t.Fatalf("Run() error for empty task body: %v", err)
	}

	if strings.TrimSpace(result.Task.Content) != "" {
		t.Errorf("expected empty task content, got %q", result.Task.Content)
	}
}

// TestCommandPrecedenceAcrossSearchPaths verifies that when the same command name
// exists in two search paths, the first search path wins.  This documents the
// "first match wins" contract of findCommand.
func TestCommandPrecedenceAcrossSearchPaths(t *testing.T) {
	t.Parallel()

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	createTask(t, dir1, "task", "", "/deploy")
	createCommand(t, dir1, "deploy", "", "Deploy from path one")
	createCommand(t, dir2, "deploy", "", "Deploy from path two")

	c := New(WithSearchPaths(dir1, dir2), WithBootstrap(false))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Task.Content, "Deploy from path one") {
		t.Errorf("expected first search path command to win, got %q", result.Task.Content)
	}

	if strings.Contains(result.Task.Content, "Deploy from path two") {
		t.Error("second search path command must not appear when first path has same command")
	}
}

// TestEmptyTaskName verifies that passing an empty task name to Run() results in a
// task-not-found error rather than a panic or unexpected success.
func TestEmptyTaskName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := New(WithSearchPaths(dir), WithBootstrap(false))

	_, err := c.Run(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty task name")
	}

	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound for empty task name, got: %v", err)
	}
}

// TestUserPromptSeparatorInContent verifies that when a user prompt is appended,
// the resulting content contains the task text, the --- separator, and the prompt
// text — confirming the documented append format.
func TestUserPromptSeparatorInContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "Task body")

	c := New(WithSearchPaths(dir), WithBootstrap(false), WithUserPrompt("User text"))

	result, err := c.Run(context.Background(), "task")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	content := result.Task.Content

	taskIdx := strings.Index(content, "Task body")
	userIdx := strings.Index(content, "User text")

	if taskIdx < 0 {
		t.Error("task body missing from content")
	}

	if userIdx < 0 {
		t.Error("user prompt missing from content")
	}

	if taskIdx >= 0 && userIdx >= 0 && taskIdx >= userIdx {
		t.Error("expected task body to appear before user prompt")
	}

	// The separator "---" must appear between them
	between := content[taskIdx:userIdx]
	if !strings.Contains(between, "---") {
		t.Errorf("expected '---' separator between task and user prompt; between section: %q", between)
	}
}
