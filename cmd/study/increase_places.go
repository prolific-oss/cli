package study

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	studyui "github.com/benmatselby/prolificli/ui/study"
	"github.com/spf13/cobra"
)

// IncreasePlacesOptions represents the options for the increase-places command.
type IncreasePlacesOptions struct {
	Args   []string
	Places int
}

// NewIncreasePlacesCommand creates a new `study increase-places` command to
// allow you to increase the places on a study.
func NewIncreasePlacesCommand(client client.API, w io.Writer) *cobra.Command {
	var opts IncreasePlacesOptions

	cmd := &cobra.Command{
		Use: "increase-places",
		Long: `Increase the places on your study.

You can only increase places on your study, not decrease. This is helpful if you
run a trial study with a smaller group of participants, and then want to expand
to a wider audience.`,
		Example: `
$ prolific study increase-places 64395e9c2332b8a59a65d51e -p 300
$ prolific study increase-places 64395e9c2332b8a59a65d51e --places 5000`,
		Short: "Increase the total available places on a study",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			study, err := client.GetStudy(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			if study.TotalAvailablePlaces > opts.Places {
				return fmt.Errorf("study currently has %v places, and you cannot decrease the available places to %v", study.TotalAvailablePlaces, opts.Places)
			}

			updatedStudy, err := client.UpdateStudy(study.ID, model.UpdateStudy{TotalAvailablePlaces: opts.Places})
			if err != nil {
				return err
			}

			fmt.Fprintln(w, studyui.RenderStudy(*updatedStudy))

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Places, "places", "p", 0, "The number of places you want to set on your study.")

	return cmd
}
