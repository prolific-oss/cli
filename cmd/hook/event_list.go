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
	Limit          int
	Offset         int
}

// NewListCommand creates a new command to deal with listing events
func NewEventListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts EventListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide a list of events for your subscription",
		Long: `List all events sent to your subscription

If you have a subscription for a Prolific Platform event, we will deliver a
payload to your target URL. We save this audit record. This means you can query
the Prolific Platform to get events for a given subscription. This maybe be
useful for reconciliation or testing.
		`,
		Example: `
Get the last 200 events for the 637e081185389c0ca5595915 subscription
$ prolific hook events -s 637e081185389c0ca5595915

You can also use the standard limit and offset parameters. This will get you
the last 10 events for your subscription.
$ prolific hook events -s 637e081185389c0ca5595915 -l 10

This will offset by 10 events, and get the next 10 events.
$ prolific hook events -s 637e081185389c0ca5595915 -l 10 -o 10
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderEvents(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.SubscriptionID, "subscription", "s", "", "List the events for a subscription")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of events returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of events to offset")

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

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(events.Results), events.Meta.Count))

	return nil
}
