package filters

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewFiltersCommand creates the `filters` parent command.
func NewFiltersCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filters",
		Short: "Browse and search the filters available for your study",
		Long: `Filters allow you to restrict access to your study based on
participant demographics and attributes.

You can save combinations of filters, known as filter sets, to re-use across
studies. These are useful if you're running multiple studies with the same
audience filters.

There are two types of filters:

- A select type filter allows you to select one or more options from a list of
  pre-defined choices.
- A range type filter allows you to select an upper and / or a lower bound for
  a given participant attribute.`,
		Example: `
List all filters
$ prolific filters list

Search for filters by keyword
$ prolific filters search "software developer"`,
	}

	cmd.AddCommand(
		NewListCommand(client, w),
		NewSearchCommand(client, w),
	)

	return cmd
}
