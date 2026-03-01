package clientcmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered clients",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Clients) == 0 {
			fmt.Println("No clients registered. Run 'agent-pulse client add' to register one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL\tTIMEOUT\tEVENTS")
		for _, c := range cfg.Clients {
			events := "all"
			if len(c.Events) > 0 {
				events = strings.Join(c.Events, ",")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.URL, c.Timeout, events)
		}
		w.Flush()

		return nil
	},
}
