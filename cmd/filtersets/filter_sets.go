package filtersets

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/config"
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
		NewViewCommand("view", client, w),
	)
	return cmd
}

// GetFilterSetPath returns the URL path to a filter set, agnostic of domain
func GetFilterSetPath(workspaceID, FilterSetID string) string {
	return fmt.Sprintf("researcher/workspaces/%s/screener-sets/%s", workspaceID, FilterSetID)
}

// GetFilterSetURL returns the full URL to a filter set using configuration
func GetFilterSetURL(workspaceID, FilterSetID string) string {
	return fmt.Sprintf("%s/%s", config.GetApplicationURL(), GetFilterSetPath(workspaceID, FilterSetID))
}
