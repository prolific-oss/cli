package filters

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui/filter"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func NewListCommand(client client.API, w io.Writer) *cobra.Command {
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
			err := renderList(client)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

func renderList(client client.API) error {
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
