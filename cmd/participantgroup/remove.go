package participantgroup

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
)

// RemoveOptions are the options for removing participants from a group.
type RemoveOptions struct {
	Args           []string
	ParticipantIDs []string
	File           string
}

// NewRemoveCommand creates a new command for removing participants from a participant group.
func NewRemoveCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts RemoveOptions

	cmd := &cobra.Command{
		Use:   commandName + " <group-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Remove participants from a participant group",
		Long: `Remove one or more participants from an existing participant group.

You can specify participants in two ways:

1. Inline: provide one or more --participant-id flags
2. File: provide a --file flag with a file containing one participant ID per line

The two methods are mutually exclusive.`,
		Example: `
  # Remove participants by ID
  prolific participant remove <group_id> -p <participant_id> -p <participant_id>

  # Remove participants from a file (one ID per line)
  prolific participant remove <group_id> -f participants.csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := removeParticipants(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&opts.ParticipantIDs, "participant-id", "p", nil, "The ID of a participant to remove. Can be specified multiple times.")
	flags.StringVarP(&opts.File, "file", "f", "", "Path to a file containing one participant ID per line.")

	return cmd
}

func removeParticipants(c client.API, opts RemoveOptions, w io.Writer) error {
	groupID := opts.Args[0]

	hasInlineIDs := len(opts.ParticipantIDs) > 0
	hasFile := opts.File != ""

	if hasFile && hasInlineIDs {
		return fmt.Errorf("cannot use --file together with --participant-id")
	}

	if hasFile {
		ids, err := shared.ParseIDFile(opts.File)
		if err != nil {
			return err
		}
		opts.ParticipantIDs = ids
	}

	if len(opts.ParticipantIDs) == 0 {
		return fmt.Errorf("you must provide at least one participant ID via --participant-id or --file")
	}

	response, err := c.RemoveParticipantGroupMembers(groupID, opts.ParticipantIDs)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Removed %d participant(s) from group %s (%d remaining)\n", len(opts.ParticipantIDs), groupID, len(response.Results))

	return nil
}
