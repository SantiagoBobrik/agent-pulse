package clientcmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
	"github.com/spf13/cobra"
)

var (
	flagName    string
	flagURL     string
	flagPort    string
	flagTimeout string
	flagEvents  string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Register a new event client",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var c *client.Client

		if flagName != "" && flagURL != "" {
			c, err = buildClientFromFlags()
			if err != nil {
				return err
			}
		} else {
			c, err = client.RunWizard(os.Stdin, os.Stdout)
			if err != nil {
				return err
			}
		}

		if !client.IsUniqueName(c.Name, cfg.Clients) {
			return fmt.Errorf("client %q already exists. Edit ~/.config/agent-pulse/config.yaml to modify it", c.Name)
		}

		cfg.Clients = append(cfg.Clients, *c)

		if err := config.Save(cfg); err != nil {
			return err
		}

		logger.Info("client registered", "name", c.Name, "url", c.URL)
		logger.Info("to set timeout or headers, edit ~/.config/agent-pulse/config.yaml")
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&flagName, "name", "", "client name")
	addCmd.Flags().StringVar(&flagURL, "url", "", "client URL or IP")
	addCmd.Flags().StringVar(&flagPort, "port", "", "client port (default 80)")
	addCmd.Flags().StringVar(&flagTimeout, "timeout", "", "delivery timeout in ms (default 2000)")
	addCmd.Flags().StringVar(&flagEvents, "events", "", "comma-separated event types or 'all'")
}

func buildClientFromFlags() (*client.Client, error) {
	url := flagURL
	if !strings.Contains(url, "://") {
		url = "http://" + url
	}
	if flagPort != "" && flagPort != "80" {
		url = url + ":" + flagPort
	}

	timeout := 2000
	if flagTimeout != "" {
		parsed, err := strconv.Atoi(flagTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout %q: must be milliseconds", flagTimeout)
		}
		timeout = parsed
	}

	var events []string
	if flagEvents != "" && flagEvents != "all" {
		events = strings.Split(flagEvents, ",")
		for i := range events {
			events[i] = strings.TrimSpace(events[i])
		}
	}

	c := &client.Client{
		Name:    flagName,
		URL:     url,
		Timeout: timeout,
		Events:  events,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}
