package study

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"

	"github.com/spf13/cobra"
)

// IncreasePlacesOptions represents the options for the increase-places command.
type IncreasePlacesOptions struct {
	Args              []string
	Places            int
	AnnotatorsPerTask int
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
to a wider audience.

The places on a Task Builder study are worked out from its batch: the number of task
groups times the annotators per task. To increase them, set --annotators-per-task to a
higher number instead of --places. The new total is shown once it is updated.`,
		Example: `
$ prolific study increase-places 64395e9c2332b8a59a65d51e -p 300
$ prolific study increase-places 64395e9c2332b8a59a65d51e --places 5000

For a Task Builder study:
$ prolific study increase-places 64395e9c2332b8a59a65d51e --annotators-per-task 3`,
		Short: "Increase the total available places on a study",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			study, err := client.GetStudy(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			if study.IsTaskBuilderBatch() {
				return increaseTaskBuilderPlaces(client, w, *study, cmd.Flags().Changed("places"), opts.AnnotatorsPerTask)
			}

			if opts.AnnotatorsPerTask != 0 {
				return errors.New("--annotators-per-task only applies to a Task Builder study; use --places for this one")
			}

			if study.TotalAvailablePlaces > opts.Places {
				return fmt.Errorf("study currently has %v places, and you cannot decrease the available places to %v", study.TotalAvailablePlaces, opts.Places)
			}

			updatedStudy, err := client.UpdateStudy(study.ID, model.UpdateStudy{TotalAvailablePlaces: opts.Places})
			if err != nil {
				return err
			}

			fmt.Fprintln(w, RenderStudy(*updatedStudy))

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Places, "places", "p", 0, "The number of places you want to set on your study.")
	flags.IntVar(&opts.AnnotatorsPerTask, "annotators-per-task", 0, "For a Task Builder study: the number of participants for each task, which sets its places.")

	return cmd
}

// increaseTaskBuilderPlaces raises the annotators per task of a Task Builder study. Its places are calculated by the API
// from that number, so the places cannot be set directly, and the API refuses a decrease on a published study.
func increaseTaskBuilderPlaces(c client.API, w io.Writer, study model.Study, placesGiven bool, annotatorsPerTask int) error {
	if placesGiven {
		return errors.New("the places of a Task Builder study are worked out from its batch, so use --annotators-per-task to increase them, not --places")
	}
	if annotatorsPerTask < 1 {
		return errors.New("a Task Builder study's places are set with --annotators-per-task, a number of 1 or more")
	}
	if current := study.AnnotatorsPerTask(); annotatorsPerTask < current {
		return fmt.Errorf("study currently has %v annotators per task, and you cannot decrease it to %v", current, annotatorsPerTask)
	}

	updatedStudy, err := c.UpdateStudy(study.ID, model.UpdateStudy{
		DataCollectionMetadata: map[string]any{"annotators_per_task": annotatorsPerTask},
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(w, RenderStudy(*updatedStudy))

	return nil
}
