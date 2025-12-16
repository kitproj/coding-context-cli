package codingcontext

import (
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/mcp"
)

func TestResult_MCPServers(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   mcp.MCPServerConfigs
	}{
		{
			name: "no MCP servers",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: mcp.MCPServerConfigs{},
		},
		{
			name: "MCP servers from task only",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{
						MCPServers: mcp.MCPServerConfigs{
							"filesystem": {Type: mcp.TransportTypeStdio, Command: "filesystem"},
							"git":        {Type: mcp.TransportTypeStdio, Command: "git"},
						},
					},
				},
			},
			want: mcp.MCPServerConfigs{
				"filesystem": {Type: mcp.TransportTypeStdio, Command: "filesystem"},
				"git":        {Type: mcp.TransportTypeStdio, Command: "git"},
			},
		},
		{
			name: "MCP servers from rules only",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServers: mcp.MCPServerConfigs{
								"jira": {Type: mcp.TransportTypeStdio, Command: "jira"},
							},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServers: mcp.MCPServerConfigs{
								"api": {Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"},
							},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{},
				},
			},
			want: mcp.MCPServerConfigs{
				"jira": {Type: mcp.TransportTypeStdio, Command: "jira"},
				"api":  {Type: mcp.TransportTypeHTTP, URL: "https://api.example.com"},
			},
		},
		{
			name: "MCP servers from both task and rules",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServers: mcp.MCPServerConfigs{
								"jira": {Type: mcp.TransportTypeStdio, Command: "jira"},
							},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{
						MCPServers: mcp.MCPServerConfigs{
							"filesystem": {Type: mcp.TransportTypeStdio, Command: "filesystem"},
						},
					},
				},
			},
			want: mcp.MCPServerConfigs{
				"filesystem": {Type: mcp.TransportTypeStdio, Command: "filesystem"},
				"jira":       {Type: mcp.TransportTypeStdio, Command: "jira"},
			},
		},
		{
			name: "multiple rules with MCP servers",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServers: mcp.MCPServerConfigs{
								"server1": {Type: mcp.TransportTypeStdio, Command: "server1"},
							},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServers: mcp.MCPServerConfigs{
								"server2": {Type: mcp.TransportTypeStdio, Command: "server2"},
							},
						},
					},
					{
						FrontMatter: markdown.RuleFrontMatter{},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{
						MCPServers: mcp.MCPServerConfigs{
							"task-server": {Type: mcp.TransportTypeStdio, Command: "task-server"},
						},
					},
				},
			},
			want: mcp.MCPServerConfigs{
				"task-server": {Type: mcp.TransportTypeStdio, Command: "task-server"},
				"server1":     {Type: mcp.TransportTypeStdio, Command: "server1"},
				"server2":     {Type: mcp.TransportTypeStdio, Command: "server2"},
			},
		},
		{
			name: "task overrides rule server with same name",
			result: Result{
				Rules: []markdown.Markdown[markdown.RuleFrontMatter]{
					{
						FrontMatter: markdown.RuleFrontMatter{
							MCPServers: mcp.MCPServerConfigs{
								"filesystem": {Type: mcp.TransportTypeStdio, Command: "rule-filesystem"},
							},
						},
					},
				},
				Task: markdown.Markdown[markdown.TaskFrontMatter]{
					FrontMatter: markdown.TaskFrontMatter{
						MCPServers: mcp.MCPServerConfigs{
							"filesystem": {Type: mcp.TransportTypeStdio, Command: "task-filesystem"},
						},
					},
				},
			},
			want: mcp.MCPServerConfigs{
				"filesystem": {Type: mcp.TransportTypeStdio, Command: "task-filesystem"},
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

			for name, wantServer := range tt.want {
				gotServer, exists := got[name]
				if !exists {
					t.Errorf("MCPServers() missing server %q", name)
					continue
				}

				if gotServer.Type != wantServer.Type {
					t.Errorf("MCPServers()[%q].Type = %v, want %v", name, gotServer.Type, wantServer.Type)
				}
				if gotServer.Command != wantServer.Command {
					t.Errorf("MCPServers()[%q].Command = %q, want %q", name, gotServer.Command, wantServer.Command)
				}
				if gotServer.URL != wantServer.URL {
					t.Errorf("MCPServers()[%q].URL = %q, want %q", name, gotServer.URL, wantServer.URL)
				}
			}
		})
	}
}
