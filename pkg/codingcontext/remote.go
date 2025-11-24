package codingcontext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	getter "github.com/hashicorp/go-getter/v2"
)

// downloadPath downloads or copies a path (local or remote) using go-getter
// and returns the local path where it was downloaded/copied
func downloadPath(ctx context.Context, src string) (string, error) {
	// Create a temporary directory for the download
	tmpBase, err := os.MkdirTemp("", "coding-context-remote-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// go-getter requires the destination to not exist, so create a subdirectory
	tmpDir := filepath.Join(tmpBase, "download")

	// Use go-getter to download the directory
	_, err = getter.Get(ctx, tmpDir, src)
	if err != nil {
		os.RemoveAll(tmpBase)
		return "", fmt.Errorf("failed to download from %s: %w", src, err)
	}

	return tmpDir, nil
}
