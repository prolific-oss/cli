package filtersets

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewFilterSetCommand creates a new `filters` command
func NewFilterSetCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter-sets",
		Short: "Manage and view your filter sets",
		Long: `List your filter sets

Filters allow you to restrict access to your study based on participant
demographics and attributes.

Combine these filters into filter sets which allow you to re-use preset filter
settings across multiple studies.

Filters are broadly found in two distinct types:
- A select type filter allows you to select one or more options from a list of pre-defined choices.
- A range type filter allows you to select an upper and / or a lower bound for a given participant attribute.
`,
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
	)
	return cmd
}
