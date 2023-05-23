package hook

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewEventTypeCommand creates a new `hook event-types` command to give you details about
// your which events you can register subscriptions for.
func NewEventTypeCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   commandName,
		Short: "List of event types you can subscribe to",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := renderEventTypes(client, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

// renderEventTypes will show all of the event types that can be registered.
func renderEventTypes(client client.API, w io.Writer) error {
	eventTypes, err := client.GetHookEventTypes()
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "Event Type", "Description")

	for _, event := range eventTypes.Results {
		fmt.Fprintf(tw, "%s\t%s\n", event.EventType, event.Description)
	}

	return tw.Flush()
}
