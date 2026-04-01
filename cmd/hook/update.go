package hook

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// UpdateOptions are the options for updating a hook subscription.
type UpdateOptions struct {
	EventType string
	TargetURL string
	Enable    bool
	Disable   bool
}

// NewUpdateSubscriptionCommand creates a new `hook update` command to update a hook subscription.
func NewUpdateSubscriptionCommand(c client.API, w io.Writer) *cobra.Command {
	var opts UpdateOptions

	cmd := &cobra.Command{
		Use:   "update <subscription-id>",
		Short: "Update a hook subscription",
		Long: `Update an existing hook subscription.

You can update the event type, target URL, or enable/disable the subscription.
All fields are optional; only the flags you provide will be updated.`,
		Example: `
Disable a subscription:
$ prolific hook update sub-id-123 --disable

Update the target URL:
$ prolific hook update sub-id-123 -u https://example.com/api/v2/hook/

Update event type and re-enable:
$ prolific hook update sub-id-123 -e study.status.change --enable`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Enable && opts.Disable {
				return errors.New("--enable and --disable are mutually exclusive")
			}

			subscriptionID := args[0]
			payload := client.UpdateHookPayload{}

			if cmd.Flags().Changed("event-type") {
				payload.EventType = &opts.EventType
			}
			if cmd.Flags().Changed("target-url") {
				payload.TargetURL = &opts.TargetURL
			}
			if opts.Enable {
				v := true
				payload.IsEnabled = &v
			} else if opts.Disable {
				v := false
				payload.IsEnabled = &v
			}

			hook, err := c.UpdateHookSubscription(subscriptionID, payload)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintf(w, "Subscription updated successfully\n")
			fmt.Fprintf(w, "ID:           %s\n", hook.ID)
			fmt.Fprintf(w, "Event Type:   %s\n", hook.EventType)
			fmt.Fprintf(w, "Target URL:   %s\n", hook.TargetURL)
			fmt.Fprintf(w, "Enabled:      %v\n", hook.IsEnabled)
			fmt.Fprintf(w, "Workspace ID: %s\n", hook.WorkspaceID)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.EventType, "event-type", "e", "", "The event type to subscribe to")
	flags.StringVarP(&opts.TargetURL, "target-url", "u", "", "The URL to notify when the event is triggered")
	flags.BoolVar(&opts.Enable, "enable", false, "Enable the subscription")
	flags.BoolVar(&opts.Disable, "disable", false, "Disable the subscription")

	return cmd
}
