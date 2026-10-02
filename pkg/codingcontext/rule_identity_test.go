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
