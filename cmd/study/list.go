package study

import (
	"fmt"
	"io"
	"os"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// NewListCommand creates a new `study list` command to give you details about
// your studies.
func NewListCommand(client client.API) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Provide details about your studies",
		Run: func(cmd *cobra.Command, args []string) {

			err := renderList(client, os.Stdout)
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func renderList(client client.API, w io.Writer) error {
	studies, err := client.GetStudies()
	if err != nil {
		return err
	}

	var items []list.Item

	for _, study := range studies.Results {
		items = append(items, study)
	}

	lv := ui.ListView{List: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	lv.List.Title = "My studies"

	p := tea.NewProgram(lv, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		return fmt.Errorf("cannot render studies: %s", err)
	}

	return nil
}
