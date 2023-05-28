package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	studyui "github.com/prolific-oss/cli/ui/study"
	"github.com/spf13/cobra"
)

// NewViewCommand creates a new `study view` command to give you details about
// your studies.
func NewViewCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Provide details about your study, requires a Study ID",
		Long:  `View study details`,
		Example: `
To get details about a study
$ prolific study view 64395e9c2332b8a59a65d51e`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			study, err := client.GetStudy(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintln(w, studyui.RenderStudy(*study))

			return nil
		},
	}

	return cmd
}
