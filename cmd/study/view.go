package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/prolificli/client"
	studyui "github.com/prolific-oss/prolificli/ui/study"
	"github.com/spf13/cobra"
)

// NewViewCommand creates a new `study view` command to give you details about
// your studies.
func NewViewCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Provide details about your study, requires a Study ID",
		Args:  cobra.MinimumNArgs(1),
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
