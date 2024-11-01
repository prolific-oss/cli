package study

import (
	"fmt"
	"io"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	studyui "github.com/benmatselby/prolificli/ui/study"
	"github.com/spf13/cobra"
)

// TransitionOptions is the options for transitioning a study command.
type TransitionOptions struct {
	Args   []string
	Action string
	Silent bool
}

// NewTransitionCommand creates a new `study transition` command to allow you
// change the study status.
func NewTransitionCommand(client client.API, w io.Writer) *cobra.Command {
	var opts TransitionOptions

	cmd := &cobra.Command{
		Use:     "transition",
		Short:   "Transition the status of a study",
		Long:    `You can pause, start, stop or publish a study`,
		Example: ``,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := transitionStudy(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Action, "action", "a", "", fmt.Sprintf("Transition a study, it can be one of %s", strings.Join(model.TransitionList, ", ")))
	flags.BoolVarP(&opts.Silent, "silent", "s", false, "Silently create the study. It will not render the study once created.")

	return cmd
}

func transitionStudy(client client.API, opts TransitionOptions, w io.Writer) error {
	if opts.Action == "" {
		return fmt.Errorf("you must provide an action to transition the study to")
	}

	_, err := client.TransitionStudy(opts.Args[0], opts.Action)
	if err != nil {
		return err
	}

	if !opts.Silent {
		study, err := client.GetStudy(opts.Args[0])
		if err != nil {
			return err
		}

		fmt.Fprintln(w, studyui.RenderStudy(*study))
	}

	return nil
}
