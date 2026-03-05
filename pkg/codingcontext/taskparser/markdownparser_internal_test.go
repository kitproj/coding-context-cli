package taskparser

import "testing"

// TestBodyOffset_NoPrefixReturnsZero verifies that content without a "---\n"
// prefix is treated as having no frontmatter (offset 0).
func TestBodyOffset_NoPrefixReturnsZero(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"plain content", "Hello world\n"},
		{"dashes without newline", "---"},
		{"dashes-only line followed by content", "---\nfoo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// For "---\nfoo" — starts with "---\n" but has no closing ---
			// For others — no "---\n" prefix at all.
		})
	}

	// Plain content: no "---\n" prefix → offset must be 0
	if got := bodyOffset([]byte("plain content\n")); got != 0 {
		t.Errorf("bodyOffset(plain) = %d, want 0", got)
	}

	// Empty source → offset 0
	if got := bodyOffset([]byte("")); got != 0 {
		t.Errorf("bodyOffset(empty) = %d, want 0", got)
	}
}

// TestBodyOffset_NoClosingDelimiter verifies that source starting with "---\n"
// but lacking a closing "---" returns 0 (no valid frontmatter block).
func TestBodyOffset_NoClosingDelimiter(t *testing.T) {
	t.Parallel()

	// Source has opening "---\n" and content lines but no closing "---"
	source := []byte("---\nkey: value\nmore: data\n")
	got := bodyOffset(source)

	if got != 0 {
		t.Errorf("bodyOffset(no closing ---) = %d, want 0", got)
	}
}

// TestBodyOffset_NoNewlineAfterOpening verifies that source starting with "---\n"
// followed by content without any subsequent newline causes the loop to break
// (IndexByte returns -1) and returns 0.
func TestBodyOffset_NoNewlineAfterOpening(t *testing.T) {
	t.Parallel()

	// "---\n" prefix, but the remaining bytes "noNewline" contain no '\n'
	source := []byte("---\nnoNewline")
	got := bodyOffset(source)

	if got != 0 {
		t.Errorf("bodyOffset(no newline after opening) = %d, want 0", got)
	}
}

// TestBodyOffset_ValidFrontmatter verifies that a proper frontmatter block
// returns the byte offset immediately after the closing "---" line.
func TestBodyOffset_ValidFrontmatter(t *testing.T) {
	t.Parallel()

	// "---\nkey: val\n---\nbody\n"
	// Offset should point to 'b' in "body"
	source := []byte("---\nkey: val\n---\nbody\n")
	got := bodyOffset(source)

	body := string(source[got:])
	if body != "body\n" {
		t.Errorf("bodyOffset(valid frontmatter): body = %q, want \"body\\n\"", body)
	}
}

// TestBodyOffset_EmptyFrontmatter verifies that an empty frontmatter block
// (just "---\n---\n") returns the offset after the second delimiter.
func TestBodyOffset_EmptyFrontmatter(t *testing.T) {
	t.Parallel()

	source := []byte("---\n---\ncontent\n")
	got := bodyOffset(source)

	body := string(source[got:])
	if body != "content\n" {
		t.Errorf("bodyOffset(empty frontmatter): body = %q, want \"content\\n\"", body)
	}
}
