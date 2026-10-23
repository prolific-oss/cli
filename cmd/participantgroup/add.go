package participantgroup

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewAddCommand creates a new command for adding participants to a participant group.
func NewAddCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	return newMemberCommand(memberCommand{
		short: "Add participants to a participant group",
		long: `Add one or more participants to an existing participant group.

You can specify participants in two ways:

1. Inline: provide one or more --participant-id flags
2. File: provide a --file flag with a file containing one participant ID per line

The two methods are mutually exclusive.`,
		example: `
  # Add participants by ID
  prolific participant-group add <group_id> -p <participant_id> -p <participant_id>

  # Add participants from a file (one ID per line)
  prolific participant-group add <group_id> -f participants.csv`,
		flagUsage: "The ID of a participant to add. Can be specified multiple times.",
		apply: func(c client.API, groupID string, ids []string) (*client.ViewParticipantGroupResponse, error) {
			return c.AddParticipantGroupMembers(groupID, ids)
		},
		summary: "Added %d participant(s) to group %s (%d total)",
	}, commandName, c, w)
}
