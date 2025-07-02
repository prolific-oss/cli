package participantgroup

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
)

// ViewOptions is the options for the detail view of a participant group command.
type ViewOptions struct {
	Args []string
}

// NewViewCommand creates a new command to show a participant group.
func NewViewCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(1),
		Short: "Provide details about your participant group",
		Long: `View your participant group

A participant group contains one or more participants.
`,
		Example: `
List the participants in your participant group

$ prolific participant view 6429b0ea05b2a24cac83c3a4
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderGroup(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

// renderGroup will show your participant group
func renderGroup(client client.API, opts ListOptions, w io.Writer) error {
	if len(opts.Args) < 1 || opts.Args[0] == "" {
		return errors.New("please provide a participant group ID")
	}

	membership, err := client.GetParticipantGroup(opts.Args[0])
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "Participant ID", "Date added")
	for _, participant := range membership.Results {
		fmt.Fprintf(tw, "%s\t%s\n", participant.ParticipantID, participant.DatetimeCreated.Format(ui.AppDateTimeFormat))
	}

	return tw.Flush()
}
