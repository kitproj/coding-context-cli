package markdown

import (
	"errors"
	"fmt"
	"testing"
)

// TestContentStartOffset_NoFrontmatter verifies that source not starting with
// "---\n" returns offset 0 (no frontmatter to skip).
func TestContentStartOffset_NoFrontmatter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"plain text", []byte("Hello world\n")},
		{"dashes without newline", []byte("---")},
		{"indented dashes", []byte("  ---\ncontent\n---\n")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := contentStartOffset(tc.input); got != 0 {
				t.Errorf("contentStartOffset(%q) = %d, want 0", tc.input, got)
			}
		})
	}
}

// TestContentStartOffset_NoClosingDelimiter verifies that source starting with
// "---\n" but lacking a closing "---" line returns 0.
func TestContentStartOffset_NoClosingDelimiter(t *testing.T) {
	t.Parallel()

	source := []byte("---\nkey: value\nmore: data\n")
	if got := contentStartOffset(source); got != 0 {
		t.Errorf("contentStartOffset(no closing ---) = %d, want 0", got)
	}
}

// TestContentStartOffset_NoNewlineAfterOpening verifies that source starting with
// "---\n" but with no subsequent newline returns 0 (loop break path).
func TestContentStartOffset_NoNewlineAfterOpening(t *testing.T) {
	t.Parallel()

	// "---\n" prefix followed by bytes with no '\n'
	source := []byte("---\nnoNewlineHere")
	if got := contentStartOffset(source); got != 0 {
		t.Errorf("contentStartOffset(no newline in body) = %d, want 0", got)
	}
}

// TestContentStartOffset_ValidFrontmatter verifies that a proper "---\n...\n---\n"
// block returns the byte offset immediately after the closing delimiter.
func TestContentStartOffset_ValidFrontmatter(t *testing.T) {
	t.Parallel()

	source := []byte("---\nkey: val\n---\nbody content\n")
	got := contentStartOffset(source)
	body := string(source[got:])

	if body != "body content\n" {
		t.Errorf("contentStartOffset: body = %q, want \"body content\\n\"", body)
	}
}

// TestContentStartOffset_EmptyFrontmatter verifies that an empty frontmatter
// block ("---\n---\n") returns the offset after the closing delimiter.
func TestContentStartOffset_EmptyFrontmatter(t *testing.T) {
	t.Parallel()

	source := []byte("---\n---\ncontent\n")
	got := contentStartOffset(source)
	body := string(source[got:])

	if body != "content\n" {
		t.Errorf("contentStartOffset(empty frontmatter): body = %q, want \"content\\n\"", body)
	}
}

// TestYamlErrorPosition_GoccyFormat verifies that errors formatted as "[line:col] msg"
// are parsed correctly by yamlErrorPosition.
func TestYamlErrorPosition_GoccyFormat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		errMsg string
		wantL  int
		wantC  int
	}{
		{
			name:   "standard goccy format",
			errMsg: "[3:7] unexpected key",
			wantL:  3,
			wantC:  7,
		},
		{
			name:   "line 1 col 1",
			errMsg: "[1:1] some error",
			wantL:  1,
			wantC:  1,
		},
		{
			name:   "multi-line error (only first line parsed)",
			errMsg: "[2:5] first line error\nsecond line context\n",
			wantL:  2,
			wantC:  5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := fmt.Errorf("%s", tc.errMsg) //nolint:err113
			l, c := yamlErrorPosition(err)

			if l != tc.wantL || c != tc.wantC {
				t.Errorf("yamlErrorPosition(%q) = (%d, %d), want (%d, %d)",
					tc.errMsg, l, c, tc.wantL, tc.wantC)
			}
		})
	}
}

// TestYamlErrorPosition_YamlV2Fallback verifies that errors formatted as
// "yaml: line N: msg" (goldmark-meta style) are parsed via the fallback path.
func TestYamlErrorPosition_YamlV2Fallback(t *testing.T) {
	t.Parallel()

	err := errors.New("yaml: line 4: some yaml error") //nolint:err113
	l, c := yamlErrorPosition(err)

	if l != 4 {
		t.Errorf("yamlErrorPosition(yaml v2 format) line = %d, want 4", l)
	}

	if c != 0 {
		t.Errorf("yamlErrorPosition(yaml v2 format) col = %d, want 0", c)
	}
}

// TestYamlErrorPosition_UnknownFormat verifies that an unrecognized error string
// returns (0, 0) without panicking.
func TestYamlErrorPosition_UnknownFormat(t *testing.T) {
	t.Parallel()

	err := errors.New("some unrelated error with no position info") //nolint:err113
	l, c := yamlErrorPosition(err)

	if l != 0 || c != 0 {
		t.Errorf("yamlErrorPosition(unknown) = (%d, %d), want (0, 0)", l, c)
	}
}
