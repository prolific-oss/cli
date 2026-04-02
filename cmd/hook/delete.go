package hook

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewDeleteSubscriptionCommand creates a new `hook delete` command to delete a hook subscription.
func NewDeleteSubscriptionCommand(c client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <subscription-id>",
		Short: "Delete a hook subscription",
		Long: `Delete an existing hook subscription.

If you want to temporarily pause notifications instead of permanently deleting
the subscription, use the 'hook update --disable' command.`,
		Example: `
Delete a subscription:
$ prolific hook delete 6261321e223a605c7a4f7564`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subscriptionID := args[0]

			if err := c.DeleteHookSubscription(subscriptionID); err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintf(w, "Subscription %s deleted successfully\n", subscriptionID)

			return nil
		},
	}

	return cmd
}
