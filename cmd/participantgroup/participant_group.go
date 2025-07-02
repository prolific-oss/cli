package participantgroup

import (
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewParticipantCommand creates a new `participant` command
func NewParticipantCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "participant",
		Short: "Manage and view your participant groups",
		Long: `List your participant groups

Participant Groups allow you to create, modify, and use lists of participants
directly within the Prolific ecosystem.

Participant groups allow you do the following:

- Create a new participant group within the scope of a project.
- Add and remove users manually to / from the participant group.
- Use one or more participant groups as eligibility requirements for a new study.
- Combined with study completion codes, automatically add or remove participants
  from a group when they submit a response to your study with the correct code.`,
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewViewCommand("view", client, w),
	)
	return cmd
}
