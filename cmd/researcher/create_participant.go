package researcher

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewCreateParticipantCommand creates a new `researcher create-participant` command to
// create a test participant for the researcher.
func NewCreateParticipantCommand(client client.API, w io.Writer) *cobra.Command {
	var email string

	cmd := &cobra.Command{
		Use:   "create-participant",
		Short: "Create a test participant",
		Long: `Create a test participant with the same details as the researcher.

The participant is created with the supplied email and bypasses fraud checks
and on-boarding steps. A randomly generated password is assigned; reset it to
log in as the participant.

The feature must be enabled for the workspace before the endpoint can be used.
The participant is limited to taking studies only in workspaces associated with
the researcher where the feature is enabled.`,
		Example: `
To create a test participant:
$ prolific researcher create-participant -e test@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := client.CreateTestParticipant(email)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintf(w, "Created test participant: %s\n", response.ParticipantID)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&email, "email", "e", "", "The email of the test participant (required)")

	_ = cmd.MarkFlagRequired("email")

	return cmd
}
