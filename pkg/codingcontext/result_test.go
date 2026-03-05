package codingcontext

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

func TestResult_Prompt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		taskName string
		want     string
	}{
		{
			name: "task only without rules",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "test-task", "task_name: test-task", "Task content\n")
			},
			taskName: "test-task",
			want:     "Task content\n",
		},
		{
			name: "single rule and task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "test-task", "task_name: test-task", "Task content\n")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule 1 content\n")
			},
			taskName: "test-task",
			want:     "Rule 1 content\n\nTask content\n",
		},
		{
			name: "multiple rules and task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "test-task", "task_name: test-task", "Task content\n")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule 1 content\n")
				createRule(t, dir, ".agents/rules/rule2.md", "", "Rule 2 content\n")
			},
			taskName: "test-task",
			want:     "Rule 1 content\n\nRule 2 content\n\nTask content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			ctx := New(
				WithSearchPaths("file://"+tmpDir),
				WithLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))),
			)

			result, err := ctx.Run(context.Background(), tt.taskName)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			if result.Name != tt.taskName {
				t.Errorf("Result.Name = %q, want %q", result.Name, tt.taskName)
			}

			if result.Prompt != tt.want {
				t.Errorf("Result.Prompt = %q, want %q", result.Prompt, tt.want)
			}
		})
	}
}

func rule(cfg mcp.MCPServerConfig) markdown.Markdown[markdown.RuleFrontMatter] {
	return markdown.Markdown[markdown.RuleFrontMatter]{
		FrontMatter: markdown.RuleFrontMatter{MCPServer: cfg},
	}
}

func emptyRule() markdown.Markdown[markdown.RuleFrontMatter] {
	return markdown.Markdown[markdown.RuleFrontMatter]{FrontMatter: markdown.RuleFrontMatter{}}
}

func resultWithRules(name string, rules ...markdown.Markdown[markdown.RuleFrontMatter]) Result {
	return Result{Name: name, Rules: rules, Task: markdown.Markdown[markdown.TaskFrontMatter]{}}
}

func mcpServersCases() []struct {
	name   string
	result Result
	want   map[string]mcp.MCPServerConfig
} {
	return []struct {
		name   string
		result Result
		want   map[string]mcp.MCPServerConfig
	}{
		{
			name:   "no MCP servers",
			result: resultWithRules("test-task"),
			want:   map[string]mcp.MCPServerConfig{},
		},
		{
			name: "MCP servers from rules with URNs",
			result: resultWithRules("test-task",
				rule(mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "jira"}),
				rule(mcp.MCPServerConfig{Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"}),
			),
			want: map[string]mcp.MCPServerConfig{
				"": {Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"},
			},
		},
		{
			name: "multiple rules with MCP servers and empty rule",
			result: resultWithRules("test-task",
				rule(mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server1"}),
				rule(mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server2"}),
				emptyRule(),
			),
			want: map[string]mcp.MCPServerConfig{
				"": {Type: mcp.TransportTypeStdio, Command: "server2"},
			},
		},
		{
			name:   "rule without MCP server",
			result: resultWithRules("test-task", emptyRule()),
			want:   map[string]mcp.MCPServerConfig{},
		},
		{
			name: "mixed rules with URNs",
			result: resultWithRules("",
				rule(mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server1"}),
				rule(mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server2"}),
				rule(mcp.MCPServerConfig{Type: mcp.TransportTypeHTTP, URL: "https://example.com"}),
			),
			want: map[string]mcp.MCPServerConfig{
				"": {Type: mcp.TransportTypeHTTP, URL: "https://example.com"},
			},
		},
	}
}

func TestResult_MCPServers(t *testing.T) {
	t.Parallel()

	for _, tt := range mcpServersCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.result.MCPServers()

			if len(got) != len(tt.want) {
				t.Errorf("MCPServers() returned %d servers, want %d", len(got), len(tt.want))
				t.Logf("Got keys: %v", mapKeys(got))
				t.Logf("Want keys: %v", mapKeys(tt.want))

				return
			}

			for key, wantServer := range tt.want {
				gotServer, ok := got[key]
				if !ok {
					t.Errorf("MCPServers() missing key %q", key)

					continue
				}

				if gotServer.Type != wantServer.Type {
					t.Errorf("MCPServers()[%q].Type = %v, want %v", key, gotServer.Type, wantServer.Type)
				}

				if gotServer.Command != wantServer.Command {
					t.Errorf("MCPServers()[%q].Command = %q, want %q", key, gotServer.Command, wantServer.Command)
				}

				if gotServer.URL != wantServer.URL {
					t.Errorf("MCPServers()[%q].URL = %q, want %q", key, gotServer.URL, wantServer.URL)
				}
			}
		})
	}
}

// Helper function to get map keys for debugging.
func mapKeys(m map[string]mcp.MCPServerConfig) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
