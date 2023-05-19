package hook

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/ui"
	"github.com/spf13/cobra"
)

// EventListOptions is the options for the listing events for a subscription command.
type EventListOptions struct {
	Args           []string
	SubscriptionID string
	Limit          int
	Offset         int
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
	flags.IntVarP(&opts.Limit, "limit", "l", 1, "Limit the number of events returned")
	flags.IntVarP(&opts.Offset, "offset", "o", 0, "The number of events to offset")

	return cmd
}

// renderEvents will show your projects
func renderEvents(client client.API, opts EventListOptions, w io.Writer) error {
	if opts.SubscriptionID == "" {
		return errors.New("please provide a subscription ID")
	}

	events, err := client.GetEvents(opts.SubscriptionID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", "ID", "Created", "Updated", "Status", "Resource ID")
	for _, event := range events.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", event.ID, event.DateCreated.Format(ui.AppDateTimeFormat), event.DateUpdated.Format(ui.AppDateTimeFormat), event.Status, event.ResourceID)
	}

	_ = tw.Flush()

	eventCount := len(events.Results)
	entityName := "event"
	if eventCount > 1 {
		entityName = "events"
	}

	fmt.Fprintf(w, "\nShowing %v %s of %v\n", eventCount, entityName, events.Meta.Count)

	return nil
}
