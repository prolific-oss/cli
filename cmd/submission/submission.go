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
		Short: "Submission related commands",
	}

	cmd.AddCommand(
		NewListCommand(client, w),
	)
	return cmd
}
