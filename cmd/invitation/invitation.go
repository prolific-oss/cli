package invitation

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewInvitationCommand creates a new `invitation` command
func NewInvitationCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invitation",
		Short: "Manage workspace invitations",
		Long: `Manage invitations on Prolific

Invitations are issued to invite users to collaborate or become admins of a
shared workspace. Each invitation contains an association to a workspace and
a role that determines the invitee's permissions.
`,
	}

	cmd.AddCommand(
		NewCreateCommand("create", client, w),
	)
	return cmd
}
