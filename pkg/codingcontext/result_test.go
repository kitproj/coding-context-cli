package codingcontext

import (
	"testing"
)

func TestMarkdown_BootstrapPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "md file",
			path: "/path/to/task.md",
			want: "/path/to/task-bootstrap",
		},
		{
			name: "mdc file",
			path: "/path/to/rule.mdc",
			want: "/path/to/rule-bootstrap",
		},
		{
			name: "empty path",
			path: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Markdown[FrontMatter]{Path: tt.path}
			got := m.BootstrapPath()
			if got != tt.want {
				t.Errorf("BootstrapPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
