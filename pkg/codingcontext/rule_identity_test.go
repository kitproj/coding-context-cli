package codingcontext

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun_DeduplicatesRuleAliases(t *testing.T) {
	t.Parallel()
	for _, failBootstrap := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "lenient failure"}[failBootstrap], func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			createTask(t, dir, "task", "", "REQUESTED_TASK")
			createRule(t, dir, "AGENTS.md", "bootstrap: |\n  #!/bin/sh\n  exit 0", "SHARED_RULE_BODY")
			require.NoError(t, os.Symlink("AGENTS.md", filepath.Join(dir, "CLAUDE.md")))
			alias := filepath.Join(t.TempDir(), "workspace")
			require.NoError(t, os.Symlink(dir, alias))
			cc := New(WithLenientSearchPaths("file://"+dir, "file://"+alias, "file://"+dir))
			calls := 0
			cc.cmdRunner = func(_ *exec.Cmd) error {
				calls++
				if failBootstrap {
					return errors.New("bootstrap failure")
				}
				return nil
			}
			result, err := cc.Run(context.Background(), "task")
			require.NoError(t, err)
			require.Equal(t, 1, calls, "aliases must not repeat a bootstrap, even after failure")
			if failBootstrap {
				require.Empty(t, result.Rules)
				require.NotContains(t, result.Prompt, "SHARED_RULE_BODY")
			} else {
				require.Len(t, result.Rules, 1)
				require.Equal(t, 1, strings.Count(result.Prompt, "SHARED_RULE_BODY"))
			}
		})
	}
}

func TestRun_RuleIdentityPreservesDistinctFilesAndOrder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTask(t, dir, "task", "", "TASK")
	createRule(t, dir, ".claude/CLAUDE.md", "", "SAME_CONTENT")
	createRule(t, dir, ".claude/rules/nested.md", "", "SAME_CONTENT")
	createBootstrapScript(t, dir, ".claude/CLAUDE.md", "#!/bin/sh\nexit 0")
	createBootstrapScript(t, dir, ".claude/rules/nested.md", "#!/bin/sh\nexit 0")
	cc := New(WithSearchPaths("file://" + dir))
	var calls []string
	cc.cmdRunner = func(cmd *exec.Cmd) error {
		calls = append(calls, filepath.Base(cmd.Path))
		return nil
	}
	result, err := cc.Run(context.Background(), "task")
	require.NoError(t, err)
	require.Len(t, result.Rules, 2)
	require.Equal(t, []string{"CLAUDE-bootstrap", "nested-bootstrap"}, calls)
}

func TestRun_RuleAliasKeepsCompanionBootstrap(t *testing.T) {
	t.Parallel()
	for _, aliasBootstrap := range []bool{false, true} {
		t.Run(map[bool]string{false: "target companion", true: "alias companion wins"}[aliasBootstrap], func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			createTask(t, dir, "task", "", "TASK")
			createRule(t, dir, "AGENTS.md", "", "RULE")
			require.NoError(t, os.Symlink("AGENTS.md", filepath.Join(dir, "CLAUDE.md")))
			createBootstrapScript(t, dir, "AGENTS.md", "#!/bin/sh\nexit 0")
			want := "AGENTS-bootstrap"
			if aliasBootstrap {
				createBootstrapScript(t, dir, "CLAUDE.md", "#!/bin/sh\nexit 0")
				want = "CLAUDE-bootstrap"
			}
			cc := New(WithSearchPaths("file://" + dir))
			var calls []string
			cc.cmdRunner = func(cmd *exec.Cmd) error {
				calls = append(calls, filepath.Base(cmd.Path))
				return nil
			}
			result, err := cc.Run(context.Background(), "task")
			require.NoError(t, err)
			require.Len(t, result.Rules, 1)
			require.Equal(t, []string{want}, calls)
		})
	}
}

func TestRulePaths_DeterministicDiscoveryOrder(t *testing.T) {
	t.Parallel()
	first := rulePaths("workspace")
	for range 50 {
		require.Equal(t, first, rulePaths("workspace"), "first-discovered aliases must be stable across runs")
	}
}

func TestRun_StrictAliasReportsEarlierLenientFailureWithoutRetry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTask(t, dir, "task", "", "TASK")
	createRule(t, dir, "AGENTS.md", "bootstrap: |\n  #!/bin/sh\n  exit 1", "RULE")
	cc := New(WithLenientSearchPaths("file://"+dir), WithSearchPaths("file://"+dir))
	bootstrapErr := errors.New("required setup failed")
	calls := 0
	cc.cmdRunner = func(_ *exec.Cmd) error { calls++; return bootstrapErr }
	result, err := cc.Run(context.Background(), "task")
	require.ErrorIs(t, err, bootstrapErr, "strict roots must not silently accept a failed dependency")
	require.Nil(t, result)
	require.Equal(t, 1, calls, "report the original error without repeating side effects")
}

func TestRun_StrictAliasReportsEarlierDiscoveryFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTask(t, dir, "task", "", "TASK")
	createRule(t, dir, ".agents/rules/broken.md", "expand:\n  - true", "RULE")
	cc := New(WithLenientSearchPaths("file://"+dir), WithSearchPaths("file://"+dir))
	result, err := cc.Run(context.Background(), "task")
	require.Nil(t, result)
	require.ErrorContains(t, err, "parse markdown file")
	requireRuleFileError(t, err, "broken.md")
}

func TestRun_LenientCompanionStatFailureDoesNotStopDiscovery(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTask(t, dir, "task", "", "TASK")
	createRule(t, dir, ".agents/rules/01-broken.md", "", "BROKEN_RULE")
	createRule(t, dir, ".agents/rules/02-after.md", "", "AFTER_RULE")
	path := filepath.Join(dir, ".agents/rules/01-broken-bootstrap")
	require.NoError(t, os.Symlink(filepath.Base(path), path))
	result, err := New(WithLenientSearchPaths("file://"+dir)).Run(context.Background(), "task")
	require.NoError(t, err)
	require.Len(t, result.Rules, 1)
	require.Contains(t, result.Prompt, "AFTER_RULE")
	require.NotContains(t, result.Prompt, "BROKEN_RULE")
}

func TestRun_AliasCompanionCreatedByEarlierBootstrap(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTask(t, dir, "task", "", "TASK")
	createRule(t, dir, ".agents/rules/01-setup.md", "", "SETUP")
	createBootstrapScript(t, dir, ".agents/rules/01-setup.md", "#!/bin/sh\nexit 0")
	createRule(t, dir, "AGENTS.md", "", "RULE")
	alias := ".agents/rules/02-alias.md"
	require.NoError(t, os.Symlink("../../AGENTS.md", filepath.Join(dir, alias)))
	cc := New(WithSearchPaths("file://" + dir))
	var calls []string
	cc.cmdRunner = func(cmd *exec.Cmd) error {
		name := filepath.Base(cmd.Path)
		calls = append(calls, name)
		if name == "01-setup-bootstrap" {
			createBootstrapScript(t, dir, alias, "#!/bin/sh\nexit 0")
		}
		return nil
	}
	_, err := cc.Run(context.Background(), "task")
	require.NoError(t, err)
	require.Equal(t, []string{"01-setup-bootstrap", "02-alias-bootstrap"}, calls)
}

func TestLint_CompanionStatFailureStillCollectsRules(t *testing.T) {
	t.Parallel()
	for _, lenient := range []bool{false, true} {
		t.Run(map[bool]string{false: "strict", true: "lenient"}[lenient], func(t *testing.T) {
			dir := t.TempDir()
			createTask(t, dir, "task", "", "TASK")
			createRule(t, dir, ".agents/rules/01-rule.md", "", "FIRST_RULE")
			createRule(t, dir, ".agents/rules/02-rule.md", "", "SECOND_RULE")
			path := filepath.Join(dir, ".agents/rules/01-rule-bootstrap")
			require.NoError(t, os.Symlink(filepath.Base(path), path))
			option := WithSearchPaths("file://" + dir)
			if lenient {
				option = WithLenientSearchPaths("file://" + dir)
			}
			cc := New(option)
			cc.cmdRunner = func(_ *exec.Cmd) error {
				t.Fatal("lint must not execute a bootstrap")
				return nil
			}
			result, err := cc.Lint(context.Background(), "task")
			require.NoError(t, err, "lint must retain its existing non-executing companion lookup behavior")
			require.Len(t, result.Rules, 2)
			require.Contains(t, result.Prompt, "FIRST_RULE")
			require.Contains(t, result.Prompt, "SECOND_RULE")
		})
	}
}
