package filters

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewFiltersCommand creates the `filters` parent command. It still lists the
// catalogue when it is called without a subcommand, because that is what
// `prolific filters` did before the subcommands existed.
func NewFiltersCommand(client client.API, w io.Writer) *cobra.Command {
	var detail bool

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
  a given participant attribute.

Run without a subcommand, this lists the catalogue, the same as
` + "`prolific filters list`" + `.`,
		Example: `
List all filters
$ prolific filters list

Search for filters by keyword
$ prolific filters search "software developer"

List the choices belonging to a filter, to get the choice IDs to select
$ prolific filters choices current-job-role

Search within a filter's choices
$ prolific filters choices search current-job-role nurse`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if detail {
				if err := renderFilterDetails(client, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}

				return nil
			}

			if err := renderList(client, ListOptions{}, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	cmd.AddCommand(
		NewListCommand(client, w),
		NewSearchCommand(client, w),
		NewChoicesCommand(client, w),
		NewRuleTreeCommand(client, w),
	)

	// `prolific filters -n` printed every filter's details before the command
	// gained subcommands. It keeps working, hidden, so the old invocation is
	// not broken and the help still points at `filters list` instead.
	cmd.Flags().BoolVarP(&detail, "non-interactive", "n", false, "Render the filter details straight to the terminal.")
	_ = cmd.Flags().MarkHidden("non-interactive")

	return cmd
}

// renderFilterDetails writes a detail block per filter, which is the output
// `prolific filters -n` has always produced.
func renderFilterDetails(c client.API, w io.Writer) error {
	filters, err := c.GetFilters("")
	if err != nil {
		return err
	}

	for _, f := range filters.Results {
		if _, err := fmt.Fprintln(w, RenderFilter(f)); err != nil {
			return err
		}
	}

	return nil
}
