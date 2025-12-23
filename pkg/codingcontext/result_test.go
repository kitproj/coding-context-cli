package codingcontext

import (
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

func TestResult_Prompt(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name: "empty result",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					Content: "Task content",
				},
				Prompt: "Task content",
			},
			want: "Task content",
		},
		{
			name: "single rule and task",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						Content: "Rule 1 content",
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					Content: "Task content",
				},
				Prompt: "Rule 1 content\nTask content",
			},
			want: "Rule 1 content\nTask content",
		},
		{
			name: "multiple rules and task",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						Content: "Rule 1 content",
					},
					{
						Content: "Rule 2 content",
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					Content: "Task content",
				},
				Prompt: "Rule 1 content\nRule 2 content\nTask content",
			},
			want: "Rule 1 content\nRule 2 content\nTask content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Prompt != tt.want {
				t.Errorf("Result.Prompt = %q, want %q", tt.result.Prompt, tt.want)
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
