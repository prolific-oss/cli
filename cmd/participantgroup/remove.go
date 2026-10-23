package participantgroup

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewRemoveCommand creates a new command for removing participants from a participant group.
func NewRemoveCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	return newMemberCommand(memberCommand{
		short: "Remove participants from a participant group",
		long: `Remove one or more participants from an existing participant group.

You can specify participants in two ways:

1. Inline: provide one or more --participant-id flags
2. File: provide a --file flag with a file containing one participant ID per line

The two methods are mutually exclusive.`,
		example: `
  # Remove participants by ID
  prolific participant-group remove <group_id> -p <participant_id> -p <participant_id>

  # Remove participants from a file (one ID per line)
  prolific participant-group remove <group_id> -f participants.csv`,
		flagUsage: "The ID of a participant to remove. Can be specified multiple times.",
		apply: func(c client.API, groupID string, ids []string) (*client.ViewParticipantGroupResponse, error) {
			return c.RemoveParticipantGroupMembers(groupID, ids)
		},
		summary: "Removed %d participant(s) from group %s (%d remaining)",
	}, commandName, c, w)
}
