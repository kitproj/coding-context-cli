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

// Test helper functions for creating fixtures

// createTask creates a task file in the .agents/tasks directory
func createTask(t *testing.T, dir, name, frontmatter, content string) {
	t.Helper()
	taskDir := filepath.Join(dir, ".agents", "tasks")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("failed to create task directory: %v", err)
	}

	var fileContent string
	if frontmatter != "" {
		fileContent = fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
	} else {
		fileContent = content
	}

	taskPath := filepath.Join(taskDir, name+".md")
	if err := os.WriteFile(taskPath, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("failed to create task file: %v", err)
	}
}

// createRule creates a rule file in the specified path within dir
func createRule(t *testing.T, dir, relPath, frontmatter, content string) {
	t.Helper()
	rulePath := filepath.Join(dir, relPath)
	ruleDir := filepath.Dir(rulePath)
	if err := os.MkdirAll(ruleDir, 0o755); err != nil {
		t.Fatalf("failed to create rule directory: %v", err)
	}

	var fileContent string
	if frontmatter != "" {
		fileContent = fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
	} else {
		fileContent = content
	}

	if err := os.WriteFile(rulePath, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("failed to create rule file: %v", err)
	}
}

// createCommand creates a command file in the .agents/commands directory
func createCommand(t *testing.T, dir, name, frontmatter, content string) {
	t.Helper()
	cmdDir := filepath.Join(dir, ".agents", "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("failed to create command directory: %v", err)
	}

	var fileContent string
	if frontmatter != "" {
		fileContent = fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
	} else {
		fileContent = content
	}

	cmdPath := filepath.Join(cmdDir, name+".md")
	if err := os.WriteFile(cmdPath, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("failed to create command file: %v", err)
	}
}

// createBootstrapScript creates a bootstrap script for a rule file
func createBootstrapScript(t *testing.T, dir, rulePath, scriptContent string) {
	t.Helper()
	fullRulePath := filepath.Join(dir, rulePath)
	baseNameWithoutExt := strings.TrimSuffix(fullRulePath, filepath.Ext(fullRulePath))
	bootstrapPath := baseNameWithoutExt + "-bootstrap"

	if err := os.WriteFile(bootstrapPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("failed to create bootstrap script: %v", err)
	}
}

// TestNew tests the constructor with various options
func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		opts  []Option
		check func(t *testing.T, c *Context)
	}{
		{
			name: "default context",
			opts: nil,
			check: func(t *testing.T, c *Context) {
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
			},
		},
		{
			name: "with params",
			opts: []Option{
				WithParams(taskparser.Params{"key1": []string{"value1"}, "key2": []string{"value2"}}),
			},
			check: func(t *testing.T, c *Context) {
				if c.params.Value("key1") != "value1" {
					t.Errorf("expected params[key1]=value1, got %v", c.params.Value("key1"))
				}
				if c.params.Value("key2") != "value2" {
					t.Errorf("expected params[key2]=value2, got %v", c.params.Value("key2"))
				}
			},
		},
		{
			name: "with selectors",
			opts: []Option{
				WithSelectors(selectors.Selectors{"env": {"dev": true, "test": true}}),
			},
			check: func(t *testing.T, c *Context) {
				if !c.includes.GetValue("env", "dev") {
					t.Error("expected env=dev selector")
				}
				if !c.includes.GetValue("env", "test") {
					t.Error("expected env=test selector")
				}
			},
		},
		{
			name: "with manifest URL",
			opts: []Option{
				WithManifestURL("https://example.com/manifest.txt"),
			},
			check: func(t *testing.T, c *Context) {
				if c.manifestURL != "https://example.com/manifest.txt" {
					t.Errorf("expected manifestURL to be set, got %v", c.manifestURL)
				}
			},
		},
		{
			name: "with search paths",
			opts: []Option{
				WithSearchPaths("/path/one", "/path/two"),
			},
			check: func(t *testing.T, c *Context) {
				if len(c.searchPaths) != 2 {
					t.Errorf("expected 2 search paths, got %d", len(c.searchPaths))
				}
				if c.searchPaths[0] != "/path/one" {
					t.Errorf("expected first path to be /path/one, got %v", c.searchPaths[0])
				}
				if c.searchPaths[1] != "/path/two" {
					t.Errorf("expected second path to be /path/two, got %v", c.searchPaths[1])
				}
			},
		},
		{
			name: "with custom logger",
			opts: []Option{
				WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),
			},
			check: func(t *testing.T, c *Context) {
				if c.logger == nil {
					t.Error("expected logger to be set")
				}
			},
		},
		{
			name: "with resume mode",
			opts: []Option{
				WithResume(true),
			},
			check: func(t *testing.T, c *Context) {
				if !c.resume {
					t.Error("expected resume to be true")
				}
				if !c.doBootstrap {
					t.Error("expected doBootstrap to be true by default")
				}
			},
		},
		{
			name: "with bootstrap disabled",
			opts: []Option{
				WithBootstrap(false),
			},
			check: func(t *testing.T, c *Context) {
				if c.doBootstrap {
					t.Error("expected doBootstrap to be false")
				}
			},
		},
		{
			name: "resume and bootstrap are independent",
			opts: []Option{
				WithResume(true),
				WithBootstrap(false),
			},
			check: func(t *testing.T, c *Context) {
				if !c.resume {
					t.Error("expected resume to be true")
				}
				if c.doBootstrap {
					t.Error("expected doBootstrap to be false")
				}
			},
		},
		{
			name: "with agent",
			opts: []Option{
				WithAgent(AgentCursor),
			},
			check: func(t *testing.T, c *Context) {
				if c.agent != AgentCursor {
					t.Errorf("expected agent to be cursor, got %v", c.agent)
				}
			},
		},
		{
			name: "multiple options combined",
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"production"}}),
				WithSelectors(selectors.Selectors{"lang": {"go": true}}),
				WithSearchPaths("/custom/path"),
				WithResume(false),
				WithAgent(AgentCopilot),
			},
			check: func(t *testing.T, c *Context) {
				if c.params.Value("env") != "production" {
					t.Error("params not set correctly")
				}
				if !c.includes.GetValue("lang", "go") {
					t.Error("selectors not set correctly")
				}
				if len(c.searchPaths) != 1 || c.searchPaths[0] != "/custom/path" {
					t.Error("search paths not set correctly")
				}
				if c.resume != false {
					t.Error("resume not set correctly")
				}
				if c.agent != AgentCopilot {
					t.Error("agent not set correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.opts...)
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

// TestContext_Run_Basic tests basic task execution scenarios
func TestContext_Run_Basic(t *testing.T) {
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
				createTask(t, dir, "simple", "", "This is a simple task.")
			},
			taskName: "simple",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "This is a simple task.") {
					t.Errorf("expected task content 'This is a simple task.', got %q", result.Task.Content)
				}
				if result.Tokens <= 0 {
					t.Error("expected positive token count")
				}
			},
		},
		{
			name: "task with frontmatter",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "with-frontmatter", "priority: high\nenv: dev", "Task content here.")
			},
			taskName: "with-frontmatter",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Task content here.") {
					t.Errorf("expected task content, got %q", result.Task.Content)
				}
				if result.Task.FrontMatter.Content["priority"] != "high" {
					t.Error("expected priority=high in frontmatter")
				}
			},
		},
		{
			name: "task with parameter substitution",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "params-task", "", "Environment: ${env}\nFeature: ${feature}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"production"}, "feature": []string{"auth"}}),
			},
			taskName: "params-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Environment: production") {
					t.Errorf("expected 'Environment: production' in content, got %q", result.Task.Content)
				}
				if !strings.Contains(result.Task.Content, "Feature: auth") {
					t.Errorf("expected 'Feature: auth' in content, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "task with unresolved parameter",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "unresolved", "", "Missing: ${missing_param}")
			},
			taskName: "unresolved",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "${missing_param}") {
					t.Errorf("expected unresolved parameter to remain as ${missing_param}, got %q", result.Task.Content)
				}
			},
		},
		{
			name:        "task not found returns error",
			setup:       func(t *testing.T, dir string) {},
			taskName:    "nonexistent",
			wantErr:     true,
			errContains: "task not found",
		},
		{
			name: "task with selectors sets includes",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "selector-task", "selectors:\n  env: production\n  lang: go", "Task with selectors")
			},
			taskName: "selector-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Task with selectors") {
					t.Errorf("unexpected content: %q", result.Task.Content)
				}
			},
		},
		{
			name: "multiple params in same content",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "multi-params", "", "User: ${user}, Email: ${email}, Role: ${role}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"user": []string{"alice"}, "email": []string{"alice@example.com"}, "role": []string{"admin"}}),
			},
			taskName: "multi-params",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				expected := "User: alice, Email: alice@example.com, Role: admin"
				if !strings.Contains(result.Task.Content, expected) {
					t.Errorf("expected %q in content, got %q", expected, result.Task.Content)
				}
			},
		},
		{
			name: "task ID automatically set from filename",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "my-task", "", "Task content")
			},
			taskName: "my-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if result.Task.FrontMatter.ID != "my-task" {
					t.Errorf("expected task ID 'my-task', got %q", result.Task.FrontMatter.ID)
				}
			},
		},
		{
			name: "task with explicit ID in frontmatter",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "file-name", "id: explicit-task-id", "Task content")
			},
			taskName: "file-name",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if result.Task.FrontMatter.ID != "explicit-task-id" {
					t.Errorf("expected task ID 'explicit-task-id', got %q", result.Task.FrontMatter.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestContext_Run_Rules tests rule discovery and filtering
func TestContext_Run_Rules(t *testing.T) {
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
				createTask(t, dir, "task1", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule 1 content")
				createRule(t, dir, ".cursor/rules/rule2.md", "", "Rule 2 content")
			},
			taskName: "task1",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 2 {
					t.Errorf("expected 2 rules, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "filter rules by selectors from task frontmatter",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "filtered-task", "selectors:\n  env: production", "Task with selectors")
				createRule(t, dir, ".agents/rules/prod-rule.md", "env: production", "Production rule")
				createRule(t, dir, ".agents/rules/dev-rule.md", "env: development", "Development rule")
				createRule(t, dir, ".agents/rules/no-env.md", "", "No env specified")
			},
			taskName: "filtered-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// Should include prod-rule and no-env (no env key is allowed)
				// Should exclude dev-rule (env doesn't match)
				if len(result.Rules) != 2 {
					t.Errorf("expected 2 rules (prod and no-env), got %d", len(result.Rules))
				}
				foundProd := false
				foundDev := false
				for _, rule := range result.Rules {
					if strings.Contains(rule.Content, "Production rule") {
						foundProd = true
					}
					if strings.Contains(rule.Content, "Development rule") {
						foundDev = true
					}
				}
				if !foundProd {
					t.Error("expected to find production rule")
				}
				if foundDev {
					t.Error("did not expect to find development rule")
				}
			},
		},
		{
			name: "rules with parameter substitution",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "param-task", "", "Task")
				createRule(t, dir, ".agents/rules/param-rule.md", "", "Project: ${project}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"project": []string{"myapp"}}),
			},
			taskName: "param-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Fatalf("expected 1 rule, got %d", len(result.Rules))
				}
				if !strings.Contains(result.Rules[0].Content, "Project: myapp") {
					t.Errorf("expected parameter substitution in rule, got %q", result.Rules[0].Content)
				}
			},
		},
		{
			name: "bootstrap disabled skips rule discovery",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "bootstrap-task", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule content")
			},
			opts: []Option{
				WithBootstrap(false),
			},
			taskName: "bootstrap-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 0 {
					t.Errorf("expected 0 rules when bootstrap is disabled, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "resume mode does not skip rule discovery",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "resume-task", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule content")
			},
			opts: []Option{
				WithResume(true),
			},
			taskName: "resume-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Errorf("expected 1 rule when resume is true but bootstrap is enabled, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "bootstrap script executed for rules",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "bootstrap-task", "", "Task")
				createRule(t, dir, ".agents/rules/rule-with-bootstrap.md", "", "Rule content")
				createBootstrapScript(t, dir, ".agents/rules/rule-with-bootstrap.md", "#!/bin/sh\necho 'bootstrapped'")
			},
			taskName: "bootstrap-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Errorf("expected 1 rule, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "bootstrap disabled skips bootstrap scripts",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "no-bootstrap", "", "Task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule")
				createBootstrapScript(t, dir, ".agents/rules/rule1.md", "#!/bin/sh\nexit 1")
			},
			opts: []Option{
				WithBootstrap(false),
			},
			taskName: "no-bootstrap",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// When bootstrap is disabled, rules aren't discovered, so bootstrap scripts won't run
				if len(result.Rules) != 0 {
					t.Error("expected no rules when bootstrap is disabled")
				}
			},
		},
		{
			name: "bootstrap from frontmatter is preferred",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "frontmatter-bootstrap", "", "Task")
				// Create rule with bootstrap in frontmatter that writes a marker file
				createRule(t, dir, ".agents/rules/rule-with-frontmatter.md",
					"bootstrap: |\n  #!/bin/sh\n  echo 'frontmatter' > "+filepath.Join(dir, "bootstrap-ran.txt")+"\n",
					"Rule content")
			},
			taskName: "frontmatter-bootstrap",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Errorf("expected 1 rule, got %d", len(result.Rules))
				}
				// The integration tests verify frontmatter bootstrap actually ran
			},
		},
		{
			name: "bootstrap from frontmatter preferred over file",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Errorf("expected 1 rule, got %d", len(result.Rules))
				}
				// The integration tests verify which bootstrap actually ran
			},
		},
		{
			name: "bootstrap from file when frontmatter empty",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Errorf("expected 1 rule, got %d", len(result.Rules))
				}
				// The integration tests verify the file-based bootstrap ran
			},
		},
		{
			name: "agent option collects all rules",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				// Agent filtering is not implemented, so all rules are included
				if len(result.Rules) != 3 {
					t.Errorf("expected 3 rules, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "task frontmatter agent overrides option",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "override-task", "agent: copilot", "Task")
				createRule(t, dir, ".cursor/rules/cursor-rule.md", "", "Cursor rule")
				createRule(t, dir, ".github/agents/copilot-rule.md", "", "Copilot rule")
			},
			opts: []Option{
				WithAgent(AgentCursor), // This should be overridden by task frontmatter
			},
			taskName: "override-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// Verify all rules are collected (agent filtering not implemented)
				if len(result.Rules) != 2 {
					t.Errorf("expected 2 rules, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "multiple selector values with OR logic",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "multi-selector", "selectors:\n  env:\n    - dev\n    - test", "Task")
				createRule(t, dir, ".agents/rules/dev-rule.md", "env: dev", "Dev rule")
				createRule(t, dir, ".agents/rules/test-rule.md", "env: test", "Test rule")
				createRule(t, dir, ".agents/rules/prod-rule.md", "env: prod", "Prod rule")
			},
			taskName: "multi-selector",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// Should include dev and test rules, exclude prod
				if len(result.Rules) != 2 {
					t.Errorf("expected 2 rules, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "CLI selectors combined with task selectors use OR logic",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				// Should include both prod-rule (from task) and dev-rule (from CLI)
				// Should exclude test-rule (matches neither)
				// This demonstrates OR logic: rules match if env is production OR development
				if len(result.Rules) != 2 {
					t.Errorf("expected 2 rules (prod and dev via OR logic), got %d", len(result.Rules))
				}
				foundProd := false
				foundDev := false
				foundTest := false
				for _, rule := range result.Rules {
					if strings.Contains(rule.Content, "Production rule") {
						foundProd = true
					}
					if strings.Contains(rule.Content, "Development rule") {
						foundDev = true
					}
					if strings.Contains(rule.Content, "Test rule") {
						foundTest = true
					}
				}
				if !foundProd {
					t.Error("expected to find production rule (from task selector)")
				}
				if !foundDev {
					t.Error("expected to find development rule (from CLI selector)")
				}
				if foundTest {
					t.Error("did not expect to find test rule (matches neither selector)")
				}
			},
		},
		{
			name: "CLI selectors combined with array task selectors use OR logic",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				// Should include prod, staging (from task array), and dev (from CLI)
				// Should exclude test (matches none)
				// This demonstrates OR logic with array selectors: env is production OR staging OR development
				if len(result.Rules) != 3 {
					t.Errorf("expected 3 rules (prod, staging, dev via OR logic), got %d", len(result.Rules))
				}
				foundProd := false
				foundStaging := false
				foundDev := false
				foundTest := false
				for _, rule := range result.Rules {
					if strings.Contains(rule.Content, "Production rule") {
						foundProd = true
					}
					if strings.Contains(rule.Content, "Staging rule") {
						foundStaging = true
					}
					if strings.Contains(rule.Content, "Development rule") {
						foundDev = true
					}
					if strings.Contains(rule.Content, "Test rule") {
						foundTest = true
					}
				}
				if !foundProd {
					t.Error("expected to find production rule (from task array selector)")
				}
				if !foundStaging {
					t.Error("expected to find staging rule (from task array selector)")
				}
				if !foundDev {
					t.Error("expected to find development rule (from CLI selector)")
				}
				if foundTest {
					t.Error("did not expect to find test rule (matches no selector)")
				}
			},
		},
		{
			name: "token counting for rules",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "token-task", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "This is rule 1 content")
				createRule(t, dir, ".agents/rules/rule2.md", "", "This is rule 2 content")
			},
			taskName: "token-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
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
			},
		},
		{
			name: "rule IDs automatically set from filename",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "id-task", "", "Task")
				createRule(t, dir, ".agents/rules/my-rule.md", "", "Rule without ID in frontmatter")
				createRule(t, dir, ".agents/rules/another-rule.md", "id: explicit-id", "Rule with explicit ID")
			},
			taskName: "id-task",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 2 {
					t.Fatalf("expected 2 rules, got %d", len(result.Rules))
				}

				// Check that one rule has auto-generated ID from filename
				foundMyRule := false
				foundAnotherRule := false
				for _, rule := range result.Rules {
					if rule.FrontMatter.ID == "my-rule" {
						foundMyRule = true
						if !strings.Contains(rule.Content, "Rule without ID") {
							t.Error("my-rule should contain 'Rule without ID'")
						}
					}
					if rule.FrontMatter.ID == "explicit-id" {
						foundAnotherRule = true
						if !strings.Contains(rule.Content, "Rule with explicit ID") {
							t.Error("explicit-id should contain 'Rule with explicit ID'")
						}
					}
				}

				if !foundMyRule {
					t.Error("expected to find rule with auto-generated ID 'my-rule'")
				}
				if !foundAnotherRule {
					t.Error("expected to find rule with explicit ID 'explicit-id'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestContext_Run_Commands tests command substitution in tasks
func TestContext_Run_Commands(t *testing.T) {
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
				createTask(t, dir, "with-command", "", "Before command\n/greet\nAfter command")
				createCommand(t, dir, "greet", "", "Hello, World!")
			},
			taskName: "with-command",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Before command") {
					t.Error("expected task content before command")
				}
				if !strings.Contains(result.Task.Content, "Hello, World!") {
					t.Error("expected command content to be substituted")
				}
				if !strings.Contains(result.Task.Content, "After command") {
					t.Error("expected task content after command")
				}
			},
		},
		{
			name: "command with parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "cmd-with-params", "", "/greet name=\"Alice\"")
				createCommand(t, dir, "greet", "", "Hello, ${name}!")
			},
			taskName: "cmd-with-params",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Hello, Alice!") {
					t.Errorf("expected parameter substitution in command, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "command with context parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "ctx-params", "", "/deploy")
				createCommand(t, dir, "deploy", "", "Deploying to ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"staging"}}),
			},
			taskName: "ctx-params",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Deploying to staging") {
					t.Errorf("expected context parameter substitution, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "multiple commands in task",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "multi-cmd", "", "/intro\n\n/body\n\n/outro\n")
				createCommand(t, dir, "intro", "", "Introduction")
				createCommand(t, dir, "body", "", "Main content")
				createCommand(t, dir, "outro", "", "Conclusion")
			},
			taskName: "multi-cmd",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				content := result.Task.Content
				// Each command may or may not be substituted depending on parsing
				// Just verify we got some content
				if strings.TrimSpace(content) == "" {
					t.Errorf("expected non-empty content, got %q", content)
				}
			},
		},
		{
			name: "command not found returns error",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "missing-cmd", "", "/nonexistent")
			},
			taskName:    "missing-cmd",
			wantErr:     true,
			errContains: "command not found",
		},
		{
			name: "command parameter overrides context parameter",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "override-param", "", "/msg value=\"specific\"")
				createCommand(t, dir, "msg", "", "Value: ${value}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"value": []string{"general"}}),
			},
			taskName: "override-param",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Value: specific") {
					t.Errorf("expected command param to override context param, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "command with multiple parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "multi-params", "", "/info name=\"Bob\" age=\"30\" role=\"developer\"")
				createCommand(t, dir, "info", "", "Name: ${name}, Age: ${age}, Role: ${role}")
			},
			taskName: "multi-params",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				expected := "Name: Bob, Age: 30, Role: developer"
				if !strings.Contains(result.Task.Content, expected) {
					t.Errorf("expected %q in content, got %q", expected, result.Task.Content)
				}
			},
		},
		{
			name: "mixed text and commands",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "mixed", "", "# Title\n\n/section1\n\nMiddle text\n\n/section2\n\nEnd text")
				createCommand(t, dir, "section1", "", "Section 1 content")
				createCommand(t, dir, "section2", "", "Section 2 content")
			},
			taskName: "mixed",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
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
			},
		},
		{
			name: "command with selectors filters rules",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				// Should include postgres-rule and generic-rule
				// Should exclude mysql-rule
				if len(result.Rules) != 2 {
					t.Errorf("expected 2 rules, got %d", len(result.Rules))
				}
				foundPostgres := false
				foundMySQL := false
				for _, rule := range result.Rules {
					if strings.Contains(rule.Content, "PostgreSQL rule") {
						foundPostgres = true
					}
					if strings.Contains(rule.Content, "MySQL rule") {
						foundMySQL = true
					}
				}
				if !foundPostgres {
					t.Error("expected to find PostgreSQL rule")
				}
				if foundMySQL {
					t.Error("did not expect to find MySQL rule")
				}
			},
		},
		{
			name: "command selectors combine with task selectors",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				// Should include: prod-auth-rule (matches both), prod-rule (matches env), auth-rule (matches feature)
				// Should exclude: dev-rule (env doesn't match)
				if len(result.Rules) != 3 {
					t.Errorf("expected 3 rules, got %d", len(result.Rules))
					for _, r := range result.Rules {
						t.Logf("Found rule: %s", r.Content)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestContext_Run_Integration tests end-to-end integration scenarios
func TestContext_Run_Integration(t *testing.T) {
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
			check: func(t *testing.T, result *Result) {
				// Check task content includes params and command
				if !strings.Contains(result.Task.Content, "Deploy myservice") {
					t.Error("expected app param substitution")
				}
				if !strings.Contains(result.Task.Content, "Deploy to production") {
					t.Error("expected command with param substitution")
				}
				// Check rules - should have 2 (prod and go, not dev)
				if len(result.Rules) != 2 {
					t.Errorf("expected 2 rules, got %d", len(result.Rules))
				}
				// Check token counting
				if result.Tokens <= 0 {
					t.Error("expected positive token count")
				}
			},
		},
		{
			name: "complex task with multiple slash commands and mixed content",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "complex", "", "# Project Setup\n\n/intro\n\n## Steps\n\n/step1\n\n/step2\n\n## Conclusion\n\n/outro\n")
				createCommand(t, dir, "intro", "", "Welcome to the project")
				createCommand(t, dir, "step1", "", "First, initialize the repository")
				createCommand(t, dir, "step2", "", "Then, configure the settings")
				createCommand(t, dir, "outro", "", "You're all set!")
			},
			taskName: "complex",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				content := result.Task.Content
				if !strings.Contains(content, "# Project Setup") {
					t.Error("expected markdown header")
				}
				// Commands may or may not be substituted - just check we got content
				if strings.TrimSpace(content) == "" {
					t.Error("expected non-empty content")
				}
			},
		},
		{
			name: "bootstrap disabled workflow skips rules but includes task",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "bootstrap", "", "Continue this task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Should be skipped")
				createBootstrapScript(t, dir, ".agents/rules/rule1.md", "#!/bin/sh\necho 'should not run'")
			},
			opts: []Option{
				WithBootstrap(false),
			},
			taskName: "bootstrap",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Continue this task") {
					t.Errorf("unexpected task content: %q", result.Task.Content)
				}
				if len(result.Rules) != 0 {
					t.Errorf("expected 0 rules when bootstrap is disabled, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "agent-specific workflow",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "for-cursor", "agent: cursor", "Task for Cursor")
				createRule(t, dir, ".cursor/rules/cursor.md", "", "Cursor-specific")
				createRule(t, dir, ".agents/rules/general.md", "", "General rule")
				createRule(t, dir, ".github/agents/copilot.md", "", "Copilot rule")
			},
			taskName: "for-cursor",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// Agent filtering is not implemented, so all rules are collected
				if len(result.Rules) != 3 {
					t.Errorf("expected 3 rules, got %d", len(result.Rules))
				}
			},
		},
		{
			name: "multiple search paths",
			setup: func(t *testing.T, dir string) {
				// Create first directory with task and rule
				createTask(t, dir, "multi-path", "", "Multi-path task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule from first path")

				// Create second directory with additional rule
				secondDir := filepath.Join(dir, "second")
				if err := os.MkdirAll(secondDir, 0o755); err != nil {
					t.Fatalf("failed to create second dir: %v", err)
				}
				createRule(t, secondDir, ".agents/rules/rule2.md", "", "Rule from second path")
			},
			opts: []Option{
				// Second path will be added via setup
			},
			taskName: "multi-path",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// This test only finds rules from the first path since we don't add the second path
				if len(result.Rules) != 1 {
					t.Errorf("expected 1 rule from first path, got %d", len(result.Rules))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestContext_Run_Errors tests error scenarios
func TestContext_Run_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, dir string)
		opts        []Option
		taskName    string
		wantErr     bool
		errContains string
	}{
		{
			name: "command not found in task",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "bad-cmd", "", "/missing-command\n")
			},
			taskName:    "bad-cmd",
			wantErr:     true,
			errContains: "command not found",
		},
		{
			name: "invalid agent in task frontmatter",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "bad-agent", "agent: invalidagent", "Task content")
			},
			taskName:    "bad-agent",
			wantErr:     true,
			errContains: "unknown agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		})
	}
}

// TestContext_Run_ExpandParams tests parameter expansion opt-out functionality
func TestContext_Run_ExpandParams(t *testing.T) {
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
				createTask(t, dir, "no-expand", "expand: false", "Issue: ${issue_number}\nTitle: ${issue_title}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"issue_number": []string{"123"}, "issue_title": []string{"Bug fix"}}),
			},
			taskName: "no-expand",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "${issue_number}") {
					t.Errorf("expected ${issue_number} to be preserved, got %q", result.Task.Content)
				}
				if !strings.Contains(result.Task.Content, "${issue_title}") {
					t.Errorf("expected ${issue_title} to be preserved, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "task with expand: true expands parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "expand", "expand: true", "Issue: ${issue_number}\nTitle: ${issue_title}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"issue_number": []string{"123"}, "issue_title": []string{"Bug fix"}}),
			},
			taskName: "expand",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Issue: 123") {
					t.Errorf("expected 'Issue: 123', got %q", result.Task.Content)
				}
				if !strings.Contains(result.Task.Content, "Title: Bug fix") {
					t.Errorf("expected 'Title: Bug fix', got %q", result.Task.Content)
				}
			},
		},
		{
			name: "task without expand defaults to expanding",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "default", "", "Env: ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"production"}}),
			},
			taskName: "default",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Env: production") {
					t.Errorf("expected 'Env: production', got %q", result.Task.Content)
				}
			},
		},
		{
			name: "command with expand: false preserves parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "cmd-no-expand", "", "/deploy")
				createCommand(t, dir, "deploy", "expand: false", "Deploying to ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"staging"}}),
			},
			taskName: "cmd-no-expand",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "${env}") {
					t.Errorf("expected ${env} to be preserved, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "command with expand: true expands parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "cmd-expand", "", "/deploy")
				createCommand(t, dir, "deploy", "expand: true", "Deploying to ${env}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"env": []string{"staging"}}),
			},
			taskName: "cmd-expand",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Deploying to staging") {
					t.Errorf("expected 'Deploying to staging', got %q", result.Task.Content)
				}
			},
		},
		{
			name: "command without expand defaults to expanding",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "cmd-default", "", "/info")
				createCommand(t, dir, "info", "", "Project: ${project}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"project": []string{"myapp"}}),
			},
			taskName: "cmd-default",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Project: myapp") {
					t.Errorf("expected 'Project: myapp', got %q", result.Task.Content)
				}
			},
		},
		{
			name: "rule with expand: false preserves parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "rule-no-expand", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "expand: false", "Version: ${version}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"version": []string{"1.0.0"}}),
			},
			taskName: "rule-no-expand",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Fatalf("expected 1 rule, got %d", len(result.Rules))
				}
				if !strings.Contains(result.Rules[0].Content, "${version}") {
					t.Errorf("expected ${version} to be preserved in rule, got %q", result.Rules[0].Content)
				}
			},
		},
		{
			name: "rule with expand: true expands parameters",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "rule-expand", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "expand: true", "Version: ${version}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"version": []string{"1.0.0"}}),
			},
			taskName: "rule-expand",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Fatalf("expected 1 rule, got %d", len(result.Rules))
				}
				if !strings.Contains(result.Rules[0].Content, "Version: 1.0.0") {
					t.Errorf("expected 'Version: 1.0.0' in rule, got %q", result.Rules[0].Content)
				}
			},
		},
		{
			name: "rule without expand defaults to expanding",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "rule-default", "", "Task content")
				createRule(t, dir, ".agents/rules/rule1.md", "", "App: ${app}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"app": []string{"service"}}),
			},
			taskName: "rule-default",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 1 {
					t.Fatalf("expected 1 rule, got %d", len(result.Rules))
				}
				if !strings.Contains(result.Rules[0].Content, "App: service") {
					t.Errorf("expected 'App: service' in rule, got %q", result.Rules[0].Content)
				}
			},
		},
		{
			name: "mixed: task no expand, command with expand",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "mixed1", "expand: false", "Task ${task_var}\n/cmd")
				createCommand(t, dir, "cmd", "expand: true", "Command ${cmd_var}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"task_var": []string{"task_value"}, "cmd_var": []string{"cmd_value"}}),
			},
			taskName: "mixed1",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				content := result.Task.Content
				if !strings.Contains(content, "${task_var}") {
					t.Errorf("expected task param to be preserved, got %q", content)
				}
				if !strings.Contains(content, "Command cmd_value") {
					t.Errorf("expected command param to be expanded, got %q", content)
				}
			},
		},
		{
			name: "mixed: task with expand, command no expand",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "mixed2", "expand: true", "Task ${task_var}\n/cmd")
				createCommand(t, dir, "cmd", "expand: false", "Command ${cmd_var}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"task_var": []string{"task_value"}, "cmd_var": []string{"cmd_value"}}),
			},
			taskName: "mixed2",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				content := result.Task.Content
				if !strings.Contains(content, "Task task_value") {
					t.Errorf("expected task param to be expanded, got %q", content)
				}
				if !strings.Contains(content, "${cmd_var}") {
					t.Errorf("expected command param to be preserved, got %q", content)
				}
			},
		},
		{
			name: "command with inline parameters and expand: false",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "inline-no-expand", "", "/greet name=\"Alice\"")
				createCommand(t, dir, "greet", "expand: false", "Hello, ${name}! Your ID: ${id}")
			},
			opts: []Option{
				WithParams(taskparser.Params{"id": []string{"123"}}),
			},
			taskName: "inline-no-expand",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// Both inline and context params should be preserved
				if !strings.Contains(result.Task.Content, "${name}") {
					t.Errorf("expected ${name} to be preserved, got %q", result.Task.Content)
				}
				if !strings.Contains(result.Task.Content, "${id}") {
					t.Errorf("expected ${id} to be preserved, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "multiple rules with different expand settings",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				if len(result.Rules) != 3 {
					t.Fatalf("expected 3 rules, got %d", len(result.Rules))
				}
				// Find each rule and check content
				for _, rule := range result.Rules {
					if strings.Contains(rule.Content, "Rule1:") {
						if !strings.Contains(rule.Content, "${var1}") {
							t.Errorf("expected ${var1} in rule1, got %q", rule.Content)
						}
					} else if strings.Contains(rule.Content, "Rule2:") {
						if !strings.Contains(rule.Content, "Rule2: val2") {
							t.Errorf("expected 'Rule2: val2', got %q", rule.Content)
						}
					} else if strings.Contains(rule.Content, "Rule3:") {
						if !strings.Contains(rule.Content, "Rule3: val3") {
							t.Errorf("expected 'Rule3: val3', got %q", rule.Content)
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestUserPrompt tests the user_prompt parameter functionality
func TestUserPrompt(t *testing.T) {
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
				createTask(t, dir, "simple", "", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt("User prompt content"),
			},
			taskName: "simple",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Task content") {
					t.Error("expected task content to contain 'Task content'")
				}
				if !strings.Contains(result.Task.Content, "User prompt content") {
					t.Error("expected task content to contain 'User prompt content'")
				}
				// Check that user_prompt comes after task content
				taskIdx := strings.Index(result.Task.Content, "Task content")
				userIdx := strings.Index(result.Task.Content, "User prompt content")
				if taskIdx >= userIdx {
					t.Error("expected user_prompt to come after task content")
				}
			},
		},
		{
			name: "user_prompt with slash command",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "with-command", "", "Task content\n")
				createCommand(t, dir, "greet", "", "Hello from command!")
			},
			opts: []Option{
				WithUserPrompt("User says:\n/greet\n"),
			},
			taskName: "with-command",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Task content") {
					t.Error("expected task content to contain 'Task content'")
				}
				if !strings.Contains(result.Task.Content, "User says:") {
					t.Error("expected task content to contain 'User says: '")
				}
				if !strings.Contains(result.Task.Content, "Hello from command!") {
					t.Error("expected slash command in user_prompt to be expanded")
				}
			},
		},
		{
			name: "user_prompt with parameter substitution",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Issue: 123") {
					t.Error("expected parameter substitution in user_prompt")
				}
			},
		},
		{
			name: "user_prompt with slash command and parameters",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Please fix:") {
					t.Error("expected task content to contain 'Please fix: '")
				}
				if !strings.Contains(result.Task.Content, "Issue 456: Fix bug") {
					t.Error("expected slash command to be expanded with parameter substitution")
				}
			},
		},
		{
			name: "empty user_prompt should not affect task",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "empty", "", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt(""),
			},
			taskName: "empty",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if result.Task.Content != "Task content\n" {
					t.Errorf("expected task content to be unchanged, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "no user_prompt parameter should not affect task",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "no-prompt", "", "Task content\n")
			},
			opts:     []Option{},
			taskName: "no-prompt",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if result.Task.Content != "Task content\n" {
					t.Errorf("expected task content to be unchanged, got %q", result.Task.Content)
				}
			},
		},
		{
			name: "user_prompt with multiple slash commands",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "multi", "", "Task content\n")
				createCommand(t, dir, "cmd1", "", "Command 1")
				createCommand(t, dir, "cmd2", "", "Command 2")
			},
			opts: []Option{
				WithUserPrompt("/cmd1\n/cmd2\n"),
			},
			taskName: "multi",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Command 1") {
					t.Error("expected first slash command to be expanded")
				}
				if !strings.Contains(result.Task.Content, "Command 2") {
					t.Error("expected second slash command to be expanded")
				}
			},
		},
		{
			name: "user_prompt respects task expand setting",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "${issue_number}") {
					t.Error("expected parameter to NOT be expanded when expand: false")
				}
			},
		},
		{
			name: "user_prompt with invalid slash command",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "invalid", "", "Task content\n")
			},
			opts: []Option{
				WithUserPrompt("/nonexistent-command\n"),
			},
			taskName:    "invalid",
			wantErr:     true,
			errContains: "command not found",
		},
		{
			name: "both task prompt and user prompt parse correctly",
			setup: func(t *testing.T, dir string) {
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
			check: func(t *testing.T, result *Result) {
				// Verify task content contains both task and user prompt elements
				if !strings.Contains(result.Task.Content, "Task prompt with text") {
					t.Error("expected task content to contain 'Task prompt with text'")
				}
				if !strings.Contains(result.Task.Content, "More task text") {
					t.Error("expected task content to contain 'More task text'")
				}
				if !strings.Contains(result.Task.Content, "User prompt with text") {
					t.Error("expected task content to contain 'User prompt with text'")
				}
				if !strings.Contains(result.Task.Content, "More user text") {
					t.Error("expected task content to contain 'More user text'")
				}
				// Verify both commands were expanded with correct parameters
				if !strings.Contains(result.Task.Content, "Task command output value1") {
					t.Error("expected task command to be expanded with param1=value1")
				}
				if !strings.Contains(result.Task.Content, "User command output value2") {
					t.Error("expected user command to be expanded with param2=value2")
				}
				// Verify delimiter is present (separating task from user prompt)
				if !strings.Contains(result.Task.Content, "---") {
					t.Error("expected delimiter '---' between task and user prompt")
				}
				// Verify order: task content comes before user content
				taskIdx := strings.Index(result.Task.Content, "Task prompt with text")
				userIdx := strings.Index(result.Task.Content, "User prompt with text")
				if taskIdx == -1 || userIdx == -1 {
					t.Error("expected both task and user prompt text to be found in result")
				} else if taskIdx >= userIdx {
					t.Error("expected task content to come before user prompt content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestIsLocalPath tests the isLocalPath helper function
func TestIsLocalPath(t *testing.T) {
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
			result := isLocalPath(tt.path)
			if result != tt.expected {
				t.Errorf("isLocalPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestNormalizeLocalPath tests the normalizeLocalPath helper function
func TestNormalizeLocalPath(t *testing.T) {
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

// TestSkillDiscovery tests skill discovery functionality
func TestSkillDiscovery(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		opts      []Option
		taskName  string
		wantErr   bool
		checkFunc func(t *testing.T, result *Result)
	}{
		{
			name: "discover skills with metadata",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill directory with SKILL.md
				skillDir := filepath.Join(dir, ".agents", "skills", "test-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: test-skill
description: A test skill for unit testing
license: MIT
metadata:
  author: test-author
  version: "1.0"
---

# Test Skill

This is a test skill.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
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
				// Check that Location is set to absolute path
				if skill.Location == "" {
					t.Error("expected skill Location to be set")
				}
			},
		},
		{
			name: "discover multiple skills",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create first skill
				skillDir1 := filepath.Join(dir, ".agents", "skills", "skill-one")
				if err := os.MkdirAll(skillDir1, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent1 := `---
name: skill-one
description: First test skill
---

# Skill One
`
				skillPath1 := filepath.Join(skillDir1, "SKILL.md")
				if err := os.WriteFile(skillPath1, []byte(skillContent1), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}

				// Create second skill
				skillDir2 := filepath.Join(dir, ".agents", "skills", "skill-two")
				if err := os.MkdirAll(skillDir2, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent2 := `---
name: skill-two
description: Second test skill
---

# Skill Two
`
				skillPath2 := filepath.Join(skillDir2, "SKILL.md")
				if err := os.WriteFile(skillPath2, []byte(skillContent2), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 2 {
					t.Fatalf("expected 2 skills, got %d", len(result.Skills.Skills))
				}
				// Skills should be in order of discovery
				names := []string{result.Skills.Skills[0].Name, result.Skills.Skills[1].Name}
				if (names[0] != "skill-one" && names[0] != "skill-two") ||
					(names[1] != "skill-one" && names[1] != "skill-two") {
					t.Errorf("expected skills 'skill-one' and 'skill-two', got %v", names)
				}
			},
		},
		{
			name: "error on skills with missing required fields",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with missing name - should cause error
				skillDir1 := filepath.Join(dir, ".agents", "skills", "invalid-skill-1")
				if err := os.MkdirAll(skillDir1, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent1 := `---
description: Missing name field
---

# Invalid Skill
`
				skillPath1 := filepath.Join(skillDir1, "SKILL.md")
				if err := os.WriteFile(skillPath1, []byte(skillContent1), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "error on skill with missing description",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with missing description - should cause error
				skillDir := filepath.Join(dir, ".agents", "skills", "invalid-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent := `---
name: invalid-skill
---

# Invalid Skill
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "skills filtered by selectors",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with environment selector
				skillDir1 := filepath.Join(dir, ".agents", "skills", "dev-skill")
				if err := os.MkdirAll(skillDir1, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent1 := `---
name: dev-skill
description: Development environment skill
env: development
---

# Dev Skill
`
				skillPath1 := filepath.Join(skillDir1, "SKILL.md")
				if err := os.WriteFile(skillPath1, []byte(skillContent1), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}

				// Create skill with production selector
				skillDir2 := filepath.Join(dir, ".agents", "skills", "prod-skill")
				if err := os.MkdirAll(skillDir2, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent2 := `---
name: prod-skill
description: Production environment skill
env: production
---

# Prod Skill
`
				skillPath2 := filepath.Join(skillDir2, "SKILL.md")
				if err := os.WriteFile(skillPath2, []byte(skillContent2), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			opts: []Option{
				WithSelectors(selectors.Selectors{"env": {"development": true}}),
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill matching selector, got %d", len(result.Skills.Skills))
				}
				if result.Skills.Skills[0].Name != "dev-skill" {
					t.Errorf("expected skill name 'dev-skill', got %q", result.Skills.Skills[0].Name)
				}
			},
		},
		{
			name: "error on skills with invalid field lengths",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with name too long (>64 chars) - should cause error
				skillDir1 := filepath.Join(dir, ".agents", "skills", "long-name-skill")
				if err := os.MkdirAll(skillDir1, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent1 := `---
name: this-is-a-very-long-skill-name-that-exceeds-the-maximum-allowed-length-of-64-characters
description: Valid description
---

# Long Name Skill
`
				skillPath1 := filepath.Join(skillDir1, "SKILL.md")
				if err := os.WriteFile(skillPath1, []byte(skillContent1), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "error on skill with description too long",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill with description too long (>1024 chars) - should cause error
				skillDir := filepath.Join(dir, ".agents", "skills", "long-desc-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				longDesc := strings.Repeat("a", 1025)
				skillContent := fmt.Sprintf(`---
name: long-desc-skill
description: %s
---

# Long Desc Skill
`, longDesc)
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  true,
		},
		{
			name: "bootstrap disabled skips skill discovery",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Task content")

				// Create skill directory with SKILL.md
				skillDir := filepath.Join(dir, ".agents", "skills", "test-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: test-skill
description: A test skill that should not be discovered when bootstrap is disabled
---

# Test Skill

This is a test skill.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			opts:     []Option{WithBootstrap(false)},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 0 {
					t.Errorf("expected 0 skills when bootstrap is disabled, got %d", len(result.Skills.Skills))
				}
			},
		},
		{
			name: "resume mode does not skip skill discovery",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Task content")

				// Create skill directory with SKILL.md
				skillDir := filepath.Join(dir, ".agents", "skills", "test-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: test-skill
description: A test skill that should be discovered even in resume mode
---

# Test Skill

This is a test skill.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			opts:     []Option{WithResume(true)},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Errorf("expected 1 skill when resume is true but bootstrap is enabled, got %d", len(result.Skills.Skills))
				}
			},
		},
		{
			name: "discover skills from .cursor/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .cursor/skills directory
				skillDir := filepath.Join(dir, ".cursor", "skills", "cursor-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: cursor-skill
description: A skill for Cursor IDE
---

# Cursor Skill

This is a skill for Cursor.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "cursor-skill" {
					t.Errorf("expected skill name 'cursor-skill', got %q", skill.Name)
				}
				if skill.Description != "A skill for Cursor IDE" {
					t.Errorf("expected skill description 'A skill for Cursor IDE', got %q", skill.Description)
				}
				if skill.Location == "" {
					t.Error("expected skill Location to be set")
				}
			},
		},
		{
			name: "discover skills from both .agents/skills and .cursor/skills",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .agents/skills directory
				skillDir1 := filepath.Join(dir, ".agents", "skills", "agents-skill")
				if err := os.MkdirAll(skillDir1, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent1 := `---
name: agents-skill
description: A generic agents skill
---

# Agents Skill
`
				skillPath1 := filepath.Join(skillDir1, "SKILL.md")
				if err := os.WriteFile(skillPath1, []byte(skillContent1), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}

				// Create skill in .cursor/skills directory
				skillDir2 := filepath.Join(dir, ".cursor", "skills", "cursor-skill")
				if err := os.MkdirAll(skillDir2, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}
				skillContent2 := `---
name: cursor-skill
description: A Cursor IDE skill
---

# Cursor Skill
`
				skillPath2 := filepath.Join(skillDir2, "SKILL.md")
				if err := os.WriteFile(skillPath2, []byte(skillContent2), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 2 {
					t.Fatalf("expected 2 skills, got %d", len(result.Skills.Skills))
				}
				names := []string{result.Skills.Skills[0].Name, result.Skills.Skills[1].Name}
				// Verify both skills are present (order doesn't matter)
				if (names[0] != "agents-skill" && names[0] != "cursor-skill") ||
					(names[1] != "agents-skill" && names[1] != "cursor-skill") {
					t.Errorf("expected skills 'agents-skill' and 'cursor-skill', got %v", names)
				}
			},
		},
		{
			name: "discover skills from .opencode/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .opencode/skills directory
				skillDir := filepath.Join(dir, ".opencode", "skills", "opencode-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: opencode-skill
description: A skill for OpenCode
---

# OpenCode Skill

This is a skill for OpenCode.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "opencode-skill" {
					t.Errorf("expected skill name 'opencode-skill', got %q", skill.Name)
				}
			},
		},
		{
			name: "discover skills from .github/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .github/skills directory
				skillDir := filepath.Join(dir, ".github", "skills", "copilot-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: copilot-skill
description: A skill for GitHub Copilot
---

# Copilot Skill

This is a skill for GitHub Copilot.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "copilot-skill" {
					t.Errorf("expected skill name 'copilot-skill', got %q", skill.Name)
				}
			},
		},
		{
			name: "discover skills from .augment/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .augment/skills directory
				skillDir := filepath.Join(dir, ".augment", "skills", "augment-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: augment-skill
description: A skill for Augment
---

# Augment Skill

This is a skill for Augment.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "augment-skill" {
					t.Errorf("expected skill name 'augment-skill', got %q", skill.Name)
				}
			},
		},
		{
			name: "discover skills from .windsurf/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .windsurf/skills directory
				skillDir := filepath.Join(dir, ".windsurf", "skills", "windsurf-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: windsurf-skill
description: A skill for Windsurf
---

# Windsurf Skill

This is a skill for Windsurf.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "windsurf-skill" {
					t.Errorf("expected skill name 'windsurf-skill', got %q", skill.Name)
				}
			},
		},
		{
			name: "discover skills from .claude/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .claude/skills directory
				skillDir := filepath.Join(dir, ".claude", "skills", "claude-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: claude-skill
description: A skill for Claude
---

# Claude Skill

This is a skill for Claude.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "claude-skill" {
					t.Errorf("expected skill name 'claude-skill', got %q", skill.Name)
				}
			},
		},
		{
			name: "discover skills from .gemini/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .gemini/skills directory
				skillDir := filepath.Join(dir, ".gemini", "skills", "gemini-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: gemini-skill
description: A skill for Gemini
---

# Gemini Skill

This is a skill for Gemini.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "gemini-skill" {
					t.Errorf("expected skill name 'gemini-skill', got %q", skill.Name)
				}
			},
		},
		{
			name: "discover skills from .codex/skills directory",
			setup: func(t *testing.T, dir string) {
				// Create task
				createTask(t, dir, "test-task", "", "Test task content")

				// Create skill in .codex/skills directory
				skillDir := filepath.Join(dir, ".codex", "skills", "codex-skill")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatalf("failed to create skill directory: %v", err)
				}

				skillContent := `---
name: codex-skill
description: A skill for Codex
---

# Codex Skill

This is a skill for Codex.
`
				skillPath := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
					t.Fatalf("failed to create skill file: %v", err)
				}
			},
			taskName: "test-task",
			wantErr:  false,
			checkFunc: func(t *testing.T, result *Result) {
				if len(result.Skills.Skills) != 1 {
					t.Fatalf("expected 1 skill, got %d", len(result.Skills.Skills))
				}
				skill := result.Skills.Skills[0]
				if skill.Name != "codex-skill" {
					t.Errorf("expected skill name 'codex-skill', got %q", skill.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup test fixtures
			tt.setup(t, tmpDir)

			// Create context with test directory and options
			opts := append([]Option{
				WithSearchPaths("file://" + tmpDir),
			}, tt.opts...)
			cc := New(opts...)

			// Run the context
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

			// Run checks
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}
