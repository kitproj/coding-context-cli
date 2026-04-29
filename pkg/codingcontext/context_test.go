package codingcontext

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/selectors"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
)

const cursorSkillName = "cursor-skill"

// Test helper functions for creating fixtures

// buildMarkdownContent wraps content with a YAML frontmatter block when
// frontmatter is non-empty. Returns content unchanged when frontmatter is empty.
func buildMarkdownContent(frontmatter, content string) string {
	if frontmatter == "" {
		return content
	}
	return fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
}

// writeMarkdownFile writes a markdown file (with optional frontmatter) to path,
// creating any missing parent directories.
func writeMarkdownFile(t *testing.T, path, frontmatter, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(buildMarkdownContent(frontmatter, content)), 0o600); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// createTask creates a task file in the .agents/tasks directory.
func createTask(t *testing.T, dir, name, frontmatter, content string) {
	t.Helper()
	writeMarkdownFile(t, filepath.Join(dir, ".agents", "tasks", name+".md"), frontmatter, content)
}

// createRule creates a rule file at relPath within dir.
func createRule(t *testing.T, dir, relPath, frontmatter, content string) {
	t.Helper()
	writeMarkdownFile(t, filepath.Join(dir, relPath), frontmatter, content)
}

// createCommand creates a command file in the .agents/commands directory.
func createCommand(t *testing.T, dir, name, frontmatter, content string) {
	t.Helper()
	writeMarkdownFile(t, filepath.Join(dir, ".agents", "commands", name+".md"), frontmatter, content)
}

// createBootstrapScript creates a bootstrap script for a rule file.
func createBootstrapScript(t *testing.T, dir, rulePath, scriptContent string) {
	t.Helper()

	fullRulePath := filepath.Join(dir, rulePath)
	baseNameWithoutExt := strings.TrimSuffix(fullRulePath, filepath.Ext(fullRulePath))
	bootstrapPath := baseNameWithoutExt + "-bootstrap"

	// Bootstrap scripts are executed directly (support shebangs); require 0755
	// #nosec G306 -- bootstrap scripts require 0755 for direct execution
	if err := os.WriteFile(bootstrapPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("failed to create bootstrap script: %v", err)
	}
}

func createSkill(t *testing.T, dir, subdir, content string) {
	t.Helper()

	skillDir := filepath.Join(dir, subdir)

	if err := os.MkdirAll(skillDir, 0o750); err != nil {
		t.Fatalf("failed to create skill directory: %v", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")

	if err := os.WriteFile(skillPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create skill file: %v", err)
	}
}

// TestNew tests the constructor with various options.
func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		opts  []Option
		check func(t *testing.T, c *Context)
	}{
		{name: "default context", opts: nil, check: checkNewDefault},
		{
			name:  "with params",
			opts:  []Option{WithParams(taskparser.Params{"key1": []string{"value1"}, "key2": []string{"value2"}})},
			check: checkNewWithParams,
		},
		{
			name:  "with selectors",
			opts:  []Option{WithSelectors(selectors.Selectors{"env": {"dev": true, "test": true}})},
			check: checkNewWithSelectors,
		},
		{
			name:  "with manifest URL",
			opts:  []Option{WithManifestURL("https://example.com/manifest.txt")},
			check: checkNewWithManifestURL,
		},
		{
			name:  "with search paths",
			opts:  []Option{WithSearchPaths("/path/one", "/path/two")},
			check: checkNewWithSearchPaths,
		},
		{
			name:  "with custom logger",
			opts:  []Option{WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))},
			check: checkNewWithLogger,
		},
		{name: "with resume mode", opts: []Option{WithResume(true)}, check: checkNewWithResume},
		{name: "with bootstrap disabled", opts: []Option{WithBootstrap(false)}, check: checkNewBootstrapDisabled},
		{
			name:  "resume and bootstrap are independent",
			opts:  []Option{WithResume(true), WithBootstrap(false)},
			check: checkNewResumeAndBootstrapIndependent,
		},
		{name: "with agent", opts: []Option{WithAgent(AgentCursor)}, check: checkNewWithAgent},
		{
			name: "multiple options combined",
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"production"}}),
				WithSelectors(selectors.Selectors{"lang": {"go": true}}),
				WithSearchPaths("/custom/path"),
				WithResume(false),
				WithAgent(AgentCopilot),
			},
			check: checkNewMultipleCombined,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := New(tt.opts...)
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

func checkRunBasicSimpleTask(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "This is a simple task.") {
		t.Errorf("expected task content 'This is a simple task.', got %q", result.Task.Content)
	}

	if result.Tokens <= 0 {
		t.Error("expected positive token count")
	}
}

func checkRunBasicFrontmatter(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Task content here.") {
		t.Errorf("expected task content, got %q", result.Task.Content)
	}

	if result.Task.FrontMatter.Content["priority"] != "high" {
		t.Error("expected priority=high in frontmatter")
	}
}

func checkRunBasicParamSubstitution(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Environment: production") {
		t.Errorf("expected 'Environment: production' in content, got %q", result.Task.Content)
	}

	if !strings.Contains(result.Task.Content, "Feature: auth") {
		t.Errorf("expected 'Feature: auth' in content, got %q", result.Task.Content)
	}
}

func checkRunBasicUnresolvedParam(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "${missing_param}") {
		t.Errorf("expected unresolved parameter to remain as ${missing_param}, got %q", result.Task.Content)
	}
}

func checkRunBasicSelectors(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Task with selectors") {
		t.Errorf("unexpected content: %q", result.Task.Content)
	}
}

func checkRunBasicMultipleParams(t *testing.T, result *Result) {
	t.Helper()

	expected := "User: alice, Email: alice@example.com, Role: admin"
	if !strings.Contains(result.Task.Content, expected) {
		t.Errorf("expected %q in content, got %q", expected, result.Task.Content)
	}
}

func checkNewDefault(t *testing.T, c *Context) {
	t.Helper()

	if c.params == nil {
		t.Error("expected params to be initialized")
	}

	if c.includes == nil {
		t.Error("expected includes to be initialized")
	}

	if c.logger == nil {
		t.Error("expected logger to be initialized")
	}

	if c.cmdRunner == nil {
		t.Error("expected cmdRunner to be initialized")
	}
}

func checkNewWithParams(t *testing.T, c *Context) {
	t.Helper()

	if c.params.Value("key1") != "value1" {
		t.Errorf("expected params[key1]=value1, got %v", c.params.Value("key1"))
	}

	if c.params.Value("key2") != "value2" {
		t.Errorf("expected params[key2]=value2, got %v", c.params.Value("key2"))
	}
}

func checkNewWithSelectors(t *testing.T, c *Context) {
	t.Helper()

	if !c.includes.GetValue("env", "dev") {
		t.Error("expected env=dev selector")
	}

	if !c.includes.GetValue("env", "test") {
		t.Error("expected env=test selector")
	}
}

func checkNewWithManifestURL(t *testing.T, c *Context) {
	t.Helper()

	if c.manifestURL != "https://example.com/manifest.txt" {
		t.Errorf("expected manifestURL to be set, got %v", c.manifestURL)
	}
}

func checkNewWithSearchPaths(t *testing.T, c *Context) {
	t.Helper()

	if len(c.searchPaths) != 2 {
		t.Errorf("expected 2 search paths, got %d", len(c.searchPaths))
	}

	if c.searchPaths[0].Path != "/path/one" {
		t.Errorf("expected first path to be /path/one, got %v", c.searchPaths[0].Path)
	}

	if c.searchPaths[1].Path != "/path/two" {
		t.Errorf("expected second path to be /path/two, got %v", c.searchPaths[1].Path)
	}
}

func checkNewWithLogger(t *testing.T, c *Context) {
	t.Helper()

	if c.logger == nil {
		t.Error("expected logger to be set")
	}
}

func checkNewWithResume(t *testing.T, c *Context) {
	t.Helper()

	if !c.resume {
		t.Error("expected resume to be true")
	}

	if !c.doBootstrap {
		t.Error("expected doBootstrap to be true by default")
	}
}

func checkNewBootstrapDisabled(t *testing.T, c *Context) {
	t.Helper()

	if c.doBootstrap {
		t.Error("expected doBootstrap to be false")
	}
}

func checkNewResumeAndBootstrapIndependent(t *testing.T, c *Context) {
	t.Helper()

	if !c.resume {
		t.Error("expected resume to be true")
	}

	if c.doBootstrap {
		t.Error("expected doBootstrap to be false")
	}
}

func checkNewWithAgent(t *testing.T, c *Context) {
	t.Helper()

	if c.agent != AgentCursor {
		t.Errorf("expected agent to be cursor, got %v", c.agent)
	}
}

func checkNewMultipleCombined(t *testing.T, c *Context) {
	t.Helper()

	if c.params.Value("env") != "production" {
		t.Error("params not set correctly")
	}

	if !c.includes.GetValue("lang", "go") {
		t.Error("selectors not set correctly")
	}

	if len(c.searchPaths) != 1 || c.searchPaths[0].Path != "/custom/path" {
		t.Error("search paths not set correctly")
	}

	if c.resume != false {
		t.Error("resume not set correctly")
	}

	if c.agent != AgentCopilot {
		t.Error("agent not set correctly")
	}
}

// TestContext_Run_Basic tests basic task execution scenarios.
//
//nolint:funlen
func TestContext_Run_Basic(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		opts        []Option
		taskName    string
		wantErr     bool
		errContains string
		check       func(t *testing.T, result *Result)
	}{
		{
			name: "simple task with plain text",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "simple", "", "This is a simple task.")
			},
			taskName: "simple",
			wantErr:  false,
			check:    checkRunBasicSimpleTask,
		},
		{
			name: "task with frontmatter",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "with-frontmatter", "priority: high\nenv: dev", "Task content here.")
			},
			taskName: "with-frontmatter",
			wantErr:  false,
			check:    checkRunBasicFrontmatter,
		},
		{
			name: "task with parameter substitution",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "params-task", "", "Environment: ${env}\nFeature: ${feature}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"production"}, "feature": []string{"auth"}}),
			},
			taskName: "params-task",
			wantErr:  false,
			check:    checkRunBasicParamSubstitution,
		},
		{
			name: "task with unresolved parameter",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "unresolved", "", "Missing: ${missing_param}")
			},
			taskName: "unresolved",
			wantErr:  false,
			check:    checkRunBasicUnresolvedParam,
		},
		{
			name:        "task not found returns error",
			setup:       func(t *testing.T, _ string) { t.Helper() },
			taskName:    "nonexistent",
			wantErr:     true,
			errContains: "task not found",
		},
		{
			name: "task with selectors sets includes",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "selector-task", "selectors:\n  env: production\n  lang: go", "Task with selectors")
			},
			taskName: "selector-task",
			wantErr:  false,
			check:    checkRunBasicSelectors,
		},
		{
			name: "multiple params in same content",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "multi-params", "", "User: ${user}, Email: ${email}, Role: ${role}")
			},
			opts: []Option{
				WithParams(taskparser.Params{
					"user": []string{"alice"}, "email": []string{"alice@example.com"}, "role": []string{"admin"},
				}),
			},
			taskName: "multi-params",
			wantErr:  false,
			check:    checkRunBasicMultipleParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			opts := append([]Option{WithSearchPaths(tmpDir)}, tt.opts...)
			c := New(opts...)

			result, err := c.Run(context.Background(), tt.taskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func checkRulesCount(n int) func(t *testing.T, result *Result) {
	return func(t *testing.T, result *Result) {
		t.Helper()

		if len(result.Rules) != n {
			t.Errorf("expected %d rules, got %d", n, len(result.Rules))
		}
	}
}

func checkRulesFilteredBySelectors(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(result.Rules))
	}

	foundProd := false

	for _, rule := range result.Rules {
		if strings.Contains(rule.Content, "Production rule") {
			foundProd = true

			break
		}
	}

	if !foundProd {
		t.Error("expected to find production rule")
	}
}

func checkRulesParamSubstitution(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result.Rules))
	}

	if !strings.Contains(result.Rules[0].Content, "Project: myapp") {
		t.Errorf("expected parameter substitution in rule, got %q", result.Rules[0].Content)
	}
}

func checkRulesTokenCounting(t *testing.T, result *Result) {
	t.Helper()

	if result.Tokens <= 0 {
		t.Error("expected positive total token count")
	}

	totalRuleTokens := 0

	for _, rule := range result.Rules {
		if rule.Tokens <= 0 {
			t.Error("expected positive token count for each rule")
		}

		totalRuleTokens += rule.Tokens
	}

	if result.Task.Tokens <= 0 {
		t.Error("expected positive token count for task")
	}

	expectedTotal := totalRuleTokens + result.Task.Tokens
	if result.Tokens != expectedTotal {
		t.Errorf("expected total tokens %d, got %d", expectedTotal, result.Tokens)
	}
}

// TestContext_Run_Rules tests rule discovery and filtering.
//
//nolint:funlen,maintidx
func TestContext_Run_Rules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		opts     []Option
		taskName string
		wantErr  bool
		check    func(t *testing.T, result *Result)
	}{
		{
			name: "discover rules in standard paths",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "task1", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule 1 content")
				createRule(t, dir, ".cursor/rules/rule2.md", "", "Rule 2 content")
			},
			taskName: "task1",
			wantErr:  false,
			check:    checkRulesCount(2),
		},
		{
			name: "filter rules by selectors from task frontmatter",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "filtered-task", "selectors:\n  env: production", "Task with selectors")
				createRule(t, dir, ".agents/rules/prod-rule.md", "env: production", "Production rule")
				createRule(t, dir, ".agents/rules/dev-rule.md", "env: development", "Development rule")
				createRule(t, dir, ".agents/rules/no-env.md", "", "No env specified")
			},
			taskName: "filtered-task",
			wantErr:  false,
			check:    checkRulesFilteredBySelectors,
		},
		{
			name: "rules with parameter substitution",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "param-task", "", "Task")
				createRule(t, dir, ".agents/rules/param-rule.md", "", "Project: ${project}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"project": []string{"myapp"}}),
			},
			taskName: "param-task",
			wantErr:  false,
			check:    checkRulesParamSubstitution,
		},
		{
			name: "bootstrap disabled skips rule discovery",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "bootstrap-task", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule content")
			},
			opts: []Option{
				WithBootstrap(false),
			},
			taskName: "bootstrap-task",
			wantErr:  false,
			check:    checkRulesCount(0),
		},
		{
			name: "resume mode does not skip rule discovery",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "resume-task", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule content")
			},
			opts: []Option{
				WithResume(true),
			},
			taskName: "resume-task",
			wantErr:  false,
			check:    checkRulesCount(1),
		},
		{
			name: "bootstrap script executed for rules",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "bootstrap-task", "", "Task")
				createRule(t, dir, ".agents/rules/rule-with-bootstrap.md", "", "Rule content")
				createBootstrapScript(t, dir, ".agents/rules/rule-with-bootstrap.md", "#!/bin/sh\necho 'bootstrapped'")
			},
			taskName: "bootstrap-task",
			wantErr:  false,
			check:    checkRulesCount(1),
		},
		{
			name: "bootstrap disabled skips bootstrap scripts",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "no-bootstrap", "", "Task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule")
				createBootstrapScript(t, dir, ".agents/rules/rule1.md", "#!/bin/sh\nexit 1")
			},
			opts: []Option{
				WithBootstrap(false),
			},
			taskName: "no-bootstrap",
			wantErr:  false,
			check:    checkRulesCount(0),
		},
		{
			name: "bootstrap from frontmatter is preferred",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "frontmatter-bootstrap", "", "Task")
				// Create rule with bootstrap in frontmatter that writes a marker file
				createRule(t, dir, ".agents/rules/rule-with-frontmatter.md",
					"bootstrap: |\n  #!/bin/sh\n  echo 'frontmatter' > "+filepath.Join(dir, "bootstrap-ran.txt")+"\n",
					"Rule content")
			},
			taskName: "frontmatter-bootstrap",
			wantErr:  false,
			check:    checkRulesCount(1),
		},
		{
			name: "bootstrap from frontmatter preferred over file",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "frontmatter-priority", "", "Task")
				// Create rule with BOTH frontmatter and file bootstrap
				// Frontmatter writes "frontmatter", file writes "file"
				markerPath := filepath.Join(dir, "bootstrap-marker.txt")
				createRule(t, dir, ".agents/rules/priority-rule.md",
					"bootstrap: |\n  #!/bin/sh\n  echo 'frontmatter' > "+markerPath+"\n",
					"Rule content")
				// Also create a file-based bootstrap (should be ignored)
				createBootstrapScript(t, dir, ".agents/rules/priority-rule.md",
					"#!/bin/sh\necho 'file' > "+markerPath)
			},
			taskName: "frontmatter-priority",
			wantErr:  false,
			check:    checkRulesCount(1),
		},
		{
			name: "bootstrap from file when frontmatter empty",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "file-fallback", "", "Task")
				// Create rule WITHOUT frontmatter bootstrap
				markerPath := filepath.Join(dir, "bootstrap-marker.txt")
				createRule(t, dir, ".agents/rules/fallback-rule.md", "", "Rule content")
				// Create file-based bootstrap (should be used)
				createBootstrapScript(t, dir, ".agents/rules/fallback-rule.md",
					"#!/bin/sh\necho 'file' > "+markerPath)
			},
			taskName: "file-fallback",
			wantErr:  false,
			check:    checkRulesCount(1),
		},
		{
			name: "agent option collects all rules",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "agent-task", "", "Task")
				createRule(t, dir, ".agents/rules/generic.md", "", "Generic rule")
				createRule(t, dir, ".cursor/rules/cursor-rule.md", "", "Cursor rule")
				createRule(t, dir, ".github/agents/copilot-rule.md", "", "Copilot rule")
			},
			opts: []Option{
				WithAgent(AgentCursor),
			},
			taskName: "agent-task",
			wantErr:  false,
			check:    checkRulesCount(3),
		},
		{
			name: "task frontmatter agent overrides option",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "override-task", "agent: copilot", "Task")
				createRule(t, dir, ".cursor/rules/cursor-rule.md", "", "Cursor rule")
				createRule(t, dir, ".github/agents/copilot-rule.md", "", "Copilot rule")
			},
			opts: []Option{
				WithAgent(AgentCursor), // This should be overridden by task frontmatter
			},
			taskName: "override-task",
			wantErr:  false,
			check:    checkRulesCount(2),
		},
		{
			name: "multiple selector values with OR logic",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "multi-selector", "selectors:\n  env:\n    - dev\n    - test", "Task")
				createRule(t, dir, ".agents/rules/dev-rule.md", "env: dev", "Dev rule")
				createRule(t, dir, ".agents/rules/test-rule.md", "env: test", "Test rule")
				createRule(t, dir, ".agents/rules/prod-rule.md", "env: prod", "Prod rule")
			},
			taskName: "multi-selector",
			wantErr:  false,
			check:    checkRulesCount(2),
		},
		{
			name: "CLI selectors combined with task selectors use OR logic",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "or-task", "selectors:\n  env: production", "Task with env=production")
				createRule(t, dir, ".agents/rules/prod-rule.md", "env: production", "Production rule")
				createRule(t, dir, ".agents/rules/dev-rule.md", "env: development", "Development rule")
				createRule(t, dir, ".agents/rules/test-rule.md", "env: test", "Test rule")
			},
			opts: []Option{
				WithSelectors(selectors.Selectors{"env": {"development": true}}), // CLI selector for development
			},
			taskName: "or-task",
			wantErr:  false,
			check:    checkRulesCount(2), // prod + dev; test excluded
		},
		{
			name: "CLI selectors combined with array task selectors use OR logic",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "array-or", "selectors:\n  env:\n    - production\n    - staging", "Task with array selectors")
				createRule(t, dir, ".agents/rules/prod-rule.md", "env: production", "Production rule")
				createRule(t, dir, ".agents/rules/staging-rule.md", "env: staging", "Staging rule")
				createRule(t, dir, ".agents/rules/dev-rule.md", "env: development", "Development rule")
				createRule(t, dir, ".agents/rules/test-rule.md", "env: test", "Test rule")
			},
			opts: []Option{
				WithSelectors(selectors.Selectors{"env": {"development": true}}), // CLI adds development
			},
			taskName: "array-or",
			wantErr:  false,
			check:    checkRulesCount(3), // prod + staging + dev; test excluded
		},
		{
			name: "token counting for rules",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "token-task", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "This is rule 1 content")
				createRule(t, dir, ".agents/rules/rule2.md", "", "This is rule 2 content")
			},
			taskName: "token-task",
			wantErr:  false,
			check:    checkRulesTokenCounting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			opts := append([]Option{WithSearchPaths(tmpDir)}, tt.opts...)
			c := New(opts...)

			result, err := c.Run(context.Background(), tt.taskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func checkTaskContains(s string) func(t *testing.T, result *Result) {
	return func(t *testing.T, result *Result) {
		t.Helper()

		if !strings.Contains(result.Task.Content, s) {
			t.Errorf("expected %q in task content, got %q", s, result.Task.Content)
		}
	}
}

func checkTaskNotEmpty(t *testing.T, result *Result) {
	t.Helper()

	if strings.TrimSpace(result.Task.Content) == "" {
		t.Errorf("expected non-empty content, got %q", result.Task.Content)
	}
}

func checkCommandsSingleRef(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Before command") {
		t.Error("expected task content before command")
	}

	if !strings.Contains(result.Task.Content, "Hello, World!") {
		t.Error("expected command content to be substituted")
	}

	if !strings.Contains(result.Task.Content, "After command") {
		t.Error("expected task content after command")
	}
}

func checkCommandsMixedText(t *testing.T, result *Result) {
	t.Helper()

	content := result.Task.Content
	if !strings.Contains(content, "# Title") {
		t.Error("expected title text")
	}

	if !strings.Contains(content, "Middle text") {
		t.Error("expected middle text")
	}

	if !strings.Contains(content, "End text") {
		t.Error("expected end text")
	}
}

func checkCommandsSelectorsFilterRules(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(result.Rules))
	}

	foundPostgres := false

	for _, rule := range result.Rules {
		if strings.Contains(rule.Content, "PostgreSQL rule") {
			foundPostgres = true

			break
		}
	}

	if !foundPostgres {
		t.Error("expected to find PostgreSQL rule")
	}
}

// TestContext_Run_Commands tests command substitution in tasks.
//
//nolint:funlen
func TestContext_Run_Commands(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		opts        []Option
		taskName    string
		wantErr     bool
		errContains string
		check       func(t *testing.T, result *Result)
	}{
		{
			name: "task with single command reference",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "with-command", "", "Before command\n/greet\nAfter command")
				createCommand(t, dir, "greet", "", "Hello, World!")
			},
			taskName: "with-command",
			wantErr:  false,
			check:    checkCommandsSingleRef,
		},
		{
			name: "command with parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "cmd-with-params", "", "/greet name=\"Alice\"")
				createCommand(t, dir, "greet", "", "Hello, ${name}!")
			},
			taskName: "cmd-with-params",
			wantErr:  false,
			check:    checkTaskContains("Hello, Alice!"),
		},
		{
			name: "command with context parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "ctx-params", "", "/deploy")
				createCommand(t, dir, "deploy", "", "Deploying to ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"staging"}}),
			},
			taskName: "ctx-params",
			wantErr:  false,
			check:    checkTaskContains("Deploying to staging"),
		},
		{
			name: "multiple commands in task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "multi-cmd", "", "/intro\n\n/body\n\n/outro\n")
				createCommand(t, dir, "intro", "", "Introduction")
				createCommand(t, dir, "body", "", "Main content")
				createCommand(t, dir, "outro", "", "Conclusion")
			},
			taskName: "multi-cmd",
			wantErr:  false,
			check:    checkTaskNotEmpty,
		},
		{
			name: "command not found passes through as-is",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "missing-cmd", "", "/nonexistent")
			},
			taskName: "missing-cmd",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				t.Helper()
				if !strings.Contains(result.Prompt, "/nonexistent") {
					t.Errorf("expected pass-through of /nonexistent, got %q", result.Prompt)
				}
			},
		},
		{
			name: "command parameter overrides context parameter",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "override-param", "", "/msg value=\"specific\"")
				createCommand(t, dir, "msg", "", "Value: ${value}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"value": []string{"general"}}),
			},
			taskName: "override-param",
			wantErr:  false,
			check:    checkTaskContains("Value: specific"),
		},
		{
			name: "command with multiple parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "multi-params", "", "/info name=\"Bob\" age=\"30\" role=\"developer\"")
				createCommand(t, dir, "info", "", "Name: ${name}, Age: ${age}, Role: ${role}")
			},
			taskName: "multi-params",
			wantErr:  false,
			check:    checkTaskContains("Name: Bob, Age: 30, Role: developer"),
		},
		{
			name: "mixed text and commands",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "mixed", "", "# Title\n\n/section1\n\nMiddle text\n\n/section2\n\nEnd text")
				createCommand(t, dir, "section1", "", "Section 1 content")
				createCommand(t, dir, "section2", "", "Section 2 content")
			},
			taskName: "mixed",
			wantErr:  false,
			check:    checkCommandsMixedText,
		},
		{
			name: "command with selectors filters rules",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Task uses a command that has selectors
				createTask(t, dir, "task-with-cmd", "", "/setup-db")
				// Command has selectors that should be applied to rule filtering
				createCommand(t, dir, "setup-db", "selectors:\n  database: postgres", "Setting up database...")
				// Rules with different database values
				createRule(t, dir, ".agents/rules/postgres-rule.md", "database: postgres", "PostgreSQL rule")
				createRule(t, dir, ".agents/rules/mysql-rule.md", "database: mysql", "MySQL rule")
				createRule(t, dir, ".agents/rules/generic-rule.md", "", "Generic rule")
			},
			taskName: "task-with-cmd",
			wantErr:  false,
			check:    checkCommandsSelectorsFilterRules,
		},
		{
			name: "command selectors combine with task selectors",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Task has its own selectors
				createTask(t, dir, "combined-selectors", "selectors:\n  env: production", "/enable-feature")
				// Command also has selectors
				createCommand(t, dir, "enable-feature", "selectors:\n  feature: auth", "Enabling authentication...")
				// Rules with different combinations
				createRule(t, dir, ".agents/rules/prod-auth-rule.md", "env: production\nfeature: auth", "Production auth rule")
				createRule(t, dir, ".agents/rules/prod-rule.md", "env: production", "Production rule")
				createRule(t, dir, ".agents/rules/auth-rule.md", "feature: auth", "Auth rule")
				createRule(t, dir, ".agents/rules/dev-rule.md", "env: development", "Development rule")
			},
			taskName: "combined-selectors",
			wantErr:  false,
			check:    checkRulesCount(3), // prod-auth + prod + auth; dev excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			opts := append([]Option{WithSearchPaths(tmpDir)}, tt.opts...)
			c := New(opts...)

			result, err := c.Run(context.Background(), tt.taskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func checkIntegrationFullWorkflow(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Deploy myservice") {
		t.Error("expected app param substitution")
	}

	if !strings.Contains(result.Task.Content, "Deploy to production") {
		t.Error("expected command with param substitution")
	}

	if len(result.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(result.Rules))
	}

	if result.Tokens <= 0 {
		t.Error("expected positive token count")
	}
}

func checkIntegrationComplexTask(t *testing.T, result *Result) {
	t.Helper()

	content := result.Task.Content
	if !strings.Contains(content, "# Project Setup") {
		t.Error("expected markdown header")
	}

	if strings.TrimSpace(content) == "" {
		t.Error("expected non-empty content")
	}
}

func checkIntegrationBootstrapDisabled(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Continue this task") {
		t.Errorf("unexpected task content: %q", result.Task.Content)
	}

	if len(result.Rules) != 0 {
		t.Errorf("expected 0 rules when bootstrap is disabled, got %d", len(result.Rules))
	}
}

// TestContext_Run_Integration tests end-to-end integration scenarios.
//
//nolint:funlen
func TestContext_Run_Integration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		opts     []Option
		taskName string
		wantErr  bool
		check    func(t *testing.T, result *Result)
	}{
		{
			name: "full workflow with task, rules, commands, and parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "fullworkflow", "selectors:\n  env: production\n  lang: go", "Deploy ${app}\n/deploy-steps")
				createCommand(t, dir, "deploy-steps", "", "1. Build\n2. Test\n3. Deploy to ${env}")
				createRule(t, dir, ".agents/rules/prod.md", "env: production", "Production guidelines")
				createRule(t, dir, ".agents/rules/go.md", "lang: go", "Go best practices")
				createRule(t, dir, ".agents/rules/dev.md", "env: development", "Dev only rule")
			},
			opts: []Option{
				WithParams(taskparser.Params{"app": []string{"myservice"}, "env": []string{"production"}}),
			},
			taskName: "fullworkflow",
			wantErr:  false,
			check:    checkIntegrationFullWorkflow,
		},
		{
			name: "complex task with multiple slash commands and mixed content",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "complex", "",
					"# Project Setup\n\n/intro\n\n## Steps\n\n/step1\n\n/step2\n\n## Conclusion\n\n/outro\n")
				createCommand(t, dir, "intro", "", "Welcome to the project")
				createCommand(t, dir, "step1", "", "First, initialize the repository")
				createCommand(t, dir, "step2", "", "Then, configure the settings")
				createCommand(t, dir, "outro", "", "You're all set!")
			},
			taskName: "complex",
			wantErr:  false,
			check:    checkIntegrationComplexTask,
		},
		{
			name: "bootstrap disabled workflow skips rules but includes task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "bootstrap", "", "Continue this task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Should be skipped")
				createBootstrapScript(t, dir, ".agents/rules/rule1.md", "#!/bin/sh\necho 'should not run'")
			},
			opts: []Option{
				WithBootstrap(false),
			},
			taskName: "bootstrap",
			wantErr:  false,
			check:    checkIntegrationBootstrapDisabled,
		},
		{
			name: "agent-specific workflow",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "for-cursor", "agent: cursor", "Task for Cursor")
				createRule(t, dir, ".cursor/rules/cursor.md", "", "Cursor-specific")
				createRule(t, dir, ".agents/rules/general.md", "", "General rule")
				createRule(t, dir, ".github/agents/copilot.md", "", "Copilot rule")
			},
			taskName: "for-cursor",
			wantErr:  false,
			check:    checkRulesCount(3),
		},
		{
			name: "multiple search paths",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create first directory with task and rule
				createTask(t, dir, "multi-path", "", "Multi-path task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule from first path")

				// Create second directory with additional rule
				secondDir := filepath.Join(dir, "second")
				if err := os.MkdirAll(secondDir, 0o750); err != nil {
					t.Fatalf("failed to create second dir: %v", err)
				}

				createRule(t, secondDir, ".agents/rules/rule2.md", "", "Rule from second path")
			},
			opts: []Option{
				// Second path will be added via setup
			},
			taskName: "multi-path",
			wantErr:  false,
			check:    checkRulesCount(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			opts := append([]Option{WithSearchPaths(tmpDir)}, tt.opts...)
			c := New(opts...)

			result, err := c.Run(context.Background(), tt.taskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestContext_Run_Errors tests error scenarios.
func TestContext_Run_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		opts        []Option
		taskName    string
		wantErr     bool
		errContains string
		check       func(t *testing.T, result *Result)
	}{
		{
			name: "command not found in task passes through as-is",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "bad-cmd", "", "/missing-command\n")
			},
			taskName: "bad-cmd",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				t.Helper()
				if !strings.Contains(result.Prompt, "/missing-command") {
					t.Errorf("expected pass-through of /missing-command, got %q", result.Prompt)
				}
			},
		},
		{
			name: "invalid agent in task frontmatter",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "bad-agent", "agent: invalidagent", "Task content")
			},
			taskName:    "bad-agent",
			wantErr:     true,
			errContains: "unknown agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			opts := append([]Option{WithSearchPaths(tmpDir)}, tt.opts...)
			c := New(opts...)

			result, err := c.Run(context.Background(), tt.taskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				if result != nil {
					t.Error("expected nil result on error")
				}

				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func checkOneRuleContains(s string) func(t *testing.T, result *Result) {
	return func(t *testing.T, result *Result) {
		t.Helper()

		if len(result.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(result.Rules))
		}

		if !strings.Contains(result.Rules[0].Content, s) {
			t.Errorf("expected %q in rule, got %q", s, result.Rules[0].Content)
		}
	}
}

func checkExpandIssueAndTitle(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Issue: 123") {
		t.Errorf("expected 'Issue: 123', got %q", result.Task.Content)
	}

	if !strings.Contains(result.Task.Content, "Title: Bug fix") {
		t.Errorf("expected 'Title: Bug fix', got %q", result.Task.Content)
	}
}

// checkMixedNoExpandCommandExpand: task has expand:false (text unexpanded), command has expand:true (expanded).
func checkMixedNoExpandCommandExpand(t *testing.T, result *Result) {
	t.Helper()

	content := result.Task.Content
	if !strings.Contains(content, "Task ${task_var}") {
		t.Errorf("expected task param unexpanded (expand:false), got %q", content)
	}

	if !strings.Contains(content, "Command cmd_value") {
		t.Errorf("expected command param expanded (expand:true), got %q", content)
	}
}

// checkMixedTaskExpandNoCommand: task has expand:true (text expanded), command has expand:false (unexpanded).
func checkMixedTaskExpandNoCommand(t *testing.T, result *Result) {
	t.Helper()

	content := result.Task.Content
	if !strings.Contains(content, "Task task_value") {
		t.Errorf("expected task param expanded (expand:true), got %q", content)
	}

	if !strings.Contains(content, "Command ${cmd_var}") {
		t.Errorf("expected command param unexpanded (expand:false), got %q", content)
	}
}

func checkExpandInlineNoExpand(t *testing.T, result *Result) {
	t.Helper()

	// expand:false means inline params and global params are NOT substituted
	if !strings.Contains(result.Task.Content, "${name}") {
		t.Errorf("expected '${name}' unexpanded in output, got %q", result.Task.Content)
	}

	if !strings.Contains(result.Task.Content, "${id}") {
		t.Errorf("expected '${id}' unexpanded in output, got %q", result.Task.Content)
	}
}

func checkExpandMultipleRules(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(result.Rules))
	}

	for _, rule := range result.Rules {
		switch {
		case strings.Contains(rule.Content, "Rule1:"):
			// Rule1 has expand:false — parameter should NOT be substituted
			if !strings.Contains(rule.Content, "Rule1: ${var1}") {
				t.Errorf("expected 'Rule1: ${var1}' (unexpanded), got %q", rule.Content)
			}
		case strings.Contains(rule.Content, "Rule2:"):
			if !strings.Contains(rule.Content, "Rule2: val2") {
				t.Errorf("expected 'Rule2: val2', got %q", rule.Content)
			}
		case strings.Contains(rule.Content, "Rule3:"):
			if !strings.Contains(rule.Content, "Rule3: val3") {
				t.Errorf("expected 'Rule3: val3', got %q", rule.Content)
			}
		}
	}
}

// TestContext_Run_ExpandParams tests parameter expansion opt-out functionality.
//
//nolint:funlen
func TestContext_Run_ExpandParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		opts        []Option
		taskName    string
		wantErr     bool
		errContains string
		check       func(t *testing.T, result *Result)
	}{
		{
			name: "task with expand: false preserves parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "no-expand", "expand: false", "Issue: ${issue_number}\nTitle: ${issue_title}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"issue_number": []string{"123"}, "issue_title": []string{"Bug fix"}}),
			},
			taskName: "no-expand",
			wantErr:  false,
			check:    checkTaskContains("${issue_number}"), // expand:false means no substitution
		},
		{
			name: "task with expand: true expands parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "expand", "expand: true", "Issue: ${issue_number}\nTitle: ${issue_title}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"issue_number": []string{"123"}, "issue_title": []string{"Bug fix"}}),
			},
			taskName: "expand",
			wantErr:  false,
			check:    checkExpandIssueAndTitle,
		},
		{
			name: "task without expand defaults to expanding",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "default", "", "Env: ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"production"}}),
			},
			taskName: "default",
			wantErr:  false,
			check:    checkTaskContains("Env: production"),
		},
		{
			name: "command with expand: false preserves parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "cmd-no-expand", "", "/deploy")
				createCommand(t, dir, "deploy", "expand: false", "Deploying to ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"staging"}}),
			},
			taskName: "cmd-no-expand",
			wantErr:  false,
			check:    checkTaskContains("Deploying to ${env}"), // expand:false means no substitution
		},
		{
			name: "command with expand: true expands parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "cmd-expand", "", "/deploy")
				createCommand(t, dir, "deploy", "expand: true", "Deploying to ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"staging"}}),
			},
			taskName: "cmd-expand",
			wantErr:  false,
			check:    checkTaskContains("Deploying to staging"),
		},
		{
			name: "command without expand defaults to expanding",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "cmd-default", "", "/info")
				createCommand(t, dir, "info", "", "Project: ${project}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"project": []string{"myapp"}}),
			},
			taskName: "cmd-default",
			wantErr:  false,
			check:    checkTaskContains("Project: myapp"),
		},
		{
			name: "rule with expand: false preserves parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "rule-no-expand", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "expand: false", "Version: ${version}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"version": []string{"1.0.0"}}),
			},
			taskName: "rule-no-expand",
			wantErr:  false,
			check:    checkOneRuleContains("Version: ${version}"), // expand:false means no substitution
		},
		{
			name: "rule with expand: true expands parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "rule-expand", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "expand: true", "Version: ${version}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"version": []string{"1.0.0"}}),
			},
			taskName: "rule-expand",
			wantErr:  false,
			check:    checkOneRuleContains("Version: 1.0.0"),
		},
		{
			name: "rule without expand defaults to expanding",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "rule-default", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "App: ${app}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"app": []string{"service"}}),
			},
			taskName: "rule-default",
			wantErr:  false,
			check:    checkOneRuleContains("App: service"),
		},
		{
			name: "mixed: task no expand, command with expand",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "mixed1", "expand: false", "Task ${task_var}\n/cmd")
				createCommand(t, dir, "cmd", "expand: true", "Command ${cmd_var}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"task_var": []string{"task_value"}, "cmd_var": []string{"cmd_value"}}),
			},
			taskName: "mixed1",
			wantErr:  false,
			check:    checkMixedNoExpandCommandExpand, // task text unexpanded, command expanded
		},
		{
			name: "mixed: task with expand, command no expand",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "mixed2", "expand: true", "Task ${task_var}\n/cmd")
				createCommand(t, dir, "cmd", "expand: false", "Command ${cmd_var}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"task_var": []string{"task_value"}, "cmd_var": []string{"cmd_value"}}),
			},
			taskName: "mixed2",
			wantErr:  false,
			check:    checkMixedTaskExpandNoCommand, // task text expanded, command unexpanded
		},
		{
			name: "command with inline parameters and expand: false",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "inline-no-expand", "", "/greet name=\"Alice\"")
				createCommand(t, dir, "greet", "expand: false", "Hello, ${name}! Your ID: ${id}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"id": []string{"123"}}),
			},
			taskName: "inline-no-expand",
			wantErr:  false,
			check:    checkExpandInlineNoExpand, // expand:false means all ${...} stay unexpanded
		},
		{
			name: "multiple rules with different expand settings",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "multi-rules", "", "Task")
				createRule(t, dir, ".agents/rules/rule1.md", "expand: false", "Rule1: ${var1}")
				createRule(t, dir, ".agents/rules/rule2.md", "expand: true", "Rule2: ${var2}")
				createRule(t, dir, ".agents/rules/rule3.md", "", "Rule3: ${var3}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"var1": []string{"val1"}, "var2": []string{"val2"}, "var3": []string{"val3"}}),
			},
			taskName: "multi-rules",
			wantErr:  false,
			check:    checkExpandMultipleRules,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			opts := append([]Option{WithSearchPaths(tmpDir)}, tt.opts...)
			c := New(opts...)

			result, err := c.Run(context.Background(), tt.taskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestUserPrompt tests the user_prompt parameter functionality.
func checkUserPromptSimple(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Task content") {
		t.Error("expected task content to contain 'Task content'")
	}

	if !strings.Contains(result.Task.Content, "User prompt content") {
		t.Error("expected task content to contain 'User prompt content'")
	}

	taskIdx := strings.Index(result.Task.Content, "Task content")

	userIdx := strings.Index(result.Task.Content, "User prompt content")
	if taskIdx >= userIdx {
		t.Error("expected user_prompt to come after task content")
	}
}

func checkUserPromptWithCommand(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Task content") {
		t.Error("expected task content to contain 'Task content'")
	}

	if !strings.Contains(result.Task.Content, "User says:") {
		t.Error("expected task content to contain 'User says: '")
	}

	if !strings.Contains(result.Task.Content, "Hello from command!") {
		t.Error("expected slash command in user_prompt to be expanded")
	}
}

func checkUserPromptWithComplexCommand(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Please fix:") {
		t.Error("expected task content to contain 'Please fix: '")
	}

	if !strings.Contains(result.Task.Content, "Issue 456: Fix bug") {
		t.Error("expected slash command to be expanded with parameter substitution")
	}
}

func checkUserPromptUnchanged(t *testing.T, result *Result) {
	t.Helper()

	if result.Task.Content != "Task content\n" {
		t.Errorf("expected task content to be unchanged, got %q", result.Task.Content)
	}
}

func checkUserPromptMultipleCommands(t *testing.T, result *Result) {
	t.Helper()

	if !strings.Contains(result.Task.Content, "Command 1") {
		t.Error("expected first slash command to be expanded")
	}

	if !strings.Contains(result.Task.Content, "Command 2") {
		t.Error("expected second slash command to be expanded")
	}
}

func checkUserPromptBothParsed(t *testing.T, result *Result) {
	t.Helper()

	checks := []struct{ substr, msg string }{
		{"Task prompt with text", "expected task content to contain 'Task prompt with text'"},
		{"More task text", "expected task content to contain 'More task text'"},
		{"User prompt with text", "expected task content to contain 'User prompt with text'"},
		{"More user text", "expected task content to contain 'More user text'"},
		{"Task command output value1", "expected task command to be expanded with param1=value1"},
		{"User command output value2", "expected user command to be expanded with param2=value2"},
	}
	for _, c := range checks {
		if !strings.Contains(result.Task.Content, c.substr) {
			t.Error(c.msg)
		}
	}
}

//nolint:funlen
func TestUserPrompt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		opts        []Option
		taskName    string
		wantErr     bool
		errContains string
		check       func(t *testing.T, result *Result)
	}{
		{
			name: "simple user_prompt appended to task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "simple", "", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt("User prompt content"),
			},
			taskName: "simple",
			wantErr:  false,
			check:    checkUserPromptSimple,
		},
		{
			name: "user_prompt with slash command",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "with-command", "", "Task content\n")
				createCommand(t, dir, "greet", "", "Hello from command!")
			},
			opts: []Option{
				WithUserPrompt("User says:\n/greet\n"),
			},
			taskName: "with-command",
			wantErr:  false,
			check:    checkUserPromptWithCommand,
		},
		{
			name: "user_prompt with parameter substitution",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "with-params", "", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt("Issue: ${issue_number}"),
				WithParams(taskparser.Params{
					"issue_number": []string{"123"},
				}),
			},
			taskName: "with-params",
			wantErr:  false,
			check:    checkTaskContains("Issue: 123"),
		},
		{
			name: "user_prompt with slash command and parameters",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "complex", "", "Task content\n")
				createCommand(t, dir, "issue-info", "", "Issue ${issue_number}: ${issue_title}")
			},
			opts: []Option{
				WithUserPrompt("Please fix:\n/issue-info\n"),
				WithParams(taskparser.Params{
					"issue_number": []string{"456"},
					"issue_title":  []string{"Fix bug"},
				}),
			},
			taskName: "complex",
			wantErr:  false,
			check:    checkUserPromptWithComplexCommand,
		},
		{
			name: "empty user_prompt should not affect task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "empty", "", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt(""),
			},
			taskName: "empty",
			wantErr:  false,
			check:    checkUserPromptUnchanged,
		},
		{
			name: "no user_prompt parameter should not affect task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "no-prompt", "", "Task content\n")
			},
			opts:     []Option{},
			taskName: "no-prompt",
			wantErr:  false,
			check:    checkUserPromptUnchanged,
		},
		{
			name: "user_prompt with multiple slash commands",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "multi", "", "Task content\n")
				createCommand(t, dir, "cmd1", "", "Command 1")
				createCommand(t, dir, "cmd2", "", "Command 2")
			},
			opts: []Option{
				WithUserPrompt("/cmd1\n/cmd2\n"),
			},
			taskName: "multi",
			wantErr:  false,
			check:    checkUserPromptMultipleCommands,
		},
		{
			name: "user_prompt respects task expand setting",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "no-expand", "expand: false", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt("Issue ${issue_number}"),
				WithParams(taskparser.Params{
					"issue_number": []string{"789"},
				}),
			},
			taskName: "no-expand",
			wantErr:  false,
			check:    checkTaskContains("${issue_number}"), // expand:false applies to user_prompt too
		},
		{
			name: "user_prompt with invalid slash command passes through as-is",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "invalid", "", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt("/nonexistent-command\n"),
			},
			taskName: "invalid",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				t.Helper()
				if !strings.Contains(result.Prompt, "/nonexistent-command") {
					t.Errorf("expected pass-through of /nonexistent-command, got %q", result.Prompt)
				}
			},
		},
		{
			name: "both task prompt and user prompt parse correctly",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Task has text and slash command
				createTask(t, dir, "parse-test", "", "Task prompt with text\n/task-command arg1\nMore task text\n")
				createCommand(t, dir, "task-command", "", "Task command output ${param1}")
				createCommand(t, dir, "user-command", "", "User command output ${param2}")
			},
			opts: []Option{
				WithUserPrompt("User prompt with text\n/user-command arg2\nMore user text"),
				WithParams(taskparser.Params{
					"param1": []string{"value1"},
					"param2": []string{"value2"},
				}),
			},
			taskName: "parse-test",
			wantErr:  false,
			check:    checkUserPromptBothParsed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			tt.setup(t, tmpDir)

			allOpts := append([]Option{
				WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
				WithSearchPaths("file://" + tmpDir),
			}, tt.opts...)

			c := New(allOpts...)

			result, err := c.Run(context.Background(), tt.taskName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestIsLocalPath tests the isLocalPath helper function.
func TestIsLocalPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "file:// protocol",
			path:     "file:///path/to/local",
			expected: true,
		},
		{
			name:     "absolute path",
			path:     "/path/to/local",
			expected: true,
		},
		{
			name:     "relative path - ./",
			path:     "./relative/path",
			expected: true,
		},
		{
			name:     "relative path - ../",
			path:     "../relative/path",
			expected: true,
		},
		{
			name:     "relative path - no prefix",
			path:     "relative/path",
			expected: true,
		},
		{
			name:     "git protocol",
			path:     "git::https://github.com/user/repo.git",
			expected: false,
		},
		{
			name:     "https protocol",
			path:     "https://example.com/file.tar.gz",
			expected: false,
		},
		{
			name:     "http protocol",
			path:     "http://example.com/file.tar.gz",
			expected: false,
		},
		{
			name:     "s3 protocol",
			path:     "s3::https://s3.amazonaws.com/bucket/key",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isLocalPath(tt.path)
			if result != tt.expected {
				t.Errorf("isLocalPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestNormalizeLocalPath tests the normalizeLocalPath helper function.
func TestNormalizeLocalPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "file:// protocol - absolute path",
			path:     "file:///path/to/local",
			expected: "/path/to/local",
		},
		{
			name:     "file:// protocol - relative path",
			path:     "file://./relative/path",
			expected: "./relative/path",
		},
		{
			name:     "absolute path without protocol",
			path:     "/path/to/local",
			expected: "/path/to/local",
		},
		{
			name:     "relative path - ./",
			path:     "./relative/path",
			expected: "./relative/path",
		},
		{
			name:     "relative path - ../",
			path:     "../relative/path",
			expected: "../relative/path",
		},
		{
			name:     "relative path - no prefix",
			path:     "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := normalizeLocalPath(tt.path)
			if result != tt.expected {
				t.Errorf("normalizeLocalPath(%q) = %q, expected %q", tt.path, result, tt.expected)
			}
		})
	}
}

// TestLogsParametersAndSelectors verifies that parameters and selectors are logged
// exactly once after the task is found (which may add selectors from task frontmatter).
func TestLogsParametersAndSelectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		params          taskparser.Params
		selectors       selectors.Selectors
		resume          bool
		expectParamsLog bool
		expectSelectors bool // Always true since task_name is added
	}{
		{
			name:            "with parameters and selectors",
			params:          taskparser.Params{"key": []string{"value"}},
			selectors:       selectors.Selectors{"env": {"dev": true}},
			resume:          false,
			expectParamsLog: true,
			expectSelectors: true,
		},
		{
			name:            "with only parameters",
			params:          taskparser.Params{"key": []string{"value"}},
			selectors:       selectors.Selectors{},
			resume:          false,
			expectParamsLog: true,
			expectSelectors: true, // task_name is always added
		},
		{
			name:            "with only selectors",
			params:          taskparser.Params{},
			selectors:       selectors.Selectors{"env": {"dev": true}},
			resume:          false,
			expectParamsLog: true, // Always logged (may be empty)
			expectSelectors: true,
		},
		{
			name:            "with resume mode",
			params:          taskparser.Params{},
			selectors:       selectors.Selectors{},
			resume:          true,
			expectParamsLog: true, // Always logged (may be empty)
			expectSelectors: true, // resume=true + task_name are added
		},
		{
			name:            "with no parameters or selectors",
			params:          taskparser.Params{},
			selectors:       selectors.Selectors{},
			resume:          false,
			expectParamsLog: true, // Always logged (may be empty)
			expectSelectors: true, // task_name is always added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			// Create a simple task
			createTask(t, tmpDir, "test-task", "task_name: test-task", "Test task content")

			// Create a custom logger that captures log output
			var logOutput strings.Builder

			logger := slog.New(slog.NewTextHandler(&logOutput, nil))

			// Create context with test options
			cc := New(
				WithParams(tt.params),
				WithSelectors(tt.selectors),
				WithSearchPaths("file://"+tmpDir),
				WithLogger(logger),
				WithResume(tt.resume),
			)

			// Run the context
			_, err := cc.Run(context.Background(), "test-task")
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			// Check log output
			logs := logOutput.String()

			// Count occurrences of "Parameters" and "Selectors" log messages
			paramsCount := strings.Count(logs, "msg=Parameters")
			selectorsCount := strings.Count(logs, "msg=Selectors")

			// Verify parameters logging - always exactly once
			if tt.expectParamsLog {
				if paramsCount != 1 {
					t.Errorf("expected exactly 1 Parameters log, got %d", paramsCount)
				}
			} else {
				if paramsCount != 0 {
					t.Errorf("expected no Parameters log, got %d", paramsCount)
				}
			}

			// Verify selectors logging - always exactly once
			if tt.expectSelectors {
				if selectorsCount != 1 {
					t.Errorf("expected exactly 1 Selectors log, got %d", selectorsCount)
				}
			} else {
				if selectorsCount != 0 {
					t.Errorf("expected no Selectors log, got %d", selectorsCount)
				}
			}
		})
	}
}

type skillDiscoveryCase struct {
	name      string
	setup     func(t *testing.T, dir string)
	opts      []Option
	taskName  string
	wantErr   bool
	checkFunc func(t *testing.T, result *Result)
}

func checkSkillMetadata(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Skills.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
	}

	skill := result.Skills.Skills[0]
	if skill.Name != "test-skill" {
		t.Errorf("expected skill name 'test-skill', got %q", skill.Name)
	}

	if skill.Description != "A test skill for unit testing" {
		t.Errorf("expected skill description 'A test skill for unit testing', got %q", skill.Description)
	}

	if skill.Location == "" {
		t.Error("expected skill Location to be set")
	}
}

func checkSkillsOneAndTwo(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Skills.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(result.Skills.Skills))
	}

	names := []string{result.Skills.Skills[0].Name, result.Skills.Skills[1].Name}
	if (names[0] != "skill-one" && names[0] != "skill-two") || //nolint:gosec
		(names[1] != "skill-one" && names[1] != "skill-two") {
		t.Errorf("expected skills 'skill-one' and 'skill-two', got %v", names)
	}
}

func checkSkillFilteredBySelector(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Skills.Skills) != 1 {
		t.Fatalf("expected 1 skill matching selector, got %d", len(result.Skills.Skills))
	}

	if result.Skills.Skills[0].Name != "dev-skill" {
		t.Errorf("expected skill name 'dev-skill', got %q", result.Skills.Skills[0].Name)
	}
}

func checkSkillsBootstrapDisabled(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Skills.Skills) != 0 {
		t.Errorf("expected 0 skills when bootstrap is disabled, got %d", len(result.Skills.Skills))
	}
}

func checkSkillResumeMode(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Skills.Skills) != 1 {
		t.Errorf("expected 1 skill when resume is true but bootstrap is enabled, got %d", len(result.Skills.Skills))
	}
}

func checkCursorSkill(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Skills.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
	}

	skill := result.Skills.Skills[0]
	if skill.Name != cursorSkillName {
		t.Errorf("expected skill name %q, got %q", cursorSkillName, skill.Name)
	}

	if skill.Description != "A skill for Cursor IDE" {
		t.Errorf("expected skill description 'A skill for Cursor IDE', got %q", skill.Description)
	}

	if skill.Location == "" {
		t.Error("expected skill Location to be set")
	}
}

func checkAgentsAndCursorSkills(t *testing.T, result *Result) {
	t.Helper()

	if len(result.Skills.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(result.Skills.Skills))
	}

	names := []string{result.Skills.Skills[0].Name, result.Skills.Skills[1].Name}
	if (names[0] != "agents-skill" && names[0] != cursorSkillName) || //nolint:gosec
		(names[1] != "agents-skill" && names[1] != cursorSkillName) {
		t.Errorf("expected skills 'agents-skill' and %q, got %v", cursorSkillName, names)
	}
}

func checkSingleSkillNamed(name string) func(t *testing.T, result *Result) {
	return func(t *testing.T, result *Result) {
		t.Helper()

		if len(result.Skills.Skills) != 1 {
			t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
		}

		if result.Skills.Skills[0].Name != name {
			t.Errorf("expected skill name %q, got %q", name, result.Skills.Skills[0].Name)
		}
	}
}

//nolint:funlen,maintidx
func skillDiscoveryCases() []skillDiscoveryCase {
	return []skillDiscoveryCase{
		{
			name: "discover skills with metadata",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill directory with SKILL.md
				createSkill(t, dir, filepath.Join(".agents", "skills", "test-skill"), `---
name: test-skill
description: A test skill for unit testing
license: MIT
metadata:
  author: test-author
  version: "1.0"
---

# Test Skill

This is a test skill.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSkillMetadata,
		},
		{
			name: "discover multiple skills",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create first skill
				createSkill(t, dir, filepath.Join(".agents", "skills", "skill-one"), `---
name: skill-one
description: First test skill
---

# Skill One
`)

				// Create second skill
				createSkill(t, dir, filepath.Join(".agents", "skills", "skill-two"), `---
name: skill-two
description: Second test skill
---

# Skill Two
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSkillsOneAndTwo,
		},
		{
			name: "error on skills with missing required fields",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with missing name - should cause error
				createSkill(t, dir, filepath.Join(".agents", "skills", "invalid-skill-1"), `---
description: Missing name field
---

# Invalid Skill
`)
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "error on skill with missing description",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with missing description - should cause error
				createSkill(t, dir, filepath.Join(".agents", "skills", "invalid-skill"), `---
name: invalid-skill
---

# Invalid Skill
`)
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "skills filtered by selectors",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with environment selector
				createSkill(t, dir, filepath.Join(".agents", "skills", "dev-skill"), `---
name: dev-skill
description: Development environment skill
env: development
---

# Dev Skill
`)

				// Create skill with production selector
				createSkill(t, dir, filepath.Join(".agents", "skills", "prod-skill"), `---
name: prod-skill
description: Production environment skill
env: production
---

# Prod Skill
`)
			},
			opts: []Option{
				WithSelectors(selectors.Selectors{"env": {"development": true}}),
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSkillFilteredBySelector,
		},
		{
			name: "error on skills with invalid field lengths",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with name too long (>64 chars) - should cause error
				createSkill(t, dir, filepath.Join(".agents", "skills", "long-name-skill"), `---
name: this-is-a-very-long-skill-name-that-exceeds-the-maximum-allowed-length-of-64-characters
description: Valid description
---

# Long Name Skill
`)
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "error on skill with description too long",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with description too long (>1024 chars) - should cause error
				longDesc := strings.Repeat("a", 1025)
				createSkill(t, dir, filepath.Join(".agents", "skills", "long-desc-skill"), fmt.Sprintf(`---
name: long-desc-skill
description: %s
---

# Long Desc Skill
`, longDesc))
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "bootstrap disabled skips skill discovery",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Task content")

				// Create skill directory with SKILL.md
				createSkill(t, dir, filepath.Join(".agents", "skills", "test-skill"), `---
name: test-skill
description: A test skill that should not be discovered when bootstrap is disabled
---

# Test Skill

This is a test skill.
`)
			},
			opts:      []Option{WithBootstrap(false)},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSkillsBootstrapDisabled,
		},
		{
			name: "resume mode does not skip skill discovery",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Task content")

				// Create skill directory with SKILL.md
				createSkill(t, dir, filepath.Join(".agents", "skills", "test-skill"), `---
name: test-skill
description: A test skill that should be discovered even in resume mode
---

# Test Skill

This is a test skill.
`)
			},
			opts:      []Option{WithResume(true)},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSkillResumeMode,
		},
		{
			name: "discover skills from .cursor/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .cursor/skills directory
				cursorContent := "---\nname: " + cursorSkillName +
					"\ndescription: A skill for Cursor IDE\n---\n\n# Cursor Skill\n\nThis is a skill for Cursor.\n"
				createSkill(t, dir, filepath.Join(".cursor", "skills", cursorSkillName), cursorContent)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkCursorSkill,
		},
		{
			name: "discover skills from both .agents/skills and .cursor/skills",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .agents/skills directory
				createSkill(t, dir, filepath.Join(".agents", "skills", "agents-skill"), `---
name: agents-skill
description: A generic agents skill
---

# Agents Skill
`)

				// Create skill in .cursor/skills directory
				createSkill(t, dir, filepath.Join(".cursor", "skills", cursorSkillName),
					"---\nname: "+cursorSkillName+"\ndescription: A Cursor IDE skill\n---\n\n# Cursor Skill\n")
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkAgentsAndCursorSkills,
		},
		{
			name: "discover skills from .opencode/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .opencode/skills directory
				createSkill(t, dir, filepath.Join(".opencode", "skills", "opencode-skill"), `---
name: opencode-skill
description: A skill for OpenCode
---

# OpenCode Skill

This is a skill for OpenCode.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSingleSkillNamed("opencode-skill"),
		},
		{
			name: "discover skills from .github/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .github/skills directory
				createSkill(t, dir, filepath.Join(".github", "skills", "copilot-skill"), `---
name: copilot-skill
description: A skill for GitHub Copilot
---

# Copilot Skill

This is a skill for GitHub Copilot.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSingleSkillNamed("copilot-skill"),
		},
		{
			name: "discover skills from .augment/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .augment/skills directory
				createSkill(t, dir, filepath.Join(".augment", "skills", "augment-skill"), `---
name: augment-skill
description: A skill for Augment
---

# Augment Skill

This is a skill for Augment.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSingleSkillNamed("augment-skill"),
		},
		{
			name: "discover skills from .windsurf/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .windsurf/skills directory
				createSkill(t, dir, filepath.Join(".windsurf", "skills", "windsurf-skill"), `---
name: windsurf-skill
description: A skill for Windsurf
---

# Windsurf Skill

This is a skill for Windsurf.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSingleSkillNamed("windsurf-skill"),
		},
		{
			name: "discover skills from .claude/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .claude/skills directory
				createSkill(t, dir, filepath.Join(".claude", "skills", "claude-skill"), `---
name: claude-skill
description: A skill for Claude
---

# Claude Skill

This is a skill for Claude.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSingleSkillNamed("claude-skill"),
		},
		{
			name: "discover skills from .gemini/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .gemini/skills directory
				createSkill(t, dir, filepath.Join(".gemini", "skills", "gemini-skill"), `---
name: gemini-skill
description: A skill for Gemini
---

# Gemini Skill

This is a skill for Gemini.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSingleSkillNamed("gemini-skill"),
		},
		{
			name: "discover skills from .codex/skills directory",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .codex/skills directory
				createSkill(t, dir, filepath.Join(".codex", "skills", "codex-skill"), `---
name: codex-skill
description: A skill for Codex
---

# Codex Skill

This is a skill for Codex.
`)
			},
			taskName:  "test-task",
			wantErr:   false,
			checkFunc: checkSingleSkillNamed("codex-skill"),
		},
	}
}

// TestSkillDiscovery tests skill discovery functionality.
func TestSkillDiscovery(t *testing.T) {
	t.Parallel()

	for _, tt := range skillDiscoveryCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			tt.setup(t, tmpDir)

			opts := append([]Option{
				WithSearchPaths("file://" + tmpDir),
			}, tt.opts...)
			cc := New(opts...)

			result, err := cc.Run(context.Background(), tt.taskName)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestLenientSearchPaths tests that WithLenientSearchPaths makes a best effort
// to recover or skip problematic files instead of returning errors.
func TestLenientSearchPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T, strictDir, lenientDir string)
		taskName  string
		wantErr   bool
		checkFunc func(t *testing.T, result *Result)
	}{
		{
			name: "lenient: infer skill name from directory when name is missing",
			setup: func(t *testing.T, strictDir, lenientDir string) {
				t.Helper()
				createTask(t, strictDir, "test-task", "", "Test task content")

				// Skill missing name — should infer "analyze-transcripts" from directory
				createSkill(t, lenientDir, filepath.Join(".agents", "skills", "analyze-transcripts"), `---
description: Analyzes call transcripts
---

# Analyze Transcripts
`)
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				t.Helper()
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				if result.Skills.Skills[0].Name != "analyze-transcripts" {
					t.Errorf("expected inferred skill name 'analyze-transcripts', got %q", result.Skills.Skills[0].Name)
				}
			},
		},
		{
			name: "lenient: skip skill when description is missing",
			setup: func(t *testing.T, strictDir, lenientDir string) {
				t.Helper()
				createTask(t, strictDir, "test-task", "", "Test task content")

				// Skill missing description — should be skipped
				createSkill(t, lenientDir, filepath.Join(".agents", "skills", "no-desc-skill"), `---
name: no-desc-skill
---

# No Description Skill
`)
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				t.Helper()
				if len(result.Skills.Skills) != 0 {
					t.Errorf("expected 0 skills (skipped due to missing description), got %d", len(result.Skills.Skills))
				}
			},
		},
		{
			name: "lenient: include skill when name exceeds max length",
			setup: func(t *testing.T, strictDir, lenientDir string) {
				t.Helper()
				createTask(t, strictDir, "test-task", "", "Test task content")

				createSkill(t, lenientDir, filepath.Join(".agents", "skills", "long-name-skill"), `---
name: this-is-a-very-long-skill-name-that-exceeds-the-maximum-allowed-length-of-64-characters
description: Valid description
---

# Long Name Skill
`)
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				t.Helper()
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill (lenient mode allows oversized name), got %d", len(result.Skills.Skills))
				}
				wantName := "this-is-a-very-long-skill-name-that-exceeds-the-maximum-allowed-length-of-64-characters"
				if result.Skills.Skills[0].Name != wantName {
					t.Errorf("expected skill name %q, got %q", wantName, result.Skills.Skills[0].Name)
				}
			},
		},
		{
			name: "lenient: include skill when description exceeds max length",
			setup: func(t *testing.T, strictDir, lenientDir string) {
				t.Helper()
				createTask(t, strictDir, "test-task", "", "Test task content")

				longDesc := strings.Repeat("a", 1025)
				createSkill(t, lenientDir, filepath.Join(".agents", "skills", "long-desc-skill"), fmt.Sprintf(`---
name: long-desc-skill
description: %s
---

# Long Desc Skill
`, longDesc))
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				t.Helper()
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill (lenient mode allows oversized description), got %d", len(result.Skills.Skills))
				}
				if got := len(result.Skills.Skills[0].Description); got != 1025 {
					t.Errorf("expected description length 1025, got %d", got)
				}
			},
		},
		{
			name: "lenient: skip skill with bad YAML frontmatter",
			setup: func(t *testing.T, strictDir, lenientDir string) {
				t.Helper()
				createTask(t, strictDir, "test-task", "", "Test task content")

				createSkill(t, lenientDir, filepath.Join(".agents", "skills", "bad-yaml-skill"), `---
name: [invalid yaml
description: this won't parse
---

# Bad YAML Skill
`)
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				t.Helper()
				if len(result.Skills.Skills) != 0 {
					t.Errorf("expected 0 skills (skipped due to bad YAML), got %d", len(result.Skills.Skills))
				}
			},
		},
		{
			name: "strict path still errors on skill missing name",
			setup: func(t *testing.T, strictDir, _ string) {
				t.Helper()
				createTask(t, strictDir, "test-task", "", "Test task content")

				// Same broken skill on strict path — should still error
				createSkill(t, strictDir, filepath.Join(".agents", "skills", "invalid-skill"), `---
description: Missing name field
---

# Invalid Skill
`)
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "lenient: valid skills from lenient path are still discovered",
			setup: func(t *testing.T, strictDir, lenientDir string) {
				t.Helper()
				createTask(t, strictDir, "test-task", "", "Test task content")

				createSkill(t, lenientDir, filepath.Join(".agents", "skills", "good-skill"), `---
name: good-skill
description: A perfectly valid skill on a lenient path
---

# Good Skill
`)
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				t.Helper()
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				if result.Skills.Skills[0].Name != "good-skill" {
					t.Errorf("expected skill name 'good-skill', got %q", result.Skills.Skills[0].Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			strictDir := t.TempDir()
			lenientDir := t.TempDir()

			tt.setup(t, strictDir, lenientDir)

			opts := []Option{
				WithSearchPaths("file://" + strictDir),
				WithLenientSearchPaths("file://" + lenientDir),
			}
			cc := New(opts...)

			result, err := cc.Run(context.Background(), tt.taskName)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestLenientAgent tests that WithLenientAgent makes agent paths lenient.
func TestLenientAgent(t *testing.T) {
	t.Parallel()

	t.Run("lenient agent skips skill with missing description", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		createTask(t, tmpDir, "test-task", "", "Test task content")

		// Skill missing description in a claude agent path
		createSkill(t, tmpDir, filepath.Join(".claude", "skills", "broken-skill"), `---
name: broken-skill
---

# Broken Skill
`)

		cc := New(
			WithSearchPaths("file://"+tmpDir),
			WithLenientAgent(AgentClaude),
		)

		result, err := cc.Run(context.Background(), "test-task")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Skills.Skills) != 0 {
			t.Errorf("expected 0 skills (skipped due to missing description), got %d", len(result.Skills.Skills))
		}
	})

	t.Run("lenient agent infers skill name from directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		createTask(t, tmpDir, "test-task", "", "Test task content")

		// Skill missing name in a claude agent path
		createSkill(t, tmpDir, filepath.Join(".claude", "skills", "inferred-skill"), `---
description: Should infer name from directory
---

# Inferred Skill
`)

		cc := New(
			WithSearchPaths("file://"+tmpDir),
			WithLenientAgent(AgentClaude),
		)

		result, err := cc.Run(context.Background(), "test-task")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Skills.Skills) != 1 {
			t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
		}

		if result.Skills.Skills[0].Name != "inferred-skill" {
			t.Errorf("expected inferred skill name 'inferred-skill', got %q", result.Skills.Skills[0].Name)
		}
	})
}

// TestLenientAgentMutualExclusion tests that WithAgent and WithLenientAgent are mutually exclusive.
func TestLenientAgentMutualExclusion(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createTask(t, tmpDir, "test-task", "", "Test task content")

	cc := New(
		WithSearchPaths("file://"+tmpDir),
		WithAgent(AgentClaude),
		WithLenientAgent(AgentClaude),
	)

	_, err := cc.Run(context.Background(), "test-task")
	if err == nil {
		t.Fatal("expected error when both WithAgent and WithLenientAgent are set, but got none")
	}
}
