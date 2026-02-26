package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "claude-pulse",
	Short:   "Bridge between Claude Code and physical hardware",
	Long:    "claude-pulse forwards Claude Code lifecycle events to connected devices via WebSocket in real-time.",
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}
