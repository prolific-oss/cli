package survey

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewSurveyCommand creates a new `survey` command
func NewSurveyCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "survey",
		Short: "Manage and view your surveys",
		Long: `Manage your surveys

Surveys allow you to define screening questions for participants.
You can create surveys with questions organised into sections, or as a flat list of questions.
`,
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewViewCommand("view", client, w),
		NewCreateCommand("create", client, w),
		NewDeleteCommand("delete", client, w),
	)
	return cmd
}
