package survey

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// ViewOptions is the options for the detail view of a survey.
type ViewOptions struct {
	Args []string
}

// NewViewCommand creates a new command to show a survey.
func NewViewCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ViewOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(1),
		Short: "Provide details about your survey",
		Long: `View your survey

A detailed view of how your survey is configured.
`,
		Example: `
View the details of a specific survey

$ prolific survey view 6261321e223a605c7a4f7678
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderSurvey(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

// renderSurvey will show your survey
func renderSurvey(client client.API, opts ViewOptions, w io.Writer) error {
	if len(opts.Args) < 1 || opts.Args[0] == "" {
		return errors.New("please provide a survey ID")
	}

	survey, err := client.GetSurvey(opts.Args[0])
	if err != nil {
		return err
	}

	fmt.Fprint(w, renderSurveyString(*survey))

	return nil
}
