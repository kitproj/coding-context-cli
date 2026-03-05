package codingcontext

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ── fixture helpers ──────────────────────────────────────────────────────────

// createNamespaceTask creates a task file under .agents/namespaces/<ns>/tasks/.
func createNamespaceTask(t *testing.T, dir, namespace, name, content string) {
	t.Helper()

	taskDir := filepath.Join(dir, ".agents", "namespaces", namespace, "tasks")
	if err := os.MkdirAll(taskDir, 0o750); err != nil {
		t.Fatalf("failed to create namespace task dir: %v", err)
	}

	fileContent := content

	if err := os.WriteFile(filepath.Join(taskDir, name+".md"), []byte(fileContent), 0o600); err != nil {
		t.Fatalf("failed to write namespace task file: %v", err)
	}
}

// createNamespaceRule creates a rule file under .agents/namespaces/<ns>/rules/.
func createNamespaceRule(t *testing.T, dir, namespace, name, content string) {
	t.Helper()

	relPath := filepath.Join(".agents", "namespaces", namespace, "rules", name+".md")
	createRule(t, dir, relPath, "", content)
}

// createNamespaceCommand creates a command file under .agents/namespaces/<ns>/commands/.
func createNamespaceCommand(t *testing.T, dir, namespace, name, frontmatter, content string) {
	t.Helper()
	relPath := filepath.Join(".agents", "namespaces", namespace, "commands", name+".md")
	createRule(t, dir, relPath, frontmatter, content)
}

// createNamespaceSkill creates a skill under .agents/namespaces/<ns>/skills/<subdir>/SKILL.md.
func createNamespaceSkill(t *testing.T, dir, namespace, subdir, content string) {
	t.Helper()

	skillDir := filepath.Join(dir, ".agents", "namespaces", namespace, "skills", subdir)
	if err := os.MkdirAll(skillDir, 0o750); err != nil {
		t.Fatalf("failed to create namespace skill dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write namespace SKILL.md: %v", err)
	}
}

// newRunContext creates a Context scoped to a single temp directory, with
// bootstrap disabled (to avoid accidentally running scripts in unit tests).
func newRunContext(dir string) *Context {
	return New(WithSearchPaths(dir), WithBootstrap(false))
}

// newFullContext creates a Context scoped to a single temp directory with
// bootstrap enabled (for tests that need rule/skill discovery).
func newFullContext(dir string) *Context {
	return New(WithSearchPaths(dir))
}

// ── parseNamespacedTaskName ──────────────────────────────────────────────────

func TestParseNamespacedTaskName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		wantNS    string
		wantBase  string
		wantErr   bool
		errSubstr string
	}{
		{input: "fix-bug", wantNS: "", wantBase: "fix-bug"},
		{input: "myteam/fix-bug", wantNS: "myteam", wantBase: "fix-bug"},
		{input: "team-a/deploy", wantNS: "team-a", wantBase: "deploy"},
		{input: "a/b/c", wantErr: true, errSubstr: "one level"},
		{input: "/task", wantErr: true, errSubstr: "namespace must not be empty"},
		{input: "ns/", wantErr: true, errSubstr: "task base name must not be empty"},
		{input: "a/b/c/d", wantErr: true, errSubstr: "one level"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			ns, base, err := parseNamespacedTaskName(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.input)
				}

				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error to contain %q, got %q", tt.errSubstr, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}

			if ns != tt.wantNS {
				t.Errorf("namespace: got %q, want %q", ns, tt.wantNS)
			}

			if base != tt.wantBase {
				t.Errorf("baseName: got %q, want %q", base, tt.wantBase)
			}
		})
	}
}

// ── namespace-aware path functions ───────────────────────────────────────────

func TestNamespacedTaskSearchPaths(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	ns := testNamespace

	paths := namespacedTaskSearchPaths(dir, ns)

	nsPath := filepath.Join(dir, ".agents", "namespaces", ns, "tasks")
	globalPath := filepath.Join(dir, ".agents", "tasks")

	if !slices.Contains(paths, nsPath) {
		t.Errorf("expected namespace task path %q in result", nsPath)
	}

	if !slices.Contains(paths, globalPath) {
		t.Errorf("expected global task path %q in result", globalPath)
	}
	// namespace path must come before global path
	nsIdx := slices.Index(paths, nsPath)

	globalIdx := slices.Index(paths, globalPath)
	if nsIdx >= globalIdx {
		t.Errorf("expected namespace path (idx %d) before global path (idx %d)", nsIdx, globalIdx)
	}
}

func TestNamespacedTaskSearchPaths_NoNamespace(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	paths := namespacedTaskSearchPaths(dir, "")

	// Should equal plain taskSearchPaths output
	plain := taskSearchPaths(dir)
	for _, p := range plain {
		if !slices.Contains(paths, p) {
			t.Errorf("expected global task path %q in no-namespace result", p)
		}
	}

	// Must not include any namespaces directory
	for _, p := range paths {
		if strings.Contains(p, "namespaces") {
			t.Errorf("no-namespace call should not include 'namespaces' path, got %q", p)
		}
	}
}

func TestNamespacedRuleSearchPaths(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	ns := testNamespace

	paths := namespacedRuleSearchPaths(dir, ns)

	nsPath := filepath.Join(dir, ".agents", "namespaces", ns, "rules")
	globalPath := filepath.Join(dir, ".agents", "rules")

	if !slices.Contains(paths, nsPath) {
		t.Errorf("expected namespace rule path %q", nsPath)
	}

	if !slices.Contains(paths, globalPath) {
		t.Errorf("expected global rule path %q", globalPath)
	}
	// namespace path must come first
	if slices.Index(paths, nsPath) >= slices.Index(paths, globalPath) {
		t.Error("expected namespace rule path to precede global rule path")
	}
}

func TestNamespacedCommandSearchPaths(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	ns := testNamespace

	paths := namespacedCommandSearchPaths(dir, ns)

	nsPath := filepath.Join(dir, ".agents", "namespaces", ns, "commands")
	globalPath := filepath.Join(dir, ".agents", "commands")

	if !slices.Contains(paths, nsPath) {
		t.Errorf("expected namespace command path %q", nsPath)
	}

	if !slices.Contains(paths, globalPath) {
		t.Errorf("expected global command path %q", globalPath)
	}

	if slices.Index(paths, nsPath) >= slices.Index(paths, globalPath) {
		t.Error("expected namespace command path to precede global command path")
	}
}

func TestNamespacedSkillSearchPaths(t *testing.T) {
	t.Parallel()

	dir := testProjectDir
	ns := testNamespace

	paths := namespacedSkillSearchPaths(dir, ns)

	nsPath := filepath.Join(dir, ".agents", "namespaces", ns, "skills")
	globalPath := filepath.Join(dir, ".agents", "skills")

	if !slices.Contains(paths, nsPath) {
		t.Errorf("expected namespace skill path %q", nsPath)
	}

	if !slices.Contains(paths, globalPath) {
		t.Errorf("expected global skill path %q", globalPath)
	}

	if slices.Index(paths, nsPath) >= slices.Index(paths, globalPath) {
		t.Error("expected namespace skill path to precede global skill path")
	}
}

// ── Run() with namespace ─────────────────────────────────────────────────────

func TestRun_NamespacedTask_Found(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "build", "Namespace build task.")

	cc := newRunContext(dir)

	result, err := cc.Run(context.Background(), "myteam/build")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Task.Content, "Namespace build task.") {
		t.Errorf("unexpected task content: %q", result.Task.Content)
	}
}

func TestRun_NamespacedTask_FallsBackToGlobal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Only a global task exists; no namespace task file
	createTask(t, dir, "deploy", "", "Global deploy task.")

	cc := newRunContext(dir)

	result, err := cc.Run(context.Background(), "myteam/deploy")
	if err != nil {
		t.Fatalf("Run() error (expected global fallback): %v", err)
	}

	if !strings.Contains(result.Task.Content, "Global deploy task.") {
		t.Errorf("expected global task content, got %q", result.Task.Content)
	}
}

func TestRun_NamespacedTask_NamespaceTakesPrecedenceOverGlobal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "build", "", "Global build task.")
	createNamespaceTask(t, dir, "myteam", "build", "Namespace build task.")

	cc := newRunContext(dir)

	result, err := cc.Run(context.Background(), "myteam/build")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Task.Content, "Namespace build task.") {
		t.Errorf("expected namespace task to win, got %q", result.Task.Content)
	}

	if strings.Contains(result.Task.Content, "Global build task.") {
		t.Errorf("global task should not appear when namespace task exists")
	}
}

func TestRun_NamespacedTask_NotFound_GlobalNotFound_Error(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cc := newRunContext(dir)

	_, err := cc.Run(context.Background(), "myteam/missing")
	if err == nil {
		t.Fatal("expected error for missing task")
	}

	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestRun_InvalidNamespace_TooManySlashes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cc := newRunContext(dir)

	_, err := cc.Run(context.Background(), "a/b/c")
	if err == nil {
		t.Fatal("expected error for deeply nested task name")
	}

	if !strings.Contains(err.Error(), "one level") {
		t.Errorf("expected 'one level' in error, got %v", err)
	}
}

func TestRun_InvalidNamespace_EmptyNamespace(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cc := newRunContext(dir)

	_, err := cc.Run(context.Background(), "/task")
	if err == nil {
		t.Fatal("expected error for empty namespace")
	}
}

func TestRun_InvalidNamespace_EmptyBaseName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cc := newRunContext(dir)

	_, err := cc.Run(context.Background(), "ns/")
	if err == nil {
		t.Fatal("expected error for empty base name")
	}
}

// ── Rules with namespace ─────────────────────────────────────────────────────

func TestRun_NamespaceRules_IncludedBeforeGlobalRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")
	createNamespaceRule(t, dir, "myteam", "ns-rule", "Namespace rule content.")
	createRule(t, dir, ".agents/rules/global-rule.md", "", "Global rule content.")

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Rules) < 2 {
		t.Fatalf("expected at least 2 rules (ns + global), got %d", len(result.Rules))
	}

	// Namespace rule must appear before global rule in the assembled prompt
	nsIdx := strings.Index(result.Prompt, "Namespace rule content.")
	globalIdx := strings.Index(result.Prompt, "Global rule content.")

	if nsIdx < 0 {
		t.Error("namespace rule content missing from prompt")
	}

	if globalIdx < 0 {
		t.Error("global rule content missing from prompt")
	}

	if nsIdx >= globalIdx {
		t.Error("expected namespace rule to appear before global rule in prompt")
	}
}

func TestRun_GlobalRules_AlwaysIncludedWithNamespace(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")
	createRule(t, dir, ".agents/rules/global-rule.md", "", "Global rule content.")
	// No namespace-specific rules

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Prompt, "Global rule content.") {
		t.Error("global rule should always be included even when using a namespace")
	}
}

func TestRun_NamespaceSelector_FiltersRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")
	// This rule restricts itself to myteam via the selector system
	createRule(t, dir, ".agents/rules/myteam-only.md", "namespace: myteam", "myteam-only rule.")
	// This rule restricts itself to otherteam
	createRule(t, dir, ".agents/rules/otherteam-only.md", "namespace: otherteam", "otherteam-only rule.")
	// Unrestricted global rule
	createRule(t, dir, ".agents/rules/everyone.md", "", "everyone rule.")

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Prompt, "myteam-only rule.") {
		t.Error("myteam-scoped rule should be included for myteam task")
	}

	if strings.Contains(result.Prompt, "otherteam-only rule.") {
		t.Error("otherteam-scoped rule must not be included for myteam task")
	}

	if !strings.Contains(result.Prompt, "everyone rule.") {
		t.Error("unrestricted global rule should always be included")
	}
}

func TestRun_NoNamespace_NamespaceScopedRulesExcluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "plain", "", "Plain task.")
	// A rule restricted to myteam — should NOT appear for a non-namespaced task
	createRule(t, dir, ".agents/rules/myteam-only.md", "namespace: myteam", "myteam-only rule.")
	createRule(t, dir, ".agents/rules/global.md", "", "Global rule.")

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "plain")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if strings.Contains(result.Prompt, "myteam-only rule.") {
		t.Error("namespace-scoped rule must not be included for non-namespaced tasks")
	}

	if !strings.Contains(result.Prompt, "Global rule.") {
		t.Error("global rule should be included for non-namespaced tasks")
	}
}

func TestRun_DifferentNamespaces_DoNotShareRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "teamA", "work", "TeamA work.")
	createNamespaceTask(t, dir, "teamB", "work", "TeamB work.")
	createNamespaceRule(t, dir, "teamA", "rule", "TeamA namespace rule.")
	createNamespaceRule(t, dir, "teamB", "rule", "TeamB namespace rule.")

	// Run as teamA
	ccA := newFullContext(dir)

	resultA, err := ccA.Run(context.Background(), "teamA/work")
	if err != nil {
		t.Fatalf("teamA Run() error: %v", err)
	}

	if !strings.Contains(resultA.Prompt, "TeamA namespace rule.") {
		t.Error("teamA should see its own namespace rule")
	}

	if strings.Contains(resultA.Prompt, "TeamB namespace rule.") {
		t.Error("teamA must not see teamB namespace rules")
	}

	// Run as teamB (needs a fresh Context)
	ccB := newFullContext(dir)

	resultB, err := ccB.Run(context.Background(), "teamB/work")
	if err != nil {
		t.Fatalf("teamB Run() error: %v", err)
	}

	if !strings.Contains(resultB.Prompt, "TeamB namespace rule.") {
		t.Error("teamB should see its own namespace rule")
	}

	if strings.Contains(resultB.Prompt, "TeamA namespace rule.") {
		t.Error("teamB must not see teamA namespace rules")
	}
}

// ── Commands with namespace ──────────────────────────────────────────────────

func TestRun_NamespaceCommand_OverridesGlobalCommand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "deploy", "/deploy")
	createCommand(t, dir, "deploy", "", "Global deploy command.")
	createNamespaceCommand(t, dir, "myteam", "deploy", "", "Namespace deploy command.")

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "myteam/deploy")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Task.Content, "Namespace deploy command.") {
		t.Error("namespace command should override global command")
	}

	if strings.Contains(result.Task.Content, "Global deploy command.") {
		t.Error("global command must not appear when namespace command has same name")
	}
}

func TestRun_NamespaceCommand_FallsBackToGlobalCommand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "/shared-cmd")
	createCommand(t, dir, "shared-cmd", "", "Shared global command.")
	// No namespace override for this command

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Task.Content, "Shared global command.") {
		t.Error("should fall back to global command when no namespace override exists")
	}
}

// ── Skills with namespace ────────────────────────────────────────────────────

func TestRun_NamespaceSkills_DiscoveredAlongsideGlobal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")
	createNamespaceSkill(t, dir, "myteam", "team-tool",
		"---\nname: team-tool\ndescription: A team-specific tool skill.\n---\n")
	createSkill(t, dir, ".agents/skills/global-tool",
		"---\nname: global-tool\ndescription: A global tool skill.\n---\n")

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	names := make([]string, 0, len(result.Skills.Skills))
	for _, s := range result.Skills.Skills {
		names = append(names, s.Name)
	}

	if !slices.Contains(names, "team-tool") {
		t.Errorf("namespace skill 'team-tool' not discovered; got %v", names)
	}

	if !slices.Contains(names, "global-tool") {
		t.Errorf("global skill 'global-tool' not discovered; got %v", names)
	}
}

func TestRun_NamespaceSkills_ListedBeforeGlobalSkills(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")
	createNamespaceSkill(t, dir, "myteam", "alpha-skill",
		"---\nname: alpha-skill\ndescription: Namespace alpha skill.\n---\n")
	createSkill(t, dir, ".agents/skills/beta-skill",
		"---\nname: beta-skill\ndescription: Global beta skill.\n---\n")

	cc := newFullContext(dir)

	result, err := cc.Run(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if len(result.Skills.Skills) < 2 {
		t.Fatalf("expected at least 2 skills, got %d", len(result.Skills.Skills))
	}

	// Namespace skill must appear first
	if result.Skills.Skills[0].Name != "alpha-skill" {
		t.Errorf("expected namespace skill first, got %q", result.Skills.Skills[0].Name)
	}
}

// ── Selector state after findTask ────────────────────────────────────────────

func TestRun_NamespacedTask_SetsNamespaceSelector(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")

	cc := newRunContext(dir)

	_, err := cc.Run(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !cc.includes.GetValue("namespace", "myteam") {
		t.Error("expected namespace=myteam to be set in selectors after Run()")
	}
}

func TestRun_NamespacedTask_SetsBothTaskNameForms(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "fix-bug", "Fix the bug.")

	cc := newRunContext(dir)

	_, err := cc.Run(context.Background(), "myteam/fix-bug")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Both full and base task names should be present
	if !cc.includes.GetValue("task_name", "myteam/fix-bug") {
		t.Error("expected task_name=myteam/fix-bug in selectors")
	}

	if !cc.includes.GetValue("task_name", "fix-bug") {
		t.Error("expected task_name=fix-bug (base) in selectors")
	}
}

func TestRun_NonNamespacedTask_NamespaceSelectorIsEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "plain", "", "Plain task.")

	cc := newRunContext(dir)

	_, err := cc.Run(context.Background(), "plain")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// For non-namespaced tasks, namespace selector is set to "" (empty string sentinel)
	// so that rules declaring a specific namespace in frontmatter are excluded.
	if !cc.includes.GetValue("namespace", "") {
		t.Error("expected namespace=\"\" (empty sentinel) for non-namespaced tasks")
	}

	if !cc.includes.GetValue("task_name", "plain") {
		t.Error("expected task_name=plain in selectors")
	}
}

// ── Lint() with namespace ────────────────────────────────────────────────────

func TestLint_NamespacedTask_NoErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")
	createNamespaceRule(t, dir, "myteam", "ns-rule", "Namespace rule.")
	createRule(t, dir, ".agents/rules/global.md", "", "Global rule.")

	cc := New(WithSearchPaths(dir))

	result, err := cc.Lint(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Lint() error: %v", err)
	}

	if len(result.Errors) != 0 {
		t.Errorf("expected no lint errors, got %+v", result.Errors)
	}
}

func TestLint_NamespacedTask_TracksNamespaceTaskFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")

	cc := New(WithSearchPaths(dir))

	result, err := cc.Lint(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Lint() error: %v", err)
	}

	if !hasLoadedFile(result, filepath.Join("namespaces", "myteam", "tasks", "work.md"), LoadedFileKindTask) {
		t.Errorf("expected namespace task file in LoadedFiles, got %+v", result.LoadedFiles)
	}
}

func TestLint_NamespacedTask_NamespaceNotFlaggedAsUnmatchedSelector(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")

	cc := New(WithSearchPaths(dir))

	result, err := cc.Lint(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Lint() error: %v", err)
	}

	if hasLintError(result, LintErrorKindSelectorNoMatch, "namespace") {
		t.Error("'namespace' selector must not produce a SelectorNoMatch lint error")
	}
}

func TestLint_InvalidNamespacedTaskName_FatalError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cc := New(WithSearchPaths(dir))

	_, err := cc.Lint(context.Background(), "a/b/c")
	if err == nil {
		t.Fatal("expected fatal error for invalid task name")
	}

	if !strings.Contains(err.Error(), "one level") {
		t.Errorf("expected 'one level' in error, got %v", err)
	}
}

func TestLint_NamespacedTask_TracksBothNamespaceAndGlobalRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createNamespaceTask(t, dir, "myteam", "work", "Do work.")
	createNamespaceRule(t, dir, "myteam", "ns-rule", "NS rule.")
	createRule(t, dir, ".agents/rules/global-rule.md", "", "Global rule.")

	cc := New(WithSearchPaths(dir))

	result, err := cc.Lint(context.Background(), "myteam/work")
	if err != nil {
		t.Fatalf("Lint() error: %v", err)
	}

	if !hasLoadedFile(result, "ns-rule.md", LoadedFileKindRule) {
		t.Errorf("expected namespace rule in LoadedFiles, got %+v", result.LoadedFiles)
	}

	if !hasLoadedFile(result, "global-rule.md", LoadedFileKindRule) {
		t.Errorf("expected global rule in LoadedFiles, got %+v", result.LoadedFiles)
	}
}
