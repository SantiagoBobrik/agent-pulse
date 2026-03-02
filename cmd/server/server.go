package server

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the event bridge server",
}

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(downCmd)
	Cmd.AddCommand(logsCmd)
}
