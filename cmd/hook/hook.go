package hook

import (
	"fmt"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
	"github.com/SantiagoBobrik/agent-pulse/internal/events"
	"github.com/spf13/cobra"
)

var (
	flagProvider string
	flagEvent    string
)

var Cmd = &cobra.Command{
	Use:   "hook",
	Short: "Handle agent lifecycle hooks",
	RunE:  runHook,
}

func init() {
	Cmd.Flags().StringVar(&flagProvider, "provider", "", "agent provider (e.g. claude, gemini)")
	Cmd.Flags().StringVar(&flagEvent, "event", "", "lifecycle event name")
	Cmd.MarkFlagRequired("provider")
	Cmd.MarkFlagRequired("event")
}

func runHook(cmd *cobra.Command, args []string) error {
	provider := domain.Provider(flagProvider)
	if !provider.IsValid() {
		return fmt.Errorf("unknown provider %q", flagProvider)
	}
	event := domain.EventType(flagEvent)
	if !event.IsValidFor(provider) {
		return fmt.Errorf("unknown event %q for provider %q", flagEvent, flagProvider)
	}
	return events.HandleEvent(provider, event)
}
