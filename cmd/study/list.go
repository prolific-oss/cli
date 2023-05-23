package study

import (
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/model"
	"github.com/prolific-oss/prolificli/ui/study"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing studies command.
type ListOptions struct {
	Args           []string
	Csv            bool
	Fields         string
	NonInteractive bool
	ProjectID      string
	Status         string
}

// NewListCommand creates a new `study list` command to give you details about
// your studies.
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "List all of your studies",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			renderer := study.ListRenderer{}

			if opts.Csv {
				renderer.SetStrategy(&study.CsvRenderer{})
			} else if opts.NonInteractive {
				renderer.SetStrategy(&study.NonInteractiveRenderer{})
			} else {
				renderer.SetStrategy(&study.InteractiveRenderer{})
			}

			err := renderer.Render(client, study.ListUsedOptions{
				Status: opts.Status, NonInteractive: opts.NonInteractive, Fields: opts.Fields, ProjectID: opts.ProjectID,
			}, w)

			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Status, "status", "s", model.StatusAll, fmt.Sprintf("The status you want to filter on: %s.", strings.Join(model.StudyListStatus, ", ")))
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", false, "Render the list details straight to the terminal.")
	flags.BoolVarP(&opts.Csv, "csv", "c", false, "Render the list details in a CSV format.")
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in non-interactive/csv mode.")
	flags.StringVarP(&opts.ProjectID, "project", "p", "", "Get studies for a given project ID.")

	return cmd
}
