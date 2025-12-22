package codingcontext

import (
	"testing"
)

func TestResult_MCPServers(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   []MCPServerConfig
	}{
		{
			name: "no MCP servers",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: []MCPServerConfig{},
		},
		{
			name: "MCP servers from rules only",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServer: MCPServerConfig{Type: TransportTypeStdio, Command: "jira"},
						},
					},
					{
						FrontMatter: RuleFrontMatter{
							MCPServer: MCPServerConfig{Type: TransportTypeHTTP, URL: "https://api.example.com"},
						},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: []MCPServerConfig{
				{Type: TransportTypeStdio, Command: "jira"},
				{Type: TransportTypeHTTP, URL: "https://api.example.com"},
			},
		},
		{
			name: "multiple rules with MCP servers",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServer: MCPServerConfig{Type: TransportTypeStdio, Command: "server1"},
						},
					},
					{
						FrontMatter: RuleFrontMatter{
							MCPServer: MCPServerConfig{Type: TransportTypeStdio, Command: "server2"},
						},
					},
					{
						FrontMatter: RuleFrontMatter{},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: []MCPServerConfig{
				{Type: TransportTypeStdio, Command: "server1"},
				{Type: TransportTypeStdio, Command: "server2"},
				{}, // Empty rule MCP server
			},
		},
		{
			name: "rule without MCP server",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: []MCPServerConfig{
				{}, // Empty rule MCP server
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
