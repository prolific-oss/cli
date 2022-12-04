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
		Args:  cobra.MinimumNArgs(1),
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
