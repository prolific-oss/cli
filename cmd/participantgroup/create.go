package participantgroup

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateOptions are the options to be able to create a participant group.
type CreateOptions struct {
	Args           []string
	Name           string
	WorkspaceID    string
	Description    string
	ParticipantIDs []string
}

// NewCreateCommand creates a new command for creating a participant group.
func NewCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create a participant group",
		Long: `Create a participant group in your workspace

Participant groups allow you to create and manage lists of participants
directly within the Prolific ecosystem. You can then use these groups as
eligibility requirements for studies.`,
		Example: `
To create a participant group in a workspace
$ prolific participant create -N "My Group" -w <workspace_id>

To create a participant group with a description
$ prolific participant create -N "My Group" -w <workspace_id> -d "A group for repeat participants"

To create a participant group with initial participants
$ prolific participant create -N "My Group" -w <workspace_id> -p <participant_id> -p <participant_id>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createParticipantGroup(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Name, "name", "N", "", "The name of the participant group.")
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The ID of the workspace to create the participant group in.")
	flags.StringVarP(&opts.Description, "description", "d", "", "The description of the participant group.")
	flags.StringArrayVarP(&opts.ParticipantIDs, "participant-id", "p", nil, "The ID of a participant to add to the group. Can be specified multiple times.")

	return cmd
}

// createParticipantGroup will create a participant group
func createParticipantGroup(client client.API, opts CreateOptions, w io.Writer) error {
	if opts.Name == "" {
		return errors.New("name is required")
	}

	if opts.WorkspaceID == "" {
		return errors.New("workspace is required")
	}

	group := model.CreateParticipantGroup{
		Name:           opts.Name,
		WorkspaceID:    opts.WorkspaceID,
		Description:    opts.Description,
		ParticipantIDs: opts.ParticipantIDs,
	}

	record, err := client.CreateParticipantGroup(group)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Created participant group: %s\n", record.ID)

	return nil
}
