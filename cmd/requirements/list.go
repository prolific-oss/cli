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
