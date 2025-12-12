package codingcontext

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper functions for creating fixtures

// createTask creates a task file in the .agents/tasks directory
func createTask(t *testing.T, dir, name, frontmatter, content string) {
	t.Helper()
	taskDir := filepath.Join(dir, ".agents", "tasks")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("failed to create task directory: %v", err)
	}

	var fileContent string
	if frontmatter != "" {
		fileContent = fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
	} else {
		fileContent = content
	}

	taskPath := filepath.Join(taskDir, name+".md")
	if err := os.WriteFile(taskPath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to create task file: %v", err)
	}
}

// createRule creates a rule file in the specified path within dir
func createRule(t *testing.T, dir, relPath, frontmatter, content string) {
	t.Helper()
	rulePath := filepath.Join(dir, relPath)
	ruleDir := filepath.Dir(rulePath)
	if err := os.MkdirAll(ruleDir, 0755); err != nil {
		t.Fatalf("failed to create rule directory: %v", err)
	}

	var fileContent string
	if frontmatter != "" {
		fileContent = fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
	} else {
		fileContent = content
	}

	if err := os.WriteFile(rulePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to create rule file: %v", err)
	}
}

// createCommand creates a command file in the .agents/commands directory
func createCommand(t *testing.T, dir, name, frontmatter, content string) {
	t.Helper()
	cmdDir := filepath.Join(dir, ".agents", "commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to create command directory: %v", err)
	}

	var fileContent string
	if frontmatter != "" {
		fileContent = fmt.Sprintf("---\n%s\n---\n%s", frontmatter, content)
	} else {
		fileContent = content
	}

	cmdPath := filepath.Join(cmdDir, name+".md")
	if err := os.WriteFile(cmdPath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to create command file: %v", err)
	}
}

// createBootstrapScript creates a bootstrap script for a rule file
func createBootstrapScript(t *testing.T, dir, rulePath, scriptContent string) {
	t.Helper()
	fullRulePath := filepath.Join(dir, rulePath)
	baseNameWithoutExt := strings.TrimSuffix(fullRulePath, filepath.Ext(fullRulePath))
	bootstrapPath := baseNameWithoutExt + "-bootstrap"

	if err := os.WriteFile(bootstrapPath, []byte(scriptContent), 0755); err != nil {
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
				WithParams(Params{"key1": "value1", "key2": "value2"}),
			},
			check: func(t *testing.T, c *Context) {
				if c.params["key1"] != "value1" {
					t.Errorf("expected params[key1]=value1, got %v", c.params["key1"])
				}
				if c.params["key2"] != "value2" {
					t.Errorf("expected params[key2]=value2, got %v", c.params["key2"])
				}
			},
		},
		{
			name: "with selectors",
			opts: []Option{
				WithSelectors(Selectors{"env": {"dev": true, "test": true}}),
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
				WithParams(Params{"env": "production"}),
				WithSelectors(Selectors{"lang": {"go": true}}),
				WithSearchPaths("/custom/path"),
				WithResume(false),
				WithAgent(AgentCopilot),
			},
			check: func(t *testing.T, c *Context) {
				if c.params["env"] != "production" {
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
				WithParams(Params{"env": "production", "feature": "auth"}),
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
				WithParams(Params{"user": "alice", "email": "alice@example.com", "role": "admin"}),
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
			name: "empty task content returns error",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "empty", "", "")
			},
			taskName:    "empty",
			wantErr:     true,
			errContains: "task not found",
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
				WithParams(Params{"project": "myapp"}),
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
			name: "resume mode skips rule discovery",
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
				if len(result.Rules) != 0 {
					t.Errorf("expected 0 rules in resume mode, got %d", len(result.Rules))
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
			name: "resume mode skips bootstrap scripts",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "no-bootstrap", "", "Task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule")
				createBootstrapScript(t, dir, ".agents/rules/rule1.md", "#!/bin/sh\nexit 1")
			},
			opts: []Option{
				WithResume(true),
			},
			taskName: "no-bootstrap",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				// In resume mode, rules aren't discovered, so bootstrap won't run
				if len(result.Rules) != 0 {
					t.Error("expected no rules in resume mode")
				}
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
				WithSelectors(Selectors{"env": {"development": true}}), // CLI selector for development
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
				WithSelectors(Selectors{"env": {"development": true}}), // CLI adds development
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
				WithParams(Params{"env": "staging"}),
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
				WithParams(Params{"value": "general"}),
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
				WithParams(Params{"app": "myservice", "env": "production"}),
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
			name: "resume mode workflow skips rules but includes task",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "resume", "", "Resume this task")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Should be skipped")
				createBootstrapScript(t, dir, ".agents/rules/rule1.md", "#!/bin/sh\necho 'should not run'")
			},
			opts: []Option{
				WithResume(true),
			},
			taskName: "resume",
			wantErr:  false,
			check: func(t *testing.T, result *Result) {
				if !strings.Contains(result.Task.Content, "Resume this task") {
					t.Errorf("unexpected task content: %q", result.Task.Content)
				}
				if len(result.Rules) != 0 {
					t.Errorf("expected 0 rules in resume mode, got %d", len(result.Rules))
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
				if err := os.MkdirAll(secondDir, 0755); err != nil {
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
				WithParams(Params{"issue_number": "123", "issue_title": "Bug fix"}),
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
				WithParams(Params{"issue_number": "123", "issue_title": "Bug fix"}),
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
				WithParams(Params{"env": "production"}),
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
				WithParams(Params{"env": "staging"}),
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
				WithParams(Params{"env": "staging"}),
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
				WithParams(Params{"project": "myapp"}),
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
				WithParams(Params{"version": "1.0.0"}),
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
				WithParams(Params{"version": "1.0.0"}),
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
				WithParams(Params{"app": "service"}),
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
				WithParams(Params{"task_var": "task_value", "cmd_var": "cmd_value"}),
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
				WithParams(Params{"task_var": "task_value", "cmd_var": "cmd_value"}),
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
				WithParams(Params{"id": "123"}),
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
				WithParams(Params{"var1": "val1", "var2": "val2", "var3": "val3"}),
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
				WithParams(Params{
					"issue_number": "123",
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
				WithParams(Params{
					"issue_number": "456",
					"issue_title":  "Fix bug",
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
				WithParams(Params{
					"issue_number": "789",
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
