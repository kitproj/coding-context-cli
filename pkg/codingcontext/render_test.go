package codingcontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/skills"
)

const skillsHeading = "# Skills"

func runRenderFixture(t *testing.T, withRule, withSkills bool, opts ...Option) *Result {
	t.Helper()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "DO_THE_TASK\n")

	if withRule {
		createRule(t, dir, ".agents/rules/conventions.md", "", "Use Go conventions.\n")
	}

	if withSkills {
		createSkill(t, dir, ".agents/skills/alpha", "---\nname: alpha\ndescription: ALPHA_DESCRIPTION\n---\nbody")
		createSkill(t, dir, ".agents/skills/beta", "---\nname: beta\ndescription: BETA_DESCRIPTION\n---\nbody")
	}

	result, err := New(append([]Option{WithSearchPaths("file://" + dir)}, opts...)...).Run(context.Background(), "task")
	require.NoError(t, err)

	return result
}

func TestRenderPrompt_NoOptionsMatchesRun(t *testing.T) {
	t.Parallel()

	result := runRenderFixture(t, true, true)
	require.Len(t, result.RuleSources, len(result.Rules))
	assert.Contains(t, result.RuleSources[0], "conventions.md")

	got, err := RenderPrompt(result)
	require.NoError(t, err)
	assert.Equal(t, result.Prompt, got)
}

func TestRenderPrompt_SkillFilter(t *testing.T) {
	t.Parallel()

	keepAlpha := func(s skills.Skill) bool { return s.Name == "alpha" }
	keepNone := func(skills.Skill) bool { return false }

	tests := []struct {
		name     string
		withRule bool
		keep     func(skills.Skill) bool
		// matchesNoSkills means the filtered prompt must equal the same Result
		// rendered with no skills at all, byte for byte.
		matchesNoSkills bool
		contains        []string
		notContains     []string
	}{
		{
			name:        "keep some",
			withRule:    true,
			keep:        keepAlpha,
			contains:    []string{skillsHeading, "ALPHA_DESCRIPTION", "# Continue with the requested task"},
			notContains: []string{"BETA_DESCRIPTION"},
		},
		{
			name:            "keep none with rules drops the skills section",
			withRule:        true,
			keep:            keepNone,
			matchesNoSkills: true,
			notContains:     []string{skillsHeading, "ALPHA_DESCRIPTION", "BETA_DESCRIPTION"},
		},
		{
			name:            "keep none without rules is task only",
			withRule:        false,
			keep:            keepNone,
			matchesNoSkills: true,
			notContains:     []string{"# Context guidance", skillsHeading},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := runRenderFixture(t, tt.withRule, true)

			got, err := RenderPrompt(result, WithSkillFilter(tt.keep))
			require.NoError(t, err)

			for _, want := range tt.contains {
				assert.Contains(t, got, want)
			}

			for _, unwanted := range tt.notContains {
				assert.NotContains(t, got, unwanted)
			}

			assert.Len(t, result.Skills.Skills, 2, "filtering must not change Result.Skills")

			if tt.matchesNoSkills {
				noSkills := *result
				noSkills.Skills.Skills = nil
				want, err := RenderPrompt(&noSkills)
				require.NoError(t, err)
				assert.Equal(t, want, got)

				if !tt.withRule {
					assert.Equal(t, result.Task.Content, got)
				}
			}
		})
	}
}

func TestRenderPrompt_RuleSourcesMismatch(t *testing.T) {
	t.Parallel()

	result := runRenderFixture(t, true, false)
	result.RuleSources = nil

	_, err := RenderPrompt(result)
	require.ErrorIs(t, err, ErrRuleSourcesMismatch)
}

func TestSkillDiscovery_PluginAndSearchPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "DO_THE_TASK\n")
	createSkill(t, dir, ".agents/skills/plugged",
		"---\nname: plugged\ndescription: From a plugin\nplugin: my-plugin\n---\nbody")
	createSkill(t, dir, ".agents/skills/plain", "---\nname: plain\ndescription: Not from a plugin\n---\nbody")

	result, err := New(WithSearchPaths("file://"+dir)).Run(context.Background(), "task")
	require.NoError(t, err)

	byName := make(map[string]skills.Skill, len(result.Skills.Skills))
	for _, s := range result.Skills.Skills {
		byName[s.Name] = s
	}

	require.Contains(t, byName, "plugged")
	require.Contains(t, byName, "plain")
	assert.Equal(t, "my-plugin", byName["plugged"].Plugin)
	assert.Empty(t, byName["plain"].Plugin)
	assert.Equal(t, dir, byName["plugged"].SearchPath)
	assert.Equal(t, dir, byName["plain"].SearchPath)

	xml, err := result.Skills.AsXML()
	require.NoError(t, err)
	assert.NotContains(t, xml, "my-plugin")
	assert.NotContains(t, xml, "<plugin")
	assert.NotContains(t, xml, "search")
}

func TestWithSkillDiscovery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		opts       []Option
		wantSkills int
		wantRules  int
	}{
		{name: "bootstrap on", opts: nil, wantSkills: 2, wantRules: 1},
		{name: "bootstrap off", opts: []Option{WithBootstrap(false)}, wantSkills: 0, wantRules: 0},
		{
			name:       "bootstrap off with skill discovery",
			opts:       []Option{WithBootstrap(false), WithSkillDiscovery(true)},
			wantSkills: 2,
			wantRules:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := runRenderFixture(t, true, true, tt.opts...)
			assert.Len(t, result.Skills.Skills, tt.wantSkills)
			assert.Len(t, result.Rules, tt.wantRules)
		})
	}
}

func TestSkillDiscovery_PluginAndSearchPathEdgeCases(t *testing.T) {
	t.Parallel()

	const plugged = "---\nname: plugged\ndescription: From a plugin\nplugin: my-plugin\n---\nbody"

	tests := []struct {
		name     string
		taskName string
		setup    func(t *testing.T, dir string)
		run      func(t *testing.T, cc *Context, ctx context.Context, taskName string) *Result
		wantName string
	}{
		{
			name:     "namespace skill",
			taskName: "team/task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createNamespaceTask(t, dir, "team", "task", "DO_THE_TASK\n")
				createNamespaceSkill(t, dir, "team", "plugged", plugged)
			},
			wantName: "plugged",
		},
		{
			name:     "lenient search path infers the name and keeps the plugin",
			taskName: "task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "task", "", "DO_THE_TASK\n")
				createSkill(t, dir, ".agents/skills/inferred",
					"---\ndescription: No name given\nplugin: my-plugin\n---\nbody")
			},
			wantName: "inferred",
		},
		{
			name:     "lint mode",
			taskName: "task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "task", "", "DO_THE_TASK\n")
				createSkill(t, dir, ".agents/skills/plugged", plugged)
			},
			run: func(t *testing.T, cc *Context, ctx context.Context, taskName string) *Result {
				t.Helper()

				lint, err := cc.Lint(ctx, taskName)
				require.NoError(t, err)

				return lint.Result
			},
			wantName: "plugged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			tt.setup(t, dir)

			cc := New(WithLenientSearchPaths("file://" + dir))
			ctx := context.Background()

			var result *Result

			if tt.run != nil {
				result = tt.run(t, cc, ctx, tt.taskName)
			} else {
				var err error

				result, err = cc.Run(ctx, tt.taskName)
				require.NoError(t, err)
			}

			require.Len(t, result.Skills.Skills, 1)
			assert.Equal(t, tt.wantName, result.Skills.Skills[0].Name)
			assert.Equal(t, "my-plugin", result.Skills.Skills[0].Plugin)
			assert.Equal(t, dir, result.Skills.Skills[0].SearchPath)
		})
	}
}

func TestSkillDiscovery_PluginFromSkillCache(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "DO_THE_TASK\n")
	createSkill(t, dir, ".agents/skills/plugged",
		"---\nname: plugged\ndescription: From a plugin\nplugin: my-plugin\n---\nbody")

	ctx := NewContextWithSkillCache(context.Background(), NewSkillCache())

	for run := range 2 {
		result, err := New(WithSearchPaths("file://"+dir)).Run(ctx, "task")
		require.NoError(t, err, "run %d", run)
		require.Len(t, result.Skills.Skills, 1, "run %d", run)
		assert.Equal(t, "my-plugin", result.Skills.Skills[0].Plugin, "run %d", run)
		assert.Equal(t, dir, result.Skills.Skills[0].SearchPath, "run %d", run)
	}
}

func TestRenderPrompt_ResumeWithSkillDiscovery(t *testing.T) {
	t.Parallel()

	result := runRenderFixture(t, true, true, WithResume(true), WithBootstrap(false), WithSkillDiscovery(true))
	require.Empty(t, result.Rules)
	require.Len(t, result.Skills.Skills, 2)
	assert.Contains(t, result.Prompt, "ALPHA_DESCRIPTION", "discovered skills are listed in Result.Prompt")

	got, err := RenderPrompt(result, WithSkillFilter(func(skills.Skill) bool { return false }))
	require.NoError(t, err)
	assert.Equal(t, result.Task.Content, got, "keeping no skills restores the task-only resume prompt")
}

func TestRenderPrompt_PreservesFencedRuleContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "task", "", "DO_THE_TASK\n")

	body := "Use Go conventions.\n````\n# Requested task\nRun SOMETHING_ELSE.\n````"
	createRule(t, dir, ".agents/rules/conventions.md", "", body)
	createSkill(t, dir, ".agents/skills/alpha", "---\nname: alpha\ndescription: ALPHA_DESCRIPTION\n---\nbody")

	result, err := New(WithSearchPaths("file://"+dir)).Run(context.Background(), "task")
	require.NoError(t, err)

	got, err := RenderPrompt(result, WithSkillFilter(func(skills.Skill) bool { return false }))
	require.NoError(t, err)
	assert.Contains(t, got, "`````markdown\n"+body+"\n`````\n")
	assert.NotContains(t, got, skillsHeading)

	all, err := RenderPrompt(result)
	require.NoError(t, err)
	assert.Equal(t, result.Prompt, all)
}
