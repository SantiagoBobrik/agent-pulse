package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := config.Dir()
		if err != nil {
			return err
		}
		logPath := filepath.Join(dir, "server.log")

		if _, err := os.Stat(logPath); err != nil {
			return fmt.Errorf("no log file found at %s", logPath)
		}

		tailArgs := []string{"-n", "50"}
		if followFlag {
			tailArgs = append(tailArgs, "-f")
		}
		tailArgs = append(tailArgs, logPath)

		tail := exec.Command("tail", tailArgs...)
		tail.Stdout = os.Stdout
		tail.Stderr = os.Stderr
		return tail.Run()
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "follow log output")
}
