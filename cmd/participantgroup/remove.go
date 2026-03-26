package participantgroup

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// RemoveOptions is the options for the remove participants command.
type RemoveOptions struct {
	Args []string
}

// NewRemoveCommand creates a new command to remove participants from a participant group.
func NewRemoveCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts RemoveOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(2),
		Short: "Remove participants from a participant group",
		Long: `Remove one or more participants from a participant group

The first argument is the participant group ID. All subsequent arguments are
participant IDs to remove. Non-member participant IDs are silently ignored.
`,
		Example: `
Remove one participant from a group:

$ prolific participant remove 6429b0ea05b2a24cac83c3a4 abc123def456

Remove multiple participants at once:

$ prolific participant remove 6429b0ea05b2a24cac83c3a4 abc123def456 789xyz012
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := removeParticipants(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

func removeParticipants(client client.API, opts RemoveOptions, w io.Writer) error {
	groupID := opts.Args[0]
	participantIDs := opts.Args[1:]

	membership, err := client.RemoveParticipantsFromGroup(groupID, participantIDs)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "Participant ID", "Date added")
	for _, p := range membership.Results {
		fmt.Fprintf(tw, "%s\t%s\n", p.ParticipantID, p.DatetimeCreated.Format(ui.AppDateTimeFormat))
	}

	return tw.Flush()
}
