package survey

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// DeleteOptions is the options for the delete survey command.
type DeleteOptions struct {
	Args []string
}

// NewDeleteCommand creates a new command to delete a survey.
func NewDeleteCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts DeleteOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(1),
		Short: "Delete a survey",
		Long: `Delete a survey

Permanently removes the specified survey.
`,
		Example: `
Delete a specific survey

$ prolific survey delete 6261321e223a605c7a4f7678
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := deleteSurvey(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

func deleteSurvey(client client.API, opts DeleteOptions, w io.Writer) error {
	err := client.DeleteSurvey(opts.Args[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Deleted survey: %s\n", opts.Args[0])

	return nil
}
