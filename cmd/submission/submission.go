package submission

import (
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewSubmissionCommand creates a new `submission` command
func NewSubmissionCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submission",
		Short: "Manage and view your study submissions",
		Long: `Manage study submissions

A published study will have submissions taken by the Prolific Participants. These
commands allow you to manage those submissions.
`,
	}

	cmd.AddCommand(
		NewListCommand(client, w),
	)
	return cmd
}
