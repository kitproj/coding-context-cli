package main

import (
	stdcontext "context"
	"fmt"
	"os"

	"github.com/kitproj/coding-context-cli/pkg/context"
)

func (cc *codingContext) downloadRemoteDirectories(ctx stdcontext.Context) error {
	for _, remotePath := range cc.remotePaths {
		fmt.Fprintf(cc.logOut, "⪢ Downloading remote directory: %s\n", remotePath)
		localPath, err := context.DownloadRemoteDirectory(ctx, remotePath)
		if err != nil {
			return fmt.Errorf("failed to download remote directory %s: %w", remotePath, err)
		}
		cc.downloadedDirs = append(cc.downloadedDirs, localPath)
		fmt.Fprintf(cc.logOut, "⪢ Downloaded to: %s\n", localPath)
	}

	return nil
}

func (cc *codingContext) cleanupDownloadedDirectories() {
	for _, dir := range cc.downloadedDirs {
		if dir == "" {
			continue
		}

		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(cc.logOut, "⪢ Error cleaning up downloaded directory %s: %v\n", dir, err)
		}
	}
}
