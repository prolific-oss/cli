package hook

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
)

// EventListOptions is the options for the listing events for a subscription command.
type EventListOptions struct {
	Args           []string
	SubscriptionID string
}

// NewListCommand creates a new command to deal with listing events
func NewEventListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts EventListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide a list of events for your subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderEvents(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.SubscriptionID, "subscription", "s", "", "List the events for a subscription")

	return cmd
}

// renderEvents will show your projects
func renderEvents(client client.API, opts EventListOptions, w io.Writer) error {
	if opts.SubscriptionID == "" {
		return errors.New("please provide a subscription ID")
	}

	events, err := client.GetEvents(opts.SubscriptionID)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", "ID", "Created", "Updated", "Status", "Resource ID")
	for _, event := range events.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", event.ID, event.DateCreated.Format(ui.AppDateTimeFormat), event.DateUpdated.Format(ui.AppDateTimeFormat), event.Status, event.ResourceID)
	}

	tw.Flush()

	return nil
}
