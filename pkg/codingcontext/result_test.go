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
		want   map[string]mcp.MCPServerConfig
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
			want: map[string]mcp.MCPServerConfig{},
		},
		{
			name: "MCP servers from rules with URNs",
			result: Result{
				Name: "test-task",
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:jira-server"},
							MCPServer:       mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "jira"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:api-server"},
							MCPServer:       mcp.MCPServerConfig{Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: map[string]mcp.MCPServerConfig{
				"urn:agents:rule:jira-server": {Type: mcp.TransportTypeStdio, Command: "jira"},
				"urn:agents:rule:api-server":  {Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"},
			},
		},
		{
			name: "multiple rules with MCP servers and empty rule",
			result: Result{
				Name: "test-task",
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:server1"},
							MCPServer:       mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server1"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:server2"},
							MCPServer:       mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server2"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:empty"},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: map[string]mcp.MCPServerConfig{
				"urn:agents:rule:server1": {Type: mcp.TransportTypeStdio, Command: "server1"},
				"urn:agents:rule:server2": {Type: mcp.TransportTypeStdio, Command: "server2"},
				// Empty rule MCP server is filtered out
			},
		},
		{
			name: "rule without MCP server",
			result: Result{
				Name: "test-task",
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:no-server"},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: map[string]mcp.MCPServerConfig{
				// Empty rule MCP server is filtered out
			},
		},
		{
			name: "mixed rules with URNs",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:explicit"},
							MCPServer:       mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server1"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:some-rule"},
							MCPServer:       mcp.MCPServerConfig{Type: mcp.TransportTypeStdio, Command: "server2"},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							BaseFrontMatter: markdown.BaseFrontMatter{URN: "urn:agents:rule:another"},
							MCPServer:       mcp.MCPServerConfig{Type: mcp.TransportTypeHTTP, URL: "https://example.com"},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: map[string]mcp.MCPServerConfig{
				"urn:agents:rule:explicit":  {Type: mcp.TransportTypeStdio, Command: "server1"},
				"urn:agents:rule:some-rule": {Type: mcp.TransportTypeStdio, Command: "server2"},
				"urn:agents:rule:another":   {Type: mcp.TransportTypeHTTP, URL: "https://example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// Helper function to get map keys for debugging
func mapKeys(m map[string]mcp.MCPServerConfig) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
