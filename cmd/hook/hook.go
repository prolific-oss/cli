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
		Short: "Hook related commands",
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
	)
	return cmd
}
