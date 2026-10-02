package codingcontext

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/tokencount"
)

func bootstrapIsolationRuleNames(rules []markdown.Markdown[markdown.RuleFrontMatter]) []string {
	names := make([]string, 0, len(rules))
	for _, rule := range rules {
		names = append(names, rule.FrontMatter.Name)
	}

	return names
}

func bootstrapIsolationRuleContent(
	tb testing.TB,
	rules []markdown.Markdown[markdown.RuleFrontMatter],
	name string,
) string {
	tb.Helper()

	for _, rule := range rules {
		if rule.FrontMatter.Name == name {
			return rule.Content
		}
	}

	tb.Fatalf("rule %q not found", name)

	return ""
}

func rewriteBootstrapIsolationRule(tb testing.TB, path, name, content string) {
	tb.Helper()

	source := fmt.Sprintf("---\nname: %s\n---\n%s", name, content)
	require.NoError(tb, os.WriteFile(path, []byte(source), 0o600))
}

func requireRuleFileError(tb testing.TB, err error, name string) {
	tb.Helper()

	var rfe *ruleFileError
	require.ErrorAs(tb, err, &rfe)
	require.Equal(tb, name, filepath.Base(rfe.path))
}

func TestRun_LenientBootstrapFailureSkipsOnlyFailingRule(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "test-task", "", "Task content")
	createRule(t, dir, ".agents/rules/01-before.md", "name: before", "Before content")
	createRule(t, dir, ".agents/rules/02-failing.md", "name: failing", "Failing content")
	createRule(t, dir, ".agents/rules/03-after.md", "name: after", "After content")
	createBootstrapScript(t, dir, ".agents/rules/01-before.md", "#!/bin/sh\nexit 0")
	createBootstrapScript(t, dir, ".agents/rules/02-failing.md", "#!/bin/sh\nexit 1")
	createBootstrapScript(t, dir, ".agents/rules/03-after.md", "#!/bin/sh\nexit 0")

	bootstrapErr := errors.New("bootstrap sentinel")
	var calls []string
	var logs strings.Builder
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	cc := New(
		WithLenientSearchPaths("file://"+dir),
		WithLogger(logger),
	)
	cc.cmdRunner = func(cmd *exec.Cmd) error {
		name := filepath.Base(cmd.Path)
		calls = append(calls, name)
		if name == "02-failing-bootstrap" {
			return bootstrapErr
		}
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.NoError(t, err)
	require.Equal(t, []string{"before", "after"}, bootstrapIsolationRuleNames(result.Rules))
	require.Equal(t, []string{
		"01-before-bootstrap",
		"02-failing-bootstrap",
		"03-after-bootstrap",
	}, calls)
	require.Contains(t, result.Prompt, "Before content")
	require.NotContains(t, result.Prompt, "Failing content")
	require.Contains(t, result.Prompt, "After content")
	require.Contains(t, logs.String(), "skipping rule file after bootstrap failure")
	require.Contains(t, logs.String(), "02-failing.md")
	require.Contains(t, logs.String(), bootstrapErr.Error())

	expectedTokens := result.Task.Tokens
	for _, rule := range result.Rules {
		expectedTokens += tokencount.EstimateTokens(rule.Content)
	}
	require.Equal(t, expectedTokens, result.Tokens)
}

func TestRun_StrictBootstrapFailureIsFatal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "test-task", "", "Task content")
	createRule(t, dir, ".agents/rules/01-before.md", "name: before", "Before content")
	createRule(t, dir, ".agents/rules/02-failing.md", "name: failing", "Failing content")
	createRule(t, dir, ".agents/rules/03-after.md", "name: after", "After content")
	createBootstrapScript(t, dir, ".agents/rules/01-before.md", "#!/bin/sh\nexit 0")
	createBootstrapScript(t, dir, ".agents/rules/02-failing.md", "#!/bin/sh\nexit 1")
	createBootstrapScript(t, dir, ".agents/rules/03-after.md", "#!/bin/sh\nexit 0")

	bootstrapErr := errors.New("bootstrap sentinel")
	var calls []string

	cc := New(WithSearchPaths("file://" + dir))
	cc.cmdRunner = func(cmd *exec.Cmd) error {
		name := filepath.Base(cmd.Path)
		calls = append(calls, name)
		if name == "02-failing-bootstrap" {
			return bootstrapErr
		}
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.Nil(t, result)
	require.ErrorIs(t, err, bootstrapErr)
	requireRuleFileError(t, err, "02-failing.md")
	require.Equal(t, []string{
		"01-before-bootstrap",
		"02-failing-bootstrap",
	}, calls)
	require.Equal(t, []string{"before"}, bootstrapIsolationRuleNames(cc.rules))
}

func TestRun_RuleDirectoryDiscoveryCompletesBeforeBootstrap(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	targetPath := filepath.Join(dir, ".agents", "rules", "02-target.md")

	createTask(t, dir, "test-task", "", "Task content")
	createRule(t, dir, ".agents/rules/01-writer.md", "name: writer", "Writer content")
	createRule(t, dir, ".agents/rules/02-target.md", "name: target", "Original target content")
	createBootstrapScript(t, dir, ".agents/rules/01-writer.md", "#!/bin/sh\nexit 0")

	cc := New(WithSearchPaths("file://" + dir))
	cc.cmdRunner = func(_ *exec.Cmd) error {
		rewriteBootstrapIsolationRule(t, targetPath, "target", "Rewritten target content")
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.NoError(t, err)
	require.Contains(t, bootstrapIsolationRuleContent(t, result.Rules, "target"), "Original target content")
	require.NotContains(t, bootstrapIsolationRuleContent(t, result.Rules, "target"), "Rewritten target content")
}

func TestRun_BootstrapCompletesBeforeNextRuleDirectoryDiscovery(t *testing.T) {
	t.Parallel()

	firstDir := t.TempDir()
	secondDir := t.TempDir()
	targetPath := filepath.Join(secondDir, ".agents", "rules", "01-target.md")

	createTask(t, firstDir, "test-task", "", "Task content")
	createRule(t, firstDir, ".agents/rules/01-writer.md", "name: writer", "Writer content")
	createBootstrapScript(t, firstDir, ".agents/rules/01-writer.md", "#!/bin/sh\nexit 0")
	createRule(t, secondDir, ".agents/rules/01-target.md", "name: target", "Original target content")

	cc := New(
		WithSearchPaths("file://"+firstDir),
		WithSearchPaths("file://"+secondDir),
	)
	cc.cmdRunner = func(_ *exec.Cmd) error {
		rewriteBootstrapIsolationRule(t, targetPath, "target", "Rewritten before later discovery")
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.NoError(t, err)
	require.Contains(t, bootstrapIsolationRuleContent(t, result.Rules, "target"), "Rewritten before later discovery")
}

func TestRun_LenientRuleParseFailurePreservesEarlierRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "test-task", "", "Task content")
	createRule(t, dir, ".agents/rules/01-before.md", "name: before", "Before content")
	createBootstrapScript(t, dir, ".agents/rules/01-before.md", "#!/bin/sh\nexit 0")
	createRule(t, dir, ".agents/rules/02-failing.md", "name: failing\nexpand:\n  - true", "Failing content")
	createRule(t, dir, ".agents/rules/03-after.md", "name: after", "After content")

	var calls []string
	var logs strings.Builder
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	cc := New(
		WithLenientSearchPaths("file://"+dir),
		WithLogger(logger),
	)
	cc.cmdRunner = func(cmd *exec.Cmd) error {
		calls = append(calls, filepath.Base(cmd.Path))
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.NoError(t, err)
	require.Equal(t, []string{"before", "after"}, bootstrapIsolationRuleNames(result.Rules))
	require.Equal(t, []string{"01-before-bootstrap"}, calls)
	require.Contains(t, logs.String(), "skipping rule file after discovery failure")
	require.Contains(t, logs.String(), "02-failing.md")
	require.Contains(t, result.Prompt, "Before content")
	require.NotContains(t, result.Prompt, "Failing content")
	require.Contains(t, result.Prompt, "After content")
}

func TestRun_LenientRuleParseFailureDoesNotBlockLaterLenientSearchPath(t *testing.T) {
	t.Parallel()

	firstDir := t.TempDir()
	secondDir := t.TempDir()

	createTask(t, firstDir, "test-task", "", "Task content")
	createRule(t, firstDir, ".agents/rules/01-before.md", "name: before", "Before content")
	createBootstrapScript(t, firstDir, ".agents/rules/01-before.md", "#!/bin/sh\nexit 0")
	createRule(t, firstDir, ".agents/rules/02-failing.md", "name: failing\nexpand:\n  - true", "Failing content")
	createRule(t, firstDir, ".agents/rules/03-after.md", "name: after", "After content")
	createRule(t, secondDir, ".agents/rules/01-target.md", "name: target", "Target content")
	createBootstrapScript(t, secondDir, ".agents/rules/01-target.md", "#!/bin/sh\nexit 0")

	var calls []string
	var logs strings.Builder
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	cc := New(
		WithLenientSearchPaths("file://"+firstDir),
		WithLenientSearchPaths("file://"+secondDir),
		WithLogger(logger),
	)
	cc.cmdRunner = func(cmd *exec.Cmd) error {
		calls = append(calls, filepath.Base(cmd.Path))
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.NoError(t, err)
	require.Equal(t, []string{"before", "after", "target"}, bootstrapIsolationRuleNames(result.Rules))
	require.Equal(t, []string{"01-before-bootstrap", "01-target-bootstrap"}, calls)
	require.Contains(t, result.Prompt, "Before content")
	require.NotContains(t, result.Prompt, "Failing content")
	require.Contains(t, result.Prompt, "After content")
	require.Contains(t, result.Prompt, "Target content")
	require.Contains(t, logs.String(), "skipping rule file after discovery failure")
	require.Contains(t, logs.String(), "02-failing.md")
}

func TestRun_StrictRuleParseFailureIsFatalBeforeDirectoryBootstrap(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "test-task", "", "Task content")
	createRule(t, dir, ".agents/rules/01-before.md", "name: before", "Before content")
	createBootstrapScript(t, dir, ".agents/rules/01-before.md", "#!/bin/sh\nexit 0")
	createRule(t, dir, ".agents/rules/02-failing.md", "name: failing\nexpand:\n  - true", "Failing content")

	var calls []string
	cc := New(WithSearchPaths("file://" + dir))
	cc.cmdRunner = func(cmd *exec.Cmd) error {
		calls = append(calls, filepath.Base(cmd.Path))
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.Nil(t, result)
	require.Error(t, err)
	requireRuleFileError(t, err, "02-failing.md")
	require.Empty(t, calls)
	require.Empty(t, cc.rules)
}

func TestRun_RuleExpansionUsesDirectoryDiscoverySnapshot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "test-task", "", "Task content")
	createRule(t, dir, ".agents/rules/01-writer.md", "name: writer", "Writer content")
	createBootstrapScript(t, dir, ".agents/rules/01-writer.md", "#!/bin/sh\nexit 0")
	createRule(t, dir, ".agents/rules/02-target.md", "name: target\nexpand: true", "Value: ${value}")

	cc := New(
		WithSearchPaths("file://"+dir),
		WithParams(taskparser.Params{"value": []string{"before-bootstrap"}}),
	)
	cc.cmdRunner = func(_ *exec.Cmd) error {
		cc.params = taskparser.Params{"value": []string{"after-bootstrap"}}
		return nil
	}

	result, err := cc.Run(context.Background(), "test-task")
	require.NoError(t, err)
	require.Contains(t, bootstrapIsolationRuleContent(t, result.Rules, "target"), "before-bootstrap")
	require.NotContains(t, bootstrapIsolationRuleContent(t, result.Rules, "target"), "after-bootstrap")
}
