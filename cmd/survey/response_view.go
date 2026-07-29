package survey

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// ResponseViewOptions is the options for viewing a survey response.
type ResponseViewOptions struct {
	Args []string
}

// NewResponseViewCommand creates a new command to show a survey response.
func NewResponseViewCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ResponseViewOptions

	cmd := &cobra.Command{
		Use:   commandName + " <survey_id> <response_id>",
		Args:  cobra.MinimumNArgs(2),
		Short: "View a survey response",
		Long: `View a survey response

A detailed view of a participant's response to a survey.
`,
		Example: `
View the details of a specific survey response

$ prolific survey response view 6261321e223a605c7a4f7678 7372432f334b716d8b5g8789
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderSurveyResponse(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

func renderSurveyResponse(client client.API, opts ResponseViewOptions, w io.Writer) error {
	if len(opts.Args) < 2 || opts.Args[0] == "" || opts.Args[1] == "" {
		return errors.New("please provide a survey ID and response ID")
	}

	response, err := client.GetSurveyResponse(opts.Args[0], opts.Args[1])
	if err != nil {
		return err
	}

	fmt.Fprint(w, renderResponseString(*response))

	return nil
}
