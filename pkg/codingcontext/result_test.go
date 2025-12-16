package codingcontext

import (
	"testing"
)

func TestResult_MCPServer(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name: "no MCP server",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: "",
		},
		{
			name: "MCP server from task",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{
						MCPServer: "filesystem",
					},
				},
			},
			want: "filesystem",
		},
		{
			name: "task MCP server takes precedence over rules",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServer: "git",
						},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{
						MCPServer: "filesystem",
					},
				},
			},
			want: "filesystem",
		},
		{
			name: "rules have MCP server but task doesn't",
			result: Result{
				Rules: []Markdown[RuleFrontMatter]{
					{
						FrontMatter: RuleFrontMatter{
							MCPServer: "jira",
						},
					},
				},
				Task: Markdown[TaskFrontMatter]{
					FrontMatter: TaskFrontMatter{},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.MCPServer()

			if got != tt.want {
				t.Errorf("MCPServer() = %q, want %q", got, tt.want)
			}
		})
	}
}
