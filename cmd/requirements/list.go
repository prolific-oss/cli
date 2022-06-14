package requirement

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui/requirement"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// NewListCommand creates a new `requirement list` command to give you details about
// eligibility requirements.
func NewListCommand(client client.API) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all eligibility requirements",
		Run: func(cmd *cobra.Command, args []string) {

			err := renderList(client, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %s", strings.ReplaceAll(err.Error(), "\n", ""))
				os.Exit(1)
			}
		},
	}

	return cmd
}

func renderList(client client.API, w io.Writer) error {
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

	if err := tea.NewProgram(lv).Start(); err != nil {
		return fmt.Errorf("cannot render requirements: %s", err)
	}

	return nil
}
