package clientcmd

import (
	"fmt"
	"strings"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		found := -1
		for i, c := range cfg.Clients {
			if strings.EqualFold(c.Name, name) {
				found = i
				break
			}
		}

		if found == -1 {
			return fmt.Errorf("client %q not found", name)
		}

		cfg.Clients = append(cfg.Clients[:found], cfg.Clients[found+1:]...)

		if err := config.Save(cfg); err != nil {
			return err
		}

		logger.Info("client removed", "name", name)
		return nil
	},
}
