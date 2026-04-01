package hook

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateOptions are the options for creating a hook subscription.
type CreateOptions struct {
	WorkspaceID string
	EventType   string
	TargetURL   string
}

// NewCreateSubscriptionCommand creates a new `hook create` command to create a hook subscription.
func NewCreateSubscriptionCommand(c client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a hook subscription",
		Long: `Create a subscription for an event type.

When an event is triggered in the Prolific system, the hook will automatically
notify the specified target URL. Before creating a subscription, you must ensure
that you have created a secret for your workspace.`,
		Example: `
Create a subscription for study status changes:
$ prolific hook create -w 63722982f9cc073ecc730f6b -e study.status.change -u https://example.com/api/v1/studies/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := client.CreateHookPayload{
				WorkspaceID: opts.WorkspaceID,
				EventType:   opts.EventType,
				TargetURL:   opts.TargetURL,
			}

			// The secret here is a one-time secret used for confirming the subscription received via X-Hook-Secret header,
			// and is not the same as the hook secret used for signing webhook requests.
			hook, secret, err := c.CreateHookSubscription(payload)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			confirmedHook, err := c.ConfirmHookSubscription(hook.ID, secret)
			if err != nil {
				return fmt.Errorf("subscription created (ID: %s) but confirmation failed: %s", hook.ID, err.Error())
			}

			fmt.Fprintf(w, "Subscription created successfully\n")
			fmt.Fprintf(w, "ID:           %s\n", confirmedHook.ID)
			fmt.Fprintf(w, "Event Type:   %s\n", confirmedHook.EventType)
			fmt.Fprintf(w, "Target URL:   %s\n", confirmedHook.TargetURL)
			fmt.Fprintf(w, "Enabled:      %v\n", confirmedHook.IsEnabled)
			fmt.Fprintf(w, "Workspace ID: %s\n", confirmedHook.WorkspaceID)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace to create the subscription in (required)")
	flags.StringVarP(&opts.EventType, "event-type", "e", "", "The event type to subscribe to (required)")
	flags.StringVarP(&opts.TargetURL, "target-url", "u", "", "The URL to notify when the event is triggered (required)")

	_ = cmd.MarkFlagRequired("event-type")
	_ = cmd.MarkFlagRequired("target-url")

	return cmd
}
