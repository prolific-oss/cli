package researcher

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewResearcherCommand creates a new `researcher` command
func NewResearcherCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "researcher",
		Short: "Manage researcher resources",
		Long:  `Manage researcher resources such as test participants.`,
	}

	cmd.AddCommand(
		NewCreateParticipantCommand(client, w),
	)

	return cmd
}
