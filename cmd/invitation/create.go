package invitation

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

var validRoles = []string{"WORKSPACE_ADMIN", "WORKSPACE_COLLABORATOR"}

// CreateOptions are the options to be able to create an invitation.
type CreateOptions struct {
	Args      []string
	Workspace string
	Emails    []string
	Role      string
}

// NewCreateCommand creates a new command for creating invitations.
func NewCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create workspace invitations",
		Long: `Create invitations to invite users to collaborate on a workspace

Invitations are sent to the provided email addresses with the specified role.
The role determines whether the invitee will be a workspace admin or collaborator.
`,
		Example: `
To invite a user as a collaborator
$ prolific invitation create -w 60d9aa5fa100c40b8c3fac61 -e user@example.com -r WORKSPACE_COLLABORATOR

To invite multiple users as admins
$ prolific invitation create -w 60d9aa5fa100c40b8c3fac61 -e user1@example.com -e user2@example.com -r WORKSPACE_ADMIN
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createInvitation(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Workspace, "workspace", "w", "", "The ID of the workspace to invite users to.")
	flags.StringArrayVarP(&opts.Emails, "email", "e", nil, "Email address to invite (can be specified multiple times).")
	flags.StringVarP(&opts.Role, "role", "r", "", "The role for the invitee: WORKSPACE_ADMIN or WORKSPACE_COLLABORATOR.")

	return cmd
}

// createInvitation will create invitations for the given emails
func createInvitation(client client.API, opts CreateOptions, w io.Writer) error {
	if opts.Workspace == "" {
		return errors.New("workspace is required")
	}

	if len(opts.Emails) == 0 {
		return errors.New("at least one email is required")
	}

	if opts.Role == "" {
		return errors.New("role is required")
	}

	if !isValidRole(opts.Role) {
		return fmt.Errorf("invalid role %q: must be one of %s", opts.Role, strings.Join(validRoles, ", "))
	}

	invitation := model.CreateInvitation{
		Association: opts.Workspace,
		Emails:      opts.Emails,
		Role:        opts.Role,
	}

	response, err := client.CreateInvitation(invitation)
	if err != nil {
		return err
	}

	for _, inv := range response.Invitations {
		fmt.Fprintf(w, "Invited %s as %s\n", inv.Invitee.Email, inv.Role)
	}

	return nil
}

func isValidRole(role string) bool {
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}
