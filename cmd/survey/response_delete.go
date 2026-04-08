package survey

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// ResponseDeleteOptions is the options for the delete survey response command.
type ResponseDeleteOptions struct {
	Args []string
}

// NewResponseDeleteCommand creates a new command to delete a survey response.
func NewResponseDeleteCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ResponseDeleteOptions

	cmd := &cobra.Command{
		Use:   commandName + " <survey_id> <response_id>",
		Args:  cobra.MinimumNArgs(2),
		Short: "Delete a survey response",
		Long: `Delete a survey response

Permanently removes the specified response from a survey.
`,
		Example: `
Delete a specific survey response

$ prolific survey response delete 6261321e223a605c7a4f7678 7372432f334b716d8b5g8789
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := deleteSurveyResponse(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

func deleteSurveyResponse(client client.API, opts ResponseDeleteOptions, w io.Writer) error {
	err := client.DeleteSurveyResponse(opts.Args[0], opts.Args[1])
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Deleted survey response: %s\n", opts.Args[1])

	return nil
}
