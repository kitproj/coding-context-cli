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
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		taskName string
		want     string
	}{
		{
			name: "task only without rules",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "test-task", "task_name: test-task", "Task content\n")
			},
			taskName: "test-task",
			want:     "Task content\n",
		},
		{
			name: "single rule and task",
			setup: func(t *testing.T, dir string) {
				createTask(t, dir, "test-task", "task_name: test-task", "Task content\n")
				createRule(t, dir, ".agents/rules/rule1.md", "", "Rule 1 content\n")
			},
			taskName: "test-task",
			want:     "Rule 1 content\n\nTask content\n",
		},
		{
			name: "multiple rules and task",
			setup: func(t *testing.T, dir string) {
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

func TestResult_MCPServers(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   []mcp.MCPServerConfig
	}{
		{
			name: "no MCP servers",
			result: Result{
				Name:  "test-task",
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: []mcp.MCPServerConfig{},
		},
		{
			name: "MCP servers from rules only",
			result: Result{
				Name: "test-task",
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServer: mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "jira"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServer: mcp.MCPServerConfig{Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: []mcp.MCPServerConfig{
				{Type: mcp.TransportTypeStdio, Command: "jira"},
				{Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"},
			},
		},
		{
			name: "multiple rules with MCP servers and empty rule",
			result: Result{
				Name: "test-task",
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServer: mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server1"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServer: mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server2"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: []mcp.MCPServerConfig{
				{Type: mcp.TransportTypeStdio, Command: "server1"},
				{Type: mcp.TransportTypeStdio, Command: "server2"},
				// Empty rule MCP server is filtered out
			},
		},
		{
			name: "rule without MCP server",
			result: Result{
				Name: "test-task",
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: []mcp.MCPServerConfig{
				// Empty rule MCP server is filtered out
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.MCPServers()

			if len(got) != len(tt.want) {
				t.Errorf("MCPServers() returned %d servers, want %d", len(got), len(tt.want))
				return
			}

			for i, wantServer := range tt.want {
				gotServer := got[i]

				if gotServer.Type != wantServer.Type {
					t.Errorf("MCPServers()[%d].Type = %v, want %v", i, gotServer.Type, wantServer.Type)
				}
				if gotServer.Command != wantServer.Command {
					t.Errorf("MCPServers()[%d].Command = %q, want %q", i, gotServer.Command, wantServer.Command)
				}
				if gotServer.URL != wantServer.URL {
					t.Errorf("MCPServers()[%d].URL = %q, want %q", i, gotServer.URL, wantServer.URL)
				}
			}
		})
	}
}
