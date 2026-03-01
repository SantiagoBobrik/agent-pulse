package clientcmd

import "github.com/spf13/cobra"

// Cmd is the parent command for all client subcommands.
var Cmd = &cobra.Command{
	Use:   "client",
	Short: "Manage registered event clients",
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(removeCmd)
}
