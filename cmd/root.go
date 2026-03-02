package cmd

import (
	clientcmd "github.com/SantiagoBobrik/agent-pulse/cmd/client"
	"github.com/SantiagoBobrik/agent-pulse/cmd/hook"
	"github.com/SantiagoBobrik/agent-pulse/cmd/server"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "agent-pulse",
	Short:   "Bridge between AI agents and the outside world",
	Long:    "agent-pulse captures agent lifecycle events and distributes them to any number of registered clients — physical devices, webhooks, scripts, or any HTTP endpoint.",
	Version: version,
}

func init() {
	rootCmd.AddCommand(hook.Cmd)
	rootCmd.AddCommand(clientcmd.Cmd)
	rootCmd.AddCommand(server.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}
