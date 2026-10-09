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
		Long: `Understand the audience of participants eligible for a study, without
creating the study or saving a filter set.

"count" totals how many participants match a set of filters. "breakdown"
does the same but splits the result by the values of one additional
filter. Both take filters either as a -p/--template-path file or directly
via --filters (see "prolific study create --help" for the filter format).`,
		Example: `  # Count participants matching a set of filters
  prolific audience count -p /path/to/filters.json -w <workspace-id>

  # Count, split by the values of one filter
  prolific audience breakdown -p /path/to/filters.json -w <workspace-id>`,
	}

	cmd.AddCommand(
		NewBreakdownCommand(client, w),
		NewCountCommand(client, w),
	)

	return cmd
}
