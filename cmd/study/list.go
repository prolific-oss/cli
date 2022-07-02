package study

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/benmatselby/prolificli/ui/study"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing studies command.
type ListOptions struct {
	Args           []string
	Status         string
	NonInteractive bool
}

// NewListCommand creates a new `study list` command to give you details about
// your studies.
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your studies",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Args = args

			renderer := study.ListRenderer{}

			if opts.NonInteractive {
				renderer.SetStrategy(&study.NonInteractiveRenderer{})
			} else {
				renderer.SetStrategy(&study.InteractiveRenderer{})
			}

			err := renderer.Render(client, study.ListUsedOptions{
				Status: opts.Status, NonInteractive: opts.NonInteractive,
			}, w)

			if err != nil {
				fmt.Printf("Error: %s", strings.ReplaceAll(err.Error(), "\n", ""))
				os.Exit(1)
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Status, "status", "s", model.StatusAll, fmt.Sprintf("The status you want to filter on: %s.", strings.Join(model.StudyListStatus, ", ")))
	flags.BoolVarP(&opts.NonInteractive, "non-interactive", "n", false, "Render the list details straight to the terminal.")

	return cmd
}
