package participantgroup

import (
	"errors"
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
	File string
}

// NewRemoveCommand creates a new command to remove participants from a participant group.
func NewRemoveCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts RemoveOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(1),
		Short: "Remove participants from a participant group",
		Long: `Remove one or more participants from a participant group

The first argument is the participant group ID. Participant IDs to remove can
be provided as additional arguments or via a CSV file (one ID per line).
Non-member participant IDs are silently ignored.
`,
		Example: `
Remove one participant from a group:

$ prolific participant remove 6429b0ea05b2a24cac83c3a4 abc123def456

Remove multiple participants at once:

$ prolific participant remove 6429b0ea05b2a24cac83c3a4 abc123def456 789xyz012

Remove participants listed in a CSV file:

$ prolific participant remove 6429b0ea05b2a24cac83c3a4 --file participants.csv
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

	flags := cmd.Flags()
	flags.StringVarP(&opts.File, "file", "f", "", "Path to a file containing one participant ID per line")

	return cmd
}

func removeParticipants(client client.API, opts RemoveOptions, w io.Writer) error {
	groupID := opts.Args[0]
	positionalIDs := opts.Args[1:]

	if opts.File != "" && len(positionalIDs) > 0 {
		return errors.New("cannot use both --file and positional participant IDs")
	}

	if opts.File == "" && len(positionalIDs) == 0 {
		return errors.New("provide participant IDs as arguments or via --file")
	}

	var participantIDs []string
	var err error

	if opts.File != "" {
		participantIDs, err = parseParticipantFile(opts.File)
		if err != nil {
			return err
		}
	} else {
		participantIDs = positionalIDs
	}

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
