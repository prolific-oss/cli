package audience

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewAudienceCommand creates a new `audience` command
func NewAudienceCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audience",
		Short: "Understand your audience of eligible participants",
	}

	cmd.AddCommand(
		NewBreakdownCommand(client, w),
	)

	return cmd
}
