package participantgroup

import (
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewParticipantCommand creates a new `participant` command
func NewParticipantCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "participant",
		Short: "Manage and view your participant groups",
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
	)
	return cmd
}
