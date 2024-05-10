package requirement

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui/requirement"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// NewListCommand creates a new `requirement list` command to give you details about
// eligibility requirements.
func NewListCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "requirements",
		Short: "List all eligibility requirements available for your study",
		Long: `List eligibility requirements to filter participants for your study

When you run a study, you can decide who is eligible from Prolific's pool of
participants. These requirements are called eligibility requirements. From the
list view you can press enter to get more details about that requirement. Those
details can then be used in the "study create" command.`,
		Example: `
To list all the requirements
$ prolific requirements

This will provide you with an interactive list you can search. Once you have picked
a requirement, press enter to get more details`,
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
	reqs, err := client.GetEligibilityRequirements()
	if err != nil {
		return err
	}

	var items []list.Item
	var reqMap = make(map[string]model.Requirement)

	for _, req := range reqs.Results {
		items = append(items, req)
		reqMap[req.ID] = req
	}

	lv := requirement.ListView{
		List:         list.New(items, list.NewDefaultDelegate(), 0, 0),
		Requirements: reqMap,
		Client:       client,
	}
	lv.List.Title = "Eligibility requirements"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render studies: %s", err)
	}

	return nil
}
