package server

import (
	"fmt"
	"syscall"

	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
	"github.com/SantiagoBobrik/agent-pulse/internal/pid"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the running server",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := pid.Read()
		if err != nil {
			return fmt.Errorf("no running server found: %w", err)
		}

		if err := syscall.Kill(p, syscall.SIGTERM); err != nil {
			pid.Remove()
			return fmt.Errorf("failed to stop server (pid %d): %w", p, err)
		}

		pid.Remove()
		logger.Info("server stopped", "pid", p)
		return nil
	},
}
