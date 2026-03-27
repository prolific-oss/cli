package filters

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui"
	"github.com/prolific-oss/cli/ui/filter"
	"github.com/spf13/cobra"
)

// ListOptions are the options for the list filters command.
type ListOptions struct {
	Args           []string
	NonInteractive bool
}

func NewListCommand(client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   "filters",
		Short: "List all filters available for your study",
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
		Example: ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			var err error
			if opts.NonInteractive {
				err = renderNonInteractive(client, w)
			} else {
				err = renderInteractive(client)
			}

			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", false, "Render the list details straight to the terminal.")

	return cmd
}

func renderInteractive(client client.API) error {
	filters, err := client.GetFilters()
	if err != nil {
		return err
	}

	var items []list.Item

	for _, filter := range filters.Results {
		items = append(items, filter)
	}

	lv := filter.ListView{
		List:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		Client: client,
	}
	lv.List.Title = "Filters"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render filters: %s", err)
	}

	return nil
}

func renderNonInteractive(client client.API, w io.Writer) error {
	filters, err := client.GetFilters()
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "Title", "FilterID", "Type", "DataType")

	for _, f := range filters.Results {
		dataType := formatDataType(f.DataType, f.Min, f.Max)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", f.Title(), f.FilterID, f.Type, dataType)
	}

	_ = tw.Flush()

	total := len(filters.Results)
	if filters.JSONAPIMeta != nil {
		total = filters.JSONAPIMeta.Meta.Count
	}
	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(filters.Results), total))

	return nil
}

// formatDataType returns the data type string, appending range bounds in
// parentheses when present, e.g. "integer (18–100)", "integer (min: 18)".
func formatDataType(dataType string, min, max any) string {
	hasMin := min != nil
	hasMax := max != nil

	if hasMin && hasMax {
		return fmt.Sprintf("%s (%v\u2013%v)", dataType, min, max)
	}

	if hasMin {
		return fmt.Sprintf("%s (min: %v)", dataType, min)
	}

	if hasMax {
		return fmt.Sprintf("%s (max: %v)", dataType, max)
	}

	return dataType
}
