package survey

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// ResponseDeleteAllOptions is the options for the delete all survey responses command.
type ResponseDeleteAllOptions struct {
	Args []string
}

// NewResponseDeleteAllCommand creates a new command to delete all survey responses.
func NewResponseDeleteAllCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ResponseDeleteAllOptions

	cmd := &cobra.Command{
		Use:   commandName + " <survey_id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Delete all responses for a survey",
		Long: `Delete all responses for a survey

Permanently removes all responses from the specified survey.
`,
		Example: `
Delete all responses for a survey

$ prolific survey response delete-all 6261321e223a605c7a4f7678
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := deleteAllSurveyResponses(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

func deleteAllSurveyResponses(client client.API, opts ResponseDeleteAllOptions, w io.Writer) error {
	surveyID := opts.Args[0]

	err := client.DeleteAllSurveyResponses(surveyID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Deleted all responses for survey: %s\n", surveyID)

	return nil
}
