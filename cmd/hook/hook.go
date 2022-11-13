package hook

import (
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewHookCommand creates a new `hook` command
func NewHookCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage and view your hook subscriptions",
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewEventTypeCommand("event-types", client, w),
		NewListSecretCommand("secrets", client, w),
	)
	return cmd
}
