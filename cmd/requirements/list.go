package requirement

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/benmatselby/prolificli/client"
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

	for _, req := range reqs.Results {
		title := req.Query.Question
		if title == "" {
			title = req.Query.Title
		}

		fmt.Fprintf(w, "- %s\n", title)
	}

	return nil
}
