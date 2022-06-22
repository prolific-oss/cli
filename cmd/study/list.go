package study

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui/study"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing studies command.
type ListOptions struct {
	Args   []string
	Status string
}

// NewListCommand creates a new `study list` command to give you details about
// your studies.
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your studies",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Args = args

			err := renderList(client, opts, w)
			if err != nil {
				fmt.Printf("Error: %s", strings.ReplaceAll(err.Error(), "\n", ""))
				os.Exit(1)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Status, "status", "s", model.StatusAll, fmt.Sprintf("The status you want to filter on: %s.", strings.Join(model.StudyListStatus, ", ")))

	return cmd
}

func renderList(client client.API, opts ListOptions, w io.Writer) error {
	studies, err := client.GetStudies(opts.Status)
	if err != nil {
		return err
	}

	var items []list.Item
	var studyMap = make(map[string]model.Study)

	for _, study := range studies.Results {
		items = append(items, study)
		studyMap[study.ID] = study
	}

	lv := study.ListView{
		List:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		Studies: studyMap,
		Client:  client,
	}
	lv.List.Title = "My studies"

	if err := tea.NewProgram(lv).Start(); err != nil {
		return fmt.Errorf("cannot render studies: %s", err)
	}

	return nil
}
