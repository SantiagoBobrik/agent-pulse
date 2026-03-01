package setup

import (
	"fmt"
	"os"

	"github.com/SantiagoBobrik/agent-pulse/internal/hooks"
	"github.com/spf13/cobra"
)

// Cmd is the setup command.
var Cmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure Claude Code hooks for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}

		if err := hooks.Setup(cwd); err != nil {
			return err
		}

		fmt.Println("agent-pulse hooks configured in .claude/settings.json")
		return nil
	},
}
