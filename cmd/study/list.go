package study

import (
	"fmt"
	"io"
	"os"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewListCommand creates a new `study list` command to give you details about
// your studies.
func NewListCommand(client client.API) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Provide details about your studies",
		Run: func(cmd *cobra.Command, args []string) {

			err := renderList(client, os.Stdout)
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func renderList(client client.API, w io.Writer) error {
	studies, err := client.GetStudies()
	if err != nil {
		return err
	}

	for _, study := range studies.Results {
		fmt.Fprintf(w, "%s - %s - %s\n", study.ID, study.Status, study.Name)
	}

	return nil
}
