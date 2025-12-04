package codingcontext

import (
	"testing"
)

func TestResult_MCPServers(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   MCPServerConfigs
	}{
		{
			name: "no MCP servers",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: MCPServerConfigs{},
		},
		{
			name: "MCP servers from task only",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{
						MCPServers: MCPServerConfigs{
							"filesystem": {Type: TransportTypeStdio, Command: "filesystem"},
							"git":        {Type: TransportTypeStdio, Command: "git"},
						},
					},
				},
			},
			want: MCPServerConfigs{
				"filesystem": {Type: TransportTypeStdio, Command: "filesystem"},
				"git":        {Type: TransportTypeStdio, Command: "git"},
			},
		},
		{
			name: "MCP servers from rules only",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServers: MCPServerConfigs{
								"jira": {Type: TransportTypeStdio, Command: "jira"},
							},
						},
					},
					{
						FrontMatter: RuleFrontMatter{
							MCPServers: MCPServerConfigs{
								"api": {Type: TransportTypeHTTP, URL: "https://api.example.com"},
							},
						},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: MCPServerConfigs{
				"jira": {Type: TransportTypeStdio, Command: "jira"},
				"api":  {Type: TransportTypeHTTP, URL: "https://api.example.com"},
			},
		},
		{
			name: "MCP servers from both task and rules",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServers: MCPServerConfigs{
								"jira": {Type: TransportTypeStdio, Command: "jira"},
							},
						},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{
						MCPServers: MCPServerConfigs{
							"filesystem": {Type: TransportTypeStdio, Command: "filesystem"},
						},
					},
				},
			},
			want: MCPServerConfigs{
				"filesystem": {Type: TransportTypeStdio, Command: "filesystem"},
				"jira":       {Type: TransportTypeStdio, Command: "jira"},
			},
		},
		{
			name: "multiple rules with MCP servers",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServers: MCPServerConfigs{
								"server1": {Type: TransportTypeStdio, Command: "server1"},
							},
						},
					},
					{
						FrontMatter: RuleFrontMatter{
							MCPServers: MCPServerConfigs{
								"server2": {Type: TransportTypeStdio, Command: "server2"},
							},
						},
					},
					{
						FrontMatter: RuleFrontMatter{},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{
						MCPServers: MCPServerConfigs{
							"task-server": {Type: TransportTypeStdio, Command: "task-server"},
						},
					},
				},
			},
			want: MCPServerConfigs{
				"task-server": {Type: TransportTypeStdio, Command: "task-server"},
				"server1":     {Type: TransportTypeStdio, Command: "server1"},
				"server2":     {Type: TransportTypeStdio, Command: "server2"},
			},
		},
		{
			name: "task overrides rule server with same name",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServers: MCPServerConfigs{
								"filesystem": {Type: TransportTypeStdio, Command: "rule-filesystem"},
							},
						},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{
						MCPServers: MCPServerConfigs{
							"filesystem": {Type: TransportTypeStdio, Command: "task-filesystem"},
						},
					},
				},
			},
			want: MCPServerConfigs{
				"filesystem": {Type: TransportTypeStdio, Command: "task-filesystem"},
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
