package filters

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewListCommand creates the `filters list` command.
func NewListCommand(client client.API, w io.Writer) *cobra.Command {
	var nonInteractive bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all filters available for your study",
		Long: `List every filter in the catalogue.

Use this when you want to browse the full set of filters. To find filters by
keyword, use ` + "`prolific filters search`" + ` instead.`,
		Example: `
List all filters in an interactive, searchable interface
$ prolific filters list

List all filters in a non-interactive format for scripting or AI agents
$ prolific filters list -n`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if nonInteractive {
				err = renderNonInteractiveList(client, w)
			} else {
				err = renderInteractiveList(client)
			}
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&nonInteractive, "non-interactive", "n", false, "Render the filter details straight to the terminal.")

	return cmd
}

func renderNonInteractiveList(client client.API, w io.Writer) error {
	filters, err := client.GetFilters()
	if err != nil {
		return err
	}

	for _, f := range filters.Results {
		fmt.Fprintln(w, RenderFilter(f))
	}

	return nil
}

func renderInteractiveList(client client.API) error {
	filters, err := client.GetFilters()
	if err != nil {
		return err
	}

	var items []list.Item

	for _, f := range filters.Results {
		items = append(items, f)
	}

	lv := ListView{
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
