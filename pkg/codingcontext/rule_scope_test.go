package codingcontext

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun_RuleDiscoveryDoesNotInlineAgentContent(t *testing.T) {
	t.Parallel()

	for _, agentDir := range []string{".claude", ".codex", ".gemini"} {
		t.Run(agentDir, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			createTask(t, dir, "edit-readme", "single_shot: true", "Add hi to README.")
			createRule(t, dir, filepath.Join(agentDir, "rules", "nested", "conventions.md"), "", "REPOSITORY_RULE")
			createBootstrapScript(t, dir, filepath.Join(agentDir, "rules", "nested", "conventions.md"), "#!/bin/sh\nexit 0")
			for i := range 100 {
				skillDir := filepath.Join(agentDir, "skills", fmt.Sprintf("skill-%d", i))
				createSkill(t, dir, skillDir, fmt.Sprintf("---\nname: skill-%d\ndescription: A useful skill\nbootstrap: |\n  #!/bin/sh\n  exit 0\n---\n%s", i, strings.Repeat("SKILL_BODY ", 1500)))
				createRule(t, dir, filepath.Join(skillDir, "references", "guide.md"), "", "SKILL_REFERENCE")
			}
			for _, path := range []string{"commands/edit.md", "plugins/cache/plugin/skills/tool/SKILL.md", "notes.md"} {
				createRule(t, dir, filepath.Join(agentDir, path), "", "UNRELATED_CONTENT")
			}

			cc := New(
				WithLenientSearchPaths("file://"+dir),
				WithLenientAgent(AgentCursor),
				WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
			)
			var bootstraps []string
			cc.cmdRunner = func(cmd *exec.Cmd) error {
				bootstraps = append(bootstraps, cmd.Path)
				return nil
			}
			result, err := cc.Run(context.Background(), "edit-readme")
			require.NoError(t, err)
			require.Equal(t, 1, len(result.Rules))
			require.Len(t, result.Skills.Skills, 100)
			require.Contains(t, result.Prompt, "REPOSITORY_RULE")
			require.Contains(t, result.Prompt, "skill-99")
			require.NotContains(t, result.Prompt, "SKILL_BODY")
			require.NotContains(t, result.Prompt, "SKILL_REFERENCE")
			require.NotContains(t, result.Prompt, "UNRELATED_CONTENT")
			require.Less(t, len(result.Prompt), 50000, "prompt size should depend on skill metadata, not bodies")
			require.Equal(t, []string{filepath.Join(dir, agentDir, "rules", "nested", "conventions-bootstrap")}, bootstraps,
				"only the real rule bootstrap should execute")
		})
	}
}

func TestRun_AgentInstructionsBootstrapBeforeNestedRules(t *testing.T) {
	t.Parallel()
	for _, instruction := range []string{".claude/CLAUDE.md", ".codex/AGENTS.md", ".gemini/GEMINI.md"} {
		t.Run(instruction, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			createTask(t, dir, "task", "", "TASK_INSTRUCTIONS")
			createRule(t, dir, instruction, "", "INITIALIZATION_RULE")
			createBootstrapScript(t, dir, instruction, "#!/bin/sh\nexit 0")
			nested := filepath.Join(filepath.Dir(instruction), "rules", "dependent.md")
			createRule(t, dir, nested, "", "DEPENDENT_RULE")
			createBootstrapScript(t, dir, nested, "#!/bin/sh\nexit 0")
			cc := New(WithSearchPaths("file://" + dir))
			initialized := false
			cc.cmdRunner = func(cmd *exec.Cmd) error {
				if filepath.Base(cmd.Path) == "dependent-bootstrap" {
					if !initialized {
						return errors.New("agent instructions have not initialized the workspace")
					}
					return nil
				}
				initialized = true
				return nil
			}
			result, err := cc.Run(context.Background(), "task")
			require.NoError(t, err)
			require.Len(t, result.Rules, 2)
			require.Equal(t, []string{"INITIALIZATION_RULE", "DEPENDENT_RULE"}, []string{result.Rules[0].Content, result.Rules[1].Content})
		})
	}
}

func TestRun_ScopedRulesPreserveNamespaceAndSelectors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createRule(t, dir, ".agents/namespaces/team/tasks/task.md", "selectors:\n  language: go", "TASK_INSTRUCTIONS")
	createRule(t, dir, ".agents/namespaces/team/rules/team.md", "", "TEAM_RULE")
	createRule(t, dir, ".claude/rules/go.md", "language: go", "GO_RULE")
	createRule(t, dir, ".claude/rules/python.md", "language: python", "PYTHON_RULE")
	createRule(t, dir, ".codex/AGENTS.md", "", "CODEX_INSTRUCTIONS")

	result, err := New(WithSearchPaths("file://"+dir), WithAgent(AgentCodex)).Run(context.Background(), "team/task")
	require.NoError(t, err)
	require.Contains(t, result.Prompt, "TASK_INSTRUCTIONS")
	require.Contains(t, result.Prompt, "TEAM_RULE")
	require.Contains(t, result.Prompt, "GO_RULE")
	require.NotContains(t, result.Prompt, "PYTHON_RULE")
	require.Contains(t, result.Prompt, "CODEX_INSTRUCTIONS")
}

func TestRun_PreservesExplicitInstructionFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	createTask(t, dir, "task", "", "Do the task.")
	paths := []string{
		"CLAUDE.md", "CLAUDE.local.md", ".claude/CLAUDE.md", ".claude/CLAUDE.local.md",
		"AGENTS.md", ".codex/AGENTS.md",
		"GEMINI.md", ".gemini/GEMINI.md", ".gemini/styleguide.md",
	}
	for i, path := range paths {
		createRule(t, dir, path, "", fmt.Sprintf("instruction-%d", i))
	}
	result, err := New(WithSearchPaths("file://"+dir)).Run(context.Background(), "task")
	require.NoError(t, err)
	require.Len(t, result.Rules, len(paths), "each instruction file should be included exactly once")
	for i := range paths {
		require.Equal(t, 1, strings.Count(result.Prompt, fmt.Sprintf("instruction-%d", i)))
	}
}
