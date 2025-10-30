package cmd

import (
	"fmt"
	"os"

	"github.com/kitproj/coding-agent-context-cli/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "coding-agent-context-cli",
	Short: "CLI tool for coding agent context management",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("coding-agent-context-cli version %s\n", version.Version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
