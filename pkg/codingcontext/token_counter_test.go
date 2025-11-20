package codingcontext

import "testing"

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "empty string",
			text: "",
			want: 0,
		},
		{
			name: "short text",
			text: "Hi",
			want: 0, // 2 chars / 4 = 0
		},
		{
			name: "simple sentence",
			text: "This is a test.",
			want: 3, // 15 chars / 4 = 3
		},
		{
			name: "paragraph",
			text: "This is a longer paragraph with multiple words that should result in more tokens being counted by our estimation algorithm.",
			want: 30, // 123 chars / 4 = 30
		},
		{
			name: "multiline text",
			text: `This is line one.
This is line two.
This is line three.`,
			want: 13, // 55 chars / 4 = 13
		},
		{
			name: "code snippet",
			text: `func main() {
    fmt.Println("Hello, World!")
}`,
			want: 12, // 49 chars / 4 = 12
		},
		{
			name: "markdown with frontmatter",
			text: `---
title: Test
---
# Heading

This is content.`,
			want: 11, // 47 chars / 4 = 11
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateTokens(tt.text)
			if got != tt.want {
				t.Errorf("estimateTokens() = %d, want %d", got, tt.want)
			}
		})
	}
}
