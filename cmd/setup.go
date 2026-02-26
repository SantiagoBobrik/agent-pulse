package cmd

import (
	"fmt"
	"os"

	"github.com/SantiagoBobrik/claude-pulse/internal/config"
	"github.com/SantiagoBobrik/claude-pulse/internal/hooks"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure Claude Code hooks for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}

		if err := hooks.Setup(cwd, cfg.Port); err != nil {
			return err
		}

		fmt.Println("claude-pulse hooks configured in .claude/settings.json")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
