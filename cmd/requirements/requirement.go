package requirement

import (
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewRequirementCommand creates a new `requirement` command
func NewRequirementCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "requirement",
		Short: "View all the requirements you can add to your studies",
	}

	cmd.AddCommand(
		NewListCommand(client, w),
	)
	return cmd
}
