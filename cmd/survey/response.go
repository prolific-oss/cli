package survey

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewResponseCommand creates a new `survey response` parent command
func NewResponseCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "response",
		Short: "Manage survey responses",
		Long: `Manage responses to your surveys

Survey responses contain participant answers to your survey questions.
You can list, view, create, delete individual responses, or delete all responses for a survey.
`,
	}

	cmd.AddCommand(
		NewResponseListCommand("list", client, w),
		NewResponseViewCommand("view", client, w),
		NewResponseCreateCommand("create", client, w),
		NewResponseDeleteCommand("delete", client, w),
		NewResponseDeleteAllCommand("delete-all", client, w),
		NewResponseSummaryCommand("summary", client, w),
	)
	return cmd
}
