package main

import (
	"strings"
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			expected: 0,
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 0,
		},
		{
			name:     "one empty string",
			s1:       "hello",
			s2:       "",
			expected: 5,
		},
		{
			name:     "single character difference",
			s1:       "hello",
			s2:       "hallo",
			expected: 1,
		},
		{
			name:     "multiple differences",
			s1:       "kitten",
			s2:       "sitting",
			expected: 3,
		},
		{
			name:     "completely different",
			s1:       "abc",
			s2:       "xyz",
			expected: 3,
		},
		{
			name:     "insertion",
			s1:       "cat",
			s2:       "cats",
			expected: 1,
		},
		{
			name:     "deletion",
			s1:       "cats",
			s2:       "cat",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := levenshteinDistance(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

func TestHashContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "simple content",
			content: "Hello World",
		},
		{
			name:    "multiline content",
			content: "Line 1\nLine 2\nLine 3",
		},
		{
			name:    "empty content",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hashContent(tt.content)
			hash2 := hashContent(tt.content)

			// Same content should produce same hash
			if hash1 != hash2 {
				t.Errorf("hashContent produced different hashes for same content")
			}

			// Hash should be 64 characters (SHA256 in hex)
			if len(hash1) != 64 {
				t.Errorf("hashContent produced hash of length %d, want 64", len(hash1))
			}

			// Different content should produce different hash
			if tt.content != "" {
				differentHash := hashContent(tt.content + "different")
				if hash1 == differentHash {
					t.Errorf("hashContent produced same hash for different content")
				}
			}
		})
	}
}

func TestSplitIntoChunks(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedCount int
		checkFirst    bool
		firstHeading  string
		firstContent  string
	}{
		{
			name: "single chunk with heading",
			content: `# Introduction
This is the introduction.
It has multiple lines.
`,
			expectedCount: 1,
			checkFirst:    true,
			firstHeading:  "# Introduction",
			firstContent:  "This is the introduction.\nIt has multiple lines.\n",
		},
		{
			name: "multiple chunks",
			content: `# First Section
Content of first section.

# Second Section
Content of second section.
`,
			expectedCount: 2,
			checkFirst:    true,
			firstHeading:  "# First Section",
			firstContent:  "Content of first section.\n",
		},
		{
			name: "content without heading",
			content: `Just some content
without any headings.
`,
			expectedCount: 1,
			checkFirst:    true,
			firstHeading:  "",
			firstContent:  "Just some content\nwithout any headings.\n",
		},
		{
			name: "H2 headings should not split",
			content: `# Main Heading
Content here.

## Subheading
More content.

# Another Main Heading
Final content.
`,
			expectedCount: 2,
			checkFirst:    true,
			firstHeading:  "# Main Heading",
		},
		{
			name:          "empty content",
			content:       "",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := splitIntoChunks(tt.content)

			if len(chunks) != tt.expectedCount {
				t.Errorf("splitIntoChunks() returned %d chunks, want %d", len(chunks), tt.expectedCount)
			}

			if tt.checkFirst && len(chunks) > 0 {
				if chunks[0].heading != tt.firstHeading {
					t.Errorf("first chunk heading = %q, want %q", chunks[0].heading, tt.firstHeading)
				}
				if tt.firstContent != "" && chunks[0].content != tt.firstContent {
					t.Errorf("first chunk content = %q, want %q", chunks[0].content, tt.firstContent)
				}
			}
		})
	}
}

func TestRuleDeduplicator_ExactDuplicates(t *testing.T) {
	dedup := NewRuleDeduplicator()

	content1 := `# Test Rule
This is a test rule.
`

	// First addition should succeed
	result1 := dedup.AddContent(content1)
	if result1 == "" {
		t.Errorf("First addition should not be deduplicated")
	}

	// Second addition of same content should be deduplicated
	result2 := dedup.AddContent(content1)
	if result2 != "" {
		t.Errorf("Second addition of same content should be deduplicated, got: %q", result2)
	}
}

func TestRuleDeduplicator_DifferentContent(t *testing.T) {
	dedup := NewRuleDeduplicator()

	content1 := `# Test Rule 1
This is test rule one.
`

	content2 := `# Test Rule 2
This is test rule two, completely different.
`

	// Both should be added
	result1 := dedup.AddContent(content1)
	if result1 == "" {
		t.Errorf("First unique content should not be deduplicated")
	}

	result2 := dedup.AddContent(content2)
	if result2 == "" {
		t.Errorf("Second unique content should not be deduplicated")
	}
}

func TestRuleDeduplicator_SimilarChunks(t *testing.T) {
	dedup := NewRuleDeduplicator()

	// First content
	content1 := `# Coding Standards
- Use tabs for indentation
- Write tests for all functions
- Use meaningful variable names
`

	// Very similar content (small typo changes)
	content2 := `# Coding Standards
- Use tabs for indentations
- Write tests for all function
- Use meaningful variable name
`

	// First addition should succeed
	result1 := dedup.AddContent(content1)
	if result1 == "" {
		t.Errorf("First addition should not be deduplicated")
	}

	// Second addition should be deduplicated as similar
	result2 := dedup.AddContent(content2)
	if result2 != "" {
		t.Errorf("Similar content should be deduplicated, got: %q", result2)
	}
}

func TestRuleDeduplicator_MultipleChunks(t *testing.T) {
	dedup := NewRuleDeduplicator()

	content1 := `# Section A
Content A here.

# Section B
Content B here.
`

	content2 := `# Section A
Content A here.

# Section C
Content C here - completely different.
`

	// Add first content
	result1 := dedup.AddContent(content1)
	if result1 == "" {
		t.Errorf("First content should not be deduplicated")
	}

	// Add second content - Section A should be deduplicated, Section C should remain
	result2 := dedup.AddContent(content2)

	// Result should only contain Section C
	if result2 == "" {
		t.Errorf("Second content should have at least Section C")
	}

	// Check that Section A is not in result2
	if strings.Contains(result2, "Section A") {
		t.Errorf("Section A should be deduplicated from result2")
	}

	// Check that Section C is in result2
	if !strings.Contains(result2, "Section C") {
		t.Errorf("Section C should be in result2, got: %q", result2)
	}
}

func TestRuleDeduplicator_ExactChunkDuplicates(t *testing.T) {
	dedup := NewRuleDeduplicator()

	content1 := `# Duplicated Section
This is duplicated content.
`

	content2 := `# Different Heading
Other content here.

# Duplicated Section
This is duplicated content.
`

	// Add first content
	result1 := dedup.AddContent(content1)
	if result1 == "" {
		t.Errorf("First content should not be deduplicated")
	}

	// Add second content - "Duplicated Section" should be removed
	result2 := dedup.AddContent(content2)

	// Result should only contain "Different Heading" section
	if !strings.Contains(result2, "Different Heading") {
		t.Errorf("Result should contain 'Different Heading', got: %q", result2)
	}

	// Result should not contain "Duplicated Section"
	if strings.Contains(result2, "Duplicated Section") {
		t.Errorf("Result should not contain 'Duplicated Section', got: %q", result2)
	}
}

func TestRuleDeduplicator_ChunkWithoutHeading(t *testing.T) {
	dedup := NewRuleDeduplicator()

	content1 := `Some content without heading.
Just plain text.
`

	content2 := `Some content without heading.
Just plain text.
`

	// Add first content
	result1 := dedup.AddContent(content1)
	if result1 == "" {
		t.Errorf("First content should not be deduplicated")
	}

	// Add same content again - should be deduplicated
	result2 := dedup.AddContent(content2)
	if result2 != "" {
		t.Errorf("Duplicate content should be deduplicated, got: %q", result2)
	}
}
