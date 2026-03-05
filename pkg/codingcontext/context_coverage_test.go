package codingcontext

import (
	"encoding/hex"
	"testing"
)

// TestDownloadDir_Deterministic verifies that downloadDir produces the same output
// for the same input (SHA-256 is deterministic) and different outputs for different inputs.
func TestDownloadDir_Deterministic(t *testing.T) {
	t.Parallel()

	path := "github.com/example/repo//some/path"
	got1 := downloadDir(path)
	got2 := downloadDir(path)

	if got1 != got2 {
		t.Errorf("downloadDir is not deterministic: %q != %q", got1, got2)
	}

	other := downloadDir("github.com/other/repo")
	if got1 == other {
		t.Errorf("downloadDir should produce different results for different inputs")
	}
}

// TestDownloadDir_ContainsHex verifies that the last component of the path
// returned by downloadDir is a valid hex string (SHA-256 hash encoded).
func TestDownloadDir_ContainsHex(t *testing.T) {
	t.Parallel()

	result := downloadDir("https://example.com/repo")
	// filepath.Base of result should be a hex-encoded 32-byte SHA-256 hash (64 hex chars)
	base := result[len(result)-64:]
	if _, err := hex.DecodeString(base); err != nil {
		t.Errorf("downloadDir last 64 chars %q is not valid hex: %v", base, err)
	}
}
