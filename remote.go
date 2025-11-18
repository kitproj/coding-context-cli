package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	getter "github.com/hashicorp/go-getter/v2"
)

func (cc *codingContext) downloadRemoteDirectories(ctx context.Context) error {
	for _, remotePath := range cc.remotePaths {
		cc.logger.Info("Downloading remote directory", "path", remotePath)
		localPath, err := downloadRemoteDirectory(ctx, remotePath)
		if err != nil {
			return fmt.Errorf("failed to download remote directory %s: %w", remotePath, err)
		}
		cc.downloadedDirs = append(cc.downloadedDirs, localPath)
		cc.logger.Info("Downloaded to", "path", localPath)
	}

	return nil
}

func (cc *codingContext) cleanupDownloadedDirectories() {
	for _, dir := range cc.downloadedDirs {
		if dir == "" {
			continue
		}

		if err := os.RemoveAll(dir); err != nil {
			cc.logger.Error("Error cleaning up downloaded directory", "path", dir, "error", err)
		}
	}
}

// downloadRemoteDirectory downloads a remote directory using go-getter
// and returns the local path where it was downloaded
func downloadRemoteDirectory(ctx context.Context, src string) (string, error) {
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
