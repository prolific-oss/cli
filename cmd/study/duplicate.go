package study

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewDuplicateCommand creates a new `study duplicate` command to duplicate
// an existing study.
func NewDuplicateCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "duplicate",
		Short: "Duplicate an existing study",
		Long: `Duplicate an existing study

This may be useful if you have a templated study in the web application. If you
are mainly using the CLI, you can define a JSON/YAML file to pass into the "study
create" command`,
		Example: `
To duplicate a study, you need the ID
$ prolific study duplicate 64395e9c2332b8a59a65d51e`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			study, err := client.DuplicateStudy(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintln(w, study.ID)

			return nil
		},
	}

	return cmd
}
