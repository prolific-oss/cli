package study

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	studyui "github.com/benmatselby/prolificli/ui/study"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// ViewOptions is the options for the detail view of a project.
type ViewOptions struct {
	Args []string
	Web  bool
}

// NewViewCommand creates a new `study view` command to give you details about
// your studies.
func NewViewCommand(client client.API, w io.Writer) *cobra.Command {
	var opts ViewOptions

	cmd := &cobra.Command{
		Use:   "view",
		Short: "Provide details about your study, requires a Study ID",
		Long:  `View study details`,
		Example: `
To get details about a study
$ prolific study view 64395e9c2332b8a59a65d51e`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.Web {
				return browser.OpenURL(studyui.GetStudyURL(opts.Args[0]))
			}

			study, err := client.GetStudy(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintln(w, studyui.RenderStudy(*study))

			return nil
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.Web, "web", "W", false, "Open the study in the web application")

	return cmd
}
